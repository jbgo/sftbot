package trading

import ()

type PlxMarket struct {
}

func (market *PlxMarket) Buy(order *Order) error {
	return nil
}

func (market *PlxMarket) Exists() bool {
	return false
}

func (market *PlxMarket) GetCurrency() string {
	return ""
}

func (market *PlxMarket) GetCurrentPrice() (float64, error) {
	return 0.0, nil
}

func (market *PlxMarket) GetName() string {
	return ""
}

func (market *PlxMarket) GetPendingOrders() ([]*Order, error) {
	return nil, nil
}

func (market *PlxMarket) GetSummaryData(startTime, endTime int64) (summaryData []*SummaryData, err error) {
	return nil, nil
}

func (market *PlxMarket) Sell(order *Order) error {
	return nil
}
