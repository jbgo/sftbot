package trading

import ()

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
