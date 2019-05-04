package control

import (
	"fmt"
	"math"
	"strings"

	i2c "github.com/d2r2/go-i2c"
	"github.com/only1isus/ADS1115"

	"github.com/ghodss/yaml"
	"github.com/only1isus/majorProj/config"
	"github.com/only1isus/majorProj/consts"
	"github.com/stianeikeland/go-rpio"
)

// Devices is a slice of all the output devices.
type Devices struct {
	Devices []OutputDevice `yaml:"devices"`
}

// OutputDevice ...
type OutputDevice struct {
	Name      consts.OutputDevice `yaml:"name"`
	Pins      DriverPins          `yaml:"pins"`
	Rate      float64             `yaml:"rate"`
	OnTime    int64               `yaml:"onTime"`
	Every     int64               `yaml:"every"`
	Automatic bool                `yaml:"automatic"`
}

// DriverPins ...
type DriverPins struct {
	EN  uint8 `yaml:"en"`
	IN1 uint8 `yaml:"in1"`
	IN2 uint8 `yaml:"in2"`
}

type ADCSensor struct {
	Name       consts.AnalogSensor `yaml:"name"`
	Every      int64               `yaml:"every"`
	AnalogPin  int                 `yaml:"analogPin"`
	connection *ADS1115.ADS1115
}

// ADCSensor is a struct of the input sensor connected to the ADS1115 module.
type ADCSensors struct {
	AnalogSensor []ADCSensor `yaml:"analogSensor"`
}

type I2CSensor struct {
	Name       string `yaml:"name"`
	Bus        int    `yaml:"bus"`
	Address    uint8  `yaml:"address"`
	Every      int64  `yaml:"every"`
	connection *i2c.I2C
}

type I2CSensors struct {
	I2CSensors []I2CSensor `yaml:"i2cSensors"`
}

type WaterLevel ADCSensor

func NewAnalogSensor(sensorName consts.AnalogSensor) (*ADCSensor, error) {
	var analogSensors ADCSensors

	configFile, err := config.ReadConfigFile()
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(configFile, &analogSensors); err != nil {
		fmt.Println("error unmarshalling", err)
		return nil, err
	}
	for _, sensor := range analogSensors.AnalogSensor {
		if strings.ToLower(string(sensor.Name)) == strings.ToLower(string(sensorName)) {
			return &sensor, nil
		}
	}
	return nil, fmt.Errorf("cannot find %s in the config file", sensorName)
}

func NewI2CSensor(sensorName consts.I2CSensor) (*I2CSensor, error) {
	var i2cSensors I2CSensors

	configFile, err := config.ReadConfigFile()
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(configFile, &i2cSensors); err != nil {
		fmt.Println("error unmarshalling", err)
		return nil, err
	}
	for _, sensor := range i2cSensors.I2CSensors {
		if strings.ToLower(sensor.Name) == strings.ToLower(string(sensorName)) {

			return &sensor, err
		}
	}

	return nil, fmt.Errorf("cannot find %s in i2cSensor setting", sensorName)
}

func NewADS1115Device(deviceName consts.ADS1115Device) (*ADS1115.ADS1115, error) {
	type ADSDevice struct {
		Name    consts.ADS1115Device `yaml:"name"`
		Address int                  `yaml:"address"`
		Bus     int                  `yaml:"bus"`
	}
	type ADSDevices struct {
		ADSDevices []ADSDevice `yaml:"adsDevices"`
	}

	var adsDev ADSDevices
	configFile, err := config.ReadConfigFile()
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(configFile, &adsDev); err != nil {
		fmt.Println("error unmarshalling", err)
		return nil, err
	}

	fmt.Println("device ", adsDev)
	for _, adsdevice := range adsDev.ADSDevices {
		if strings.ToLower(string(adsdevice.Name)) == strings.ToLower(string(deviceName)) {
			i2cconn, err := ADS1115.NewConnection(adsdevice.Address, adsdevice.Bus)
			if err != nil {
				return nil, err
			}
			ads1115Device := ADS1115.NewADS1115Device(i2cconn)
			if err != nil {
				return nil, err
			}
			return ads1115Device, nil
		}
	}
	return nil, fmt.Errorf("failed to create ADS1115 device using %v from config file", deviceName)
}

// Get reads the config file and returns a list of nodes and error.
// When using thOutputDeviceis method, check if the slice of node is nil and handle it to avoid
// "invalid memory address or nil pointer dereference" error
func (d *Devices) Get(deviceName consts.OutputDevice) (*OutputDevice, error) {
	configFile, err := config.ReadConfigFile()
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(configFile, &d); err != nil {
		fmt.Println("error unmarshalling", err)
		return nil, err
	}
	for _, device := range d.Devices {
		if device.Name == deviceName {
			return &device, nil
		}
	}
	return nil, nil
}

// NewOutputDevice takes the name of the device and returns an instance of that device.
// Note: to avoid errors check that the name of device in the config file is the same as
// the ones in consts.go.
func NewOutputDevice(name consts.OutputDevice) (*OutputDevice, error) {
	d := Devices{}
	device, err := d.Get(name)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (o OutputDevice) On() error {
	err := rpio.Open()
	if err != nil {
		return err
	}

	defer rpio.Close()

	en := rpio.Pin(o.Pins.EN)
	in1 := rpio.Pin(o.Pins.IN1)
	in2 := rpio.Pin(o.Pins.IN2)
	// en.Pwm()
	rpio.StartPwm()
	en.Pwm()
	in1.Output()
	in2.Output()

	en.Freq(1920000)
	en.DutyCycle(uint32(o.Rate*128), 128)
	in1.High()
	in2.Low()

	return nil
}

func (o OutputDevice) OnNoPWM() error {

	err := rpio.Open()
	if err != nil {
		return err
	}

	defer rpio.Close()

	en := rpio.Pin(o.Pins.EN)
	in1 := rpio.Pin(o.Pins.IN1)
	in2 := rpio.Pin(o.Pins.IN2)

	en.Output()
	in1.Output()
	in2.Output()

	en.Freq(1920000)
	en.DutyCycle(uint32(o.Rate*128), 128)
	en.High()
	in1.High()
	in2.Low()

	return nil
}

func (o OutputDevice) ChangePWM(rate float64) error {
	err := rpio.Open()
	if err != nil {
		return err
	}
	defer rpio.Close()

	rpio.StartPwm()
	en := rpio.Pin(o.Pins.EN)
	en.Pwm()
	en.DutyCycle(uint32(rate*128), 128)
	return nil
}

// Off method turns the fan off.
func (o OutputDevice) Off() error {
	err := rpio.Open()
	defer rpio.Close()

	if err != nil {
		return err
	}

	en := rpio.Pin(o.Pins.EN)
	in1 := rpio.Pin(o.Pins.IN1)
	in2 := rpio.Pin(o.Pins.IN2)
	en.Output()
	in1.Output()
	in2.Output()

	en.Low()
	in1.Low()
	in2.Low()
	rpio.StopPwm()

	return nil
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
