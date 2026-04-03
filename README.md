# pbd-cli

CLI tool for PaleBlueDot AI platform.

## Installation

### Option 1: One-click Install (Recommended)

**Linux / macOS:**

```bash
curl -sSL https://raw.githubusercontent.com/PaleBlueDot-AI-Open/pbd-cli/main/install.sh | bash
```

This script automatically detects your OS and architecture, downloads the latest release, and installs to `/usr/local/bin`.

### Option 2: Download Binary

Download from [GitHub Releases](https://github.com/PaleBlueDot-AI-Open/pbd-cli/releases):

| Platform | Architecture | Download |
|----------|-------------|----------|
| macOS (Intel) | amd64 | `pbd-cli_*_darwin_amd64.tar.gz` |
| macOS (Apple Silicon) | arm64 | `pbd-cli_*_darwin_arm64.tar.gz` |
| Linux | amd64 | `pbd-cli_*_linux_amd64.tar.gz` |
| Linux | arm64 | `pbd-cli_*_linux_arm64.tar.gz` |
| Windows | amd64 | `pbd-cli_*_windows_amd64.zip` |
| Windows | arm64 | `pbd-cli_*_windows_arm64.zip` |

Extract and place the binary in your PATH.

### Option 3: Build from Source

Requirements:
- Go 1.23+

```bash
# Clone the repository
git clone https://github.com/PaleBlueDot-AI-Open/pbd-cli.git
cd pbd-cli

# Install dependencies
go mod download

# Build
make build

# Install to /usr/local/bin
make install
```

## Usage

```bash
pbd-cli login                    # Authenticate with browser (default)
pbd-cli login --manual           # Authenticate with session cookie
pbd-cli token list [-f]          # List tokens (-f for formatted)
pbd-cli token create --name KEY   # Create new token
pbd-cli token delete <id>         # Delete token
pbd-cli token get-key <id>       # Get token key
pbd-cli usage balance [-f]       # Show balance
pbd-cli usage logs               # Show usage logs
pbd-cli wallet [-f]              # Show wallet balance
pbd-cli models list [-f]         # List available models
pbd-cli logout                   # Clear session
```

## Development

```bash
make build      # Build binary
make clean      # Remove binaries
make install    # Install to /usr/local/bin
```

## License

MIT