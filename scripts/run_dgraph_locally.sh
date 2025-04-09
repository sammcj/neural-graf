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
    if ! command -v dgraph &> /dev/null; then
        print_error "DGraph is not installed. Please install it first."
        echo "You can install DGraph using the following command:"
        echo "curl -s https://get.dgraph.io | bash"
        exit 1
    fi
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
    dgraph zero --my=localhost:5080 --wal=dgraph_data/zero/wal --zero=localhost:5080 &
    ZERO_PID=$!
    echo $ZERO_PID > dgraph_data/zero.pid
    print_info "DGraph Zero started with PID: $ZERO_PID"

    # Wait for Zero to start
    sleep 5
}

# Start DGraph Alpha
start_dgraph_alpha() {
    print_info "Starting DGraph Alpha..."
    dgraph alpha --my=localhost:7080 --zero=localhost:5080 --postings=dgraph_data/alpha/p --wal=dgraph_data/alpha/wal --security whitelist=0.0.0.0/0 &
    ALPHA_PID=$!
    echo $ALPHA_PID > dgraph_data/alpha.pid
    print_info "DGraph Alpha started with PID: $ALPHA_PID"

    # Wait for Alpha to start
    sleep 5
}

# Start DGraph Ratel (UI)
start_dgraph_ratel() {
    print_info "Starting DGraph Ratel (UI)..."
    dgraph-ratel &
    RATEL_PID=$!
    echo $RATEL_PID > dgraph_data/ratel.pid
    print_info "DGraph Ratel started with PID: $RATEL_PID"
}

# Stop DGraph processes
stop_dgraph() {
    print_info "Stopping DGraph processes..."

    if [ -f dgraph_data/zero.pid ]; then
        ZERO_PID=$(cat dgraph_data/zero.pid)
        if ps -p $ZERO_PID > /dev/null; then
            kill $ZERO_PID
            print_info "DGraph Zero stopped"
        else
            print_warning "DGraph Zero process not found"
        fi
        rm dgraph_data/zero.pid
    fi

    if [ -f dgraph_data/alpha.pid ]; then
        ALPHA_PID=$(cat dgraph_data/alpha.pid)
        if ps -p $ALPHA_PID > /dev/null; then
            kill $ALPHA_PID
            print_info "DGraph Alpha stopped"
        else
            print_warning "DGraph Alpha process not found"
        fi
        rm dgraph_data/alpha.pid
    fi

    if [ -f dgraph_data/ratel.pid ]; then
        RATEL_PID=$(cat dgraph_data/ratel.pid)
        if ps -p $RATEL_PID > /dev/null; then
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
        if ps -p $ZERO_PID > /dev/null; then
            print_info "DGraph Zero is running with PID: $ZERO_PID"
        else
            print_warning "DGraph Zero is not running (stale PID file)"
        fi
    else
        print_warning "DGraph Zero is not running"
    fi

    if [ -f dgraph_data/alpha.pid ]; then
        ALPHA_PID=$(cat dgraph_data/alpha.pid)
        if ps -p $ALPHA_PID > /dev/null; then
            print_info "DGraph Alpha is running with PID: $ALPHA_PID"
        else
            print_warning "DGraph Alpha is not running (stale PID file)"
        fi
    else
        print_warning "DGraph Alpha is not running"
    fi

    if [ -f dgraph_data/ratel.pid ]; then
        RATEL_PID=$(cat dgraph_data/ratel.pid)
        if ps -p $RATEL_PID > /dev/null; then
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
