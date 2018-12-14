package control

import (
	"fmt"
	"log"
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
	EN  uint8 `yaml:"en"`
	IN1 uint8 `yaml:"in1"`
	IN2 uint8 `yaml:"in2"`
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
func (t *Temperature) Maintain(value float64, fan *OutputDevice) error {
	go func() error {

		temp, err := t.Get()
		if err != nil {
			return err
		}
		log.Printf("Current temperature is %v", *temp)
		for {
			temp, err := t.Get()
			if err != nil {
				return err
			}
			val := *temp
			log.Printf("Current temperature is %v", val)
			if val >= value {
				log.Printf("Turning on fan. Current temperature is %v, limit set to %v", val, value)
				if err := fan.On(); err != nil {
					return err
				}
				// maintainTemperature:
				for {
					// fan.On()
					currentTemp, err := t.Get()
					if err != nil {
						return err
					}
					if value >= *currentTemp {
						log.Printf("Temperature now at %v. Cooling for another 30s", *currentTemp)
						time.Sleep(30 * time.Second)
						log.Println("Done")
						if err := fan.Off(); err != nil {
							return err
						}
						break //maintainTemperature
					}
					// time.Sleep(3 * time.Second)
				}
			}
			time.Sleep(30 * time.Second)
			// return nil
		}
	}()
	return nil
}

// On method turns the fan on.
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
	// defer rpio.StopPwm()
	en.Pwm()
	in1.Output()
	in2.Output()

	en.Freq(19200000)
	en.DutyCycle(uint32(o.Rate*128), 128)
	in1.Low()
	in2.High()

	// log := util.LogEntry{}
	// log.Message = "Fan turned on"
	// log.Success = true
	// log.Time = time.Now().Unix()
	// err = log.Add()
	// if err != nil {
	// 	fmt.Println(err)
	// }
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

	// log := util.LogEntry{}
	// log.Message = "Fan turned off"
	// log.Success = true
	// log.Time = time.Now().Unix()
	// err = log.Add()

	// if err != nil {
	// 	fmt.Println(err)
	// }
	return nil
}
