package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
)

const (
	filename = "data/mainnet.db"
)

// Sensor is a collection of Entry data
type Sensor struct {
	Data []Entry `json:"data"`
}

// Entry is the structure of the data entered
type Entry struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Time      string  `json:"time"`
	Timestamp int64   `json:"timestamp"`
	Value     float32 `json:"value"`
}

// getdbpath returns a string of the full file path
func getdbpath() string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	dbpath := filepath.Join(pwd, filename)
	return dbpath
}

// Loaddb loads the database if it already exists and creates one
// if none exists
func Loaddb() {
	db, err := bolt.Open(getdbpath(), 0644, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}

// Initialize is the initializer for the database
// it creates an instance of the database and retruns a struct of the database
func Initialize() *bolt.DB {
	db, err := bolt.Open(getdbpath(), 0644, nil)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Encode takes a value as a sturct and returns []byte
func Encode(value interface{}) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

// Decode takes the value and converts the data to a struct
func Decode(data string, holder interface{}) error {
	err := json.Unmarshal([]byte(data), &holder)
	if err != nil {
		return fmt.Errorf("cannot unmarshal json %v ", err)
	}
	return nil
}

// CreateBucketEntry generates a new bucket entry and attaches it to the database
func CreateBucketEntry(name string, key string, value interface{}) error {
	// change data to a json structure to be saved in the database
	entrydatajson, err := Encode(value)
	if err != nil {
		return fmt.Errorf("could not encode data %v ", err)
	}

	db := Initialize()
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		// first 8 characters of the uuid used for the bucket name
		b, err := tx.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		err = b.Put([]byte(key), []byte(entrydatajson))
		if err != nil {
			return fmt.Errorf("could not put %v ", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not update %v ", err)
	}
	return nil
}

// GetFromBucket function takes a key and returns the value as a json object
func GetFromBucket(name string, key string) string {
	db := Initialize()
	defer db.Close()

	var value []byte

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(name))
		if bucket == nil {
			return fmt.Errorf("Bucket not found")
		}
		value = bucket.Get([]byte(key))
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
	return string(value)
}

// NewNestedUser - this function takes a subbucket name as a string and an User interface
func NewNestedUser(subBucketName string, in interface{}) error {
	db := Initialize()
	defer db.Close()
	err := db.Update(func(tx *bolt.Tx) error {
		// Setup the users bucket.
		bkt, err := tx.CreateBucketIfNotExists([]byte("User"))
		if err != nil {
			return err
		}
		nesbucket, err := bkt.CreateBucket([]byte(subBucketName))
		if err != nil {
			return err
		}

		encoded, err := Encode(in)
		if err != nil {
			return fmt.Errorf("couldn't encode user data %v", err)
		}
		if err := nesbucket.Put([]byte(subBucketName), encoded); err != nil {
			return fmt.Errorf("couldn't create subbucket %v ", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not update %v ", err)
	}
	return nil
}

// GetNestedUser - this function takes a subbucket name as a string and an User interface
func GetNestedUser(subBucketName string, in interface{}) error {
	db := Initialize()
	defer db.Close()

	var value []byte

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("User"))
		if bucket == nil {
			return fmt.Errorf("Bucket not found")
		}
		nestedBucket := bucket.Bucket([]byte(subBucketName))
		if nestedBucket == nil {
			return fmt.Errorf("Subbucket not found")
		}
		value = nestedBucket.Get([]byte(subBucketName))
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not update %v ", err)
	}

	if err := Decode(string(value), &in); err != nil {
		return err
	}
	return nil
}
