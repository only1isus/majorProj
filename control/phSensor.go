package control

import (
	"encoding/json"
	"time"

	"github.com/only1isus/ADS1115"
	"github.com/only1isus/majorProj/consts"
	"github.com/only1isus/majorProj/rpc"
	"github.com/only1isus/majorProj/types"
)

type PHSensor ADCSensor

var (
	defaultCalibrationValues = map[int][]float64{
		4: []float64{4, 1.993},
		7: []float64{7, 1.476},
	}
)

func calculatePH(analogValue float64) float64 {
	m := float64((defaultCalibrationValues[4][1] - defaultCalibrationValues[7][1]) / (defaultCalibrationValues[4][0] - defaultCalibrationValues[7][0]))
	c := defaultCalibrationValues[4][1] - m*defaultCalibrationValues[4][0]
	pHValue := (analogValue - c) / m
	return ToFixed(pHValue, 2)
}

func NewPHSensor(connection *ADS1115.ADS1115) (*PHSensor, error) {

	phsensor, err := NewAnalogSensor(consts.PHSensor)
	if err != nil {
		return nil, err
	}
	phsensor.connection = connection
	ph := PHSensor(*phsensor)
	return &ph, nil
}

func (ph *PHSensor) Calibrate(bufferSolution4 float64, bufferSolution7 float64) {
	defaultCalibrationValues[4][1] = bufferSolution4
	defaultCalibrationValues[7][1] = bufferSolution7
}

func (ph *PHSensor) Get() (*float64, error) {
	voltageValue, err := ph.connection.Read(ph.AnalogPin)
	if err != nil {
		return nil, err
	}
	out := new(float64)
	*out = calculatePH(float64(voltageValue))
	return out, nil
}

func (ph PHSensor) ReadAndCommit() error {
	for {
		timer := time.NewTimer(time.Minute * time.Duration(ph.Every))
		defer timer.Stop()
		// wait for the timer to reach its limit
		<-timer.C
		phvalue, err := ph.Get()
		if err != nil {
			return err
		}
		data := &types.SensorEntry{
			SensorType: consts.PH,
			Time:       time.Now().Unix(),
			Value:      *phvalue,
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

func (ph *PHSensor) Close() error {
	return ph.connection.Close()
}
