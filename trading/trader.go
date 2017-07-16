package trading

import (
	"fmt"
	"github.com/jbgo/sftbot/db"
	"log"
	"sort"
	"strings"
	"time"
)

type Trader struct {
	Market           Market
	Exchange         Exchange
	Config           *TraderConfig
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
	Bids             []*Order
	Asks             []*Order
	DB               db.Store
	StateKey         string
}

type TraderConfig struct {
	BuyThresholdStart              int64
	BuyThresholdMin                int64
	BuyThresholdMax                int64
	BuyThresholdIncrement          int64
	SellThresholdStart             float64
	SellThresholdDecrement         float64
	VolatilityFactor               float64
	VolatilityIndexUpperPercentile int64
	VolatilityIndexLowerPercentile int64
	ProfitFactor                   float64
	BTC_BuyAmount                  float64
	ALT_SellRatio                  float64
	TimeWindow                     int64
	EstimatedFee                   float64
	StateKey                       string
	Simulate                       bool
}

func DefaultTraderConfig() *TraderConfig {
	return &TraderConfig{
		BuyThresholdStart:              50,
		BuyThresholdMin:                10,
		BuyThresholdMax:                50,
		BuyThresholdIncrement:          2,
		SellThresholdStart:             1.06,
		SellThresholdDecrement:         0.01,
		VolatilityFactor:               1.02,
		VolatilityIndexUpperPercentile: 55,
		VolatilityIndexLowerPercentile: 45,
		ProfitFactor:                   1.06,
		BTC_BuyAmount:                  0.01,
		ALT_SellRatio:                  0.5,
		TimeWindow:                     24 * 60 * 60,
		EstimatedFee:                   0.005,
		StateKey:                       "trader.state",
		Simulate:                       true,
	}
}

// Trader attributes that get persisted between runs
type TraderState struct {
	BuyThreshold  int64
	SellThreshold float64
	Bids          []*Order
	Asks          []*Order
}

