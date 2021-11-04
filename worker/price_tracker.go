package worker

import (
	"fmt"
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
	"github.com/posipaka-trade/posipaka-trade-cmn/log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func (worker *Worker) trackPriceGrowth(announcedAssets symbol.Assets, annType announcement.Type) {
	var priceList []string
	switch annType {
	case announcement.Unknown:
		return
	case announcement.NewCrypto:
		log.Info.Println("[PriceTracker] -> Price growth tracker applied to pair ", announcedAssets)
		priceList = priceGetter(worker.gateioConn, announcedAssets)
	case announcement.NewTradingPair:
		operatePair := worker.selectBuyPair(announcedAssets)
		if operatePair.IsEmpty() {
			return
		}

		log.Info.Println("[PriceTracker] -> Price growth tracker applied to pair ", operatePair)
		priceList = priceGetter(worker.binanceConn, operatePair)
	}

	if priceList == nil {
		return
	}

	if fileName := storePriceList(priceList, announcedAssets); len(fileName) != 0 {
		log.Info.Print("[PriceTracker] -> Tracked price stored to file ", fileName)
	}
}

func priceGetter(exchange exchangeapi.ApiConnector, assets symbol.Assets) []string {
	startTime := time.Now()
	var priceList []string
	for time.Now().Sub(startTime) < time.Minute {
		requestTime := time.Now()
		price, err := exchange.GetCurrentPrice(assets)
		if err != nil {
			priceList = append(priceList, "Finished with an error -> "+err.Error())
			break
		}
		responseTime := time.Now()

		requestDur := responseTime.Sub(requestTime) / 2
		priceList = append(priceList, fmt.Sprintf("[%s] %s/%s -> %s Request time: %d ms\n",
			time.Now().Add((-requestDur)*time.Nanosecond).Format(time.StampMicro),
			assets.Base, assets.Quote,
			strconv.FormatFloat(price, 'f', -1, 64),
			requestDur/time.Millisecond))

		time.Sleep(time.Millisecond)
	}

	priceList = append(priceList, "Finished after one minute.")
	return priceList
}

func storePriceList(priceList []string, assets symbol.Assets) string {
	filePath := filepath.Join(".", "price_tracker")
	err := os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		log.Error.Print("[PriceTracker] -> ", err)
		return ""
	}

	filePath = filepath.Join(filePath, fmt.Sprintf("%s%s_%d.txt",
		assets.Base, assets.Quote, time.Now().UnixMilli()))
	file, err := os.Create(filePath)
	if err != nil {
		log.Error.Print("[PriceTracker] -> ", err)
		return ""
	}
	defer file.Close()

	for _, line := range priceList {
		_, err = file.WriteString(line)
		if err != nil {
			log.Error.Print("[PriceTracker] -> ", err)
			return ""
		}
	}
	err = file.Sync()

	if err != nil {
		log.Error.Print("[PriceTracker] -> ", err)
		return ""
	}

	return filePath
}
