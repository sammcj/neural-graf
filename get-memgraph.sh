#!/usr/bin/env bash
#
#                  Memgraph Installer Script
#
#   Homepage: https://memgraph.com
#   Requires: bash, curl
#
# Hello! This is a script that installs Memgraph
# into your system (which may require password authorization).
# Use it like this:
#
#	$ curl https://raw.githubusercontent.com/sammcj/mcp-graph/main/get-memgraph.sh | bash
#
# This should work on Mac and Linux systems.

set -eE
set -o functrace

DIM='\033[2m'
BOLD='\033[1m'
RED='\033[91;1m'
GREEN='\033[32;1m'
RESET='\033[0m'

sudo_cmd=""
platform="$(uname | tr '[:upper:]' '[:lower:]')"

print_instruction() {
  printf '%b\n' "$BOLD$1$RESET"
}

print_step() {
  printf '%b\n' "$DIM$1$RESET"
}

print_error() {
  printf '%b\n' "$RED$1$RESET"
}

print_good() {
  printf '%b\n' "$GREEN$1$RESET"
}

install_memgraph() {
  printf "%b" "$BOLD"
  cat << "EOF"
  __  __                                         _
 |  \/  |                                       | |
 | \  / | ___ _ __ ___   __ _ _ __ __ _ _ __ | |__
 | |\/| |/ _ \ '_ ` _ \ / _` | '__/ _` | '_ \| '_ \
 | |  | |  __/ | | | | | (_| | | | (_| | |_) | | | |
 |_|  |_|\___|_| |_| |_|\__, |_|  \__,_| .__/|_| |_|
                         __/ |         | |
                        |___/          |_|

EOF
  printf "%b" "$RESET"

  # Check curl is installed
  if ! hash curl 2>/dev/null; then
    print_error "Could not find curl. Please install curl and try again."
    exit 1
  fi

  # Check if Memgraph is already installed
  if hash memgraph 2>/dev/null; then
    print_good "Memgraph is already installed."
    memgraph --version
    exit 0
  fi

  # Install Memgraph based on platform
  case "$platform" in
    darwin)
      install_memgraph_macos
      ;;
    linux)
      install_memgraph_linux
      ;;
    *)
      print_error "Unsupported platform: $platform"
      print_error "Please install Memgraph manually from https://memgraph.com/download"
      exit 1
      ;;
  esac
}

install_memgraph_macos() {
  print_step "Installing Memgraph on macOS using Homebrew..."

  # Check if Homebrew is installed
  if ! hash brew 2>/dev/null; then
    print_error "Homebrew is not installed. Please install Homebrew first:"
    print_error "/bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
    exit 1
  fi

  # Install Memgraph
  print_step "Adding Memgraph tap..."
  brew tap memgraph/memgraph

  print_step "Installing Memgraph..."
  brew install memgraph

  # Check installation
  if hash memgraph 2>/dev/null; then
    print_good "Memgraph has been installed successfully."
    memgraph --version
  else
    print_error "Installation failed. Please try again or install manually."
    exit 1
  fi
}

install_memgraph_linux() {
  print_step "Installing Memgraph on Linux..."

  # Check sudo permissions
  if hash sudo 2>/dev/null; then
    sudo_cmd="sudo"
    echo "Requires sudo permission to install Memgraph."
    if ! $sudo_cmd -v; then
      print_error "Need sudo privileges to complete installation."
      exit 1
    fi
  fi

  # Detect distribution
  if [ -f /etc/os-release ]; then
    . /etc/os-release
    dist_id=$ID
  else
    print_error "Could not detect Linux distribution."
    print_error "Please install Memgraph manually from https://memgraph.com/download"
    exit 1
  fi

  case "$dist_id" in
    ubuntu|debian)
      install_memgraph_debian
      ;;
    centos|rhel|fedora)
      install_memgraph_rpm
      ;;
    *)
      print_error "Unsupported Linux distribution: $dist_id"
      print_error "Please install Memgraph manually from https://memgraph.com/download"
      exit 1
      ;;
  esac
}

install_memgraph_debian() {
  print_step "Installing Memgraph on Debian/Ubuntu..."

  # Add Memgraph repository
  print_step "Adding Memgraph repository..."
  $sudo_cmd wget -O- https://packages.memgraph.com/memgraph.public | $sudo_cmd apt-key add -
  $sudo_cmd add-apt-repository "deb https://packages.memgraph.com/$(lsb_release -cs) latest main"

  # Update and install
  print_step "Updating package lists..."
  $sudo_cmd apt-get update

  print_step "Installing Memgraph..."
  $sudo_cmd apt-get install -y memgraph

  # Check installation
  if hash memgraph 2>/dev/null; then
    print_good "Memgraph has been installed successfully."
    memgraph --version
  else
    print_error "Installation failed. Please try again or install manually."
    exit 1
  fi
}

install_memgraph_rpm() {
  print_step "Installing Memgraph on CentOS/RHEL/Fedora..."

  # Add Memgraph repository
  print_step "Adding Memgraph repository..."
  $sudo_cmd rpm --import https://packages.memgraph.com/memgraph.public

  $sudo_cmd tee /etc/yum.repos.d/memgraph.repo << EOF > /dev/null
[memgraph]
name=Memgraph RPM repository
baseurl=https://packages.memgraph.com/rpm
gpgcheck=1
enabled=1
EOF

  # Install Memgraph
  print_step "Installing Memgraph..."
  $sudo_cmd yum install -y memgraph

  # Check installation
  if hash memgraph 2>/dev/null; then
    print_good "Memgraph has been installed successfully."
    memgraph --version
  else
    print_error "Installation failed. Please try again or install manually."
    exit 1
  fi
}

function exit_error {
  if [ "$?" -ne 0 ]; then
    print_error "There was some problem while installing Memgraph. Please visit https://memgraph.com/docs/memgraph/installation for manual installation instructions."
  fi
}

trap 'exit_error' EXIT

install_memgraph
print_instruction "Please visit https://memgraph.com/docs/memgraph/quick-start for further instructions on usage."
