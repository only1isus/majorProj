package control

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	i2c "github.com/d2r2/go-i2c"
	sht3x "github.com/d2r2/go-sht3x"

	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/rpc"
	"github.com/only1isus/majorProj/types"
)

type HumiditySensor I2CSensor

func NewHumiditySensor() (*HumiditySensor, error) {
	humiditySensor, err := NewI2CSensor(consts.Sth3xHumidity)
	if err != nil {
		return nil, err
	}
	fmt.Println(*humiditySensor)
	hs := HumiditySensor(*humiditySensor)
	return &hs, nil
}

func (h *HumiditySensor) Get() (*float64, error) {
	fmt.Println("reading humidity")
	time.Sleep(time.Second * 1)
	i2cconn, err := i2c.NewI2C(h.Address, h.Bus)
	if err != nil {
		log.Fatal(err)
	}
	defer i2cconn.Close()

	sensor := sht3x.NewSHT3X()
	_, humidity, err := sensor.ReadTemperatureAndRelativeHumidity(i2cconn, sht3x.RepeatabilityLow)
	if err != nil {
		return nil, err
	}
	hum := new(float64)
	*hum = ToFixed(float64(humidity), 1)
	return hum, nil
}

func (h *HumiditySensor) ReadAndCommit() error {
	for {
		timer := time.NewTimer(time.Minute * time.Duration(h.Every))
		defer timer.Stop()
		// wait for the timer to reach its limit
		<-timer.C

		hum, err := h.Get()
		if err != nil {
			return err
		}
		out := &types.SensorEntry{
			Time:       time.Now().Unix(),
			SensorType: consts.Humidity,
			Value:      *hum,
		}
		data, err := json.Marshal(out)
		if err != nil {
			return err
		}
		if err := rpc.CommitSensorData(&data); err != nil {
			return err
		}
	}
}

func (h HumiditySensor) Close() error {
	return h.connection.Close()
}
