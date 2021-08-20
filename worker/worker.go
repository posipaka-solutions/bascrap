package worker

import (
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/posipaka-trade/bascrap/internal/scraper"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
	"strings"
	"time"
)

type Worker struct {
	gateioConn, binanceConn exchangeapi.ApiConnector
	initialFunds            float64

	isWorking bool
}

func New(binanceConn, gateioConn exchangeapi.ApiConnector, funds float64) *Worker {
	return &Worker{
		initialFunds: funds,
		gateioConn:   gateioConn,
		binanceConn:  binanceConn,
	}
}

func (worker *Worker) StartMonitoring() {
	worker.isWorking = true
	cryptoListingHandler := scraper.New(announcement.NewCryptoListing)
	fiatListingHandler := scraper.New(announcement.NewFiatListing)

	cmn.LogInfo.Print("Monitoring started.")
	for worker.isWorking {
		newCrypto, isOkay := checkCryptoNews(cryptoListingHandler)
		if isOkay {
			worker.buyNewCrypto(newCrypto)
			worker.isWorking = false
		}
		newFiats := checkFiatNews(fiatListingHandler)
		if newFiats != nil {
			worker.buyNewFiat(newFiats)
			worker.isWorking = false
		}

		time.Sleep(1 * time.Second)
	}
	cmn.LogInfo.Print("Monitoring finished.")
}
