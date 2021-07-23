package common

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/config"
)

// LoadP2PConfig returns a P2PConfig by reading the config file provided
func LoadP2PConfig(filepath string) (*config.P2PConfig, error) {
	cfg := &config.P2PConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadMainConfig returns a Config by reading the config file provided
func LoadMainConfig(filepath string) (*config.Config, error) {
	cfg := &config.Config{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadApiConfig returns a ApiRoutesConfig by reading the config file provided
func LoadApiConfig(filepath string) (*config.ApiRoutesConfig, error) {
	cfg := &config.ApiRoutesConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadEconomicsConfig returns a EconomicsConfig by reading the config file provided
func LoadEconomicsConfig(filepath string) (*config.EconomicsConfig, error) {
	cfg := &config.EconomicsConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadSystemSmartContractsConfig returns a SystemSmartContractsConfig by reading the config file provided
func LoadSystemSmartContractsConfig(filepath string) (*config.SystemSmartContractsConfig, error) {
	cfg := &config.SystemSmartContractsConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadRatingsConfig returns a RatingsConfig by reading the config file provided
func LoadRatingsConfig(filepath string) (*config.RatingsConfig, error) {
	cfg := &config.RatingsConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return &config.RatingsConfig{}, err
	}

	return cfg, nil
}

// LoadPreferencesConfig returns a Preferences by reading the config file provided
func LoadPreferencesConfig(filepath string) (*config.Preferences, error) {
	cfg := &config.Preferences{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadExternalConfig returns a ExternalConfig by reading the config file provided
func LoadExternalConfig(filepath string) (*config.ExternalConfig, error) {
	cfg := &config.ExternalConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, fmt.Errorf("cannot load external config: %w", err)
	}

	return cfg, nil
}

// LoadGasScheduleConfig returns a map[string]uint64 of gas costs read from the provided config file
func LoadGasScheduleConfig(filepath string) (map[string]map[string]uint64, error) {
	gasScheduleConfig, err := core.LoadTomlFileToMap(filepath)
	if err != nil {
		return nil, err
	}

	flattenedGasSchedule := make(map[string]map[string]uint64)
	for libType, costs := range gasScheduleConfig {
		flattenedGasSchedule[libType] = make(map[string]uint64)
		costsMap := costs.(map[string]interface{})
		for operationName, cost := range costsMap {
			flattenedGasSchedule[libType][operationName] = uint64(cost.(int64))
		}
	}

	return flattenedGasSchedule, nil
}

// LoadEpochConfig returns an EpochConfig by reading from the provided config file
func LoadEpochConfig(filepath string) (*config.EpochConfig, error) {
	cfg := &config.EpochConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
