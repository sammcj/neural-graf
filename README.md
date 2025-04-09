# MCP-Graph

A lightweight, self-hosted knowledge graph system built in Go with Dgraph integration and MCP server support.

## Overview

MCP-Graph is a modular knowledge graph system that provides efficient data ingestion, graph-based storage, and powerful querying capabilities using Dgraph as the primary database. The system includes the following key components:

- **Knowledge Graph**: Core graph database functionality using Dgraph
- **MCP Server**: Deployed as a Model Context Protocol server (using mark3labs/mcp-go) to provide standardised LLM tool interfaces

## Features

- Clean, modular architecture with well-defined interfaces
- Flexible deployment options (standalone, containerised)
- Powerful graph querying capabilities
- MCP server integration for AI applications
- Standardised knowledge graph operations

## System Architecture

The system follows a clean, modular design with the following components:

```tree
mcp-graph/
├── cmd/
│   └── server/                 # Main application entry point
├── internal/
│   ├── api/                    # API handlers
│   ├── config/                 # Configuration management
│   ├── graph/                  # Knowledge graph implementation
│   │   └── dgraph/             # Dgraph implementation
│   ├── mcp/                    # MCP server
│   └── service/                # Core business logic
├── pkg/
│   ├── models/                 # Shared data models
│   └── utils/                  # Utility functions
└── scripts/                    # Deployment and tooling scripts
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Dgraph (can be run via Docker)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/sammcj/mcp-graph.git
   cd mcp-graph
   ```

2. Build the application:
   ```bash
   go build -o bin/mcp-graph ./cmd/server
   ```

3. Run with Docker Compose (includes Dgraph):
   ```bash
   docker-compose up -d
   ```

   Alternatively, for macOS users with Colima who experience issues with Docker Compose, see [Running Without Docker Compose](docs/running_without_docker_compose.md).

## Configuration

Configuration can be provided via a YAML file or environment variables. See `.env.example` and `config.yaml.example` for available options.

## Usage

### API Endpoints

The service provides RESTful API endpoints for interacting with the knowledge graph.

### MCP Server

The MCP server can be used with compatible LLM applications like Claude Desktop or Cline.

## Future Enhancements

- Visual graph explorer
- Import/export functionality
- Advanced query capabilities
- Performance optimisations

## License

[MIT](LICENSE)
