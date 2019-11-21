package core

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("core")

// LoadP2PConfig returns a P2PConfig by reading the config file provided
func LoadP2PConfig(filepath string) (*config.P2PConfig, error) {
	cfg := &config.P2PConfig{}
	err := LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadServersPConfig returns a ServersConfig by reading the config file provided
func LoadServersPConfig(filepath string) (*config.ServersConfig, error) {
	cfg := &config.ServersConfig{}
	err := LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadGasScheduleConfig returns a map[string]uint64 of gas costs read from the provided config file
func LoadGasScheduleConfig(filepath string) (map[string]uint64, error) {
	gasScheduleConfig, err := LoadTomlFileToMap(filepath)
	if err != nil {
		return nil, err
	}

	flattenedGasSchedule := make(map[string]uint64)
	for _, costs := range gasScheduleConfig {
		costsMap := costs.(map[string]interface{})
		for operationName, cost := range costsMap {
			flattenedGasSchedule[operationName] = uint64(cost.(int64))
		}
	}

	return flattenedGasSchedule, nil
}
