package main

import (
	"github.com/jbgo/sftbot/command"
	"github.com/mitchellh/cli"
	"log"
	"os"
)

const VERSION = "0.1.0"

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
