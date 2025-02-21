graph TD
    A[Web UI] <--> B[Go HTTP Server]
    B <--> C[Flow Engine]
    C --> D[Node Registry]
    C --> E[Flow Storage]
    C --> F[Runtime Execution]
    F --> G[Standard Nodes]
    F --> H[Custom Nodes]
    B --> I[WebSocket Server]
    I <--> A
    J[CLI] --> B
