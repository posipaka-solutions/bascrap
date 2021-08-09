package scraper

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/posipaka-trade/bascrap/scraper/announcement"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	"net/http"
)

func closeBody(response *http.Response) {
	if err := response.Body.Close(); err != nil {
		panic(err.Error())
	}
}

type ScrapHandler struct {
	LatestNewsHeader      News
	announceForMonitoring announcement.Type
}

func New(announceType announcement.Type) ScrapHandler {
	defer cmn.LogInfo.Print("scraper -> New Scraper instance created.")
	handler := ScrapHandler{
		announceForMonitoring: announceType,
	}

	var err error
	handler.LatestNewsHeader, err = handler.GetLatestNews()
	if err != nil {
		panic(err.Error())
	}

	return handler
}

func (handler ScrapHandler) GetLatestNews() (News, error) {
	res, err := http.Get(announcement.Links[handler.announceForMonitoring])
	if err != nil {
		return News{}, err
	}
	defer closeBody(res)

	if res.StatusCode != 200 {
		return News{}, errors.New(fmt.Sprintf("[scraper] -> Error when get html page. %d: %s", res.StatusCode, res.Status))
	}

	news, err := parseHtml(res, handler.announceForMonitoring)
	if err != nil {
		return News{}, err
	}

	if handler.LatestNewsHeader.Equal(news) {
		return News{}, &NoNewsUpdate{}
	}

	handler.LatestNewsHeader = news
	return news, nil
}

func parseHtml(response *http.Response, annType announcement.Type) (News, error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return News{}, err
	}

	div := doc.Find(announcementDiv)
	if div == nil {
		return News{}, errors.New("[scraper] -> Announcement div not found")
	}
	div = div.Find(newsListDiv)
	if div == nil {
		return News{}, errors.New("[scraper] -> News list div not found")
	}

	var news News
	div.Find("a").Each(func(i int, s *goquery.Selection) {
		if i != 0 {
			return
		}

		news.Type = annType
		news.Header = s.Text()
		link, exist := s.Attr("href")
		if exist {
			news.Link = fmt.Sprint(binanceUrl, link)
		}
	})

	return news, nil
}
