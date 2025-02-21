package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yourusername/go-red/internal/engine"
)

// WebSocketManager manages WebSocket connections
type WebSocketManager struct {
	clients    map[*WebSocketClient]bool
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	broadcast  chan []byte
	mu         sync.RWMutex
}

// WebSocketClient represents a WebSocket client
type WebSocketClient struct {
	manager  *WebSocketManager
	conn     *websocket.Conn
	send     chan []byte
	flowID   string
	userID   string
	lastPing time.Time
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// NewWebSocketManager creates a new WebSocketManager
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*WebSocketClient]bool),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		broadcast:  make(chan []byte),
	}
}

// Run starts the WebSocketManager
func (m *WebSocketManager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client] = true
			m.mu.Unlock()
		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.send)
			}
			m.mu.Unlock()
		case message := <-m.broadcast:
			m.mu.RLock()
			for client := range m.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(m.clients, client)
				}
			}
			m.mu.RUnlock()
		}
	}
}

// BroadcastToAll sends a message to all clients
func (m *WebSocketManager) BroadcastToAll(message []byte) {
	m.broadcast <- message
}

// BroadcastToFlow sends a message to all clients subscribed to a flow
func (m *WebSocketManager) BroadcastToFlow(flowID string, message []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for client := range m.clients {
		if client.flowID == flowID {
			select {
			case client.send <- message:
			default:
				// Client send buffer is full, skip
			}
		}
	}
}

// HandleWebSocket handles WebSocket connections
func (m *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for now
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &WebSocketClient{
		manager:  m,
		conn:     conn,
		send:     make(chan []byte, 256),
		lastPing: time.Now(),
	}

	// Get flowID from query parameters
	flowID := r.URL.Query().Get("flowId")
	if flowID != "" {
		client.flowID = flowID
	}
	
	// Get userID from query parameters
	userID := r.URL.Query().Get("userId")
	if userID != "" {
		client.userID = userID
	}
	
	// Register client
	m.register <- client
	
	// Start goroutines for reading and writing
	go client.readPump()
	go client.writePump()
	
	// Send welcome message
	welcome := WebSocketMessage{
		Type: "welcome",
		Payload: json.RawMessage(`{"message": "Connected to go-red server"}`),
	}
	
	welcomeJSON, _ := json.Marshal(welcome)
	client.send <- welcomeJSON
}

// readPump pumps messages from the WebSocket connection to the manager
func (c *WebSocketClient) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()
	
	c.conn.SetReadLimit(4096) // Maximum message size
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.lastPing = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// Handle received message
		var wsMessage WebSocketMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			log.Printf("Failed to unmarshal WebSocket message: %v", err)
			continue
		}
		
		// Process message based on type
		switch wsMessage.Type {
		case "ping":
			// Send pong response
			pong := WebSocketMessage{
				Type: "pong",
				Payload: json.RawMessage(`{"time": "` + time.Now().Format(time.RFC3339) + `"}`),
			}
			pongJSON, _ := json.Marshal(pong)
			c.send <- pongJSON
			
		case "subscribe":
			// Subscribe to a flow
			var payload struct {
				FlowID string `json:"flowId"`
			}
			if err := json.Unmarshal(wsMessage.Payload, &payload); err != nil {
				log.Printf("Invalid subscribe payload: %v", err)
				continue
			}
			
			c.flowID = payload.FlowID
			
		case "unsubscribe":
			// Unsubscribe from a flow
			c.flowID = ""
			
		default:
			// Unknown message type, ignore
		}
	}
}

// writePump pumps messages from the client to the WebSocket connection
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// Add queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}
			
			if err := w.Close(); err != nil {
				return
			}
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
