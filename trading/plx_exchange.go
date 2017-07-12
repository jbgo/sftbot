package trading

import (
	"fmt"
	"github.com/jbgo/sftbot/plx"
)

type PlxExchange struct {
	Client *plx.Client
}

func NewPlxExchange(client *plx.Client) Exchange {
	return &PlxExchange{Client: client}
}

func (exchange *PlxExchange) GetMarket(marketName string) (market Market, err error) {
	ticker, err := exchange.Client.GetTickerMap()
	_, ok := ticker[marketName]

	if !ok {
		return nil, fmt.Errorf("unknown market: %s", marketName)
	}

	market = NewPlxMarket(marketName)
	return market, nil
}

func (exchange *PlxExchange) GetBalance(currency string) (*Balance, error) {
	plxBalance, err := exchange.Client.GetBalance(currency)
	if err != nil {
		return nil, err
	}

	return &Balance{
		Available: plxBalance.Available,
		OnOrders:  plxBalance.OnOrders,
		BtcValue:  plxBalance.BtcValue,
	}, nil
}
