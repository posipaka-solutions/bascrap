/*
Package cfg implements bascrap configuration file parser
*/
package cfg

import (
	"errors"
	"github.com/pelletier/go-toml"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi"
	"github.com/posipaka-trade/posipaka-trade-cmn/log"
)

const (
	BinanceEx = "binance"
	GateioEx  = "gate"
)

// ApiCredentials parse info about api keys for a specified exchange
func ApiCredentials(cfgPath, exchange string) (exchangeapi.ApiKey, error) {
	tml, err := toml.LoadFile(cfgPath)
	if err != nil {
		return exchangeapi.ApiKey{}, errors.New("[cfg] Config file (" + cfgPath + ") loading error " + err.Error())
	}

	apiKey := tml.Get(exchange + "_api_cred.key")
	if apiKey == nil {
		log.Error.Print("[cfg] Api key for ", exchange, " exchange was not recognized.")
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

// InitialFunds return a value of initial funds specified in configuration file
func InitialFunds(cfgPath string) (float64, error) {
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
