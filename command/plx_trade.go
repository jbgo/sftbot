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
	"flag"
	"fmt"
	"github.com/jbgo/sftbot/trading"
	"log"
	"time"
)

type TradeCommand struct {
	Flags    *flag.FlagSet
	Market   string
	StopTime time.Time
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

	return c.Flags
}

func (c *TradeCommand) Run(args []string) int {
	c.InitFlags()
	c.Flags.Parse(args)

	if len(c.Market) == 0 {
		fmt.Println(c.Help())
		return 1
	}

	plxExchange := &trading.PlxExchange{}
	trader, err := trading.NewTrader(c.Market, plxExchange)

	if err != nil {
		fmt.Println("[error]", err)
		return 1
	}

	log.Printf("market=%s action=trading.start\n", c.Market)

	for {
		err = trader.Trade()

		if err != nil {
			log.Printf("level=error market=%s error=\"%s\"\n", c.Market, err)
		}

		time.Sleep(60 * time.Second)
	}

	return 0
}
