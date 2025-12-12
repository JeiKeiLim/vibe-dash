# vibe-dash

A terminal dashboard for tracking AI-assisted coding projects.

## Features

- Track multiple AI coding projects from a single dashboard
- Detect when AI agents are waiting for user input
- Support for multiple AI coding methodologies (Speckit, BMAD, etc.)
- Centralized configuration and state management

## Requirements

- Go 1.21 or later
- CGO enabled (required for SQLite)

## Installation

```bash
# Clone the repository
git clone https://github.com/JeiKeiLim/vibe-dash.git
cd vibe-dash

# Build the binary
make build

# Or install globally
make install
```

## Usage

```bash
# Run the dashboard
vibe

# Or run directly after building
./bin/vibe
```

## Development

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Run all tests including integration tests
make test-all

# Build
make build

# Clean build artifacts
make clean
```

## Architecture

vibe-dash follows a hexagonal architecture pattern:

```
internal/
├── core/              # Domain layer - ZERO external dependencies
│   ├── domain/        # Entities
│   ├── ports/         # Interfaces only
│   └── services/      # Use cases
└── adapters/          # Infrastructure layer
    ├── cli/           # Cobra commands
    ├── tui/           # Bubble Tea components
    ├── persistence/   # SQLite + YAML
    ├── filesystem/    # OS abstraction
    └── detectors/     # MethodDetector implementations
```

## License

MIT License - see [LICENSE](LICENSE) for details.
