import { useState, useEffect } from 'react';
import { QuantitativePrompt, QuantitativeAnalysis } from '../../types/ai/prompts';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8000';

export function useQuantitativeAnalysis(data: QuantitativePrompt) {
  const [analysis, setAnalysis] = useState<QuantitativeAnalysis | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    const fetchAnalysis = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const response = await fetch(`${API_URL}/api/v1/ai/analyze`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(data),
        });

        if (!response.ok) {
          throw new Error('Failed to fetch analysis');
        }

        const result = await response.json();
        if (result.status === 'success') {
          setAnalysis(result.analysis);
        } else {
          throw new Error(result.detail || 'Failed to analyze trading opportunity');
        }
      } catch (error) {
        setError(error instanceof Error ? error.message : 'Failed to fetch analysis');
      } finally {
        setIsLoading(false);
      }
    };

    if (data.metrics.price > 0 && data.metrics.volume > 0) {
      fetchAnalysis();
    }
  }, [data]);

  const refreshAnalysis = () => {
    if (data.metrics.price > 0 && data.metrics.volume > 0) {
      setAnalysis(null);
      setError(null);
    }
  };

  return { analysis, error, isLoading, refreshAnalysis };
}
