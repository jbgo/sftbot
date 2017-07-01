package trading

import (
	"math"
	"testing"
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

	balances := map[string]*Balance{
		"BTC": btcBalance,
	}

	market := &FakeMarket{Name: "BTC_ABC", ExistsValue: true}
	exchange := &FakeExchange{Balance: balances, Market: market}

	trader, err := NewTrader("BTC_ABC", exchange)

	if err != nil {
		t.Fatal(err)
	}

	trader.EstimatedFee = 0.0025

	order := &Order{Price: 0.025, Amount: 100}

	if trader.CanBuy(order) {
		t.Error("expect CanBuy() to be false when not enough money is available")
	}

	btcBalance.Available = 2.51

	if !trader.CanBuy(order) {
		t.Error("expect CanBuy() to be true when plenty of money is available")
	}

	balances["BTC"] = nil

	if trader.CanBuy(order) {
		t.Error("expect CanBuy() to be false when unable to retrieve balance")
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
