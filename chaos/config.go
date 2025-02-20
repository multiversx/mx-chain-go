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
	Name           string              `json:"name"`
	Failures       []failureDefinition `json:"failures"`
	failuresByName map[string]failureDefinition
}

type failureDefinition struct {
	Name       string                 `json:"name"`
	Enabled    bool                   `json:"enabled"`
	Type       string                 `json:"type"`
	OnPoints   []string               `json:"onPoints"`
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
	for _, failure := range profile.Failures {
		name := failure.Name
		failType := failType(failure.Type)

		if len(name) == 0 {
			return fmt.Errorf("all failures must have a name")
		}

		if len(failType) == 0 {
			return fmt.Errorf("failure '%s' has no fail type", name)
		}

		if _, ok := knownFailTypes[failType]; !ok {
			return fmt.Errorf("failure '%s' has unknown fail type: '%s'", name, failType)
		}

		if len(failure.OnPoints) == 0 {
			return fmt.Errorf("failure '%s' has no points configured", name)
		}

		for _, point := range failure.OnPoints {
			if _, ok := knownPoints[pointName(point)]; !ok {
				return fmt.Errorf("failure '%s' has unknown activation point: '%s'", name, point)
			}
		}

		if len(failure.Triggers) == 0 {
			return fmt.Errorf("failure '%s' has no triggers configured", name)
		}

		if failType == failTypeSleep {
			if failure.getParameterAsFloat64("duration") == 0 {
				return fmt.Errorf("failure '%s', with fail type '%s', requires the parameter 'duration'", name, failType)
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

func (profile *chaosProfile) getFailureByName(name string) (failureDefinition, bool) {
	failure, ok := profile.failuresByName[name]
	return failure, ok
}

func (profile *chaosProfile) getFailureParameterAsFloat64(failureName string, parameterName string) float64 {
	failure, ok := profile.getFailureByName(failureName)
	if !ok {
		return 0
	}

	return failure.getParameterAsFloat64(parameterName)
}

func (profile *chaosProfile) getFailuresOnPoint(point string) []failureDefinition {
	var failures []failureDefinition

	for _, failure := range profile.Failures {
		if failure.Enabled && failure.isOnPoint(point) {
			failures = append(failures, failure)
		}
	}

	return failures
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

func (failure *failureDefinition) isOnPoint(pointName string) bool {
	for _, point := range failure.OnPoints {
		if point == pointName {
			return true
		}
	}

	return false
}
