package scraper

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"net/http"
)

const binanceUrl = "https://www.binance.com"
const (
	announcementDiv = ".css-6f91y1"
	newsListDiv     = ".css-vurnku"
)

func closeBody(response *http.Response) {
	if err := response.Body.Close(); err != nil {
		panic(err.Error())
	}
}

type ScrapHandler struct {
	latestAnnounce        announcement.Details
	announceForMonitoring announcement.Type
}

func New(announceType announcement.Type) ScrapHandler {
	handler := ScrapHandler{
		announceForMonitoring: announceType,
	}

	var err error
	handler.latestAnnounce, err = handler.GetLatestAnnounce()
	if err != nil {
		panic(err.Error())
	}

	return handler
}

func (handler ScrapHandler) GetLatestAnnounce() (announcement.Details, error) {
	res, err := http.Get(announcement.Links[handler.announceForMonitoring])
	if err != nil {
		return announcement.Details{}, err
	}
	defer closeBody(res)

	if res.StatusCode != 200 {
		return announcement.Details{}, errors.New(fmt.Sprintf("[scraper] -> Error when get html page. %d: %s", res.StatusCode, res.Status))
	}

	newsInfo, err := parseHtml(res, handler.announceForMonitoring)
	if err != nil {
		return announcement.Details{}, err
	}

	if handler.latestAnnounce.Equal(newsInfo) {
		return announcement.Details{}, &NoNewsUpdate{}
	}

	handler.latestAnnounce = newsInfo
	return newsInfo, nil
}

func parseHtml(response *http.Response, annType announcement.Type) (announcement.Details, error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return announcement.Details{}, err
	}

	div := doc.Find(announcementDiv)
	if div == nil {
		return announcement.Details{}, errors.New("[scraper] -> Announcement div not found")
	}
	div = div.Find(newsListDiv)
	if div == nil {
		return announcement.Details{}, errors.New("[scraper] -> Info list div not found")
	}

	announce := announcement.Details{}
	div.Find("a").Each(func(i int, s *goquery.Selection) {
		if i != 0 {
			return
		}

		announce.Type = annType
		announce.Header = s.Text()
		link, exist := s.Attr("href")
		if exist {
			announce.Link = fmt.Sprint(binanceUrl, link)
		}
	})

	return announce, nil
}
