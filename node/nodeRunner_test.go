//go:build !race
// +build !race

package node

import (
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createConfigs(tb testing.TB) *config.Configs {
	tempDir := tb.TempDir()

	originalConfigsPath := "../cmd/node/config"
	newConfigsPath := path.Join(tempDir, "config")

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

func TestNewNodeRunner(t *testing.T) {
	t.Parallel()

	t.Run("nil configs should error", func(t *testing.T) {
		t.Parallel()

		expectedErrorString := "nil configs provided"
		runner, err := NewNodeRunner(nil)
		assert.Nil(t, runner)
		assert.Equal(t, expectedErrorString, err.Error())
	})
	t.Run("with valid configs should work", func(t *testing.T) {
		t.Parallel()

		configs := createConfigs(t)
		runner, err := NewNodeRunner(configs)
		assert.NotNil(t, runner)
		assert.Nil(t, err)
	})
}

type applicationRunningTrigger struct {
	chanClose chan struct{}
}

func newApplicationRunningTrigger() *applicationRunningTrigger {
	return &applicationRunningTrigger{
		chanClose: make(chan struct{}),
	}
}

func (trigger *applicationRunningTrigger) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), "application is now running") {
		log.Info("got signal, trying to gracefully close the node")
		close(trigger.chanClose)
	}

	return 0, nil
}

func TestNodeRunner_StartAndCloseNode(t *testing.T) {
	t.Parallel()

	configs := createConfigs(t)
	runner, _ := NewNodeRunner(configs)

	trigger := newApplicationRunningTrigger()
	err := logger.AddLogObserver(trigger, &logger.PlainFormatter{})
	require.Nil(t, err)

	// start a go routine that will send the SIGINT message after 1 minute
	go func() {
		timeout := time.Minute * 5
		select {
		case <-trigger.chanClose:
		case <-time.After(timeout):
			require.Fail(t, "timeout waiting for application to start")
		}
		time.Sleep(time.Second)

		log.Info("sending SIGINT to self")
		errKill := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		assert.Nil(t, errKill)
	}()

	err = runner.Start()
	assert.Nil(t, err)
}
