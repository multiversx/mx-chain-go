package factory

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/stretchr/testify/require"
)

// ------------ Test CoreComponents --------------------
func TestCoreComponents_Create_ShouldWork(t *testing.T) {
	//t.Parallel()

	generalConfig, _ := loadMainConfig(configPath)
	ratingsConfig, _ := loadRatingsConfig(ratingsPath)
	economicsConfig, _ := loadEconomicsConfig(economicsPath)

	ccf, err := createCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	require.Nil(t, err)
	require.NotNil(t, ccf)
}

func TestCoreComponents_Create_Close_ShouldWork(t *testing.T) {
	//t.Parallel()

	nrBefore := runtime.NumGoroutine()
	generalConfig, _ := loadMainConfig(configPath)
	ratingsConfig, _ := loadRatingsConfig(ratingsPath)
	economicsConfig, _ := loadEconomicsConfig(economicsPath)

	ccf, _ := createCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	err := ccf.Close()
	time.Sleep(2 * time.Second)
	nrAfter := runtime.NumGoroutine()

	require.Nil(t, err)
	if nrBefore != nrAfter {
		printStack()
	}

	require.Equal(t, nrBefore, nrAfter)
}

func createCoreComponents(
	generalConfig config.Config,
	ratingsConfig config.RatingsConfig,
	economicsConfig config.EconomicsConfig) (mainFactory.CoreComponentsHandler, error) {
	chanCreateViews := make(chan struct{}, 1)
	chanLogRewrite := make(chan struct{}, 1)

	statusHandlersFactoryArgs := &factory.StatusHandlersFactoryArgs{
		UseTermUI:      false,
		ChanStartViews: chanCreateViews,
		ChanLogRewrite: chanLogRewrite,
	}

	statusHandlersFactory, err := factory.NewStatusHandlersFactory(statusHandlersFactoryArgs)
	if err != nil {
		return nil, err
	}

	chanStopNodeProcess := make(chan endProcess.ArgEndProcess, 1)

	coreArgs := mainFactory.CoreComponentsFactoryArgs{
		Config:                generalConfig,
		RatingsConfig:         ratingsConfig,
		EconomicsConfig:       economicsConfig,
		NodesFilename:         nodesSetupPath,
		WorkingDirectory:      "workingDir",
		ChanStopNodeProcess:   chanStopNodeProcess,
		StatusHandlersFactory: statusHandlersFactory,
	}

	coreComponentsFactory, err := mainFactory.NewCoreComponentsFactory(coreArgs)
	if err != nil {
		return nil, fmt.Errorf("NewCoreComponentsFactory failed: %w", err)
	}

	managedCoreComponents, err := mainFactory.NewManagedCoreComponents(coreComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedCoreComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedCoreComponents, nil
}

func loadMainConfig(filepath string) (*config.Config, error) {
	cfg := &config.Config{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadApiConfig(filepath string) (*config.ApiRoutesConfig, error) {
	cfg := &config.ApiRoutesConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadEconomicsConfig(filepath string) (*config.EconomicsConfig, error) {
	cfg := &config.EconomicsConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadSystemSmartContractsConfig(filepath string) (*config.SystemSmartContractsConfig, error) {
	cfg := &config.SystemSmartContractsConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadRatingsConfig(filepath string) (*config.RatingsConfig, error) {
	cfg := &config.RatingsConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return &config.RatingsConfig{}, err
	}

	return cfg, nil
}

func loadPreferencesConfig(filepath string) (*config.Preferences, error) {
	cfg := &config.Preferences{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadExternalConfig(filepath string) (*config.ExternalConfig, error) {
	cfg := &config.ExternalConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, fmt.Errorf("cannot load external config: %w", err)
	}

	return cfg, nil
}

func printStack() {
	stackSlice := make([]byte, 10240)
	s := runtime.Stack(stackSlice, true)
	fmt.Printf("\n%s", stackSlice[0:s])
}
