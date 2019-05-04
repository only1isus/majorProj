package control

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/only1isus/ADS1115"

	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/rpc"
	"github.com/only1isus/majorProj/types"
)

type WaterLevelSensor ADCSensor

func NewWaterLevelSensor(address, bus int) (*WaterLevelSensor, error) {
	wlSensor, err := NewAnalogSensor(consts.WaterLevelSensor)
	if err != nil {
		return nil, err
	}
	ads, err := ADS1115.NewConnection(address, bus)
	if err != nil {
		return nil, err
	}
	wlSensor.connection = ADS1115.NewADS1115Device(ads)
	wl := WaterLevelSensor(*wlSensor)
	return &wl, nil
}

func (wl *WaterLevelSensor) Get() (*float64, error) {
	value, err := wl.connection.Read(wl.AnalogPin)
	if err != nil {
		return nil, err
	}
	out := new(float64)
	*out = ToFixed(float64(value), 1)
	return out, nil
}

// CheckAndNotify takes the level of water as an int and a channel to send responses to.
// if the level of the water in the container is less than the amount specified (0 - 100%)
// then a message is sent over the channel
func (wl *WaterLevelSensor) CheckAndNotify(level float64, entry chan *types.LogEntry) {
	go func(ent chan *types.LogEntry) {
		for {
			currentLevel, err := wl.Get()
			if err != nil {
				fmt.Printf("%v", err)

			}
			timer := time.NewTimer(time.Minute * time.Duration(wl.Every))
			defer timer.Stop()
			<-timer.C

			if err != nil {
				entry <- &types.LogEntry{
					Message: fmt.Sprintf("Something went wrong reading the water level %v", err),
					Success: false,
					Time:    time.Now().Unix(),
					Type:    string(consts.WaterLevel),
				}
			}
			if *currentLevel < level {
				entry <- &types.LogEntry{
					Message: fmt.Sprintf("The current water level is %v. Please consider refilling.", *currentLevel),
					Success: true,
					Time:    time.Now().Unix(),
					Type:    string(consts.WaterLevel),
				}
			}
		}
	}(entry)
}

func (wl WaterLevelSensor) ReadAndNotify() error {

	voltageValue, err := wl.Get()
	if err != nil {
		return err
	}
	data := &types.SensorEntry{
		SensorType: consts.WaterLevel,
		Time:       time.Now().Unix(),
		Value:      *voltageValue,
	}
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err := rpc.CommitSensorData(&out); err != nil {
		return err
	}
	return nil
}

func (wl *WaterLevelSensor) Close() error {
	wl.connection.Close()
	return nil
}
