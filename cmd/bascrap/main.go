package main

import (
	"github.com/pelletier/go-toml"
	"github.com/posipaka-trade/bascrap/worker"
	"github.com/posipaka-trade/binance-api-go/pkg/binance"
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

	//gateioApiKey, isOkay := parseApiCred(gateioEx)
	//if !isOkay {
	//	return
	//}

	w := worker.New(binance.New(binanceApiKey), nil)
	w.StartMonitoring()
}

func parseApiCred(exchange string) (exchangeapi.ApiKey, bool) {
	tml, err := toml.LoadFile(configPath)
	if err != nil {
		cmn.LogError.Printf("[cfg] Config parsing error %s", err.Error())
		return exchangeapi.ApiKey{}, false
	}

	apiKey := tml.Get(exchange + "api_cred.key")
	if apiKey == nil {
		cmn.LogError.Print("[cfg] ", exchange, " key not specified.")
		return exchangeapi.ApiKey{}, false
	}

	apiSecret := tml.Get(exchange + "api_cred.secret")
	if apiSecret == nil {
		cmn.LogError.Print("[cfg] ", exchange, " secret not specified.")
		return exchangeapi.ApiKey{}, false
	}

	return exchangeapi.ApiKey{
		Key:    apiKey.(string),
		Secret: apiSecret.(string),
	}, true
}
