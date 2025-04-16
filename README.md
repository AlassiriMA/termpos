# TermPOS - Terminal Point of Sale System

TermPOS is a minimal, terminal-based Point of Sale (POS) system written in Go. It provides three operating modes:

1. **Classic CLI Mode** — Command-based interface (e.g., `add`, `sell`, `report`)
2. **Agent Mode** — HTTP server for remote commands
3. **AI Assistant Mode** — Natural language interface (e.g., "add 3 lattes at $5")

## Features

- Product management (add, update stock)
- Sales recording
- Reporting (sales, inventory, revenue)
- SQLite storage for data persistence
- Clean terminal-formatted tables and outputs

## Installation

### Prerequisites

- Go 1.21 or higher
- SQLite

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/termpos.git
cd termpos

# Build the application
go build -o termpos ./cmd/pos
