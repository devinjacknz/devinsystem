from typing import Dict, Any, List
import aiohttp
from ..models.prompts import TradingMode, QuantitativePrompt

class AIIntegrationService:
    def __init__(self):
        self.ollama_url = "http://localhost:11434/api/generate"
        self.deepseek_url = "http://localhost:8080/v1/completions"

    async def analyze_trading_opportunity(self, prompt: QuantitativePrompt) -> Dict[str, Any]:
        if prompt.mode == TradingMode.DEX:
            return await self._analyze_dex_opportunity(prompt)
        else:
            return await self._analyze_pump_opportunity(prompt)

    async def _analyze_dex_opportunity(self, prompt: QuantitativePrompt) -> Dict[str, Any]:
        analysis_prompt = f"""
        Analyze DEX trading opportunity for {prompt.symbol}:
        - Price: ${prompt.metrics.price}
        - Volume: ${prompt.metrics.volume}
        - Volatility: {prompt.metrics.volatility}
        - Momentum: {prompt.metrics.momentum}
        - Liquidity: ${prompt.metrics.liquidity}
        """
        
        async with aiohttp.ClientSession() as session:
            async with session.post(
                self.ollama_url,
                json={
                    "model": "quantitative",
                    "prompt": analysis_prompt,
                    "stream": False
                }
            ) as response:
                if response.status != 200:
                    raise Exception("Failed to get Ollama analysis")
                result = await response.json()
                return self._parse_dex_analysis(result.get("response", ""))

    async def _analyze_pump_opportunity(self, prompt: QuantitativePrompt) -> Dict[str, Any]:
        analysis_prompt = f"""
        Analyze Pump.fun trading opportunity for {prompt.symbol}:
        - Price: ${prompt.metrics.price}
        - Volume: ${prompt.metrics.volume}
        - Volatility: {prompt.metrics.volatility}
        - Momentum: {prompt.metrics.momentum}
        - Liquidity: ${prompt.metrics.liquidity}
        """
        
        async with aiohttp.ClientSession() as session:
            async with session.post(
                self.deepseek_url,
                json={
                    "model": "deepseek-coder-6.7b-instruct",
                    "prompt": analysis_prompt,
                    "max_tokens": 1000,
                    "temperature": 0.1
                }
            ) as response:
                if response.status != 200:
                    raise Exception("Failed to get DeepSeek analysis")
                result = await response.json()
                return self._parse_pump_analysis(result.get("choices", [{}])[0].get("text", ""))

    def _parse_dex_analysis(self, response: str) -> Dict[str, Any]:
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
                "overall": self._calculate_overall_risk_score(response)
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

    def _parse_pump_analysis(self, response: str) -> Dict[str, Any]:
        return {
            "entryPoints": {
                "optimal": self._extract_price(response, "optimal entry"),
                "stopLoss": self._extract_price(response, "stop-loss"),
                "takeProfit": self._extract_price(response, "take-profit")
            },
            "position": {
                "size": self._calculate_position_size(response),
                "riskPercentage": self._extract_percentage(response, "risk"),
                "maxExposure": self._extract_number(response, "exposure")
            },
            "risk": {
                "volatility": self._extract_score(response, "volatility"),
                "liquidity": self._extract_score(response, "liquidity"),
                "overall": self._calculate_overall_risk_score(response)
            },
            "signals": {
                "volumeProfile": self._extract_signal(response, "volume"),
                "priceAction": self._extract_signal(response, "price action"),
                "momentum": self._extract_signal(response, "momentum")
            },
            "execution": {
                "timeframe": self._extract_timeframe(response),
                "orderType": self._determine_order_type(response),
                "slippageTolerance": self._calculate_slippage_tolerance(response)
            }
        }

    def _extract_price(self, text: str, key: str) -> float:
        try:
            import re
            pattern = f"{key}.*?(\d+\.?\d*)"
            match = re.search(pattern, text.lower())
            return float(match.group(1)) if match else 0.0
        except:
            return 0.0

    def _extract_number(self, text: str, key: str) -> float:
        try:
            import re
            pattern = f"{key}.*?(\d+\.?\d*)"
            match = re.search(pattern, text.lower())
            return float(match.group(1)) if match else 0.0
        except:
            return 0.0

    def _extract_percentage(self, text: str, key: str) -> float:
        try:
            import re
            pattern = f"{key}.*?(\d+\.?\d*)%?"
            match = re.search(pattern, text.lower())
            return float(match.group(1)) if match else 0.0
        except:
            return 0.0

    def _extract_score(self, text: str, key: str) -> float:
        try:
            import re
            pattern = f"{key}.*?(\d+\.?\d*)/10"
            match = re.search(pattern, text.lower())
            return float(match.group(1)) if match else 5.0
        except:
            return 5.0

    def _extract_signal(self, text: str, key: str) -> str:
        try:
            import re
            pattern = f"{key}.*?:\s*(.*?)(?:\n|$)"
            match = re.search(pattern, text.lower())
            return match.group(1).strip().capitalize() if match else "Neutral"
        except:
            return "Neutral"

    def _extract_timeframe(self, text: str) -> str:
        timeframes = ["1m", "5m", "15m", "30m", "1h", "4h", "1d"]
        for tf in timeframes:
            if tf in text.lower():
                return tf
        return "4h"

    def _extract_order_type(self, text: str) -> str:
        order_types = ["market", "limit", "stop-limit"]
        for ot in order_types:
            if ot in text.lower():
                return ot.capitalize()
        return "Market"

    def _calculate_overall_risk_score(self, text: str) -> float:
        volatility = self._extract_score(text, "volatility")
        liquidity = self._extract_score(text, "liquidity")
        return round((volatility + liquidity) / 2, 1)

    def _calculate_position_size(self, text: str) -> float:
        base = self._extract_number(text, "position size")
        return max(100.0, base)

    def _determine_order_type(self, text: str) -> str:
        if "high volatility" in text.lower() or "rapid price movement" in text.lower():
            return "Market"
        return "Limit"

    def _calculate_slippage_tolerance(self, text: str) -> float:
        base = self._extract_percentage(text, "slippage")
        volatility = self._extract_score(text, "volatility")
        return min(max(base, volatility / 2), 5.0)

ai_service = AIIntegrationService()
