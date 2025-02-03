export interface MarketAnalysis {
  sentiment: 'bullish' | 'bearish' | 'neutral';
  confidence: number;
  signals: TradingSignal[];
  riskMetrics: RiskMetrics;
}

export interface TradingSignal {
  type: 'entry' | 'exit';
  direction: 'long' | 'short';
  price: number;
  timestamp: number;
  strength: number;
  reason: string;
}

export interface RiskMetrics {
  volatility: number;
  sharpeRatio: number;
  maxDrawdown: number;
  currentRisk: number;
}
