import React, { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card'
import { WalletInfo, WalletTransfer } from '../../types/wallet'

interface WalletManagerProps {
  tradingWallet: WalletInfo | null;
  profitWallet: WalletInfo | null;
  onConnect: () => Promise<void>;
  onTransfer: (transfer: WalletTransfer) => Promise<void>;
  isConnecting: boolean;
  error: string | null;
}

export function WalletManager({
  tradingWallet,
  profitWallet,
  onConnect,
  onTransfer,
  isConnecting,
  error
}: WalletManagerProps) {
  const [amount, setAmount] = useState('')
  const [isTransferring, setIsTransferring] = useState(false)

  const handleTransfer = async (fromType: WalletInfo['type'], toType: WalletInfo['type']) => {
    if (!amount) return
    
    setIsTransferring(true)
    try {
      await onTransfer({
        fromType,
        toType,
        amount: parseFloat(amount)
      })
      setAmount('')
    } finally {
      setIsTransferring(false)
    }
  }

  if (!tradingWallet || !profitWallet) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Wallet Connection</CardTitle>
        </CardHeader>
        <CardContent>
          {error && (
            <div className="text-red-500 text-sm mb-4" role="alert">
              {error}
            </div>
          )}
          <button
            onClick={onConnect}
            disabled={isConnecting}
            className="w-full p-2 bg-primary text-primary-foreground rounded hover:bg-primary/90 disabled:opacity-50"
          >
            {isConnecting ? 'Connecting...' : 'Connect Wallets'}
          </button>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Wallet Management</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <h3 className="text-sm font-medium">Trading Wallet</h3>
              <p className="text-xs text-muted-foreground break-all">{tradingWallet.address}</p>
              <p className="font-medium">{tradingWallet.balance.toFixed(4)} SOL</p>
            </div>
            <div className="space-y-2">
              <h3 className="text-sm font-medium">Profit Wallet</h3>
              <p className="text-xs text-muted-foreground break-all">{profitWallet.address}</p>
              <p className="font-medium">{profitWallet.balance.toFixed(4)} SOL</p>
            </div>
          </div>

          <div className="space-y-4">
            <div className="space-y-2">
              <label htmlFor="amount" className="text-sm font-medium">
                Transfer Amount
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
                onClick={() => handleTransfer('trading', 'profit')}
                disabled={isTransferring || !amount}
                className="p-2 bg-primary text-primary-foreground rounded hover:bg-primary/90 disabled:opacity-50"
              >
                {isTransferring ? 'Processing...' : 'Trading → Profit'}
              </button>
              <button
                onClick={() => handleTransfer('profit', 'trading')}
                disabled={isTransferring || !amount}
                className="p-2 bg-primary text-primary-foreground rounded hover:bg-primary/90 disabled:opacity-50"
              >
                {isTransferring ? 'Processing...' : 'Profit → Trading'}
              </button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