type MarketData struct {
	CurrentPrice    float64
	Percentiles     []float64
	VolatilityIndex float64
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

func NewTrader(marketName string, exchange Exchange, dbStore db.Store, config *TraderConfig) (trader *Trader, err error) {
	market, err := exchange.GetMarket(marketName)
	if err != nil {
		return nil, err
	}

	if market == nil || !market.Exists() {
		return nil, fmt.Errorf("market not found: %s", market.GetName())
	}

	return &Trader{
		Market:           market,
		Config:           config,
		BuyThreshold:     config.BuyThresholdStart,
		SellThreshold:    config.SellThresholdStart,
		VolatilityFactor: config.VolatilityFactor,
		ProfitFactor:     config.ProfitFactor,
		BTC_BuyAmount:    config.BTC_BuyAmount,
		ALT_SellRatio:    config.ALT_SellRatio,
		TimeWindow:       config.TimeWindow,
		EstimatedFee:     config.EstimatedFee,
		Exchange:         exchange,
		DB:               dbStore,
		StateKey:         config.StateKey,
	}, nil
}

func (t *Trader) Trade() error {
	err := t.LoadState()
	if err != nil {
		return err
	}

	marketData, err := t.LoadMarketData()
	if err != nil {
		return err
	}

	err = t.Reconcile()
	if err != nil {
		return err
	}

	err = t.LoadBalances()
	if err != nil {
		return err
	}

	formatString := strings.Join([]string{
		"market=%s",
		"action=consider",
		"price=%0.9f",
		"buy_threshold=%d",
		"buy_pct=%0.9f",
		"volatility_index=%0.9f",
		"volatility_factor=%0.9f",
		"sell_threshold=%0.9f",
		"btc_balance=%0.9f",
		"alt_balance=%0.9f",
		"bid_count=%d",
		"filled_count=%d",
	}, " ")

	filledCount := 0
	for _, bid := range t.Bids {
		if bid.Filled {
			filledCount += 1
		}
	}

	log.Printf(formatString+"\n",
		t.Market.GetName(),
		marketData.CurrentPrice,
		t.BuyThreshold,
		marketData.Percentiles[t.BuyThreshold],
		marketData.VolatilityIndex,
		t.VolatilityFactor,
		t.SellThreshold,
		t.BTC_Balance.Available,
		t.ALT_Balance.Available,
		len(t.Bids),
		filledCount,
	)

	buyOrder, err := t.Buy(marketData)
	if err != nil {
		return err
	}

	if buyOrder != nil {
		log.Printf("market=%s action=order id=%s type=%s price=%0.9f amount=%0.9f total=%0.9f\n",
			t.Market.GetName(),
			buyOrder.Id,
			buyOrder.Type,
			buyOrder.Price,
			buyOrder.Amount,
			buyOrder.Total)
	}

	sellOrder, err := t.Sell(marketData)
	if err != nil {
		return err
	}

	if sellOrder != nil {
		log.Printf("market=%s action=order id=%s type=%s price=%0.9f amount=%0.9f total=%0.9f\n",
			t.Market.GetName(),
			sellOrder.Id,
			sellOrder.Type,
			sellOrder.Price,
			sellOrder.Amount,
			sellOrder.Total)
	}

	return t.SaveState()
}

func (t *Trader) LoadBalances() (err error) {
	t.BTC_Balance, err = t.Exchange.GetBalance("BTC")
	if err != nil {
		return err
	}

	t.ALT_Balance, err = t.Exchange.GetBalance(t.Market.GetCurrency())
	if err != nil {
		return err
	}

	return nil
}

func (t *Trader) Reconcile() error {
	pendingOrders, err := t.Market.GetPendingOrders()

	if err != nil {
		return err
	}

	markFilledBids(pendingOrders, t.Bids)

	t.Asks = removeFilledAsks(pendingOrders, t.Asks)

	return nil
}

func markFilledBids(pendingOrders, bids []*Order) {
	for _, bid := range bids {
		bid.Filled = true

		for _, order := range pendingOrders {
			if bid.Id == order.Id {
				bid.Filled = false
				break
			}
		}
	}
}

func removeFilledAsks(pendingOrders, staleAsks []*Order) (freshAsks []*Order) {
	freshAsks = make([]*Order, 0, len(pendingOrders))

	for _, ask := range staleAsks {
		for _, order := range pendingOrders {
			if ask.Id == order.Id {
				freshAsks = append(freshAsks, order)
				break
			}
		}
	}

	return freshAsks
}

func (t *Trader) Buy(marketData *MarketData) (order *Order, err error) {
	if !t.ShouldBuy(marketData) {
		log.Printf("should not buy")
		return nil, nil
	}

	order = t.BuildBuyOrder(marketData)

	if !t.CanBuy(order) {
		log.Printf("cannot not buy")
		return nil, nil
	}

	if t.Config.Simulate {
		order.Id = fmt.Sprintf("sim%d", time.Now().Unix())
	} else {
		err = t.Market.Buy(order)
		if err != nil {
			return order, err
		}
	}

	t.Bids = append(t.Bids, order)

	if t.BuyThreshold > t.Config.BuyThresholdMin {
		t.BuyThreshold -= t.Config.BuyThresholdIncrement
	}

	if t.SellThreshold > t.Config.ProfitFactor {
		t.SellThreshold -= t.Config.SellThresholdDecrement
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
	desiredPrice := marketData.CurrentPrice * (1 - t.EstimatedFee)
	altAmount := t.BTC_BuyAmount / desiredPrice

	return &Order{
		Type:   "buy",
		Price:  desiredPrice,
		Amount: altAmount,
		Total:  desiredPrice * altAmount,
	}
}

func (t *Trader) Sell(marketData *MarketData) (order *Order, err error) {
	if !t.ShouldSell(marketData) {
		log.Printf("should not sell")
		return nil, nil
	}

	order = t.BuildSellOrder(marketData)

	if !t.CanSell(order) {
		log.Printf("cannot not sell")
		return nil, nil
	}

	if t.Config.Simulate {
		order.Id = fmt.Sprintf("sim%d", time.Now().Unix())
	} else {
		err = t.Market.Sell(order)
		if err != nil {
			return order, err
		}
	}

	t.Bids = removeLastFilledBid(t.Bids)

	if t.BuyThreshold < t.Config.BuyThresholdMax {
		t.BuyThreshold += t.Config.BuyThresholdIncrement
	}

	t.SellThreshold += t.Config.SellThresholdDecrement

	return order, nil
}

func removeLastFilledBid(bids []*Order) []*Order {
	index := -1

	for i, bid := range bids {
		if bid.Filled {
			index = i
		}
	}

	return append(bids[:index], bids[index+1:]...)
}

func (t *Trader) ShouldSell(marketData *MarketData) bool {
	var lastTrade *Order

	for _, bid := range t.Bids {
		if bid.Filled {
			lastTrade = bid
		}
	}

	if lastTrade == nil {
		return false
	}

	return marketData.CurrentPrice > lastTrade.Price*t.SellThreshold
}

func (t *Trader) BuildSellOrder(marketData *MarketData) *Order {
	order := &Order{Type: "sell"}
	order.Amount = t.ALT_Balance.Available * t.ALT_SellRatio
	order.Price = marketData.CurrentPrice * (1 + t.EstimatedFee)

	if order.Amount*order.Price < t.BTC_BuyAmount {
		order.Amount = t.BTC_BuyAmount / order.Price
	}

	order.Total = order.Price * order.Amount

	return order
}

func (t *Trader) CanSell(order *Order) bool {
	return order.Amount <= t.ALT_Balance.Available
}

func (t *Trader) LoadMarketData() (marketData *MarketData, err error) {
	endTime := time.Now().Unix()
	startTime := endTime - t.TimeWindow

	summaryData, err := t.Market.GetSummaryData(startTime, endTime)

	if err != nil {
		return nil, err
	}

	marketData = &MarketData{}
	marketData.Percentiles = calculatePercentiles(summaryData)
	marketData.VolatilityIndex = marketData.Percentiles[t.Config.VolatilityIndexUpperPercentile] / marketData.Percentiles[t.Config.VolatilityIndexLowerPercentile]

	currentPrice, err := t.Market.GetCurrentPrice()

	if err != nil {
		return nil, err
	}

	marketData.CurrentPrice = currentPrice

	return marketData, nil
}

func (t *Trader) LoadState() error {
	traderState := &TraderState{}

	err, hasData := t.DB.HasData(t.StateKey)
	if err != nil {
		return err
	}

	if !hasData {
		// No state to load when first run of a new currency
		return nil
	}

	err = t.DB.Read(t.StateKey, &traderState)
	if err != nil {
		return err
	}

	t.BuyThreshold = traderState.BuyThreshold
	t.SellThreshold = traderState.SellThreshold
	t.Bids = traderState.Bids
	t.Asks = traderState.Asks

	return nil
}

func (t *Trader) SaveState() error {
	return t.DB.Write(t.StateKey, &TraderState{
		BuyThreshold:  t.BuyThreshold,
		SellThreshold: t.SellThreshold,
		Bids:          t.Bids,
		Asks:          t.Asks,
	})
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
