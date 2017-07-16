package command

import (
	"flag"
	"fmt"
	"github.com/jbgo/sftbot/plx"
	"log"
	"time"
)

type MyTradesCommand struct {
	Flags  *flag.FlagSet
	Market string

	StartTimeVar string
	StartTime    time.Time

	EndTimeVar string
	EndTime    time.Time
}

func (c *MyTradesCommand) Synopsis() string {
	return "get my recent trades for a given market"
}

func (c *MyTradesCommand) Help() string {
	return formatHelpText(`
Usage: sftbot plx my-trades [options]

  ` + c.Synopsis() + `

` + helpOptions(c))
}

func (c *MyTradesCommand) InitFlags() *flag.FlagSet {
	c.Flags = flag.NewFlagSet("plx my-trades", flag.PanicOnError)

	c.Flags.StringVar(&c.Market, "market", "", "PLX market specified as currency pair (e.g. BTC_XYZ)")
	c.Flags.StringVar(&c.StartTimeVar, "start-time", "2017-05-01 00:00:00 CST", "Start of time range. YYYY-MM-DD HH:MM:SS (TZ)")
	c.Flags.StringVar(&c.EndTimeVar, "end-time", "", "End of time range. YYYY-MM-DD HH:MM:SS (TZ)")

	return c.Flags
}

func (c *MyTradesCommand) Validate() (err error) {
	if len(c.Market) == 0 {
		return fmt.Errorf(c.Help())
	}

	c.StartTime, err = time.Parse(TIME_VAR_FORMAT, c.StartTimeVar)
	if err != nil {
		return err
	}

	if len(c.EndTimeVar) > 0 {
		c.EndTime, err = time.Parse(TIME_VAR_FORMAT, c.EndTimeVar)
		if err != nil {
			return err
		}
	} else {
		c.EndTime = time.Now()
	}

	return nil
}

func (c *MyTradesCommand) Run(args []string) int {
	c.InitFlags()
	c.Flags.Parse(args)

	err := c.Validate()
	if err != nil {
		log.Println(err)
		return 1
	}

	client := plx.NewLiveClient()

	tradeHistory, err := client.MyTradeHistory(c.Market, c.StartTime.Unix(), c.EndTime.Unix())

	if err != nil {
		log.Println(err)
		return 1
	}

	for market, trades := range tradeHistory {
		fmt.Printf("=== %s ===\n", market)
		for _, t := range trades {
			fmt.Printf("%-4s %s    rate=%0.9f    amount=%0.9f    fee=%0.9f    total=%0.9f    order=%s\n",
				t.Type, t.Date, t.Rate, t.Amount, t.Fee, t.Total, t.OrderNumber)
		}
	}

	return 0
}
