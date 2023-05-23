package testscommon

import (
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

// CreateTestConfigs will try to copy the whole configs directory to a temp directory and return the configs after load
// The copying of the configs is required because minor adjustments of their contents is required for the tests to pass
func CreateTestConfigs(tb testing.TB, originalConfigsPath string) *config.Configs {
	tempDir := tb.TempDir()

	newConfigsPath := path.Join(tempDir, "config")

	// TODO refactor this cp to work on all OSes
	cmd := exec.Command("cp", "-r", originalConfigsPath, newConfigsPath)
	err := cmd.Run()
	require.Nil(tb, err)

	newGenesisSmartContractsFilename := path.Join(newConfigsPath, "genesisSmartContracts.json")
	correctTestPathInGenesisSmartContracts(tb, tempDir, newGenesisSmartContractsFilename)

	apiConfig, err := common.LoadApiConfig(path.Join(newConfigsPath, "api.toml"))
	require.Nil(tb, err)

	generalConfig, err := common.LoadMainConfig(path.Join(newConfigsPath, "config.toml"))
	require.Nil(tb, err)

	ratingsConfig, err := common.LoadRatingsConfig(path.Join(newConfigsPath, "ratings.toml"))
	require.Nil(tb, err)

	economicsConfig, err := common.LoadEconomicsConfig(path.Join(newConfigsPath, "economics.toml"))
	require.Nil(tb, err)

	prefsConfig, err := common.LoadPreferencesConfig(path.Join(newConfigsPath, "prefs.toml"))
	require.Nil(tb, err)

	p2pConfig, err := common.LoadP2PConfig(path.Join(newConfigsPath, "p2p.toml"))
	require.Nil(tb, err)

	externalConfig, err := common.LoadExternalConfig(path.Join(newConfigsPath, "external.toml"))
	require.Nil(tb, err)

	systemSCConfig, err := common.LoadSystemSmartContractsConfig(path.Join(newConfigsPath, "systemSmartContractsConfig.toml"))
	require.Nil(tb, err)

	epochConfig, err := common.LoadEpochConfig(path.Join(newConfigsPath, "enableEpochs.toml"))
	require.Nil(tb, err)

	roundConfig, err := common.LoadRoundConfig(path.Join(newConfigsPath, "enableRounds.toml"))
	require.Nil(tb, err)

	// make the node pass the network wait constraints
	p2pConfig.Node.MinNumPeersToWaitForOnBootstrap = 0
	p2pConfig.Node.ThresholdMinConnectedPeers = 0

	return &config.Configs{
		GeneralConfig:     generalConfig,
		ApiRoutesConfig:   apiConfig,
		EconomicsConfig:   economicsConfig,
		SystemSCConfig:    systemSCConfig,
		RatingsConfig:     ratingsConfig,
		PreferencesConfig: prefsConfig,
		ExternalConfig:    externalConfig,
		P2pConfig:         p2pConfig,
		FlagsConfig: &config.ContextFlagsConfig{
			WorkingDir:    tempDir,
			NoKeyProvided: true,
			Version:       "test version",
			DbDir:         path.Join(tempDir, "db"),
		},
		ImportDbConfig: &config.ImportDbConfig{},
		ConfigurationPathsHolder: &config.ConfigurationPathsHolder{
			GasScheduleDirectoryName: path.Join(newConfigsPath, "gasSchedules"),
			Nodes:                    path.Join(newConfigsPath, "nodesSetup.json"),
			Genesis:                  path.Join(newConfigsPath, "genesis.json"),
			SmartContracts:           newGenesisSmartContractsFilename,
			ValidatorKey:             "validatorKey.pem",
		},
		EpochConfig: epochConfig,
		RoundConfig: roundConfig,
	}
}

func correctTestPathInGenesisSmartContracts(tb testing.TB, tempDir string, newGenesisSmartContractsFilename string) {
	input, err := ioutil.ReadFile(newGenesisSmartContractsFilename)
	require.Nil(tb, err)

	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, "./config") {
			lines[i] = strings.Replace(line, "./config", path.Join(tempDir, "config"), 1)
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(newGenesisSmartContractsFilename, []byte(output), 0644)
	require.Nil(tb, err)
}
