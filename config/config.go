package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	configName = "config.yaml"
)

type out struct {
}

type databaseConnection struct {
	DatabaseConnection []map[string]interface{} `yaml:"databaseConnection"`
}

func getPath() string {

	path, err := filepath.Abs("./")
	if err != nil {
		fmt.Printf("cannot find root. %+v", err)
		os.Exit(2)
	}
	return path
}

// ReadConfig reads value for the key provided for the .yaml file and parse the data
func ReadConfig(key string) (databaseConnection, bool) {
	path := getPath()
	fullpath := strings.Join([]string{path, configName}, "/")
	yamlfile, err := ioutil.ReadFile(fullpath)
	if err != nil {
		fmt.Printf("cannot open %s , make sure the file exists.", configName)
		os.Exit(2)
	}

	switch key {
	case "database":
		var db databaseConnection
		err := yaml.Unmarshal(yamlfile, &db)
		if err != nil {
			log.Println(err)
		}

		fmt.Println(db)
		return db, true
	}
	return databaseConnection{}, false
}
