from enum import Enum
from pydantic import BaseModel, Field
from typing import Optional

class TradingMode(str, Enum):
    DEX = "dex"
    PUMP = "pump"

class AgentStatus(str, Enum):
    ACTIVE = "active"
    PAUSED = "paused"
    STOPPED = "stopped"

class AgentStrategy(BaseModel):
    mode: TradingMode
    risk_tolerance: float = Field(ge=0.0, le=1.0)
    max_trade_size: float = Field(gt=0.0)
    stop_loss_percentage: float = Field(gt=0.0, le=100.0)
    take_profit_percentage: float = Field(gt=0.0, le=1000.0)
    auto_rebalance: bool = False
    rebalance_threshold: Optional[float] = Field(default=None, ge=0.0, le=100.0)
    wallet_address: str = Field(..., description="Solana wallet address for trading")
    network: str = Field(default="mainnet", pattern="^(mainnet|devnet|testnet)$")

class Agent(BaseModel):
    id: str
    name: str
    status: AgentStatus
    strategy: AgentStrategy
    created_at: str
    updated_at: str
    total_trades: int = 0
    successful_trades: int = 0
    current_position: Optional[dict] = None
