from fastapi import APIRouter, HTTPException, Depends
from typing import Dict, List
from ..models.agent import Agent, AgentStrategy, AgentStatus
from ..services.ai_integration import ai_service
from ..services.agent_service import AgentService
from internal.wallet.solana import SolanaWallet

router = APIRouter()

@router.post("/create")
async def create_agent(agent: Agent, agent_service: AgentService = Depends()):
    try:
        # Validate Solana wallet
        if not agent.strategy.wallet_address:
            raise HTTPException(status_code=400, detail="Solana wallet address is required")
            
        # Initialize wallet
        try:
            wallet = SolanaWallet(agent.strategy.wallet_address)
            if not await wallet.is_valid():
                raise HTTPException(status_code=400, detail="Invalid Solana wallet address")
        except Exception as wallet_error:
            raise HTTPException(status_code=400, detail=f"Wallet validation failed: {str(wallet_error)}")
            
        # Set initial status
        agent.status = AgentStatus.PAUSED
        
        # Create agent
        created_agent = await agent_service.create_agent(agent)
        return created_agent
    except HTTPException as he:
        raise he
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to create agent: {str(e)}")

@router.put("/{agent_id}")
async def update_agent(agent_id: str, agent: Agent, agent_service: AgentService = Depends()):
    try:
        # Validate agent exists
        existing_agent = await agent_service.get_agent(agent_id)
        if not existing_agent:
            raise HTTPException(status_code=404, detail="Agent not found")
            
        # Validate Solana wallet if changed
        if agent.strategy.wallet_address != existing_agent.strategy.wallet_address:
            try:
                wallet = SolanaWallet(agent.strategy.wallet_address)
                if not await wallet.is_valid():
                    raise HTTPException(status_code=400, detail="Invalid Solana wallet address")
            except Exception as wallet_error:
                raise HTTPException(status_code=400, detail=f"Wallet validation failed: {str(wallet_error)}")
        
        # Update agent
        updated_agent = await agent_service.update_agent(agent_id, agent)
        return updated_agent
    except HTTPException as he:
        raise he
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to update agent: {str(e)}")

@router.get("/{agent_id}/status")
async def get_agent_status(agent_id: str, agent_service: AgentService = Depends()):
    try:
        # Get agent
        agent = await agent_service.get_agent(agent_id)
        if not agent:
            raise HTTPException(status_code=404, detail="Agent not found")
            
        # Get agent status
        status = await agent_service.get_agent_status(agent_id)
        return status
    except HTTPException as he:
        raise he
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to get agent status: {str(e)}")

@router.get("/")
async def list_agents(agent_service: AgentService = Depends()):
    try:
        agents = await agent_service.list_agents()
        return agents
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to list agents: {str(e)}")
