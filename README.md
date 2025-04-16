# TermPOS - Terminal Point of Sale System

TermPOS is a minimal, terminal-based Point of Sale (POS) system written in Go. It provides three operating modes:

1. **Classic CLI Mode** — Command-based interface (e.g., `add`, `sell`, `report`)
2. **Agent Mode** — HTTP server for remote commands
3. **AI Assistant Mode** — Natural language interface (e.g., "add 3 lattes at $5")

## Features

- Product management (add, update stock)
- Sales recording with multiple payment methods (cash, card, mobile)
- Comprehensive reporting (sales, inventory, revenue, daily, top products, summary)
- Staff management with role-based access control
- Customer profiles with loyalty program
- Receipt generation for sales transactions
- Inventory tracking with low stock alerts
- Configurable business settings
- Automated database backups with encryption
- Audit logging for compliance
- SQLite storage for data persistence
- Docker support for containerized deployment
- Clean terminal-formatted tables and outputs
- Multiple operating modes for different use cases

## Installation

### Prerequisites

- Go 1.19 or higher
- SQLite

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/termpos.git
cd termpos

# Build the application
go build -o termpos ./cmd/pos
```

### Using Provided Makefile

```bash
# Build for all platforms (Linux, Windows, macOS)
make build

# Build for a specific platform
make build-linux
make build-windows
make build-macos

# Clean build artifacts
make clean

# Run tests
make test
```

### Using Docker

```bash
# Build the Docker image
docker build -t termpos:latest .

# Run in classic mode with mounted volumes for persistence
docker run -v ./data:/app/data -v ./config:/app/config termpos:latest

# Run in agent mode and expose the API port
docker run -p 8000:8000 -v ./data:/app/data -v ./config:/app/config termpos:latest --mode agent --port 8000

# Run with environment variables
docker run -e POS_ENCRYPTION_KEY="your-secure-key" -v ./data:/app/data termpos:latest
```

### Using Docker Compose

```bash
# Start the services defined in docker-compose.yml
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the services
docker-compose down
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

### Advanced Sales Features

```bash
# Sell with discount, tax rate, and printing receipt
./termpos sell 1 2 --discount 0.5 --tax-rate 8.5 --print-receipt

# Sell with customer information for loyalty program
./termpos sell 1 1 --customer-id 1 --print-receipt

# Sell with payment method details
./termpos sell 1 2 --payment-method "card" --payment-ref "TX123456" --email "customer@example.com"
```

### Staff Management

```bash
# Add a staff member
./termpos staff add johndoe manager --full-name "John Doe" --position "Store Manager"

# List all staff
./termpos staff list

# Get details for a staff member
./termpos staff get johndoe

# Find staff by search term
./termpos staff find manager
```

### Agent Mode (HTTP Server)

```bash
# Start the agent server
./termpos --mode agent --port 8000

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
./termpos --mode assistant

# Then use natural language commands like:
> add 5 muffins at $2.50
> sell 2 coffee
> show inventory
> what are my top selling products?
> show sales report
> show daily sales
```

### Backup and Security

```bash
# Create an encrypted backup
./termpos backup-enhanced --encrypt --verify

# Run the backup workflow
./termpos workflow backup run

# Check if data is sensitive
./termpos sensitive is-sensitive "api_key"
```

## Docker Deployment

### Docker Configuration

The project includes Docker support for containerized deployment. The Docker setup includes:

1. **Multi-stage build** for optimized image size
2. **Volume mounts** for persistent data storage
3. **Environment variable support** for configuration
4. **Security best practices** following container security guidelines

### Docker Image Architecture

The Docker image is built using a multi-stage approach:

1. **Builder stage**: Compiles the Go application with optimizations
2. **Runtime stage**: Minimal Debian-based image with only required dependencies

### Data Persistence

To ensure your data persists between container restarts:

```bash
# Create required directories on host
mkdir -p ./data ./config ./backups

# Run with mounted volumes
docker run -v $(pwd)/data:/app/data -v $(pwd)/config:/app/config -v $(pwd)/backups:/app/backups termpos:latest
```

### Environment Variables

The Docker container supports the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| POS_DB_PATH | Path to the SQLite database | /app/data/pos.db |
| POS_CONFIG_PATH | Path to the configuration file | /app/config/config.json |
| POS_ENCRYPTION_KEY | Key for encrypted backups | (randomly generated) |

### Production Deployment Example

For production deployment, consider using Docker Compose with environment variables:

```yaml
# Production docker-compose.yml example
version: '3.8'

services:
  termpos:
    image: termpos:latest
    restart: unless-stopped
    ports:
      - "8000:8000"
    volumes:
      - ./data:/app/data
      - ./config:/app/config
      - ./backups:/app/backups
    environment:
      - POS_DB_PATH=/app/data/pos.db
      - POS_CONFIG_PATH=/app/config/config.json
      - POS_ENCRYPTION_KEY=${POS_ENCRYPTION_KEY}
    command: ["--mode", "agent", "--port", "8000"]
```

Run with:
```bash
# Set encryption key (should be stored securely, not in plaintext)
export POS_ENCRYPTION_KEY="your-secure-key-here"

# Start the service
docker-compose up -d
```
