package control

import (
	"github.com/only1isus/majorProj/consts"
)

type CirculationPump OutputDevice

func NewCirculationPump() (CirculationPump, error) {
	cp, err := NewOutputDevice(consts.CoolingFan)
	if err != nil {
		return CirculationPump{}, err
	}
	return CirculationPump(*cp), nil
}

func (cp CirculationPump) On() {
	pump := OutputDevice(cp)
	pump.On()
}

func (cp CirculationPump) OnNoPWM() {
	pump := OutputDevice(cp)
	pump.OnNoPWM()
}

func (cp CirculationPump) Off() {
	pump := OutputDevice(cp)
	pump.Off()
}

func (cp *CirculationPump) ChangePWM(rate float64) {
	pump := OutputDevice(*cp)
	pump.Rate = rate
}
