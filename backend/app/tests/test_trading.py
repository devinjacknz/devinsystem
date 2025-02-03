import pytest
from ..models.agent import TradingMode

def test_execute_dex_trade(client):
    trade_params = {
        "symbol": "SOL/USD",
        "amount": 1000,
        "price": 100.50,
        "slippage": 0.5
    }
    response = client.post(f"/api/v1/trading/trade/{TradingMode.DEX.value}", json=trade_params)
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "success"
    assert data["mode"] == "dex"
    assert "trade_id" in data
    assert data["params"] == trade_params

def test_execute_pump_trade(client):
    trade_params = {
        "symbol": "PEPE/USD",
        "amount": 500,
        "price": 0.0001,
        "slippage": 1.0
    }
    response = client.post(f"/api/v1/trading/trade/{TradingMode.PUMP.value}", json=trade_params)
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "success"
    assert data["mode"] == "pump"
    assert "trade_id" in data
    assert data["params"] == trade_params

def test_invalid_trade_mode(client):
    trade_params = {
        "symbol": "SOL/USD",
        "amount": 1000,
        "price": 100.50,
        "slippage": 0.5
    }
    response = client.post("/api/v1/trading/trade/invalid", json=trade_params)
    assert response.status_code == 422

def test_missing_trade_params(client):
    trade_params = {
        "symbol": "SOL/USD",
        "amount": 1000
    }
    response = client.post(f"/api/v1/trading/trade/{TradingMode.DEX.value}", json=trade_params)
    assert response.status_code == 400
    assert "Missing required trade parameters" in response.json()["detail"]

def test_get_market_data(client):
    response = client.get(f"/api/v1/trading/market-data/{TradingMode.DEX.value}?symbol=SOL/USD")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "success"
    assert data["mode"] == TradingMode.DEX.value
    assert "price" in data["data"]
    assert "volume" in data["data"]
    assert "liquidity" in data["data"]

def test_get_trade_status(client):
    trade_id = "dex_sol_usd"
    response = client.get(f"/api/v1/trading/trade/{trade_id}")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "success"
    assert data["trade_id"] == trade_id
    assert "execution_status" in data
    assert "filled_amount" in data
    assert "filled_price" in data
