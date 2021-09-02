package analyzer

import (
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"testing"
)

func TestAnalyzer(t *testing.T) {
	announDetail := announcement.Details{}
	t.Run("NewCryptoAnnouncement", func(t *testing.T) {
		announDetail.SourceUrl = announcement.NewCryptoListingUrl
		t.Run("OnlyOneCrypto", func(t *testing.T) {
			announDetail.Header = "Binance Will List Gnosis (GNO)"
			announSymbol, announType := AnnouncementSymbol(announDetail)
			if announType != announcement.NewCrypto {
				t.Errorf("Incorrect type recognition. Expacted: %s\tActual: %s",
					announcement.TypeAlias[announcement.NewCrypto], announcement.TypeAlias[announType])
			}

			if announSymbol.Base != "GNO" {
				t.Errorf("Incorrect BASE symbol. Expected: %s\tActual:%s", "GNO", announSymbol.Base)
			}

			if announSymbol.Quote != "USDT" {
				t.Errorf("Incorrect QUOTE symbol. Expected: %s\tActual:%s", "USDT", announSymbol.Quote)
			}
		})

		t.Run("SeveralCrypto", func(t *testing.T) {
			announDetail.Header = "Binance Will List Alpaca Finance (ALPACA) and Harvest Finance (FARM) in the Innovation Zone"
			announSymbol, announType := AnnouncementSymbol(announDetail)
			if announType != announcement.NewCrypto {
				t.Errorf("Incorrect type recognition. Expacted: %s\tActual: %s",
					announcement.TypeAlias[announcement.NewCrypto], announcement.TypeAlias[announType])
			}

			if announSymbol.Base != "ALPACA" {
				t.Errorf("Incorrect BASE symbol. Expected: %s\tActual: %s", "ALPACA", announSymbol.Base)
			}

			if announSymbol.Quote != "USDT" {
				t.Errorf("Incorrect QUOTE symbol. Expected: %s\tActual: %s", "USDT", announSymbol.Quote)
			}
		})
	})

	t.Run("NewTradingPairAnnouncement", func(t *testing.T) {
		announDetail.SourceUrl = announcement.NewFiatListingUrl
		announDetail.Header = "Binance Adds AXS/AUD, AXS/USDT & TVK/BUSD Trading Pairs"
		announSymbol, announType := AnnouncementSymbol(announDetail)
		if announType != announcement.NewTradingPair {
			t.Errorf("Incorrect announcement type. Expected: %s\tActual: %s",
				announcement.TypeAlias[announcement.NewTradingPair], announcement.TypeAlias[announType])
		}

		if announSymbol.Base != "AXS" && announSymbol.Quote != "USDT" {
			t.Errorf("Incorret trading pair sellection. Expected: %s\tActual: %s/%s", "AXS/USDT",
				announSymbol.Base, announSymbol.Quote)
		}
	})

	t.Run("NewTradingPairOnMarginMarkets", func(t *testing.T) {
		announDetail.SourceUrl = announcement.NewCryptoListingUrl
		announDetail.Header = "Binance Adds ALPACA, MINA, QUICK & RAY on Isolated Margin, Trade to Win Interest Free Vouchers and Special Merch"
		_, announType := AnnouncementSymbol(announDetail)
		if announType != announcement.Unknown {
			t.Errorf("Incorrect announcement type. Expected: %s\tActual: %s",
				announcement.TypeAlias[announcement.Unknown], announcement.TypeAlias[announType])
		}
	})

	t.Run("NoSuitableTradingPairs", func(t *testing.T) {
		announDetail.SourceUrl = announcement.NewFiatListingUrl
		announDetail.Header = "Binance Adds DOGE/BIDR, ETC/BRL, ETC/UAH Trading Pairs"
		announSymbol, announType := AnnouncementSymbol(announDetail)
		if announType != announcement.NewTradingPair {
			t.Errorf("Incorrect announcement type. Expected: %s\tActual: %s",
				announcement.TypeAlias[announcement.NewTradingPair], announcement.TypeAlias[announType])
		}

		if !announSymbol.IsEmpty() {
			t.Errorf("Selected unsuitable pair from list of new ones. Symbols: %s/%s",
				announSymbol.Base, announSymbol.Quote)
		}
	})

	t.Run("UnusefulAnnouncement", func(t *testing.T) {
		announDetail.SourceUrl = announcement.NewCryptoListingUrl
		announDetail.Header = "FOR, FXS, HNT, KLAY, LPT and PERP Listing on Isolated Margin, $1M Margin 0% Interest Vouchers Giveaway"
		_, announType := AnnouncementSymbol(announDetail)
		if announType != announcement.Unknown {
			t.Errorf("Incorrect announcement type. Expected: %s\tActual: %s",
				announcement.TypeAlias[announcement.Unknown], announcement.TypeAlias[announType])
		}

		announDetail.Header = "Earn Up to 30% APY on C98 and ARPA with Binance Savings"
		_, announType = AnnouncementSymbol(announDetail)
		if announType != announcement.Unknown {
			t.Errorf("Incorrect announcement type. Expected: %s\tActual: %s",
				announcement.TypeAlias[announcement.Unknown], announcement.TypeAlias[announType])
		}

		announDetail.Header = "Binance Completes Axie Infinity (AXS) & Smooth Love Potion (SLP) Ronin Network Integration"
		_, announType = AnnouncementSymbol(announDetail)
		if announType != announcement.Unknown {
			t.Errorf("Incorrect announcement type. Expected: %s\tActual: %s",
				announcement.TypeAlias[announcement.Unknown], announcement.TypeAlias[announType])
		}

		announDetail.Header = "Introducing Mobox (MBOX) on Binance Launchpool! Farm MBOX by Staking BNB, MBOX and BUSD Tokens"
		_, announType = AnnouncementSymbol(announDetail)
		if announType != announcement.Unknown {
			t.Errorf("Incorrect announcement type. Expected: %s\tActual: %s",
				announcement.TypeAlias[announcement.Unknown], announcement.TypeAlias[announType])
		}
	})
}
