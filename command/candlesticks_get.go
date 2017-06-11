package command

import (
	"encoding/json"
	"fmt"
	"github.com/jbgo/sftbot/candlesticks"
	"log"
	"os"
	"time"
)

type CandlesticksGetCommand struct {
}

func (c *CandlesticksGetCommand) Synopsis() string {
	return "get latest candlesticks data"
}

func (c *CandlesticksGetCommand) Help() string {
	return formatHelpText(`
Usage: sftbot candlesticks get [options]

  Get the latest PLX candlesticks data.

Options:

  TBD
  `)
}

func (c *CandlesticksGetCommand) Run(args []string) int {
	endTime := time.Now().Unix()
	startTime := endTime - (60 * 60 * 24 * 1)

	params := candlesticks.ChartDataParams{
		CurrencyPair: "BTC_XRP",
		Start:        startTime,
		End:          endTime,
		Period:       300,
	}

	data, err := candlesticks.GetChartData(&params)

	if err != nil {
		log.Println(err)
		return 1
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return 1
	}

	fmt.Println(string(encoded))
	os.Stdout.Write(encoded)

	return 0
}
