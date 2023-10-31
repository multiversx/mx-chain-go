package components

import (
	"os"

	"github.com/pelletier/go-toml"
)

// LoadConfigFromFile will try to load the config from the specified file
func LoadConfigFromFile(filename string, config interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(data, config)

	return err
}
