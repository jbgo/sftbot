package trading

import ()

type PlxExchange struct {
}

func (exchange *PlxExchange) GetMarket(marketName string) (market Market, err error) {
	return nil, nil
}

func (exchange *PlxExchange) GetTicker(marketName string) (ticker []*TickerEntry, err error) {
	return ticker, nil
}

func (exchange *PlxExchange) GetBalance(currency string) (*Balance, error) {
	return nil, nil
}
