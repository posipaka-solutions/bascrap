package worker

import (
	"github.com/posipaka-trade/bascrap/internal/assets"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/order"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
)

// buyNewCrypto returns bought quantity of base
func (worker *Worker) buyNewCrypto(newSymbol symbol.Assets) float64 {
	price, err := worker.gateioConn.GetCurrentPrice(newSymbol)
	if err != nil {
		cmn.LogError.Print(err.Error())
		return 0
	}
	cmn.LogInfo.Printf("Price for %s%s at gate.io -> %f", newSymbol.Base, newSymbol.Quote, price)

	return worker.setCryptoOrder(newSymbol, price)
}

func (worker *Worker) setCryptoOrder(newSymbol symbol.Assets, price float64) float64 {
	parameters := order.Parameters{
		Assets:   newSymbol,
		Side:     order.Buy,
		Type:     order.Limit,
		Quantity: worker.initialFunds / (price * 1.5),
		Price:    price * 1.5,
	}

	cmn.LogInfo.Printf("Limit order on gate.io:Quantity value - %f, Price value - %f", parameters.Quantity, parameters.Price)

	_, err := worker.gateioConn.SetOrder(parameters)
	if err != nil {
		cmn.LogError.Print(err.Error())
		worker.notificationsQueue = append(worker.notificationsQueue, err.Error())
		return 0
	}

	return worker.initialFunds / (price * 1.05)
}

// buyNewFiat perform buy of base asset of new fiat pair. Returns symbol and quantity of buy transactions
func (worker *Worker) buyNewFiat(newTradingPair symbol.Assets) (symbol.Assets, float64) {
	buyPair := worker.selectBuyPair(newTradingPair)
	if buyPair.IsEmpty() {
		cmn.LogError.Print("New trading pair are missing because suitable buy pair for it not found.")
		return symbol.Assets{}, 0
	}

	if buyPair.Quote != assets.Busd {
		newQuoteQuantity := worker.transferFunds(buyPair)
		if newQuoteQuantity == 0 {
			return symbol.Assets{}, 0
		}
		worker.initialFunds = newQuoteQuantity
	}

	params := order.Parameters{
		Assets:   buyPair,
		Side:     order.Buy,
		Type:     order.Market,
		Quantity: worker.initialFunds,
	}
	quantity, err := worker.binanceConn.SetOrder(params)
	if err != nil {
		worker.notificationsQueue = append(worker.notificationsQueue, err.Error())
		cmn.LogError.Print(err)
		return symbol.Assets{}, 0
	}

	return buyPair, quantity
}

func (worker *Worker) transferFunds(buyPair symbol.Assets) float64 {
	params := order.Parameters{
		Assets: symbol.Assets{
			Base:  buyPair.Quote,
			Quote: assets.Busd,
		},
		Side:     order.Buy,
		Type:     order.Market,
		Quantity: worker.initialFunds,
	}

	if buyPair.Quote == assets.Usdt {
		params.Side = order.Sell
		params.Assets.Base = assets.Busd
		params.Assets.Quote = assets.Usdt
	}

	quantity, err := worker.binanceConn.SetOrder(params)
	if err != nil {
		worker.notificationsQueue = append(worker.notificationsQueue, err.Error())
		cmn.LogError.Print(err)
		return 0
	}

	return quantity
}

func (worker *Worker) selectBuyPair(newTradingPair symbol.Assets) symbol.Assets {
	allExchangeSymbols := worker.binanceConn.GetSymbolsList()
	if allExchangeSymbols == nil ||
		len(allExchangeSymbols) == 0 {
		cmn.LogError.Print("Symbols list for exchange is empty.")
		return symbol.Assets{}
	}

	suitableArrayIndex := 0
	for _, symb := range allExchangeSymbols {
		if symb.Base == newTradingPair.Base {
			allExchangeSymbols[suitableArrayIndex] = symb
			suitableArrayIndex++
		}
	}
	allExchangeSymbols = allExchangeSymbols[:suitableArrayIndex]

	for _, priorityAsset := range assets.BuyPriority {
		if priorityAsset == newTradingPair.Quote {
			continue
		}

		for idx := range allExchangeSymbols {
			if allExchangeSymbols[idx].Quote != priorityAsset {
				continue
			}
			return allExchangeSymbols[idx]
		}
	}
	return symbol.Assets{}
}
