package main

import (
	"github.com/pelletier/go-toml"
	"github.com/posipaka-trade/bascrap/worker"
	"github.com/posipaka-trade/binance-api-go/pkg/binance"
	"github.com/posipaka-trade/gate-api-go/pkg/gate"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi"
)

const (
	binanceEx = "binance"
	gateioEx  = "gateio"
)

const configPath = "./configs/bascrap.toml"

func main() {
	cmn.InitLoggers("bascrap")

	binanceApiKey, isOkay := parseApiCred(binanceEx)
	if !isOkay {
		return
	}

	gateioApiKey, isOkay := parseApiCred(gateioEx)
	if !isOkay {
		return
	}

	tml, err := toml.LoadFile(configPath)
	if err != nil {
		panic("[cfg] Config parsing error " + err.Error())
	}

	quantity := tml.Get("quantityToSpend")
	if quantity == nil {
		panic("[cfg] quantityToSpend not specified")
	}

	w := worker.New(binance.New(binanceApiKey), gate.New(gateioApiKey), quantity.(float64))
	w.StartMonitoring()
}

func parseApiCred(exchange string) (exchangeapi.ApiKey, bool) {
	tml, err := toml.LoadFile(configPath)
	if err != nil {
		cmn.LogError.Printf("[cfg] Config parsing error %s", err.Error())
		return exchangeapi.ApiKey{}, false
	}

	apiKey := tml.Get(exchange + "_api_cred.key")
	if apiKey == nil {
		cmn.LogError.Print("[cfg] ", exchange, " key not specified.")
		return exchangeapi.ApiKey{}, false
	}

	apiSecret := tml.Get(exchange + "_api_cred.secret")
	if apiSecret == nil {
		cmn.LogError.Print("[cfg] ", exchange, " secret not specified.")
		return exchangeapi.ApiKey{}, false
	}

	return exchangeapi.ApiKey{
		Key:    apiKey.(string),
		Secret: apiSecret.(string),
	}, true
}
