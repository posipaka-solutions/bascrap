package worker

import (
	"github.com/posipaka-trade/bascrap/scraper"
	"github.com/posipaka-trade/bascrap/scraper/announcement"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"io/ioutil"
	"time"
)

func StartMonitoring() {
	for {
		handler := scraper.New(announcement.NewCryptoListing)
		err := ioutil.WriteFile("./announcement.html", []byte(handler.LatestNewsHeader), 0666)
		if err != nil {
			panic(err)
		}
		cmn.LogInfo.Print("New announcement page stored.")
		time.Sleep(1 * time.Minute)
	}
}
