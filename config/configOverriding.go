package config

import (
	"fmt"
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common/reflectcommon"
)

const (
	configTomlFile       = "config.toml"
	enableEpochsTomlFile = "enableEpochs.toml"
	p2pTomlFile          = "p2p.toml"
	externalTomlFile     = "external.toml"
)

var (
	availableConfigFilesForOverriding = []string{configTomlFile, enableEpochsTomlFile, p2pTomlFile, externalTomlFile}
	log                               = logger.GetOrCreate("config")
)

// OverrideConfigValues will override config values for the specified configurations
func OverrideConfigValues(newConfigs []OverridableConfig, configs *Configs) error {
	var err error
	for _, newConfig := range newConfigs {
		switch newConfig.File {
		case configTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.GeneralConfig, newConfig.Path, newConfig.Value)
		case enableEpochsTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.EpochConfig, newConfig.Path, newConfig.Value)
		case p2pTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.P2pConfig, newConfig.Path, newConfig.Value)
		case externalTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.ExternalConfig, newConfig.Path, newConfig.Value)
		default:
			err = fmt.Errorf("invalid config file <%s>. Available options are %s", newConfig.File, strings.Join(availableConfigFilesForOverriding, ","))
		}

		if err != nil {
			return err
		}

		log.Info("updated config value", "file", newConfig.File, "path", newConfig.Path, "value", newConfig.Value)
	}

	return nil
}
