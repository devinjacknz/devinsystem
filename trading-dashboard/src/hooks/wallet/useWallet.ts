import { useState, useCallback } from 'react'
import { WalletInfo, WalletType, WalletState } from '../../types/wallet'
import { API_URL } from '../../utils/env'

export function useWallet() {
  const [state, setState] = useState<WalletState>({
    tradingWallet: null,
    profitWallet: null,
    isConnecting: false,
    error: null
  })

  const transfer = useCallback(async (
    fromType: WalletType,
    toType: WalletType,
    amount: number
  ) => {
    try {
      const response = await fetch(`${API_URL}/wallet/transfer`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        },
        body: JSON.stringify({ fromType, toType, amount })
      })

      const data = await response.json()
      if (!response.ok) {
        throw new Error(data.error || 'Transfer failed')
      }

      return data
    } catch (error) {
      throw error instanceof Error ? error : new Error('Transfer failed')
    }
  }, [])

  return {
    ...state,
    transfer
  }
}
