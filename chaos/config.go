package chaos

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
)

type chaosConfig struct {
	NumCallsDivisor_processTransaction_shouldReturnError       int `json:"numCallsDivisorProcessTransactionShouldReturnError"`
	NumCallsDivisor_maybeCorruptSignature                      int `json:"numCallsDivisorMaybeCorruptSignature"`
	NumCallsDivisor_shouldSkipWaitingForSignatures             int `json:"numCallsDivisorShouldSkipWaitingForSignatures"`
	NumCallsDivisor_shouldReturnErrorInCheckSignaturesValidity int `json:"numCallsDivisorShouldReturnErrorInCheckSignaturesValidity"`
	NumCallsDivisor_consensusV2_maybeCorruptLeaderSignature    int `json:"numCallsDivisorConsensusV2MaybeCorruptLeaderSignature"`
	NumCallsDivisor_consensusV2_shouldSkipSendingBlock         int `json:"numCallsDivisorConsensusV2ShouldSkipSendingBlock"`
}

func (config *chaosConfig) applyDefaults() {
	if config.NumCallsDivisor_processTransaction_shouldReturnError == 0 {
		config.NumCallsDivisor_processTransaction_shouldReturnError = math.MaxInt
	}
	if config.NumCallsDivisor_maybeCorruptSignature == 0 {
		config.NumCallsDivisor_maybeCorruptSignature = math.MaxInt
	}
	if config.NumCallsDivisor_shouldSkipWaitingForSignatures == 0 {
		config.NumCallsDivisor_shouldSkipWaitingForSignatures = math.MaxInt
	}
	if config.NumCallsDivisor_shouldReturnErrorInCheckSignaturesValidity == 0 {
		config.NumCallsDivisor_shouldReturnErrorInCheckSignaturesValidity = math.MaxInt
	}
	if config.NumCallsDivisor_consensusV2_maybeCorruptLeaderSignature == 0 {
		config.NumCallsDivisor_consensusV2_maybeCorruptLeaderSignature = math.MaxInt
	}
	if config.NumCallsDivisor_consensusV2_shouldSkipSendingBlock == 0 {
		config.NumCallsDivisor_consensusV2_shouldSkipSendingBlock = math.MaxInt
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

	config.applyDefaults()

	return config, nil
}
