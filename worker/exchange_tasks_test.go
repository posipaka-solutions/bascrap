package worker

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/posipaka-trade/bascrap/internal/announcement"
	mockexchangeapi "github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/mock"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/order"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
	"github.com/posipaka-trade/posipaka-trade-cmn/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewFiatAnnouncementBuy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	var haggler hagglingParameters
	log.Init("", true)

	t.Run("PurchaseWithoutMoneyTransfer", func(t *testing.T) {
		exchange := mockexchangeapi.NewMockApiConnector(ctrl)
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{
			{"KMA", "EUR"}, {"KMA", "BUSD"},
			{"BTC", "EUR"}, {"KMA", "BTC"},
		})

		initialFunds := 22.8
		orderInfo := order.OrderInfo{
			Price:    124.2,
			Quantity: 1548.58,
		}
		newSymbol := symbol.Assets{
			Base:  "KMA",
			Quote: "USDT",
		}

		exchange.EXPECT().SetOrder(gomock.Any()).MaxTimes(1).DoAndReturn(func(parameters order.Parameters) (order.OrderInfo, error) {
			if parameters.Quantity != initialFunds {
				return order.OrderInfo{}, errors.New("funds value is incorrect")
			}
			if parameters.Side != order.Buy {
				return order.OrderInfo{}, errors.New("incorrect order side")
			}
			if parameters.Type != order.Market {
				return order.OrderInfo{}, errors.New("incorrect order type")
			}
			if !parameters.Assets.IsEqual(symbol.Assets{Base: "KMA", Quote: "BUSD"}) {
				return order.OrderInfo{}, errors.New("incorrect trading pair")
			}
			return orderInfo, nil
		})

		worker := New(exchange, nil, initialFunds, false)
		haggler = worker.buyNewFiat(newSymbol)
		assert.Equal(t, symbol.Assets{Base: "KMA", Quote: "BUSD"}, haggler.symbol)
		assert.Equal(t, orderInfo.Quantity, haggler.boughtQuantity)
		assert.Equal(t, orderInfo.Price, haggler.boughtPrice)
		assert.Equal(t, announcement.NewTradingPair, haggler.announcementType)
	})

	t.Run("TransferMoneyBeforeNewPairBuy", func(t *testing.T) {
		exchange := mockexchangeapi.NewMockApiConnector(ctrl)
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{
			{"KMA", "EUR"},
			{"BTC", "EUR"},
			{"KMA", "BTC"},
		})

		initialFunds := 22.8
		newSymbol := symbol.Assets{
			Base:  "KMA",
			Quote: "USDT",
		}
		transferInfo := order.OrderInfo{
			Price:    123.123,
			Quantity: 123.456,
		}
		buyInfo := order.OrderInfo{
			Price:    0.000154,
			Quantity: 12.058453,
		}

		exchange.EXPECT().SetOrder(gomock.Any()).MaxTimes(2).DoAndReturn(func(parameters order.Parameters) (order.OrderInfo, error) {
			if parameters.Side != order.Buy {
				return order.OrderInfo{}, errors.New("incorrect order side")
			}
			if parameters.Type != order.Market {
				return order.OrderInfo{}, errors.New("incorrect order type")
			}

			if parameters.Assets.IsEqual(symbol.Assets{Base: "BTC", Quote: "BUSD"}) {
				if parameters.Quantity != initialFunds {
					return order.OrderInfo{}, errors.New("funds value for transfer is incorrect")
				}
				return transferInfo, nil
			} else if parameters.Assets.IsEqual(symbol.Assets{Base: "KMA", Quote: "BTC"}) {
				if parameters.Quantity != transferInfo.Quantity {
					return order.OrderInfo{}, errors.New("funds value for buy is incorrect")
				}
				return buyInfo, nil
			}
			return order.OrderInfo{}, errors.New("incorrect trading pair")
		})

		worker := New(exchange, nil, initialFunds, false)
		haggler = worker.buyNewFiat(newSymbol)

		assert.Equal(t, symbol.Assets{Base: "KMA", Quote: "BTC"}, haggler.symbol)
		assert.Equal(t, buyInfo.Quantity, haggler.boughtQuantity)
		assert.Equal(t, buyInfo.Price, haggler.boughtPrice)
		assert.Equal(t, announcement.NewTradingPair, haggler.announcementType)
	})

	t.Run("NewPairQuoteBusd", func(t *testing.T) {
		exchange := mockexchangeapi.NewMockApiConnector(ctrl)
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{
			{"KMA", "USDT"},
			{"BTC", "EUR"},
			{"KMA", "BTC"},
		})

		initialFunds := 22.8
		newSymbol := symbol.Assets{
			Base:  "KMA",
			Quote: "BUSD",
		}
		transferInfo := order.OrderInfo{
			Price:    123.456,
			Quantity: 678.00458,
		}
		buyInfo := order.OrderInfo{
			Price:    0.054965,
			Quantity: 0.3231,
		}

		exchange.EXPECT().SetOrder(gomock.Any()).MaxTimes(2).DoAndReturn(func(parameters order.Parameters) (order.OrderInfo, error) {
			if parameters.Type != order.Market {
				return order.OrderInfo{}, errors.New("incorrect order type")
			}
			if parameters.Side == order.Buy {
				if !parameters.Assets.IsEqual(symbol.Assets{Base: "KMA", Quote: "USDT"}) {
					return order.OrderInfo{}, errors.New("incorrect trading pair for BUY order")
				}
				if parameters.Quantity != transferInfo.Quantity {
					return order.OrderInfo{}, errors.New("funds value for buy is incorrect")
				}
				return buyInfo, nil
			} else if parameters.Side == order.Sell {
				if !parameters.Assets.IsEqual(symbol.Assets{Base: "BUSD", Quote: "USDT"}) {
					return order.OrderInfo{}, errors.New("incorrect trading pair for SELL order")
				}
				if parameters.Quantity != initialFunds {
					return order.OrderInfo{}, errors.New("funds value for transfer is incorrect")
				}
				return transferInfo, nil
			}

			return order.OrderInfo{}, errors.New("incorrect order side")
		})

		worker := New(exchange, nil, initialFunds, false)
		haggler = worker.buyNewFiat(newSymbol)

		assert.Equal(t, symbol.Assets{Base: "KMA", Quote: "USDT"}, haggler.symbol)
		assert.Equal(t, buyInfo.Quantity, haggler.boughtQuantity)
		assert.Equal(t, buyInfo.Price, haggler.boughtPrice)
		assert.Equal(t, announcement.NewTradingPair, haggler.announcementType)
	})

	t.Run("NoSuitablePairForBuy", func(t *testing.T) {
		exchange := mockexchangeapi.NewMockApiConnector(ctrl)
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{
			{"KMA", "LOC"},
			{"BTC", "EUR"},
		})
		smb := symbol.Assets{
			Base:  "KMA",
			Quote: "BUSD",
		}

		exchange.EXPECT().SetOrder(gomock.Any()).Times(0)
		worker := New(exchange, nil, 15, false)
		haggler = worker.buyNewFiat(smb)

		if !haggler.symbol.IsEmpty() {
			t.Errorf("Symbol value is not empty. Expected: empty. Actual: %s%s", haggler.symbol.Base, haggler.symbol.Quote)
			return
		}
		if haggler.boughtQuantity != 0 {
			t.Errorf("Quantity value is not zero. Expected: zero. Actual: %f", haggler.boughtQuantity)
			return
		}
	})

	t.Run("EmptySymbolsList", func(t *testing.T) {
		exchange := mockexchangeapi.NewMockApiConnector(ctrl)
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{})
		exchange.EXPECT().SetOrder(gomock.Any()).Times(0)

		worker := New(exchange, nil, 15, false)
		haggler = worker.buyNewFiat(symbol.Assets{
			Base:  "KMA",
			Quote: "BUSD",
		})

		if !haggler.symbol.IsEmpty() {
			t.Errorf("Symbol value is not empty. Expected: empty. Actual: %s%s", haggler.symbol.Base, haggler.symbol.Quote)
			return
		}

		if haggler.boughtQuantity != 0 {
			t.Errorf("Quantity value is not zero. Expected: zero. Actual: %f", haggler.boughtQuantity)
			return
		}
	})
}

