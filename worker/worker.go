package worker

import (
	"github.com/posipaka-trade/bascrap/scraper"
	"github.com/posipaka-trade/bascrap/scraper/announcement"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
	"net/http"
	"strings"
	"time"
)

const usdt = "USDT"

type Worker struct {
	gateioConn, binanceConn exchangeapi.ApiConnector

	quantityToSpend float64
	isWorking       bool
	client          *http.Client
}

func New(binanceConn, gateioConn exchangeapi.ApiConnector, quantity float64) *Worker {
	return &Worker{
		quantityToSpend: quantity,
		gateioConn:      gateioConn,
		binanceConn:     binanceConn,
		client:          &http.Client{},
	}
}

func (worker *Worker) StartMonitoring() {
	worker.isWorking = true
	cryptoListingHandler := scraper.New(announcement.NewCryptoListing)
	fiatListingHandler := scraper.New(announcement.NewFiatListing)

	cmn.LogInfo.Print("Monitoring started.")
	for worker.isWorking {
		if !(time.Now().Hour() >= 3 && time.Now().Hour() <= 13) {
			time.Sleep(1 * time.Minute)
		}
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

func checkCryptoNews(handler scraper.ScrapHandler) (symbol.Assets, bool) {
	news, err := handler.GetLatestNews()
	if err != nil {
		_, isOkay := err.(*scraper.NoNewsUpdate)
		if !isOkay {
			cmn.LogError.Print(err.Error())
			return symbol.Assets{}, false
		}
	}

	if !strings.Contains(news.Header, "Binance Will List") {
		return symbol.Assets{}, false
	}

	headerWords := strings.Fields(news.Header)
	for _, word := range headerWords {
		if strings.HasPrefix(word, "(") && strings.HasSuffix(word, ")") {
			cmn.LogInfo.Print("New crypto ", word[1:len(word)-1])
			return symbol.Assets{
				Base:  word[1 : len(word)-1],
				Quote: usdt,
			}, true
		}
	}

	return symbol.Assets{}, false
}

func checkFiatNews(handler scraper.ScrapHandler) []symbol.Assets {
	news, err := handler.GetLatestNews()
	if err != nil {
		_, isOkay := err.(*scraper.NoNewsUpdate)
		if !isOkay {
			cmn.LogError.Print(err.Error())
			return nil
		}
	}

	if !strings.Contains(news.Header, "Binance Adds") {
		return nil
	}

	fiatList := make([]string, 0)
	headerWords := strings.Fields(news.Header)
	for _, word := range headerWords {
		if strings.Contains(word, "/") {
			if strings.HasSuffix(word, ",") {
				fiatList = append(fiatList, word[:len(word)-1])
			} else {
				fiatList = append(fiatList, word[:])
			}
			cmn.LogInfo.Print("New fiat ", fiatList[len(fiatList)-1])
		}
	}

	if len(fiatList) == 0 {
		return nil
	}

	return splitSymbols(fiatList, "/")
}

func splitSymbols(symbolsList []string, delimiter string) []symbol.Assets {
	symbolsAssetsList := make([]symbol.Assets, 0)
	for _, symb := range symbolsList {
		idx := strings.Index(symb, delimiter)
		if idx == -1 {
			continue
		}

		symbolsAssetsList = append(symbolsAssetsList, symbol.Assets{
			Base:  symb[:idx],
			Quote: symb[idx+1:],
		})
	}

	if len(symbolsList) == 0 {
		return nil
	}

	return symbolsAssetsList
}
