package worker

import (
	"github.com/posipaka-trade/bascrap/internal/assets"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/order"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
)

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

	return worker.MakeOrder(newSymbol, price)
}

func (worker *Worker) MakeOrder(newSymbol symbol.Assets, price float64) bool {
	parameters := order.Parameters{
		Assets:   newSymbol,
		Side:     order.Buy,
		Type:     order.Limit,
		Quantity: worker.initialFunds / (price * 1.05),
		Price:    price * 1.05,
	}
	cmn.LogInfo.Printf("Quantity value - %f, Price value - %f", parameters.Quantity, parameters.Price)
	_, err := worker.gateioConn.SetOrder(parameters)
	if err != nil {
		cmn.LogError.Print(err.Error())
		return false
	}
	cmn.LogInfo.Print("Order was set at gate.io")
	return true
}

func (worker *Worker) buyNewFiat(symbolsList []symbol.Assets) bool {
	var newFiat symbol.Assets
	fiatMatched := false
	for _, quote := range assets.Priorities {
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
	for idx, quote := range assets.Priorities {
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
			if idx == len(assets.Priorities)-1 {
				cmn.LogError.Print("Suitable fiat for prebuy not found.")
				return false
			}
			continue
		}
		worker.binanceConn.AddLimits(limits)
		break
	}

	fundsQuantity, err := worker.binanceConn.GetAssetBalance(buyFiat.Quote)
	if err != nil {
		cmn.LogError.Print("Failed to get fiat balance.")
		return false
	}

	newQuantity := 0.0
	if fundsQuantity < 1 {
		newQuantity, err = worker.binanceConn.SetOrder(order.Parameters{
			Assets: symbol.Assets{
				Base:  buyFiat.Quote,
				Quote: assets.Usdt,
			},
			Side:     order.Buy,
			Type:     order.Market,
			Quantity: worker.initialFunds,
		})
		if err != nil {
			cmn.LogError.Print("Failed to set order for buy .")
			return false
		}
	}

	parameters := order.Parameters{
		Assets:   buyFiat,
		Side:     order.Buy,
		Type:     order.Market,
		Quantity: newQuantity, // todo change quantity depending on it fiat
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
