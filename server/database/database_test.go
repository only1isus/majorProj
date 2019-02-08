package db

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/types"
)

var info = []types.User{
	types.User{
		Email:    "testing1235@gmail.com",
		Name:     "adam Doe",
		Phone:    "00000000",
		Password: "qwerty",
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
	name consts.BucketFilter
}{
	{name: consts.Humidity},
	{name: consts.WaterLevel},
	{name: consts.Temperature},
	{name: consts.PH},
	{name: consts.All},
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
			t.Log(*user)
		})
	}
}
func TestWriteSenorData(t *testing.T) {
	for _, test := range sensorData {
		time.Sleep(2 * time.Second) // just to makes sure the data isn't overwritten
		t.Run(string(test.SensorType), func(t *testing.T) {
			err := AddSensorEntry([]byte("1GYJU7OD2KFJRBUWDPP5I8P5VCL"), []byte(time.Now().Format(time.RFC3339)), test)
			if err != nil {
				t.Fatalf(err.Error())
			}
		})
	}
}

func TestGetSenorData(t *testing.T) {
	for _, test := range sensorTT {
		// time.Sleep(time.Second * 2)
		t.Run(string(test.name), func(t *testing.T) {
			data, err := GetSensorData([]byte("1GYJU7OD2KFJRBUWDPP5I8P5VCL"), test.name, 0)
			if err != nil {
				t.Fatalf("got an error trying to get data %v", err.Error())
			}
			// if (*data)[0].SensorType != test.name {
			// 	t.Errorf("expeced %v got %v instead", test.name, (*data)[0].SensorType)
			// }
			t.Log(len(*data))
		})
	}
}

func TestWriteLog(t *testing.T) {
	l := types.LogEntry{
		Message: "hello I am romaine",
		Success: false,
		Time:    time.Now().Unix(),
	}

	err := AddLogEntry([]byte("1GYJU7OD2KFJRBUWDPP5I8P5VCL"), []byte(time.Now().Format(time.RFC3339)), l)
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
	t.Log(len(*logs))
}

func TestFarmDetails(t *testing.T) {
	fd := &types.FarmDetails{
		Configured:   true,
		CropType:     "lettuce",
		MaturityTime: 23,
	}
	out, err := json.Marshal(fd)
	if err != nil {
		t.Fatalf("cannot format the data")
	}
	if err := AddFarmEntry([]byte("1GYJU7OD2KFJRBUWDPP5I8P5VCL"), []byte("1GYJU7OD2KFJRBUWDPP5I8P5VCL"), out); err != nil {
		t.Fatalf("got an error adding farm details to the database, %v", err)
	}
}

func TestGetFarmDetails(t *testing.T) {
	fd, err := GetFarmDetails([]byte("1GYJU7OD2KFJRBUWDPP5I8P5VCL"))
	if err != nil {
		t.Fatalf("got an error adding farm details to the database, %v", err)
	}
	if (*fd).CropType != "lettuce" {
		t.Fatal("failed, the information is not the same.", fd)
	}
}
