import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { SwapInterface } from '../../dex/SwapInterface'
import { useWallet } from '../../../hooks/wallet/useWallet'
import { useAIAnalysis } from '../../../hooks/ai/useAIAnalysis'

jest.mock('../../../hooks/wallet/useWallet')
jest.mock('../../../hooks/ai/useAIAnalysis')

describe('SwapInterface', () => {
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
    confidence: 0.85
  }

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useWallet as jest.Mock).mockReturnValue(mockWallet)
    ;(useAIAnalysis as jest.Mock).mockReturnValue({ analysis: mockAIAnalysis, isLoading: false, error: null })
  })

  it('renders swap interface correctly', () => {
    render(<SwapInterface />)
    
    expect(screen.getByLabelText(/from token/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/to token/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/amount/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /swap/i })).toBeInTheDocument()
  })

  it('handles token selection', () => {
    render(<SwapInterface />)
    
    const fromTokenSelect = screen.getByLabelText(/from token/i)
    const toTokenSelect = screen.getByLabelText(/to token/i)

    fireEvent.change(fromTokenSelect, { target: { value: 'SOL' } })
    fireEvent.change(toTokenSelect, { target: { value: 'USDC' } })

    expect(fromTokenSelect).toHaveValue('SOL')
    expect(toTokenSelect).toHaveValue('USDC')
  })

  it('executes swap transaction', async () => {
    render(<SwapInterface />)
    
    const fromTokenSelect = screen.getByLabelText(/from token/i)
    const toTokenSelect = screen.getByLabelText(/to token/i)
    const amountInput = screen.getByLabelText(/amount/i)

    fireEvent.change(fromTokenSelect, { target: { value: 'SOL' } })
    fireEvent.change(toTokenSelect, { target: { value: 'USDC' } })
    fireEvent.change(amountInput, { target: { value: '10' } })

    const swapButton = screen.getByRole('button', { name: /swap/i })
    fireEvent.click(swapButton)

    await waitFor(() => {
      expect(mockWallet.executeTrade).toHaveBeenCalledWith({
        type: 'swap',
        fromToken: 'SOL',
        toToken: 'USDC',
        amount: 10
      })
    })
  })

  it('displays AI analysis insights', () => {
    render(<SwapInterface />)
    
    expect(screen.getByText(/sentiment: bullish/i)).toBeInTheDocument()
    expect(screen.getByText(/risk score: 0.7/i)).toBeInTheDocument()
    expect(screen.getByText(/confidence: 85%/i)).toBeInTheDocument()
  })

  it('validates input amount against wallet balance', () => {
    render(<SwapInterface />)
    
    const amountInput = screen.getByLabelText(/amount/i)
    fireEvent.change(amountInput, { target: { value: '2000' } })

    expect(screen.getByText(/insufficient balance/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /swap/i })).toBeDisabled()
  })

  it('handles loading states', () => {
    ;(useAIAnalysis as jest.Mock).mockReturnValue({ analysis: null, isLoading: true, error: null })
    
    render(<SwapInterface />)
    
    expect(screen.getByText(/loading analysis/i)).toBeInTheDocument()
  })

  it('displays error states', () => {
    ;(useWallet as jest.Mock).mockReturnValue({
      ...mockWallet,
      error: 'Failed to connect wallet'
    })
    
    render(<SwapInterface />)
    
    expect(screen.getByText(/failed to connect wallet/i)).toBeInTheDocument()
  })

  it('updates price impact warning based on trade size', () => {
    render(<SwapInterface />)
    
    const amountInput = screen.getByLabelText(/amount/i)
    fireEvent.change(amountInput, { target: { value: '500' } })

    expect(screen.getByText(/high price impact/i)).toBeInTheDocument()
    expect(screen.getByText(/proceed with caution/i)).toBeInTheDocument()
  })
})
