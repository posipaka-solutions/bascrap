package assets

const (
	Usdt = "USDT"
	Busd = "BUSD"
	Eur  = "EUR"
	Aud  = "AUD"
	Rub  = "RUB"
	Btc  = "BTC"
	Eth  = "ETH"
	Bnb  = "BNB"
)

var AnnouncePriority = []string{
	Usdt,
	Busd,
	Eur,
	Aud,
	Rub,
}

var BuyPriority = []string{
	Usdt,
	Busd,
	Btc,
	Eth,
	Bnb,
}
