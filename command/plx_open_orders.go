package command

import (
	"fmt"
	"github.com/jbgo/sftbot/plx"
	"log"
	"strings"
)

type OpenOrdersCommand struct {
}

func (c *OpenOrdersCommand) Synopsis() string {
	return "list open orders"
}

func (c *OpenOrdersCommand) Help() string {
	return c.Synopsis()
}

func (c *OpenOrdersCommand) Run(args []string) int {
	client := plx.Client{BaseUrl: plx.LIVE_URL}

	marketOrders, err := client.AllOpenOrders()

	if err != nil {
		log.Fatal(err)
	}

	for market, orders := range marketOrders {
		if len(orders) == 0 {
			continue
		}

		fmt.Printf("=== %s ===\n", market)

		for _, order := range orders {
			fmt.Printf(
				"%s rate=%0.9f amount=%0.9f total=%0.9f\n orderNumber=%d",
				strings.ToUpper(order.Type),
				order.Rate,
				order.Amount,
				order.Total,
				order.Number)
		}
	}

	return 0
}
