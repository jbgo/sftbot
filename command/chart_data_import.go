package command

import (
	"github.com/jbgo/sftbot/data"
	"github.com/jbgo/sftbot/plx"
	"log"
	"time"
)

type ChartDataImportCommand struct {
}

func (c *ChartDataImportCommand) Synopsis() string {
	return "import PLX chart data to trading bot DB"
}

func (c *ChartDataImportCommand) Help() string {
	return formatHelpText(`
Usage: sftbot chart-data import [options]

  Import Poloniex chart data (a.k.a. candlesticks).

Options:

  -days=N                 The number of days worth of chart data to import.
                          Default: 7

  -resolution=T           Resolution of chart data.
                          Default: 15m
                          Choices: 5m, 15m, 30m, 30m, 60m, ...

  -currenctPair=BTC_XYZ   PLX currency pair for chart data.
                          Must be in the format BTC_XYZ

  -continuous             Continuously run and import new candlesticks data at
                          the resolution interval.
                          Default: false
  `)
}

func (c *ChartDataImportCommand) Run(args []string) int {
	db, err := data.OpenDB()

	if err != nil {
		log.Println(err)
		return 1
	}

	defer db.Close()

	endTime := time.Now().Unix()
	startTime := endTime - (60 * 60 * 24 * 1)

	params := plx.ChartDataParams{
		CurrencyPair: "BTC_XRP",
		Start:        startTime,
		End:          endTime,
		Period:       300,
	}

	chartData, err := plx.GetChartData(&params)

	for _, p := range *chartData {
		err = db.Write("candlesticks.BTC_XRP", p.Date, p)

		if err != nil {
			log.Println(err)
			return 1
		}
	}

	return 0
}
