package types

import (
	"github.com/only1isus/majorProj/consts"
)

// Log holds []LogEntry
type Log struct {
	Entry []LogEntry `json:"entry"`
}

// LogEntry ...
type LogEntry struct {
	Type    string `json:"type"`
	Time    int64  `json:"time"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DatabaseConnection ...
type Database struct {
	DBConnection Connection `json:"databaseConnection"`
}

type Connection struct {
	Port                  string `yaml:"port"`
	Host                  string `yaml:"host"`
	Secret                string `yaml:"secret"`
	RequireAuthentication bool   `yaml:"requireAuthentication"`
}

// type DatabaseConnection struct {
// 	Port                  string `yaml:"port"`
// 	Host                  string `yaml:"host"`
// 	Secret                string `yaml:"secret"`
// 	RequireAuthentication bool   `yaml:"requireAuthentication"`
// }

// User ...
type User struct {
	CreatedAt int64  `json:"createdAt,omitempty"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Role      string `json:"role,omitempty"`
	Password  string `json:"password,omitempty"`
	Key       string `json:"key,omitempty"`
}

// SensorEntry is the structure ofrhe sensor data
type SensorEntry struct {
	Time       int64               `json:"time"`
	SensorType consts.BucketFilter `json:"sensorType"`
	Value      float64             `json:"value"`
}

// Sensor struct holds []SensorEntry
type Sensor struct {
	Data []SensorEntry `json:"data"`
}

// OutputDeviceSetting tells how an output device is suppose to behave
type OutputDeviceSetting struct {
	Name      string // the name of the device.
	Pin       int    // GPIO pin numner the device is connected to.
	Ontime    int64  // off time in minutes
	Every     int64  // repeat every X minutes
	Automatic bool   // if set to 'true' then the ontime and every has no effect.
}

// ChangeTiming method takes the timing (onTime and Every and make the changes in the config.yaml file.)
func (o *OutputDeviceSetting) ChangeTiming(onTime, every int64) {
	o.Every = every
	o.Ontime = onTime
}
