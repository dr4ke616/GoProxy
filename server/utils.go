package server

import (
	"encoding/json"
	"log"
	"os"
)

// Parses the config file
func LoadFromConfig(out interface{}, config ...string) error {
	var err error
	var file *os.File
	var config_file = "config.json"

	if len(config) > 0 {
		config_file = config[0]
	}

	if file, err = os.Open(config_file); err != nil {
		log.Println("Failed to open config file:")
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	if err = decoder.Decode(out); err != nil {
		log.Println("Failed to decode JSON config file:")
		return err
	}
	return nil
}
