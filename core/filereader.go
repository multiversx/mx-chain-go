package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/pelletier/go-toml"
)

// LoadFile method to open file from given path - does not close the file
func LoadFile(relativePath string, log *logger.Logger) (*os.File, error) {
	path, err := filepath.Abs(relativePath)
	fmt.Println(path)
	if err != nil {
		log.Error("cannot create absolute path for the provided file", err.Error())
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// LoadTomlFile method to open and decode toml file
func LoadTomlFile(dest interface{}, relativePath string, log *logger.Logger) error {
	f, err := LoadFile(relativePath, log)
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
		if err != nil {
			log.Error("cannot close file: ", err.Error())
		}
	}()

	return toml.NewDecoder(f).Decode(dest)
}

// LoadJsonFile method to open and decode json file
func LoadJsonFile(dest interface{}, relativePath string, log *logger.Logger) error {
	f, err := LoadFile(relativePath, log)
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
		if err != nil {
			log.Error("cannot close file: ", err.Error())
		}
	}()

	return json.NewDecoder(f).Decode(dest)
}
