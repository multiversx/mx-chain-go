package api_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory/api"
	"github.com/multiversx/mx-chain-go/factory/bootstrap"
	"github.com/multiversx/mx-chain-go/factory/mock"
	"github.com/multiversx/mx-chain-go/process/sync/disabled"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestCreateApiResolver(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	coreComponents := componentsMock.GetCoreComponents()
	cryptoComponents := componentsMock.GetCryptoComponents(coreComponents)
	networkComponents := componentsMock.GetNetworkComponents(cryptoComponents)
	dataComponents := componentsMock.GetDataComponents(coreComponents, shardCoordinator)
	stateComponents := componentsMock.GetStateComponents(coreComponents, shardCoordinator)
	processComponents := componentsMock.GetProcessComponents(shardCoordinator, coreComponents, networkComponents, dataComponents, cryptoComponents, stateComponents)
	argsB := componentsMock.GetBootStrapFactoryArgs()

	bcf, _ := bootstrap.NewBootstrapComponentsFactory(argsB)
	mbc, err := bootstrap.NewManagedBootstrapComponents(bcf)
	require.Nil(t, err)
	err = mbc.Create()
	require.Nil(t, err)

	gasSchedule, _ := common.LoadGasScheduleConfig("../../cmd/node/config/gasSchedules/gasScheduleV1.toml")
	economicsConfig := testscommon.GetEconomicsConfig()
	cfg := componentsMock.GetGeneralConfig()
	args := &api.ApiResolverArgs{
		Configs: &config.Configs{
			FlagsConfig: &config.ContextFlagsConfig{
				WorkingDir: "",
			},
			GeneralConfig:   &cfg,
			EpochConfig:     &config.EpochConfig{},
			EconomicsConfig: &economicsConfig,
		},
		CoreComponents:       coreComponents,
		DataComponents:       dataComponents,
		StateComponents:      stateComponents,
		BootstrapComponents:  mbc,
		CryptoComponents:     cryptoComponents,
		ProcessComponents:    processComponents,
		StatusCoreComponents: componentsMock.GetStatusCoreComponents(),
		GasScheduleNotifier: &testscommon.GasScheduleNotifierMock{
			GasSchedule: gasSchedule,
		},
		Bootstrapper:       disabled.NewDisabledBootstrapper(),
		AllowVMQueriesChan: common.GetClosedUnbufferedChannel(),
	}

	apiResolver, err := api.CreateApiResolver(args)
	require.Nil(t, err)
	require.NotNil(t, apiResolver)
}
