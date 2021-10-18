package worker

import (
	"github.com/posipaka-trade/bascrap/internal/scraper"
	"github.com/posipaka-trade/posipaka-trade-cmn/log"
	"time"
)

func (worker *Worker) webpageScanner(handler *scraper.ScrapHandler) {
	defer worker.Wg.Done()
	for worker.isWorking {
		time.Sleep(2 * time.Second)

		newsTitle, err := handler.LatestWebsiteNews()
		if err != nil {
			if _, isOkay := err.(*scraper.NoNewsUpdate); !isOkay {
				log.Error.Print(err.Error())
			}
			continue
		}

		worker.messageMutex.Lock()
		if worker.latestHandleMessage != newsTitle {
			log.Info.Print("New announcement on Binance. (Mad website)")
			worker.latestHandleMessage = newsTitle
			worker.newAnnouncement <- newsTitle
		}
		worker.messageMutex.Unlock()
	}
}

func (worker *Worker) telegramScanner(handler *scraper.ScrapHandler) {
	defer worker.Wg.Done()
	time.Sleep(3000 * time.Millisecond)
	for worker.isWorking {
		time.Sleep(time.Millisecond)

		newsTitle, err := handler.LatestTelegramNews()
		if err != nil {
			if _, isOkay := err.(*scraper.NoNewsUpdate); !isOkay {
				log.Error.Print(err.Error())
			}
			continue
		}

		worker.messageMutex.Lock()
		if worker.latestHandleMessage != newsTitle {
			log.Info.Print("New announcement on Binance. (Mad telegram channel)")
			worker.latestHandleMessage = newsTitle
			worker.newAnnouncement <- newsTitle
		}
		worker.messageMutex.Unlock()
	}
}
