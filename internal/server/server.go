package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/yourusername/go-red/internal/config"
	"github.com/yourusername/go-red/internal/engine"
	"github.com/yourusername/go-red/internal/storage"
)

// Server represents the HTTP server
type Server struct {
	config  *config.Config
	engine  *engine.Engine
	storage storage.Storage
	router  *mux.Router
}

// New creates a new Server instance
func New(cfg *config.Config, eng *engine.Engine, store storage.Storage) *Server {
	srv := &Server{
		config:  cfg,
		engine:  eng,
		storage: store,
		router:  mux.NewRouter(),
	}

	// Register routes
	srv.setupRoutes()

	return srv
}

// Start starts the HTTP server
func (s *Server) Start() error {
	port := s.config.GetInt("http.port")
	if port == 0 {
		port = 1880 // Default port
	}

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Handler:      s.router,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return server.ListenAndServe()
}

// setupRoutes registers all HTTP routes
func (s *Server) setupRoutes() {
	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	
	// Flows API
	api.HandleFunc("/flows", s.handleListFlows).Methods("GET")
	api.HandleFunc("/flows", s.handleCreateFlow).Methods("POST")
	api.HandleFunc("/flows/{id}", s.handleGetFlow).Methods("GET")
	api.HandleFunc("/flows/{id}", s.handleUpdateFlow).Methods("PUT")
	api.HandleFunc("/flows/{id}", s.handleDeleteFlow).Methods("DELETE")
	api.HandleFunc("/flows/{id}/start", s.handleStartFlow).Methods("POST")
	api.HandleFunc("/flows/{id}/stop", s.handleStopFlow).Methods("POST")
	
	// Nodes API
	api.HandleFunc("/nodes", s.handleListNodeTypes).Methods("GET")
	
	// Settings API
	api.HandleFunc("/settings", s.handleGetSettings).Methods("GET")
	api.HandleFunc("/settings", s.handleUpdateSettings).Methods("PUT")
	
	// Static files (Web UI)
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("web/dist")))
}

// handleListFlows handles GET /api/flows
func (s *Server) handleListFlows(w http.ResponseWriter, r *http.Request) {
	flowIDs := s.engine.ListFlows()
	flows := make([]map[string]interface{}, 0, len(flowIDs))
	
	for _, id := range flowIDs {
		flow, exists := s.engine.GetFlow(id)
		if !exists {
			continue
		}
		
		flowJSON, err := flow.ToJSON()
		if err != nil {
			continue
		}
		
		var flowMap map[string]interface{}
		if err := json.Unmarshal(flowJSON, &flowMap); err != nil {
			continue
		}
		
		// Add status
		flowMap["status"] = string(flow.GetStatus())
		flows = append(flows, flowMap)
	}
	
	respond(w, http.StatusOK, map[string]interface{}{
		"flows": flows,
	})
}

// handleCreateFlow handles POST /api/flows
func (s *Server) handleCreateFlow(w http.ResponseWriter, r *http.Request) {
	var flowDef map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&flowDef); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid flow definition")
		return
	}
	
	// Generate ID if not provided
	id, ok := flowDef["id"].(string)
	if !ok || id == "" {
		id = fmt.Sprintf("flow-%d", time.Now().UnixNano())
		flowDef["id"] = id
	}
	
	// Convert to JSON
	flowJSON, err := json.Marshal(flowDef)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to marshal flow definition")
		return
	}
	
	// Deploy flow
	if err := s.engine.DeployFlow(id, flowJSON); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to deploy flow: %v", err))
		return
	}
	
	respond(w, http.StatusCreated, map[string]interface{}{
		"id": id,
	})
}

// handleGetFlow handles GET /api/flows/{id}
func (s *Server) handleGetFlow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	flow, exists := s.engine.GetFlow(id)
	if !exists {
		respondError(w, http.StatusNotFound, "Flow not found")
		return
	}
	
	flowJSON, err := flow.ToJSON()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to marshal flow")
		return
	}
	
	var flowMap map[string]interface{}
	if err := json.Unmarshal(flowJSON, &flowMap); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to unmarshal flow")
		return
	}
	
	// Add status
	flowMap["status"] = string(flow.GetStatus())
	
	respond(w, http.StatusOK, flowMap)
}

// handleUpdateFlow handles PUT /api/flows/{id}
func (s *Server) handleUpdateFlow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	var flowDef map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&flowDef); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid flow definition")
		return
	}
	
	// Ensure ID matches
	flowDef["id"] = id
	
	// Convert to JSON
	flowJSON, err := json.Marshal(flowDef)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to marshal flow definition")
		return
	}
	
	// Deploy flow
	if err := s.engine.DeployFlow(id, flowJSON); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to deploy flow: %v", err))
		return
	}
	
	respond(w, http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

// handleDeleteFlow handles DELETE /api/flows/{id}
func (s *Server) handleDeleteFlow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if err := s.engine.DeleteFlow(id); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete flow: %v", err))
		return
	}
	
	respond(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// handleStartFlow handles POST /api/flows/{id}/start
func (s *Server) handleStartFlow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	flow, exists := s.engine.GetFlow(id)
	if !exists {
		respondError(w, http.StatusNotFound, "Flow not found")
		return
	}
	
	if err := flow.Start(r.Context()); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start flow: %v", err))
		return
	}
	
	respond(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"status":  string(flow.GetStatus()),
	})
}

// handleStopFlow handles POST /api/flows/{id}/stop
func (s *Server) handleStopFlow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	flow, exists := s.engine.GetFlow(id)
	if !exists {
		respondError(w, http.StatusNotFound, "Flow not found")
		return
	}
	
	flow.Stop()
	
	respond(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"status":  string(flow.GetStatus()),
	})
}

// handleListNodeTypes handles GET /api/nodes
func (s *Server) handleListNodeTypes(w http.ResponseWriter, r *http.Request) {
	nodeTypes := s.engine.GetRegistry().GetAllNodeTypes()
	types := make([]map[string]interface{}, 0, len(nodeTypes))
	
	for _, nt := range nodeTypes {
		types = append(types, map[string]interface{}{
			"name":        nt.Name,
			"description": nt.Description,
			"category":    nt.Category,
			"defaults":    nt.Defaults,
		})
	}
	
	respond(w, http.StatusOK, map[string]interface{}{
		"nodes": types,
	})
}

// handleGetSettings handles GET /api/settings
func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	// For now, just return a dummy response
	respond(w, http.StatusOK, map[string]interface{}{
		"httpPort": s.config.GetInt("http.port"),
		"version":  "0.1.0",
	})
}

// handleUpdateSettings handles PUT /api/settings
func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	// For now, just return a success response
	respond(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// respond sends a JSON response
func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respond(w, status, map[string]interface{}{
		"error": message,
	})
}
