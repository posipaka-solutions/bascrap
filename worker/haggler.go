package worker

import (
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/posipaka-trade/bascrap/internal/assets"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/order"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
	"github.com/posipaka-trade/posipaka-trade-cmn/log"
)

const (
	cryptoGrowthPercent   = 1.12
	usdtPairGrowthPercent = 1.05
	busdPairGrowthPercent = 1.10
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

	if parameters.announcementType == announcement.NewCrypto {
		orderParameters.Quantity, err = worker.gateioConn.GetAssetBalance(orderParameters.Assets.Base)
		if err != nil {
			worker.notificationsQueue = append(worker.notificationsQueue, err.Error())
			log.Error.Print(err)
		}
		orderParameters.Price = parameters.boughtPrice * cryptoGrowthPercent
		_, err = worker.gateioConn.SetOrder(orderParameters)
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

		_, err = worker.binanceConn.SetOrder(orderParameters)
		if err != nil {
			worker.notificationsQueue = append(worker.notificationsQueue, err.Error())
			log.Error.Print(err)
		}
	}
}
