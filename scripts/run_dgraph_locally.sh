#!/bin/bash
# Script to run DGraph locally without Docker Compose
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

# Check if DGraph is installed
check_dgraph_installed() {
    if ! command -v dgraph &>/dev/null; then
        print_error "DGraph is not installed. Please install it first."
        echo "You can install DGraph using the following command:"
        echo "curl -s https://get.dgraph.io | bash"
        exit 1
    fi
}

# Check if a port is in use
check_port_in_use() {
    local port=$1
    if lsof -i :$port -sTCP:LISTEN -t >/dev/null; then
        return 0 # Port is in use
    else
        return 1 # Port is free
    fi
}

# Check if DGraph ports are available
check_dgraph_ports() {
    local ports=(5080 6080 7080 8080 9080)
    local port_names=("DGraph Zero internal" "DGraph Zero external" "DGraph Alpha internal" "DGraph Alpha HTTP API" "DGraph Alpha gRPC API")
    local has_conflict=false

    print_info "Checking if DGraph ports are available..."

    for i in "${!ports[@]}"; do
        if check_port_in_use ${ports[$i]}; then
            print_error "Port ${ports[$i]} (${port_names[$i]}) is already in use."
            has_conflict=true
        fi
    done

    if $has_conflict; then
        print_error "Some DGraph ports are already in use. Please stop any running DGraph instances or other services using these ports."
        print_info "You can use the following command to check what's using the ports:"
        print_info "  lsof -i :5080,6080,7080,8080,9080"
        print_info "And to stop existing DGraph processes:"
        print_info "  ./scripts/run_dgraph_locally.sh stop"
        exit 1
    fi

    print_info "All DGraph ports are available."
}

# Create directories for DGraph data
create_directories() {
    print_info "Creating directories for DGraph data..."
    mkdir -p dgraph_data/zero
    mkdir -p dgraph_data/alpha
    mkdir -p dgraph_data/ratel
}

# Start DGraph Zero
start_dgraph_zero() {
    print_info "Starting DGraph Zero..."
    dgraph zero --my=localhost:5080 --wal=dgraph_data/zero/wal --bindall=true &
    ZERO_PID=$!
    echo $ZERO_PID >dgraph_data/zero.pid
    print_info "DGraph Zero started with PID: $ZERO_PID"

    # Wait for Zero to start and initialize
    print_info "Waiting for DGraph Zero to initialize..."
    sleep 15

    # Check if Zero is running
    if ! ps -p $ZERO_PID >/dev/null; then
        print_error "DGraph Zero failed to start. Check the logs for errors."
        exit 1
    fi

    print_info "DGraph Zero initialized"
}

# Start DGraph Alpha
start_dgraph_alpha() {
    print_info "Starting DGraph Alpha..."
    dgraph alpha --my=localhost:7080 --zero=localhost:5080 --postings=dgraph_data/alpha/p --wal=dgraph_data/alpha/wal --security whitelist=0.0.0.0/0 --bindall=true &
    ALPHA_PID=$!
    echo $ALPHA_PID >dgraph_data/alpha.pid
    print_info "DGraph Alpha started with PID: $ALPHA_PID"

    # Wait for Alpha to start and initialize
    print_info "Waiting for DGraph Alpha to initialize..."
    sleep 20

    # Check if Alpha is running
    if ! ps -p $ALPHA_PID >/dev/null; then
        print_error "DGraph Alpha failed to start. Check the logs for errors."
        exit 1
    fi

    # Check if Alpha is accessible
    print_info "Checking if DGraph Alpha is accessible..."
    MAX_RETRIES=5
    RETRY_COUNT=0
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health | grep -q "200"; then
            print_info "DGraph Alpha is accessible"
            break
        else
            RETRY_COUNT=$((RETRY_COUNT + 1))
            if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
                print_warning "DGraph Alpha is not responding, but process is running. Continuing anyway..."
            else
                print_info "Waiting for DGraph Alpha to become accessible (attempt $RETRY_COUNT/$MAX_RETRIES)..."
                sleep 2
            fi
        fi
    done
}

