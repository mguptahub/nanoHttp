# justHttp

justHttp is a lightweight HTTP server manager that allows you to run multiple HTTP server instances on different ports. It's designed to be simple to use while providing essential features for serving static files.

## Features

- Run multiple HTTP server instances on different ports
- Configuration management at `~/.justHttp/config`
- Customizable instance settings:
  - Port number (default: 8080)
  - Web root folder
  - Directory listing (optional)
  - SSL certificates (optional)
- Instance management (add, delete, start, stop)
- Self-update capability
- Simple CLI interface with comprehensive help system

## Installation

### Quick Install (Recommended)

```bash
curl -fsSL https://justhttp.mguptahub.com/install.sh | bash
```

This will automatically detect your OS and architecture, download the appropriate binary, and install it to `/usr/local/bin`.

### Manual Installation

If you prefer to install manually, you can download the binary directly:

#### Linux

```bash
# For AMD64
curl -L -o justHttp https://github.com/mguptahub/justHttp/releases/latest/download/justHttp-linux-amd64

# For ARM64
curl -L -o justHttp https://github.com/mguptahub/justHttp/releases/latest/download/justHttp-linux-arm64

# Make it executable
chmod +x justHttp

# Move to a directory in your PATH
sudo mv justHttp /usr/local/bin/
```

#### macOS

```bash
# For Intel (AMD64)
curl -L -o justHttp https://github.com/mguptahub/justHttp/releases/latest/download/justHttp-darwin-amd64

# For Apple Silicon (ARM64)
curl -L -o justHttp https://github.com/mguptahub/justHttp/releases/latest/download/justHttp-darwin-arm64

# Make it executable
chmod +x justHttp

# Move to a directory in your PATH
sudo mv justHttp /usr/local/bin/
```

## Usage

justHttp provides a simple command-line interface. Use `--help` with any command to see detailed usage information:

```bash
# Show general help
justHttp --help

# Show help for specific commands
justHttp add --help
justHttp start --help
justHttp stop --help
justHttp delete --help
justHttp list --help
justHttp update --help
justHttp version --help
```

### Adding a new instance

```bash
# Basic usage with required flags
justHttp add -name myserver -web-folder /path/to/files

# Full usage with all options
justHttp add -name myserver \
  -port 8080 \
  -web-folder /path/to/files \
  -allow-dir-listing \
  -ssl-cert-folder /path/to/certs
```

### Managing instances

```bash
# Start an instance
justHttp start myserver

# Stop an instance
justHttp stop myserver

# Delete an instance
justHttp delete myserver

# List all instances
justHttp list
```

### System commands

```bash
# Check for updates
justHttp update

# Show version
justHttp version
```

## Configuration

The configuration file is stored at `~/.justHttp/config` in JSON format. Here's an example configuration:

```json
{
  "instances": {
    "myserver": {
      "name": "myserver",
      "port": 8080,
      "web_folder": "/path/to/files",
      "allow_dir_listing": true,
      "ssl_cert_folder": "",
      "is_running": false
    }
  }
}
```

## Building from source

```bash
# Clone the repository
git clone https://github.com/mguptahub/justHttp.git
cd justHttp

# Build the binary
go build -o justHttp cmd/justhttp/main.go
```

## Command-line Options

### Add Command
- `-name` (required): Instance name
- `-port` (default: 8080): Port number
- `-web-folder` (required): Web root folder
- `-allow-dir-listing` (default: false): Allow directory listing
- `-ssl-cert-folder`: SSL certificates folder

### Other Commands
- `start <instance-name>`: Start an instance
- `stop <instance-name>`: Stop an instance
- `delete <instance-name>`: Delete an instance
- `list`: List all instances
- `update`: Check for and install updates
- `version`: Show version information

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 