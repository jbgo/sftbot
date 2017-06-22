package command

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jbgo/sftbot/data"
	"log"
	"math"
	"sort"
	"time"
)

const WINDOW = 24 * 60 * 60
const BUY_COOLOFF = 0
const SELL_COOLOFF = 0

const BTC_BUY_AMOUNT = 0.01
const XRP_SELL_AMOUNT = 90
const XRP_SELL_PCT = 0.5
const PLX_FEE = 0.00025
const VOLATILITY = 1.02
const PROFIT_FACTOR = 1.06

type SimulateCommand struct {
	DB    *data.Store
	Flags *flag.FlagSet

	BTC_Balance   float64
	XRP_Balance   float64
	PLX_Balance   float64
	Bids          []float64
	LastBuyTime   int64
	LastSellTime  int64
	BuyThreshold  int
	SellThreshold float64

	CurrencyPair string

	StartTimeVar string
	StartTime    time.Time

	EndTimeVar string
	EndTime    time.Time
}

const TIME_VAR_FORMAT = "2006-01-02 15:04:05 MST"

func (c *SimulateCommand) Synopsis() string {
	return "simulate trading the given currency pair"
}
func (c *SimulateCommand) Help() string {
	return formatHelpText(`
Usage: sftbot simulate [options]

  Simulate trading the given currency pair.

` + helpOptions(c))
}

func (c *SimulateCommand) InitFlags() *flag.FlagSet {
	c.Flags = flag.NewFlagSet("simulate", flag.PanicOnError)

	c.Flags.StringVar(&c.CurrencyPair, "currency-pair", "", "PLX currency pair. Must be in the format BTC_XYZ")
	c.Flags.StringVar(&c.StartTimeVar, "start-time", "2017-05-01 00:00:00 CST", "Simulation start time. YYYY-MM-DD HH:MM:SS (TZ)")
	c.Flags.StringVar(&c.EndTimeVar, "end-time", "2017-06-21 00:00:00 CST", "Simulation end time. YYYY-MM-DD HH:MM:SS (TZ)")

	return c.Flags
}

func (c *SimulateCommand) Validate() error {
	if len(c.CurrencyPair) == 0 {
		return fmt.Errorf("missing -currency-pair")
	}

	var err error

	c.StartTime, err = time.Parse(TIME_VAR_FORMAT, c.StartTimeVar)
	if err != nil {
		return err
	}

	c.EndTime, err = time.Parse(TIME_VAR_FORMAT, c.EndTimeVar)
	if err != nil {
		return err
	}

	return nil
}

func (c *SimulateCommand) ForTimeRange(startTime, endTime time.Time, callback func(d *data.ChartData)) error {
	db, err := data.OpenDB()

	if err != nil {
		log.Println(err)
		return err
	}

	defer db.Close()

	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("chart_data." + c.CurrencyPair))

		if b == nil {
			return fmt.Errorf("unsupported currency pair: %s", c.CurrencyPair)
		}

		minTime := db.EncodeKey(startTime.Unix())
		maxTime := db.EncodeKey(endTime.Unix())

		cursor := b.Cursor()

		for k, v := cursor.Seek(minTime); k != nil && bytes.Compare(k, maxTime) <= 0; k, v = cursor.Next() {
			d := data.ChartData{}
			db.DecodeValue(v, &d)
			callback(&d)
		}

		return nil
	})
}

func (c *SimulateCommand) GetPrices(startTime, endTime time.Time) []float64 {
	prices := []float64{}

	c.ForTimeRange(startTime, endTime, func(d *data.ChartData) {
		prices = append(prices, d.WeightedAverage)
	})

	return prices
}

func CalcPercentiles(prices []float64) map[int]float64 {
	percentiles := make(map[int]float64)
	count := len(prices)

	sortedPrices := make([]float64, len(prices))
	copy(sortedPrices, prices)
	sort.Float64s(sortedPrices)

	for i := 1; i <= 99; i += 1 {
		percentiles[i] = sortedPrices[i*count/100]
	}

	return percentiles
}

