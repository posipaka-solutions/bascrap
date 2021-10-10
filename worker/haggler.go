package worker

import (
	"fmt"
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/posipaka-trade/bascrap/internal/assets"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/order"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
	"github.com/posipaka-trade/posipaka-trade-cmn/log"
)

const (
	cryptoGrowthPercent   = 1.20
	usdtPairGrowthPercent = 1.10
	busdPairGrowthPercent = 1.05
)

type hagglingParameters struct {
	announcementType            announcement.Type
	boughtPrice, boughtQuantity float64
	symbol                      symbol.Assets
}

func (worker *Worker) sellCrypto(parameters *hagglingParameters) {
	worker.notificationsQueue = append(worker.notificationsQueue, "Setting profit order immediately after announcement")
	log.Info.Print(worker.notificationsQueue[len(worker.notificationsQueue)-1])

	orderParameters := order.Parameters{
		Assets: parameters.symbol,
		Side:   order.Sell,
		Type:   order.Limit,
	}
	var err error
	var result float64

	if parameters.announcementType == announcement.NewCrypto {
		orderParameters.Quantity, err = worker.gateioConn.GetAssetBalance(orderParameters.Assets.Base)
		if err != nil {
			worker.notificationsQueue = append(worker.notificationsQueue, err.Error())
			log.Error.Print(err)
		}
		orderParameters.Price = parameters.boughtPrice * cryptoGrowthPercent
		result, err = worker.gateioConn.SetOrder(orderParameters)
		if err != nil {
			worker.notificationsQueue = append(worker.notificationsQueue, err.Error())
			log.Error.Print(err)
		}
	} else if parameters.announcementType == announcement.NewTradingPair {
		orderParameters.Quantity, err = worker.binanceConn.GetAssetBalance(orderParameters.Assets.Base)
		if err != nil {
			worker.notificationsQueue = append(worker.notificationsQueue, err.Error())
			log.Error.Print(err)
		}

		if orderParameters.Assets.Quote == assets.Busd {
			orderParameters.Price = parameters.boughtPrice * usdtPairGrowthPercent
		} else {
			orderParameters.Price = parameters.boughtPrice * busdPairGrowthPercent
		}

		result, err = worker.binanceConn.SetOrder(orderParameters)
		if err != nil {
			worker.notificationsQueue = append(worker.notificationsQueue, err.Error())
			log.Error.Print(err)
		}
	}
	if result > 0 {
		worker.notificationsQueue = append(worker.notificationsQueue, fmt.Sprintf("Profit order was placed at the price -> %f", orderParameters.Price))
		log.Info.Print(worker.notificationsQueue[len(worker.notificationsQueue)-1])
	}
}
