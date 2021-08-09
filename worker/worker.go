package worker

import (
	"github.com/posipaka-trade/bascrap/scraper"
	"github.com/posipaka-trade/bascrap/scraper/announcement"
)

func StartMonitoring() {
	cryptoListingHandler := scraper.New(announcement.NewCryptoListing)
	fiatListingHandler := scraper.New(announcement.NewFiatListing)

	for {

	}
}
