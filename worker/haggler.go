package worker

import (
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/order"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
	"github.com/posipaka-trade/posipaka-trade-cmn/log"
)

const cryptoGrowthPercent = 1.12
const tradingPairGrowthPercent = 1.05

type hagglingParameters struct {
	announcementType            announcement.Type
	boughtPrice, boughtQuantity float64
	symbol                      symbol.Assets
}

func (worker *Worker) sellCrypto(parameters *hagglingParameters) {
	worker.notificationsQueue = append(worker.notificationsQueue, "Setting profit order immediately after announcement")
	log.Info.Print(worker.notificationsQueue[len(worker.notificationsQueue)-1])

	orderParameters := order.Parameters{
		Assets:   parameters.symbol,
		Side:     order.Sell,
		Type:     order.Limit,
		Quantity: parameters.boughtQuantity,
	}

	if parameters.announcementType == announcement.NewCrypto {
		orderParameters.Price = parameters.boughtPrice * cryptoGrowthPercent
		_, err := worker.gateioConn.SetOrder(orderParameters)
		if err != nil {
			worker.notificationsQueue = append(worker.notificationsQueue, err.Error())
			log.Error.Print(err)
		}
	} else if parameters.announcementType == announcement.NewTradingPair {
		orderParameters.Price = parameters.boughtPrice * tradingPairGrowthPercent
		_, err := worker.binanceConn.SetOrder(orderParameters)
		if err != nil {
			worker.notificationsQueue = append(worker.notificationsQueue, err.Error())
			log.Error.Print(err)
		}
	}
}
