package db

import (
	"testing"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/types"
)

var info = []types.User{
	types.User{
		Email:    "isuspisus1@gmail.com",
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

const (
	bucket = "1GYJU7OD2KFJRBUWDPP5I8P5VCL"
)

func convertDate(date string) int64 {
	// igmore error as the time will be provided as a int64 value
	// testing purpose
	t, _ := time.Parse(time.RFC3339, date)
	return t.Unix()
}

func TestWriteUserToDatabase(t *testing.T) {
	for _, test := range info {
		t.Run(test.Name, func(t *testing.T) {
			err := AddUserEntry(test)
			if err != nil {
				t.Log(err)
			}
		})
	}
}
func TestGetUserFromDatabase(t *testing.T) {
	for _, test := range info {
		t.Run(test.Name, func(t *testing.T) {
			user, err := GetUserData(test.Email)
			if err != nil {
				t.Log(err)
				t.FailNow()
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
		t.Run(string(test.SensorType), func(t *testing.T) {
			err := AddSensorEntry([]byte(bucket), []byte(ksuid.New().String()), test)
			if err != nil {
				t.Fatalf(err.Error())
			}
		})
	}
}

func TestGetSenorData(t *testing.T) {
	for _, test := range sensorTT {
		t.Run(string(test.name), func(t *testing.T) {
			data, err := GetSensorData([]byte(bucket), test.name, convertDate("2019-04-18T00:00:00-05:00"), convertDate("2019-04-18T06:00:00-05:00"))
			if err != nil {
				t.Fatalf("got an error trying to get data %v", err.Error())
			}
			t.Logf("length of response %d ", len(*data))
		})
	}
}

func TestWriteLog(t *testing.T) {
	l := types.LogEntry{
		Message: "hello I am romaine",
		Success: false,
		Time:    time.Now().Unix(),
	}

	err := AddLogEntry([]byte(bucket), []byte(ksuid.New().String()), l)
	if err != nil {
		t.Fail()
	}
}

func TestGetLogs(t *testing.T) {
	logs, err := GetLogs([]byte("1GYJU7OD2KFJRBUWDPP5I8P5VCL"), convertDate("2019-04-18T00:00:00-05:00"), convertDate("2019-04-18T06:00:00-05:00"))
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
		PlantedOn:    convertDate("2019-03-03T00:00:00+00:00"),
		HarvestOn:    convertDate("2019-04-03T00:00:00+00:00"),
		NPK:          "generic",
		CropType:     "spinach",
		MaturityTime: 30,
	}
	if err := AddFarmEntry([]byte(bucket), []byte(bucket), *fd); err != nil {
		t.Fatalf("got an error adding farm details to the database, %v", err)
	}
}

func TestGetFarmDetails(t *testing.T) {
	fd, err := GetFarmDetails([]byte(bucket))
	if err != nil {
		t.Fatalf("got an error adding farm details to the database, %v", err)
	}
	if (*fd).CropType != "spinach" {
		t.Fatal("failed, the information is not the same.", fd)
	}
}

func TestWriteSummary(t *testing.T) {
	_, err := GetSummaries()
	if err != nil {
		t.Fatalf("got an error adding farm details to the database, %v", err)
	}
}
