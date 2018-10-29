package main

import (
	"fmt"
	"time"

	db "github.com/only1isus/majorProj/database"
)

func main() {
	logEntry := db.Log{
		Time:    time.Now().Unix(),
		Success: true,
		Message: "hello mate",
	}

	formatted, err := db.Encode(logEntry)
	if err != nil {
		fmt.Printf("got the following error: %v", err)
	}
	db.NewEnty("test1", formatted)

	unformatted, err := db.GetEntry("test1")
	if err != nil {
		fmt.Printf("got an errorr %v", err)
	}
	fmt.Println(unformatted)
	db.Decode(unformatted)

}
