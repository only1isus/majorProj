package config

import (
	"testing"
)

func TestGetConfig(t *testing.T) {

	data, err := ReadConfigFile()
	if err != nil {
		t.Fail()
	}
	if data == nil {
		t.Log("reading file failed.")
		t.Fail()
	}
}
