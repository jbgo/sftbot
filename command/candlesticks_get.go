package command

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
	return 0
}
