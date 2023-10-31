package testscommon

import (
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
)

// CreateTestConfigs will try to copy the whole configs directory to a temp directory and return the configs after load
// The copying of the configs is required because minor adjustments of their contents is required for the tests to pass
func CreateTestConfigs(originalConfigsPath string) (*config.Configs, error) {
	tempDir := os.TempDir()

	newConfigsPath := path.Join(tempDir, "config")

	// TODO refactor this cp to work on all OSes
	cmd := exec.Command("cp", "-r", originalConfigsPath, newConfigsPath)
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	newGenesisSmartContractsFilename := path.Join(newConfigsPath, "genesisSmartContracts.json")
	err = correctTestPathInGenesisSmartContracts(tempDir, newGenesisSmartContractsFilename)
	if err != nil {
		return nil, err
	}

	apiConfig, err := common.LoadApiConfig(path.Join(newConfigsPath, "api.toml"))
	if err != nil {
		return nil, err
	}

	generalConfig, err := common.LoadMainConfig(path.Join(newConfigsPath, "config.toml"))
	if err != nil {
		return nil, err
	}

	ratingsConfig, err := common.LoadRatingsConfig(path.Join(newConfigsPath, "ratings.toml"))
	if err != nil {
		return nil, err
	}

	economicsConfig, err := common.LoadEconomicsConfig(path.Join(newConfigsPath, "economics.toml"))
	if err != nil {
		return nil, err
	}

	prefsConfig, err := common.LoadPreferencesConfig(path.Join(newConfigsPath, "prefs.toml"))
	if err != nil {
		return nil, err
	}

	mainP2PConfig, err := common.LoadP2PConfig(path.Join(newConfigsPath, "p2p.toml"))
	if err != nil {
		return nil, err
	}

	fullArchiveP2PConfig, err := common.LoadP2PConfig(path.Join(newConfigsPath, "fullArchiveP2P.toml"))
	if err != nil {
		return nil, err
	}

	externalConfig, err := common.LoadExternalConfig(path.Join(newConfigsPath, "external.toml"))
	if err != nil {
		return nil, err
	}

	systemSCConfig, err := common.LoadSystemSmartContractsConfig(path.Join(newConfigsPath, "systemSmartContractsConfig.toml"))
	if err != nil {
		return nil, err
	}

	epochConfig, err := common.LoadEpochConfig(path.Join(newConfigsPath, "enableEpochs.toml"))
	if err != nil {
		return nil, err
	}

	roundConfig, err := common.LoadRoundConfig(path.Join(newConfigsPath, "enableRounds.toml"))
	if err != nil {
		return nil, err
	}

	// make the node pass the network wait constraints
	mainP2PConfig.Node.MinNumPeersToWaitForOnBootstrap = 0
	mainP2PConfig.Node.ThresholdMinConnectedPeers = 0
	fullArchiveP2PConfig.Node.MinNumPeersToWaitForOnBootstrap = 0
	fullArchiveP2PConfig.Node.ThresholdMinConnectedPeers = 0

	return &config.Configs{
		GeneralConfig:        generalConfig,
		ApiRoutesConfig:      apiConfig,
		EconomicsConfig:      economicsConfig,
		SystemSCConfig:       systemSCConfig,
		RatingsConfig:        ratingsConfig,
		PreferencesConfig:    prefsConfig,
		ExternalConfig:       externalConfig,
		MainP2pConfig:        mainP2PConfig,
		FullArchiveP2pConfig: fullArchiveP2PConfig,
		FlagsConfig: &config.ContextFlagsConfig{
			WorkingDir: tempDir,
			Version:    "test version",
			DbDir:      path.Join(tempDir, "db"),
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
	}, nil
}

func correctTestPathInGenesisSmartContracts(tempDir string, newGenesisSmartContractsFilename string) error {
	input, err := os.ReadFile(newGenesisSmartContractsFilename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, "./config") {
			lines[i] = strings.Replace(line, "./config", path.Join(tempDir, "config"), 1)
		}
	}
	output := strings.Join(lines, "\n")
	return os.WriteFile(newGenesisSmartContractsFilename, []byte(output), 0644)
}
