package control

import (
	"fmt"
	"log"
	"time"

	"github.com/only1isus/majorProj/consts"
)

// GrowLight struct
type GrowLight OutputDevice

// NewGrowLight return a struct based off the parameters set in the config file
func NewGrowLight() (GrowLight, error) {
	gl, err := NewOutputDevice(consts.GrowLight)
	if err != nil {
		return GrowLight{}, err
	}
	return GrowLight(*gl), err
}

// WaitThenTurnOn waits for the amount of time set in the config file "every" to pass then the
// light is turned on. The light stays on according the at amount of time set "onTime"
func (gl GrowLight) WaitThenTurnOn() {
	growLight := OutputDevice(gl)
	for {
		timer := time.NewTimer(time.Minute * time.Duration(growLight.Every))
		defer timer.Stop()
		// wait for the timer to reach its limit
		<-timer.C
		if err := growLight.On(); err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Minute * time.Duration(growLight.OnTime))
		growLight.Off()
	}
}

// TurnOnThenWait turns the light on for the anount of time set in the config file "onTime"
// The light is then turned off for the amount of time set in the config file "every"
func (gl GrowLight) TurnOnThenWait() {
	fmt.Println("called the light function")
	growLight := OutputDevice(gl)
	for {
		growLight.On()
		time.Sleep(time.Minute * time.Duration(growLight.OnTime))
		if err := gl.Off(); err != nil {
			log.Println(err)
		}
		timer := time.NewTimer(time.Minute * time.Duration(growLight.Every))
		defer timer.Stop()
		// wait for the timer to reach its limit
		<-timer.C
	}
}

// Off turns the growlight off
func (gl GrowLight) Off() error {
	growLight := OutputDevice(gl)
	if err := growLight.Off(); err != nil {
		return err
	}
	return nil
}
