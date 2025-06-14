#!/bin/bash
# Script to run MCP-Graph with a local DGraph instance
# Useful for macOS users who use Colima and have issues with Docker Compose

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored messages
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if DGraph is running
check_dgraph_running() {
    print_info "Checking if DGraph is running..."

    # Check if DGraph Alpha is accessible
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health >/dev/null 2>&1; then
        print_info "DGraph Alpha is running and accessible."
        return 0
    else
        print_warning "DGraph Alpha is not accessible. Starting DGraph..."
        return 1
    fi
}

# Main function
main() {
    # Check if the DGraph script exists
    if [ ! -f "./scripts/run_dgraph_locally.sh" ]; then
        print_error "DGraph script not found. Make sure you're running this script from the project root directory."
        exit 1
    fi

    # Check if DGraph is running, start it if not
    if ! check_dgraph_running; then
        # Use the fix script to ensure proper DGraph setup
        print_info "Using fix script to ensure proper DGraph setup..."
        ./scripts/fix_dgraph_setup.sh
    fi

    # Set environment variables for MCP-Graph
    export MCPGRAPH_DGRAPH_ADDRESS=localhost:9080
    export MCPGRAPH_MCP_USESSE=true
    export MCPGRAPH_MCP_ADDRESS=:3000

    # Run MCP-Graph
    print_info "Starting MCP-Graph..."
    go run cmd/server/main.go
}

main "$@"
