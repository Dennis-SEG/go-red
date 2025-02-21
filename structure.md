# go-red Project Structure

```
go-red/
├── cmd/
│   └── go-red/
│       └── main.go           # Application entry point
├── internal/
│   ├── server/               # HTTP and WebSocket server
│   │   ├── server.go
│   │   ├── handlers.go
│   │   └── websocket.go
│   ├── engine/               # Flow execution engine
│   │   ├── engine.go
│   │   ├── node.go
│   │   ├── flow.go
│   │   └── message.go
│   ├── registry/             # Node registry
│   │   ├── registry.go
│   │   └── loader.go
│   ├── storage/              # Flow storage
│   │   ├── storage.go
│   │   └── filesystem.go
│   └── config/               # Configuration
│       └── config.go
├── pkg/
│   └── nodes/                # Standard nodes
│       ├── input/
│       │   ├── http.go
│       │   └── websocket.go
│       ├── process/
│       │   ├── function.go
│       │   └── switch.go
│       └── output/
│           ├── http.go
│           └── debug.go
├── web/                      # Web UI
│   ├── src/
│   ├── public/
│   └── package.json
├── examples/                 # Example flows
├── docs/                     # Documentation
├── go.mod
├── go.sum
├── README.md
└── Makefile
```

## Directory Structure Explanation

### cmd/
Contains the application entry points. The `go-red` directory holds the main executable.

### internal/
Contains packages that are private to this application and not meant to be imported by other applications.

- **server/**: HTTP and WebSocket server implementation
  - `server.go`: Main server implementation
  - `handlers.go`: HTTP request handlers
  - `websocket.go`: WebSocket communication

- **engine/**: Core flow execution engine
  - `engine.go`: Main engine implementation
  - `node.go`: Node representation and execution
  - `flow.go`: Flow representation and management
  - `message.go`: Message structure passed between nodes

- **registry/**: Node type registry
  - `registry.go`: Registry implementation
  - `loader.go`: Node type loader

- **storage/**: Flow storage
  - `storage.go`: Storage interface
  - `filesystem.go`: File-based storage implementation

- **config/**: Configuration management
  - `config.go`: Configuration loading and access

### pkg/
Contains packages that are public and may be imported by other applications.

- **nodes/**: Standard node implementations
  - **input/**: Input nodes
    - `http.go`: HTTP input node
    - `websocket.go`: WebSocket input node
  - **process/**: Processing nodes
    - `function.go`: JavaScript function node
    - `switch.go`: Switch/router node
  - **output/**: Output nodes
    - `http.go`: HTTP output node
    - `debug.go`: Debug output node

### web/
Web UI files for the editor.
- **src/**: Source code
- **public/**: Public assets
- **package.json**: Dependencies and scripts

### examples/
Example flows for demonstration purposes.

### docs/
Documentation files.

### Other files
- **go.mod**: Go module definition
- **go.sum**: Go module checksum
- **README.md**: Project overview
- **Makefile**: Build and development tasks
