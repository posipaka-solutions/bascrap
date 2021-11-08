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

type Funds struct {
	CryptoFunds      float64
	TradingPairFunds float64
}

// InitialFunds return a value of initial funds specified in configuration file
func InitialFunds(cfgPath string) (Funds, error) {
	tml, err := toml.LoadFile(cfgPath)
	if err != nil {
		return Funds{}, errors.New("[cfg] Config file (" + cfgPath + ") loading error " + err.Error())
	}

	cryptoFunds := tml.Get("crypto_funds")
	if cryptoFunds == nil {
		return Funds{}, errors.New("[cfg] Initial funds(`crypto_funds`) not specified")
	}

	tradingPairFunds := tml.Get("trading_pair_funds")
	if tradingPairFunds == nil {
		return Funds{}, errors.New("[cfg] Initial funds(`trading_pair_funds`) not specified")
	}

	return Funds{
		CryptoFunds:      cryptoFunds.(float64),
		TradingPairFunds: tradingPairFunds.(float64),
	}, nil
}
