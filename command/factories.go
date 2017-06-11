package command

import (
	"github.com/mitchellh/cli"
	"strings"
)

func formatHelpText(helpText string) string {
	return strings.TrimSpace(helpText)
}

func Candlesticks() (cli.Command, error) {
	return &CandlesticksCommand{}, nil
}

func CandlesticksGet() (cli.Command, error) {
	return &CandlesticksGetCommand{}, nil
}

func CandlesticksImport() (cli.Command, error) {
	return &CandlesticksImportCommand{}, nil
}