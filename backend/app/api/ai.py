from fastapi import APIRouter, HTTPException
from typing import Dict, Any
from ..models.prompts import QuantitativePrompt
from ..services.ai_integration import ai_service

router = APIRouter()

@router.post("/analyze")
async def analyze_trading_opportunity(prompt: QuantitativePrompt) -> Dict[str, Any]:
    try:
        analysis = await ai_service.analyze_trading_opportunity(prompt)
        return {
            "status": "success",
            "analysis": analysis,
            "mode": prompt.mode,
            "symbol": prompt.symbol
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/health")
async def health_check() -> Dict[str, str]:
    return {"status": "healthy"}
