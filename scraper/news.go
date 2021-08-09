package scraper

import "github.com/posipaka-trade/bascrap/scraper/announcement"

const binanceUrl = "https://www.binance.com"
const (
	announcementDiv = ".css-6f91y1"
	newsListDiv     = ".css-vurnku"
)

type News struct {
	Type   announcement.Type
	Header string
	Link   string
}

func (news News) Equal(otherNews News) bool {
	return news.Header == otherNews.Header &&
		news.Link == otherNews.Link &&
		news.Type == otherNews.Type
}
