package worker

import (
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/posipaka-trade/bascrap/internal/announcement/analyzer"
	"github.com/posipaka-trade/bascrap/internal/scraper"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi"
	"sync"
	"time"
)

type Worker struct {
	gateioConn, binanceConn exchangeapi.ApiConnector
	initialFunds            float64
	Wg                      sync.WaitGroup
	isWorking               bool
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
	err := worker.binanceConn.UpdateSymbolsList()
	if err != nil {
		cmn.LogInfo.Print("Failed to get symbols list from Binance")
		return
	}

	worker.Wg.Add(2)
	go worker.monitorController(announcement.NewCryptoListingUrl)
	go worker.monitorController(announcement.NewFiatListingUrl)

	cmn.LogInfo.Print("Monitoring started.")
}

func (worker *Worker) monitorController(monitoringUrl string) {
	defer worker.Wg.Done()
	handler := scraper.New(monitoringUrl)
	for worker.isWorking {
		time.Sleep(5 * time.Second)

		announcedDetails, err := handler.GetLatestAnnounce()
		if err != nil {
			cmn.LogError.Print(err.Error())
			continue
		}

		cmn.LogInfo.Print("New announcement on Binance.")
		worker.processAnnouncement(announcedDetails)

		err = worker.binanceConn.UpdateSymbolsList()
		if err != nil {
			cmn.LogInfo.Print("Failed to get symbols list from Binance")
		}
	}
}

func (worker *Worker) processAnnouncement(announcedDetails announcement.Details) {
	symbolAssets, announcedType := analyzer.AnnouncementSymbol(announcedDetails)
	switch announcedType {
	case announcement.Unknown:
		cmn.LogWarning.Print("This new announcement is unuseful for Bascrap")
		break
	case announcement.NewCrypto:
		if symbolAssets.IsEmpty() {
			cmn.LogWarning.Print("New crypto did not get form latest announcement header. -- " +
				announcedDetails.Header)
		} else {
			worker.buyNewCrypto(symbolAssets)
		}
		break
	case announcement.NewTradingPair:
		if symbolAssets.IsEmpty() {
			cmn.LogWarning.Print("New trading pair did not get form latest announcement header. -- " +
				announcedDetails.Header)
		} else {
			worker.buyNewFiat(symbolAssets)
		}
		break
	}
}
