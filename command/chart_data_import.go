package command

import (
	"flag"
	"fmt"
	"github.com/jbgo/sftbot/data"
	"github.com/jbgo/sftbot/plx"
	"log"
	"time"
)

type ChartDataImportCommand struct {
	Flags *flag.FlagSet

	CurrencyPair string

	Continuous bool
	Days       int64
	Resolution int64
}

func (c *ChartDataImportCommand) Synopsis() string {
	return "import PLX chart data to trading bot DB"
}

func (c *ChartDataImportCommand) Help() string {
	return formatHelpText(`
Usage: sftbot chart-data import [options]

  Import Poloniex chart data (a.k.a. candlesticks).

` + helpOptions(c))
}

func (c *ChartDataImportCommand) InitFlags() *flag.FlagSet {
	c.Flags = flag.NewFlagSet("chart-data import", flag.PanicOnError)

	c.Flags.StringVar(&c.CurrencyPair, "currency-pair", "", "PLX currency pair for chart data. Must be in the format BTC_XYZ")
	c.Flags.BoolVar(&c.Continuous, "continuous", false, "If true, continuously run and import new chart data at the specified resolution.")
	c.Flags.Int64Var(&c.Days, "days", 7, "The number of days worth of chart data to import.")
	c.Flags.Int64Var(&c.Resolution, "resolution", 300, "Resolution of chart data in seconds. Choices: 300, 900, 1800, 7200, 1440, 86400")

	return c.Flags
}

func (c *ChartDataImportCommand) Parse(args []string) error {
	c.Flags.Parse(args)

	if len(c.CurrencyPair) == 0 {
		return fmt.Errorf("-currency-pair is required")
	}

	return nil
}

func (c *ChartDataImportCommand) Run(args []string) int {
	c.InitFlags()

	err := c.Parse(args)

	if err != nil {
		fmt.Println(c.Help())
		return 1
	}

	db, err := data.OpenDB()

	if err != nil {
		log.Println(err)
		return 1
	}

	defer db.Close()

	client := plx.Client{BaseUrl: plx.LIVE_URL}

	params := plx.ChartDataParams{
		CurrencyPair: c.CurrencyPair,
		Period:       c.Resolution,
	}

	referenceTime := time.Now().Unix()
	oneDay, _ := time.ParseDuration("24h")
	secondsPerDay := int64(oneDay.Seconds())

	var numDays int64 = 0

	for numDays < c.Days {
		params.Start = referenceTime - ((c.Days - numDays) * secondsPerDay)
		params.End = params.Start + secondsPerDay

		log.Printf("loading chart data for %v - %v", time.Unix(params.Start, 0), time.Unix(params.End, 0))

		chartData, err := client.GetChartData(&params)

		if err != nil {
			log.Println(err)
			return 1
		}

		for _, p := range chartData {
			err = db.Write("chart_data."+c.CurrencyPair, p.Date, p)

			if err != nil {
				log.Println(err)
				return 1
			}
		}

		numDays += 1

		// avoid public API rate limits
		time.Sleep(250 * time.Millisecond)
	}

	return 0
}
