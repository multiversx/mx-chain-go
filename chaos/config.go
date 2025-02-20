package chaos

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type chaosConfig struct {
	Profiles            []chaosProfile `json:"profiles"`
	SelectedProfileName string         `json:"selectedProfile"`
	selectedProfile     chaosProfile
}

type chaosProfile struct {
	Name             string              `json:"name"`
	Failures         []failureDefinition `json:"failures"`
	ReusableTriggers []string            `json:"reusableTriggers"`
	failuresByName   map[string]failureDefinition
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

	for _, profile := range config.Profiles {
		profile.populateFailuresByName()

		if profile.Name == config.SelectedProfileName {
			config.selectedProfile = profile
		}
	}

	return &config, nil
}

func (config *chaosConfig) verify() error {
	if len(config.SelectedProfileName) == 0 {
		return fmt.Errorf("no selected profile")
	}

	selectedProfileExists := false

	for _, profile := range config.Profiles {
		err := profile.verify()
		if err != nil {
			return fmt.Errorf("profile %s verification failed: %v", profile.Name, err)
		}

		if profile.Name == config.SelectedProfileName {
			selectedProfileExists = true
		}
	}

	if !selectedProfileExists {
		return fmt.Errorf("selected profile does not exist: %s", config.SelectedProfileName)
	}

	return nil
}

func (profile *chaosProfile) verify() error {
	knownFailures := make(map[failureName]struct{})

	knownFailures[failureCreatingBlockError] = struct{}{}
	knownFailures[failureProcessingBlockError] = struct{}{}
	knownFailures[failurePanicOnEpochChange] = struct{}{}
	knownFailures[failureConsensusCorruptSignature] = struct{}{}
	knownFailures[failureConsensusV1ReturnErrorInCheckSignaturesValidity] = struct{}{}
	knownFailures[failureConsensusV1DelayBroadcastingFinalBlockAsLeader] = struct{}{}
	knownFailures[failureConsensusV2CorruptLeaderSignature] = struct{}{}
	knownFailures[failureConsensusV2DelayLeaderSignature] = struct{}{}
	knownFailures[failureConsensusV2SkipSendingBlock] = struct{}{}

	for _, failure := range profile.Failures {
		name := failureName(failure.Name)

		if _, ok := knownFailures[name]; !ok {
			return fmt.Errorf("unknown failure: %s", name)
		}

		if len(failure.Triggers) == 0 {
			return fmt.Errorf("failure %s has no triggers", name)
		}

		if name == failureConsensusV1DelayBroadcastingFinalBlockAsLeader {
			if failure.getParameterAsFloat64("duration") == 0 {
				return fmt.Errorf("failure %s requires the parameter 'duration'", name)
			}
		}

		if name == failureConsensusV2DelayLeaderSignature {
			if failure.getParameterAsFloat64("duration") == 0 {
				return fmt.Errorf("failure %s requires the parameter 'duration'", name)
			}
		}
	}

	return nil
}

func (profile *chaosProfile) populateFailuresByName() {
	profile.failuresByName = make(map[string]failureDefinition)

	for _, failure := range profile.Failures {
		profile.failuresByName[failure.Name] = failure
	}
}

func (profile *chaosProfile) getFailureByName(name failureName) (failureDefinition, bool) {
	failure, ok := profile.failuresByName[string(name)]
	return failure, ok
}

func (profile *chaosProfile) getFailureParameterAsFloat64(failureName failureName, parameterName string) float64 {
	failure, ok := profile.getFailureByName(failureName)
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
