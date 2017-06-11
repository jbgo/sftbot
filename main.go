package main

import (
	"encoding/json"
	"fmt"
	"github.com/jbgo/sftbot/command"
	"github.com/mitchellh/cli"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const VERSION = "0.1.0"

const PUBLIC_API_URL = "https://poloniex.com/public"

type ChartDataParams struct {
	command      string
	currencyPair string
	start        uint
	end          uint
	period       uint
}

type Candlestick struct {
	Date            uint64
	High            float64
	Low             float64
	Open            float64
	Close           float64
	Volume          float64
	QuoteVolume     float64
	WeightedAverage float64
}

func (p *ChartDataParams) ToQueryString() string {
	return fmt.Sprintf(
		"command=%s&currencyPair=%s&start=%d&end=%d&period=%d",
		p.command,
		p.currencyPair,
		p.start,
		p.end,
		p.period)
}

func candlesticks() {
	params := ChartDataParams{"returnChartData", "BTC_XRP", 1497033851, 1497120251, 1800}

	url := fmt.Sprintf("%s?%s", PUBLIC_API_URL, params.ToQueryString())

	resp, err := http.Get(url)
	if err != nil {
		log.Panic(err)
	}

	if resp.StatusCode != 200 {
		log.Panic("request failed:", resp.Status)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	log.Printf("%d %s GET %s", len(body), resp.Status, url)

	candlesticks := make([]Candlestick, 24)
	err = json.Unmarshal(body, &candlesticks)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}

func usage(subCommand string) {
	log.Printf("TODO: usage for %s", subCommand)
	os.Exit(1)
}

func main() {
	c := cli.NewCLI("sftbot", VERSION)

	c.Args = os.Args[1:]

	c.Commands = map[string]cli.CommandFactory{
		"candlesticks":        command.Candlesticks,
		"candlesticks get":    command.CandlesticksGet,
		"candlesticks import": command.CandlesticksImport,
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
