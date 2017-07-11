package command

import (
	"flag"
	"fmt"
	"github.com/jbgo/sftbot/plx"
	"log"
	"time"
)

type TradeHistoryCommand struct {
	Flags  *flag.FlagSet
	Market string
}

func (c *TradeHistoryCommand) Synopsis() string {
	return "get recent trades for a given market"
}

func (c *TradeHistoryCommand) Help() string {
	return formatHelpText(`
Usage: sftbot plx trade-history [options]

  ` + c.Synopsis() + `

` + helpOptions(c))
}

func (c *TradeHistoryCommand) InitFlags() *flag.FlagSet {
	c.Flags = flag.NewFlagSet("plx trade-history", flag.PanicOnError)

	c.Flags.StringVar(&c.Market, "market", "", "PLX market specified as currency pair (e.g. BTC_XYZ)")

	return c.Flags
}

func (c *TradeHistoryCommand) Run(args []string) int {
	c.InitFlags()
	c.Flags.Parse(args)

	if len(c.Market) == 0 {
		fmt.Println(c.Help())
		return 1
	}

	endTime := time.Now().Unix() + 10
	startTime := endTime - 100

	params := plx.TradeHistoryParams{
		Market:    c.Market,
		StartTime: startTime,
		EndTime:   endTime,
	}

	client := plx.Client{BaseUrl: plx.LIVE_URL}

	trades, err := client.GetTradeHistory(&params)

	if err != nil {
		log.Println(err)
		return 1
	}

	for _, t := range trades {
		fmt.Printf("%-4s %s    rate=%0.9f    amount=%0.9f    total=%0.9f\n",
			t.Type, t.Date, t.Rate, t.Amount, t.Total)
	}

	return 0
}
