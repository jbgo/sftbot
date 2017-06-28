package trading

import (
	"fmt"
	"github.com/jbgo/sftbot/data"
	"github.com/jbgo/sftbot/plx"
	"log"
	"math"
	"sort"
	"time"
)

type Trader struct {
	Market           string
	Budget           float64
	BuyThreshold     int64
	SellThreshold    float64
	VolatilityFactor float64
	ProfitFactor     float64
	BTC_BuyAmount    float64
	ALT_SellRatio    float64
	TimeWindow       int64
	EstimatedPlxFee  float64
	Simulation       bool
	Bids             []*Order
	Asks             []*Order
	BTC_Balance      float64
}

type MarketData struct {
	CurrentPrice    float64
	Percentiles     []float64
	VolatilityIndex float64
	High            float64
	Low             float64
	SortedAverages  []float64
}

type Order struct {
	Type   string
	Price  float64
	Amount float64
}

func NewTrader(market string) (trader *Trader, err error) {
	ticker, err := plx.GetTicker()

	if err != nil {
		return nil, err
	}

	marketFound := false
	for _, entry := range ticker {
		if entry.Market == market {
			marketFound = true
			break
		}
	}

	if !marketFound {
		return nil, fmt.Errorf("market not found: %s", market)
	}

	return &Trader{
		Market:           market,
		Budget:           0.05,
		BuyThreshold:     50,
		SellThreshold:    1.06,
		VolatilityFactor: 1.02,
		ProfitFactor:     1.06,
		BTC_BuyAmount:    0.01,
		ALT_SellRatio:    0.5,
		TimeWindow:       24 * 60 * 60,
		EstimatedPlxFee:  0.005,
		Simulation:       true,
		// TODO get real balance each interation
		BTC_Balance: 0.02,
	}, nil
}

func (t *Trader) Trade() error {
	err := t.Reconcile()
	if err != nil {
		return err
	}

	marketData, err := t.LoadMarketData()
	if err != nil {
		return err
	}

	log.Printf("market=%s price=%0.9f 45_pct=%0.9f, 55_pct=%0.9f volatility=%0.9f high=%0.9f low=%0.9f\n",
		t.Market,
		marketData.CurrentPrice,
		marketData.Percentiles[45],
		marketData.Percentiles[55],
		marketData.VolatilityIndex,
		marketData.High,
		marketData.Low)

	order, err := t.Buy(marketData)
	if err != nil {
		return err
	}

	if order != nil {
		// TODO: save order to DB
		log.Printf("market=%s action=order type=buy\n", t.Market)
	}

	order, err = t.Sell(marketData)
	if err != nil {
		return err
	}

	if order != nil {
		// TODO: save order to DB
		log.Printf("market=%s action=order type=selln\n", t.Market)
	}

	return nil
}

func (t *Trader) Reconcile() error {
	return nil
}

func (t *Trader) Buy(marketData *MarketData) (order *Order, err error) {
	shouldBuy := marketData.CurrentPrice < marketData.Percentiles[t.BuyThreshold] &&
		marketData.VolatilityIndex > t.VolatilityFactor

	if !shouldBuy {
		return
	}

	order = t.BuildBuyOrder(marketData)
	availableToTrade := t.BTC_AvailableToTrade(marketData)

	tradeValue := order.Price * order.Amount * (1 + t.EstimatedPlxFee)
	canBuy := availableToTrade > tradeValue

	if !canBuy {
		return
	}

	if t.Simulation {
		t.Bids = append(t.Bids, order)
	} else {
		// TODO place the order
	}

	if t.BuyThreshold > 10 {
		t.BuyThreshold -= 2
	}

	if t.SellThreshold > t.ProfitFactor {
		t.SellThreshold -= 0.01
	}

	return order, nil
}

