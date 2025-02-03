import { renderHook, act } from '@testing-library/react'
import { useWallet } from '../wallet/useWallet'
import type { WalletTransfer } from '../../types/wallet'

describe('useWallet', () => {
  const mockTradingWallet = {
    address: 'trading-wallet-address',
    balance: 1000,
    type: 'trading'
  }

  const mockProfitWallet = {
    address: 'profit-wallet-address',
    balance: 500,
    type: 'profit'
  }

  beforeEach(() => {
    jest.clearAllMocks()
    global.fetch = jest.fn()
  })

  it('initializes wallet state correctly', async () => {
    ;(global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({
        tradingWallet: mockTradingWallet,
        profitWallet: mockProfitWallet
      })
    })

    const { result } = renderHook(() => useWallet())

    await act(async () => {
      await result.current.connectWallets()
    })

    expect(result.current.tradingWallet).toEqual(mockTradingWallet)
    expect(result.current.profitWallet).toEqual(mockProfitWallet)
    expect(result.current.isConnecting).toBe(false)
  })

  it('handles wallet transfer between trading and profit wallets', async () => {
    const mockTransferResponse = { hash: 'tx-hash' }
    ;(global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          tradingWallet: mockTradingWallet,
          profitWallet: mockProfitWallet
        })
      })
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockTransferResponse)
      })

    const { result } = renderHook(() => useWallet())

    await act(async () => {
      await result.current.connectWallets()
    })

    const transfer: WalletTransfer = {
      fromType: 'trading',
      toType: 'profit',
      amount: 100
    }

    await act(async () => {
      await result.current.transfer(transfer)
    })

    expect(global.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/wallet/transfer'),
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify(transfer)
      })
    )
  })

  it('handles wallet connection errors', async () => {
    const mockError = new Error('Failed to connect wallet')
    ;(global.fetch as jest.Mock).mockRejectedValueOnce(mockError)

    const { result } = renderHook(() => useWallet())

    await act(async () => {
      await result.current.connectWallets()
    })

    expect(result.current.error).toBe('Failed to connect wallet')
    expect(result.current.isConnecting).toBe(false)
  })

  it('validates transfer amounts against wallet balances', async () => {
    ;(global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({
        tradingWallet: mockTradingWallet,
        profitWallet: mockProfitWallet
      })
    })

    const { result } = renderHook(() => useWallet())

    await act(async () => {
      await result.current.connectWallets()
    })

    const transfer: WalletTransfer = {
      fromType: 'trading',
      toType: 'profit',
      amount: 2000
    }

    await expect(
      result.current.transfer(transfer)
    ).rejects.toThrow('Insufficient balance')
  })
})
