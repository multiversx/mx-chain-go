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
	Name       string                 `json:"name"`
	Enabled    bool                   `json:"enabled"`
	Triggers   []string               `json:"triggers"`
	Parameters map[string]interface{} `json:"parameters"`
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

	err = config.verify()
	if err != nil {
		return nil, fmt.Errorf("config verification failed: %v", err)
	}

	config.populateFailuresByName()
	return &config, nil
}

func (config *chaosConfig) verify() error {
	knownFailures := make(map[failureName]struct{})

	knownFailures[failureProcessTransactionShouldReturnError] = struct{}{}
	knownFailures[failureShouldCorruptSignature] = struct{}{}
	knownFailures[failureShouldSkipWaitingForSignatures] = struct{}{}
	knownFailures[failureShouldReturnErrorInCheckSignaturesValidity] = struct{}{}
	knownFailures[failureShouldDelayBroadcastingFinalBlockAsLeader] = struct{}{}
	knownFailures[failureShouldCorruptLeaderSignature] = struct{}{}
	knownFailures[failureShouldDelayLeaderSignature] = struct{}{}
	knownFailures[failureShouldSkipSendingBlock] = struct{}{}

	for _, failure := range config.Failures {
		name := failureName(failure.Name)

		if _, ok := knownFailures[name]; !ok {
			return fmt.Errorf("unknown failure: %s", name)
		}

		if len(failure.Triggers) == 0 {
			return fmt.Errorf("failure %s has no triggers", name)
		}

		if name == failureShouldDelayBroadcastingFinalBlockAsLeader {
			if failure.getParameterAsFloat64("duration") == 0 {
				return fmt.Errorf("failure %s requires the parameter 'duration'", name)
			}
		}

		if name == failureShouldDelayLeaderSignature {
			if failure.getParameterAsFloat64("duration") == 0 {
				return fmt.Errorf("failure %s requires the parameter 'duration'", name)
			}
		}
	}

	return nil
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

func (config *chaosConfig) getFailureParameterAsFloat64(failureName failureName, parameterName string) float64 {
	failure, ok := config.getFailureByName(failureName)
	if !ok {
		return 0
	}

	return failure.getParameterAsFloat64(parameterName)
}

func (failure *failureDefinition) getParameterAsFloat64(parameterName string) float64 {
	value, ok := failure.Parameters[parameterName]
	if !ok {
		return 0
	}

	floatValue, ok := value.(float64)
	if !ok {
		return 0
	}

	return floatValue
}
