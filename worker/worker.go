package worker

import (
	"fmt"
	"github.com/posipaka-trade/bascrap/scraper"
	"github.com/posipaka-trade/bascrap/scraper/announcement"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"strings"
	"time"
)

const usdt = "USDT"

func StartMonitoring() {
	cryptoListingHandler := scraper.New(announcement.NewCryptoListing)
	fiatListingHandler := scraper.New(announcement.NewFiatListing)

	for {
		if !(time.Now().Hour() >= 6 && time.Now().Hour() <= 13) {
			time.Sleep(1 * time.Minute)
		}
		checkCryptoNews(cryptoListingHandler)
		checkFiatNews(fiatListingHandler)
		time.Sleep(1 * time.Second)
	}
}

func checkCryptoNews(handler scraper.ScrapHandler) string {
	news, err := handler.GetLatestNews()
	if err != nil {
		_, isOkay := err.(*scraper.NoNewsUpdate)
		if !isOkay {
			cmn.LogError.Print(err.Error())
			return ""
		}
	}

	if !strings.Contains(news.Header, "Binance Will List") {
		return ""
	}

	cryptoList := make([]string, 0)
	headerWords := strings.Fields(news.Header)
	for _, word := range headerWords {
		if strings.HasPrefix(word, "(") && strings.HasSuffix(word, ")") {
			cryptoList = append(cryptoList, fmt.Sprint(word[1:len(word)-1], "/", usdt))
			cmn.LogInfo.Print("New crypto ", cryptoList[len(cryptoList)-1])
		}
	}

	if len(cryptoList) == 0 {
		return ""
	}
	return cryptoList[0]
}

func checkFiatNews(handler scraper.ScrapHandler) string {
	news, err := handler.GetLatestNews()
	if err != nil {
		_, isOkay := err.(*scraper.NoNewsUpdate)
		if !isOkay {
			cmn.LogError.Print(err.Error())
			return ""
		}
	}

	if !strings.Contains(news.Header, "Binance Adds") {
		return ""
	}

	fiatList := make([]string, 0)
	headerWords := strings.Fields(news.Header)
	for _, word := range headerWords {
		if strings.Contains(word, "/") {
			if strings.HasSuffix(word, ",") {
				fiatList = append(fiatList, word[:len(word)-1])
			} else {
				fiatList = append(fiatList, word[:len(word)])
			}
			cmn.LogInfo.Print("New fiat ", fiatList[len(fiatList)-1])
		}
	}

	if len(fiatList) == 0 {
		return ""
	}
	return fiatList[0]
}
