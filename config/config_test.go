package config

import (
	"testing"
)

func TestGetConfig(t *testing.T) {
	fan := Devices{}
	deviceName := "phuppump"
	fanSetting, err := fan.Get(deviceName)
	if err != nil {
		t.Log(err)
	}
	if fanSetting.Name != deviceName {
		t.Error("cannot find the fan setting")
	}

}
