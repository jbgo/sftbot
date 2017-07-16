package command

import (
	"github.com/mitchellh/cli"
	"strings"
)

func formatHelpText(helpText string) string {
	return strings.TrimSpace(helpText)
}

func Balances() (cli.Command, error) {
	return &BalancesCommand{}, nil
}

func ChartDataList() (cli.Command, error) {
	return &ChartDataListCommand{}, nil
}

func ChartDataImport() (cli.Command, error) {
	return &ChartDataImportCommand{}, nil
}

func MyTrades() (cli.Command, error) {
	return &MyTradesCommand{}, nil
}

func OpenOrders() (cli.Command, error) {
	return &OpenOrdersCommand{}, nil
}

func Simulate() (cli.Command, error) {
	return &SimulateCommand{}, nil
}

func Ticker() (cli.Command, error) {
	return &TickerCommand{}, nil
}

func Trade() (cli.Command, error) {
	return &TradeCommand{}, nil
}

func TradeHistory() (cli.Command, error) {
	return &TradeHistoryCommand{}, nil
}
