import { PriceData } from '../types'

export const generateMockPriceHistory = (timeRange: string): PriceData[] => {
  const now = Date.now()
  const points = 50
  let duration = 86400000 // 24H default

  switch (timeRange) {
    case '1H':
      duration = 3600000
      break
    case '7D':
      duration = 604800000
      break
    case '30D':
      duration = 2592000000
      break
  }

  const interval = duration / points
  return Array.from({ length: points + 1 }, (_, i) => ({
    timestamp: now - (i * interval),
    price: 100 + (Math.random() - 0.5) * 10
  })).reverse()
}
