package config

import (
	"encoding/json"
	"os"
)

type JsonConfig struct {
	Port  int
	Peers []string
	Size  int
}

func (config *JsonConfig) ReadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}

	return nil
}
