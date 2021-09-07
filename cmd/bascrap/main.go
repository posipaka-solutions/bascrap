package main

import (
	"github.com/posipaka-trade/bascrap/internal/cfg"
	"github.com/posipaka-trade/bascrap/worker"
	"github.com/posipaka-trade/binance-api-go/pkg/binance"
	"github.com/posipaka-trade/gate-api-go/pkg/gate"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
)

const configPath = "./configs/bascrap.toml"

func main() {
	cmn.InitLoggers("bascrap")
	cmn.LogInfo.Print("Bascrap execution started.")

	binanceApiKey, err := cfg.ApiCredentials(configPath, cfg.BinanceEx)
	if err != nil {
		panic(err.Error())
	}

	gateioApiKey, err := cfg.ApiCredentials(configPath, cfg.GateioEx)
	if err != nil {
		panic(err.Error())
	}

	initialFunds, err := cfg.InitialFunds(configPath)
	if err != nil {
		panic(err.Error())
	}

	w := worker.New(binance.New(binanceApiKey), gate.New(gateioApiKey), initialFunds)
	w.StartMonitoring()
	w.Wg.Wait()
	cmn.LogInfo.Print("Bascrap execution finished.")
}
