package trading

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

/**
 * FakeMarket
 *
 * A stub to use for tests that interact with markets.
 */

type FakeMarket struct {
	Name             string
	ExistsValue      bool
	CurrentPrice     float64
	SummaryData      []*SummaryData
	PendingOrders    []*Order
	TriggerBuyError  bool
	TriggerSellError bool
}

func (market *FakeMarket) GetName() string {
	return market.Name
}

func (market *FakeMarket) GetCurrency() string {
	return strings.Split(market.Name, "_")[1]
}

func (market *FakeMarket) Exists() bool {
	return market.ExistsValue
}

func (market *FakeMarket) GetCurrentPrice() (float64, error) {
	return market.CurrentPrice, nil
}

func (market *FakeMarket) GetSummaryData(startTime, endTime int64) ([]*SummaryData, error) {
	return market.SummaryData, nil
}

func (market *FakeMarket) GetPendingOrders() ([]*Order, error) {
	return market.PendingOrders, nil
}

func (market *FakeMarket) Buy(order *Order) error {
	if market.TriggerBuyError {
		return fmt.Errorf("fake buy error")
	}

	order.Id = strconv.FormatInt(rand.Int63(), 10)
	return nil
}

func (market *FakeMarket) Sell(order *Order) error {
	if market.TriggerSellError {
		return fmt.Errorf("fake sell error")
	}

	order.Id = strconv.FormatInt(rand.Int63(), 10)
	return nil
}

/**
 * FakeExchange
 *
 * A stub to use for tests that interact with exchanges.
 */

type FakeExchange struct {
	Market   Market
	Balances map[string]*Balance
}

func (exchange *FakeExchange) GetMarket(marketName string) (Market, error) {
	return exchange.Market, nil
}

func (exchange *FakeExchange) GetBalance(currency string) (*Balance, error) {
	balance, _ := exchange.Balances[currency]
	return balance, nil
}
