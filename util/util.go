package util

import (
	"fmt"

	"github.com/only1isus/majorProj/database"
)

// Log holds []LogEntry
type Log struct {
	Entry []LogEntry `json:"entry"`
}

type LogEntry struct {
	Time    int64  `json:"time"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// User is the user
type User struct {
	CreatedAt int64  `json:"createdAt"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
}

// SensorEntry is the structure ofrhe sensor data
type SensorEntry struct {
	Time       int64  `json:"time"`
	SensorType string `json:"sensorType"`
	Value      string `json:"value"`
}

// Sensor struct holds []SensorEntry
type Sensor struct {
	Data []SensorEntry `json:"data"`
}

// AddEntry takes a bucketname, key and a value
func (l Log) AddEntry(bucketName string, key string, value LogEntry) error {
	valueFromBucket := db.GetFromBucket(bucketName, key)

	holder := []LogEntry{}
	err := db.Decode(valueFromBucket, &holder)
	if err != nil {
		return err
	}
	l.Entry = append(l.Entry, value)
	err = db.CreateBucketEntry(bucketName, key, l.Entry)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

// CreateUser by passing in the user struct
func CreateUser(u *User) error {
	if err := db.NewNestedUser(u.Email, &u); err != nil {
		return fmt.Errorf("can't do this shit %v", err)
	}

	return nil
}

// GetUser takes a key and return a user struct
func GetUser(key string) (map[string]interface{}, error) {
	// user := interface{}
	var user map[string]interface{}
	err := db.GetNestedUser(key, &user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