func (t *Trader) BuildBuyOrder(marketData *MarketData) *Order {
	desiredPrice := marketData.CurrentPrice * 0.995
	altAmount := t.BTC_BuyAmount / desiredPrice

	return &Order{
		Type:   "buy",
		Price:  desiredPrice,
		Amount: altAmount,
	}
}

func (t *Trader) BTC_AvailableToTrade(marketData *MarketData) float64 {
	onOrders := 0.0

	for _, bid := range t.Bids {
		onOrders += bid.Price * bid.Amount * t.EstimatedPlxFee
	}

	return math.Min(t.Budget, t.BTC_Balance) - onOrders
}

func (t *Trader) Sell(marketData *MarketData) (order *Order, err error) {
	if len(t.Bids) == 0 {
		return nil, nil
	}

	lastBuyPrice := t.Bids[len(t.Bids)-1].Price
	shouldSell := marketData.CurrentPrice > lastBuyPrice*t.SellThreshold

	if !shouldSell {
		return nil, nil
	}

	availableAltBalance := 0.0 // TODO: get available alt balance
	altAmount := availableAltBalance * t.ALT_SellRatio
	desiredPrice := marketData.CurrentPrice * 1.001
	btcAmount := altAmount * desiredPrice

	if altAmount*btcAmount < 0.1*t.BTC_BuyAmount {
		// Enforce minimum sell amount
		altAmount = 0.5 * t.BTC_BuyAmount / marketData.CurrentPrice
		btcAmount = altAmount * desiredPrice
	}

	if altAmount > availableAltBalance {
		// Don't sell. Wait for more buys.
		return nil, nil
	}

	if t.Simulation {
		order = &Order{
			Type:   "sell",
			Price:  desiredPrice,
			Amount: altAmount,
		}

		t.Asks = append(t.Asks, order)
	} else {
		// TODO place the order
	}

	// remove last bid from list - yuck! y no slice.Pop()?
	if len(t.Bids) > 0 {
		i := len(t.Bids) - 1
		t.Bids = append(t.Bids[:i], t.Bids[i+1:]...)
	}

	if t.BuyThreshold < 50 {
		t.BuyThreshold += 2
	}

	t.SellThreshold += 0.01

	return order, nil
}

func (t *Trader) LoadMarketData() (marketData *MarketData, err error) {
	marketData = &MarketData{}

	endTime := time.Now().Unix()
	startTime := endTime - (60 * 60 * 24 * 1)

	params := plx.ChartDataParams{
		CurrencyPair: t.Market,
		Start:        startTime,
		End:          endTime,
		Period:       300,
	}

	chartData, err := plx.GetChartData(&params)

	if err != nil {
		return nil, err
	}

	marketData.SortedAverages = getSortedAverages(chartData)
	marketData.Percentiles = calculatePercentiles(marketData.SortedAverages)
	marketData.VolatilityIndex = marketData.Percentiles[55] / marketData.Percentiles[45]
	marketData.Low = marketData.SortedAverages[0]
	marketData.High = marketData.SortedAverages[len(marketData.SortedAverages)-1]

	ticker, err := plx.GetTicker()

	if err != nil {
		return nil, err
	}

	for _, entry := range ticker {
		if entry.Market == t.Market {
			marketData.CurrentPrice = entry.Last
		}
	}

	return marketData, nil
}

func getSortedAverages(chartData []data.ChartData) []float64 {
	sortedAverages := make([]float64, 0, len(chartData))

	for _, data := range chartData {
		sortedAverages = append(sortedAverages, data.WeightedAverage)
	}

	sort.Float64s(sortedAverages)

	return sortedAverages
}

func calculatePercentiles(sortedAverages []float64) []float64 {
	percentiles := make([]float64, 101)

	percentiles[0] = 0.0
	percentiles[100] = sortedAverages[len(sortedAverages)-1]

	for i := 1; i < 100; i += 1 {
		percentiles[i] = sortedAverages[i*len(sortedAverages)/100]
	}

	return percentiles
}
