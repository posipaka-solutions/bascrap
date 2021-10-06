package worker

import (
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/posipaka-trade/bascrap/internal/assets"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/order"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
	"github.com/posipaka-trade/posipaka-trade-cmn/log"
)

// buyNewCrypto returns bought quantity of base
func (worker *Worker) buyNewCrypto(newSymbol symbol.Assets) (hagglingParameters, error) {
	price, err := worker.gateioConn.GetCurrentPrice(newSymbol)
	if err != nil {
		return hagglingParameters{}, err
	}

	hagglingParams := hagglingParameters{
		announcementType: announcement.NewCrypto,
		boughtPrice:      price,
		symbol:           newSymbol,
	}
	hagglingParams.boughtQuantity, err = worker.setCryptoOrder(newSymbol, price)
	if err != nil {
		return hagglingParameters{}, err
	}
	return hagglingParams, nil
}

func (worker *Worker) setCryptoOrder(newSymbol symbol.Assets, price float64) (float64, error) {
	parameters := order.Parameters{
		Assets:   newSymbol,
		Side:     order.Buy,
		Type:     order.Limit,
		Quantity: worker.initialFunds / (price * 1.5),
		Price:    price * 1.5,
	}

	_, err := worker.gateioConn.SetOrder(parameters)
	log.Info.Printf("Limit order on gate.io:Quantity value - %f, Price value - %f", parameters.Quantity, parameters.Price)
	if err != nil {
		return 0, err
	}

	return worker.initialFunds / (price * 1.05), nil
}

// buyNewFiat perform buy of base asset of new fiat pair. Returns symbol and quantity of buy transactions
func (worker *Worker) buyNewFiat(newTradingPair symbol.Assets) hagglingParameters {
	buyPair := worker.selectBuyPair(newTradingPair)
	if buyPair.IsEmpty() {
		log.Error.Print("New trading pair are missing because suitable buy pair for it not found.")
		return hagglingParameters{}
	}

	if buyPair.Quote != assets.Busd {
		newQuoteQuantity := worker.transferFunds(buyPair)
		if newQuoteQuantity == 0 {
			return hagglingParameters{}
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
		log.Error.Print(err)
		return hagglingParameters{}
	}

	price, err := worker.binanceConn.GetCurrentPrice(buyPair)
	if err != nil {
		worker.notificationsQueue = append(worker.notificationsQueue, err.Error())
		log.Error.Print(err)
		return hagglingParameters{}
	}

	return hagglingParameters{
		announcementType: announcement.NewTradingPair,
		boughtPrice:      price,
		boughtQuantity:   quantity,
		symbol:           buyPair,
	}
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
		log.Error.Print(err)
		return 0
	}

	return quantity
}

func (worker *Worker) selectBuyPair(newTradingPair symbol.Assets) symbol.Assets {
	allExchangeSymbols := worker.binanceConn.GetSymbolsList()
	if allExchangeSymbols == nil ||
		len(allExchangeSymbols) == 0 {
		log.Error.Print("Symbols list for exchange is empty.")
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
