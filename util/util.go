package util

// Log defines the structure of the log entry s
type Log struct {
	Time    int64
	Success bool
	Message string
}

// User - this is the structure of the users info to be stored in the database
type User struct {
	name  string
	email string
	phone int64
	role  string
}

// PlaceHolder stores the structured data to be used after Decode function call
var PlaceHolder []byte
