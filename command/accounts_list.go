package command

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jbgo/sftbot/data"
	"log"
)

type AccountsListCommand struct {
}

func (c *AccountsListCommand) Synopsis() string {
	return "list all currency accounts and balances"
}

func (c *AccountsListCommand) Help() string {
	return c.Synopsis()
}

func (c *AccountsListCommand) Run(args []string) int {
	db, err := data.OpenDB()

	if err != nil {
		log.Println(err)
		return 1
	}

	defer db.Close()

	var accounts []data.Account

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("accounts"))

		txErr := b.ForEach(func(k, v []byte) error {
			var account data.Account

			loopErr := json.Unmarshal(v, &account)
			if loopErr != nil {
				return loopErr
			}

			accounts = append(accounts, account)

			return nil
		})

		return txErr
	})

	if err != nil {
		log.Println(err)
		return 1
	}

	for _, account := range accounts {
		fmt.Printf("Currency: %-8s Balance: %0.9f\n", account.Name, account.Balance)
	}

	return 0
}
