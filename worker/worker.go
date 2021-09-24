package worker

import (
	"fmt"
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/posipaka-trade/bascrap/internal/announcement/analyzer"
	"github.com/posipaka-trade/bascrap/internal/scraper"
	"github.com/posipaka-trade/bascrap/internal/telegram"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
	"github.com/zelenin/go-tdlib/client"
	"strings"
	"sync"
	"time"
)

type Worker struct {
	gateioConn, binanceConn exchangeapi.ApiConnector
	tdClient                *client.Client

	initialFunds       float64
	notificationsQueue []string

	Wg        sync.WaitGroup
	isWorking bool
}

func New(binanceConn, gateioConn exchangeapi.ApiConnector, funds float64) *Worker {
	worker := &Worker{
		initialFunds:       funds,
		gateioConn:         gateioConn,
		binanceConn:        binanceConn,
		notificationsQueue: make([]string, 15),
	}

	worker.notificationsQueue = worker.notificationsQueue[:0]
	return worker
}

func (worker *Worker) StartMonitoring() {
	worker.isWorking = true

	limits, err := worker.binanceConn.GetSymbolsLimits()
	if err != nil {
		cmn.LogInfo.Print("Failed to get symbols limits from Binance")
	}
	worker.binanceConn.StoreSymbolsLimits(limits)

	worker.tdClient = telegram.NewTDLibClient()
	if worker.tdClient == nil {
		return
	}

	worker.Wg.Add(1)
	go worker.monitorController(worker.tdClient)

	monitoringInfo := "Monitoring started."
	cmn.LogInfo.Print(monitoringInfo)
	telegram.SendMessageToChannel(monitoringInfo, worker.tdClient)
}

func (worker *Worker) monitorController(tclient *client.Client) {
	defer worker.Wg.Done()
	handler := scraper.New(tclient)
	for worker.isWorking {
		time.Sleep(time.Millisecond)

		announcedDetails, err := handler.GetLatestAnnounce()
		if err != nil {
			if _, isOkay := err.(*scraper.NoNewsUpdate); !isOkay {
				cmn.LogError.Print(err.Error())
			}
			continue
		}

		cmn.LogInfo.Print("New announcement on Binance.")
		worker.processAnnouncement(announcedDetails)

		worker.sendTelegramNotifications()

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
			if strings.Contains(announcedDetails.Link, "Innovation Zone") {
				worker.notificationsQueue = append(worker.notificationsQueue,
					fmt.Sprintf("New crypto %s appears in the Innovation Zone", symbolAssets.Base))
				return
			}
			worker.processCryptoAnnouncement(symbolAssets)
		}
		break
	case announcement.NewTradingPair:
		if symbolAssets.IsEmpty() {
			cmn.LogWarning.Print("New trading pair did not get form latest announcement header. -- " +
				announcedDetails.Header)
		} else {
			worker.processTradingPairAnnouncement(symbolAssets)
		}
		break
	}
}

func (worker *Worker) processCryptoAnnouncement(symbolAssets symbol.Assets) {
	announcementInfo := fmt.Sprintf("%s/%s new crypto pair was announced.", symbolAssets.Base, symbolAssets.Quote)
	cmn.LogInfo.Printf(announcementInfo)
	worker.notificationsQueue = append(worker.notificationsQueue, announcementInfo)
	quantity := worker.buyNewCrypto(symbolAssets)
	if quantity != 0 {
		cryptoInfo := fmt.Sprintf("Bascrap bought new crypto %s at gate.io. Bought quantity %f",
			symbolAssets.Base, quantity)
		cmn.LogInfo.Printf(cryptoInfo)
		worker.notificationsQueue = append(worker.notificationsQueue, cryptoInfo)
	} else {
		cmn.LogWarning.Print("New crypto buying failed.")
	}
}

func (worker *Worker) processTradingPairAnnouncement(symbolAssets symbol.Assets) {
	announcementInfo := fmt.Sprintf("%s/%s new trading pair was announced.", symbolAssets.Base, symbolAssets.Quote)
	cmn.LogInfo.Printf(announcementInfo)
	worker.notificationsQueue = append(worker.notificationsQueue, announcementInfo)
	buyPair, quantity := worker.buyNewFiat(symbolAssets)
	if !buyPair.IsEmpty() && quantity != 0 {
		fiatInfo := fmt.Sprintf("Bascrap bought %s using %s after new fiat announcement. Bought quantity %f", buyPair.Base, buyPair.Quote, quantity)
		cmn.LogInfo.Printf(fiatInfo)
		worker.notificationsQueue = append(worker.notificationsQueue, fiatInfo)
	} else {
		cmn.LogWarning.Print("New fiat buy failed.")
	}
}

func (worker *Worker) sendTelegramNotifications() {
	for i := 0; len(worker.notificationsQueue) > i; i++ {
		telegram.SendMessageToChannel(worker.notificationsQueue[i], worker.tdClient)
		time.Sleep(1 * time.Second)
	}
	worker.notificationsQueue = worker.notificationsQueue[:0]
}
