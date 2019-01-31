package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/only1isus/majorProj/consts"

	"github.com/only1isus/majorProj/control"
	"github.com/only1isus/majorProj/rpc"

	"google.golang.org/grpc"
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

	address := flag.String("host", "", "database host address")
	flag.Parse()
	if flag.NFlag() == 0 {
		log.Println("Please enter the host address and try again")
		os.Exit(1)
	}

	grpcconn, err := grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		log.Fatalf("got an error making the connection %v", err)
	}
	defer grpcconn.Close()

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

			if err := rpc.CommitSensorData(grpcconn, t); err != nil {
				log.Println("cannot send data to the database server", err)
			}

		}
	}(&temperature, &Humidity)

	if err != nil {
		log.Println("got an error creating the fan output device ", err)
	}
	err = temperature.Maintain(29.5, fan, notification)

	// cleaning up. Closing all pins and turning off all devices whenever Ctrl ^ C is recieved.
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			select {
			case n, ok := <-notification:
				if !ok || n == nil {
					log.Println("something went wrong getting the nofifications")
				}
				err := rpc.CommitLog(grpcconn, &n)
				if err != nil {
					log.Println("got an errir from the commit log ", err)
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
}
