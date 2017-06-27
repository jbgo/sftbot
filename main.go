package main

import (
	"github.com/jbgo/sftbot/command"
	"github.com/jbgo/sftbot/data"
	"github.com/mitchellh/cli"
	"log"
	"os"
)

const VERSION = "0.1.0"

func main() {
	c := cli.NewCLI("sftbot", VERSION)

	c.Args = os.Args[1:]

	c.Commands = map[string]cli.CommandFactory{
		"plx balances":      command.Balances,
		"chart-data get":    command.ChartDataGet,
		"chart-data list":   command.ChartDataList,
		"chart-data import": command.ChartDataImport,
		"plx orders":        command.OpenOrders,
		"simulate":          command.Simulate,
		"plx ticker":        command.Ticker,
		"plx trade-history": command.TradeHistory,
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
