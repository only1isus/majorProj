package db

import (
	"testing"
	"time"

	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/types"
)

var info = []types.User{
	types.User{
		Email: "testing123@gmail.com",
		Name:  "adam Doe",
		Phone: "00000000",
	},
	types.User{
		Email: "testagain@gmail.com",
		Name:  "eve Doe",
		Phone: "00000000",
	},
	types.User{
		Email: "moretesting@gmail.com",
		Name:  "John Doe",
		Phone: "00000000",
	},
	types.User{

		Email: "evenmoretesting@gmail.com",
		Name:  "dick Doe",
		Phone: "00000000",
	},
	types.User{
		Email: "evenevenmoretesting@gmail.com",
		Name:  "romaine Doe",
		Phone: "00000000",
	},
}

var sensorData = []types.SensorEntry{
	types.SensorEntry{SensorType: consts.Humidity, Time: time.Now().Unix(), Value: 66.8},
	types.SensorEntry{SensorType: consts.Temperature, Time: time.Now().Unix(), Value: 31},
	types.SensorEntry{SensorType: consts.Humidity, Time: time.Now().Unix(), Value: 54.7},
	types.SensorEntry{SensorType: consts.WaterLevel, Time: time.Now().Unix(), Value: 80.0},
	types.SensorEntry{SensorType: consts.Humidity, Time: time.Now().Unix(), Value: 34.7},
	types.SensorEntry{SensorType: consts.Temperature, Time: time.Now().Unix(), Value: 28.8},
	types.SensorEntry{SensorType: consts.Humidity, Time: time.Now().Unix(), Value: 84.7},
	types.SensorEntry{SensorType: consts.PH, Time: time.Now().Unix(), Value: 7.1},
	types.SensorEntry{SensorType: consts.PH, Time: time.Now().Unix(), Value: 7.8},
	types.SensorEntry{SensorType: consts.Humidity, Time: time.Now().Unix(), Value: 65.1},
}

var sensorTT = []struct {
	name     consts.BucketFilter
	expected int
}{
	{name: consts.Humidity, expected: 5},
	{name: consts.WaterLevel, expected: 1},
	{name: consts.Temperature, expected: 2},
	{name: consts.PH, expected: 2},
	{name: consts.All, expected: 10},
}

func TestCreateBucket(t *testing.T) {
	if err := CreateBucket("1GYJU7OD2KFJRBUWDPP5I8P5VCL"); err != nil {
		t.Fail()
	}
}
func TestWriteToDatabase(t *testing.T) {
	for _, test := range info {
		t.Run(test.Name, func(t *testing.T) {
			err := AddEntry(nil, consts.User, []byte(test.Email), test)
			if err != nil {
				t.Error("failed. User already exists")
			}

		})
	}
}

func TestGetFromDatabase(t *testing.T) {
	for _, test := range info {
		t.Run(test.Name, func(t *testing.T) {
			user, err := GetUserData(test.Email)
			if err != nil {
				t.Fail()
			}
			if user.Email != test.Email {
				t.Errorf("user %v not found", test.Email)
			}
		})
	}
}
func TestWriteSenorData(t *testing.T) {
	for _, test := range sensorData {
		time.Sleep(2 * time.Second) // just to makes sure the data isn't overwritten
		t.Run(string(test.SensorType), func(t *testing.T) {
			err := AddEntry([]byte("1GYJU7OD2KFJRBUWDPP5I8P5VCL"), consts.Sensor, []byte(time.Now().Format(time.RFC3339)), test)
			if err != nil {
				t.Fail()
			}
		})
	}
}

func TestGetSenorData(t *testing.T) {
	for _, test := range sensorTT {

		t.Run(string(test.name), func(t *testing.T) {
			data, err := GetSensorData([]byte("1GYJU7OD2KFJRBUWDPP5I8P5VCL"), test.name, 0)
			if err != nil {
				t.Fail()
			}
			if len(data) != test.expected {
				t.Errorf("expeced %v got %v instead", test.expected, len(data))
			}
		})
	}
}
func TestWriteLog(t *testing.T) {
	l := types.LogEntry{
		Message: "hello I am romaine",
		Success: false,
		Time:    time.Now().Unix(),
	}

	err := AddEntry([]byte("1GYJU7OD2KFJRBUWDPP5I8P5VCL"), consts.Log, []byte(time.Now().Format(time.RFC3339)), l)
	if err != nil {
		t.Fail()
	}
}

func TestGetLogs(t *testing.T) {
	logs, err := GetLogs([]byte("1GYJU7OD2KFJRBUWDPP5I8P5VCL"), 0)
	if err != nil {
		t.Fail()
	}
	if len(*logs) == 0 {
		t.Errorf("cannot get the logs")
	}
}
