package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/process/sync/disabled"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestCreateApiResolver(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	coreComponents := getCoreComponents()
	coreComponents.StatusHandlerUtils().Metrics()
	networkComponents := getNetworkComponents()
	dataComponents := getDataComponents(coreComponents, shardCoordinator)
	cryptoComponents := getCryptoComponents(coreComponents)
	stateComponents := getStateComponents(coreComponents, shardCoordinator)
	processComponents := getProcessComponents(shardCoordinator, coreComponents, networkComponents, dataComponents, cryptoComponents, stateComponents)
	argsB := getBootStrapArgs()

	bcf, _ := factory.NewBootstrapComponentsFactory(argsB)
	mbc, err := factory.NewManagedBootstrapComponents(bcf)
	require.Nil(t, err)
	err = mbc.Create()
	require.Nil(t, err)

	gasSchedule, _ := common.LoadGasScheduleConfig("../cmd/node/config/gasSchedules/gasScheduleV1.toml")
	economicsConfig := testscommon.GetEconomicsConfig()
	cfg := getGeneralConfig()
	args := &factory.ApiResolverArgs{
		Configs: &config.Configs{
			FlagsConfig: &config.ContextFlagsConfig{
				WorkingDir: "",
			},
			GeneralConfig:   &cfg,
			EpochConfig:     &config.EpochConfig{},
			EconomicsConfig: &economicsConfig,
		},
		CoreComponents:      coreComponents,
		DataComponents:      dataComponents,
		StateComponents:     stateComponents,
		BootstrapComponents: mbc,
		CryptoComponents:    cryptoComponents,
		ProcessComponents:   processComponents,
		GasScheduleNotifier: &mock.GasScheduleNotifierMock{
			GasSchedule: gasSchedule,
		},
		Bootstrapper:       disabled.NewDisabledBootstrapper(),
		AllowVMQueriesChan: common.GetClosedUnbufferedChannel(),
	}

	apiResolver, err := factory.CreateApiResolver(args)
	require.Nil(t, err)
	require.NotNil(t, apiResolver)
}
