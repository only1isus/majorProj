package control

import (
	"fmt"
	"testing"

	"github.com/stianeikeland/go-rpio"
	// "github.com/only1isus/majorProj/control"
)

func TestGetConfig(t *testing.T) {
	deviceName := "light"
	a := Devices{}
	fan, err := a.Get(deviceName)
	if err != nil {
		fmt.Printf("error %v", err)
		t.Fail()
	}
	if fan.Name != deviceName {
		t.Fail()
	}

}
func TestControl(t *testing.T) {
	d := Devices{}
	var err error
	fan, err := d.Get("fan")
	if err != nil {
		t.Log("fail to get config file")
		t.Fail()
	}
	_, err = initializePins(fan.Pins, rpio.Pwm, rpio.Output)
	if err != nil {
		// this is going to return an error on any device other than a pi.
		t.Log("fail to initialize the pins ", err)
		// t.Fail()
	}
}
