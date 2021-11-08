package worker

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/posipaka-trade/bascrap/internal/announcement/analyzer"
	"github.com/posipaka-trade/bascrap/internal/scraper"
	"github.com/posipaka-trade/bascrap/internal/telegram"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
	"github.com/posipaka-trade/posipaka-trade-cmn/log"
	"github.com/zelenin/go-tdlib/client"
	"sync"
	"time"
)

const cryptoAlarmPort = 6969

type Worker struct {
	gateioConn, binanceConn exchangeapi.ApiConnector
	tdClient                *client.Client

	initialFunds             float64
	notificationsQueue       []string
	sendTelegramNotification bool

	zctx        *zmq4.Context
	alarmSocket *zmq4.Socket

	messageMutex        sync.Mutex
	latestHandleMessage string
	newAnnouncement     chan string

	Wg        sync.WaitGroup
	isWorking bool
}

func New(binanceConn, gateioConn exchangeapi.ApiConnector, funds float64, sendTelegramNotification bool) *Worker {
	worker := &Worker{
		initialFunds:             funds,
		gateioConn:               gateioConn,
		binanceConn:              binanceConn,
		notificationsQueue:       make([]string, 15),
		sendTelegramNotification: sendTelegramNotification,
		newAnnouncement:          make(chan string),
	}

	worker.notificationsQueue = worker.notificationsQueue[:0]
	return worker
}

func (worker *Worker) StartMonitoring() {
	worker.isWorking = true

	limits, err := worker.binanceConn.GetSymbolsLimits()
	if err != nil {
		log.Error.Fatal("Failed to get symbols limits from Binance. ", err)
	}
	worker.binanceConn.StoreSymbolsLimits(limits)

	worker.tdClient = telegram.NewTDLibClient()
	if worker.tdClient == nil {
		log.Error.Fatal("Failed to create telegram lib client.")
	}

	err = worker.bindAlarmSocket()
	if err != nil {
		log.Error.Fatal("Failed to bind CryptoAlarm socket. ", err)
	}

	handler := scraper.New(worker.tdClient)
	worker.Wg.Add(3)
	go worker.webpageScanner(&handler)
	go worker.telegramScanner(&handler)
	go worker.monitorController()

	log.Info.Print("Monitoring started.")
	if worker.sendTelegramNotification {
		telegram.SendMessageToChannel("Monitoring started.", worker.tdClient)
	}
}

func (worker *Worker) monitorController() {
	defer worker.Wg.Done()
	for worker.isWorking {
		newsTitle, isOkay := <-worker.newAnnouncement
		if !isOkay {
			break
		}

		worker.processAnnouncement(newsTitle)
		worker.sendTelegramNotifications()

		limits, err := worker.binanceConn.GetSymbolsLimits()
		if err != nil {
			log.Info.Print("Failed to get symbols limits from Binance")
		}

		worker.binanceConn.StoreSymbolsLimits(limits)
	}
}

func (worker *Worker) processAnnouncement(newsTitle string) {
	symbolAssets, announcedType := analyzer.AnnouncementSymbol(newsTitle)

	switch announcedType {
	case announcement.Unknown:
		log.Warning.Print("This new announcement is unuseful for Bascrap")
		break
	case announcement.NewCrypto:
		if symbolAssets.IsEmpty() {
			log.Warning.Print("New crypto did not get form latest announcement header. -- " +
				newsTitle)
		} else {
			//if strings.Contains(announcedDetails.Header, "Innovation Zone") {
			//	worker.notificationsQueue = append(worker.notificationsQueue,
			//		fmt.Sprintf("New crypto %s appears in the Innovation Zone", symbolAssets.Base))
			//	log.Info.Print(len(worker.notificationsQueue) - 1)
			//	return
			//}
			worker.ProcessCryptoAnnouncement(symbolAssets)
			worker.causeCryptoAlarm()
		}
		break
	case announcement.NewTradingPair:
		if symbolAssets.IsEmpty() {
			log.Warning.Print("New trading pair did not get form latest announcement header. -- " +
				newsTitle)
		} else {
			worker.ProcessTradingPairAnnouncement(symbolAssets)
			worker.causeCryptoAlarm()
		}
		break
	}
}

