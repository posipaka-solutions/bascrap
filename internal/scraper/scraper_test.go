package scraper

import (
	"bytes"
	"os"
	"testing"
)

func BenchmarkAnnouncementPageScraping(b *testing.B) {
	htmlBody, err := os.ReadFile("testdata/new_crypto_listing.html")
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		details, err := parseHtml(bytes.NewReader(htmlBody))
		if err != nil {
			b.Error(err)
		}

		if details.Header != "Binance Will List Flow (FLOW)" {
			b.Errorf("Incorrect latest news header. Expected: %s\tActual %s",
				"Binance Will List Flow (FLOW)", details.Header)
		}
	}
}
