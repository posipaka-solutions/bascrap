package worker

import (
	"fmt"
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/posipaka-trade/bascrap/internal/announcement/analyzer"
	"github.com/posipaka-trade/bascrap/internal/scraper"
	"github.com/posipaka-trade/bascrap/internal/telegram"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi"
	"github.com/zelenin/go-tdlib/client"
	"strings"
	"sync"
	"time"
)

type Worker struct {
	gateioConn, binanceConn exchangeapi.ApiConnector
	initialFunds            float64
	Wg                      sync.WaitGroup
	isWorking               bool
	TdClient                *client.Client
	NotificationsQueue      []string
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

	worker.TdClient = telegram.NewTDLibClient()
	if worker.TdClient == nil {
		return
	}

	worker.Wg.Add(1)
	go worker.monitorController(worker.TdClient)

	monitoringInfo := "Monitoring started."
	cmn.LogInfo.Print(monitoringInfo)
	telegram.SendMessageToChannel(monitoringInfo, worker.TdClient)
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
				worker.NotificationsQueue = append(worker.NotificationsQueue,
					fmt.Sprintf("New crypto %s appears in the Innovation Zone", symbolAssets.Base))
				return
			}
			announcementInfo := fmt.Sprintf("%s/%s new crypto pair was announced.", symbolAssets.Base, symbolAssets.Quote)
			cmn.LogInfo.Printf(announcementInfo)
			worker.NotificationsQueue = append(worker.NotificationsQueue, announcementInfo)
			quantity := worker.buyNewCrypto(symbolAssets)
			if quantity != 0 {
				cryptoInfo := fmt.Sprintf("Bascrap bought new crypto %s at gate.io. Bought quantity %f",
					symbolAssets.Base, quantity)
				cmn.LogInfo.Printf(cryptoInfo)
				worker.NotificationsQueue = append(worker.NotificationsQueue, cryptoInfo)
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
			announcementInfo := fmt.Sprintf("%s/%s new trading pair was announced.", symbolAssets.Base, symbolAssets.Quote)
			cmn.LogInfo.Printf(announcementInfo)
			worker.NotificationsQueue = append(worker.NotificationsQueue, announcementInfo)
			buyPair, quantity := worker.buyNewFiat(symbolAssets)
			if !buyPair.IsEmpty() && quantity != 0 {
				fiatInfo := fmt.Sprintf("Bascrap bought %s using %s after new fiat announcement. Bought quantity %f", buyPair.Base, buyPair.Quote, quantity)
				cmn.LogInfo.Printf(fiatInfo)
				worker.NotificationsQueue = append(worker.NotificationsQueue, fiatInfo)
			} else {
				cmn.LogWarning.Print("New fiat buy failed.")
			}
		}
		break
	}
}

func (worker *Worker) sendTelegramNotifications() {

	for i := 0; len(worker.NotificationsQueue) > i; i++ {
		telegram.SendMessageToChannel(worker.NotificationsQueue[i], worker.TdClient)
		time.Sleep(1 * time.Second)
	}
	worker.NotificationsQueue = nil
}
