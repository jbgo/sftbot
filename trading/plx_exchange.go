package trading

import (
	"fmt"
	"github.com/jbgo/sftbot/plx"
)

type PlxExchange struct {
}

func NewPlxExchange() Exchange {
	return &PlxExchange{}
}

func (exchange *PlxExchange) GetMarket(marketName string) (market Market, err error) {
	ticker, err := plx.GetTickerMap()
	_, ok := ticker[marketName]

	if !ok {
		return nil, fmt.Errorf("unknown market: %s", marketName)
	}

	// TODO return NewPlxMarket(marketName)
	return nil, nil
}

func (exchange *PlxExchange) GetBalance(currency string) (*Balance, error) {
	return nil, nil
}

// type PlxMarket struct {
// 	name string
// }

// func NewPlxMarket(marketName string) (Market, error) {
// 	market := &PlxMarket{name: marketName}
// 	return market, nil
// }
