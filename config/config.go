package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	ConfigName     = "config.yaml"
	configFilePath = ""
)

// GetPath returns the path of the file.
func getPath() string {
	path, err := filepath.Abs("../")

	if err != nil {
		fmt.Printf("cannot find root. %s", err.Error())
		os.Exit(2)
	}
	return path
	// return strings.Join([]string{filepath.Dir(path), configFilePath}, "/")
}

// ReadConfigFile - reads the .yaml file and returns the []byte
func ReadConfigFile() ([]byte, error) {
	filePath := getPath()
	fullpath := strings.Join([]string{filePath, ConfigName}, "/")
	yamlfile, err := ioutil.ReadFile(fullpath)
	var fileByte []byte
	fileByte = append(fileByte, yamlfile...)

	if err != nil {
		return nil, err
	}
	fmt.Println(len(fileByte), len(yamlfile))
	return fileByte, nil
}
