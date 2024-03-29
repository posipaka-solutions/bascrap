package announcement

type Type int

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
