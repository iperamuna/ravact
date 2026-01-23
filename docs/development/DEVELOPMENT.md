# Ravact Development Guide

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Make
- Docker (optional, for testing)

### Setup

```bash
# Clone the repository
git clone https://github.com/iperamuna/ravact.git
cd ravact/ravact-go

# Download dependencies
go mod download

# Build the project
make build

# Run tests
make test
```

## Project Structure

```
ravact-go/
├── cmd/ravact/              # Application entry point
├── internal/
│   ├── models/              # Data structures and types
│   ├── system/              # System detection and utilities
│   ├── setup/               # Setup script execution
│   ├── config/              # Configuration management
│   └── ui/                  # TUI components
│       ├── screens/         # Screen implementations
│       ├── components/      # Reusable UI components (future)
│       └── theme/           # Styling and themes
├── assets/
│   ├── scripts/             # Setup scripts (.sh files)
│   └── configs/             # Configuration templates (.json files)
├── tests/                   # Integration tests
└── scripts/                 # Development scripts
```

## Development Workflow

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build for specific platform
make build-linux        # Linux x64
make build-linux-arm64  # Linux ARM64
make build-darwin       # macOS x64
make build-darwin-arm64 # macOS ARM64
```

### Testing

```bash
# Run unit tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration

# Run tests in Docker (Ubuntu 24.04)
make docker-test

# Or use the test script
./scripts/test.sh
```

### Running

```bash
# Run the application
./ravact

# Or directly with go run
go run ./cmd/ravact
```

### Docker Testing

Test in Ubuntu 24.04 environment:

```bash
# Build and run tests in Docker
make docker-test

# Or manually
docker build -t ravact-test -f Dockerfile.test .
docker run --rm ravact-test

# Open interactive shell for manual testing
make docker-shell
```

## Adding New Features

### Adding a New Setup Script

1. Create a shell script in `assets/scripts/`:

```bash
#!/bin/bash
# assets/scripts/myservice.sh

echo "Installing MyService..."
# Installation commands here
```

2. Make it executable:

```bash
chmod +x assets/scripts/myservice.sh
```

3. The script will automatically appear in the Setup menu

### Adding a New Configuration Template

1. Create a JSON template in `assets/configs/`:

```json
{
  "id": "myservice-config",
  "service_id": "myservice",
  "name": "MyService Configuration",
  "description": "Configure MyService settings",
  "file_path": "/etc/myservice/config.conf",
  "fields": [
    {
      "key": "port",
      "label": "Port",
      "type": "int",
      "required": true,
      "default": 8080,
      "description": "Port to listen on"
    }
  ]
}
```

### Adding a New Quick Command

Edit `internal/ui/screens/quick_commands.go` and add to the `commands` slice:

```go
{
    ID:          "my-command",
    Name:        "My Command",
    Description: "Description of what it does",
    Command:     "mycommand",
    Args:        []string{"arg1", "arg2"},
    RequireRoot: false,
    Confirm:     false,
}
```

## Code Style

- Follow Go conventions and idiomatic patterns
- Run `gofmt` before committing
- Add tests for new functionality
- Document exported functions and types

### Running Linters

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linters
golangci-lint run
```

## Testing Guidelines

### Unit Tests

- Test individual functions and methods
- Mock external dependencies
- Place tests in `*_test.go` files alongside the code
- Use table-driven tests where appropriate

Example:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case 1", "input1", "output1"},
        {"case 2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MyFunction(tt.input)
            if result != tt.expected {
                t.Errorf("expected %s, got %s", tt.expected, result)
            }
        })
    }
}
```

### Integration Tests

- Tag with `//go:build integration`
- Test interactions between components
- Can use temporary files and directories
- Place in `tests/` directory

Example:

```go
//go:build integration
// +build integration

package tests

import "testing"

func TestIntegration(t *testing.T) {
    // Integration test code
}
```

## Debugging

### Enable Debug Logging

Set environment variable:

```bash
export RAVACT_DEBUG=1
./ravact
```

### TUI Debugging

The TUI takes over the terminal, making debugging difficult. Options:

1. Run in a separate terminal with output redirection
2. Use a debugger like Delve
3. Write output to a log file

```go
// Log to file for debugging
f, _ := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
defer f.Close()
log.SetOutput(f)
log.Println("Debug message")
```

## Platform-Specific Development

### macOS (Development)

- Use for development and testing the TUI
- Cannot test actual Linux-specific features
- Use Docker for Linux testing

### Linux (Target Platform)

- Primary deployment target
- Test actual setup scripts and system commands
- Use Ubuntu 24.04 for testing (Docker or VM)

### Cross-Platform Considerations

- Use `runtime.GOOS` to detect platform
- Test Linux-specific code in Docker
- Ensure paths work on both platforms

## Release Process

1. Update version in code if needed
2. Run all tests: `make test && make test-integration`
3. Build all platforms: `make build-all`
4. Test binaries on target platforms
5. Create git tag: `git tag -a v0.2.0 -m "Release v0.2.0"`
6. Push tag: `git push origin v0.2.0`
7. GitHub Actions will create release automatically

## Troubleshooting

### Build Errors

```bash
# Clean and rebuild
make clean
make build
```

### Test Failures

```bash
# Run specific test
go test -v ./internal/system -run TestSystemInfo

# Run with verbose output
go test -v ./...
```

### Docker Issues

```bash
# Clean Docker images
docker rmi ravact-test

# Rebuild from scratch
docker build --no-cache -t ravact-test -f Dockerfile.test .
```

## Contributing

See main README.md for contribution guidelines.

## Resources

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [Go Testing](https://golang.org/pkg/testing/)
- [Effective Go](https://golang.org/doc/effective_go)
