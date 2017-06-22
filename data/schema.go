package data

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

const DATABASE_NAME = "sftbot.db"

var BUCKET_NAMES = []string{
	"accounts",
	"transactions",
}

var CURRENCIES = []string{
	"BTC",
	"ETH",
	"LTC",
	"XRP",
	"STEEM",
	"ZEC",
	"BTS",
	"ETC",
	"SC",
	"DASH",
	"GNT",
	"DGB",
	"FCT",
	"XMR",
	"MAID",
}

type Store struct {
	*bolt.DB
}

// NOTE: expects caller to close the database (e.g. `defer db.Close()`)
func OpenDB() (*Store, error) {
	boltdb, err := bolt.Open(DATABASE_NAME, 0600, &bolt.Options{Timeout: 1 * time.Second})

	if err != nil {
		return nil, err
	}

	store := Store{boltdb}

	return &store, err
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

	err = db.Update(createCurrencyDataBuckets)
	if err != nil {
		return err
	}

	return err
}

func (db *Store) Write(bucketName string, key interface{}, value interface{}) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Put(db.EncodeKey(key), db.EncodeValue(value))
	})
}

func (db *Store) ForEachPeriod(currencyPair string, callback func(chartData *ChartData)) error {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("chart_data." + currencyPair))

		if b == nil {
			return fmt.Errorf("unsupported currency pair: %s", currencyPair)
		}

		var chartData ChartData

		return b.ForEach(func(k, v []byte) error {
			db.DecodeValue(v, &chartData)

			callback(&chartData)

			return nil
		})
	})
}

func (db *Store) EncodeKey(data interface{}) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, data)

	if err != nil {
		log.Panic(err)
	}

	return buf.Bytes()
}

func (db *Store) EncodeValue(data interface{}) []byte {
	encoded, err := json.Marshal(data)

	if err != nil {
		log.Panic(err)
	}

	return []byte(encoded)
}

func (db *Store) DecodeValue(data []byte, out interface{}) {
	err := json.Unmarshal(data, out)

	if err != nil {
		log.Panic(err)
	}
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

func createCurrencyDataBuckets(tx *bolt.Tx) error {
	for _, currency := range CURRENCIES {
		if currency == "BTC" {
			continue
		}

		_, err := tx.CreateBucketIfNotExists([]byte("chart_data.BTC_" + currency))

		if err != nil {
			return err
		}
	}

	return nil
}
