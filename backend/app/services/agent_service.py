from typing import Dict, List, Optional
from ..models.agent import Agent, AgentStatus
from internal.wallet.solana import SolanaWallet
import uuid
from datetime import datetime

class AgentService:
    def __init__(self):
        self._agents: Dict[str, Agent] = {}
        
    async def create_agent(self, agent: Agent) -> Agent:
        if not agent.id:
            agent.id = str(uuid.uuid4())
        
        now = datetime.utcnow().isoformat()
        agent.created_at = now
        agent.updated_at = now
        
        self._agents[agent.id] = agent
        return agent
        
    async def update_agent(self, agent_id: str, agent: Agent) -> Agent:
        if agent_id not in self._agents:
            raise ValueError("Agent not found")
            
        agent.updated_at = datetime.utcnow().isoformat()
        self._agents[agent_id] = agent
        return agent
        
    async def get_agent(self, agent_id: str) -> Optional[Agent]:
        return self._agents.get(agent_id)
        
    async def get_agent_status(self, agent_id: str) -> dict:
        agent = self._agents.get(agent_id)
        if not agent:
            raise ValueError("Agent not found")
            
        return {
            "status": agent.status,
            "total_trades": agent.total_trades,
            "successful_trades": agent.successful_trades,
            "current_position": agent.current_position,
            "last_updated": agent.updated_at
        }
        
    async def list_agents(self) -> List[Agent]:
        return list(self._agents.values())
