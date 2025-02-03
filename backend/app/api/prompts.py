from fastapi import APIRouter, HTTPException
from typing import Dict, Any
from ..models.prompts import QuantitativePrompt, generate_quant_prompt
from ..services.ai import get_ai_analysis

router = APIRouter()

@router.post("/prompts/generate")
async def generate_trading_prompt(data: QuantitativePrompt):
    try:
        prompt = generate_quant_prompt(data)
        analysis = await get_ai_analysis(prompt, data.mode)
        return {
            "prompt": prompt,
            "analysis": analysis,
            "mode": data.mode,
            "symbol": data.symbol
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/prompts/templates/{mode}")
async def get_prompt_template(mode: str):
    if mode not in ["dex", "pump"]:
        raise HTTPException(status_code=400, detail="Invalid trading mode")
    
    from ..models.prompts import DEX_PROMPT_TEMPLATE, PUMP_PROMPT_TEMPLATE
    template = DEX_PROMPT_TEMPLATE if mode == "dex" else PUMP_PROMPT_TEMPLATE
    return {"template": template}
