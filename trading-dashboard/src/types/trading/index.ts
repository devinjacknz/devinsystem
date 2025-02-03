export interface MarketDepth {
  bids: [number, number][] // [price, amount]
  asks: [number, number][] // [price, amount]
}

export interface TokenInfo {
  symbol: string
  name: string
  price: number
  volume24h: number
  change24h: number
}

export interface TradeAction {
  type: 'buy' | 'sell'
  token: string
  amount: number
  price: number
}
