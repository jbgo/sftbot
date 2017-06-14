package data

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"time"
)

const DATABASE_NAME = "sftbot.db"

var BUCKET_NAMES = []string{
	"accounts",
	"transactions",
	"candlesticks",
}

var CURRENCIES = []string{
	"BTC",
	"ETH",
	"LTC",
	"XRP",
	"STEEM",
	"ZEC",
}

// NOTE: expects caller to close the database (e.g. `defer db.Close()`)
func OpenDB() (*bolt.DB, error) {
	return bolt.Open(DATABASE_NAME, 0600, &bolt.Options{Timeout: 1 * time.Second})
}

func InitSchema() error {
	db, err := OpenDB()
	if err != nil {
		return err
	}

	defer db.Close()

	err = db.Update(createBuckets)
	if err != nil {
		return err
	}

	err = db.Update(createEmptyAccounts)
	if err != nil {
		return err
	}

	return err
}

func createBuckets(tx *bolt.Tx) error {
	for _, bucketName := range BUCKET_NAMES {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))

		if err != nil {
			return err
		}
	}

	return nil
}

func createEmptyAccounts(tx *bolt.Tx) error {
	b := tx.Bucket([]byte("accounts"))

	for _, currency := range CURRENCIES {
		account := Account{currency, 0.0}
		key := []byte(account.Name)

		v := b.Get(key)
		if v != nil {
			continue
		}

		data, err := json.Marshal(account)
		if err != nil {
			return err
		}

		err = b.Put([]byte(account.Name), data)
		if err != nil {
			return err
		}
	}

	return nil
}
