package command

import (
	"flag"
	"fmt"
	"github.com/jbgo/sftbot/plx"
	"log"
	"strings"
)

type TickerCommand struct {
	Flags   *flag.FlagSet
	Markets string
}

func (c *TickerCommand) Synopsis() string {
	return "view current price and volume for given markets"
}

func (c *TickerCommand) Help() string {
	return formatHelpText(`
Usage: sftbot plx ticker [options]

  ` + c.Synopsis() + `

` + helpOptions(c))
}

func (c *TickerCommand) InitFlags() *flag.FlagSet {
	c.Flags = flag.NewFlagSet("plx ticker", flag.ContinueOnError)

	c.Flags.StringVar(&c.Markets, "markets", "", "Comma-separated list of PLX markets specified as currency pair (e.g. BTC_XYZ)")

	return c.Flags
}

func (c *TickerCommand) Run(args []string) int {
	c.InitFlags()
	c.Flags.Parse(args)

	client := plx.NewLiveClient()

	ticker, err := client.GetTicker()

	if err != nil {
		log.Panic(err)
	}

	for _, t := range ticker {
		if !strings.Contains(c.Markets, t.Market) && len(c.Markets) != 0 {
			continue
		}

		fmt.Printf(
			"%s last=%0.9f lowestAsk=%0.9f highestBid=%0.9f percentChange=%0.9f baseVolume=%0.9f quoteVolume=%0.9f\n",
			t.Market,
			t.Last,
			t.LowestAsk,
			t.HighestBid,
			t.PercentChange,
			t.BaseVolume,
			t.QuoteVolume)
	}

	return 0
}
