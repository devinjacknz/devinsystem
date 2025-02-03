from fastapi import FastAPI, WebSocket, WebSocketDisconnect
from fastapi.middleware.cors import CORSMiddleware
from typing import Dict, List
from .prompts import router as prompts_router
from .ai import router as ai_router
from ..services.ai_integration import ai_service
from ..services.ai_analysis import AIAnalysisService

app = FastAPI(title="Trading System API", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

class ConnectionManager:
    def __init__(self):
        self.active_connections: List[WebSocket] = []
        self.ai_service = ai_service
        self.analysis_service = AIAnalysisService()

    async def connect(self, websocket: WebSocket):
        await websocket.accept()
        self.active_connections.append(websocket)

    def disconnect(self, websocket: WebSocket):
        self.active_connections.remove(websocket)

    async def broadcast(self, message: str):
        for connection in self.active_connections:
            await connection.send_text(message)

    async def broadcast_analysis(self, analysis: Dict):
        for connection in self.active_connections:
            await connection.send_json(analysis)

manager = ConnectionManager()

app.include_router(prompts_router, prefix="/api/v1", tags=["prompts"])
app.include_router(ai_router, prefix="/api/v1/ai", tags=["ai"])

@app.get("/")
async def root():
    return {"message": "Trading System API"}

@app.get("/api/health")
async def health_check():
    return {"status": "healthy", "services": ["ai", "websocket", "prompts"]}

@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    await manager.connect(websocket)
    try:
        while True:
            data = await websocket.receive_text()
            await manager.broadcast(data)
    except WebSocketDisconnect:
        manager.disconnect(websocket)

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
