# Backend API Documentation

## Authentication
All protected endpoints require JWT authentication via the Authorization header. Tokens expire after 24 hours.

### Login
- **Endpoint**: POST /api/auth/login
- **Access**: Public
- **Description**: Authenticate user and receive JWT token
- **Request Body**:
```json
{
  "username": "string",
  "password": "string"
}
```
- **Success Response (200 OK)**:
```json
{
  "token": "string",
  "user": {
    "id": "string",
    "username": "string"
  }
}
```
- **Error Responses**:
  - 400 Bad Request: Invalid request format
  - 401 Unauthorized: Invalid credentials
  - 500 Internal Server Error: Server error

### Register
- **Endpoint**: POST /api/auth/register
- **Access**: Public
- **Description**: Create new user account
- **Request Body**:
```json
{
  "username": "string",
  "password": "string",
  "email": "string"
}
```
- **Success Response (201 Created)**:
```json
{
  "token": "string",
  "user": {
    "id": "string",
    "username": "string",
    "email": "string"
  }
}
```
- **Error Responses**:
  - 400 Bad Request: Invalid request format or username exists
  - 500 Internal Server Error: Server error

### Verify Token
- **Endpoint**: GET /api/auth/verify
- **Access**: Protected
- **Description**: Verify JWT token validity
- **Headers**: 
  - Authorization: Bearer {token}
- **Success Response (200 OK)**:
```json
{
  "valid": true,
  "user": {
    "id": "string",
    "username": "string"
  }
}
```
- **Error Responses**:
  - 401 Unauthorized: Invalid or expired token
  - 500 Internal Server Error: Server error

### Refresh Token
- **Endpoint**: POST /api/auth/refresh
- **Access**: Protected
- **Description**: Get new JWT token before expiration
- **Headers**: 
  - Authorization: Bearer {token}
- **Success Response (200 OK)**:
```json
{
  "token": "string"
}
```
- **Error Responses**:
  - 401 Unauthorized: Invalid or expired token
  - 500 Internal Server Error: Server error

## Trading Operations

### Solana DEX Endpoints

#### Get Pools
- **Endpoint**: GET /api/dex/pools
- **Access**: Protected
- **Description**: Get available liquidity pools on Solana DEX
- **Response (200 OK)**:
```json
{
  "pools": [
    {
      "id": "string",
      "token0": "string",
      "token1": "string",
      "liquidity": "number",
      "fee": "number"
    }
  ]
}
```

#### Get Pool Liquidity
- **Endpoint**: GET /api/dex/liquidity/{poolId}
- **Access**: Protected
- **Description**: Get detailed liquidity information for a specific pool
- **Parameters**:
  - poolId: Pool identifier
- **Response (200 OK)**:
```json
{
  "poolId": "string",
  "token0Balance": "number",
  "token1Balance": "number",
  "totalLiquidity": "number",
  "price": "number"
}
```

#### Execute Swap
- **Endpoint**: POST /api/dex/swap
- **Access**: Protected
- **Description**: Execute a token swap on Solana DEX
- **Request Body**:
```json
{
  "fromToken": "string",
  "toToken": "string",
  "amount": "number",
  "slippage": "number",
  "walletType": "trading|profit"
}
```
- **Response (201 Created)**:
```json
{
  "txHash": "string",
  "fromAmount": "number",
  "toAmount": "number",
  "price": "number",
  "fee": "number"
}
```

### Pump.fun Endpoints

#### Get Markets
- **Endpoint**: GET /api/pump/markets
- **Access**: Protected
- **Description**: Get available trading markets on Pump.fun
- **Response (200 OK)**:
```json
{
  "markets": [
    {
      "symbol": "string",
      "baseToken": "string",
      "quoteToken": "string",
      "price": "number",
      "volume24h": "number"
    }
  ]
}
```

#### Get Market Depth
- **Endpoint**: GET /api/pump/depth/{symbol}
- **Access**: Protected
- **Description**: Get order book depth for a specific market
- **Parameters**:
  - symbol: Trading pair symbol
- **Response (200 OK)**:
```json
{
  "symbol": "string",
  "bids": [["price", "amount"]],
  "asks": [["price", "amount"]],
  "timestamp": "number"
}
```

