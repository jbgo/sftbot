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
		"accounts":            command.Accounts,
		"candlesticks":        command.Candlesticks,
		"candlesticks get":    command.CandlesticksGet,
		"candlesticks import": command.CandlesticksImport,
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
