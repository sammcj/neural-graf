# Running MCP-Graph Without Docker Compose

This guide provides instructions for running MCP-Graph without Docker Compose, which is particularly useful for users who want to run Memgraph directly on their machine.

## Prerequisites

- Go 1.21 or higher
- Memgraph installed locally (see instructions below)

## Installing Memgraph Locally

You can install Memgraph directly on your system:

### macOS

```bash
brew install memgraph/memgraph/memgraph
```

### Linux

Follow the instructions at https://memgraph.com/docs/memgraph/installation

## Using the Local Memgraph Script

We've provided a script to help you run Memgraph locally without Docker Compose.

### Script Location

The script is located at `scripts/run_memgraph_locally.sh` in the project directory.

### Script Usage

Make the script executable:

```bash
chmod +x scripts/run_memgraph_locally.sh
```

#### Starting Memgraph

To start Memgraph:

```bash
./scripts/run_memgraph_locally.sh start
```

This will start Memgraph with the Bolt protocol enabled.

#### Checking Status

To check the status of Memgraph:

```bash
./scripts/run_memgraph_locally.sh status
```

#### Stopping Memgraph

To stop Memgraph:

```bash
./scripts/run_memgraph_locally.sh stop
```

#### Restarting Memgraph

To restart Memgraph:

```bash
./scripts/run_memgraph_locally.sh restart
```

### Accessing Memgraph

Once started, you can access:

- Memgraph Bolt protocol: localhost:7687
- Memgraph HTTP API: http://localhost:7444
- Memgraph Lab UI: http://localhost:3000

## Configuring MCP-Graph to Use Local Memgraph

To configure MCP-Graph to use your locally running Memgraph instance, you need to set the Neo4j connection parameters in your configuration.

### Using Environment Variables

```bash
export MCPGRAPH_NEO4J_URI=bolt://localhost:7687
export MCPGRAPH_NEO4J_USERNAME=
export MCPGRAPH_NEO4J_PASSWORD=
```

### Using Config File

The application will automatically create a `config.yaml` file with default values if it doesn't exist. You can also create or modify it manually:

```yaml
neo4j:
  uri: bolt://localhost:7687
  username: ""
  password: ""
```

The default configuration includes settings for the application name, API port, Neo4j connection parameters, MCP server, and shutdown timeout.

## Running MCP-Graph

After starting Memgraph locally and configuring MCP-Graph, you can run the application:

### Option 1: Using the Convenience Script

We've provided a convenience script that will check if Memgraph is running, start it if needed, and then run MCP-Graph with the correct configuration:

```bash
./scripts/run_with_local_memgraph.sh
```

This script:
1. Checks if Memgraph is running, and starts it if not
2. Sets the necessary environment variables
3. Runs the MCP-Graph application

### Option 2: Manual Setup

If you prefer to set things up manually:

```bash
# Set the Neo4j connection environment variables
export MCPGRAPH_NEO4J_URI=bolt://localhost:7687
export MCPGRAPH_NEO4J_USERNAME=
export MCPGRAPH_NEO4J_PASSWORD=

# Run the application
go run cmd/server/main.go
```

## Data Persistence

The script creates a `memgraph_data` directory in the current working directory to store Memgraph data. This ensures your data persists between restarts.

## Troubleshooting

### Port Conflicts

The script automatically checks for port conflicts before starting Memgraph and will provide helpful error messages if any ports are already in use.

Memgraph requires the following ports to be available:
- 7687: Memgraph Bolt protocol
- 7444: Memgraph HTTP API
- 3000: Memgraph Lab UI

If you encounter port conflicts, you can use the following command to see what processes are using the ports:
```bash
lsof -i :7687,7444,3000
```

And to stop any existing Memgraph processes:
```bash
./scripts/run_memgraph_locally.sh stop
```

### Process Management

The script manages processes using PID files stored in the `memgraph_data` directory. If you encounter issues with processes not being properly tracked, you can manually check for and kill Memgraph processes:

```bash
ps aux | grep memgraph
kill <PID>
```

### Logs

Memgraph logs are output to the console. If you want to save logs, you can redirect the output when starting:

```bash
./scripts/run_memgraph_locally.sh start > memgraph_logs.txt 2>&1
```
