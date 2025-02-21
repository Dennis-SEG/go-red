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
