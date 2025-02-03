import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { PumpTrading } from '../../pump/PumpTrading'
import { useWallet } from '../../../hooks/wallet/useWallet'
import { useAIAnalysis } from '../../../hooks/ai/useAIAnalysis'

jest.mock('../../../hooks/wallet/useWallet')
jest.mock('../../../hooks/ai/useAIAnalysis')

describe('PumpTrading', () => {
  const mockWallet = {
    tradingWallet: {
      address: 'trading-wallet-address',
      balance: 1000,
      type: 'trading'
    },
    isConnected: true,
    error: null,
    executeTrade: jest.fn()
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

  const mockMarketDepth = {
    bids: [
      [0.00001400, 1000000],
      [0.00001300, 2000000]
    ],
    asks: [
      [0.00001600, 800000],
      [0.00001700, 1500000]
    ]
  }

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useWallet as jest.Mock).mockReturnValue(mockWallet)
    ;(useAIAnalysis as jest.Mock).mockReturnValue({ analysis: mockAIAnalysis, isLoading: false, error: null })
  })

  it('renders trading interface correctly', () => {
    render(<PumpTrading marketDepth={mockMarketDepth} />)
    
    expect(screen.getByTestId('market-depth-chart')).toBeInTheDocument()
    expect(screen.getByLabelText(/amount/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /buy/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sell/i })).toBeInTheDocument()
  })

  it('executes buy order', async () => {
    render(<PumpTrading marketDepth={mockMarketDepth} />)
    
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
        price: 0.00001600
      })
    })
  })

  it('executes sell order', async () => {
    render(<PumpTrading marketDepth={mockMarketDepth} />)
    
    const amountInput = screen.getByLabelText(/amount/i)
    fireEvent.change(amountInput, { target: { value: '300' } })

    const sellButton = screen.getByRole('button', { name: /sell/i })
    fireEvent.click(sellButton)

    const confirmButton = screen.getByRole('button', { name: /confirm/i })
    fireEvent.click(confirmButton)

    await waitFor(() => {
      expect(mockWallet.executeTrade).toHaveBeenCalledWith({
        type: 'sell',
        token: 'PEPE',
        amount: 300,
        price: 0.00001400
      })
    })
  })

  it('displays market depth information', () => {
    render(<PumpTrading marketDepth={mockMarketDepth} />)
    
    expect(screen.getByText(/best bid: 0.0000140/i)).toBeInTheDocument()
    expect(screen.getByText(/best ask: 0.0000160/i)).toBeInTheDocument()
    expect(screen.getByText(/spread: 12.5%/i)).toBeInTheDocument()
  })

  it('shows AI analysis insights', () => {
    render(<PumpTrading marketDepth={mockMarketDepth} />)
    
    expect(screen.getByText(/sentiment: bullish/i)).toBeInTheDocument()
    expect(screen.getByText(/risk score: 0.7/i)).toBeInTheDocument()
    expect(screen.getByText(/confidence: 85%/i)).toBeInTheDocument()
  })

  it('validates input amount against wallet balance', () => {
    render(<PumpTrading marketDepth={mockMarketDepth} />)
    
    const amountInput = screen.getByLabelText(/amount/i)
    fireEvent.change(amountInput, { target: { value: '2000' } })

    expect(screen.getByText(/insufficient balance/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /buy/i })).toBeDisabled()
    expect(screen.getByRole('button', { name: /sell/i })).toBeDisabled()
  })

  it('displays loading states', () => {
    ;(useAIAnalysis as jest.Mock).mockReturnValue({ analysis: null, isLoading: true, error: null })
    
    render(<PumpTrading marketDepth={mockMarketDepth} />)
    
    expect(screen.getByText(/loading analysis/i)).toBeInTheDocument()
  })

  it('handles error states', () => {
    ;(useWallet as jest.Mock).mockReturnValue({
      ...mockWallet,
      error: 'Failed to connect wallet'
    })
    
    render(<PumpTrading marketDepth={mockMarketDepth} />)
    
    expect(screen.getByText(/failed to connect wallet/i)).toBeInTheDocument()
  })

  it('updates price impact warning based on order size', () => {
    render(<PumpTrading marketDepth={mockMarketDepth} />)
    
    const amountInput = screen.getByLabelText(/amount/i)
    fireEvent.change(amountInput, { target: { value: '1500000' } })

    expect(screen.getByText(/high price impact/i)).toBeInTheDocument()
    expect(screen.getByText(/proceed with caution/i)).toBeInTheDocument()
  })
})
