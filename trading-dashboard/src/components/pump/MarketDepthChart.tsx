import React from 'react'
import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip } from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card'
import { MarketDepth } from '../../types/trading'

interface MarketDepthChartProps {
  data: MarketDepth
  symbol: string
}

export function MarketDepthChart({ data, symbol }: MarketDepthChartProps) {
  const chartData = [
    ...data.bids.map(([price, amount]) => ({
      price,
      bids: amount,
      asks: null
    })),
    ...data.asks.map(([price, amount]) => ({
      price,
      bids: null,
      asks: amount
    }))
  ].sort((a, b) => a.price - b.price)

  return (
    <Card>
      <CardHeader>
        <CardTitle>Market Depth - {symbol}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-[400px]">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData}>
              <XAxis 
                dataKey="price"
                tickFormatter={(value) => `$${value.toFixed(4)}`}
              />
              <YAxis />
              <Tooltip
                formatter={(value: number) => value.toFixed(4)}
                labelFormatter={(label) => `Price: $${label.toFixed(4)}`}
              />
              <Area
                type="monotone"
                dataKey="bids"
                stackId="1"
                stroke="#22c55e"
                fill="#22c55e"
                fillOpacity={0.3}
              />
              <Area
                type="monotone"
                dataKey="asks"
                stackId="2"
                stroke="#ef4444"
                fill="#ef4444"
                fillOpacity={0.3}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}
