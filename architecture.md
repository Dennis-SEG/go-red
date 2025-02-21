# Go-RED Architecture

```mermaid
graph TD;
    A[Web UI];
    B[Go HTTP Server];
    C[Flow Engine];
    D[Node Registry];
    E[Flow Storage];
    F[Runtime Execution];
    G[Standard Nodes];
    H[Custom Nodes];
    I[WebSocket Server];
    J[CLI];
    A  B;
    B  C;
    C --> D;
    C --> E;
    C --> F;
    F --> G;
    F --> H;
    B --> I;
    I  A;
    J --> B;
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
