package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
)

const (
	filename = "data/main.db"
)

// SensorEntry is the structure ofrhe sensor data
type SensorEntry struct {
	Time       int64   `json:"time"`
	SensorType string  `json:"sensorType"`
	Value      float64 `json:"value"`
}

// Sensor struct holds []SensorEntry
type Sensor struct {
	Data []SensorEntry `json:"data"`
}

var (
	holder = make(map[string]string)
)

// getdbpath returns a string of the full file path
func getdbpath() string {
	var pwd string
	var err error
	pwd, err = os.Getwd()
	if !strings.Contains(pwd, "data") {
		fmt.Println("Creating the directory")
		err := os.MkdirAll(strings.Join([]string{pwd, "data"}, "/"), os.ModePerm)
		if err != nil {
			fmt.Println("please consider making the directory manually")
			os.Exit(1)
		}
		p, _ := os.Getwd()
		pwd = p
	}
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

// GetNestedUser takes a subbucket name as a string and an User interface
func GetNestedUser(subBucketName string, in interface{}) error {
	bucketName := strings.ToUpper("User")
	db := Initialize()
	defer db.Close()

	var value []byte

	err := db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(bucketName))
		if root == nil {
			return fmt.Errorf("Bucket not found")
		}
		value = root.Get([]byte(subBucketName))
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

// nestedEntry function takes a bucketname
// bucketName is rootbucket as a string
func nestedEntry(bucketName string, key string, value *[]byte) error {
	rootName := strings.ToUpper(bucketName)
	db := Initialize()
	defer db.Close()

	err := db.Update(func(tx *bolt.Tx) error {
		// to avoid overwriting the user data if the key already exists compare the bucketName to
		// and if they are a match user the CreateBucket method instead.
		if rootName == "USER" {
			root, err := tx.CreateBucket([]byte(rootName))
			if err != nil {
				return fmt.Errorf("user already exists")
			}
			if err := root.Put([]byte(key), *value); err != nil {
				return err
			}
		} else {
			root, err := tx.CreateBucketIfNotExists([]byte(rootName))
			if err != nil {
				return err
			}
			if err := root.Put([]byte(key), *value); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("got an error while trying to create enrty, %v", err)
	}
	return nil
}

// AddEntry takes a bucketName and key and data to be stored
// bucketname is the name GROUP of data being collected. Key is the time in the format (time.RFC3339)
func AddEntry(bucketName string, key string, value interface{}) error {
	encoded, err := Encode(&value)
	if err != nil {
		fmt.Printf("cannot encode data entered, %v ", err)
	}
	if err := nestedEntry(bucketName, key, &encoded); err != nil {
		return err
	}
	return nil
}

// GetFromDatabase takes a bucket name
func GetFromDatabase(bucketName string) (map[string]string, error) {
	rootBucket := strings.ToUpper(bucketName)

	db := Initialize()
	defer db.Close()

	err := db.View(func(tx *bolt.Tx) error {
		// check if bucket is empty, if it is then return nil otherwise the root.ForEach function panics
		root := tx.Bucket([]byte(rootBucket))
		if root == nil {
			return nil
		}
		err := root.ForEach(func(k, v []byte) error {
			value := SensorEntry{}
			if err := Decode(string(v), &value); err != nil {
				return err
			}
			holder[string(k)] = string(v)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return holder, nil
}
