package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/yourusername/go-red/internal/registry"
	"github.com/yourusername/go-red/internal/storage"
)

// Engine represents the flow execution engine
type Engine struct {
	registry *registry.Registry
	storage  storage.Storage
	flows    map[string]*Flow
	status   Status
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.RWMutex
}

// Status represents the engine status
type Status string

const (
	StatusStopped Status = "stopped"
	StatusRunning Status = "running"
	StatusError   Status = "error"
)

// New creates a new Engine instance
func New(reg *registry.Registry, store storage.Storage) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{
		registry: reg,
		storage:  store,
		flows:    make(map[string]*Flow),
		status:   StatusStopped,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Initialize prepares the engine for operation
func (e *Engine) Initialize() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Load all flows from storage
	flowIDs, err := e.storage.ListFlows()
	if err != nil {
		return fmt.Errorf("failed to list flows: %w", err)
	}

	for _, id := range flowIDs {
		flowDef, err := e.storage.LoadFlow(id)
		if err != nil {
			log.Printf("Warning: Failed to load flow %s: %v", id, err)
			continue
		}

		flow, err := NewFlow(id, flowDef, e)
		if err != nil {
			log.Printf("Warning: Failed to create flow %s: %v", id, err)
			continue
		}

		e.flows[id] = flow
	}

	return nil
}

// Start starts the engine and all flows
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status == StatusRunning {
		return errors.New("engine is already running")
	}

	for id, flow := range e.flows {
		if err := flow.Start(e.ctx); err != nil {
			log.Printf("Warning: Failed to start flow %s: %v", id, err)
		}
	}

	e.status = StatusRunning
	return nil
}

// Stop stops the engine and all flows
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != StatusRunning {
		return errors.New("engine is not running")
	}

	e.cancel()
	e.ctx, e.cancel = context.WithCancel(context.Background())

	for _, flow := range e.flows {
		flow.Stop()
	}

	e.status = StatusStopped
	return nil
}

// DeployFlow deploys a new or updated flow
func (e *Engine) DeployFlow(id string, flowDef []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Stop existing flow if it exists
	if existingFlow, exists := e.flows[id]; exists {
		existingFlow.Stop()
	}

	// Save flow to storage
	if err := e.storage.SaveFlow(id, flowDef); err != nil {
		return fmt.Errorf("failed to save flow: %w", err)
	}

	// Create new flow
	flow, err := NewFlow(id, flowDef, e)
	if err != nil {
		return fmt.Errorf("failed to create flow: %w", err)
	}

	e.flows[id] = flow

	// Start the flow if engine is running
	if e.status == StatusRunning {
		if err := flow.Start(e.ctx); err != nil {
			return fmt.Errorf("failed to start flow: %w", err)
		}
	}

	return nil
}

// GetFlow returns a flow by ID
func (e *Engine) GetFlow(id string) (*Flow, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	flow, exists := e.flows[id]
	return flow, exists
}

// ListFlows returns a list of all flow IDs
func (e *Engine) ListFlows() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	flows := make([]string, 0, len(e.flows))
	for id := range e.flows {
		flows = append(flows, id)
	}
	return flows
}

// DeleteFlow removes a flow
func (e *Engine) DeleteFlow(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Stop and remove flow if it exists
	if flow, exists := e.flows[id]; exists {
		flow.Stop()
		delete(e.flows, id)
	}

	// Remove from storage
	return e.storage.DeleteFlow(id)
}

// GetRegistry returns the node registry
func (e *Engine) GetRegistry() *registry.Registry {
	return e.registry
}

// Status returns the current engine status
func (e *Engine) Status() Status {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}
