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
const TRADE_INTERVAL = 300 // 5 minutes

type TradeCommand struct {
	Flags *flag.FlagSet

	Market string
	Config string
	Offset int64

	DBStore db.Store
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
	c.Flags.StringVar(&c.Config, "config", "", "Trader config file (JSON)")

	// This allows us to run multiple traders and stagger them so they don't trigger
	// PLX API authentication conflicts.
	c.Flags.Int64Var(&c.Offset, "offset", 150, "Offset seconds between 0 and 300")

	return c.Flags
}

func (c *TradeCommand) Validate() {
	if len(c.Market) == 0 {
		log.Fatal(c.Help())
	}

	if c.Offset < 0 || c.Offset > 300 {
		log.Fatal(c.Help())
	}
}

func (c *TradeCommand) Run(args []string) int {
	c.InitFlags()
	c.Flags.Parse(args)
	c.Validate()
	c.InitDB()

	c.TradeContinuously(TRADE_INTERVAL)

	return 0
}

func (c *TradeCommand) TradeContinuously(interval int64) {
	// This is a slow frequency trader, and 2 seconds is the absolute minimum
	// we can support.
	if interval < 2 {
		log.Fatal("SLOW DOWN!!! Minimum trading interval is 2 seconds.")
	}

	lastRunTime := int64(0)

	for {
		currentRunTime := time.Now().Unix()
		isRuntime := currentRunTime%interval == c.Offset && currentRunTime-lastRunTime >= interval/2

		if !isRuntime {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		lastRunTime = currentRunTime

		err := c.TradeOnce()
		if err != nil {
			log.Printf("ERROR: %s" + err.Error())
		}
	}
}

func (c *TradeCommand) InitDB() {
	dbPath := LIVE_DB

	dbStore, err := db.NewBoltStore(c.Market, dbPath)
	if err != nil {
		log.Fatal(err)
	}

	c.DBStore = dbStore
}

func (c *TradeCommand) TradeOnce() error {
	trader, err := c.InitTrader()

	if err != nil {
		return err
	}

	return trader.Trade()
}

func (c *TradeCommand) InitTrader() (*trading.Trader, error) {
	plxClient := plx.NewLiveClient()
	plxExchange := trading.NewPlxExchange(plxClient)

	traderConfig, err := c.LoadTraderConfig()
	if err != nil {
		return nil, err
	}

	trader, err := trading.NewTrader(c.Market, plxExchange, c.DBStore, traderConfig)
	if err != nil {
		return nil, err
	}

	return trader, nil
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
