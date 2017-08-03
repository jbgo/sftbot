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
	"runtime"
	"sort"
	"strings"
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
	c.Flags.StringVar(&c.Config, "config", "", "Trader config file (JSON)")
	return c.Flags
}

func (c *TradeCommand) Run(args []string) int {
	c.InitFlags()
	c.Flags.Parse(args)
	c.InitDB()

	c.TradeContinuously(TRADE_INTERVAL)

	return 0
}

func checkAndLog(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Printf(`evt=error msg="%s" loc=%s:%d`, err.Error(), file, line)
	}
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
		isRuntime := currentRunTime%interval == 0 && currentRunTime-lastRunTime >= interval/2

		if !isRuntime {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		lastRunTime = currentRunTime

		err := c.TradeOnce()
		if err != nil {
			checkAndLog(err)
		}
	}
}

func (c *TradeCommand) InitDB() {
	dbPath := LIVE_DB

	dbStore, err := db.NewBoltStore("trading", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	c.DBStore = dbStore
}
func (c *TradeCommand) TradeOnce() error {
	client := plx.NewLiveClient()
	ticker, err := client.GetTicker()

	if err != nil {
		return err
	}

	sort.Sort(ByVolumeDesc(ticker))

	for _, t := range ticker {
		interesting := strings.Contains(t.Market, "BTC_") && t.BaseVolume >= 500

		if !interesting {
			continue
		}

		trader, err := c.InitTrader(t.Market)
		checkAndLog(err)

		if err != nil {
			continue
		}

		err = trader.Trade()
		checkAndLog(err)
	}

	return nil
}

func (c *TradeCommand) InitTrader(marketName string) (*trading.Trader, error) {
	plxClient := plx.NewLiveClient()
	plxExchange := trading.NewPlxExchange(plxClient)

	traderConfig, err := c.LoadTraderConfig()
	if err != nil {
		return nil, err
	}

	trader, err := trading.NewTrader(marketName, plxExchange, c.DBStore, traderConfig)
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
