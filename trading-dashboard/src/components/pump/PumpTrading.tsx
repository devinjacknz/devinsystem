import React, { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card'
import { TokenInfo, TradeAction } from '../../types/trading'

interface PumpTradingProps {
  tokens: TokenInfo[]
  onTrade: (action: TradeAction) => Promise<void>
}

export function PumpTrading({ tokens, onTrade }: PumpTradingProps) {
  const [selectedToken, setSelectedToken] = useState<string>('')
  const [amount, setAmount] = useState<string>('')
  const [isLoading, setIsLoading] = useState(false)

  const handleTrade = async (type: 'buy' | 'sell') => {
    if (!selectedToken || !amount) return

    const token = tokens.find(t => t.symbol === selectedToken)
    if (!token) return

    setIsLoading(true)
    try {
      await onTrade({
        type,
        token: selectedToken,
        amount: parseFloat(amount),
        price: token.price
      })
      setAmount('')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Pump.fun Trading</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="space-y-2">
            <label htmlFor="token" className="text-sm font-medium">
              Select Token
            </label>
            <select
              id="token"
              value={selectedToken}
              onChange={(e) => setSelectedToken(e.target.value)}
              className="w-full p-2 border rounded"
              required
            >
              <option value="">Select a token</option>
              {tokens.map(token => (
                <option key={token.symbol} value={token.symbol}>
                  {token.name} ({token.symbol}) - ${token.price.toFixed(4)}
                </option>
              ))}
            </select>
          </div>

          <div className="space-y-2">
            <label htmlFor="amount" className="text-sm font-medium">
              Amount
            </label>
            <input
              id="amount"
              type="number"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              className="w-full p-2 border rounded"
              min="0"
              step="0.000001"
              required
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <button
              onClick={() => handleTrade('buy')}
              disabled={isLoading || !amount || !selectedToken}
              className="p-2 bg-green-500 text-white rounded hover:bg-green-600 disabled:opacity-50"
            >
              {isLoading ? 'Processing...' : 'Buy'}
            </button>
            <button
              onClick={() => handleTrade('sell')}
              disabled={isLoading || !amount || !selectedToken}
              className="p-2 bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50"
            >
              {isLoading ? 'Processing...' : 'Sell'}
            </button>
          </div>

          {selectedToken && (
            <>
              <div className="grid grid-cols-2 gap-4">
                {tokens.filter(t => t.symbol === selectedToken).map(token => (
                  <div key={token.symbol} className="space-y-1">
                    <p className="text-sm text-muted-foreground">24h Volume</p>
                    <p className="font-medium">${token.volume24h.toLocaleString()}</p>
                    <p className="text-sm text-muted-foreground">24h Change</p>
                    <p className={`font-medium ${token.change24h >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                      {token.change24h >= 0 ? '+' : ''}{token.change24h.toFixed(2)}%
                    </p>
                  </div>
                ))}
              </div>

              <div className="space-y-2">
                <label htmlFor="amount" className="text-sm font-medium">
                  Amount
                </label>
                <input
                  id="amount"
                  type="number"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="w-full p-2 border rounded"
                  min="0"
                  step="0.000001"
                  required
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <button
                  onClick={() => handleTrade('buy')}
                  disabled={isLoading || !amount}
                  className="p-2 bg-green-500 text-white rounded hover:bg-green-600 disabled:opacity-50"
                >
                  {isLoading ? 'Processing...' : 'Buy'}
                </button>
                <button
                  onClick={() => handleTrade('sell')}
                  disabled={isLoading || !amount}
                  className="p-2 bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50"
                >
                  {isLoading ? 'Processing...' : 'Sell'}
                </button>
              </div>
            </>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
