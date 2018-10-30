package util

// Log
type Log struct {
	Time    int64
	Success bool
	Message string
}

// User
type User struct {
	name  string
	email string
	phone int64
	role  string
}

// Sensordata
type Sensordata struct {
	time       int64
	sensortype string
	value      string
}

// PlaceHolder stores the structured data to be used after Decode function call
var PlaceHolder []byte
