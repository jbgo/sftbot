package trading

import (
	"strings"
)

/**
 * FakeMarket
 *
 * A stub to use for tests that interact with markets.
 */

type FakeMarket struct {
	Name          string
	ExistsValue   bool
	CurrentPrice  float64
	SummaryData   []*SummaryData
	TradeHistory  []*Trade
	PendingOrders []*Order
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

func (market *FakeMarket) GetTradeHistory(startTime, endTime int64) ([]*Trade, error) {
	return market.TradeHistory, nil
}

func (market *FakeMarket) GetPendingOrders() ([]*Order, error) {
	return market.PendingOrders, nil
}

/**
 * FakeExchange
 *
 * A stub to use for tests that interact with exchanges.
 */

type FakeExchange struct {
	Market   Market
	Ticker   map[string][]*TickerEntry
	Balances map[string]*Balance
}

func (exchange *FakeExchange) GetMarket(marketName string) (Market, error) {
	return exchange.Market, nil
}

func (exchange *FakeExchange) GetTicker(marketName string) ([]*TickerEntry, error) {
	ticker, _ := exchange.Ticker[marketName]
	return ticker, nil
}

func (exchange *FakeExchange) GetBalance(currency string) (*Balance, error) {
	balance, _ := exchange.Balances[currency]
	return balance, nil
}
