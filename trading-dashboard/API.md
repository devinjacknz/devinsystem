# Trading System API Documentation

## Testing

### Component Testing
```typescript
// Example: Testing SwapInterface
import { render, screen, fireEvent } from '@testing-library/react'
import { SwapInterface } from '../components/dex/SwapInterface'

describe('SwapInterface', () => {
  it('executes swap transaction', async () => {
    render(<SwapInterface />)
    
    fireEvent.change(screen.getByLabelText(/from token/i), { 
      target: { value: 'SOL' } 
    })
    fireEvent.change(screen.getByLabelText(/to token/i), { 
      target: { value: 'USDC' } 
    })
    fireEvent.change(screen.getByLabelText(/amount/i), { 
      target: { value: '10' } 
    })

    fireEvent.click(screen.getByRole('button', { name: /swap/i }))
    
    await waitFor(() => {
      expect(mockWallet.executeTrade).toHaveBeenCalledWith({
        type: 'swap',
        fromToken: 'SOL',
        toToken: 'USDC',
        amount: 10
      })
    })
  })
})
```

### Hook Testing
```typescript
// Example: Testing useWebSocket
import { renderHook } from '@testing-library/react'
import { useWebSocket } from '../hooks/useWebSocket'

describe('useWebSocket', () => {
  it('handles connection lifecycle', async () => {
    const { result } = renderHook(() => useWebSocket('ws://test'))
    
    expect(result.current.readyState).toBe(WebSocket.CONNECTING)
    
    // Simulate connection open
    act(() => {
      mockWebSocket.onopen({})
    })
    
    expect(result.current.readyState).toBe(WebSocket.OPEN)
  })
})
```

### Integration Testing
```typescript
// Example: Testing complete trading flow
describe('Trading Flow', () => {
  it('completes DEX swap workflow', async () => {
    render(<TradingDashboard />)
    
    // Select DEX trading tab
    fireEvent.click(screen.getByRole('tab', { name: /dex trading/i }))
    
    // Configure swap
    fireEvent.change(screen.getByLabelText(/from token/i), {
      target: { value: 'SOL' }
    })
    fireEvent.change(screen.getByLabelText(/to token/i), {
      target: { value: 'USDC' }
    })
    
    // Execute trade
    fireEvent.click(screen.getByRole('button', { name: /swap/i }))
    
    await waitFor(() => {
      expect(mockWallet.executeTrade).toHaveBeenCalled()
    })
  })
})
```

## Components

### SwapInterface
Component for DEX token swapping operations.

```typescript
interface SwapProps {
  fromToken: string;
  toToken: string;
  slippage: number;
  onSwap: (from: string, to: string, amount: number) => void;
}

// Usage Example
<SwapInterface
  fromToken="SOL"
  toToken="USDC"
  slippage={0.5}
  onSwap={(from, to, amount) => handleSwap(from, to, amount)}
/>
```

### PumpTrading
Component for trading meme coins on Pump.fun.

```typescript
interface PumpTradingProps {
  tokens: string[];
  marketDepth: MarketDepth;
  onTrade: (token: string, amount: number, type: 'buy' | 'sell') => void;
}

// Usage Example
<PumpTrading
  tokens={['PEPE', 'DOGE']}
  marketDepth={marketDepthData}
  onTrade={(token, amount, type) => handleTrade(token, amount, type)}
/>
```

### WalletManager
Component for managing AB wallet system.

```typescript
interface WalletProps {
  tradingWallet: WalletInfo;
  profitWallet: WalletInfo;
  onTransfer: (from: WalletType, to: WalletType, amount: number) => void;
}

// Usage Example
<WalletManager
  tradingWallet={tradingWalletInfo}
  profitWallet={profitWalletInfo}
  onTransfer={(from, to, amount) => handleTransfer(from, to, amount)}
/>
```

### AIInsights
Component for displaying AI model analysis.

```typescript
interface AIInsightsProps {
  marketData: MarketData;
  analysis: Analysis;
  riskMetrics: RiskAnalysis;
}

// Usage Example
<AIInsights
  marketData={currentMarketData}
  analysis={aiAnalysis}
  riskMetrics={riskMetrics}
/>
```

## Hooks

### useWebSocket
Hook for WebSocket connections with automatic reconnection.

```typescript
const { send, lastMessage, readyState } = useWebSocket(url, {
  onOpen: () => console.log('Connected'),
  onMessage: (data) => handleMessage(data),
  onClose: () => console.log('Disconnected'),
  onError: (error) => handleError(error)
});
```

### useWallet
Hook for wallet management and transactions.

```typescript
const {
  tradingWallet,
  profitWallet,
  isConnected,
  error,
  transfer,
  executeTrade
} = useWallet();
```

### useAIAnalysis
Hook for AI model integration and market analysis.

```typescript
const {
  analysis,
  isLoading,
  error
} = useAIAnalysis(marketData);
```

## Trading Flows

### DEX Swap Flow
1. Connect wallet
2. Select tokens and amount
3. Review AI analysis
4. Confirm swap
5. Handle transaction result

### Pump.fun Trading Flow
1. Connect wallet
2. Search and select token
3. View market depth
4. Review AI insights
5. Place buy/sell order
6. Monitor order status

### Wallet Transfer Flow
1. Select source wallet (trading/profit)
2. Select destination wallet
3. Enter amount
4. Validate balance
5. Confirm transfer
6. Handle transaction result

## WebSocket Events

### Price Updates
```typescript
interface PriceUpdate {
  type: 'price';
  data: {
    symbol: string;
    price: number;
    timestamp: number;
  }
}
```

### Order Book Updates
```typescript
interface OrderBookUpdate {
  type: 'orderBook';
  data: {
    symbol: string;
    bids: [number, number][];
    asks: [number, number][];
  }
}
```

### Trade Updates
```typescript
interface TradeUpdate {
  type: 'trade';
  data: {
    symbol: string;
    price: number;
    amount: number;
    side: 'buy' | 'sell';
    timestamp: number;
  }
}
```
