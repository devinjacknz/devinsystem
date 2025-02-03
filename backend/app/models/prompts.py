from enum import Enum
from pydantic import BaseModel
from typing import Dict, Optional

class TradingMode(str, Enum):
    DEX = "dex"
    PUMP = "pump"

class QuantMetrics(BaseModel):
    volume: float
    price: float
    volatility: float
    momentum: float
    liquidity: float

class QuantitativePrompt(BaseModel):
    mode: TradingMode
    symbol: str
    timeframe: str
    metrics: QuantMetrics

DEX_PROMPT_TEMPLATE = """Analyze DEX trading opportunity for {symbol}:
Current market conditions:
- Price: ${price}
- 24h Volume: ${volume}
- Volatility Index: {volatility}
- Momentum Score: {momentum}
- Liquidity Depth: ${liquidity}

Based on quantitative analysis:
1. Entry/Exit Points:
   - Optimal entry range
   - Stop-loss levels
   - Take-profit targets

2. Position Sizing:
   - Recommended position size
   - Risk per trade percentage
   - Maximum exposure limit

3. Risk Assessment:
   - Market volatility risk (1-10)
   - Liquidity risk (1-10)
   - Overall trade risk score

4. Technical Signals:
   - Volume profile analysis
   - Price action patterns
   - Momentum indicators

5. Trade Parameters:
   - Suggested timeframe
   - Order types to use
   - Slippage tolerance"""

PUMP_PROMPT_TEMPLATE = """Analyze Pump.fun trading opportunity for {symbol}:
Market metrics:
- Current Price: ${price}
- Trading Volume (24h): ${volume}
- Price Volatility: {volatility}
- Momentum Rating: {momentum}
- Available Liquidity: ${liquidity}

Quantitative Analysis:
1. Momentum Analysis:
   - Trend strength measurement
   - Volume/price correlation
   - Acceleration factors

2. Entry Strategy:
   - Volume profile based levels
   - Price action triggers
   - Momentum confirmation signals

3. Risk Management:
   - Position size calculation
   - Stop-loss placement
   - Risk/reward ratio

4. Market Impact:
   - Liquidity utilization
   - Slippage estimation
   - Order book depth analysis

5. Execution Plan:
   - Order type selection
   - Entry timing optimization
   - Exit strategy parameters"""

def generate_quant_prompt(data: QuantitativePrompt) -> str:
    template = DEX_PROMPT_TEMPLATE if data.mode == TradingMode.DEX else PUMP_PROMPT_TEMPLATE
    return template.format(
        symbol=data.symbol,
        price=data.metrics.price,
        volume=data.metrics.volume,
        volatility=data.metrics.volatility,
        momentum=data.metrics.momentum,
        liquidity=data.metrics.liquidity
    )