func TestNewCryptoBuy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log.Init("", true)

	price := 69.69
	initialFunds := 22.8
	gateMock := mockexchangeapi.NewMockApiConnector(ctrl)
	gateMock.EXPECT().GetCurrentPrice(gomock.Any()).Return(price, nil)
	gateMock.EXPECT().SetOrder(gomock.Any()).MaxTimes(1).DoAndReturn(func(parameters order.Parameters) (order.OrderInfo, error) {
		if parameters.Type != order.Limit {
			return order.OrderInfo{}, errors.New("incorrect order type")
		}
		if parameters.Side == order.Buy {
			if !parameters.Assets.IsEqual(symbol.Assets{Base: "TVK", Quote: "USDT"}) {
				return order.OrderInfo{}, errors.New("incorrect trading pair for BUY order")
			}
			if parameters.Quantity != initialFunds/price {
				return order.OrderInfo{}, errors.New("funds value for buy is incorrect")
			}
			if parameters.Price != price*1.15 {
				return order.OrderInfo{}, errors.New("incorrect limit order price")
			}
			return order.OrderInfo{
				Price:    price * 1.1,
				Quantity: initialFunds / (price * 1.1),
			}, nil
		}

		return order.OrderInfo{}, errors.New("incorrect order side")
	})

	worker := New(nil, gateMock, initialFunds, false)
	haggle, err := worker.buyNewCrypto(symbol.Assets{Base: "TVK", Quote: "USDT"})

	assert.Nil(t, err)
	assert.Equal(t, symbol.Assets{Base: "TVK", Quote: "USDT"}, haggle.symbol)
	assert.Equal(t, initialFunds/(price*1.1), haggle.boughtQuantity)
	assert.Equal(t, price*1.1, haggle.boughtPrice)
	assert.Equal(t, announcement.NewCrypto, haggle.announcementType)
}
