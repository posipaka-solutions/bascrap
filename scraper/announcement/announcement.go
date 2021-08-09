package announcement

type Type int

var Links = map[Type]string{
	NewCryptoListing: "https://www.binance.com/en/support/announcement/c-48?navId=48",
	NewFiatListing:   "https://www.binance.com/en/support/announcement/c-50?navId=50",
}

const (
	NewCryptoListing Type = 1 << iota
	NewFiatListing        = 1 << iota
)
