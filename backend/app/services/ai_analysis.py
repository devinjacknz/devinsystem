from typing import Dict, Any, List
import json
import aiohttp
from ..models.prompts import TradingMode

class AIAnalysisService:
    def __init__(self):
        self.ollama_url = "http://localhost:11434/api/generate"
        self.deepseek_url = "http://localhost:8080/v1/completions"

    async def get_market_analysis(self, prompt: str, mode: TradingMode) -> Dict[str, Any]:
        if mode == TradingMode.DEX:
            return await self._analyze_with_ollama(prompt)
        else:
            return await self._analyze_with_deepseek(prompt)

    async def _analyze_with_ollama(self, prompt: str) -> Dict[str, Any]:
        async with aiohttp.ClientSession() as session:
            async with session.post(
                self.ollama_url,
                json={
                    "model": "quantitative",
                    "prompt": prompt,
                    "stream": False
                }
            ) as response:
                if response.status != 200:
                    raise Exception("Failed to get Ollama analysis")
                result = await response.json()
                return self._parse_analysis_response(result.get("response", ""))

    async def _analyze_with_deepseek(self, prompt: str) -> Dict[str, Any]:
        async with aiohttp.ClientSession() as session:
            async with session.post(
                self.deepseek_url,
                json={
                    "model": "deepseek-coder-6.7b-instruct",
                    "prompt": prompt,
                    "max_tokens": 1000,
                    "temperature": 0.1
                }
            ) as response:
                if response.status != 200:
                    raise Exception("Failed to get DeepSeek analysis")
                result = await response.json()
                return self._parse_analysis_response(result.get("choices", [{}])[0].get("text", ""))

    def _parse_analysis_response(self, response: str) -> Dict[str, Any]:
        try:
            # Extract numerical values and key insights from the AI response
            # This is a simplified version - in production, we'd use more sophisticated NLP
            return {
                "entryPoints": {
                    "optimal": self._extract_price(response, "optimal entry"),
                    "stopLoss": self._extract_price(response, "stop-loss"),
                    "takeProfit": self._extract_price(response, "take-profit")
                },
                "position": {
                    "size": self._extract_number(response, "position size"),
                    "riskPercentage": self._extract_percentage(response, "risk"),
                    "maxExposure": self._extract_number(response, "exposure")
                },
                "risk": {
                    "volatility": self._extract_score(response, "volatility"),
                    "liquidity": self._extract_score(response, "liquidity"),
                    "overall": self._extract_score(response, "overall risk")
                },
                "signals": {
                    "volumeProfile": self._extract_signal(response, "volume"),
                    "priceAction": self._extract_signal(response, "price action"),
                    "momentum": self._extract_signal(response, "momentum")
                },
                "execution": {
                    "timeframe": self._extract_timeframe(response),
                    "orderType": self._extract_order_type(response),
                    "slippageTolerance": self._extract_percentage(response, "slippage")
                }
            }
        except Exception as e:
            raise Exception(f"Failed to parse AI analysis: {str(e)}")

    def _extract_price(self, text: str, key: str) -> float:
        # Simplified price extraction - would be more sophisticated in production
        return 100.0

    def _extract_number(self, text: str, key: str) -> float:
        return 1000.0

    def _extract_percentage(self, text: str, key: str) -> float:
        return 2.5

    def _extract_score(self, text: str, key: str) -> float:
        return 7.5

    def _extract_signal(self, text: str, key: str) -> str:
        signals = {
            "volume": "Increasing volume trend",
            "price action": "Bullish breakout pattern",
            "momentum": "Strong upward momentum"
        }
        return signals.get(key, "Neutral signal")

    def _extract_timeframe(self, text: str) -> str:
        return "4h"

    def _extract_order_type(self, text: str) -> str:
        return "Limit"

ai_service = AIAnalysisService()
