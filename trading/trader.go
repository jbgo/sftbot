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
	BuyThreshold     int64
	SellThreshold    float64
	VolatilityFactor float64
	ProfitFactor     float64
	BTC_BuyAmount    float64
	ALT_SellRatio    float64
	TimeWindow       int64
	EstimatedFee     float64
	Simulation       bool
	Bids             []*Order
	Asks             []*Order
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

func NewTrader(marketName string, exchange Exchange) (trader *Trader, err error) {
	market, err := exchange.GetMarket(marketName)
	if err != nil {
		return nil, err
	}

	if !market.Exists() {
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
		Simulation:       true,
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

func (t *Trader) Reconcile() error {
	// TODO get balances
	// TODO get my trade history
	// TODO load open orders (bids and asks) from DB
	// TODO update open orders (bids and asks) to be in sync with PLX
	// TODO save bids and asks to DB
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
	t.Bids = append(t.Bids, order)
	// TODO persist bids to DB

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
	btcBalance, err := t.Exchange.GetBalance("BTC")
	if err != nil || btcBalance == nil {
		// TODO don't attempt to buy if network or exchange error, but maybe we
		// should log or alert on this so we know when we're trying to buy but can't.
		return false
	}

	tradeValue := order.Price * order.Amount * (1 + t.EstimatedFee)
	canBuy := btcBalance.Available >= tradeValue

	return canBuy
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