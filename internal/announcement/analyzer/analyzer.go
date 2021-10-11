package analyzer

import (
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/posipaka-trade/bascrap/internal/assets"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
	"strings"
)

func AnnouncementSymbol(newsTitle string) (symbol.Assets, announcement.Type) {
	if strings.Contains(newsTitle, "Binance Will List") {
		return newCryptoSymbol(newsTitle), announcement.NewCrypto
	}

	if strings.Contains(newsTitle, "Binance Adds") &&
		!strings.Contains(newsTitle, "Isolated Margin") {
		return newTradingPairSymbol(newsTitle), announcement.NewTradingPair
	}

	return symbol.Assets{}, announcement.Unknown
}

func newCryptoSymbol(newsTitle string) symbol.Assets {
	headerWords := strings.Fields(newsTitle)
	for _, word := range headerWords {
		if strings.HasPrefix(word, "(") && strings.HasSuffix(word, ")") {
			return symbol.Assets{
				Base:  word[1 : len(word)-1],
				Quote: assets.Usdt,
			}
		}
	}

	return symbol.Assets{}
}

func newTradingPairSymbol(newsTitle string) symbol.Assets {
	pairsStrList := make([]string, 0)
	headerWords := strings.Fields(newsTitle)
	for _, word := range headerWords {
		if strings.Contains(word, "/") {
			if strings.HasSuffix(word, ",") {
				pairsStrList = append(pairsStrList, word[:len(word)-1])
			} else {
				pairsStrList = append(pairsStrList, word)
			}
		}
	}

	if len(pairsStrList) == 0 {
		return symbol.Assets{}
	}

	pairsAssets := splitSymbols(pairsStrList, "/")
	for _, asset := range assets.AnnouncePriority {
		for _, pair := range pairsAssets {
			if pair.Quote == asset {
				return pair
			}
		}
	}

	return symbol.Assets{}
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

	return symbolsAssetsList
}
