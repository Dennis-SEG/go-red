package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/yourusername/go-red/internal/engine"
)

// Registry manages all available node types
type Registry struct {
	nodeTypes map[string]*engine.NodeType
	mu        sync.RWMutex
}

// New creates a new Registry
func New() *Registry {
	return &Registry{
		nodeTypes: make(map[string]*engine.NodeType),
	}
}

// RegisterNodeType registers a new node type
func (r *Registry) RegisterNodeType(nodeType *engine.NodeType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodeTypes[nodeType.Name]; exists {
		return fmt.Errorf("node type %s is already registered", nodeType.Name)
	}

	r.nodeTypes[nodeType.Name] = nodeType
	return nil
}

// GetNodeType gets a node type by name
func (r *Registry) GetNodeType(name string) (*engine.NodeType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodeType, exists := r.nodeTypes[name]
	if !exists {
		return nil, fmt.Errorf("node type %s not found", name)
	}

	return nodeType, nil
}

// UnregisterNodeType removes a node type from the registry
func (r *Registry) UnregisterNodeType(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodeTypes[name]; !exists {
		return fmt.Errorf("node type %s not found", name)
	}

	delete(r.nodeTypes, name)
	return nil
}

// GetAllNodeTypes returns all registered node types
func (r *Registry) GetAllNodeTypes() []*engine.NodeType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]*engine.NodeType, 0, len(r.nodeTypes))
	for _, nodeType := range r.nodeTypes {
		types = append(types, nodeType)
	}

	return types
}

// LoadBuiltinNodes loads all built-in node types
func (r *Registry) LoadBuiltinNodes() error {
	// This will be implemented with actual node types
	// For now, just a placeholder that succeeds
	return nil
}

// LoadNodePlugin loads a node plugin from a file
func (r *Registry) LoadNodePlugin(path string) error {
	// This requires Go plugin support
	// It's a bit complex and platform-dependent
	// For now, we'll return an error
	return errors.New("plugin loading not implemented yet")
}
