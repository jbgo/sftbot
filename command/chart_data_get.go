package command

import (
	"encoding/json"
	"fmt"
	"github.com/jbgo/sftbot/plx"
	"log"
	"os"
	"time"
)

type ChartDataGetCommand struct {
}

func (c *ChartDataGetCommand) Synopsis() string {
	return "get latest candlesticks data"
}

func (c *ChartDataGetCommand) Help() string {
	return formatHelpText(`
Usage: sftbot candlesticks get [options]

  Get the latest PLX candlesticks data.

Options:

  TBD
  `)
}

func (c *ChartDataGetCommand) Run(args []string) int {
	endTime := time.Now().Unix()
	startTime := endTime - (60 * 60 * 24 * 1)

	client := plx.Client{BaseUrl: plx.LIVE_URL}

	params := plx.ChartDataParams{
		CurrencyPair: "BTC_XRP",
		Start:        startTime,
		End:          endTime,
		Period:       300,
	}

	data, err := client.GetChartData(&params)

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
