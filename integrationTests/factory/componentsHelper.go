package factory

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime/pprof"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
)

var log = logger.GetOrCreate("integrationtests")

// PrintStack -
func PrintStack() {
	buffer := new(bytes.Buffer)
	err := pprof.Lookup("goroutine").WriteTo(buffer, 2)
	if err != nil {
		log.Debug("could not dump goroutines")
	}

	log.Debug(fmt.Sprintf("\n%s", buffer.String()))
}

// CleanupWorkingDir -
func CleanupWorkingDir() {
	workingDir := WorkingDir
	if _, err := os.Stat(workingDir); !os.IsNotExist(err) {
		err = os.RemoveAll(workingDir)
		if err != nil {
			log.Debug("CleanupWorkingDir", "error", err.Error())
		}
	}
}

// CreateDefaultConfig -
func CreateDefaultConfig() *config.Configs {
	configPathsHolder := createConfigurationsPathsHolder()

	generalConfig, _ := core.LoadMainConfig(configPathsHolder.MainConfig)
	ratingsConfig, _ := core.LoadRatingsConfig(configPathsHolder.Ratings)
	economicsConfig, _ := core.LoadEconomicsConfig(configPathsHolder.Economics)
	prefsConfig, _ := core.LoadPreferencesConfig(configPathsHolder.Preferences)
	p2pConfig, _ := core.LoadP2PConfig(configPathsHolder.P2p)
	externalConfig, _ := core.LoadExternalConfig(configPathsHolder.External)
	systemSCConfig, _ := core.LoadSystemSmartContractsConfig(configPathsHolder.SystemSC)
	epochConfig, _ := core.LoadEpochConfig(configPathsHolder.Epoch)

	p2pConfig.KadDhtPeerDiscovery.Enabled = false
	prefsConfig.Preferences.DestinationShardAsObserver = "0"

	configs := &config.Configs{}
	configs.GeneralConfig = generalConfig
	configs.RatingsConfig = ratingsConfig
	configs.EconomicsConfig = economicsConfig
	configs.SystemSCConfig = systemSCConfig
	configs.PreferencesConfig = prefsConfig
	configs.P2pConfig = p2pConfig
	configs.ExternalConfig = externalConfig
	configs.EpochConfig = epochConfig
	configs.FlagsConfig = &config.ContextFlagsConfig{
		WorkingDir: "workingDir",
		UseLogView: true,
		Version:    Version,
	}
	configs.ConfigurationPathsHolder = configPathsHolder
	configs.ImportDbConfig = &config.ImportDbConfig{}

	return configs
}

func createConfigurationsPathsHolder() *config.ConfigurationPathsHolder {
	var concatPath = func(filename string) string {
		return path.Join(BaseNodeConfigPath, filename)
	}

	return &config.ConfigurationPathsHolder{
		MainConfig:               concatPath(ConfigPath),
		Ratings:                  concatPath(RatingsPath),
		Economics:                concatPath(EconomicsPath),
		Preferences:              concatPath(PrefsPath),
		External:                 concatPath(ExternalPath),
		P2p:                      concatPath(P2pPath),
		Epoch:                    concatPath(EpochPath),
		SystemSC:                 concatPath(SystemSCConfigPath),
		GasScheduleDirectoryName: concatPath(GasSchedule),
		Nodes:                    NodesSetupPath,
		Genesis:                  GenesisPath,
		SmartContracts:           GenesisSmartContracts,
		ValidatorKey:             ValidatorKeyPemPath,
		ApiRoutes:                "",
	}
}
