package trading

import (
	"math"
	"testing"
	"time"
)

func TestNewTrader(t *testing.T) {
	market := &FakeMarket{Name: "BTC_ABC", ExistsValue: true}
	exchange := &FakeExchange{Market: market}

	trader, err := NewTrader("BTC_ABC", exchange)
	if err != nil {
		t.Fatal("ERROR", err)
	}

	expected := "ABC"
	actual := trader.Market.GetCurrency()
	if actual != "ABC" {
		t.Errorf("expect to trade currency %s, got %s", expected, actual)
	}
}

func TestNewTraderUnknownMarket(t *testing.T) {
	market := &FakeMarket{Name: "BTC_ABC", ExistsValue: false}
	exchange := &FakeExchange{Market: market}

	trader, err := NewTrader("BTC_QWERTY", exchange)

	if err == nil {
		t.Error("expect NewTrader() to fail for unknown market")
	}

	if trader != nil {
		t.Error("expect trader to be nil for unknown market")
	}
}

func TestShouldBuy(t *testing.T) {
	trader := Trader{}
	marketData := &MarketData{}

	trader.BuyThreshold = 42
	trader.VolatilityFactor = 1.02

	marketData.VolatilityIndex = 1.023
	marketData.Percentiles = make([]float64, 100)
	marketData.Percentiles[trader.BuyThreshold] = 0.004132
	marketData.CurrentPrice = 0.004029

	if !trader.ShouldBuy(marketData) {
		t.Error("expect ShouldBuy() to be true when market conditions are favorable")
	}

	marketData.CurrentPrice = 0.004177

	if trader.ShouldBuy(marketData) {
		t.Error("expect ShouldBuy() to be false when current price is above threshold")
	}

	marketData.CurrentPrice = 0.004029
	marketData.VolatilityIndex = 1.019

	if trader.ShouldBuy(marketData) {
		t.Error("expect ShouldBuy() to be false when market conditions are not favorable")
	}
}

func TestCanBuy(t *testing.T) {
	btcBalance := &Balance{Available: 2.5}

	trader := &Trader{BTC_Balance: btcBalance}

	trader.EstimatedFee = 0.0025

	order := &Order{Price: 0.025, Amount: 100}

	if trader.CanBuy(order) {
		t.Error("expect CanBuy() to be false when not enough money is available")
	}

	btcBalance.Available = 2.51

	if !trader.CanBuy(order) {
		t.Error("expect CanBuy() to be true when plenty of money is available")
	}
}

func TestBuildBuyOrder(t *testing.T) {
	trader := Trader{BTC_BuyAmount: 0.0125}
	marketData := &MarketData{CurrentPrice: 0.00392}

	order := trader.BuildBuyOrder(marketData)

	if order.Type != "buy" {
		t.Error("expect order.Type to be 'buy', got", order.Type)
	}

	expected := 0.0039004
	if order.Price != expected {
		t.Errorf("expect order.Price to equal %.9f, got %.9f", expected, order.Price)
	}

	precision := math.Pow(10, 9)
	expected = math.Floor(3.204799507742796*precision) / precision
	actual := math.Floor(order.Amount*precision) / precision
	if actual != expected {
		t.Errorf("expect order.Amount to equal %.9f, got %.9f", expected, actual)
	}
}

func TestLoadMarketData(t *testing.T) {
	data := make([]*SummaryData, 0)
	for i := 0; i <= 1000; i += 1 {
		value := 0.1 - (float64(i)+500.0)*0.00002 // TODO what should this be?
		data = append(data, &SummaryData{WeightedAverage: value})
	}

	market := &FakeMarket{
		Name:         "BTC_ABC",
		ExistsValue:  true,
		CurrentPrice: 0.54,
		SummaryData:  data,
	}

	exchange := &FakeExchange{Market: market}

	trader, err := NewTrader(market.Name, exchange)
	if err != nil {
		t.Fatal(err)
	}

	marketData, err := trader.LoadMarketData()
	if err != nil {
		t.Fatal(err)
	}

	expected := market.CurrentPrice
	if marketData.CurrentPrice != expected {
		t.Errorf("expect current price %f, got %f", expected, market.CurrentPrice)
	}

	expected = 1.02531
	actual := math.Floor(marketData.VolatilityIndex*math.Pow(10, 5)) / math.Pow(10, 5)
	if actual != expected {
		t.Errorf("expect volatility index %f, got %f", expected, actual)
	}

	for i := 1; i <= 100; i += 1 {
		if marketData.Percentiles[i] < marketData.Percentiles[i-1] {
			t.Errorf("expect pct_%d (%f) to be > pct_%d (%f)", i, marketData.Percentiles[i], i-1, marketData.Percentiles[i-1])
		}
	}

	t.Logf("pct_%d -> %f", 45, marketData.Percentiles[45])
	t.Logf("pct_%d -> %f", 55, marketData.Percentiles[55])
}

