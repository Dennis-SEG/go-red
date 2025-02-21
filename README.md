# go-red

go-red is a Node-RED variant implemented in the Go programming language. It provides a flow-based programming tool for connecting hardware devices, APIs, and online services using a browser-based editor.

## Features

- Flow-based visual programming
- Web-based editor for creating and managing flows
- Real-time communication between the editor and runtime
- Built-in nodes for common tasks
- Extensible through a plugin system
- Based on Go's performance and concurrency capabilities

## Getting Started

### Prerequisites

- Go 1.17 or later
- Node.js and npm (for building the web UI)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/go-red.git
cd go-red
```

2. Build the project:

```bash
make
```

3. Run go-red:

```bash
make run
```

4. Open your browser and navigate to http://localhost:1880 to access the go-red editor.

## Project Structure

- `cmd/go-red`: Application entry point
- `internal`: Internal packages
  - `server`: HTTP and WebSocket server
  - `engine`: Flow execution engine
  - `registry`: Node registry
  - `storage`: Flow storage
  - `config`: Configuration
- `pkg`: Public packages
  - `nodes`: Standard nodes
    - `input`: Input nodes (HTTP, WebSocket, etc.)
    - `process`: Processing nodes (Function, Switch, etc.)
    - `output`: Output nodes (HTTP, Debug, etc.)
- `web`: Web UI

See [structure.md](structure.md) for a detailed project structure.

## Resource Advantages

Compared to Node-RED, go-red offers significant resource advantages:

- **Lower memory usage**: 50-70% less memory consumption
- **Better CPU efficiency**: More efficient processing with lower idle CPU usage
- **Smaller footprint**: Single binary executable that's much smaller than Node-RED
- **Faster startup**: Cold startup times under 1 second
- **Better scalability**: More efficient handling of concurrent operations

See [advantages-tradeoffs.md](advantages-tradeoffs.md) for a detailed comparison.

## Architecture

The go-red architecture consists of several components that work together to provide a seamless flow-based programming experience. The core components include the Web UI, HTTP Server, Flow Engine, and various nodes.

See [ARCHITECTURE.md](architecture.md) for the architecture diagram and detailed explanation.

## Development

To start go-red in development mode:

```bash
make dev
```

To run tests:

```bash
make test
```

## Built-in Nodes

### Input Nodes

- **HTTP Input**: Receives HTTP requests

### Process Nodes

- **Function**: Executes JavaScript code to process messages

### Output Nodes

- **Debug**: Outputs messages to the debug console

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by [Node-RED](https://nodered.org/)
