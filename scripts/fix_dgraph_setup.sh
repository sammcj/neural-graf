#!/bin/bash
# Script to fix DGraph setup issues
# This script ensures proper cleanup and initialization of DGraph

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

# Stop any running DGraph processes
stop_dgraph() {
    print_info "Stopping any running DGraph processes..."
    pkill -f dgraph || true
    sleep 2
}

# Clean up DGraph data
clean_dgraph_data() {
    print_info "Cleaning up DGraph data..."
    rm -rf dgraph_data
    mkdir -p dgraph_data/zero
    mkdir -p dgraph_data/alpha
    mkdir -p dgraph_data/alpha2
    mkdir -p dgraph_data/ratel
}

# Start DGraph Zero with proper configuration
start_dgraph_zero() {
    print_info "Starting DGraph Zero with proper configuration..."
    dgraph zero --my=127.0.0.1:5080 --wal=dgraph_data/zero/wal --bindall=true &
    ZERO_PID=$!
    echo $ZERO_PID >dgraph_data/zero.pid
    print_info "DGraph Zero started with PID: $ZERO_PID"

    # Wait for Zero to start and initialize
    print_info "Waiting for DGraph Zero to initialize..."
    sleep 20

    # Check if Zero is running
    if ! ps -p $ZERO_PID >/dev/null; then
        print_error "DGraph Zero failed to start. Check the logs for errors."
        exit 1
    fi

    print_info "DGraph Zero initialized"
}

# Start DGraph Alpha with proper configuration
start_dgraph_alpha() {
    print_info "Starting first DGraph Alpha node with proper configuration..."
    dgraph alpha --my=127.0.0.1:7080 --zero=127.0.0.1:5080 --postings=dgraph_data/alpha/p --wal=dgraph_data/alpha/wal --security whitelist=0.0.0.0/0 --bindall=true &
    ALPHA1_PID=$!
    echo $ALPHA1_PID >dgraph_data/alpha1.pid
    print_info "First DGraph Alpha started with PID: $ALPHA1_PID"

    # Wait for first Alpha to start initializing
    sleep 5

    print_info "Starting second DGraph Alpha node with proper configuration..."
    dgraph alpha --my=127.0.0.1:7081 --zero=127.0.0.1:5080 --postings=dgraph_data/alpha2/p --wal=dgraph_data/alpha2/wal --security whitelist=0.0.0.0/0 --bindall=true -o=1 --port_offset=1 &
    ALPHA2_PID=$!
    echo $ALPHA2_PID >dgraph_data/alpha2.pid
    print_info "Second DGraph Alpha started with PID: $ALPHA2_PID"

    # Wait for Alpha nodes to initialize
    print_info "Waiting for DGraph Alpha nodes to initialize..."
    sleep 30

    # Check if Alpha nodes are running
    if ! ps -p $ALPHA1_PID >/dev/null; then
        print_error "First DGraph Alpha failed to start. Check the logs for errors."
        exit 1
    fi

    if ! ps -p $ALPHA2_PID >/dev/null; then
        print_error "Second DGraph Alpha failed to start. Check the logs for errors."
        exit 1
    fi

    # Check if Alpha is accessible
    print_info "Checking if DGraph Alpha is accessible..."
    MAX_RETRIES=10
    RETRY_COUNT=0
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health | grep -q "200"; then
            print_info "DGraph Alpha is accessible"
            break
        else
            RETRY_COUNT=$((RETRY_COUNT + 1))
            if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
                print_warning "DGraph Alpha is not responding, but processes are running. Continuing anyway..."
            else
                print_info "Waiting for DGraph Alpha to become accessible (attempt $RETRY_COUNT/$MAX_RETRIES)..."
                sleep 5
            fi
        fi
    done
}

# Initialize DGraph schema
initialize_schema() {
    print_info "Initializing DGraph schema..."

    # Create a temporary schema file
    cat >/tmp/schema.graphql <<EOF
type Person {
    id: ID!
    name: String!
    age: Int
}
EOF

    # Wait for a moment to ensure DGraph is ready
    sleep 5

    # Upload the schema to DGraph
    print_info "Uploading schema to DGraph..."
    if curl -s -X POST localhost:8080/admin/schema --data-binary '@/tmp/schema.graphql' | grep -q "success"; then
        print_info "Schema uploaded successfully"
    else
        print_warning "Schema upload may have failed, but continuing anyway..."
    fi

    # Remove the temporary schema file
    rm /tmp/schema.graphql
}

# Start DGraph Ratel (UI)
start_dgraph_ratel() {
    print_info "Starting DGraph Ratel (UI)..."
    if command -v dgraph-ratel &>/dev/null; then
        dgraph-ratel &
        RATEL_PID=$!
        echo $RATEL_PID >dgraph_data/ratel.pid
        print_info "DGraph Ratel started with PID: $RATEL_PID"
        print_info "DGraph Ratel UI is available at: http://localhost:8000"
    else
        print_warning "DGraph Ratel not found. You can install it using 'curl -s https://get.dgraph.io | bash'"
        print_info "Alternatively, you can access the DGraph UI by downloading Ratel from https://github.com/dgraph-io/ratel/releases"
        print_info "Or use the DGraph Alpha HTTP API directly at http://localhost:8080"
    fi
}

# Main function
main() {
    print_info "Starting DGraph fix process..."

    # Stop any running DGraph processes
    stop_dgraph

    # Clean up DGraph data
    clean_dgraph_data

    # Start DGraph Zero
    start_dgraph_zero

    # Start DGraph Alpha
    start_dgraph_alpha

    # Initialize schema
    initialize_schema

    # Start DGraph Ratel (UI)
    start_dgraph_ratel

    print_info "DGraph setup completed. You can now run your application."
    print_info "DGraph Alpha HTTP API: http://localhost:8080"
    print_info "DGraph Alpha gRPC API: localhost:9080"
}

main "$@"
