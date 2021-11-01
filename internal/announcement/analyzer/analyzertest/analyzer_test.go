package analyzertest

import (
	"github.com/posipaka-trade/bascrap/internal/announcement"
	"github.com/posipaka-trade/bascrap/internal/announcement/analyzer"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAnalyzer(t *testing.T) {
	announDetail := ""
	t.Run("NewCryptoAnnouncement", func(t *testing.T) {
		t.Run("OnlyOneCrypto", func(t *testing.T) {
			announDetail = "Binance Will List Gnosis (GNO)"
			announSymbol, announType := analyzer.AnnouncementSymbol(announDetail)

			assert.Equal(t, announcement.NewCrypto, announType)
			assert.Equal(t, "GNO", announSymbol.Base)
			assert.Equal(t, "USDT", announSymbol.Quote)
		})

		t.Run("SeveralCrypto", func(t *testing.T) {
			announDetail = "Binance Will List Alpaca Finance (ALPACA) and Harvest Finance (FARM) in the Innovation Zone"
			announSymbol, announType := analyzer.AnnouncementSymbol(announDetail)

			assert.Equal(t, announcement.NewCrypto, announType)
			assert.Equal(t, "ALPACA", announSymbol.Base)
			assert.Equal(t, "USDT", announSymbol.Quote)
		})
	})

	t.Run("NewTradingPairAnnouncement", func(t *testing.T) {
		announDetail = "Binance Adds AXS/AUD, AXS/USDT & TVK/BUSD Trading Pairs"
		announSymbol, announType := analyzer.AnnouncementSymbol(announDetail)

		assert.Equal(t, announcement.NewTradingPair, announType)
		assert.Equal(t, "AXS", announSymbol.Base)
		assert.Equal(t, "USDT", announSymbol.Quote)
	})

	t.Run("NewTradingPairOnMarginMarkets", func(t *testing.T) {
		announDetail = "Binance Adds ALPACA, MINA, QUICK & RAY on Isolated Margin, Trade to Win Interest Free Vouchers and Special Merch"
		_, announType := analyzer.AnnouncementSymbol(announDetail)
		assert.Equal(t, announcement.Unknown, announType)
	})

	t.Run("NoSuitableTradingPairs", func(t *testing.T) {
		announDetail = "Binance Adds DOGE/BIDR, ETC/BRL, ETC/UAH Trading Pairs"
		announSymbol, announType := analyzer.AnnouncementSymbol(announDetail)

		assert.Equal(t, announcement.NewTradingPair, announType)
		assert.Equal(t, true, announSymbol.IsEmpty())
	})

	t.Run("UnusefulAnnouncement", func(t *testing.T) {
		announDetail = "FOR, FXS, HNT, KLAY, LPT and PERP Listing on Isolated Margin, $1M Margin 0% Interest Vouchers Giveaway"
		_, announType := analyzer.AnnouncementSymbol(announDetail)
		assert.Equal(t, announcement.Unknown, announType)

		announDetail = "Earn Up to 30% APY on C98 and ARPA with Binance Savings"
		_, announType = analyzer.AnnouncementSymbol(announDetail)
		assert.Equal(t, announcement.Unknown, announType)

		announDetail = "Binance Completes Axie Infinity (AXS) & Smooth Love Potion (SLP) Ronin Network Integration"
		_, announType = analyzer.AnnouncementSymbol(announDetail)
		assert.Equal(t, announcement.Unknown, announType)

		announDetail = "Introducing Mobox (MBOX) on Binance Launchpool! Farm MBOX by Staking BNB, MBOX and BUSD Tokens"
		_, announType = analyzer.AnnouncementSymbol(announDetail)
		assert.Equal(t, announcement.Unknown, announType)
	})
}
