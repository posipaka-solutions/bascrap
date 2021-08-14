package worker

import (
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/order"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
)

var quotesPriority = []string{
	"USDT",
	"BUSD",
	"EUR",
	"AUD",
	"GPB",
	"RUB",
}

func (worker *Worker) buyNewCrypto(newSymbol symbol.Assets) bool {
	limits, err := worker.gateioConn.GetSymbolLimits(newSymbol)
	if err != nil {
		cmn.LogError.Print(err.Error())
		return false
	}
	worker.gateioConn.AddLimits(limits)

	price, err := worker.gateioConn.GetCurrentPrice(newSymbol)
	if err != nil {
		cmn.LogError.Print(err.Error())
		return false
	}
	cmn.LogInfo.Print(newSymbol.Base, newSymbol.Quote, " -> ", price)

	parameters := order.Parameters{
		Assets:   newSymbol,
		Side:     order.Buy,
		Type:     order.Limit,
		Quantity: worker.quantityToSpend / (price * 1.05),
		Price:    price * 1.05,
	}
	cmn.LogInfo.Printf("Quantity value - %f, Price value - %f", parameters.Quantity, parameters.Price)
	_, err = worker.gateioConn.SetOrder(parameters)
	if err != nil {
		cmn.LogError.Print(err.Error())
		return false
	}

	cmn.LogInfo.Print("Set order at gateio to buy new crypto")
	return true
}

func (worker *Worker) buyNewFiat(symbolsList []symbol.Assets) bool {
	var newFiat symbol.Assets
	fiatMatched := false
	for _, quote := range quotesPriority {
		if fiatMatched == true {
			break
		}

		for _, symb := range symbolsList {
			if symb.Quote == quote {
				newFiat = symb
				fiatMatched = true
				break
			}
		}
	}

	if !fiatMatched {
		cmn.LogError.Print("Suitable fiat for future trading not found.")
		return false
	}
	cmn.LogInfo.Print("Selected new fiat symbol ", newFiat.Base, newFiat.Quote)

	var buyFiat symbol.Assets
	for idx, quote := range quotesPriority {
		if quote == newFiat.Quote {
			continue
		}

		buyFiat = symbol.Assets{
			Base:  newFiat.Base,
			Quote: quote,
		}
		limits, err := worker.binanceConn.GetSymbolLimits(buyFiat)
		if err != nil {
			cmn.LogError.Print(err.Error())
			if idx == len(quotesPriority)-1 {
				cmn.LogError.Print("Suitable fiat for prebuy not found.")
				return false
			}
			continue
		}
		worker.binanceConn.AddLimits(limits)
		break
	}

	//buyFiatFunds
	_, err := worker.binanceConn.GetAssetBalance(buyFiat.Quote)
	if err != nil {
		cmn.LogError.Print("Failed to transfer fiat balance.")
		return false
	}
	// add funds conversions

	parameters := order.Parameters{
		Assets:   buyFiat,
		Side:     order.Buy,
		Type:     order.Market,
		Quantity: worker.quantityToSpend, // todo change quantity depending on it fiat
	}
	_, err = worker.binanceConn.SetOrder(parameters)
	if err != nil {
		cmn.LogError.Print(err.Error())
		cmn.LogError.Print("Failed to set order on binance.")
		return false
	}
	cmn.LogInfo.Print("New fiat bought on binance")
	return true
}
