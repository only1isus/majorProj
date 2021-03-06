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
	"github.com/only1isus/majorProj/control"
	"github.com/only1isus/majorProj/rpc"
	"github.com/only1isus/majorProj/types"
)

func welcome() {
	version := "2.1.0"
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
	entry := make(chan *types.LogEntry, 1)
	log.Println("System running")

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// fan, err := control.NewOutputDevice(consts.CoolingFan)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// gl, err := control.NewGrowLight()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// go gl.TurnOnThenWait()

	wl, err := control.NewWaterLevelSensor(0x48, 1)
	if err != nil {
		fmt.Printf("got an error creating the water level sensor %v", err)
	}

	// // val, err := wl.Get()
	// // fmt.Println(val)
	// go wl.CheckAndNotify(3.6, entry)

	temperature := control.NewTemperatureSensor()
	// if err := temperature.Maintain(30, fan, notification); err != nil {
	// 	fmt.Println(err)
	// }

	go func(temp *control.TemperatureSensor, waterLevel *control.WaterLevelSensor) {
		for {
			timer := time.NewTimer(time.Minute * 5)
			defer timer.Stop()
			// wait for the timer to reach its limit
			<-timer.C
			wlValue, err := wl.Get()
			if err != nil {
				fmt.Printf("%v\n", err)
				fmt.Println("BEWARE OF THIS ERROR")
			}
			wlEntry := types.SensorEntry{
				SensorType: consts.WaterLevel,
				Time:       time.Now().Unix(),
				Value:      wlValue,
			}
			e, err := json.Marshal(wlEntry)
			if err != nil {
				fmt.Printf("%v\n", err)
			}
			fmt.Println(wlEntry)
			if err := rpc.CommitSensorData(&e); err != nil {
				log.Println("cannot send data to the database server", err)
			}

			t, err := temp.Prepare()
			if err != nil {
				log.Println("got and error from prepare method", err)
			}

			if err := rpc.CommitSensorData(t); err != nil {
				log.Println("cannot send data to the database server", err)
			}

		}
	}(&temperature, wl)

	go func() {
		for {
			select {
			case en := <-entry:
				fmt.Printf("%v\n", *en)
				e, err := json.Marshal(*en)
				if err != nil {
					fmt.Println(err)
				}
				if err := rpc.CommitLog(&e); err != nil {
					fmt.Println(err)
				}
				fmt.Println("sent the notification")
			}
		}
	}()
	go func() {
		<-c
		kill <- true
	}()

	<-kill
	log.Println("cleaning up")
	// gl.Off()
	wl.Close()
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
