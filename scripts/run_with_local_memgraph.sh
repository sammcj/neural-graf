#!/bin/bash
# Script to run MCP-Graph with a local Memgraph instance
# Useful for users who want to run Memgraph directly on their machine

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print coloured messages
print_info() {
  echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
  echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Memgraph is running
check_memgraph_running() {
  print_info "Checking if Memgraph is running..."

  # Check if Memgraph is accessible
  if nc -z localhost 7687 >/dev/null 2>&1; then
    print_info "Memgraph is running and accessible."
    return 0
  else
    print_warning "Memgraph is not accessible. Starting Memgraph..."
    return 1
  fi
}

# Main function
main() {
  # Check if the Memgraph script exists
  if [ ! -f "./scripts/run_memgraph_locally.sh" ]; then
    print_error "Memgraph script not found. Make sure you're running this script from the project root directory."
    exit 1
  fi

  # Check if Memgraph is running, start it if not
  if ! check_memgraph_running; then
    # Use the script to start Memgraph
    print_info "Using script to start Memgraph..."
    ./scripts/run_memgraph_locally.sh start
  fi

  # Set environment variables for MCP-Graph
  export MCPGRAPH_NEO4J_URI=bolt://localhost:7687
  export MCPGRAPH_NEO4J_USERNAME=
  export MCPGRAPH_NEO4J_PASSWORD=
  export MCPGRAPH_MCP_USESSE=true
  export MCPGRAPH_MCP_ADDRESS=:3000

  # Run MCP-Graph
  print_info "Starting MCP-Graph..."
  go run cmd/server/main.go
}

main "$@"
