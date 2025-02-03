import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { AIPrompt } from '../../types/ai';
import { useQuantitativeAnalysis } from '../../hooks/ai/useQuantitativeAnalysis';

interface AIPromptsProps {
  mode: 'dex' | 'pump';
  symbol: string;
}

export function AIPrompts({ mode, symbol }: AIPromptsProps) {
  const { analysis, error, isLoading } = useQuantitativeAnalysis({
    mode,
    symbol,
    timeframe: '4h',
    metrics: {
      volume: 1000000,
      price: 100.50,
      volatility: 0.15,
      momentum: 0.8,
      liquidity: 500000
    }
  });

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>AI Analysis Loading...</CardTitle>
        </CardHeader>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>AI Analysis Error</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-red-500">{error}</p>
        </CardContent>
      </Card>
    );
  }

  if (!analysis) {
    return null;
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Quantitative Trading Analysis</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-medium mb-2">Entry Points</h3>
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <p className="text-xs text-muted-foreground">Optimal</p>
                  <p className="font-medium">${analysis.entryPoints.optimal}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Stop Loss</p>
                  <p className="font-medium">${analysis.entryPoints.stopLoss}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Take Profit</p>
                  <p className="font-medium">${analysis.entryPoints.takeProfit}</p>
                </div>
              </div>
            </div>

            <div>
              <h3 className="text-sm font-medium mb-2">Risk Assessment</h3>
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <p className="text-xs text-muted-foreground">Volatility</p>
                  <p className="font-medium">{analysis.risk.volatility}/10</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Liquidity</p>
                  <p className="font-medium">{analysis.risk.liquidity}/10</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Overall</p>
                  <p className="font-medium">{analysis.risk.overall}/10</p>
                </div>
              </div>
            </div>

            <div>
              <h3 className="text-sm font-medium mb-2">Technical Signals</h3>
              <div className="space-y-2">
                <div className="bg-blue-50 p-2 rounded">
                  <p className="text-xs text-blue-600">Volume Profile</p>
                  <p className="text-sm">{analysis.signals.volumeProfile}</p>
                </div>
                <div className="bg-blue-50 p-2 rounded">
                  <p className="text-xs text-blue-600">Price Action</p>
                  <p className="text-sm">{analysis.signals.priceAction}</p>
                </div>
                <div className="bg-blue-50 p-2 rounded">
                  <p className="text-xs text-blue-600">Momentum</p>
                  <p className="text-sm">{analysis.signals.momentum}</p>
                </div>
              </div>
            </div>

            <div>
              <h3 className="text-sm font-medium mb-2">Execution Parameters</h3>
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <p className="text-xs text-muted-foreground">Timeframe</p>
                  <p className="font-medium">{analysis.execution.timeframe}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Order Type</p>
                  <p className="font-medium">{analysis.execution.orderType}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Slippage</p>
                  <p className="font-medium">{analysis.execution.slippageTolerance}%</p>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
