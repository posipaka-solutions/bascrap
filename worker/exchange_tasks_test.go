package worker

import (
	"errors"
	"github.com/golang/mock/gomock"
	mockexchangeapi "github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/mock"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/order"
	"github.com/posipaka-trade/posipaka-trade-cmn/exchangeapi/symbol"
	"github.com/posipaka-trade/posipaka-trade-cmn/log"
	"testing"
)

func TestNewFiatAnnouncementBuy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log.Init("", true)

	t.Run("PurchaseWithoutMoneyTransfer", func(t *testing.T) {
		exchange := mockexchangeapi.NewMockApiConnector(ctrl)
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{
			{"KMA", "EUR"}, {"KMA", "BUSD"},
			{"BTC", "EUR"}, {"KMA", "BTC"},
		})

		initialFunds := 22.8
		newSymbolQuantity := 3.14
		newSymbol := symbol.Assets{
			Base:  "KMA",
			Quote: "USDT",
		}

		exchange.EXPECT().SetOrder(gomock.Any()).MaxTimes(1).DoAndReturn(func(parameters order.Parameters) (float64, error) {
			if parameters.Quantity != initialFunds {
				return 0, errors.New("funds value is incorrect")
			}
			if parameters.Side != order.Buy {
				return 0, errors.New("incorrect order side")
			}
			if parameters.Type != order.Market {
				return 0, errors.New("incorrect order type")
			}
			if !parameters.Assets.IsEqual(symbol.Assets{Base: "KMA", Quote: "BUSD"}) {
				return 0, errors.New("incorrect trading pair")
			}
			return newSymbolQuantity, nil
		})

		worker := New(exchange, nil, initialFunds)
		symb, quantity := worker.buyNewFiat(newSymbol)

		if symb.IsEmpty() {
			t.Errorf("New fiat bought symbol is empty. Expected: %s%s", newSymbol.Base, newSymbol.Quote)
			return
		}

		if quantity != newSymbolQuantity {
			t.Errorf("Bought quantity incorrect. Expected: %f", newSymbolQuantity)
		}
	})

	t.Run("TransferMoneyBeforeNewPairBuy", func(t *testing.T) {
		exchange := mockexchangeapi.NewMockApiConnector(ctrl)
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{
			{"KMA", "EUR"}, {"BTC", "EUR"}, {"KMA", "BTC"},
		})

		initialFunds := 22.8
		quantityAfterTransfer := 3.14
		newSymbolQuantity := 5.89
		newSymbol := symbol.Assets{
			Base:  "KMA",
			Quote: "USDT",
		}

		exchange.EXPECT().SetOrder(gomock.Any()).MaxTimes(2).DoAndReturn(func(parameters order.Parameters) (float64, error) {
			if parameters.Side != order.Buy {
				return 0, errors.New("incorrect order side")
			}
			if parameters.Type != order.Market {
				return 0, errors.New("incorrect order type")
			}

			if parameters.Assets.IsEqual(symbol.Assets{Base: "BTC", Quote: "BUSD"}) {
				if parameters.Quantity != initialFunds {
					return 0, errors.New("funds value for transfer is incorrect")
				}
				return quantityAfterTransfer, nil
			} else if parameters.Assets.IsEqual(symbol.Assets{Base: "KMA", Quote: "BTC"}) {
				if parameters.Quantity != quantityAfterTransfer {
					return 0, errors.New("funds value for buy is incorrect")
				}
				return newSymbolQuantity, nil
			} else {
				return 0, errors.New("incorrect trading pair")

			}
		})

		worker := New(exchange, nil, initialFunds)
		symb, quantity := worker.buyNewFiat(newSymbol)

		if symb.IsEmpty() {
			t.Errorf("New fiat bought symbol is empty. Expected: %s%s", newSymbol.Base, newSymbol.Quote)
			return
		}

		if quantity != newSymbolQuantity {
			t.Errorf("Bought quantity incorrect. Expected: %f", newSymbolQuantity)
		}
	})

	t.Run("NewPairQuoteBusd", func(t *testing.T) {
		exchange := mockexchangeapi.NewMockApiConnector(ctrl)
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{
			{"KMA", "USDT"}, {"BTC", "EUR"}, {"KMA", "BTC"},
		})

		initialFunds := 22.8
		quantityAfterTransfer := 3.14
		newSymbolQuantity := 5.89
		newSymbol := symbol.Assets{
			Base:  "KMA",
			Quote: "BUSD",
		}

		exchange.EXPECT().SetOrder(gomock.Any()).MaxTimes(2).DoAndReturn(func(parameters order.Parameters) (float64, error) {
			if parameters.Type != order.Market {
				return 0, errors.New("incorrect order type")
			}
			if parameters.Side == order.Buy {
				if !parameters.Assets.IsEqual(symbol.Assets{Base: "KMA", Quote: "USDT"}) {
					return 0, errors.New("incorrect trading pair for BUY order")
				}
				if parameters.Quantity != quantityAfterTransfer {
					return 0, errors.New("funds value for buy is incorrect")
				}
				return newSymbolQuantity, nil
			} else if parameters.Side == order.Sell {
				if !parameters.Assets.IsEqual(symbol.Assets{Base: "BUSD", Quote: "USDT"}) {
					return 0, errors.New("incorrect trading pair for SELL order")
				}
				if parameters.Quantity != initialFunds {
					return 0, errors.New("funds value for transfer is incorrect")
				}
				return quantityAfterTransfer, nil
			} else {
				return 0, errors.New("incorrect order side")
			}
		})

		worker := New(exchange, nil, initialFunds)
		symb, quantity := worker.buyNewFiat(newSymbol)

		if symb.IsEmpty() {
			t.Errorf("New fiat bought symbol is empty. Expected: %s%s", newSymbol.Base, newSymbol.Quote)
			return
		}

		if quantity != newSymbolQuantity {
			t.Errorf("Bought quantity incorrect. Expected: %f", newSymbolQuantity)
		}
	})

	t.Run("NoSuitablePairForBuy", func(t *testing.T) {
		exchange := mockexchangeapi.NewMockApiConnector(ctrl)
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{
			{"KMA", "LOC"}, {"BTC", "EUR"},
		})
		exchange.EXPECT().SetOrder(gomock.Any()).Times(0)

		worker := New(exchange, nil, 15)
		symb, quantity := worker.buyNewFiat(symbol.Assets{
			Base:  "KMA",
			Quote: "BUSD",
		})

		if !symb.IsEmpty() {
			t.Errorf("Symbol value is not empty. Expected: empty. Actual: %s%s", symb.Base, symb.Quote)
			return
		}
		if quantity != 0 {
			t.Errorf("Quantity value is not zero. Expected: zero. Actual: %f", quantity)
			return
		}
	})

	t.Run("EmptySymbolsList", func(t *testing.T) {
		exchange := mockexchangeapi.NewMockApiConnector(ctrl)
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{})
		exchange.EXPECT().SetOrder(gomock.Any()).Times(0)

		worker := New(exchange, nil, 15)
		symb, quantity := worker.buyNewFiat(symbol.Assets{
			Base:  "KMA",
			Quote: "BUSD",
		})

		if !symb.IsEmpty() {
			t.Errorf("Symbol value is not empty. Expected: empty. Actual: %s%s", symb.Base, symb.Quote)
			return
		}

		if quantity != 0 {
			t.Errorf("Quantity value is not zero. Expected: zero. Actual: %f", quantity)
			return
		}
	})
}

func TestNewCryptoBuy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	price := 69.69
	initialFunds := 22.8
	log.Init("", true)
	gateMock := mockexchangeapi.NewMockApiConnector(ctrl)
	gateMock.EXPECT().GetCurrentPrice(gomock.Any()).Return(price, nil)
	gateMock.EXPECT().SetOrder(gomock.Any()).Return(initialFunds/(price*1.05), nil)

	worker := New(nil, gateMock, initialFunds)
	quantity := worker.buyNewCrypto(symbol.Assets{
		Base:  "TVK",
		Quote: "USDT",
	})
	if quantity != initialFunds/(price*1.05) {
		t.Errorf("Incorrect order quantity. Expected: %f. Actual: %f", initialFunds/(price*1.05), quantity)
		return
	}
}
