/**
Inputs:
  - plx.GetChartData() for last 24 hours to calculate percentiles
  - plx.GetOpenOrders() what orders do we have on the books waiting to fill?
  - plx.GetCompleteBalances() for currencies in market we plan to trade
  - plx.GetTradeHistory() for last 60 seconds to determine current price

Outputs:
 - trade decision (buy, sell, hold, etc.)
 - trade actions (buy, sell, cancel, etc.)
 - log full API calls in separate log from trading log
 - save current State to DB
 - timeseries data points - for analysis later (these can go in DB)
*/

package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jbgo/sftbot/db"
	"github.com/jbgo/sftbot/plx"
	"github.com/jbgo/sftbot/trading"
	"io/ioutil"
	"log"
	"time"
)

const LIVE_DB = "sftbot-live.db"
const SIMULATE_DB = "sftbot.db"

type TradeCommand struct {
	Flags  *flag.FlagSet
	Market string

	Simulate bool

	CurrentTime    time.Time
	CurrentTimeVar string

	Config string
}

func (c *TradeCommand) Synopsis() string {
	return "trade currency in the given market"
}

func (c *TradeCommand) Help() string {
	return formatHelpText(`
Usage: sftbot plx ticker [options]

  ` + c.Synopsis() + `

` + helpOptions(c))
}

func (c *TradeCommand) InitFlags() *flag.FlagSet {
	c.Flags = flag.NewFlagSet("plx ticker", flag.ContinueOnError)

	c.Flags.StringVar(&c.Market, "market", "", "Comma-separated list of PLX markets specified as currency pair (e.g. BTC_XYZ)")
	c.Flags.BoolVar(&c.Simulate, "simulate", false, "Enable simulate mode")
	c.Flags.StringVar(&c.CurrentTimeVar, "current-time", "", "When simulate mode is enabled, use timestamp as the current time. YYYY-MM-DD HH:MM:SS")
	c.Flags.StringVar(&c.Config, "config", "", "Trader config file (JSON)")

	return c.Flags
}

func (c *TradeCommand) Validate() {
	var err error

	if len(c.Market) == 0 {
		log.Fatal(c.Help())
	}

	if !c.Simulate {
		return
	}

	if len(c.CurrentTimeVar) > 0 {
		c.CurrentTime, err = time.Parse(TIME_VAR_FORMAT, c.CurrentTimeVar)
		if err != nil {
			log.Fatal(err)
		}
	} else if c.Simulate {
		log.Fatal(fmt.Errorf("-current-time required when -simulate is true"))
	}
}

func (c *TradeCommand) Run(args []string) int {
	c.InitFlags()
	c.Flags.Parse(args)
	c.Validate()

	if len(c.Market) == 0 {
		fmt.Println(c.Help())
		return 1
	}

	dbPath := LIVE_DB
	if c.Simulate {
		dbPath = SIMULATE_DB
	}

	dbStore, err := db.NewBoltStore(c.Market, dbPath)
	if err != nil {
		log.Fatal(err)
	}

	plxClient := plx.NewLiveClient()
	plxExchange := trading.NewPlxExchange(plxClient)

	traderConfig, err := c.LoadTraderConfig()
	if err != nil {
		log.Fatal(err)
	}

	trader, err := trading.NewTrader(c.Market, plxExchange, dbStore, traderConfig)
	if err != nil {
		log.Fatal(err)
	}

	if !c.Simulate {
		c.CurrentTime = time.Now()
	}

	lastRun := time.Now().Unix()

	for {
		currentRun := time.Now().Unix()

		if currentRun%300 != 150 || currentRun-lastRun < 150 {
			// run every 5 minutes at the 2:30 mark
			time.Sleep(200 * time.Millisecond)
			continue
		}

		lastRun = currentRun

		err = trader.Trade()
		if err != nil {
			log.Printf("ERROR: %s" + err.Error())
		}
	}

	return 0
}

func (c *TradeCommand) LoadTraderConfig() (*trading.TraderConfig, error) {
	if len(c.Config) == 0 {
		return nil, fmt.Errorf("-config is required")
	}

	data, err := ioutil.ReadFile(c.Config)
	if err != nil {
		return nil, err
	}

	traderConfig := &trading.TraderConfig{}
	err = json.Unmarshal(data, traderConfig)

	return traderConfig, err
}