func (worker *Worker) ProcessCryptoAnnouncement(symbolAssets symbol.Assets) {
	worker.notificationsQueue = append(worker.notificationsQueue,
		fmt.Sprintf("%s/%s new crypto pair was announced.", symbolAssets.Base, symbolAssets.Quote))
	log.Info.Print(worker.notificationsQueue[len(worker.notificationsQueue)-1])

	hagglingParams, err := worker.buyNewCrypto(symbolAssets)
	if err != nil {
		worker.notificationsQueue = append(worker.notificationsQueue, "New crypto buying failed.\n"+err.Error())
		log.Warning.Print(worker.notificationsQueue[len(worker.notificationsQueue)-1])
	} else {
		worker.notificationsQueue = append(worker.notificationsQueue,
			fmt.Sprintf("Bascrap bought new crypto %s at gate.io.\nBought quantity -> %f.\nPrice -> %f", symbolAssets.Base, hagglingParams.boughtQuantity, hagglingParams.boughtPrice))
		log.Info.Print(worker.notificationsQueue[len(worker.notificationsQueue)-1])
		worker.sellCrypto(&hagglingParams)
	}
}

func (worker *Worker) ProcessTradingPairAnnouncement(symbolAssets symbol.Assets) {
	worker.notificationsQueue = append(worker.notificationsQueue,
		fmt.Sprintf("%s/%s new trading pair was announced.", symbolAssets.Base, symbolAssets.Quote))
	log.Info.Printf(worker.notificationsQueue[len(worker.notificationsQueue)-1])

	hagglingParams := worker.buyNewFiat(symbolAssets)
	if !hagglingParams.symbol.IsEmpty() && hagglingParams.boughtQuantity != 0 {
		worker.notificationsQueue = append(worker.notificationsQueue,
			fmt.Sprintf("Bascrap bought %s using %s after new fiat announcement.\nBought quantity -> %.9f.\nPrice -> %.9f",
				hagglingParams.symbol.Base, hagglingParams.symbol.Quote, hagglingParams.boughtQuantity, hagglingParams.boughtPrice))
		log.Info.Print(worker.notificationsQueue[len(worker.notificationsQueue)-1])
		worker.sellCrypto(&hagglingParams)
	} else {
		worker.notificationsQueue = append(worker.notificationsQueue, "New fiat buy failed.")
		log.Info.Print(worker.notificationsQueue[len(worker.notificationsQueue)-1])
	}
}

func (worker *Worker) sendTelegramNotifications() {
	if !worker.sendTelegramNotification {
		return
	}

	for i := 0; len(worker.notificationsQueue) > i; i++ {
		telegram.SendMessageToChannel(worker.notificationsQueue[i], worker.tdClient)
		time.Sleep(1 * time.Second)
	}
	worker.notificationsQueue = worker.notificationsQueue[:0]
}

func (worker *Worker) bindAlarmSocket() error {
	var err error
	worker.zctx, err = zmq4.NewContext()
	if err != nil {
		return err
	}

	worker.alarmSocket, err = worker.zctx.NewSocket(zmq4.PUB)
	if err != nil {
		return err
	}

	err = worker.alarmSocket.Bind(fmt.Sprint("tcp://*:", cryptoAlarmPort))
	if err != nil {
		return err
	}

	return nil
}

func (worker *Worker) causeCryptoAlarm() {
	_, err := worker.alarmSocket.Send("ala-a-arm", 0)
	if err != nil {
		log.Warning.Print("Crypto alarm sending failed. ", err)
		return
	}
	log.Info.Print("Crypto alarm sent out.")
}
