package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Node represents a processing node in a flow
type Node struct {
	ID     string
	Name   string
	Type   *NodeType
	Config json.RawMessage
	flow   *Flow
	
	instance NodeInstance
	wires    [][]NodeInstance
	running  bool
	mu       sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

// NodeType represents a type of node (e.g., HTTP Input, Function, etc.)
type NodeType struct {
	Name        string
	Description string
	Category    string
	Defaults    json.RawMessage
	Factory     NodeFactory
}

// NodeFactory is a function that creates a specific node instance
type NodeFactory func() NodeInstance

// NodeInstance is the interface that all node implementations must satisfy
type NodeInstance interface {
	// Init initializes the node with its configuration
	Init(config json.RawMessage) error
	
	// Start starts the node
	Start(ctx context.Context) error
	
	// Stop stops the node
	Stop()
	
	// OnMessage processes a message
	OnMessage(msg *Message, port int) error
	
	// GetNode returns the parent Node structure
	GetNode() *Node
	
	// SetNode sets the parent Node structure
	SetNode(node *Node)
}

// NewNode creates a new Node instance
func NewNode(id, name string, nodeType *NodeType, config json.RawMessage, flow *Flow) (*Node, error) {
	node := &Node{
		ID:     id,
		Name:   name,
		Type:   nodeType,
		Config: config,
		flow:   flow,
		wires:  make([][]NodeInstance, 0),
	}

	// Create the node instance
	instance := nodeType.Factory()
	instance.SetNode(node)
	
	// Initialize with configuration
	if err := instance.Init(config); err != nil {
		return nil, fmt.Errorf("failed to initialize node instance: %w", err)
	}
	
	node.instance = instance
	return node, nil
}

// Start starts the node
func (n *Node) Start(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	if n.running {
		return fmt.Errorf("node %s is already running", n.ID)
	}
	
	n.ctx, n.cancel = context.WithCancel(ctx)
	if err := n.instance.Start(n.ctx); err != nil {
		n.cancel()
		return err
	}
	
	n.running = true
	return nil
}

// Stop stops the node
func (n *Node) Stop() {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	if !n.running {
		return
	}
	
	n.instance.Stop()
	if n.cancel != nil {
		n.cancel()
	}
	
	n.running = false
}

// Send sends a message to connected nodes
func (n *Node) Send(msg *Message, port int) error {
	n.mu.RLock()
	defer n.mu.RUnlock()
	
	if !n.running {
		return fmt.Errorf("node %s is not running", n.ID)
	}
	
	if port >= len(n.wires) {
		return nil // No wires connected to this port
	}
	
	for _, target := range n.wires[port] {
		// Clone the message for each target to prevent concurrent modification
		msgCopy := msg.Clone()
		
		// Send the message to the target node
		if err := target.OnMessage(msgCopy, 0); err != nil {
			return fmt.Errorf("error sending message to node: %w", err)
		}
	}
	
	return nil
}

// AddWire connects this node to another node
func (n *Node) AddWire(port int, target *Node) {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	// Ensure we have enough ports
	for len(n.wires) <= port {
		n.wires = append(n.wires, make([]NodeInstance, 0))
	}
	
	n.wires[port] = append(n.wires[port], target.instance)
}

// GetWires returns the node's wires
func (n *Node) GetWires() [][]NodeInstance {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.wires
}

// IsRunning returns whether the node is running
func (n *Node) IsRunning() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.running
}

// GetFlow returns the node's parent flow
func (n *Node) GetFlow() *Flow {
	return n.flow
}
