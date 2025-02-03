export interface PriceData {
  timestamp: number
  price: number
}

export interface Position {
  symbol: string
  size: number
  entryPrice: number
  currentPrice: number
  pnl: number
}

export type TimeRange = '1H' | '24H' | '7D' | '30D'
export const TIME_RANGES: TimeRange[] = ['1H', '24H', '7D', '30D']
export type TradingPair = 'SOL/USD' | 'SOL/USDC' | 'BONK/USD'
export type ActiveSymbol = TradingPair | 'all'

export const TRADING_PAIRS = ['SOL/USD', 'SOL/USDC', 'BONK/USD'] as const
export type TradingPairTuple = typeof TRADING_PAIRS
export type TradingPairFromTuple = TradingPairTuple[number]
