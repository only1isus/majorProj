package control

import "github.com/only1isus/majorProj/consts"

type CoolingFan OutputDevice

func NewCoolingFan() (CoolingFan, error) {
	coolingFan, err := NewOutputDevice(consts.CoolingFan)
	if err != nil {
		return CoolingFan{}, err
	}
	return CoolingFan(*coolingFan), nil
}

func (cf CoolingFan) On() {
	fan := OutputDevice(cf)
	fan.On()
}

func (cf CoolingFan) OnNoPWM() {
	fan := OutputDevice(cf)
	fan.OnNoPWM()
}

func (cf CoolingFan) Off() {
	fan := OutputDevice(cf)
	fan.Off()
}

func (cf *CoolingFan) ChangePWM(rate float64) {
	fan := OutputDevice(*cf)
	fan.Rate = rate
}
