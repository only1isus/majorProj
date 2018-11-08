package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	configName = "config.yaml"
)

// Database - database settings
type Database struct {
	DatabaseConnection []map[string]interface{} `yaml:"databaseConnection"`
}

// Pi - raspeberry pi settings
type Pi struct {
	PiConfig []PiConf `yaml:"piConfig"`
}

type PiConf struct {
	Mode     string      `yaml:"mode"`
	Username string      `yaml:"username"`
	Password interface{} `yaml:"password"`
}

type dummy interface{}

type ConfigSetting interface {
	Parse()
}

// Parse takes the setting struct to be parsed as well as the key as a string.
func (p Pi) Parse(key string) interface{} {
	r := reflect.ValueOf(p.PiConfig[0])
	val := reflect.Indirect(r).FieldByName(strings.Title(key))
	return val
}

// func (d Database) Parse(key string) []map[string]interface{} {
// 	// r := reflect.ValueOf(d.DatabaseConnection)
// 	// val := reflect.Indirect(r).FieldByName(strings.Title(key))
// 	return d.DatabaseConnection
// }

func getPath() string {

	path, err := filepath.Abs("./")
	if err != nil {
		fmt.Printf("cannot find root. %+v", err)
		os.Exit(2)
	}
	return path
}

func handleConfig(i interface{}, y []byte) (interface{}, bool) {
	err := yaml.Unmarshal(y, &i)
	if err != nil {
		log.Println(err)
		return nil, false
	}
	return i, true
}

// ReadConfig reads value for the key provided for the .yaml file and parse the data
func ReadConfig(key string) (interface{}, bool) {
	path := getPath()
	fullpath := strings.Join([]string{path, configName}, "/")
	yamlfile, err := ioutil.ReadFile(fullpath)
	if err != nil {
		fmt.Printf("cannot open %s , make sure the file exists.", configName)
		os.Exit(2)
	}
	// var i configSetting
	switch key {
	case "database":
		var dbconfig Database
		err := yaml.Unmarshal(yamlfile, &dbconfig)
		if err != nil {
			log.Println(err)
			return nil, false
		}

		return Database(dbconfig), true

	case "pi":
		var piconfig Pi
		err := yaml.Unmarshal(yamlfile, &piconfig)
		if err != nil {
			log.Println(err)
			return nil, false
		}
		return Pi(piconfig), true

	}

	return nil, false
}
