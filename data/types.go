package data

type Account struct {
	Name    string
	Balance float64
}

type ChartData struct {
	Date            int64
	High            float64
	Low             float64
	Open            float64
	Close           float64
	Volume          float64
	QuoteVolume     float64
	WeightedAverage float64
}