func (c *SimulateCommand) Run(args []string) int {
	c.InitFlags()
	c.Flags.Parse(args)

	err := c.Validate()
	if err != nil {
		log.Println(err)
		return 1
	}

	c.BTC_Balance = 0.1
	c.XRP_Balance = 0.00
	c.PLX_Balance = 0.00

	c.BuyThreshold = 50
	c.SellThreshold = PROFIT_FACTOR

	fmt.Printf("Starting Balances: BTC=%f XRP=%f PLX=%f\n", c.BTC_Balance, c.XRP_Balance, c.PLX_Balance)

	chartData := make([]*data.ChartData, 0, 8640)
	c.ForTimeRange(c.StartTime, c.EndTime, func(d *data.ChartData) {
		chartData = append(chartData, d)
	})

	fmt.Printf("found %d chart data items\n", len(chartData))

	var d *data.ChartData
	var currentPrice float64

	for _, d = range chartData {
		currentPrice = d.WeightedAverage

		// age the sell threshold in case it gets pumped up too high
		if len(c.Bids) > 0 && c.SellThreshold > PROFIT_FACTOR {
			c.SellThreshold -= 0.001
		}

		pStart := time.Unix(d.Date-WINDOW, 0)
		pEnd := time.Unix(d.Date, 0)
		prices := c.GetPrices(pStart, pEnd)
		percentiles := CalcPercentiles(prices)

		if c.ShouldBuy(currentPrice, prices, percentiles) {
			if c.Buy(currentPrice, d) {
				fmt.Printf("BUY price=%f date=%v BTC=%f XRP=%f NAV=%f bt=%d st=%f\n", currentPrice, time.Unix(d.Date, 0), c.BTC_Balance, c.XRP_Balance, c.BTC_Balance+(currentPrice*c.XRP_Balance), c.BuyThreshold, c.SellThreshold)
			}
		}

		if c.ShouldSell(currentPrice, prices, percentiles) {
			if c.Sell(currentPrice, d, false) {
				fmt.Printf("SELL price=%f date=%v BTC=%f XRP=%f NAV=%f bt=%d st=%f\n", currentPrice, time.Unix(d.Date, 0), c.BTC_Balance, c.XRP_Balance, c.BTC_Balance+(currentPrice*c.XRP_Balance), c.BuyThreshold, c.SellThreshold)
			}
		}
	}

	// Sell remaining XRP to exit position
	if c.XRP_Balance > 0 {
		c.Sell(currentPrice, d, true)
		fmt.Printf("SELL price=%f date=%v BTC=%f XRP=%f NAV=%f bt=%d st=%f\n", currentPrice, time.Unix(d.Date, 0), c.BTC_Balance, c.XRP_Balance, c.BTC_Balance+(currentPrice*c.XRP_Balance), c.BuyThreshold, c.SellThreshold)
	}

	fmt.Printf("Ending Balances: BTC=%f XRP=%f PLX=%f\n", c.BTC_Balance, c.XRP_Balance, c.PLX_Balance)

	return 0
}

func (c *SimulateCommand) ShouldBuy(price float64, prices []float64, percentiles map[int]float64) bool {
	return price < percentiles[c.BuyThreshold] && percentiles[55]/percentiles[45] > VOLATILITY
}

func (c *SimulateCommand) ShouldSell(price float64, prices []float64, percentiles map[int]float64) bool {
	return len(c.Bids) > 0 && price > c.Bids[len(c.Bids)-1]*c.SellThreshold
}

func (c *SimulateCommand) Buy(price float64, d *data.ChartData) bool {
	xrpAmount := BTC_BUY_AMOUNT / (price + (price * PLX_FEE))
	btcAmount := xrpAmount * price
	commission := btcAmount * PLX_FEE

	if c.BTC_Balance-btcAmount-commission > 0 && d.Date-c.LastBuyTime > BUY_COOLOFF {
		c.BTC_Balance -= btcAmount
		c.XRP_Balance += xrpAmount

		c.BTC_Balance -= commission
		c.PLX_Balance += commission

		c.Bids = append(c.Bids, price)
		c.LastBuyTime = d.Date
		if c.BuyThreshold > 10 {
			c.BuyThreshold -= 2
		}

		if c.SellThreshold > PROFIT_FACTOR {
			c.SellThreshold -= 0.01
		}

		return true
	}

	return false
}

func (c *SimulateCommand) Sell(price float64, d *data.ChartData, force bool) bool {
	xrpAmount := c.XRP_Balance * XRP_SELL_PCT

	xrpAmount = math.Max(xrpAmount, XRP_SELL_AMOUNT)

	// Enforce minimum XRP sell
	if xrpAmount < XRP_SELL_AMOUNT {
		return false
	}

	if force {
		xrpAmount = c.XRP_Balance
	}

	btcAmount := xrpAmount * price
	commission := btcAmount * PLX_FEE

	if force || (c.XRP_Balance > 0 && c.XRP_Balance-xrpAmount >= 0 && d.Date-c.LastSellTime > SELL_COOLOFF) {
		c.BTC_Balance += btcAmount
		c.XRP_Balance -= xrpAmount

		c.BTC_Balance -= commission
		c.PLX_Balance += commission

		if len(c.Bids) > 0 {
			i := len(c.Bids) - 1
			c.Bids = append(c.Bids[:i], c.Bids[i+1:]...)
		}
		c.LastSellTime = d.Date

		if c.BuyThreshold < 50 {
			c.BuyThreshold += 2
		}

		c.SellThreshold += 0.01

		return true
	}

	return false
}
