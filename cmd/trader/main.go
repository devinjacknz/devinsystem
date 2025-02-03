package main

import (
	"log"
	
	"github.com/devinjacknz/devintrade/internal/trading"
)

func main() {
	engine := trading.NewTradingEngine()
	log.Fatal(engine.Start())
}
