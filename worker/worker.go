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

	limits, err := worker.binanceConn.GetSymbolsLimits()
	if err != nil {
		cmn.LogInfo.Print("Failed to get symbols limits from Binance")
	}
	worker.binanceConn.StoreSymbolsLimits(limits)

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
			if _, isOkay := err.(*scraper.NoNewsUpdate); isOkay {
				cmn.LogWarning.Print(err.Error())
			} else {
				cmn.LogError.Print(err.Error())
			}
			continue
		}

		cmn.LogInfo.Print("New announcement on Binance.")
		worker.processAnnouncement(announcedDetails)

		limits, err := worker.binanceConn.GetSymbolsLimits()
		if err != nil {
			cmn.LogInfo.Print("Failed to get symbols limits from Binance")
		}

		worker.binanceConn.StoreSymbolsLimits(limits)

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
			quantity := worker.buyNewCrypto(symbolAssets)
			if quantity != 0 {
				cmn.LogInfo.Printf("Bascrap bought new crypto %s at gate.io. Bought quantity %f",
					symbolAssets.Base, quantity)
			} else {
				cmn.LogWarning.Print("New crypto buying failed.")
			}
		}
		break
	case announcement.NewTradingPair:
		if symbolAssets.IsEmpty() {
			cmn.LogWarning.Print("New trading pair did not get form latest announcement header. -- " +
				announcedDetails.Header)
		} else {
			buyPair, quantity := worker.buyNewFiat(symbolAssets)
			if !buyPair.IsEmpty() && quantity != 0 {
				cmn.LogInfo.Printf("Bascrap bought %s using %s after new fiat announcement. Bought quantity %f",
					buyPair.Base, buyPair.Quote, quantity)
			} else {
				cmn.LogWarning.Print("New fiat buy failed.")
			}
		}
		break
	}
}
