import React from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card'
import { MarketAnalysis, TradingSignal } from '../../types/ai'

interface AIInsightsProps {
  analysis: MarketAnalysis;
}

export function AIInsights({ analysis }: AIInsightsProps) {
  const { sentiment, confidence, signals, riskMetrics } = analysis

  const getSentimentColor = (sentiment: MarketAnalysis['sentiment']) => {
    switch (sentiment) {
      case 'bullish':
        return 'text-green-500';
      case 'bearish':
        return 'text-red-500';
      default:
        return 'text-yellow-500';
    }
  }

  const formatSignalTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleTimeString()
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>AI Market Analysis</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <h3 className="text-sm font-medium">Market Sentiment</h3>
              <p className={`font-medium capitalize ${getSentimentColor(sentiment)}`}>
                {sentiment} ({(confidence * 100).toFixed(1)}%)
              </p>
            </div>
            <div className="space-y-2">
              <h3 className="text-sm font-medium">Risk Level</h3>
              <p className="font-medium">
                {(riskMetrics.currentRisk * 100).toFixed(1)}%
              </p>
            </div>
          </div>

          <div className="space-y-2">
            <h3 className="text-sm font-medium">Risk Metrics</h3>
            <div className="grid grid-cols-3 gap-4">
              <div>
                <p className="text-xs text-muted-foreground">Volatility</p>
                <p className="font-medium">{(riskMetrics.volatility * 100).toFixed(1)}%</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Sharpe Ratio</p>
                <p className="font-medium">{riskMetrics.sharpeRatio.toFixed(2)}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Max Drawdown</p>
                <p className="font-medium">{(riskMetrics.maxDrawdown * 100).toFixed(1)}%</p>
              </div>
            </div>
          </div>

          <div className="space-y-2">
            <h3 className="text-sm font-medium">Trading Signals</h3>
            <div className="space-y-2">
              {signals.map((signal: TradingSignal, index: number) => (
                <div
                  key={index}
                  className={`p-2 rounded ${
                    signal.type === 'entry'
                      ? 'bg-green-100 dark:bg-green-900/20'
                      : 'bg-red-100 dark:bg-red-900/20'
                  }`}
                >
                  <div className="flex justify-between items-center">
                    <span className="font-medium capitalize">
                      {signal.type} {signal.direction}
                    </span>
                    <span className="text-sm text-muted-foreground">
                      {formatSignalTime(signal.timestamp)}
                    </span>
                  </div>
                  <div className="mt-1">
                    <p className="text-sm">
                      Price: ${signal.price.toFixed(4)} (Strength: {(signal.strength * 100).toFixed(1)}%)
                    </p>
                    <p className="text-sm text-muted-foreground">{signal.reason}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
