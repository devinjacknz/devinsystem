import { useState, useCallback } from 'react'
import { WalletInfo, WalletState, WalletTransfer } from '../../types/wallet'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export function useWallet() {
  const [state, setState] = useState<WalletState>({
    tradingWallet: null,
    profitWallet: null,
    isConnecting: false,
    error: null
  })

  const connectWallets = useCallback(async () => {
    setState(prev => ({ ...prev, isConnecting: true, error: null }))
    try {
      const response = await fetch(`${API_URL}/wallet/connect`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        }
      })

      if (!response.ok) {
        throw new Error('Failed to connect wallets')
      }

      const data: { tradingWallet: WalletInfo; profitWallet: WalletInfo } = await response.json()
      setState({
        tradingWallet: data.tradingWallet,
        profitWallet: data.profitWallet,
        isConnecting: false,
        error: null
      })
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to connect wallets',
        isConnecting: false
      }))
    }
  }, [])

  const transfer = useCallback(async (transfer: WalletTransfer) => {
    try {
      const response = await fetch(`${API_URL}/wallet/transfer`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        },
        body: JSON.stringify(transfer)
      })

      if (!response.ok) {
        throw new Error('Transfer failed')
      }

      const data: { tradingWallet: WalletInfo; profitWallet: WalletInfo } = await response.json()
      setState(prev => ({
        ...prev,
        tradingWallet: data.tradingWallet,
        profitWallet: data.profitWallet,
        error: null
      }))
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Transfer failed'
      }))
    }
  }, [])

  return {
    ...state,
    connectWallets,
    transfer
  }
}
