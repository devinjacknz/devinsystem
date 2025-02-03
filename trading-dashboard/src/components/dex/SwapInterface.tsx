import React, { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card'

interface SwapProps {
  fromToken: string
  toToken: string
  slippage: number
  onSwap: (from: string, to: string, amount: number) => void
}

export function SwapInterface({ fromToken, toToken, slippage, onSwap }: SwapProps) {
  const [amount, setAmount] = useState<string>('')
  const [isLoading, setIsLoading] = useState(false)

  const handleSwap = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!amount) return

    setIsLoading(true)
    try {
      await onSwap(fromToken, toToken, parseFloat(amount))
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Swap {fromToken} to {toToken}</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSwap} className="space-y-4">
          <div className="space-y-2">
            <label htmlFor="amount" className="text-sm font-medium">
              Amount ({fromToken})
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
          <div className="text-sm text-muted-foreground">
            Slippage Tolerance: {slippage}%
          </div>
          <button
            type="submit"
            disabled={isLoading || !amount}
            className="w-full p-2 bg-primary text-primary-foreground rounded hover:bg-primary/90 disabled:opacity-50"
          >
            {isLoading ? 'Swapping...' : 'Swap'}
          </button>
        </form>
      </CardContent>
    </Card>
  )
}
