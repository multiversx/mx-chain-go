package overridableConfig

import (
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-go/common/reflectcommon"
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	apiTomlFile            = "api.toml"
	configTomlFile         = "config.toml"
	economicsTomlFile      = "economics.toml"
	enableEpochsTomlFile   = "enableEpochs.toml"
	enableRoundsTomlFile   = "enableRounds.toml"
	externalTomlFile       = "external.toml"
	fullArchiveP2PTomlFile = "fullArchiveP2P.toml"
	p2pTomlFile            = "p2p.toml"
	ratingsTomlFile        = "ratings.toml"
	systemSCTomlFile       = "systemSmartContractsConfig.toml"
)

var (
	availableConfigFilesForOverriding = []string{
		apiTomlFile,
		configTomlFile,
		economicsTomlFile,
		enableEpochsTomlFile,
		enableRoundsTomlFile,
		externalTomlFile,
		fullArchiveP2PTomlFile,
		p2pTomlFile,
		ratingsTomlFile,
		systemSCTomlFile,
	}
	log = logger.GetOrCreate("config")
)

// OverrideConfigValues will override config values for the specified configurations
func OverrideConfigValues(newConfigs []config.OverridableConfig, configs *config.Configs) error {
	var err error
	for _, newConfig := range newConfigs {
		switch newConfig.File {
		case apiTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.ApiRoutesConfig, newConfig.Path, newConfig.Value)
		case configTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.GeneralConfig, newConfig.Path, newConfig.Value)
		case economicsTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.EconomicsConfig, newConfig.Path, newConfig.Value)
		case enableEpochsTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.EpochConfig, newConfig.Path, newConfig.Value)
		case enableRoundsTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.RoundConfig, newConfig.Path, newConfig.Value)
		case externalTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.ExternalConfig, newConfig.Path, newConfig.Value)
		case fullArchiveP2PTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.FullArchiveP2pConfig, newConfig.Path, newConfig.Value)
		case p2pTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.MainP2pConfig, newConfig.Path, newConfig.Value)
		case ratingsTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.RatingsConfig, newConfig.Path, newConfig.Value)
		case systemSCTomlFile:
			err = reflectcommon.AdaptStructureValueBasedOnPath(configs.SystemSCConfig, newConfig.Path, newConfig.Value)

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
