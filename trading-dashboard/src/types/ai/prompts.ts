export interface QuantitativePrompt {
  mode: 'dex' | 'pump';
  symbol: string;
  timeframe: string;
  metrics: QuantMetrics;
}

export interface QuantMetrics {
  volume: number;
  price: number;
  volatility: number;
  momentum: number;
  liquidity: number;
}

export interface QuantitativeAnalysis {
  entryPoints: {
    optimal: number;
    stopLoss: number;
    takeProfit: number;
  };
  position: {
    size: number;
    riskPercentage: number;
    maxExposure: number;
  };
  risk: {
    volatility: number;
    liquidity: number;
    overall: number;
  };
  signals: {
    volumeProfile: string;
    priceAction: string;
    momentum: string;
  };
  execution: {
    timeframe: string;
    orderType: string;
    slippageTolerance: number;
  };
}

export interface AIPromptResponse {
  prompt: string;
  analysis: QuantitativeAnalysis;
  mode: 'dex' | 'pump';
  symbol: string;
}
