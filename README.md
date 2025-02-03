# Solana Trading System

A SaaS-level quantitative trading system for Solana DEX and Pump.fun with local AI model integration.

## Core Components

- **Wallet Management System**
  - Secure AB wallet system (A for trading, B for profit collection)
  - HSM key storage integration
  - Solana wallet implementation

- **Trading Engine**
  - Unified exchange adapters
  - Solana DEX integration
  - Pump.fun integration
  - Order book management

- **AI Model Service**
  - Ollama integration
  - DeepSeek R1 integration
  - Market analysis service

- **Risk Control Module**
  - Stop-loss implementation
  - Slippage protection
  - Risk manager with exposure control

- **Frontend Dashboard**
  - React implementation with TypeScript
  - Real-time data visualization
  - WebSocket integration
  - Error handling and retry mechanism

- **API Gateway**
  - JWT authentication
  - WebSocket endpoints for real-time data
  - RESTful trading endpoints

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- pnpm
- Ollama
- DeepSeek R1

### Installation

1. Clone the repository:
```bash
gh repo clone devinjacknz/devintrade
cd devintrade
```

2. Install backend dependencies:
```bash
go mod download
```

3. Install frontend dependencies:
```bash
cd trading-dashboard
pnpm install
```

4. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

### Running the System

1. Start the trading engine:
```bash
go run cmd/trader/main.go
```

2. Start the API server:
```bash
go run cmd/api/main.go
```

3. Start the frontend dashboard:
```bash
cd trading-dashboard
pnpm dev
```

## Testing

Run backend tests:
```bash
go test ./... -v -race
```

Run frontend tests:
```bash
cd trading-dashboard
pnpm test
```

## Architecture

The system follows SOLID principles and a modular architecture:

- **Wallet Module**: Manages trading and profit collection wallets with secure key storage
- **Trading Engine**: Handles order execution and market interactions
- **AI Service**: Provides market analysis and trading signals
- **Risk Module**: Implements risk management and protection mechanisms
- **API Gateway**: Handles authentication and provides WebSocket/REST endpoints
- **Frontend**: Real-time dashboard for monitoring and control

## Security

- JWT-based authentication
- HSM integration for key storage
- Secure AB wallet system
- Rate limiting and request validation

## Contributing

1. Create a new branch: `git checkout -b devin/$(date +%s)-feature-name`
2. Make changes and test
3. Create a pull request

## License

MIT License
