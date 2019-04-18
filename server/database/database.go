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

func maxMinTime(start, end int64) (maxTimeUnix string, minTimeUnix string) {
	// timeNow := time.Now().Unix()
	// maxTime := timeNow + (span * 3600)
	// minTime := timeNow - (span * 3600)
	// change the times from string to the time.Time struct.
	maxTimeUnix = time.Unix(end, end/100000000).Format(time.RFC3339)
	minTimeUnix = time.Unix(start, start/100000000).Format(time.RFC3339)
	fmt.Println(maxTimeUnix)
	return maxTimeUnix, minTimeUnix
}

// AddEntry takes the data to be committed along with the key associated with the data and
// adds it the the bucket specified. The key must be unique so as to not cause the cause an
// error. The key, a bucket within a bucket, can be any unique string. If the key is passed
// as time.Now().Format(time.RFC3339) then the data can be filtered when when ready to be
// retrieved.
func AddSensorEntry(rootBucket []byte, key []byte, value types.SensorEntry) error {
	db := initialize()
	defer db.Close()

	out, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists(bytes.ToUpper(rootBucket))
		if err != nil {
			return fmt.Errorf("the root bucket name is too long or is empty")
		}
		r, err := root.CreateBucketIfNotExists(bytes.ToUpper([]byte(consts.Sensor)))
		if err != nil {
			return fmt.Errorf("the bucket name is too long or is empty")
		}
		if err := r.Put(key, out); err != nil {
			return fmt.Errorf("the key is too long")
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// GetSensorData returns a list of the sensor data
// bucketName is the name of the bucket the data should be added to. To choose which type
// of sensor data is returned set a filter.
func GetSensorData(rootBucket []byte, filter consts.BucketFilter, start int64, end int64) (*[]types.SensorEntry, error) {
	db := initialize()
	defer db.Close()

	maxTimeUnix, minTimeUnix := maxMinTime(start, end)
	sensorDataEntries := []types.SensorEntry{}

	if err := db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket(bytes.ToUpper(rootBucket))
		if root == nil {
			return fmt.Errorf("the root bucket is empty")
		}

		sensorEntries := root.Bucket(bytes.ToUpper([]byte(consts.Sensor)))
		if sensorEntries == nil {
			return fmt.Errorf("no entries found")
		}

		sensorData := types.SensorEntry{}
		if err := sensorEntries.ForEach(func(k, v []byte) error {
			if err := json.Unmarshal(v, &sensorData); err != nil {
				return err
			}
			RFC3339Time := time.Unix(sensorData.Time, sensorData.Time/100000000).Format(time.RFC3339)
			if bytes.Contains(v, []byte(filter)) {
				if RFC3339Time >= minTimeUnix && RFC3339Time <= maxTimeUnix {
					sensorDataEntries = append(sensorDataEntries, sensorData)
				}
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return &sensorDataEntries, nil
}

func AddUserEntry(user types.User) error {
	db := initialize()
	defer db.Close()
	user.CreatedAt = time.Now().Unix()
	out, err := json.Marshal(user)
	if err != nil {
		return err
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		userBucket, err := tx.CreateBucketIfNotExists(bytes.ToUpper([]byte(consts.User)))
		if err != nil {
			return err
		}

		v := userBucket.Get([]byte(user.Email))
		if v != nil {
			return fmt.Errorf("key exists")
		}
		if err = userBucket.Put([]byte(user.Email), out); err != nil {
			return fmt.Errorf("key is blank or too large")
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// GetUserData takes a key as a string and returns a User.
func GetUserData(key string) (*types.User, error) {
	db := initialize()
	defer db.Close()

	user := types.User{}
	err := db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket(bytes.ToUpper([]byte(consts.User)))
		if root == nil {
			return fmt.Errorf("the root bucket is empty")
		}
		u := root.Get(bytes.ToLower([]byte(key)))
		if u == nil {
			return fmt.Errorf("the key does not exist")
		}
		if err := json.Unmarshal(u, &user); err != nil {
			return fmt.Errorf("couldn't parse the iser info")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetFarmDetails(rootBucket []byte) (*types.FarmDetails, error) {
	db := initialize()
	defer db.Close()
	fd := types.FarmDetails{}
	err := db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket(bytes.ToUpper(rootBucket))
		if root == nil {
			return fmt.Errorf("The bucket doesn't exist")
		}
		farmDetailsBucket := root.Bucket(bytes.ToUpper([]byte(consts.FarmDetails)))
		if farmDetailsBucket == nil {
			return fmt.Errorf("The bucket doesn't exist")
		}
		result := farmDetailsBucket.Get(rootBucket)
		if err := json.Unmarshal(result, &fd); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &fd, nil
}

func AddFarmEntry(rootBucket, key, value []byte) error {
	db := initialize()
	defer db.Close()
	if key == nil {
		return fmt.Errorf("The key cannot be empty")
	}

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bytes.ToUpper(rootBucket))
		if err != nil {
			return fmt.Errorf("Bucket already exists")
		}
		b, err := bucket.CreateBucketIfNotExists(bytes.ToUpper([]byte(consts.FarmDetails)))
		if err != nil {
			return fmt.Errorf("Bucket already exists")
		}
		if err := b.Put(bytes.ToUpper(key), value); err != nil {
			return fmt.Errorf("The key used is too long")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func AddLogEntry(rootBucket []byte, key []byte, value types.LogEntry) error {
	db := initialize()
	defer db.Close()

	out, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists(bytes.ToUpper(rootBucket))
		if err != nil {
			return err
		}
		r, err := root.CreateBucketIfNotExists(bytes.ToUpper([]byte(consts.Log)))
		if err != nil {
			return err
		}
		if err := r.Put(key, out); err != nil {
			return fmt.Errorf("the key is too long or is empty")
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// GetLogs returns all the logs within the time specified with the span (number of hours) parameter.
func GetLogs(rootBucket []byte, start int64, end int64) (*[]types.LogEntry, error) {
	db := initialize()
	defer db.Close()

	maxTimeUnix, minTimeUnix := maxMinTime(start, end)
	logs := []types.LogEntry{}
	if err := db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket(bytes.ToUpper(rootBucket))
		if root == nil {
			return fmt.Errorf("the root bucket is empty")
		}
		logEntries := root.Bucket(bytes.ToUpper([]byte(consts.Log)))
		if logEntries == nil {
			return fmt.Errorf("there is no entry in the root bucket")
		}
		log := types.LogEntry{}

		if err := logEntries.ForEach(func(k, v []byte) error {
			if err := json.Unmarshal(v, &log); err != nil {
				return err
			}
			RFC3339Time := time.Unix(log.Time, log.Time/100000000).Format(time.RFC3339)
			if RFC3339Time >= minTimeUnix && RFC3339Time <= maxTimeUnix {
				logs = append(logs, log)
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &logs, nil
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