#### Place Order
- **Endpoint**: POST /api/pump/order
- **Access**: Protected
- **Description**: Place a new order on Pump.fun
- **Request Body**:
```json
{
  "symbol": "string",
  "side": "buy|sell",
  "type": "market|limit",
  "price": "number",
  "amount": "number",
  "walletType": "trading|profit"
}
```
- **Response (201 Created)**:
```json
{
  "orderId": "string",
  "status": "open|filled|partial",
  "filledAmount": "number",
  "avgPrice": "number",
  "timestamp": "number"
}
```

#### Cancel Order
- **Endpoint**: DELETE /api/pump/order/{orderId}
- **Access**: Protected
- **Description**: Cancel an existing order
- **Parameters**:
  - orderId: Order identifier
  - symbol: Trading pair symbol (query parameter)
- **Response (200 OK)**:
```json
{
  "orderId": "string",
  "status": "cancelled",
  "timestamp": "number"
}
```

## Wallet Management

### Create Wallet
- **Endpoint**: POST /api/wallet/create
- **Access**: Protected
- **Description**: Create a new wallet (trading or profit)
- **Request Body**:
```json
{
  "type": "trading|profit"
}
```
- **Response (201 Created)**:
```json
{
  "address": "string",
  "type": "trading|profit",
  "balance": 0
}
```

### Get Wallet Balance
- **Endpoint**: GET /api/wallet/balance/{type}
- **Access**: Protected
- **Description**: Get wallet balance
- **Parameters**:
  - type: Wallet type (trading|profit)
- **Response (200 OK)**:
```json
{
  "address": "string",
  "type": "trading|profit",
  "balance": "number"
}
```

### Transfer Funds
- **Endpoint**: POST /api/wallet/transfer
- **Access**: Protected
- **Description**: Transfer funds between trading and profit wallets
- **Request Body**:
```json
{
  "fromType": "trading|profit",
  "toType": "trading|profit",
  "amount": "number"
}
```
- **Response (200 OK)**:
```json
{
  "txHash": "string",
  "fromAddress": "string",
  "toAddress": "string",
  "amount": "number",
  "timestamp": "string"
}
```

### Get Wallet Transactions
- **Endpoint**: GET /api/wallet/transactions/{type}
- **Access**: Protected
- **Description**: Get wallet transaction history
- **Parameters**:
  - type: Wallet type (trading|profit)
  - limit: Maximum number of transactions (query parameter)
  - offset: Pagination offset (query parameter)
- **Response (200 OK)**:
```json
{
  "transactions": [
    {
      "txHash": "string",
      "type": "transfer|trade",
      "amount": "number",
      "timestamp": "string",
      "status": "completed|pending|failed"
    }
  ],
  "total": "number"
}
```

## AI Integration

### Analyze Market
- **Endpoint**: POST /api/ai/analyze/market
- **Access**: Protected
- **Description**: Get market analysis and trading signals using Ollama model
- **Request Body**:
```json
{
  "symbol": "string",
  "price": "number",
  "volume": "number",
  "trend": "string"
}
```
- **Response (200 OK)**:
```json
{
  "symbol": "string",
  "trend": "string",
  "confidence": "number",
  "signals": [
    {
      "type": "string",
      "symbol": "string",
      "action": "string",
      "confidence": "number"
    }
  ]
}
```

### Analyze Risk
- **Endpoint**: POST /api/ai/analyze/risk
- **Access**: Protected
- **Description**: Get risk analysis using DeepSeek model
- **Request Body**:
```json
{
  "symbol": "string",
  "price": "number",
  "volume": "number",
  "trend": "string"
}
```
- **Response (200 OK)**:
```json
{
  "symbol": "string",
  "riskLevel": "string",
  "stopLossPrice": "number",
  "confidence": "number"
}
```

### Get AI Signals
- **Endpoint**: GET /api/ai/signals/{symbol}
- **Access**: Protected
- **Description**: Get latest AI trading signals for a symbol
- **Parameters**:
  - symbol: Trading pair symbol
  - limit: Maximum number of signals (query parameter)
