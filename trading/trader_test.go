package trading

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewTrader(t *testing.T) {
	market := &FakeMarket{Name: "BTC_ABC", ExistsValue: true}
	exchange := &FakeExchange{Market: market}

	trader, err := NewTrader("BTC_ABC", exchange)
	require.Nil(t, err)

	assert.Equal(t, "ABC", trader.Market.GetCurrency())
}

func TestNewTraderUnknownMarket(t *testing.T) {
	market := &FakeMarket{Name: "BTC_ABC", ExistsValue: false}
	exchange := &FakeExchange{Market: market}

	trader, err := NewTrader("BTC_QWERTY", exchange)

	require.NotNil(t, err)
	require.Nil(t, trader)
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

	assert.True(t, trader.ShouldBuy(marketData))

	marketData.CurrentPrice = 0.004177

	assert.False(t, trader.ShouldBuy(marketData))

	marketData.CurrentPrice = 0.004029
	marketData.VolatilityIndex = 1.019

	assert.False(t, trader.ShouldBuy(marketData))
}

func TestCanBuy(t *testing.T) {
	btcBalance := &Balance{Available: 2.5}
	trader := &Trader{BTC_Balance: btcBalance}
	trader.EstimatedFee = 0.0025
	order := &Order{Price: 0.025, Amount: 100}

	assert.False(t, trader.CanBuy(order))

	btcBalance.Available = 2.51

	assert.True(t, trader.CanBuy(order))
}

func TestBuildBuyOrder(t *testing.T) {
	trader := Trader{BTC_BuyAmount: 0.0125}
	marketData := &MarketData{CurrentPrice: 0.00392}

	order := trader.BuildBuyOrder(marketData)

	assert.Equal(t, "buy", order.Type)
	assert.InDelta(t, 0.0039004, order.Price, 0.00000001)
	assert.InDelta(t, 3.2047995, order.Amount, 0.00000001)
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
	require.Nil(t, err)

	marketData, err := trader.LoadMarketData()
	require.Nil(t, err)

	assert.Equal(t, market.CurrentPrice, marketData.CurrentPrice)

	assert.InDelta(t, 1.025316, marketData.VolatilityIndex, 0.000001)

	for i := 1; i <= 100; i += 1 {
		assert.Condition(t, func() bool {
			return marketData.Percentiles[i] >= marketData.Percentiles[i-1]
		}, "expect pct_%d (%f) to be >= pct_%d (%f)", i, marketData.Percentiles[i], i-1, marketData.Percentiles[i-1])
	}

	t.Logf("pct_%d -> %f", 45, marketData.Percentiles[45])
	t.Logf("pct_%d -> %f", 55, marketData.Percentiles[55])
}

func TestShouldSell(t *testing.T) {
	trader := Trader{
		SellThreshold: 1.06,
		Bids: []*Order{
			&Order{Price: 0.1, Filled: true},
			&Order{Price: 0.9, Filled: false},
		},
	}

	marketData := &MarketData{CurrentPrice: 0.107}

	assert.True(t, trader.ShouldSell(marketData))

	marketData.CurrentPrice = 0.105

	assert.False(t, trader.ShouldSell(marketData))

	trader.Bids = []*Order{}

	assert.False(t, trader.ShouldSell(marketData))
}

func TestBuildSellOrder(t *testing.T) {
	market := &FakeMarket{Name: "BTC_ABC", ExistsValue: true}
	exchange := &FakeExchange{Market: market}

	marketData := &MarketData{
		CurrentPrice: 0.107,
	}

	trader, err := NewTrader("BTC_ABC", exchange)
	require.Nil(t, err)

	trader.ALT_SellRatio = 0.5
	trader.EstimatedFee = 0.005
	trader.ALT_Balance = &Balance{Available: 100.0}

	order, err := trader.BuildSellOrder(marketData)
	require.Nil(t, err)

	assert.Equal(t, "sell", order.Type)
	assert.Equal(t, 50.0, order.Amount)
	assert.InDelta(t, 0.1075350, order.Price, 0.00000001)

	// Test minimum sell amount
	trader.BTC_BuyAmount = 0.1
	marketData.CurrentPrice = 0.025
	trader.ALT_Balance.Available = 6.0

	order, err = trader.BuildSellOrder(marketData)
	require.Nil(t, err)

	assert.InDelta(t, 3.980, order.Amount, 0.001)
}

func TestCanSell(t *testing.T) {
	altBalance := &Balance{}
	order := &Order{}
	trader := &Trader{ALT_Balance: altBalance}

	altBalance.Available = 10.0
	order.Amount = 10.0

	assert.True(t, trader.CanSell(order))

	altBalance.Available = 9.99
	order.Amount = 10.0

	assert.False(t, trader.CanSell(order))
}

func TestLoadBalances(t *testing.T) {
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
	require.Nil(t, err)

	err = trader.LoadBalances()
	require.Nil(t, err)

	assert.Equal(t, balances["BTC"].Available, trader.BTC_Balance.Available)
	assert.Equal(t, balances["XYZ"].Available, trader.ALT_Balance.Available)
}

func TestReconcile(t *testing.T) {
	market := &FakeMarket{Name: "BTC_XYZ", ExistsValue: true}
	exchange := &FakeExchange{Market: market}

	trader, err := NewTrader(market.Name, exchange)
	require.Nil(t, err)

	trader.Bids = []*Order{
		&Order{Id: "foo", Price: 0.24},
		&Order{Id: "bar", Price: 0.19},
		&Order{Id: "baz", Price: 0.27},
	}

	trader.Asks = []*Order{
		&Order{Id: "boggle", Price: 0.29},
	}

	market.PendingOrders = []*Order{
		&Order{Id: "foo", Price: 0.24},
		&Order{Id: "baz", Price: 0.27},
	}

	err = trader.Reconcile()
	require.Nil(t, err)

	assert.Equal(t, 2, len(trader.Bids))
	assert.Equal(t, "foo", trader.Bids[0].Id)
	assert.Equal(t, "baz", trader.Bids[1].Id)

	assert.Equal(t, 0, len(trader.Asks))
}

func TestPersistence(t *testing.T) {
	market := &FakeMarket{Name: "BTC_TESTING", ExistsValue: true}
	exchange := &FakeExchange{Market: market}

	trader, err := NewTrader(market.Name, exchange)
	require.Nil(t, err)

	trader.DB.Delete(trader.StateKey)

	err = trader.LoadState()
	require.Nil(t, err)

	trader.BuyThreshold = 36
	trader.SellThreshold = 1.123
	trader.Bids = []*Order{
		&Order{Price: 0.1},
		&Order{Price: 0.2},
	}
	trader.Asks = []*Order{
		&Order{Price: 0.3},
	}

	err = trader.SaveState()
	require.Nil(t, err)

	trader, err = NewTrader(market.Name, exchange)
	require.Nil(t, err)

	err = trader.LoadState()
	require.Nil(t, err)

	assert.Equal(t, 36, int(trader.BuyThreshold))
	assert.Equal(t, 1.123, trader.SellThreshold)
	assert.Equal(t, 2, len(trader.Bids))
	assert.Equal(t, 1, len(trader.Asks))
	assert.Equal(t, 0.2, trader.Bids[1].Price)
}

// func TestBuy(t *testing.T) {
// 	t.Fatal("TODO")
// }

// func TestSell(t *testing.T) {
// 	t.Fatal("TODO")
// }
