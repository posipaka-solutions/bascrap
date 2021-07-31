package announcement

type Type int

const NewCryptoListingLink = "https://www.binance.com/en/support/announcement/c-48?navId=48"

const (
	NewCryptoListing Type = 1 << iota
	//NewFiatListing = 1 << iota
	//NewTradingPair = 1 << iota
	//All = 1 << iota
)
