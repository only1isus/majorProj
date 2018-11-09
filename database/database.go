package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

const (
	filename = "data/mainnet.db"
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
		root := tx.Bucket([]byte("User"))
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

// nestedSensorEntry function takes a bucketname, a subbucket name and a value
// bucketName = Root bucket as a string
// subBucketName = nested bucket name. A sensor name can be considered as a subBucketName.
// the value is the encoded data to be stored.
func nestedSensorEntry(bucketName string, value *[]byte) error {
	bucketName = strings.ToUpper(bucketName)
	rootName := "SENSOR"
	db := Initialize()
	defer db.Close()

	err := db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte(rootName))
		if err != nil {
			return err
		}

		timeID := time.Now()
		id := timeID.Format(time.RFC3339)
		// the time, formatted as time.RFC3339, is used as the key
		if err := root.Put([]byte(id), *value); err != nil {
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("got an error while trying to create nested bucket. %v", err)
	}
	return nil
}

// AddSensorData - this function takes s string for the bucket name and a struct of the data and returns an error
// bucketName is the name of the sensor data being stored
func AddSensorData(bucketName string, in interface{}) error {
	encoded, err := Encode(&in)
	if err != nil {
		fmt.Printf("cannot encode json %v ", err)
	}
	if err := nestedSensorEntry(bucketName, &encoded); err != nil {
		return err
	}
	return nil
}

// GetSensorData takes a bucketName as a string and returns a slice of sensor data
func GetSensorData(sensorType string) (map[string]string, error) {
	rootBucket := "SENSOR"

	db := Initialize()
	defer db.Close()

	err := db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(rootBucket))
		root.ForEach(func(k, v []byte) error {
			value := SensorEntry{}
			if err := Decode(string(v), &value); err != nil {
				return err
			}
			// Sensor = append(Sensor, value)
			// value := interface{}
			holder[string(k)] = string(v)
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	// fmt.Println(holder)
	return holder, nil
}
