package command

import (
	"flag"
	"fmt"
	"github.com/jbgo/sftbot/data"
	"log"
	"time"
)

type ChartDataListCommand struct {
	Flags  *flag.FlagSet
	Market string

	StartTimeVar string
	StartTime    time.Time

	EndTimeVar string
	EndTime    time.Time
}

func (c *ChartDataListCommand) Synopsis() string {
	return "list imported PLX chart data"
}

func (c *ChartDataListCommand) Help() string {
	return formatHelpText(`
Usage: sftbot chart-data list [options]

  List the imported PLX chart data (a.k.a. candlesticks).

` + helpOptions(c))
}

func (c *ChartDataListCommand) InitFlags() *flag.FlagSet {
	c.Flags = flag.NewFlagSet("chart-data list", flag.PanicOnError)

	c.Flags.StringVar(&c.Market, "market", "", "PLX market as a currency pair. Must be in the format BTC_XYZ")
	c.Flags.StringVar(&c.StartTimeVar, "start-time", "", "Start of time range. YYYY-MM-DD HH:MM:SS")
	c.Flags.StringVar(&c.EndTimeVar, "end-time", "", "End of time range. YYYY-MM-DD HH:MM:SS)")

	return c.Flags
}

func (c *ChartDataListCommand) Validate() {
	var err error

	if len(c.Market) == 0 {
		log.Fatal(c.Help())
	}

	if len(c.StartTimeVar) > 0 {
		c.StartTime, err = time.Parse(TIME_VAR_FORMAT, c.StartTimeVar)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		c.StartTime = time.Now().Add(-(3 * 60 * 60))
	}

	if len(c.EndTimeVar) > 0 {
		c.EndTime, err = time.Parse(TIME_VAR_FORMAT, c.EndTimeVar)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		c.EndTime = time.Now()
	}
}

func (c *ChartDataListCommand) Run(args []string) int {
	c.InitFlags()
	c.Flags.Parse(args)
	c.Validate()

	db, err := data.OpenDB()

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = db.ForEachPeriod(c.Market, func(d *data.ChartData) {
		if d.Date < c.StartTime.Unix() || d.Date > c.EndTime.Unix() {
			return
		}

		timestamp := time.Unix(d.Date, 0).Format("2006-01-02 15:04:05 MST")

		fmt.Printf("%s [%s] vol=%0.9f wavg=%0.9f open=%0.9f close=%0.9f high=%0.9f low=%0.9f qvol=%0.9f\n",
			c.Market,
			timestamp,
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
