package trading

import (
	"fmt"
	"github.com/jbgo/sftbot/plx"
	"strconv"
	"strings"
)

type PlxMarket struct {
	Name   string
	Client *plx.Client
}

func NewPlxMarket(marketName string, client *plx.Client) (*PlxMarket, error) {
	market := &PlxMarket{
		Name:   marketName,
		Client: client,
	}

	if market.Exists() {
		return market, nil
	}

	return nil, fmt.Errorf("unknown market: %s", marketName)
}

func (market *PlxMarket) Buy(order *Order) error {
	return nil
}

func (market *PlxMarket) Exists() bool {
	ticker, err := market.Client.GetTickerMap()

	if err != nil {
		return false
	}

	_, ok := ticker[market.Name]

	return ok
}

func (market *PlxMarket) GetCurrency() string {
	return strings.Split(market.Name, "_")[1]
}

func (market *PlxMarket) GetCurrentPrice() (float64, error) {
	ticker, err := market.Client.GetTickerMap()

	if err != nil {
		return 0.0, err
	}

	return ticker[market.Name].Last, nil
}

func (market *PlxMarket) GetName() string {
	return market.Name
}

func (market *PlxMarket) GetPendingOrders() ([]*Order, error) {
	plxOrders, err := market.Client.GetOpenOrders(market.Name)
	if err != nil {
		return nil, err
	}

	orders := make([]*Order, 0, len(plxOrders))
	for _, o := range plxOrders {
		orders = append(orders, &Order{
			Id:     strconv.FormatInt(o.Number, 10),
			Type:   o.Type,
			Price:  o.Rate,
			Amount: o.Amount,
			Total:  o.Total,
			Filled: false,
		})
	}

	return orders, nil
}

func (market *PlxMarket) GetSummaryData(startTime, endTime int64) (summaryData []*SummaryData, err error) {
	params := &plx.ChartDataParams{
		CurrencyPair: market.Name,
		Start:        startTime,
		End:          endTime,
		Period:       300,
	}

	chartData, err := market.Client.GetChartData(params)

	if err != nil {
		return nil, err
	}

	summaryData = make([]*SummaryData, 0, len(chartData))
	for _, d := range chartData {
		s := SummaryData(d)
		summaryData = append(summaryData, &s)
	}

	return summaryData, nil
}

func (market *PlxMarket) Sell(order *Order) error {
	return nil
}
