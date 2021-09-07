package announcement

type Type int

const (
	NewCryptoListingUrl = "https://www.binance.com/en/support/announcement/c-48?navId=48"
	NewFiatListingUrl   = "https://www.binance.com/en/support/announcement/c-50?navId=50"
)

const (
	Unknown        Type = 1 << iota
	NewCrypto           = 1 << iota
	NewTradingPair      = 1 << iota
)

var TypeAlias = map[Type]string{
	Unknown:        "unknown",
	NewCrypto:      "newCrypto",
	NewTradingPair: "newTradingPair",
}

type Details struct {
	SourceUrl string
	Header    string
	Link      string
}

func (details Details) Equal(otherDetails Details) bool {
	return details.Header == otherDetails.Header &&
		details.Link == otherDetails.Link &&
		details.SourceUrl == otherDetails.SourceUrl
}
