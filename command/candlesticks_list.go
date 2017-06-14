package command

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/jbgo/sftbot/data"
	"log"
)

type CandlesticksListCommand struct {
	Flags        *flag.FlagSet
	CurrencyPair string
}

func (c *CandlesticksListCommand) Synopsis() string {
	return "list imported candlesticks data"
}

func (c *CandlesticksListCommand) Help() string {
	return formatHelpText(`
Usage: sftbot candlesticks get [options]

  List the imported PLX candlesticks data.

Options:

  ` + c.FlagOptionsString())
}

func (c *CandlesticksListCommand) FlagOptionsString() string {
	c.InitFlags()

	options := ""
	buf := bytes.NewBufferString(options)
	c.Flags.SetOutput(buf)
	c.Flags.PrintDefaults()

	return buf.String()
}

func (c *CandlesticksListCommand) InitFlags() {
	c.Flags = flag.NewFlagSet("candlesticks list", flag.PanicOnError)
	c.Flags.StringVar(&c.CurrencyPair, "currency-pair", "", "PLX currency pair for trading data. Must be in the format BTC_XYZ")
}

func (c *CandlesticksListCommand) Run(args []string) int {
	c.InitFlags()
	c.Flags.Parse(args)

	if len(c.CurrencyPair) == 0 {
		fmt.Println(c.Help())
		return 1
	}

	db, err := data.OpenDB()

	if err != nil {
		log.Println(err)
		return 1
	}

	defer db.Close()

	err = db.ForEachPeriod(c.CurrencyPair, func(stick *data.Candlestick) {
		fmt.Printf("exchange: %-8s date: %-16d open: %0.9f    close: %0.9f\n", "BTC_XRP", stick.Date, stick.Open, stick.Close)
	})

	if err != nil {
		log.Println(err)
		return 1
	}

	return 0
}
