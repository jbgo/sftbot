package main

import (
	"github.com/jbgo/sftbot/command"
	"github.com/jbgo/sftbot/data"
	"github.com/mitchellh/cli"
	"log"
	"os"
)

const VERSION = "0.4.2"

func main() {
	c := cli.NewCLI("sftbot", VERSION)

	c.Args = os.Args[1:]

	c.Commands = map[string]cli.CommandFactory{
		"chart-data list":   command.ChartDataList,
		"chart-data import": command.ChartDataImport,
		"plx balances":      command.Balances,
		"plx my-trades":     command.MyTrades,
		"plx orders":        command.OpenOrders,
		"plx prospect":      command.Prospect,
		"plx ticker":        command.Ticker,
		"plx trade":         command.Trade,
		"plx trade-history": command.TradeHistory,
		"simulate":          command.Simulate,
	}

	err := data.InitSchema()
	if err != nil {
		log.Fatal(err)
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
