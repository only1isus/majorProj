package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	configName     = "config.yaml"
	configFilePath = ""
)

func getPath() string {
	path, err := filepath.Abs("./")

	if err != nil {
		fmt.Printf("cannot find root. %s", err.Error())
		os.Exit(2)
	}
	return path
	// return strings.Join([]string{filepath.Dir(path), configFilePath}, "/")
}

// ReadYamlFile - reads the .yaml file and returns the []byte
func readYamlFile() ([]byte, error) {
	filePath := getPath()
	fileByte := make([]byte, 2048)
	fullpath := strings.Join([]string{filePath, configName}, "/")
	yamlfile, err := ioutil.ReadFile(fullpath)
	fileByte = append(fileByte, yamlfile...)

	if err != nil {
		return nil, err
	}
	return fileByte, nil
}

// Get reads the config file and returns a list of nodes and error.
// When using thOutputDeviceis method, check if the slice of node is nil and handle it to avoid
// "invalid memory address or nil pointer dereference" error
func (o *Devices) Get(deviceName string) (*OutputDevice, error) {
	deviceName = strings.ToLower(deviceName)
	filePath := getPath()
	// fileByte := make([]byte, 2048)
	fullpath := strings.Join([]string{filePath, configName}, "/")
	yamlfile, err := ioutil.ReadFile(fullpath)
	// fileByte = append(fileByte, yamlfile...)
	fmt.Println(string(yamlfile))
	if err != nil {
		return nil, err
	}
	if err != nil {
		fmt.Println("got an error")
	}
	if err := yaml.Unmarshal(yamlfile, &o); err != nil {
		fmt.Println("error unmarshalling", err)
		return nil, err
	}
	for _, device := range o.Devices {
		if device.Name == deviceName {
			// device.Name = strings.ToLower(device.Name)
			// fmt.Println(device)
			return &device, nil
		}
	}
	return nil, nil
}
