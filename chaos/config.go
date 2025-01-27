package chaos

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type chaosConfig struct {
	NumCallsDivisor_maybeCorruptSignature                      int `json:"numCallsDivisorMaybeCorruptSignature"`
	NumCallsDivisor_shouldSkipWaitingForSignatures             int `json:"numCallsDivisorShouldSkipWaitingForSignatures"`
	NumCallsDivisor_shouldReturnErrorInCheckSignaturesValidity int `json:"numCallsDivisorShouldReturnErrorInCheckSignaturesValidity"`
	NumCallsDivisor_processTransaction_shouldReturnError       int `json:"numCallsDivisorProcessTransactionShouldReturnError"`
}

func newChaosConfig() chaosConfig {
	return chaosConfig{
		NumCallsDivisor_maybeCorruptSignature:                      5,
		NumCallsDivisor_shouldSkipWaitingForSignatures:             7,
		NumCallsDivisor_shouldReturnErrorInCheckSignaturesValidity: 11,
		NumCallsDivisor_processTransaction_shouldReturnError:       30001,
	}
}

func loadChaosConfigFromFile(filePath string) (chaosConfig, error) {
	var config chaosConfig

	file, err := os.Open(filePath)
	if err != nil {
		return config, fmt.Errorf("could not open config file: %v", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return config, fmt.Errorf("could not read config file: %v", err)
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, fmt.Errorf("could not unmarshal config JSON: %v", err)
	}

	return config, nil
}
