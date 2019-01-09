package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/types"
)

const (
	filename = "data/main.db"
)

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
func initialize() *bolt.DB {
	db, err := bolt.Open(getdbpath(), 0644, nil)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Encode takes a value as a sturct and returns []byte
func encode(value interface{}) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

// Decode takes the value and converts the data to a struct
func decode(data []byte, holder interface{}) error {
	err := json.Unmarshal([]byte(data), &holder)
	if err != nil {
		return fmt.Errorf("cannot unmarshal %v ", err)
	}
	return nil
}

// nestedEntry function takes a bucketname
// bucketName is rootbucket as a string
func nestedEntry(bucketName consts.BucketName, key string, value *[]byte) error {
	rootName := strings.ToUpper(string(bucketName))
	db := initialize()
	defer db.Close()

	err := db.Update(func(tx *bolt.Tx) error {
		// to avoid overwriting the user data if the key already exists compare the bucketName to
		// and if they are a match user the CreateBucket method instead.
		if rootName == strings.ToUpper(string(consts.User)) {
			root, err := tx.CreateBucketIfNotExists([]byte(rootName))
			if err != nil {
				return err
			}
			// because Put method doesn't tell return an error if the key exists,
			// a check has to be done manually
			v := root.Get([]byte(key))
			if v != nil {
				return fmt.Errorf("key exists %v", err)
			}
			if err = root.Put([]byte(key), *value); err != nil {
				return fmt.Errorf("key is blank or too large. %v", err)
			}

		} else {
			root, err := tx.CreateBucketIfNotExists([]byte(rootName))
			if err != nil {
				return err
			}
			if err = root.Put([]byte(key), *value); err != nil {
				return fmt.Errorf("bucket already exists")
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
// Key is the time in the format (time.RFC3339)
func AddEntry(bucketName consts.BucketName, key string, value interface{}) error {
	encoded, err := encode(&value)
	if err != nil {
		fmt.Printf("cannot encode data entered, %v ", err)
	}
	if err := nestedEntry(bucketName, key, &encoded); err != nil {
		return err
	}
	return nil
}

func getFromNestedBucket(bucketName consts.BucketName, filter consts.BucketFilter) (*[]string, error) {
	rootBucket := strings.ToUpper(string(bucketName))
	out := new([]string)
	db := initialize()
	defer db.Close()

	err := db.View(func(tx *bolt.Tx) error {
		// check if bucket is empty, if it is then return nil otherwise the root.ForEach function panics
		root := tx.Bucket([]byte(rootBucket))
		if root == nil {
			return nil
		}
		err := root.ForEach(func(k, v []byte) error {
			if bytes.Contains(v, []byte(filter)) {
				*out = append(*out, string(v))
			}
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
	return out, nil
}

// GetSensorData returns a list of the sensor data
// bucketName is the name of the bucket the data should be added to. To choose which type
// of sensor data is returned set a filter.
func GetSensorData(bucketName consts.BucketName, filter consts.BucketFilter) ([]types.SensorEntry, error) {
	d, err := getFromNestedBucket(bucketName, filter)
	if err != nil {
		return nil, err
	}
	o := types.SensorEntry{}
	out := []types.SensorEntry{}
	for _, v := range *d {
		if err := decode([]byte(v), &o); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}

// GetUserData takes a subbucket name as a string and an User interface
func GetUserData(key string) (*types.User, error) {
	users, err := getFromNestedBucket(consts.User, consts.All)
	if err != nil {
		return nil, err
	}
	fmt.Println(users)
	u := types.User{}
	for _, user := range *users {
		u := types.User{}
		if err := decode([]byte(user), &u); err != nil {
			return nil, err
		}
		if u.Email == key {
			return &u, nil
		}

	}
	return &u, err
}
