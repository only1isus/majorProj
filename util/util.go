package util

import (
	"fmt"
	"strings"
	"time"

	"github.com/only1isus/majorProj/database"
)

// Log holds []LogEntry
type Log struct {
	Entry []LogEntry `json:"entry"`
}

// LogEntry ...
type LogEntry struct {
	Time    int64  `json:"time"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// User ...
type User struct {
	CreatedAt int64  `json:"createdAt"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
}

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

// Add commits the struct to the database
func (l LogEntry) Add() error {
	bucketName := "Log"
	t := time.Now()
	key := t.Format(time.RFC3339)
	err := db.AddEntry(bucketName, key, l)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

// Add commits the struct to the database
func (u *User) Add() error {
	bucketName := "User"
	u.Email = strings.ToLower(u.Name)
	u.CreatedAt = time.Now().Unix()
	key := u.Email
	err := db.AddEntry(bucketName, key, &u)
	if err != nil {
		return err
	}
	return nil
}

// GetUser takes a key and return a user struct
func GetUser(key string) (map[string]interface{}, error) {
	var user map[string]interface{}
	err := db.GetNestedUser(key, &user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Add method commits the structured data to the database
func (sensor *SensorEntry) Add() error {
	bucketName := "Sensor"
	sensor.SensorType = strings.Title(sensor.SensorType)
	t := time.Now()
	key := t.Format(time.RFC3339)
	err := db.AddEntry(bucketName, key, &sensor)
	if err != nil {
		return err
	}
	return nil
}

// GetSensorDataByType takes a bucketName and returns an array of the sensor data
func GetSensorDataByType(sensorType string) ([]SensorEntry, error) {
	bucketName := "Sensor"
	data := SensorEntry{}
	sensorSlice := []SensorEntry{}
	sensorType = strings.Title(sensorType)
	values, err := db.GetSensorData(bucketName)
	if err != nil {
		return nil, err
	}
	for _, val := range values {
		if err := db.Decode(val, &data); err != nil {
			return nil, err
		}
		if sensorType == data.SensorType {
			sensorSlice = append(sensorSlice, data)
		}
		if sensorType == "All" {
			sensorSlice = append(sensorSlice, data)
		}
	}
	return sensorSlice, nil
}
