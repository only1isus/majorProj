package control

import (
	"fmt"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/only1isus/majorProj/config"
	"github.com/stianeikeland/go-rpio"
	"github.com/yryz/ds18b20"
)

// Devices is a slice of all the output devices.
type Devices struct {
	Devices []OutputDevice `yaml:"devices"`
}

// OutputDevice ...
type OutputDevice struct {
	Name      string     `yaml:"name"`
	Pins      DriverPins `yaml:"pins"`
	Rate      float64    `yaml:"rate"`
	OnTime    int64      `yaml:"onTime"`
	Every     int64      `yaml:"every"`
	Automatic bool       `yaml:"automatic"`
}

// DriverPins ...
type DriverPins struct {
	EN  rpio.Pin `yaml:"en"`
	IN1 rpio.Pin `yaml:"in1"`
	IN2 rpio.Pin `yaml:"in2"`
}

// ADC AS1115 ADC module
type ADC struct {
	Name string
	Addr []string
	Info string
}

type CirculationPump OutputDevice
type Fan OutputDevice
type PhPumpUp OutputDevice

type Sensor struct {
	Name  string
	value float64
	Pin   int
}

// Temperature ...
type Temperature Sensor

type PH ADC
type EC ADC
type WaterLevel ADC

type AnalogSensors interface {
	Get()
	Save()
	Log()
}

// initPin initialize the pin and returns a pin struct.
func initializePins(pins DriverPins, ENMode, INMode rpio.Mode) (*DriverPins, error) {
	// var p []rpio.Pin
	err := rpio.Open()
	defer rpio.Close()

	if err != nil {
		return nil, err
	}
	rpio.PinMode(pins.EN, ENMode)
	rpio.PinMode(pins.IN2, INMode)
	rpio.PinMode(pins.IN1, INMode)

	return &pins, nil
}

// Get reads the config file and returns a list of nodes and error.
// When using thOutputDeviceis method, check if the slice of node is nil and handle it to avoid
// "invalid memory address or nil pointer dereference" error
func (d *Devices) Get(deviceName string) (*OutputDevice, error) {
	deviceName = strings.ToLower(deviceName)
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
			fmt.Println(device)
			return &device, nil
		}
	}
	return nil, nil
}

// Get method when called returns the current temperature.
func (t *Temperature) Get() (*float64, error) {
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
	return &t.value, err
}

// Maintain method tries to keep the temperature at the value passed to the method.
func (t *Temperature) Maintain(value float64, fan Fan) error {
	go func() error {
		temp, err := t.Get()
		if err != nil {
			return err
		}
		if *temp >= value {
		maintainTemperature:
			for t.value >= value {
				fan.On()
				time.Sleep(3 * time.Second)
				if t.value <= value {
					fan.Off()
					break maintainTemperature
				}
			}
		}
		return err
	}()
	return nil
}

// On method turns the fan on.
func (f *Fan) On() {
	go func() {
		fanPins, err := initializePins(f.Pins, rpio.Pwm, rpio.Output)
		if err != nil {
			fmt.Println(err)
		}

		// need to implement the duty cycle for the output pins
		// rate := int(f.Rate * 255)
		// fanPins.EN.Pwm()
		// fanPins.EN.Du
		// fanPins.EN.Freq(rate)

		(*fanPins).EN.High()
		(*fanPins).IN1.Low()
		(*fanPins).IN2.High()
	}()
}

// Off method turns the fan off.
func (f *Fan) Off() {
	go func() {
		fanPins, err := initializePins(f.Pins, rpio.Pwm, rpio.Output)
		if err != nil {
			fmt.Println(err)
		}
		(*fanPins).EN.Low()
		(*fanPins).IN1.Low()
		(*fanPins).IN2.Low()
	}()
}

// // GetAddresses return the addresses associated with the ADC.
// func (adc *ADC) GetAddresses() ([]string, error) {
// 	var addresses []string
// 	return addresses, nil
// }
