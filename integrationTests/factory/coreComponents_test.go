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
	t.Skip()

	generalConfig, _ := core.LoadMainConfig(configPath)
	ratingsConfig, _ := core.LoadRatingsConfig(ratingsPath)
	economicsConfig, _ := core.LoadEconomicsConfig(economicsPath)

	ccf, err := createCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	require.Nil(t, err)
	require.NotNil(t, ccf)
}

func TestCoreComponents_Create_Close_ShouldWork(t *testing.T) {
	t.Skip()

	nrBefore := runtime.NumGoroutine()
	generalConfig, _ := core.LoadMainConfig(configPath)
	ratingsConfig, _ := core.LoadRatingsConfig(ratingsPath)
	economicsConfig, _ := core.LoadEconomicsConfig(economicsPath)

	ccf, _ := createCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	time.Sleep(2 * time.Second)
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

func printStack() {
	stackSlice := make([]byte, 10240)
	s := runtime.Stack(stackSlice, true)
	fmt.Printf("\n%s", stackSlice[0:s])
}
