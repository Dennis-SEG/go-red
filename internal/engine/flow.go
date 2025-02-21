package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Flow represents a complete flow with nodes and connections
type Flow struct {
	ID          string
	Name        string
	Description string
	Nodes       map[string]*Node
	Wires       map[string][]string // Source node ID -> Target node IDs
	engine      *Engine
	mu          sync.RWMutex
	status      FlowStatus
}

// FlowStatus represents the status of a flow
type FlowStatus string

const (
	FlowStatusStopped FlowStatus = "stopped"
	FlowStatusRunning FlowStatus = "running"
	FlowStatusError   FlowStatus = "error"
)

// FlowDefinition represents the JSON structure of a flow
type FlowDefinition struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Nodes       []NodeDefinition `json:"nodes"`
	Wires       []WireDefinition `json:"wires"`
}

// NodeDefinition represents the JSON structure of a node
type NodeDefinition struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Name     string          `json:"name"`
	Config   json.RawMessage `json:"config"`
	Position Position        `json:"position"`
}

// WireDefinition represents a connection between nodes
type WireDefinition struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Port   int    `json:"port"`
}

// Position represents a node's position in the editor
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NewFlow creates a new Flow from its JSON definition
func NewFlow(id string, flowDef []byte, engine *Engine) (*Flow, error) {
	var def FlowDefinition
	if err := json.Unmarshal(flowDef, &def); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flow definition: %w", err)
	}

	if def.ID == "" {
		def.ID = id
	}

	// Create flow
	flow := &Flow{
		ID:          def.ID,
		Name:        def.Name,
		Description: def.Description,
		Nodes:       make(map[string]*Node),
		Wires:       make(map[string][]string),
		engine:      engine,
		status:      FlowStatusStopped,
	}

	// Create nodes
	for _, nodeDef := range def.Nodes {
		nodeType, err := engine.GetRegistry().GetNodeType(nodeDef.Type)
		if err != nil {
			return nil, fmt.Errorf("unknown node type: %s", nodeDef.Type)
		}

		node, err := NewNode(nodeDef.ID, nodeDef.Name, nodeType, nodeDef.Config, flow)
		if err != nil {
			return nil, fmt.Errorf("failed to create node %s: %w", nodeDef.ID, err)
		}

		flow.Nodes[nodeDef.ID] = node
	}

	// Connect wires
	for _, wireDef := range def.Wires {
		sourceNode, exists := flow.Nodes[wireDef.Source]
		if !exists {
			return nil, fmt.Errorf("wire source node not found: %s", wireDef.Source)
		}

		targetNode, exists := flow.Nodes[wireDef.Target]
		if !exists {
			return nil, fmt.Errorf("wire target node not found: %s", wireDef.Target)
		}

		// Add to wires map
		flow.Wires[wireDef.Source] = append(flow.Wires[wireDef.Source], wireDef.Target)

		// Connect nodes
		sourceNode.AddWire(wireDef.Port, targetNode)
	}

	return flow, nil
}

// Start starts all nodes in the flow
func (f *Flow) Start(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.status == FlowStatusRunning {
		return fmt.Errorf("flow %s is already running", f.ID)
	}

	for _, node := range f.Nodes {
		if err := node.Start(ctx); err != nil {
			return fmt.Errorf("failed to start node %s: %w", node.ID, err)
		}
	}

	f.status = FlowStatusRunning
	return nil
}

// Stop stops all nodes in the flow
func (f *Flow) Stop() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.status != FlowStatusRunning {
		return
	}

	for _, node := range f.Nodes {
		node.Stop()
	}

	f.status = FlowStatusStopped
}

// ToJSON converts the flow to its JSON representation
func (f *Flow) ToJSON() ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	def := FlowDefinition{
		ID:          f.ID,
		Name:        f.Name,
		Description: f.Description,
	}

	// Convert nodes
	for _, node := range f.Nodes {
		nodeDef := NodeDefinition{
			ID:     node.ID,
			Type:   node.Type.Name,
			Name:   node.Name,
			Config: node.Config,
		}
		def.Nodes = append(def.Nodes, nodeDef)
	}

	// Convert wires
	for sourceID, targets := range f.Wires {
		for port, targetID := range targets {
			wireDef := WireDefinition{
				Source: sourceID,
				Target: targetID,
				Port:   port,
			}
			def.Wires = append(def.Wires, wireDef)
		}
	}

	return json.Marshal(def)
}

// GetStatus returns the current flow status
func (f *Flow) GetStatus() FlowStatus {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.status
}

// GetNode returns a node by ID
func (f *Flow) GetNode(id string) (*Node, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	node, exists := f.Nodes[id]
	return node, exists
}
