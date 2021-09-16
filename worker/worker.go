package worker

import (
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/posipaka-trade/bascrap/internal/announcement/analyzer"
	"github.com/posipaka-trade/bascrap/internal/scraper"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi"
	"github.com/zelenin/go-tdlib/client"
	"path/filepath"
	"sync"
	"time"
)

const (
	apiId   = 8061033
	apiHash = "5665589a975a637135402402542dd520"
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

	tdlibClient := newTDLibClient()
	if tdlibClient == nil {
		return
	}

	worker.Wg.Add(2)
	go worker.monitorController(tdlibClient)
	go worker.monitorController(tdlibClient)

	cmn.LogInfo.Print("Monitoring started.")
}

func newTDLibClient() *client.Client {
	authorizer := client.ClientAuthorizer()
	go client.CliInteractor(authorizer)

	authorizer.TdlibParameters <- &client.TdlibParameters{
		DatabaseDirectory:      filepath.Join(".tdlib", "database"),
		FilesDirectory:         filepath.Join(".tdlib", "files"),
		UseFileDatabase:        true,
		UseChatInfoDatabase:    true,
		UseMessageDatabase:     true,
		ApiId:                  apiId,
		ApiHash:                apiHash,
		SystemLanguageCode:     "en",
		DeviceModel:            "PosipakaServer",
		ApplicationVersion:     "0.9-development",
		EnableStorageOptimizer: true,
	}

	logVerbosity := client.WithLogVerbosity(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 0,
	})

	tdlibClient, err := client.NewClient(authorizer, logVerbosity)
	if err != nil {
		cmn.LogError.Print("TDLib client creation failed. Error: ", err)
		return nil
	}

	return tdlibClient
}

func (worker *Worker) monitorController(tclient *client.Client) {
	defer worker.Wg.Done()
	handler := scraper.New(tclient)
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
