package command

type CandlesticksCommand struct {
}

func (c *CandlesticksCommand) Synopsis() string {
	return "commands for working with PLX candlesticks data"
}

func (c *CandlesticksCommand) Help() string {
	return c.Synopsis()
}

func (c *CandlesticksCommand) Run(args []string) int {
	return 0
}
