# TermPOS - Terminal Point of Sale System

TermPOS is a minimal, terminal-based Point of Sale (POS) system written in Go. It provides three operating modes:

1. **Classic CLI Mode** — Command-based interface (e.g., `add`, `sell`, `report`)
2. **Agent Mode** — HTTP server for remote commands
3. **AI Assistant Mode** — Natural language interface (e.g., "add 3 lattes at $5")

## Features

- Product management (add, update stock)
- Sales recording
- Comprehensive reporting (sales, inventory, revenue, daily, top products, summary)
- SQLite storage for data persistence
- Clean terminal-formatted tables and outputs
- Report generation in multiple formats

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
```

## Usage

### Classic CLI Mode

```bash
# Add a product
./termpos add "Coffee" 3.50 10

# List inventory
./termpos inventory

# Sell products
./termpos sell 1 2

# Generate reports
./termpos report sales      # List all sales transactions
./termpos report inventory  # Show current inventory with values
./termpos report revenue    # Show revenue by product
./termpos report summary    # Show total revenue and items sold
./termpos report top        # Show top-selling products by quantity
./termpos report daily      # Show sales for today grouped by product
```

### Agent Mode (HTTP Server)

```bash
# Start the agent server
./termpos agent

# Use HTTP to interact with the POS system 
# (Examples using curl)
curl -X GET http://localhost:8000/products
curl -X POST http://localhost:8000/products -d '{"name":"Espresso","price":4.50,"stock":20}'
curl -X GET http://localhost:8000/reports/summary
curl -X GET http://localhost:8000/reports/daily
curl -X GET http://localhost:8000/reports/top?limit=10
```

### AI Assistant Mode

```bash
# Start the assistant mode
./termpos assistant

# Then use natural language commands like:
> add 5 muffins at $2.50
> sell 2 coffee
> show inventory
> what are my top selling products?
> show sales report
> show daily sales
```
