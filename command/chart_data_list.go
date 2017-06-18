package command

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/jbgo/sftbot/data"
	"log"
)

type ChartDataListCommand struct {
	Flags        *flag.FlagSet
	CurrencyPair string
}

func (c *ChartDataListCommand) Synopsis() string {
	return "list imported PLX chart data"
}

func (c *ChartDataListCommand) Help() string {
	return formatHelpText(`
Usage: sftbot chart-data get [options]

  List the imported PLX chart data (a.k.a. candlesticks).

Options:

  ` + c.FlagOptionsString())
}

func (c *ChartDataListCommand) FlagOptionsString() string {
	c.InitFlags()

	options := ""
	buf := bytes.NewBufferString(options)
	c.Flags.SetOutput(buf)
	c.Flags.PrintDefaults()

	return buf.String()
}

func (c *ChartDataListCommand) InitFlags() {
	c.Flags = flag.NewFlagSet("chart-data list", flag.PanicOnError)
	c.Flags.StringVar(&c.CurrencyPair, "currency-pair", "", "PLX currency pair for chart data. Must be in the format BTC_XYZ")
}

func (c *ChartDataListCommand) Run(args []string) int {
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

	err = db.ForEachPeriod(c.CurrencyPair, func(d *data.ChartData) {
		fmt.Printf("%s t=%d vol=%0.9f wavg=%0.9f open=%0.9f close=%0.9f high=%0.9f low=%0.9f qvol=%0.9f\n",
			"BTC_XRP",
			d.Date,
			d.Volume,
			d.WeightedAverage,
			d.Open,
			d.Close,
			d.High,
			d.Low,
			d.QuoteVolume)
	})

	if err != nil {
		log.Println(err)
		return 1
	}

	return 0
}
