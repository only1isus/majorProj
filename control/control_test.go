package control

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/only1isus/majorProj/consts"
// )

// func TestGetConfig(t *testing.T) {

// 	deviceName := consts.CoolingFan
// 	a := Devices{}
// 	fan, err := a.Get("coolingFan")
// 	if err != nil {
// 		fmt.Printf("error %v", err)
// 		t.Fail()
// 	}
// 	if fan.Name != deviceName {
// 		fmt.Println(fan)
// 		t.Fail()
// 	}

// }

// // func TestControl(t *testing.T) {
// // 	control.NewOutputDevice()
// // 	_, err = initializePins(fan.Pins, rpio.Pwm, rpio.Output)
// // 	if err != nil {
// // 		// this is going to return an error on any device other than a pi.
// // 		t.Log("fail to initialize the pins ", err)
// // 		// t.Fail()
// // 	}
// // }
