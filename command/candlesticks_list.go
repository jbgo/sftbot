package command

import (
	"fmt"
	"github.com/jbgo/sftbot/data"
	"log"
)

type CandlesticksListCommand struct {
}

func (c *CandlesticksListCommand) Synopsis() string {
	return "list imported candlesticks data"
}

func (c *CandlesticksListCommand) Help() string {
	return formatHelpText(`
Usage: sftbot candlesticks get [options]

  List the imported PLX candlesticks data.

Options:

  -currenctPair=BTC_XYZ   PLX currency pair for trading data.
                          Must be in the format BTC_XYZ
  `)
}

func (c *CandlesticksListCommand) Run(args []string) int {
	db, err := data.OpenDB()

	if err != nil {
		log.Println(err)
		return 1
	}

	defer db.Close()

	err = db.ForEachPeriod("BTC_XRP", func(stick *data.Candlestick) {
		fmt.Printf("exchange: %-8s date: %-16d open: %0.9f    close: %0.9f\n", "BTC_XRP", stick.Date, stick.Open, stick.Close)
	})

	if err != nil {
		log.Println(err)
		return 1
	}

	return 0
}
