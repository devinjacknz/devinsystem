from fastapi import FastAPI
from .trading import router as trading_router
from .ai import router as ai_router
from .agents import router as agent_router

app = FastAPI()

app.include_router(trading_router, prefix="/api/v1/trading", tags=["trading"])
app.include_router(ai_router, prefix="/api/v1/ai", tags=["ai"])
app.include_router(agent_router, prefix="/api/v1/agents", tags=["agents"])

@app.get("/health")
async def health_check():
    return {"status": "healthy"}
