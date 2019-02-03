package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/types"
)

const (
	filename = "data/main.db"
)

// getdbpath returns a string of the full file path
func getdbpath() string {
	var pwd string
	var err error
	pwd, err = os.Getwd()
	if !strings.Contains(pwd, "data") {
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

// Initialize is the initializer for the database
// it creates an instance of the database and retruns a struct of the database
func initialize() *bolt.DB {
	db, err := bolt.Open(getdbpath(), 0644, nil)
	if err != nil {
		log.Fatalf("go %v, location %s", err, getdbpath())
	}
	return db
}

// Encode takes a value as a sturct and returns []byte
func encode(value interface{}) (*[]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return &encoded, nil
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
func nestedEntry(rootBucket []byte, bucketName consts.BucketName, key []byte, value *[]byte) error {
	rootName := []byte(strings.ToUpper(string(bucketName)))
	db := initialize()
	defer db.Close()

	err := db.Update(func(tx *bolt.Tx) error {
		// to avoid overwriting the user data if the key already exists compare the bucketName to
		// and if they are a match user the CreateBucket method instead.
		if string(rootName) == strings.ToUpper(string(consts.User)) {
			root, err := tx.CreateBucketIfNotExists(rootName)
			if err != nil {
				return err
			}
			// because Put method doesn't tell return an error if the key exists,
			// a check has to be done manually
			v := root.Get(key)
			if v != nil {
				return fmt.Errorf("key exists %v", err)
			}
			if err = root.Put(key, *value); err != nil {
				return fmt.Errorf("key is blank or too large. %v", err)
			}

		} else {
			rb := tx.Bucket(rootBucket)
			root, err := rb.CreateBucketIfNotExists(rootName)
			if err != nil {
				return err
			}
			if err = root.Put([]byte(key), *value); err != nil {
				return fmt.Errorf("key is blank or too large. %v", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("got an error while trying to create enrty, %v", err)
	}
	return nil
}

// AddEntry takes the data to be committed along with the key associated with the data and
// adds it the the bucket specified. The key must be unique so as to not cause the cause an
// error. The key, a bucket within a bucket, can be any unique string. If the key is passed
// as time.Now().Format(time.RFC3339) then the data can be filtered when when ready to be
// retrieved.
func AddEntry(root []byte, bucketName consts.BucketName, key []byte, value interface{}) error {
	encoded, err := encode(&value)
	if err != nil {
		fmt.Printf("cannot encode data entered, %v ", err)
	}
	if err := nestedEntry(root, bucketName, key, encoded); err != nil {
		return err
	}
	return nil
}

func getFromNestedBucket(rootBucket []byte, bucketName consts.BucketName, filter consts.BucketFilter, span int64) (*[]string, error) {
	// to find the timespan, the time is taken and then the max and min times found.
	timeNow := time.Now().Unix()
	maxTime := timeNow + (span * 3600)
	minTime := timeNow - (span * 3600)
	// change the times from string to the time.Time struct.
	maxTimeUnix := time.Unix(maxTime, maxTime/100000000)
	minTimeUnix := time.Unix(minTime, minTime/100000000)
	// finally change the time from time.Time to a string
	max := maxTimeUnix.Format(time.RFC3339)
	min := minTimeUnix.Format(time.RFC3339)

	bucket := []byte(strings.ToUpper(string(bucketName)))
	out := new([]string)
	db := initialize()
	defer db.Close()

	err := db.View(func(tx *bolt.Tx) error {
		// check if bucket is empty, if it is then return nil otherwise the root.ForEach function panics
		rb := tx.Bucket(bytes.ToUpper(rootBucket))
		if rb == nil && bucketName == consts.User {
			root := tx.Bucket(bytes.ToUpper([]byte(consts.User)))
			if root == nil {
				return fmt.Errorf("bucket Empty")
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
			// return fmt.Errorf("check to make sure the Bucket Name and Root Bucket Name are correct")
		}
		root := rb.Bucket([]byte(bucket))
		if root == nil {
			return nil
		}
		// if the timespan is 0, then find and return all the entries filtered by const.BucketFilter
		if span == 0 || rootBucket == nil {
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
		}

		// otherwise find the entries within the timespan and filter by BucketFilter
		c := root.Cursor()
		for k, v := c.Seek([]byte(min)); k != nil && bytes.Compare(k, []byte(max)) <= 0; k, v = c.Next() {
			if bytes.Contains(v, []byte(filter)) {
				*out = append(*out, string(v))
			}
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
func GetSensorData(rootBucket []byte, filter consts.BucketFilter, span int64) ([]types.SensorEntry, error) {
	d, err := getFromNestedBucket(rootBucket, consts.Sensor, filter, span)
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

// GetUserData takes a key as a string and returns a User.
func GetUserData(key string) (*types.User, error) {
	users, err := getFromNestedBucket(nil, consts.User, consts.All, 0)
	if err != nil {
		return nil, err
	}
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

// GetLogs returns all the logs within the time specified with the span (number of hours) parameter.
func GetLogs(rootBucket []byte, span int64) (*[]types.LogEntry, error) {
	logs, err := getFromNestedBucket(rootBucket, consts.Log, consts.All, span)
	if err != nil {
		return nil, err
	}
	allLogs := []types.LogEntry{}
	for _, log := range *logs {
		l := types.LogEntry{}
		if err := decode([]byte(log), &l); err != nil {
			return nil, err
		}
		allLogs = append(allLogs, l)
	}
	return &allLogs, err
}

// CreateBucket takes a name and creates a bucket if none exists
func CreateBucket(bucketName string) error {
	rootName := []byte(bucketName)
	db := initialize()
	defer db.Close()

	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(rootName)
		if err != nil {
			return fmt.Errorf("the key provided already exists")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
