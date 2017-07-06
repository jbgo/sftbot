package db

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"time"
)

type Store interface {
	Read(key string, outValue interface{}) error
	Write(key string, value interface{}) error
	Delete(key string) error
	HasData(key string) (error, bool)
}

type BoltStore struct {
	BucketName string
	DBFile     string
}

func NewBoltStore(bucketName, dbFile string) (*BoltStore, error) {
	boltStore := &BoltStore{
		BucketName: bucketName,
		DBFile:     dbFile,
	}

	boltdb, err := openBoltDB(boltStore)
	defer boltdb.Close()

	return boltStore, err
}

func openBoltDB(store *BoltStore) (*bolt.DB, error) {
	boltdb, err := bolt.Open(store.DBFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return boltdb, err
	}

	err = boltdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(store.BucketName))
		return err
	})

	return boltdb, err
}

func (store *BoltStore) Read(key string, outValue interface{}) error {
	boltdb, err := openBoltDB(store)
	if err != nil {
		return err
	}

	defer boltdb.Close()

	return boltdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(store.BucketName))

		encoded := b.Get([]byte(key))

		return json.Unmarshal(encoded, &outValue)
	})
}

func (store *BoltStore) Write(key string, value interface{}) error {
	boltdb, err := openBoltDB(store)
	if err != nil {
		return err
	}

	defer boltdb.Close()

	return boltdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(store.BucketName))

		encoded, err := json.Marshal(value)
		if err != nil {
			return err
		}

		return b.Put([]byte(key), []byte(encoded))
	})
}

func (store *BoltStore) Delete(key string) error {
	boltdb, err := openBoltDB(store)
	if err != nil {
		return err
	}

	defer boltdb.Close()

	return boltdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(store.BucketName))
		return b.Delete([]byte(key))
	})
}

func (store *BoltStore) HasData(key string) (err error, hasData bool) {
	boltdb, err := openBoltDB(store)
	if err != nil {
		return err, false
	}

	defer boltdb.Close()

	err = boltdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(store.BucketName))
		data := b.Get([]byte(key))
		hasData = len(data) > 0
		return nil
	})

	return err, hasData
}
