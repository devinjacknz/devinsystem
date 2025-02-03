import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { AIPrompts } from '../../ai/AIPrompts';

const mockAnalysis = {
  entryPoints: {
    optimal: 100.50,
    stopLoss: 95.00,
    takeProfit: 110.00
  },
  position: {
    size: 1000,
    riskPercentage: 2.5,
    maxExposure: 5000
  },
  risk: {
    volatility: 7,
    liquidity: 8,
    overall: 7.5
  },
  signals: {
    volumeProfile: "Increasing volume trend",
    priceAction: "Bullish breakout pattern",
    momentum: "Strong upward momentum"
  },
  execution: {
    timeframe: "4h",
    orderType: "Limit",
    slippageTolerance: 0.5
  }
};

vi.mock('../../../hooks/ai/useQuantitativeAnalysis', () => ({
  useQuantitativeAnalysis: vi.fn(() => ({
    analysis: mockAnalysis,
    error: null,
    isLoading: false
  }))
}));

describe('AIPrompts', () => {
  it('renders loading state', () => {
    vi.mocked(useQuantitativeAnalysis).mockReturnValueOnce({
      analysis: null,
      error: null,
      isLoading: true
    });

    render(<AIPrompts mode="dex" symbol="SOL/USD" />);
    expect(screen.getByText('AI Analysis Loading...')).toBeInTheDocument();
  });

  it('renders error state', () => {
    const errorMessage = 'Failed to fetch analysis';
    vi.mocked(useQuantitativeAnalysis).mockReturnValueOnce({
      analysis: null,
      error: errorMessage,
      isLoading: false
    });

    render(<AIPrompts mode="dex" symbol="SOL/USD" />);
    expect(screen.getByText('AI Analysis Error')).toBeInTheDocument();
    expect(screen.getByText(errorMessage)).toBeInTheDocument();
  });

  it('renders analysis data correctly', () => {
    render(<AIPrompts mode="dex" symbol="SOL/USD" />);
    
    expect(screen.getByText('Quantitative Trading Analysis')).toBeInTheDocument();
    expect(screen.getByText('$100.5')).toBeInTheDocument();
    expect(screen.getByText('$95')).toBeInTheDocument();
    expect(screen.getByText('$110')).toBeInTheDocument();
    expect(screen.getByText('7/10')).toBeInTheDocument();
    expect(screen.getByText('8/10')).toBeInTheDocument();
    expect(screen.getByText('7.5/10')).toBeInTheDocument();
  });

  it('displays technical signals with correct formatting', () => {
    render(<AIPrompts mode="dex" symbol="SOL/USD" />);
    
    expect(screen.getByText('Increasing volume trend')).toBeInTheDocument();
    expect(screen.getByText('Bullish breakout pattern')).toBeInTheDocument();
    expect(screen.getByText('Strong upward momentum')).toBeInTheDocument();
  });

  it('shows execution parameters correctly', () => {
    render(<AIPrompts mode="dex" symbol="SOL/USD" />);
    
    expect(screen.getByText('4h')).toBeInTheDocument();
    expect(screen.getByText('Limit')).toBeInTheDocument();
    expect(screen.getByText('0.5%')).toBeInTheDocument();
  });
});
