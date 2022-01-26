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
	var haglgler hagglingParameters
	log.Init("", true)

	t.Run("PurchaseWithoutMoneyTransfer", func(t *testing.T) {
		exchange := mockexchangeapi.NewMockApiConnector(ctrl)
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{
			{"KMA", "EUR"}, {"KMA", "BUSD"},
			{"BTC", "EUR"}, {"KMA", "BTC"},
		})
		initialFunds := 22.8
		newSymbolQuantity := 3.14
		price := initialFunds / newSymbolQuantity
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
		exchange.EXPECT().GetCurrentPrice(gomock.Any()).Return(price, nil)

		worker := New(exchange, nil, initialFunds, false)
		// wer
		haglgler = worker.buyNewFiat(newSymbol)

		if haglgler.symbol.IsEmpty() {
			t.Errorf("New fiat bought symbol is empty. Expected: %s%s", newSymbol.Base, newSymbol.Quote)
			return
		}

		if haglgler.boughtQuantity != newSymbolQuantity {
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
		price := initialFunds / quantityAfterTransfer
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
		exchange.EXPECT().GetCurrentPrice(gomock.Any()).Return(price, nil)

		worker := New(exchange, nil, initialFunds, false)
		haglgler = worker.buyNewFiat(newSymbol)

		if haglgler.symbol.IsEmpty() {
			t.Errorf("New fiat bought symbol is empty. Expected: %s%s", newSymbol.Base, newSymbol.Quote)
			return
		}

		if haglgler.boughtQuantity != newSymbolQuantity {
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
		price := initialFunds / quantityAfterTransfer
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
		exchange.EXPECT().GetCurrentPrice(gomock.Any()).Return(price, nil)

		worker := New(exchange, nil, initialFunds, false)
		haglgler = worker.buyNewFiat(newSymbol)

		if haglgler.symbol.IsEmpty() {
			t.Errorf("New fiat bought symbol is empty. Expected: %s%s", newSymbol.Base, newSymbol.Quote)
			return
		}

		if haglgler.boughtQuantity != newSymbolQuantity {
			t.Errorf("Bought quantity incorrect. Expected: %f", newSymbolQuantity)
		}
	})

	t.Run("NoSuitablePairForBuy", func(t *testing.T) {
		exchange := mockexchangeapi.NewMockApiConnector(ctrl)
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{
			{"KMA", "LOC"}, {"BTC", "EUR"},
		})
		smb := symbol.Assets{
			Base:  "KMA",
			Quote: "BUSD"}

		exchange.EXPECT().SetOrder(gomock.Any()).Times(0)
		worker := New(exchange, nil, 15, false)
		haglgler = worker.buyNewFiat(smb)

		if !haglgler.symbol.IsEmpty() {
			t.Errorf("Symbol value is not empty. Expected: empty. Actual: %s%s", haglgler.symbol.Base, haglgler.symbol.Quote)
			return
		}
		if haglgler.boughtQuantity != 0 {
			t.Errorf("Quantity value is not zero. Expected: zero. Actual: %f", haglgler.boughtQuantity)
			return
		}
	})

	t.Run("EmptySymbolsList", func(t *testing.T) {
		exchange := mockexchangeapi.NewMockApiConnector(ctrl)
		exchange.EXPECT().GetSymbolsList().Return([]symbol.Assets{})
		exchange.EXPECT().SetOrder(gomock.Any()).Times(0)

		worker := New(exchange, nil, 15, false)
		haglgler = worker.buyNewFiat(symbol.Assets{
			Base:  "KMA",
			Quote: "BUSD",
		})

		if !haglgler.symbol.IsEmpty() {
			t.Errorf("Symbol value is not empty. Expected: empty. Actual: %s%s", haglgler.symbol.Base, haglgler.symbol.Quote)
			return
		}

		if haglgler.boughtQuantity != 0 {
			t.Errorf("Quantity value is not zero. Expected: zero. Actual: %f", haglgler.boughtQuantity)
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

	worker := New(nil, gateMock, initialFunds, false)
	hagl, err := worker.buyNewCrypto(symbol.Assets{
		Base:  "TVK",
		Quote: "USDT",
	})
	if err != nil {

	}
	if hagl.boughtQuantity != initialFunds/(price*1.05) {
		t.Errorf("Incorrect order quantity. Expected: %f. Actual: %f", initialFunds/(price*1.05), hagl.boughtQuantity)
		return
	}
}