- **Response (200 OK)**:
```json
{
  "signals": [
    {
      "type": "string",
      "symbol": "string",
      "action": "string",
      "confidence": "number",
      "timestamp": "string"
    }
  ]
}
```

## WebSocket Endpoints

### Price Updates
- **Endpoint**: /api/ws/prices
- **Access**: Protected
- **Description**: Real-time price updates for trading pairs
- **Connection**: WebSocket
- **Authentication**: Send JWT token in connection query parameter
- **Events**:
  - **Price Update**:
  ```json
  {
    "type": "price",
    "data": {
      "symbol": "SOL/USD",
      "price": 100.0,
      "timestamp": "2024-02-03T12:00:00Z"
    }
  }
  ```
  - **Error**:
  ```json
  {
    "type": "error",
    "data": {
      "code": "string",
      "message": "string"
    }
  }
  ```

### Position Updates
- **Endpoint**: /api/ws/positions
- **Access**: Protected
- **Description**: Real-time position and PnL updates
- **Connection**: WebSocket
- **Authentication**: Send JWT token in connection query parameter
- **Events**:
  - **Position Update**:
  ```json
  {
    "type": "position",
    "data": {
      "symbol": "SOL/USD",
      "size": 10.0,
      "entryPrice": 100.0,
      "currentPrice": 105.0,
      "pnl": 50.0,
      "timestamp": "2024-02-03T12:00:00Z"
    }
  }
  ```
  - **Error**:
  ```json
  {
    "type": "error",
    "data": {
      "code": "string",
      "message": "string"
    }
  }
  ```

### Order Book Updates
- **Endpoint**: /api/ws/orderbook
- **Access**: Protected
- **Description**: Real-time order book updates
- **Connection**: WebSocket
- **Authentication**: Send JWT token in connection query parameter
- **Events**:
  - **Full Order Book**:
  ```json
  {
    "type": "snapshot",
    "data": {
      "symbol": "SOL/USD",
      "bids": [["price", "amount"]],
      "asks": [["price", "amount"]],
      "timestamp": "2024-02-03T12:00:00Z"
    }
  }
  ```
  - **Order Book Update**:
  ```json
  {
    "type": "update",
    "data": {
      "symbol": "SOL/USD",
      "bids": [["price", "amount"]],
      "asks": [["price", "amount"]],
      "timestamp": "2024-02-03T12:00:00Z"
    }
  }
  ```
  - **Error**:
  ```json
  {
    "type": "error",
    "data": {
      "code": "string",
      "message": "string"
    }
  }
  ```

## Example Requests

### Authentication Examples

```bash
# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "trader",
    "password": "secure123"
  }'

# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newtrader",
    "password": "secure123",
    "email": "trader@example.com"
  }'

# Verify Token
curl -X GET http://localhost:8080/api/auth/verify \
  -H "Authorization: Bearer <token>"
```

### Trading Examples

```bash
# Get Solana DEX Pools
curl -X GET http://localhost:8080/api/dex/pools \
  -H "Authorization: Bearer <token>"

# Execute Swap on Solana DEX
curl -X POST http://localhost:8080/api/dex/swap \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "fromToken": "SOL",
    "toToken": "USDC",
    "amount": 1.0,
    "slippage": 0.5,
    "walletType": "trading"
  }'

# Get Pump.fun Markets
curl -X GET http://localhost:8080/api/pump/markets \
  -H "Authorization: Bearer <token>"

# Place Order on Pump.fun
curl -X POST http://localhost:8080/api/pump/order \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "SOL/USD",
    "side": "buy",
    "type": "limit",
    "price": 100.0,
    "amount": 1.0,
    "walletType": "trading"
  }'
```

### Wallet Examples

```bash
# Create Wallet
curl -X POST http://localhost:8080/api/wallet/create \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "trading"
  }'

# Get Wallet Balance
curl -X GET http://localhost:8080/api/wallet/balance/trading \
  -H "Authorization: Bearer <token>"

# Transfer Funds
curl -X POST http://localhost:8080/api/wallet/transfer \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "fromType": "trading",
    "toType": "profit",
    "amount": 100.0
  }'
```

### AI Analysis Examples

