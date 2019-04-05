package control

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/types"
	"github.com/yryz/ds18b20"
)

// TemperatureSensor is a type of the sensor struct
type TemperatureSensor Sensor

// NewTemperatureSensor return a TemperatureSensor struct
func NewTemperatureSensor() TemperatureSensor {
	return TemperatureSensor{}
}

// Get method when called returns the current temperature.
func (t *TemperatureSensor) Get() (*float64, error) {
	sensors, err := ds18b20.Sensors()
	if err != nil {
		panic(err)
	}

	for _, sensor := range sensors {
		var err error
		t.value, err = ds18b20.Temperature(sensor)
		if err != nil {
			return nil, err
		}
		break
	}
	return &t.value, nil
}

// Maintain method tries to keep the temperature at the value passed to the method.
func (t *TemperatureSensor) Maintain(value float64, f *OutputDevice, notify chan<- []byte) error {
	go func(n chan<- []byte, fan *OutputDevice) error {

		for {
			temp, err := t.Get()
			if err != nil {
				n <- nil
				return err
			}
			val := *temp
			if val >= value {
				log.Printf("Turning on fan. Current temperature is %f, limit set to %f", val, value)
				if err := fan.On(); err != nil {
					n <- nil
					return err
				}
				onTime := time.Now()
				for {
					currentTemp, err := t.Get()
					if err != nil {
						n <- nil
						return err
					}
					if value >= *currentTemp {
						time.Sleep(1 * time.Minute)
						log.Println("Done")
						fan.ChangePWM(0.4)
						time.Sleep(1 * time.Minute)
						if err := fan.Off(); err != nil {
							n <- nil
							return err
						}
						msg := types.LogEntry{
							Message: fmt.Sprintf("Fan was turned on for %v minute(s). Upper limit set to %vc", int64(time.Now().Sub(onTime).Minutes()), value),
							Success: true,
							Time:    time.Now().Unix(),
							Type:    "control",
						}
						out, err := json.Marshal(msg)
						if err != nil {
							n <- nil
							return err
						}
						n <- out
						break
					}
				}
			}
			time.Sleep(30 * time.Second)
		}
	}(notify, f)
	return nil
}

// Prepare gets the entry ready to be committed to the database
func (t *TemperatureSensor) Prepare() (*[]byte, error) {
	temp, err := t.Get()
	if err != nil {
		return nil, err
	}
	entry := types.SensorEntry{
		Time:       time.Now().Unix(),
		SensorType: consts.Temperature,
		Value:      *temp,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
