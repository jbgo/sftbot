package command

import ()

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

  TBD
  `)
}

func (c *CandlesticksListCommand) Run(args []string) int {
	return 0
}
