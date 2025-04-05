#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print with color
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Print error and exit
error_exit() {
    print_color "$RED" "Error: $1"
    exit 1
}

# Print success message
success() {
    print_color "$GREEN" "$1"
}

# Print info message
info() {
    print_color "$YELLOW" "$1"
}

# Detect OS and architecture
detect_os_arch() {
    local os=""
    local arch=""
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        *)          error_exit "Unsupported operating system";;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64)     arch="amd64";;
        aarch64)    arch="arm64";;
        arm64)      arch="arm64";;
        *)          error_exit "Unsupported architecture";;
    esac
    
    echo "${os}-${arch}"
}

# Download and install binary
install_binary() {
    local os_arch=$1
    local version="latest"
    local binary_name="nanoHttp-${os_arch}"
    local install_dir="/usr/local/bin"
    local binary_path="${install_dir}/nanoHttp"
    
    info "Downloading nanoHttp for ${os_arch}..."
    
    # Create a temporary directory
    local tmp_dir=$(mktemp -d)
    trap 'rm -rf "$tmp_dir"' EXIT
    
    # Download the binary
    if ! curl -L -o "${tmp_dir}/${binary_name}" "https://github.com/mguptahub/nanoHttp/releases/${version}/download/${binary_name}"; then
        error_exit "Failed to download binary"
    fi
    
    # Make the binary executable
    chmod +x "${tmp_dir}/${binary_name}"
    
    # Check if we have write permission to /usr/local/bin
    if [ ! -w "$install_dir" ]; then
        info "Need sudo permission to install to ${install_dir}"
        if ! sudo mv "${tmp_dir}/${binary_name}" "$binary_path"; then
            error_exit "Failed to install binary"
        fi
    else
        if ! mv "${tmp_dir}/${binary_name}" "$binary_path"; then
            error_exit "Failed to install binary"
        fi
    fi
    
    success "nanoHttp has been installed to ${binary_path}"
    success "You can now run 'nanoHttp --help' to get started"
}

# Main installation process
main() {
    info "Installing nanoHttp..."
    
    # Detect OS and architecture
    local os_arch=$(detect_os_arch)
    
    # Install binary
    install_binary "$os_arch"
}

# Run main function
main 