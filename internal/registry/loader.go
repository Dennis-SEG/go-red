package registry

import (
	"log"

	"github.com/yourusername/go-red/internal/engine"
	"github.com/yourusername/go-red/pkg/nodes/input"
	"github.com/yourusername/go-red/pkg/nodes/output"
	"github.com/yourusername/go-red/pkg/nodes/process"
)

// LoadBuiltinNodes loads all built-in node types
func (r *Registry) LoadBuiltinNodes() error {
	// Input nodes
	input.RegisterHTTPInputNode(r)
	log.Println("Registered HTTP input node")
	
	// Process nodes
	process.RegisterFunctionNode(r)
	log.Println("Registered Function node")
	
	// Output nodes
	output.RegisterDebugNode(r)
	log.Println("Registered Debug node")
	
	return nil
}