```bash
# Analyze Market
curl -X POST http://localhost:8080/api/ai/analyze/market \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "SOL/USD",
    "price": 100.0,
    "volume": 1000000.0,
    "trend": "bullish"
  }'

# Get AI Signals
curl -X GET http://localhost:8080/api/ai/signals/SOL%2FUSD?limit=10 \
  -H "Authorization: Bearer <token>"
```

### WebSocket Examples

```javascript
// Connect to Price Updates
const ws = new WebSocket('ws://localhost:8080/api/ws/prices?token=<jwt_token>');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Price Update:', data);
};

// Connect to Position Updates
const ws = new WebSocket('ws://localhost:8080/api/ws/positions?token=<jwt_token>');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Position Update:', data);
};

// Connect to Order Book Updates
const ws = new WebSocket('ws://localhost:8080/api/ws/orderbook?token=<jwt_token>');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Order Book Update:', data);
};
```

## Testing Guide

### Setup
```bash
# Install dependencies
npm install -g wscat  # For WebSocket testing

# Set environment variables
export API_URL=http://localhost:8080
export WS_URL=ws://localhost:8080
```

### Authentication Testing
```bash
# Login
curl -X POST ${API_URL}/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass"
  }'

# Store token
export TOKEN=$(curl -s -X POST ${API_URL}/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass"}' | jq -r '.token')
```

### Trading Testing
```bash
# Get Solana DEX pools
curl -X GET ${API_URL}/api/dex/pools \
  -H "Authorization: Bearer ${TOKEN}"

# Execute swap
curl -X POST ${API_URL}/api/dex/swap \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "fromToken": "SOL",
    "toToken": "USDC",
    "amount": 1.0,
    "slippage": 0.5,
    "walletType": "trading"
  }'

# Place order on Pump.fun
curl -X POST ${API_URL}/api/pump/order \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "SOL/USD",
    "side": "buy",
    "type": "limit",
    "price": 100.0,
    "amount": 1.0,
    "walletType": "trading"
  }'
```

### Wallet Testing
```bash
# Create trading wallet
curl -X POST ${API_URL}/api/wallet/create \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"type": "trading"}'

# Transfer funds
curl -X POST ${API_URL}/api/wallet/transfer \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "fromType": "trading",
    "toType": "profit",
    "amount": 100.0
  }'
```

### AI Analysis Testing
```bash
# Market analysis
curl -X POST ${API_URL}/api/ai/analyze/market \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "SOL/USD",
    "price": 100.0,
    "volume": 1000000.0,
    "trend": "bullish"
  }'

# Risk analysis
curl -X POST ${API_URL}/api/ai/analyze/risk \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "SOL/USD",
    "price": 100.0,
    "volume": 1000000.0,
    "trend": "bullish"
  }'
```

### WebSocket Testing
```bash
# Connect to price updates
wscat -c "${WS_URL}/api/ws/prices?token=${TOKEN}"

# Connect to order book
wscat -c "${WS_URL}/api/ws/orderbook?token=${TOKEN}"

# Connect to positions
wscat -c "${WS_URL}/api/ws/positions?token=${TOKEN}"
```

### Integration Testing
```bash
#!/bin/bash
# Full trading workflow test

# Login and get token
TOKEN=$(curl -s -X POST ${API_URL}/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass"}' | jq -r '.token')

# Create trading wallet
WALLET=$(curl -s -X POST ${API_URL}/api/wallet/create \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"type":"trading"}')

# Get market analysis
ANALYSIS=$(curl -s -X POST ${API_URL}/api/ai/analyze/market \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "SOL/USD",
    "price": 100.0,
    "volume": 1000000.0,
    "trend": "bullish"
  }')

# Place order based on analysis
curl -X POST ${API_URL}/api/pump/order \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "SOL/USD",
    "side": "buy",
    "type": "limit",
    "price": 100.0,
    "amount": 1.0,
    "walletType": "trading"
  }'
```

## Security

### Authentication
- **JWT Token Requirements**:
  - All protected endpoints require valid JWT token in Authorization header
  - Token format: `Authorization: Bearer <token>`
  - Tokens expire after 24 hours
  - Tokens are signed using RS256 algorithm
  - Token payload includes user ID, roles, and permissions

