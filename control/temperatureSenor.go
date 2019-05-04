package control

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/only1isus/majorProj/rpc"

	i2c "github.com/d2r2/go-i2c"
	sht3x "github.com/d2r2/go-sht3x"
	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/types"
)

// TemperatureSensor is a type of the sensor struct
type TemperatureSensor I2CSensor

// NewTemperatureSensor return a TemperatureSensor struct
func NewTemperatureSensor() (*TemperatureSensor, error) {
	temperatureSensor, err := NewI2CSensor(consts.Sth3xHumidity)
	if err != nil {
		return nil, err
	}
	ts := TemperatureSensor(*temperatureSensor)
	return &ts, nil
}

// Get method when called returns the current temperature.
func (t *TemperatureSensor) Get() (*float64, error) {
	fmt.Println("reading temperature")
	i2cconn, err := i2c.NewI2C(t.Address, t.Bus)
	if err != nil {
		log.Fatal(err)
	}
	defer i2cconn.Close()

	sensor := sht3x.NewSHT3X()
	temperature, _, err := sensor.ReadTemperatureAndRelativeHumidity(i2cconn, sht3x.RepeatabilityLow)
	if err != nil {
		return nil, err
	}
	tmp := new(float64)
	*tmp = ToFixed(float64(temperature), 1)
	return tmp, nil
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
func (t *TemperatureSensor) ReadAndCommit() error {
	for {
		timer := time.NewTimer(time.Minute * time.Duration(t.Every))
		defer timer.Stop()
		// wait for the timer to reach its limit
		<-timer.C

		temp, err := t.Get()
		if err != nil {
			return err
		}
		data := &types.SensorEntry{
			Time:       time.Now().Unix(),
			SensorType: consts.Temperature,
			Value:      *temp,
		}
		out, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if err := rpc.CommitSensorData(&out); err != nil {
			return err
		}
	}
}
