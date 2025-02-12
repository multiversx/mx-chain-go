package chaos

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type chaosConfig struct {
	Failures         []failureDefinition `json:"failures"`
	ReusableTriggers []string            `json:"reusableTriggers"`

	failuresByName map[string]failureDefinition
}

type failureDefinition struct {
	Name     string   `json:"name"`
	Enabled  bool     `json:"enabled"`
	Triggers []string `json:"triggers"`
}

func newChaosConfigFromFile(filePath string) (*chaosConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %v", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	var config chaosConfig

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config JSON: %v", err)
	}

	config.populateFailuresByName()
	return &config, nil
}

func (config *chaosConfig) populateFailuresByName() {
	config.failuresByName = make(map[string]failureDefinition)

	for _, failure := range config.Failures {
		config.failuresByName[failure.Name] = failure
	}
}

func (config *chaosConfig) getFailureByName(name failureName) (failureDefinition, bool) {
	failure, ok := config.failuresByName[string(name)]
	return failure, ok
}
