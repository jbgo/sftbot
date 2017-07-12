package trading

import (
	"github.com/jbgo/sftbot/plx"
)

type PlxExchange struct {
	Client *plx.Client
}

func NewPlxExchange(client *plx.Client) Exchange {
	return &PlxExchange{Client: client}
}

func (exchange *PlxExchange) GetMarket(marketName string) (market Market, err error) {
	return NewPlxMarket(marketName, exchange.Client)
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
