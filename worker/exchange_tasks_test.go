package worker

import (
	"errors"
	"github.com/golang/mock/gomock"
	cmn "github.com/posipaka-trade/posipaka-trade-cmn"
	mockexchangeapi "github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/mock"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/order"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
	"testing"
)

func TestFiatBuy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmn.InitLoggers("")

	exchange := mockexchangeapi.NewMockApiConnector(ctrl)
	worker := New(exchange, nil, 15)

	t.Run("Purchase without money transfer", func(t *testing.T) {
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{
			symbol.Assets{
				Base:  "KMA",
				Quote: "EUR",
			},
			symbol.Assets{
				Base:  "KMA",
				Quote: "BUSD",
			},
			symbol.Assets{
				Base:  "BTC",
				Quote: "EUR",
			},
			symbol.Assets{
				Base:  "KMA",
				Quote: "BTC",
			},
		})

		exchange.EXPECT().SetOrder(gomock.Any()).MaxTimes(1).DoAndReturn(func(parameters order.Parameters) (float64, error) {
			if parameters.Assets.IsEqual(symbol.Assets{
				Base:  "KMA",
				Quote: "BUSD",
			}) {
				return 123, nil
			} else {
				return 0, errors.New("not suitable pair")
			}
		})

		res := worker.buyNewFiat(symbol.Assets{
			Base:  "KMA",
			Quote: "USDT",
		})

		if !res {
			t.Error("Buying failed")
			return
		}
	})

	t.Run("Transfer money before new pair buy", func(t *testing.T) {
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{
			symbol.Assets{
				Base:  "KMA",
				Quote: "EUR",
			},
			symbol.Assets{
				Base:  "BTC",
				Quote: "EUR",
			},
			symbol.Assets{
				Base:  "KMA",
				Quote: "BTC",
			},
		})

		exchange.EXPECT().SetOrder(gomock.Any()).MaxTimes(2).DoAndReturn(func(parameters order.Parameters) (float64, error) {
			if parameters.Assets.IsEqual(symbol.Assets{
				Base:  "BTC",
				Quote: "BUSD",
			}) || parameters.Assets.IsEqual(symbol.Assets{
				Base:  "KMA",
				Quote: "BTC",
			}) {
				return 123, nil
			} else {
				return 0, errors.New("not suitable pair")
			}
		})

		res := worker.buyNewFiat(symbol.Assets{
			Base:  "KMA",
			Quote: "USDT",
		})

		if !res {
			t.Error("Buying failed")
			return
		}
	})

	t.Run("NewPairQuoteBusd", func(t *testing.T) {
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{
			symbol.Assets{
				Base:  "KMA",
				Quote: "USDT",
			},
			symbol.Assets{
				Base:  "BTC",
				Quote: "EUR",
			},
			symbol.Assets{
				Base:  "KMA",
				Quote: "BTC",
			},
		})

		exchange.EXPECT().SetOrder(gomock.Any()).MaxTimes(2).DoAndReturn(func(parameters order.Parameters) (float64, error) {
			if parameters.Assets.IsEqual(symbol.Assets{
				Base:  "BUSD",
				Quote: "USDT",
			}) || parameters.Assets.IsEqual(symbol.Assets{
				Base:  "KMA",
				Quote: "USDT",
			}) {
				return 123, nil
			} else {
				return 0, errors.New("not suitable pair, base: " + parameters.Assets.Base + " quote: " + parameters.Assets.Quote)
			}
		})

		res := worker.buyNewFiat(symbol.Assets{
			Base:  "KMA",
			Quote: "BUSD",
		})

		if !res {
			t.Error("Buying failed")
			return
		}
	})
}
