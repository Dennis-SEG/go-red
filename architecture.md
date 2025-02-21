# Go-RED Architecture

```mermaid
graph TD
    A[Web UI]  B[Go HTTP Server]
    B  C[Flow Engine]
    C --> D[Node Registry]
    C --> E[Flow Storage]
    C --> F[Runtime Execution]
    F --> G[Standard Nodes]
    F --> H[Custom Nodes]
    B --> I[WebSocket Server]
    I  A
    J[CLI] --> B
```

This architecture diagram shows the main components of the Go-RED system and how they interact:

- **Web UI**: The browser-based flow editor
- **Go HTTP Server**: Serves the web UI and API endpoints
- **Flow Engine**: Core component that manages flows
- **Node Registry**: Maintains a registry of available node types
- **Flow Storage**: Persists flows to disk
- **Runtime Execution**: Executes flows by running nodes and passing messages
- **Standard Nodes**: Built-in node types (HTTP, Function, Debug, etc.)
- **Custom Nodes**: User-created node types
- **WebSocket Server**: Provides real-time communication with the Web UI
- **CLI**: Command-line interface for controlling the server
