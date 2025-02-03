import { renderHook, act } from '@testing-library/react'
import { useAIAnalysis } from '../ai/useAIAnalysis'
import type { MarketData } from '../../types/trading'

describe('useAIAnalysis', () => {
  const mockMarketData: MarketData = {
    symbol: 'SOL/USDC',
    price: 100.50,
    volume: 1000000,
    change24h: 5.5
  }

  const mockAnalysis = {
    sentiment: 'bullish',
    riskScore: 0.7,
    priceTarget: 0.00001500,
    confidence: 0.85,
    signals: [
      { type: 'momentum', value: 0.8, description: 'Strong upward trend' },
      { type: 'volume', value: 0.6, description: 'Above average volume' }
    ]
  }

  beforeEach(() => {
    jest.clearAllMocks()
    global.fetch = jest.fn()
  })

  it('fetches analysis for market data', async () => {
    ;(global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockAnalysis)
    })

    const { result } = renderHook(() => useAIAnalysis(mockMarketData))

    expect(result.current.isLoading).toBe(true)

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 100))
    })

    expect(result.current.analysis).toEqual(mockAnalysis)
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('handles loading state', () => {
    ;(global.fetch as jest.Mock).mockImplementationOnce(
      () => new Promise(resolve => setTimeout(resolve, 1000))
    )

    const { result } = renderHook(() => useAIAnalysis(mockMarketData))

    expect(result.current.isLoading).toBe(true)
    expect(result.current.analysis).toBeNull()
    expect(result.current.error).toBeNull()
  })

  it('handles error state', async () => {
    const mockError = new Error('Failed to fetch analysis')
    ;(global.fetch as jest.Mock).mockRejectedValueOnce(mockError)

    const { result } = renderHook(() => useAIAnalysis(mockMarketData))

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 0))
    })

    expect(result.current.error).toBe('Failed to fetch analysis')
    expect(result.current.isLoading).toBe(false)
    expect(result.current.analysis).toBeNull()
  })

  it('updates analysis when market data changes', async () => {
    const updatedMarketData = { ...mockMarketData, price: 110.50 }
    const updatedAnalysis = { ...mockAnalysis, sentiment: 'bearish' }

    ;(global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockAnalysis)
      })
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(updatedAnalysis)
      })

    const { result, rerender } = renderHook(
      ({ data }) => useAIAnalysis(data),
      { initialProps: { data: mockMarketData } }
    )

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 100))
    })

    expect(result.current.analysis).toEqual(mockAnalysis)

    rerender({ data: updatedMarketData })

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 100))
    })

    expect(result.current.analysis).toEqual(updatedAnalysis)
  })

  it('cancels pending requests when unmounted', async () => {
    const mockAbortController = new AbortController()
    const mockAbort = jest.spyOn(mockAbortController, 'abort')
    ;(global as any).AbortController = jest.fn(() => mockAbortController)

    const { unmount } = renderHook(() => useAIAnalysis(mockMarketData))

    unmount()

    expect(mockAbort).toHaveBeenCalled()
  })
})