# Start DGraph Ratel (UI)
start_dgraph_ratel() {
    print_info "Starting DGraph Ratel (UI)..."
    if command -v dgraph-ratel &>/dev/null; then
        dgraph-ratel &
        RATEL_PID=$!
        echo $RATEL_PID >dgraph_data/ratel.pid
        print_info "DGraph Ratel started with PID: $RATEL_PID"
    else
        print_warning "DGraph Ratel not found. Skipping Ratel UI startup."
        print_info "You can still access DGraph Alpha HTTP API at http://localhost:8080"
    fi
}

# Stop DGraph processes
stop_dgraph() {
    print_info "Stopping DGraph processes..."

    if [ -f dgraph_data/zero.pid ]; then
        ZERO_PID=$(cat dgraph_data/zero.pid)
        if ps -p $ZERO_PID >/dev/null; then
            kill $ZERO_PID
            print_info "DGraph Zero stopped"
        else
            print_warning "DGraph Zero process not found"
        fi
        rm dgraph_data/zero.pid
    fi

    if [ -f dgraph_data/alpha.pid ]; then
        ALPHA_PID=$(cat dgraph_data/alpha.pid)
        if ps -p $ALPHA_PID >/dev/null; then
            kill $ALPHA_PID
            print_info "DGraph Alpha stopped"
        else
            print_warning "DGraph Alpha process not found"
        fi
        rm dgraph_data/alpha.pid
    fi

    if [ -f dgraph_data/ratel.pid ]; then
        RATEL_PID=$(cat dgraph_data/ratel.pid)
        if ps -p $RATEL_PID >/dev/null; then
            kill $RATEL_PID
            print_info "DGraph Ratel stopped"
        else
            print_warning "DGraph Ratel process not found"
        fi
        rm dgraph_data/ratel.pid
    fi
}

# Show status of DGraph processes
show_status() {
    print_info "DGraph Status:"

    if [ -f dgraph_data/zero.pid ]; then
        ZERO_PID=$(cat dgraph_data/zero.pid)
        if ps -p $ZERO_PID >/dev/null; then
            print_info "DGraph Zero is running with PID: $ZERO_PID"
        else
            print_warning "DGraph Zero is not running (stale PID file)"
        fi
    else
        print_warning "DGraph Zero is not running"
    fi

    if [ -f dgraph_data/alpha.pid ]; then
        ALPHA_PID=$(cat dgraph_data/alpha.pid)
        if ps -p $ALPHA_PID >/dev/null; then
            print_info "DGraph Alpha is running with PID: $ALPHA_PID"
        else
            print_warning "DGraph Alpha is not running (stale PID file)"
        fi
    else
        print_warning "DGraph Alpha is not running"
    fi

    if [ -f dgraph_data/ratel.pid ]; then
        RATEL_PID=$(cat dgraph_data/ratel.pid)
        if ps -p $RATEL_PID >/dev/null; then
            print_info "DGraph Ratel is running with PID: $RATEL_PID"
        else
            print_warning "DGraph Ratel is not running (stale PID file)"
        fi
    else
        print_warning "DGraph Ratel is not running"
    fi

    print_info "DGraph Alpha HTTP API: http://localhost:8080"
    print_info "DGraph Alpha gRPC API: localhost:9080"
    print_info "DGraph Ratel UI: http://localhost:8000"
}

# Main function
main() {
    case "$1" in
    start)
        check_dgraph_installed
        check_dgraph_ports
        create_directories
        start_dgraph_zero
        start_dgraph_alpha
        start_dgraph_ratel
        show_status
        ;;
    stop)
        stop_dgraph
        ;;
    restart)
        stop_dgraph
        sleep 2
        check_dgraph_installed
        check_dgraph_ports
        create_directories
        start_dgraph_zero
        start_dgraph_alpha
        start_dgraph_ratel
        show_status
        ;;
    status)
        show_status
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status}"
        exit 1
        ;;
    esac
}

main "$@"
