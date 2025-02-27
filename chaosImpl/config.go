package chaosImpl

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type chaosConfig struct {
	Profiles            []*chaosProfile `json:"profiles"`
	SelectedProfileName string          `json:"selectedProfile"`
	selectedProfile     *chaosProfile
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
	}

	err = config.selectProfile(config.SelectedProfileName)
	if err != nil {
		return nil, fmt.Errorf("could not select profile: %v", err)
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

func (config *chaosConfig) selectProfile(name string) error {
	profile, err := config.getProfileByName(name)
	if err != nil {
		return err
	}

	config.selectedProfile = profile
	return nil
}

func (config *chaosConfig) getProfileByName(name string) (*chaosProfile, error) {
	if len(name) == 0 {
		name = config.SelectedProfileName
	}

	for _, profile := range config.Profiles {
		if profile.Name == name {
			return profile, nil
		}
	}

	return nil, fmt.Errorf("profile not found: %s", name)
}

func (config *chaosConfig) toggleFailure(profileName string, failureName string, enabled bool) error {
	profile, err := config.getProfileByName(profileName)
	if err != nil {
		return err
	}

	return profile.toggleFailure(failureName, enabled)
}

func (config *chaosConfig) addFailure(profileName string, failure *failureDefinition) error {
	profile, err := config.getProfileByName(profileName)
	if err != nil {
		return err
	}

	return profile.addFailure(failure)
}
