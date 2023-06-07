package factory

import (
	"bytes"
	"fmt"
	"path"
	"runtime/pprof"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/p2p"
	logger "github.com/multiversx/mx-chain-logger-go"
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

// CreateDefaultConfig -
func CreateDefaultConfig(tb testing.TB) *config.Configs {
	configPathsHolder := createConfigurationsPathsHolder()

	generalConfig, _ := common.LoadMainConfig(configPathsHolder.MainConfig)
	ratingsConfig, _ := common.LoadRatingsConfig(configPathsHolder.Ratings)
	economicsConfig, _ := common.LoadEconomicsConfig(configPathsHolder.Economics)
	prefsConfig, _ := common.LoadPreferencesConfig(configPathsHolder.Preferences)
	p2pConfig, _ := common.LoadP2PConfig(configPathsHolder.P2p)
	fullArchiveP2PConfig, _ := common.LoadP2PConfig(configPathsHolder.FullArchiveP2p)
	externalConfig, _ := common.LoadExternalConfig(configPathsHolder.External)
	systemSCConfig, _ := common.LoadSystemSmartContractsConfig(configPathsHolder.SystemSC)
	epochConfig, _ := common.LoadEpochConfig(configPathsHolder.Epoch)
	roundConfig, _ := common.LoadRoundConfig(configPathsHolder.RoundActivation)

	p2pConfig.KadDhtPeerDiscovery.Enabled = false
	prefsConfig.Preferences.DestinationShardAsObserver = "0"
	prefsConfig.Preferences.ConnectionWatcherType = p2p.ConnectionWatcherTypePrint

	configs := &config.Configs{}
	configs.GeneralConfig = generalConfig
	configs.RatingsConfig = ratingsConfig
	configs.EconomicsConfig = economicsConfig
	configs.SystemSCConfig = systemSCConfig
	configs.PreferencesConfig = prefsConfig
	configs.MainP2pConfig = p2pConfig
	configs.FullArchiveP2pConfig = fullArchiveP2PConfig
	configs.ExternalConfig = externalConfig
	configs.EpochConfig = epochConfig
	configs.RoundConfig = roundConfig
	configs.FlagsConfig = &config.ContextFlagsConfig{
		WorkingDir:  tb.TempDir(),
		DbDir:       "dbDir",
		LogsDir:     "logsDir",
		UseLogView:  true,
		BaseVersion: BaseVersion,
		Version:     Version,
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
		FullArchiveP2p:           concatPath(FullArchiveP2pPath),
		Epoch:                    concatPath(EpochPath),
		SystemSC:                 concatPath(SystemSCConfigPath),
		GasScheduleDirectoryName: concatPath(GasSchedule),
		RoundActivation:          concatPath(RoundActivationPath),
		Nodes:                    NodesSetupPath,
		Genesis:                  GenesisPath,
		SmartContracts:           GenesisSmartContracts,
		ValidatorKey:             ValidatorKeyPemPath,
		ApiRoutes:                "",
		P2pKey:                   P2pKeyPath,
	}
}
