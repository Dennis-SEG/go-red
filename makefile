package server

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// WebUIHandler serves the Web UI
type WebUIHandler struct {
	rootPath string
	indexFile string
}

// NewWebUIHandler creates a new WebUIHandler
func NewWebUIHandler(rootPath string) *WebUIHandler {
	return &WebUIHandler{
		rootPath: rootPath,
		indexFile: "index.html",
	}
}

// ServeHTTP implements http.Handler
func (h *WebUIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean path to prevent directory traversal
	path := filepath.Clean(r.URL.Path)
	
	// Handle root path
	if path == "/" || path == "" {
		path = h.indexFile
	}
	
	// Prepend the root directory
	fullPath := filepath.Join(h.rootPath, path)
	
	// Serve file if it exists
	if fileExists(fullPath) {
		// Set appropriate content type based on file extension
		contentType := getContentType(fullPath)
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		
		// Set cache control headers for static assets
		if isStaticAsset(fullPath) {
			w.Header().Set("Cache-Control", "public, max-age=86400") // 1 day
		} else {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}
		
		http.ServeFile(w, r, fullPath)
		return
	}
	
	// If file doesn't exist, serve index.html for SPA routing
	indexPath := filepath.Join(h.rootPath, h.indexFile)
	if fileExists(indexPath) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		http.ServeFile(w, r, indexPath)
		return
	}
	
	// If index.html doesn't exist, return 404
	http.NotFound(w, r)
}

// AddWebSocketHandler adds the WebSocket handler to the router
func (s *Server) AddWebSocketHandler() {
	// Create WebSocket manager
	wsManager := NewWebSocketManager()
	go wsManager.Run()
	
	// Add WebSocket route
	s.router.HandleFunc("/ws", wsManager.HandleWebSocket)
	
	// Store manager for other handlers to use
	s.wsManager = wsManager
}

// AddDebugConsoleHandler adds the debug console handler
func (s *Server) AddDebugConsoleHandler() {
	// Create a new router for /debug path
	debugRouter := s.router.PathPrefix("/debug").Subrouter()
	
	// Add handlers for debug console
	debugRouter.HandleFunc("/", s.handleDebugHome)
	debugRouter.HandleFunc("/flows", s.handleDebugFlows)
	debugRouter.HandleFunc("/flows/{id}", s.handleDebugFlow)
	debugRouter.HandleFunc("/console", s.handleDebugConsole)
}

// handleDebugHome handles the debug home page
func (s *Server) handleDebugHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>go-red Debug Console</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
				h1 { color: #333; }
				ul { list-style-type: none; padding: 0; }
				li { margin: 10px 0; }
				a { color: #0078d4; text-decoration: none; }
				a:hover { text-decoration: underline; }
			</style>
		</head>
		<body>
			<h1>go-red Debug Console</h1>
			<ul>
				<li><a href="/debug/flows">Flows</a></li>
				<li><a href="/debug/console">Debug Console</a></li>
			</ul>
		</body>
		</html>
	`))
}

// handleDebugFlows handles the debug flows page
func (s *Server) handleDebugFlows(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	// Get all flows
	flowIDs := s.engine.ListFlows()
	
	// Build HTML
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>go-red Flows</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
				h1 { color: #333; }
				ul { list-style-type: none; padding: 0; }
				li { margin: 10px 0; }
				a { color: #0078d4; text-decoration: none; }
				a:hover { text-decoration: underline; }
			</style>
		</head>
		<body>
			<h1>Flows</h1>
			<ul>
	`
	
	for _, id := range flowIDs {
		flow, exists := s.engine.GetFlow(id)
		if !exists {
			continue
		}
		
		html += `<li><a href="/debug/flows/` + id + `">` + id + `</a> - Status: ` + string(flow.GetStatus()) + `</li>`
	}
	
	html += `
			</ul>
			<p><a href="/debug/">Back to Debug Home</a></p>
		</body>
		</html>
	`
	
	w.Write([]byte(html))
}

// handleDebugFlow handles the debug flow detail page
func (s *Server) handleDebugFlow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	flow, exists := s.engine.GetFlow(id)
	if !exists {
		http.NotFound(w, r)
		return
	}
	
	// Get flow JSON
	flowJSON, err := flow.ToJSON()
	if err != nil {
		http.Error(w, "Failed to get flow JSON", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "text/html")
	
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Flow: ` + id + `</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
				h1, h2 { color: #333; }
				pre { background-color: #f5f5f5; padding: 10px; border-radius: 5px; overflow: auto; }
				.button { display: inline-block; padding: 8px 16px; background-color: #0078d4; color: white; 
						 border-radius: 4px; text-decoration: none; margin-right: 10px; }
				.button:hover { background-color: #005a9e; }
			</style>
		</head>
		<body>
			<h1>Flow: ` + id + `</h1>
			<p>Status: ` + string(flow.GetStatus()) + `</p>
			<div>
				<a href="/api/flows/` + id + `/start" class="button">Start Flow</a>
				<a href="/api/flows/` + id + `/stop" class="button">Stop Flow</a>
			</div>
			<h2>Flow Definition</h2>
			<pre>` + string(flowJSON) + `</pre>
			<p><a href="/debug/flows">Back to Flows</a></p>
		</body>
		</html>
	`
	
	w.Write([]byte(html))
}

// handleDebugConsole handles the debug console page
func (s *Server) handleDebugConsole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>go-red Debug Console</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
				h1 { color: #333; }
				#console { background-color: #f5f5f5; padding: 10px; border-radius: 5px; 
						  height: 400px; overflow: auto; margin-bottom: 10px; }
				.message { margin: 5px 0; padding: 5px; border-bottom: 1px solid #ddd; }
				.time { color: #666; font-size: 0.8em; }
				.error { color: #d9534f; }
				.info { color: #5bc0de; }
				.warn { color: #f0ad4e; }
			</style>
			<script>
				let ws;
				
				function connect() {
					const host = window.location.host;
					ws = new WebSocket('ws://' + host + '/ws');
					
					ws.onopen = function() {
						addMessage('Connected to server', 'info');
					};
					
					ws.onmessage = function(event) {
						const data = JSON.parse(event.data);
						addMessage('Received: ' + JSON.stringify(data), 'info');
					};
					
					ws.onclose = function() {
						addMessage('Disconnected from server', 'warn');
						// Try to reconnect after 5 seconds
						setTimeout(connect, 5000);
					};
					
					ws.onerror = function(error) {
						addMessage('Error: ' + error, 'error');
					};
				}
				
				function addMessage(message, type) {
					const console = document.getElementById('console');
					const time = new Date().toLocaleTimeString();
					const messageElem = document.createElement('div');
					messageElem.className = 'message ' + (type || '');
					messageElem.innerHTML = '<span class="time">[' + time + ']</span> ' + message;
					console.appendChild(messageElem);
					console.scrollTop = console.scrollHeight;
				}
				
				window.onload = function() {
					connect();
				};
			</script>
		</head>
		<body>
			<h1>Debug Console</h1>
			<div id="console"></div>
			<p><a href="/debug/">Back to Debug Home</a></p>
		</body>
		</html>
	`
	
	w.Write([]byte(html))
}

// Helper functions

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := http.Dir(".").Open(path)
	if err != nil {
		return false
	}
	info.Close()
	return true
}

// getContentType returns the content type based on file extension
func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	default:
		return ""
	}
}

// isStaticAsset checks if a file is a static asset
func isStaticAsset(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	staticExts := []string{".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot"}
	
	for _, staticExt := range staticExts {
		if ext == staticExt {
			return true
		}
	}
	
	return false
}
