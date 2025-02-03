import { render, screen, fireEvent } from '@testing-library/react'
import { AIInsights } from '../../ai/AIInsights'
import { useAIAnalysis } from '../../../hooks/ai/useAIAnalysis'

jest.mock('../../../hooks/ai/useAIAnalysis')

describe('AIInsights', () => {
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

  const defaultProps = {
    marketData: {
      symbol: 'SOL/USDC',
      price: 100.50,
      volume: 1000000,
      change24h: 5.5
    }
  }

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useAIAnalysis as jest.Mock).mockReturnValue({
      analysis: mockAnalysis,
      isLoading: false,
      error: null
    })
  })

  it('renders AI analysis insights correctly', () => {
    render(<AIInsights {...defaultProps} />)
    
    expect(screen.getByText(/market analysis/i)).toBeInTheDocument()
    expect(screen.getByText(/sentiment: bullish/i)).toBeInTheDocument()
    expect(screen.getByText(/risk score: 0.7/i)).toBeInTheDocument()
    expect(screen.getByText(/price target: \$0.0000150/i)).toBeInTheDocument()
    expect(screen.getByText(/confidence: 85%/i)).toBeInTheDocument()
  })

  it('displays trading signals', () => {
    render(<AIInsights {...defaultProps} />)
    
    expect(screen.getByText(/momentum/i)).toBeInTheDocument()
    expect(screen.getByText(/strong upward trend/i)).toBeInTheDocument()
    expect(screen.getByText(/volume/i)).toBeInTheDocument()
    expect(screen.getByText(/above average volume/i)).toBeInTheDocument()
  })

  it('shows loading state', () => {
    ;(useAIAnalysis as jest.Mock).mockReturnValue({
      analysis: null,
      isLoading: true,
      error: null
    })

    render(<AIInsights {...defaultProps} />)
    
    expect(screen.getByText(/loading analysis/i)).toBeInTheDocument()
  })

  it('handles error state', () => {
    ;(useAIAnalysis as jest.Mock).mockReturnValue({
      analysis: null,
      isLoading: false,
      error: 'Failed to fetch AI analysis'
    })

    render(<AIInsights {...defaultProps} />)
    
    expect(screen.getByText(/failed to fetch ai analysis/i)).toBeInTheDocument()
  })

  it('displays risk indicators', () => {
    render(<AIInsights {...defaultProps} />)
    
    const riskScore = screen.getByTestId('risk-indicator')
    expect(riskScore).toHaveStyle({ backgroundColor: expect.any(String) })
    expect(riskScore).toHaveTextContent('0.7')
  })

  it('shows model confidence level', () => {
    render(<AIInsights {...defaultProps} />)
    
    const confidenceBar = screen.getByTestId('confidence-bar')
    expect(confidenceBar).toHaveStyle({ width: '85%' })
  })
})
