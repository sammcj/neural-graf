#!/bin/bash
# Script to run Memgraph locally without Docker Compose
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

# Check if Memgraph is installed
check_memgraph_installed() {
  if ! command -v memgraph &>/dev/null; then
    print_error "Memgraph is not installed. Please install it first."
    echo "You can install Memgraph using the following command:"
    echo "For macOS: brew install memgraph/memgraph/memgraph"
    echo "For Linux: See https://memgraph.com/docs/memgraph/installation"
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

# Check if Memgraph ports are available
check_memgraph_ports() {
  local ports=(7687 7444 3000)
  local port_names=("Bolt protocol" "HTTP API" "Memgraph Lab UI")
  local has_conflict=false

  print_info "Checking if Memgraph ports are available..."

  for i in "${!ports[@]}"; do
    if check_port_in_use ${ports[$i]}; then
      print_error "Port ${ports[$i]} (${port_names[$i]}) is already in use."
      has_conflict=true
    fi
  done

  if $has_conflict; then
    print_error "Some Memgraph ports are already in use. Please stop any running Memgraph instances or other services using these ports."
    print_info "You can use the following command to check what's using the ports:"
    print_info "  lsof -i :7687,7444,3000"
    print_info "And to stop existing Memgraph processes:"
    print_info "  ./scripts/run_memgraph_locally.sh stop"
    exit 1
  fi

  print_info "All Memgraph ports are available."
}

# Create directories for Memgraph data
create_directories() {
  print_info "Creating directories for Memgraph data..."
  mkdir -p memgraph_data
}

# Start Memgraph
start_memgraph() {
  print_info "Starting Memgraph..."
  memgraph --data-directory=./memgraph_data --bolt-port=7687 --log-level=TRACE &
  MEMGRAPH_PID=$!
  echo $MEMGRAPH_PID >memgraph_data/memgraph.pid
  print_info "Memgraph started with PID: $MEMGRAPH_PID"

  # Wait for Memgraph to start and initialize
  print_info "Waiting for Memgraph to initialize..."
  sleep 10

  # Check if Memgraph is running
  if ! ps -p $MEMGRAPH_PID >/dev/null; then
    print_error "Memgraph failed to start. Check the logs for errors."
    exit 1
  fi

  print_info "Memgraph initialized"
}

# Stop Memgraph processes
stop_memgraph() {
  print_info "Stopping Memgraph processes..."

  if [ -f memgraph_data/memgraph.pid ]; then
    MEMGRAPH_PID=$(cat memgraph_data/memgraph.pid)
    if ps -p $MEMGRAPH_PID >/dev/null; then
      kill $MEMGRAPH_PID
      print_info "Memgraph stopped"
    else
      print_warning "Memgraph process not found"
    fi
    rm memgraph_data/memgraph.pid
  fi
}

# Show status of Memgraph processes
show_status() {
  print_info "Memgraph Status:"

  if [ -f memgraph_data/memgraph.pid ]; then
    MEMGRAPH_PID=$(cat memgraph_data/memgraph.pid)
    if ps -p $MEMGRAPH_PID >/dev/null; then
      print_info "Memgraph is running with PID: $MEMGRAPH_PID"
    else
      print_warning "Memgraph is not running (stale PID file)"
    fi
  else
    print_warning "Memgraph is not running"
  fi

  print_info "Memgraph Bolt protocol: localhost:7687"
  print_info "Memgraph HTTP API: http://localhost:7444"
  print_info "Memgraph Lab UI: http://localhost:3000"
}

# Main function
main() {
  case "$1" in
  start)
    check_memgraph_installed
    check_memgraph_ports
    create_directories
    start_memgraph
    show_status
    ;;
  stop)
    stop_memgraph
    ;;
  restart)
    stop_memgraph
    sleep 2
    check_memgraph_installed
    check_memgraph_ports
    create_directories
    start_memgraph
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
