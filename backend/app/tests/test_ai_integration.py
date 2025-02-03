import pytest
from fastapi.testclient import TestClient
from ..api.main import app
from ..models.prompts import TradingMode, QuantitativePrompt, QuantMetrics

client = TestClient(app)

@pytest.fixture
def sample_dex_prompt():
    return {
        "mode": "dex",
        "symbol": "SOL/USD",
        "timeframe": "4h",
        "metrics": {
            "volume": 1000000,
            "price": 100.50,
            "volatility": 0.15,
            "momentum": 0.8,
            "liquidity": 500000
        }
    }

@pytest.fixture
def sample_pump_prompt():
    return {
        "mode": "pump",
        "symbol": "PEPE/USD",
        "timeframe": "1h",
        "metrics": {
            "volume": 500000,
            "price": 0.0001,
            "volatility": 0.25,
            "momentum": 0.9,
            "liquidity": 100000
        }
    }

def test_analyze_dex_trading():
    response = client.post("/api/v1/ai/analyze", json=sample_dex_prompt())
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "success"
    assert data["mode"] == "dex"
    assert "analysis" in data
    analysis = data["analysis"]
    assert "entryPoints" in analysis
    assert "risk" in analysis
    assert "signals" in analysis
    assert "execution" in analysis

def test_analyze_pump_trading():
    response = client.post("/api/v1/ai/analyze", json=sample_pump_prompt())
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "success"
    assert data["mode"] == "pump"
    assert "analysis" in data
    analysis = data["analysis"]
    assert "entryPoints" in analysis
    assert "risk" in analysis
    assert "signals" in analysis
    assert "execution" in analysis

def test_invalid_trading_mode():
    invalid_prompt = sample_dex_prompt()
    invalid_prompt["mode"] = "invalid"
    response = client.post("/api/v1/ai/analyze", json=invalid_prompt)
    assert response.status_code == 422

def test_missing_metrics():
    invalid_prompt = sample_dex_prompt()
    del invalid_prompt["metrics"]["volume"]
    response = client.post("/api/v1/ai/analyze", json=invalid_prompt)
    assert response.status_code == 422

def test_health_check():
    response = client.get("/api/v1/ai/health")
    assert response.status_code == 200
    assert response.json()["status"] == "healthy"
