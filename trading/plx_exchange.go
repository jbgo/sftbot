package trading

import (
	"fmt"
	"github.com/jbgo/sftbot/plx"
)

type PlxExchange struct {
	Client *plx.Client
}

func NewPlxExchange(baseUrl string) Exchange {
	client := &plx.Client{BaseUrl: baseUrl}
	return &PlxExchange{Client: client}
}

func (exchange *PlxExchange) GetMarket(marketName string) (market Market, err error) {
	ticker, err := exchange.Client.GetTickerMap()
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
