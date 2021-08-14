package main

import (
	"errors"
	"github.com/pelletier/go-toml"
	"github.com/posipaka-trade/bascrap/worker"
	"github.com/posipaka-trade/binance-api-go/pkg/binance"
	"github.com/posipaka-trade/gate-api-go/pkg/gate"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi"
)

const (
	binanceEx = "binance"
	gateioEx  = "gate"
)

const configPath = "./configs/bascrap.toml"

func main() {
	cmn.InitLoggers("bascrap")
	cmn.LogInfo.Print("Bascrap execution started.")

	binanceApiKey, err := parseApiCred(configPath, binanceEx)
	if err != nil {
		panic(err.Error())
	}

	gateioApiKey, err := parseApiCred(configPath, gateioEx)
	if err != nil {
		panic(err.Error())
	}

	initialFunds, err := parseInitialFunds(configPath)
	if err != nil {
		panic(err.Error())
	}

	w := worker.New(binance.New(binanceApiKey), gate.New(gateioApiKey), initialFunds)
	w.StartMonitoring()
	cmn.LogInfo.Print("Bascrap execution finished.")
}

func parseApiCred(cfgPath, exchange string) (exchangeapi.ApiKey, error) {
	tml, err := toml.LoadFile(cfgPath)
	if err != nil {
		return exchangeapi.ApiKey{}, errors.New("[cfg] Config file (" + cfgPath + ") loading error " + err.Error())
	}

	apiKey := tml.Get(exchange + "_api_cred.key")
	if apiKey == nil {
		cmn.LogError.Print("[cfg] Api key for ", exchange, " exchange was not recognized.")
		return exchangeapi.ApiKey{}, errors.New("[cfg] Api key for " + exchange + " exchanged was not recognized")
	}

	apiSecret := tml.Get(exchange + "_api_cred.secret")
	if apiSecret == nil {
		return exchangeapi.ApiKey{}, errors.New("[cfg] Api secret for " + exchange + " exchanged was not recognized")
	}

	return exchangeapi.ApiKey{
		Key:    apiKey.(string),
		Secret: apiSecret.(string),
	}, nil
}

func parseInitialFunds(cfgPath string) (float64, error) {
	tml, err := toml.LoadFile(cfgPath)
	if err != nil {
		return 0, errors.New("[cfg] Config file (" + cfgPath + ") loading error " + err.Error())
	}

	initialFunds := tml.Get("initial_funds")
	if initialFunds == nil {
		return 0, errors.New("[cfg] Initial funds(`initial_funds`) not specified")
	}

	return initialFunds.(float64), nil
}
