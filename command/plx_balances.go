package command

import (
	"fmt"
	"github.com/jbgo/sftbot/plx"
	"log"
)

type BalancesCommand struct {
}

func (c *BalancesCommand) Synopsis() string {
	return "list all balances"
}

func (c *BalancesCommand) Help() string {
	return c.Synopsis()
}

func (c *BalancesCommand) Run(args []string) int {
	balances, err := plx.CompleteBalances()

	if err != nil {
		log.Fatal(err)
	}

	totalBtcValue := 0.0

	fmt.Println("---")

	for _, balance := range balances {
		totalBtcValue += balance.BtcValue
		if balance.BtcValue > 0.0 {
			fmt.Printf("%-8s available=%0.9f    on_orders=%0.9f    btc_value=%0.9f\n",
				balance.Currency, balance.Available, balance.OnOrders, balance.BtcValue)
		}
	}

	fmt.Printf("---\nTotal BTC value: %0.9f\n", totalBtcValue)

	return 0
}