- **WebSocket Security**:
  - Initial authentication required via token in query parameter
  - Connection automatically closed if token expires
  - Rate limiting applied per connection
  - Message size limits enforced

### Rate Limiting
- **HTTP Endpoints**:
  - 100 requests per minute per IP
  - 1000 requests per hour per user
  - Burst allowance: 20 requests
  - Response headers include rate limit status
  - Status codes: 429 Too Many Requests

- **WebSocket Rate Limits**:
  - 10 messages per second per connection
  - 1000 messages per minute per user
  - Connection closed if limits exceeded

### Access Control
- **Role-Based Access**:
  - Admin: Full system access
  - Trader: Trading and wallet operations
  - Viewer: Read-only access
- **IP Restrictions**:
  - Whitelist for admin access
  - Geographic restrictions configurable
  - Automatic blocking of suspicious IPs

### Data Protection
- **Transport Security**:
  - SSL/TLS required in production
  - TLS 1.2 or higher required
  - Strong cipher suites enforced
- **Request Validation**:
  - Input sanitization on all endpoints
  - Request size limits enforced
  - Content-Type validation
- **CORS Security**:
  - Enabled only for trusted domains
  - Credentials mode: include
  - Methods: GET, POST, PUT, DELETE

### Monitoring
- **Security Logging**:
  - All authentication attempts logged
  - Failed attempts tracked
  - Suspicious patterns monitored
- **Alerts**:
  - Immediate notification on:
    - Multiple failed login attempts
    - Unusual trading patterns
    - Rate limit violations
    - Geographic anomalies

## Error Handling

### Error Response Format
All error responses follow this standardized format:
```json
{
  "error": {
    "code": "string",
    "message": "string",
    "details": {
      "field": "string",
      "reason": "string",
      "suggestion": "string"
    }
  }
}
```

### HTTP Status Codes

#### Authentication Errors (401, 403)
- 401 Unauthorized
  - Missing token: Token not provided in Authorization header
  - Invalid token: Token is malformed or expired
  - Invalid credentials: Username or password incorrect
- 403 Forbidden
  - Insufficient permissions: User lacks required role/permissions
  - Account suspended: User account has been suspended
  - IP blocked: Client IP address has been blocked

#### Client Errors (400, 404, 409, 429)
- 400 Bad Request
  - Invalid parameters: Request parameters fail validation
  - Missing required field: Required request field not provided
  - Invalid format: Request body format is incorrect
- 404 Not Found
  - Resource not found: Requested resource does not exist
  - Invalid endpoint: API endpoint does not exist
- 409 Conflict
  - Resource exists: Attempting to create duplicate resource
  - State conflict: Resource state prevents operation
- 429 Too Many Requests
  - Rate limit exceeded: Too many requests in time period
  - Quota exceeded: API usage quota exceeded

#### Server Errors (500, 502, 503)
- 500 Internal Server Error
  - Database error: Error executing database operation
  - Integration error: External service integration failed
  - Processing error: Error processing request
- 502 Bad Gateway
  - Upstream error: Error from upstream service
  - Network error: Network connectivity issues
- 503 Service Unavailable
  - Maintenance mode: System under maintenance
  - Overloaded: System temporarily overloaded

### Error Examples

#### Authentication Error
```json
{
  "error": {
    "code": "AUTH_001",
    "message": "Invalid authentication token",
    "details": {
      "reason": "Token has expired",
      "suggestion": "Please login again to obtain a new token"
    }
  }
}
```

#### Validation Error
```json
{
  "error": {
    "code": "VAL_001",
    "message": "Invalid request parameters",
    "details": {
      "field": "amount",
      "reason": "Amount must be greater than 0",
      "suggestion": "Provide a positive number for amount"
    }
  }
}
```

#### Rate Limit Error
```json
{
  "error": {
    "code": "RATE_001",
    "message": "Rate limit exceeded",
    "details": {
      "limit": "100 requests per minute",
      "reset": "2024-02-03T12:01:00Z",
      "suggestion": "Please wait before making more requests"
    }
  }
}
```
