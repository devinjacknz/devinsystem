import { useState, useEffect } from 'react'
import { MarketAnalysis } from '../../types/ai'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export function useAIAnalysis(symbol: string) {
  const [analysis, setAnalysis] = useState<MarketAnalysis | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    const fetchAnalysis = async () => {
      setIsLoading(true)
      setError(null)
      try {
        const response = await fetch(`${API_URL}/ai/analysis/${symbol}`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
          }
        })

        if (!response.ok) {
          throw new Error('Failed to fetch AI analysis')
        }

        const data: MarketAnalysis = await response.json()
        setAnalysis(data)
      } catch (error) {
        setError(error instanceof Error ? error.message : 'Failed to fetch AI analysis')
      } finally {
        setIsLoading(false)
      }
    }

    const interval = setInterval(fetchAnalysis, 30000) // Update every 30 seconds
    fetchAnalysis() // Initial fetch

    return () => clearInterval(interval)
  }, [symbol])

  return { analysis, error, isLoading }
}