func TestShouldSell(t *testing.T) {
	trader := Trader{
		SellThreshold: 1.06,
		LastBuy:       &Trade{Price: 0.1},
	}

	marketData := &MarketData{CurrentPrice: 0.107}

	if !trader.ShouldSell(marketData) {
		t.Error("expect ShouldSell() to be true when there is money to be made")
	}

	marketData.CurrentPrice = 0.105

	if trader.ShouldSell(marketData) {
		t.Error("expect ShouldSell() to be false when it's better to hold")
	}

	trader.LastBuy = nil

	if trader.ShouldSell(marketData) {
		t.Error("expect ShouldSell() to be false when there's nothing to sell")
	}
}

func TestBuildSellOrder(t *testing.T) {
	market := &FakeMarket{Name: "BTC_ABC", ExistsValue: true}
	exchange := &FakeExchange{Market: market}

	marketData := &MarketData{
		CurrentPrice: 0.107,
	}

	trader, err := NewTrader("BTC_ABC", exchange)
	if err != nil {
		t.Fatal(err)
	}
	trader.ALT_SellRatio = 0.5
	trader.EstimatedFee = 0.005
	trader.ALT_Balance = &Balance{Available: 100.0}

	order, err := trader.BuildSellOrder(marketData)
	if err != nil {
		t.Fatal(err)
	}

	if order.Type != "sell" {
		t.Error("expect order.Type to be 'buy', got", order.Type)
	}

	expected := 50.0
	if order.Amount != expected {
		t.Errorf("expect order.Amount to equal %f, got %f", expected, order.Amount)
	}

	precision := math.Pow(10, 7)
	expected = 0.1075350
	actual := math.Floor(order.Price*precision) / precision
	if actual != expected {
		t.Errorf("expect order.Price to equal %.9f, got %.9f", expected, actual)
	}

	// Test minimum sell amount
	trader.BTC_BuyAmount = 0.1
	marketData.CurrentPrice = 0.025
	trader.ALT_Balance.Available = 6.0

	order, err = trader.BuildSellOrder(marketData)
	if err != nil {
		t.Fatal(err)
	}

	precision = 1000
	expected = 3.980
	actual = math.Floor(order.Amount*precision) / precision
	if actual != expected {
		t.Errorf("expect order.Amount to equal %.9f, got %.9f", expected, actual)
	}
}

func TestCanSell(t *testing.T) {
	altBalance := &Balance{}
	order := &Order{}
	trader := &Trader{ALT_Balance: altBalance}

	altBalance.Available = 10.0
	order.Amount = 10.0
	if !trader.CanSell(order) {
		t.Errorf("expect CanSell() to be true when balance is sufficient")
	}

	altBalance.Available = 9.99
	order.Amount = 10.0
	if trader.CanSell(order) {
		t.Errorf("expect CanSell() to be false when balance is deficient")
	}
}

func TestReconcile(t *testing.T) {
	tradeHistory := []*Trade{
		&Trade{Date: time.Now().Unix() - (3 * 60 * 60), Price: 0.42},
	}

	market := &FakeMarket{
		Name:         "BTC_XYZ",
		ExistsValue:  true,
		TradeHistory: tradeHistory,
	}

	ticker := map[string][]*TickerEntry{
		"BTC_XYZ": []*TickerEntry{&TickerEntry{Last: 0.105}},
	}

	balances := map[string]*Balance{
		"BTC": &Balance{Available: 1.23},
		"XYZ": &Balance{Available: 142.73},
	}

	exchange := &FakeExchange{
		Market:   market,
		Ticker:   ticker,
		Balances: balances,
	}

	trader, err := NewTrader(market.Name, exchange)

	if err != nil {
		t.Fatal(err)
	}

	err = trader.Reconcile()

	if err != nil {
		t.Fatal(err)
	}

	actual := trader.BTC_Balance.Available
	expected := balances["BTC"].Available
	if actual != expected {
		t.Errorf("expect BTC balance to equal %f, got %f", expected, actual)
	}

	actual = trader.ALT_Balance.Available
	expected = balances["XYZ"].Available
	if actual != expected {
		t.Errorf("expect ALT balance to equal %f, got %f", expected, actual)
	}

	actual = trader.LastBuy.Price
	expected = tradeHistory[0].Price
	if actual != expected {
		t.Errorf("expect last buy price to equal %f, got %f", expected, actual)
	}
}
