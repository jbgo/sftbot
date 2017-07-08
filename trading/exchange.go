package trading

import ()

type Balance struct {
	Currency  string
	Available float64
	OnOrders  float64
	BtcValue  float64
}

type SummaryData struct {
	Date            int64
	High            float64
	Low             float64
	Open            float64
	Close           float64
	Volume          float64
	QuoteVolume     float64
	WeightedAverage float64
}

type TickerEntry struct {
	Market        string
	Last          float64
	LowestAsk     float64
	HighestBid    float64
	PercentChange float64
	BaseVolume    float64
	QuoteVolume   float64
}

type Order struct {
	Id     string
	Type   string
	Price  float64
	Amount float64
	Total  float64
	Filled bool
}

type Exchange interface {
	GetMarket(marketName string) (market Market, err error)
	GetTicker(marketName string) (ticker []*TickerEntry, err error)
	GetBalance(currency string) (*Balance, error)
}

type Market interface {
	GetName() string
	GetCurrency() string
	Exists() bool
	GetCurrentPrice() (float64, error)
	GetSummaryData(startTime, endTime int64) (summaryData []*SummaryData, err error)
	GetTradeHistory(startTime, endTime int64) ([]*Trade, error)
	GetPendingOrders() ([]*Order, error)
}
