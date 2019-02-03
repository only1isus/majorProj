package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/types"

	"github.com/only1isus/majorProj/control"
	"github.com/only1isus/majorProj/rpc"
)

func welcome() {
	version := "1.1.0"
	fmt.Printf(`										
                            |/
                        . - |~ .
                     .*          ', 
                    ,   #          '
                    '              *		
                    '              '  
                     '-,___.'__,.-'
     ==============================================
                   AMBOROSA - alpha
                          %v
     ==============================================
					
					`, version)
	fmt.Println("")
}

func main() {
	c := make(chan os.Signal, 1)
	notification := make(chan []byte, 1)
	kill := make(chan bool, 1)
	onTime := time.Now()
	log.Println("System running")

	temperature := control.Temperature{}
	Humidity := control.Humidity{}
	fan, err := control.NewOutputDevice(consts.CoolingFan)

	go func(temp *control.Temperature, humid *control.Humidity) {
		for {
			timer := time.NewTimer(time.Minute * 5)
			defer timer.Stop()
			// wait for the timer to reach its limit
			<-timer.C

			t, err := temp.Prepare()
			if err != nil {
				log.Println("got and error from prepare method", err)
			}

			if err := rpc.CommitSensorData(t); err != nil {
				log.Println("cannot send data to the database server", err)
			}
		}
	}(&temperature, &Humidity)

	if err != nil {
		log.Println("got an error creating the fan output device ", err)
	}
	err = temperature.Maintain(30, fan, notification)

	// cleaning up. Closing all pins and turning off all devices whenever Ctrl ^ C is recieved.
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			select {
			case n, ok := <-notification:
				if !ok || n == nil {
					log.Println("something went wrong getting the nofifications")
				}
				err := rpc.CommitLog(&n)
				if err != nil {
					log.Println("got an error from the commit log ", err)
				}
			}
		}
	}()
	go func() {
		<-c
		kill <- true
	}()

	<-kill
	log.Println("cleaning up")
	err = fan.Off()
	if err != nil {
		fmt.Println("got an error from")
	}
	msg := types.LogEntry{
		Message: fmt.Sprintf("System terminated from the command line at %v on %v. On time %v minutes.", time.Now().Format("15:04:05"), time.Now().Format("2006-01-02"), int64(time.Now().Sub(onTime).Minutes())),
		Success: true,
		Time:    time.Now().Unix(),
		Type:    "termination",
	}
	out, err := json.Marshal(msg)
	if err != nil {
		notification <- nil
	}
	err = rpc.CommitLog(&out)
	if err != nil {
		log.Println("got an error from the commit log ", err)
	}
	log.Println("Shutting down")
}
