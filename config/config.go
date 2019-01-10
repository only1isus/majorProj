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
	path, err := filepath.Abs(".")
	base := filepath.Base(path)
	dir := filepath.Dir(path)
	if err != nil {
		fmt.Printf("cannot find root. %s", err.Error())
		os.Exit(2)
	}
	return strings.Join([]string{dir, base}, "/")
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
	return fileByte, nil
}
