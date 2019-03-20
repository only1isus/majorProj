package control

import (
	"fmt"
	"time"

	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/types"
)

type WaterLevelSensor ADCSensor

func NewWaterLevelSensor(address, bus int) (WaterLevelSensor, error) {
	wlSensor, err := NewAnalogSensor(consts.WaterLevelSensor)
	if err != nil {
		return WaterLevelSensor{}, err
	}
	conn, err := i2cConnection(address, bus)
	if err != nil {
		return WaterLevelSensor{}, err
	}
	wlSensor.connection = conn

	return WaterLevelSensor(*wlSensor), nil
}

func (wl WaterLevelSensor) Get() (float64, error) {
	if err := wl.connection.Start(); err != nil {
		return 0, fmt.Errorf("cannot start the i2c connection")
	}
	value, err := wl.connection.ReadWithDefaults(wl.AnalogPin)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// CheckAndNotify takes the level of water as an int and a channel to send responses to.
// if the level of the water in the container is less than the amount specified (0 - 100%)
// then a message is sent over the channel
func (wl WaterLevelSensor) CheckAndNotify(level int, entry chan *types.LogEntry) {
	for {
		timer := time.NewTimer(time.Minute * time.Duration(wl.Every))
		defer timer.Stop()
		<-timer.C
		currentLevel, err := wl.Get()
		if err != nil {
			entry <- &types.LogEntry{
				Message: fmt.Sprintf("Something went wrong reading the water level %v", err),
				Success: false,
				Time:    time.Now().Unix(),
				Type:    string(consts.WaterLevel),
			}
		}
		fmt.Printf("current %v set %v", currentLevel, level)
		if currentLevel < float64(level) {
			entry <- &types.LogEntry{
				Message: fmt.Sprintf("The current water level is %v. Please consider refilling.", currentLevel),
				Success: true,
				Time:    time.Now().Unix(),
				Type:    string(consts.WaterLevel),
			}
		}
	}
}

func (wl *WaterLevelSensor) Close() error {
	if err := wl.connection.Start(); err != nil {
		return fmt.Errorf("cannot start the i2c connection")
	}
	wl.connection.Close()
	return nil
}
