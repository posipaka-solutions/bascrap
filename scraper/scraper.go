package scraper

import (
	"github.com/posipaka-trade/bascrap/scraper/announcement"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"io/ioutil"
	"net/http"
)

type News struct {
	Type   announcement.Type
	Header string
}

type ScrapHandler struct {
	LatestNewsHeader      string
	announceForMonitoring announcement.Type
}

func New(announceType announcement.Type) ScrapHandler {
	defer cmn.LogInfo.Print("scraper -> New Scraper instance created.")
	return ScrapHandler{
		LatestNewsHeader:      latestNewsHeader(),
		announceForMonitoring: announceType,
	}
}

func closeBody(response *http.Response) {
	if err := response.Body.Close(); err != nil {
		panic(err.Error())
	}
}

func latestNewsHeader() string {
	response, err := http.Get(announcement.NewCryptoListingLink)
	if err != nil {
		cmn.LogError.Print("scraper -> ", err.Error())
		return ""
	}

	if response.StatusCode/100 != 2 {
		cmn.LogError.Print("scraper -> ", response.Status)
		return ""
	}

	defer closeBody(response)
	html, err := ioutil.ReadAll(response.Body)
	if err != nil {
		cmn.LogError.Print("scraper -> ", err.Error())
		return ""
	}

	return string(html)
}
