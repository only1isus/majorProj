package database

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dgraph-io/badger"
	"github.com/only1isus/majorProj/util"
)

// Encode takes structured data and returns a slice of bytes
func Encode(data util.Log) ([]byte, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("got: %v", err)
	}
	return encoded, nil
}

// Decode takes a slice of bytes and returns an interface of the structured data
// func Decode(data string) (interface{}, error) {
// 	var holder util.Log
// 	unmarshalled := data.([]byte)
// 	err := json.Unmarshal(unmarshalled, &holder)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return holder, nil
// }

func intializer() *badger.DB {
	opts := badger.DefaultOptions
	opts.Dir = "./data"
	opts.ValueDir = "./data"

	db, err := badger.Open(opts)
	if err != nil {
		log.Println("couldn't create the database")
	}

	return db
}

// NewEnty takes a key and a value and appends it to the database
func NewEnty(key string, value []byte) {
	db := intializer()
	defer db.Close()
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(value))
		return err
	})

	if err != nil {
		fmt.Println("cannot open database")
	}
}

// GetEntry takes a key searches the database and returns an interface
func GetEntry(key string) (interface{}, error) {
	db := intializer()
	defer db.Close()

	err := db.View(func(txn *badger.Txn) error {
		data, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		// d := make(chan []byte)
		util.PlaceHolder, err = data.Value()
		if err != nil {
			fmt.Printf("got an error %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return string(util.PlaceHolder), nil
}
