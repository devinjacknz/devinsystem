import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { TradingDashboard } from '../../TradingDashboard'
import { useWallet } from '../../../hooks/wallet/useWallet'
import { useAIAnalysis } from '../../../hooks/ai/useAIAnalysis'
import { useWebSocket } from '../../../hooks/websocket/useWebSocket'

jest.mock('../../../hooks/wallet/useWallet')
jest.mock('../../../hooks/ai/useAIAnalysis')
jest.mock('../../../hooks/websocket/useWebSocket')

describe('Trading Flow Integration', () => {
  const mockWallet = {
    tradingWallet: {
      address: 'trading-wallet-address',
      balance: 1000,
      type: 'trading'
    },
    profitWallet: {
      address: 'profit-wallet-address',
      balance: 500,
      type: 'profit'
    },
    isConnected: true,
    error: null,
    executeTrade: jest.fn(),
    transfer: jest.fn()
  }

  const mockAIAnalysis = {
    sentiment: 'bullish',
    riskScore: 0.7,
    priceTarget: 0.00001500,
    confidence: 0.85,
    signals: [
      { type: 'momentum', value: 0.8, description: 'Strong upward trend' },
      { type: 'volume', value: 0.6, description: 'Above average volume' }
    ]
  }

  const mockWebSocket = {
    send: jest.fn(),
    lastMessage: null,
    readyState: WebSocket.OPEN
  }

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useWallet as jest.Mock).mockReturnValue(mockWallet)
    ;(useAIAnalysis as jest.Mock).mockReturnValue({ analysis: mockAIAnalysis, isLoading: false, error: null })
    ;(useWebSocket as jest.Mock).mockReturnValue(mockWebSocket)
  })

  it('completes full DEX trading flow', async () => {
    render(<TradingDashboard />)
    
    const dexTab = screen.getByRole('tab', { name: /dex trading/i })
    fireEvent.click(dexTab)

    const fromTokenSelect = screen.getByLabelText(/from token/i)
    const toTokenSelect = screen.getByLabelText(/to token/i)
    const amountInput = screen.getByLabelText(/amount/i)

    fireEvent.change(fromTokenSelect, { target: { value: 'SOL' } })
    fireEvent.change(toTokenSelect, { target: { value: 'USDC' } })
    fireEvent.change(amountInput, { target: { value: '10' } })

    expect(screen.getByText(/sentiment: bullish/i)).toBeInTheDocument()
    expect(screen.getByText(/confidence: 85%/i)).toBeInTheDocument()

    const swapButton = screen.getByRole('button', { name: /swap/i })
    fireEvent.click(swapButton)

    const confirmButton = screen.getByRole('button', { name: /confirm/i })
    fireEvent.click(confirmButton)

    await waitFor(() => {
      expect(mockWallet.executeTrade).toHaveBeenCalledWith({
        type: 'swap',
        fromToken: 'SOL',
        toToken: 'USDC',
        amount: 10
      })
    })
  })

  it('completes full Pump.fun trading flow', async () => {
    render(<TradingDashboard />)
    
    const pumpTab = screen.getByRole('tab', { name: /pump\.fun/i })
    fireEvent.click(pumpTab)

    const searchInput = screen.getByLabelText(/token search/i)
    fireEvent.change(searchInput, { target: { value: 'pepe' } })
    
    await waitFor(() => {
      fireEvent.click(screen.getByText(/PEPE \/ USDT/i))
    })

    expect(screen.getByTestId('market-depth-chart')).toBeInTheDocument()

    const amountInput = screen.getByLabelText(/amount/i)
    fireEvent.change(amountInput, { target: { value: '500' } })

    const buyButton = screen.getByRole('button', { name: /buy/i })
    fireEvent.click(buyButton)

    const confirmButton = screen.getByRole('button', { name: /confirm/i })
    fireEvent.click(confirmButton)

    await waitFor(() => {
      expect(mockWallet.executeTrade).toHaveBeenCalledWith({
        type: 'buy',
        token: 'PEPE',
        amount: 500,
        price: expect.any(Number)
      })
    })
  })

  it('completes wallet transfer flow', async () => {
    render(<TradingDashboard />)
    
    const walletTab = screen.getByRole('tab', { name: /wallet/i })
    fireEvent.click(walletTab)

    const amountInput = screen.getByLabelText(/amount/i)
    const fromSelect = screen.getByLabelText(/from wallet/i)
    const toSelect = screen.getByLabelText(/to wallet/i)
    
    fireEvent.change(amountInput, { target: { value: '100' } })
    fireEvent.change(fromSelect, { target: { value: 'trading' } })
    fireEvent.change(toSelect, { target: { value: 'profit' } })

    const transferButton = screen.getByRole('button', { name: /transfer/i })
    fireEvent.click(transferButton)

    await waitFor(() => {
      expect(mockWallet.transfer).toHaveBeenCalledWith({
        fromType: 'trading',
        toType: 'profit',
        amount: 100
      })
    })
  })

  it('handles WebSocket updates during trading', async () => {
    render(<TradingDashboard />)

    const priceUpdate = {
      type: 'price',
      data: {
        symbol: 'SOL/USDC',
        price: 100.50
      }
    }

    const messageCallback = (useWebSocket as jest.Mock).mock.calls[0][1].onMessage
    messageCallback(priceUpdate)

    await waitFor(() => {
      expect(screen.getByText(/100\.50/)).toBeInTheDocument()
    })

    const orderBookUpdate = {
      type: 'orderBook',
      data: {
        bids: [[100, 1.5]],
        asks: [[101, 2.0]]
      }
    }

    messageCallback(orderBookUpdate)

    await waitFor(() => {
      expect(screen.getByTestId('order-book')).toHaveTextContent('100.00')
      expect(screen.getByTestId('order-book')).toHaveTextContent('101.00')
    })
  })

  it('handles error states gracefully', async () => {
    ;(useWallet as jest.Mock).mockReturnValue({
      ...mockWallet,
      error: 'Connection failed',
      isConnected: false
    })

    render(<TradingDashboard />)
    
    expect(screen.getByText(/connection failed/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /reconnect/i })).toBeInTheDocument()

    const reconnectButton = screen.getByRole('button', { name: /reconnect/i })
    fireEvent.click(reconnectButton)

    await waitFor(() => {
      expect(screen.queryByText(/connection failed/i)).not.toBeInTheDocument()
    })
  })

  it('validates trading inputs', async () => {
    render(<TradingDashboard />)
    
    const dexTab = screen.getByRole('tab', { name: /dex trading/i })
    fireEvent.click(dexTab)

    const swapButton = screen.getByRole('button', { name: /swap/i })
    fireEvent.click(swapButton)

    expect(screen.getByText(/amount required/i)).toBeInTheDocument()
    expect(screen.getByText(/select tokens/i)).toBeInTheDocument()
  })

  it('updates UI based on WebSocket connection state', async () => {
    ;(useWebSocket as jest.Mock).mockReturnValue({
      ...mockWebSocket,
      readyState: WebSocket.CLOSED
    })

    render(<TradingDashboard />)
    
    expect(screen.getByText(/disconnected/i)).toBeInTheDocument()

    ;(useWebSocket as jest.Mock).mockReturnValue({
      ...mockWebSocket,
      readyState: WebSocket.OPEN
    })

    await waitFor(() => {
      expect(screen.getByText(/connected/i)).toBeInTheDocument()
    })
  })
})
