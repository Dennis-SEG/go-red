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
