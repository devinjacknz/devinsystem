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
        const response = await fetch(`${API_URL}/api/v1/prompts/generate`, {
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
        setAnalysis(result.analysis);
      } catch (error) {
        setError(error instanceof Error ? error.message : 'Failed to fetch analysis');
      } finally {
        setIsLoading(false);
      }
    };

    fetchAnalysis();
  }, [data]);

  return { analysis, error, isLoading };
}
