package control

import (
	"encoding/json"
	"time"

	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/types"
)

type HumiditySensor Sensor

func NewHumiditySensor() HumiditySensor {
	return HumiditySensor{}
}

func (h *HumiditySensor) Get() (*float64, error) {
	// sensorType := dht.DHT11
	// _, humidity, _, err := dht.ReadDHTxxWithRetry(sensorType, 4, false, 10)
	// if err != nil {
	// 	return nil, err
	// }

	// hum := float64(humidity)
	// return &hum, err
	return nil, nil
}

func (h *HumiditySensor) Prepare() (*[]byte, error) {
	hum, err := h.Get()
	if err != nil {
		return nil, err
	}
	entry := types.SensorEntry{
		Time:       time.Now().Unix(),
		SensorType: consts.Humidity,
		Value:      *hum,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
