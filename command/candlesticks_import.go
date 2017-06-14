package command

import (
	//"github.com/jbgo/sftbot/candlesticks"
	"github.com/jbgo/sftbot/data"
	"log"
)

type CandlesticksImportCommand struct {
}

func (c *CandlesticksImportCommand) Synopsis() string {
	return "import PLX candlesticks data to trading bot DB"
}

func (c *CandlesticksImportCommand) Help() string {
	return formatHelpText(`
Usage: sftbot candlesticks import [options]

  Import Poloniex candlesticks data.

Options:

  -days=N                 The number of days worth of candlesticks data to
                          import.
                          Default: 7

  -resolution=T           Resolution of candlesticks data.
                          Default: 15m
                          Choices: 5m, 15m, 30m, 30m, 60m, ...

  -currenctPair=BTC_XYZ   PLX currency pair for trading data.
                          Must be in the format BTC_XYZ

  -continuous             Continuously run and import new candlesticks data
                          at the resolution interval.
                          Default: false
  `)
}

func (c *CandlesticksImportCommand) Run(args []string) int {
	db, err := data.OpenDB()

	if err != nil {
		log.Println(err)
		return 1
	}

	defer db.Close()

	//db.Candlesticks.Put("BTC_XRP", &candlesticks)
	// Open database
	// Create buckets (if not exist)
	// loop {
	//    query API
	//    write data
	// }
	return 0
}
