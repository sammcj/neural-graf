# Running MCP-Graph Without Docker Compose

This guide provides instructions for running MCP-Graph without Docker Compose, which is particularly useful for macOS users who use Colima and experience issues with Docker Compose.

## Prerequisites

- Go 1.21 or higher
- DGraph installed locally (see instructions below)

## Installing DGraph Locally

You can install DGraph directly on your system using the official installer:

```bash
curl -s https://get.dgraph.io | bash
```

This will install the DGraph binaries (`dgraph` and `dgraph-ratel`) on your system.

## Using the Local DGraph Script

We've provided a script to help you run DGraph locally without Docker Compose. The script manages the DGraph Zero, Alpha, and Ratel components.

### Script Location

The script is located at `scripts/run_dgraph_locally.sh` in the project directory.

### Script Usage

Make the script executable:

```bash
chmod +x scripts/run_dgraph_locally.sh
```

#### Starting DGraph

To start all DGraph components:

```bash
./scripts/run_dgraph_locally.sh start
```

This will:
1. Start DGraph Zero (cluster manager)
2. Start DGraph Alpha (database server)
3. Start DGraph Ratel (UI)

#### Checking Status

To check the status of DGraph components:

```bash
./scripts/run_dgraph_locally.sh status
```

#### Stopping DGraph

To stop all DGraph components:

```bash
./scripts/run_dgraph_locally.sh stop
```

#### Restarting DGraph

To restart all DGraph components:

```bash
./scripts/run_dgraph_locally.sh restart
```

### Accessing DGraph

Once started, you can access:

- DGraph Alpha HTTP API: http://localhost:8080
- DGraph Alpha gRPC API: localhost:9080
- DGraph Ratel UI: http://localhost:8000

## Configuring MCP-Graph to Use Local DGraph

To configure MCP-Graph to use your locally running DGraph instance, you need to set the DGraph address in your configuration.

### Using Environment Variables

```bash
export MCPGRAPH_DGRAPH_ADDRESS=localhost:9080
```

### Using Config File

Create or modify your `config.yaml` file:

```yaml
dgraph:
  address: localhost:9080
```

## Running MCP-Graph

After starting DGraph locally and configuring MCP-Graph, you can run the application:

### Option 1: Using the Convenience Script

We've provided a convenience script that will check if DGraph is running, start it if needed, and then run MCP-Graph with the correct configuration:

```bash
./scripts/run_with_local_dgraph.sh
```

This script:
1. Checks if DGraph is running, and starts it if not
2. Sets the necessary environment variables
3. Runs the MCP-Graph application

### Option 2: Manual Setup

If you prefer to set things up manually:

```bash
# Set the DGraph address environment variable
export MCPGRAPH_DGRAPH_ADDRESS=localhost:9080

# Run the application
go run cmd/server/main.go
```

## Data Persistence

The script creates a `dgraph_data` directory in the current working directory to store DGraph data. This ensures your data persists between restarts.

## Troubleshooting

### Port Conflicts

If you encounter port conflicts, ensure that no other services are using the following ports:
- 5080: DGraph Zero internal port
- 6080: DGraph Zero external port
- 7080: DGraph Alpha internal port
- 8080: DGraph Alpha HTTP API
- 9080: DGraph Alpha gRPC API
- 8000: DGraph Ratel UI

### Process Management

The script manages processes using PID files stored in the `dgraph_data` directory. If you encounter issues with processes not being properly tracked, you can manually check for and kill DGraph processes:

```bash
ps aux | grep dgraph
kill <PID>
```

### Logs

DGraph logs are output to the console. If you want to save logs, you can redirect the output when starting:

```bash
./scripts/run_dgraph_locally.sh start > dgraph_logs.txt 2>&1
```
