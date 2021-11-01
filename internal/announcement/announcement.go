package announcement

type Type int

const (
	Unknown        Type = 1 << iota
	NewCrypto      Type = 1 << iota
	NewTradingPair Type = 1 << iota
)

var TypeAlias = map[Type]string{
	Unknown:        "unknown",
	NewCrypto:      "newCrypto",
	NewTradingPair: "newTradingPair",
}
