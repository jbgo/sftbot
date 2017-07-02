package trading

import (
	"fmt"
	// "github.com/jbgo/sftbot/data"
	// "github.com/jbgo/sftbot/plx"
	"log"
	// "math"
	"sort"
	"time"
)

type Trader struct {
	Market           Market
	Exchange         Exchange
	BTC_Balance      *Balance
	ALT_Balance      *Balance
	BuyThreshold     int64
	SellThreshold    float64
	VolatilityFactor float64
	ProfitFactor     float64
	BTC_BuyAmount    float64
	ALT_SellRatio    float64
	TimeWindow       int64
	EstimatedFee     float64
	LastBuy          *Trade
}

type MarketData struct {
	CurrentPrice    float64
	Percentiles     []float64
	VolatilityIndex float64
}

type Order struct {
	Type   string
	Price  float64
	Amount float64
}

type Trade struct {
	Date     int64
	Type     string
	Price    float64
	Amount   float64
	Total    float64
	Fee      float64
	Exchange string
	// For storing data specific to a particular exchange
	Metadata interface{}
}

func NewTrader(marketName string, exchange Exchange) (trader *Trader, err error) {
	market, err := exchange.GetMarket(marketName)
	if err != nil {
		return nil, err
	}

	if market == nil || !market.Exists() {
		return nil, fmt.Errorf("market not found: %s", market.GetName())
	}

	return &Trader{
		Market:           market,
		BuyThreshold:     50,
		SellThreshold:    1.06,
		VolatilityFactor: 1.02,
		ProfitFactor:     1.06,
		BTC_BuyAmount:    0.01,
		ALT_SellRatio:    0.5,
		TimeWindow:       24 * 60 * 60,
		EstimatedFee:     0.005,
		Exchange:         exchange,
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

	log.Printf("market=%s price=%0.9f 45_pct=%0.9f, 55_pct=%0.9f volatility=%0.9f\n",
		t.Market,
		marketData.CurrentPrice,
		marketData.Percentiles[45],
		marketData.Percentiles[55],
		marketData.VolatilityIndex)

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

func (t *Trader) Reconcile() (err error) {
	// TODO load state from DB

	t.BTC_Balance, err = t.Exchange.GetBalance("BTC")
	if err != nil {
		return err
	}

	t.ALT_Balance, err = t.Exchange.GetBalance(t.Market.GetCurrency())
	if err != nil {
		return err
	}

	endTime := time.Now().Unix()
	startTime := endTime - (24 * 60 * 60)

	if t.LastBuy != nil {
		startTime = t.LastBuy.Date
	}

	trades, err := t.Market.GetTradeHistory(startTime, endTime)
	if err != nil {
		return err
	}

	if len(trades) > 0 {
		t.LastBuy = trades[0]
	}

	// TODO recalculate thresholds?

	// TODO save updated state to DB

	return nil
}

func (t *Trader) Buy(marketData *MarketData) (order *Order, err error) {
	if !t.ShouldBuy(marketData) {
		return
	}

	order = t.BuildBuyOrder(marketData)

	if !t.CanBuy(order) {
		return nil, err
	}

	// TODO place the order

	if t.BuyThreshold > 10 {
		t.BuyThreshold -= 2
	}

	if t.SellThreshold > t.ProfitFactor {
		t.SellThreshold -= 0.01
	}

	return order, nil
}

func (t *Trader) ShouldBuy(marketData *MarketData) bool {
	return marketData.CurrentPrice < marketData.Percentiles[t.BuyThreshold] &&
		marketData.VolatilityIndex > t.VolatilityFactor
}

func (t *Trader) CanBuy(order *Order) bool {
	tradeValue := order.Price * order.Amount * (1 + t.EstimatedFee)
	return t.BTC_Balance.Available >= tradeValue
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

func (t *Trader) Sell(marketData *MarketData) (order *Order, err error) {
	if !t.ShouldSell(marketData) {
		return nil, nil
	}

	order, err = t.BuildSellOrder(marketData)

	if err != nil {
		return nil, nil
	}

	if !t.CanSell(order) {
		return nil, nil
	}

	// TODO place the order

	if t.BuyThreshold < 50 {
		t.BuyThreshold += 2
	}

	t.SellThreshold += 0.01

	return order, nil
}

func (t *Trader) ShouldSell(marketData *MarketData) bool {
	if t.LastBuy == nil {
		return false
	}

	return marketData.CurrentPrice > t.LastBuy.Price*t.SellThreshold
}

func (t *Trader) BuildSellOrder(marketData *MarketData) (*Order, error) {
	order := &Order{Type: "sell"}
	order.Amount = t.ALT_Balance.Available * t.ALT_SellRatio
	order.Price = marketData.CurrentPrice * (1 + t.EstimatedFee)

	if order.Amount*order.Price < t.BTC_BuyAmount {
		order.Amount = t.BTC_BuyAmount / order.Price
	}

	return order, nil
}

func (t *Trader) CanSell(order *Order) bool {
	return order.Amount <= t.ALT_Balance.Available
}

func (t *Trader) LoadMarketData() (marketData *MarketData, err error) {
	endTime := time.Now().Unix()
	startTime := endTime - (60 * 60 * 24 * 1)

	summaryData, err := t.Market.GetSummaryData(startTime, endTime)

	if err != nil {
		return nil, err
	}

	marketData = &MarketData{}
	marketData.Percentiles = calculatePercentiles(summaryData)
	marketData.VolatilityIndex = marketData.Percentiles[55] / marketData.Percentiles[45]

	currentPrice, err := t.Market.GetCurrentPrice()

	if err != nil {
		return nil, err
	}

	marketData.CurrentPrice = currentPrice

	return marketData, nil
}

func getSortedAverages(summaryData []*SummaryData) []float64 {
	sortedAverages := make([]float64, 0, len(summaryData))

	for _, data := range summaryData {
		sortedAverages = append(sortedAverages, data.WeightedAverage)
	}

	sort.Float64s(sortedAverages)

	return sortedAverages
}

func calculatePercentiles(summaryData []*SummaryData) []float64 {
	percentiles := make([]float64, 101)
	sortedAverages := getSortedAverages(summaryData)

	percentiles[0] = 0.0
	percentiles[100] = sortedAverages[len(sortedAverages)-1]

	for i := 1; i < 100; i += 1 {
		percentiles[i] = sortedAverages[i*len(sortedAverages)/100]
	}

	return percentiles
}
