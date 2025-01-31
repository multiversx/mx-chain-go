package runType_test

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"
	mockCoreComp "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	factoryMock "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/require"
)

func TestNewManagedRunTypeComponents(t *testing.T) {
	t.Parallel()

	t.Run("nil factory should error", func(t *testing.T) {
		t.Parallel()

		managedRunTypeComponents, err := runType.NewManagedRunTypeComponents(nil)
		require.Equal(t, errors.ErrNilRunTypeComponentsFactory, err)
		require.Nil(t, managedRunTypeComponents)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		managedRunTypeComponents, err := createComponents()
		require.NoError(t, err)
		require.NotNil(t, managedRunTypeComponents)
	})
}

func TestManagedRunTypeComponents_Create(t *testing.T) {
	t.Parallel()

	t.Run("should work with getters", func(t *testing.T) {
		t.Parallel()

		managedRunTypeComponents, err := createComponents()
		require.NoError(t, err)

		require.Nil(t, managedRunTypeComponents.BlockChainHookHandlerCreator())
		require.Nil(t, managedRunTypeComponents.EpochStartBootstrapperCreator())
		require.Nil(t, managedRunTypeComponents.BootstrapperFromStorageCreator())
		require.Nil(t, managedRunTypeComponents.BootstrapperCreator())
		require.Nil(t, managedRunTypeComponents.BlockProcessorCreator())
		require.Nil(t, managedRunTypeComponents.ForkDetectorCreator())
		require.Nil(t, managedRunTypeComponents.BlockTrackerCreator())
		require.Nil(t, managedRunTypeComponents.RequestHandlerCreator())
		require.Nil(t, managedRunTypeComponents.HeaderValidatorCreator())
		require.Nil(t, managedRunTypeComponents.ScheduledTxsExecutionCreator())
		require.Nil(t, managedRunTypeComponents.TransactionCoordinatorCreator())
		require.Nil(t, managedRunTypeComponents.ValidatorStatisticsProcessorCreator())
		require.Nil(t, managedRunTypeComponents.VMContextCreator())
		require.Nil(t, managedRunTypeComponents.AdditionalStorageServiceCreator())
		require.Nil(t, managedRunTypeComponents.SCProcessorCreator())
		require.Nil(t, managedRunTypeComponents.SCResultsPreProcessorCreator())
		require.Equal(t, consensus.ConsensusModelInvalid, managedRunTypeComponents.ConsensusModel())
		require.Nil(t, managedRunTypeComponents.VmContainerMetaFactoryCreator())
		require.Nil(t, managedRunTypeComponents.VmContainerShardFactoryCreator())
		require.Nil(t, managedRunTypeComponents.AccountsParser())
		require.Nil(t, managedRunTypeComponents.AccountsCreator())
		require.Nil(t, managedRunTypeComponents.OutGoingOperationsPoolHandler())
		require.Nil(t, managedRunTypeComponents.DataCodecHandler())
		require.Nil(t, managedRunTypeComponents.TopicsCheckerHandler())
		require.Nil(t, managedRunTypeComponents.ShardCoordinatorCreator())
		require.Nil(t, managedRunTypeComponents.NodesCoordinatorWithRaterCreator())
		require.Nil(t, managedRunTypeComponents.RequestersContainerFactoryCreator())
		require.Nil(t, managedRunTypeComponents.InterceptorsContainerFactoryCreator())
		require.Nil(t, managedRunTypeComponents.ShardResolversContainerFactoryCreator())
		require.Nil(t, managedRunTypeComponents.TxPreProcessorCreator())
		require.Nil(t, managedRunTypeComponents.ExtraHeaderSigVerifierHolder())
		require.Nil(t, managedRunTypeComponents.GenesisBlockCreatorFactory())
		require.Nil(t, managedRunTypeComponents.GenesisMetaBlockCheckerCreator())
		require.Nil(t, managedRunTypeComponents.NodesSetupCheckerFactory())
		require.Nil(t, managedRunTypeComponents.EpochStartTriggerFactory())
		require.Nil(t, managedRunTypeComponents.LatestDataProviderFactory())
		require.Nil(t, managedRunTypeComponents.StakingToPeerFactory())
		require.Nil(t, managedRunTypeComponents.ValidatorInfoCreatorFactory())
		require.Nil(t, managedRunTypeComponents.ApiProcessorCompsCreatorHandler())
		require.Nil(t, managedRunTypeComponents.EndOfEpochEconomicsFactoryHandler())
		require.Nil(t, managedRunTypeComponents.RewardsCreatorFactory())
		require.Nil(t, managedRunTypeComponents.SystemSCProcessorFactory())
		require.Nil(t, managedRunTypeComponents.PreProcessorsContainerFactoryCreator())
		require.Nil(t, managedRunTypeComponents.DataRetrieverContainersSetter())
		require.Nil(t, managedRunTypeComponents.BroadCastShardMessengerFactoryHandler())
		require.Nil(t, managedRunTypeComponents.ExportHandlerFactoryCreator())
		require.Nil(t, managedRunTypeComponents.ValidatorAccountsSyncerFactoryHandler())
		require.Nil(t, managedRunTypeComponents.ShardRequestersContainerCreatorHandler())
		require.Nil(t, managedRunTypeComponents.APIRewardsTxHandler())
		require.Nil(t, managedRunTypeComponents.OutportDataProviderFactory())
		require.Nil(t, managedRunTypeComponents.DelegatedListFactoryHandler())
		require.Nil(t, managedRunTypeComponents.DirectStakedListFactoryHandler())
		require.Nil(t, managedRunTypeComponents.TotalStakedValueFactoryHandler())

		err = managedRunTypeComponents.Create()
		require.NoError(t, err)
		require.NotNil(t, managedRunTypeComponents.BlockChainHookHandlerCreator())
		require.NotNil(t, managedRunTypeComponents.EpochStartBootstrapperCreator())
		require.NotNil(t, managedRunTypeComponents.BootstrapperFromStorageCreator())
		require.NotNil(t, managedRunTypeComponents.BootstrapperCreator())
		require.NotNil(t, managedRunTypeComponents.BlockProcessorCreator())
		require.NotNil(t, managedRunTypeComponents.ForkDetectorCreator())
		require.NotNil(t, managedRunTypeComponents.BlockTrackerCreator())
		require.NotNil(t, managedRunTypeComponents.RequestHandlerCreator())
		require.NotNil(t, managedRunTypeComponents.HeaderValidatorCreator())
		require.NotNil(t, managedRunTypeComponents.ScheduledTxsExecutionCreator())
		require.NotNil(t, managedRunTypeComponents.TransactionCoordinatorCreator())
		require.NotNil(t, managedRunTypeComponents.ValidatorStatisticsProcessorCreator())
		require.NotNil(t, managedRunTypeComponents.VMContextCreator())
		require.NotNil(t, managedRunTypeComponents.AdditionalStorageServiceCreator())
		require.NotNil(t, managedRunTypeComponents.SCProcessorCreator())
		require.NotNil(t, managedRunTypeComponents.SCResultsPreProcessorCreator())
		require.Equal(t, consensus.ConsensusModelV1, managedRunTypeComponents.ConsensusModel())
		require.NotNil(t, managedRunTypeComponents.VmContainerMetaFactoryCreator())
		require.NotNil(t, managedRunTypeComponents.VmContainerShardFactoryCreator())
		require.NotNil(t, managedRunTypeComponents.AccountsParser())
		require.NotNil(t, managedRunTypeComponents.AccountsCreator())
		require.NotNil(t, managedRunTypeComponents.OutGoingOperationsPoolHandler())
		require.NotNil(t, managedRunTypeComponents.DataCodecHandler())
		require.NotNil(t, managedRunTypeComponents.TopicsCheckerHandler())
		require.NotNil(t, managedRunTypeComponents.ShardCoordinatorCreator())
		require.NotNil(t, managedRunTypeComponents.NodesCoordinatorWithRaterCreator())
		require.NotNil(t, managedRunTypeComponents.RequestersContainerFactoryCreator())
		require.NotNil(t, managedRunTypeComponents.InterceptorsContainerFactoryCreator())
		require.NotNil(t, managedRunTypeComponents.ShardResolversContainerFactoryCreator())
		require.NotNil(t, managedRunTypeComponents.TxPreProcessorCreator())
		require.NotNil(t, managedRunTypeComponents.ExtraHeaderSigVerifierHolder())
		require.NotNil(t, managedRunTypeComponents.GenesisBlockCreatorFactory())
		require.NotNil(t, managedRunTypeComponents.GenesisMetaBlockCheckerCreator())
		require.NotNil(t, managedRunTypeComponents.NodesSetupCheckerFactory())
		require.NotNil(t, managedRunTypeComponents.EpochStartTriggerFactory())
		require.NotNil(t, managedRunTypeComponents.LatestDataProviderFactory())
		require.NotNil(t, managedRunTypeComponents.StakingToPeerFactory())
		require.NotNil(t, managedRunTypeComponents.ValidatorInfoCreatorFactory())
		require.NotNil(t, managedRunTypeComponents.ApiProcessorCompsCreatorHandler())
		require.NotNil(t, managedRunTypeComponents.EndOfEpochEconomicsFactoryHandler())
		require.NotNil(t, managedRunTypeComponents.RewardsCreatorFactory())
		require.NotNil(t, managedRunTypeComponents.SystemSCProcessorFactory())
		require.NotNil(t, managedRunTypeComponents.PreProcessorsContainerFactoryCreator())
		require.NotNil(t, managedRunTypeComponents.DataRetrieverContainersSetter())
		require.NotNil(t, managedRunTypeComponents.BroadCastShardMessengerFactoryHandler())
		require.NotNil(t, managedRunTypeComponents.ExportHandlerFactoryCreator())
		require.NotNil(t, managedRunTypeComponents.ValidatorAccountsSyncerFactoryHandler())
		require.NotNil(t, managedRunTypeComponents.ShardRequestersContainerCreatorHandler())
		require.NotNil(t, managedRunTypeComponents.APIRewardsTxHandler())
		require.NotNil(t, managedRunTypeComponents.OutportDataProviderFactory())
		require.NotNil(t, managedRunTypeComponents.DelegatedListFactoryHandler())
		require.NotNil(t, managedRunTypeComponents.DirectStakedListFactoryHandler())
		require.NotNil(t, managedRunTypeComponents.TotalStakedValueFactoryHandler())

		require.Equal(t, factory.RunTypeComponentsName, managedRunTypeComponents.String())
		require.NoError(t, managedRunTypeComponents.Close())
	})
}

func TestManagedRunTypeComponents_Close(t *testing.T) {
	t.Parallel()

	managedRunTypeComponents, _ := createComponents()
	require.NoError(t, managedRunTypeComponents.Close())

	err := managedRunTypeComponents.Create()
	require.NoError(t, err)

	require.NoError(t, managedRunTypeComponents.Close())
	require.Nil(t, managedRunTypeComponents.BlockChainHookHandlerCreator())
}

func TestManagedRunTypeComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	managedRunTypeComponents, _ := createComponents()
	err := managedRunTypeComponents.CheckSubcomponents()
	require.Equal(t, errors.ErrNilRunTypeComponents, err)

	err = managedRunTypeComponents.Create()
	require.NoError(t, err)

	//TODO check for nil each subcomponent - MX-15371
	err = managedRunTypeComponents.CheckSubcomponents()
	require.NoError(t, err)

	require.NoError(t, managedRunTypeComponents.Close())
}

func TestManagedRunTypeComponents_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var managedRunTypeComponents factory.RunTypeComponentsHandler
	managedRunTypeComponents, _ = runType.NewManagedRunTypeComponents(nil)
	require.True(t, managedRunTypeComponents.IsInterfaceNil())

	managedRunTypeComponents, _ = createComponents()
	require.False(t, managedRunTypeComponents.IsInterfaceNil())
}

func createComponents() (factory.RunTypeComponentsHandler, error) {
	rcf, _ := runType.NewRunTypeComponentsFactory(createArgsRunTypeComponents())
	return runType.NewManagedRunTypeComponents(rcf)
}

func createArgsRunTypeComponents() runType.ArgsRunTypeComponents {
	acc1 := data.InitialAccount{
		Address:      "erd1whq0zspt6ktnv37gqj303da0vygyqwf5q52m7erftd0rl7laygfs6rhpct",
		Supply:       big.NewInt(2),
		Balance:      big.NewInt(1),
		StakingValue: big.NewInt(1),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}

	generalConfig := testscommon.GetGeneralConfig()
	return runType.ArgsRunTypeComponents{
		CoreComponents: &factoryMock.CoreComponentsHolderMock{
			InternalMarshalizerCalled: func() marshal.Marshalizer {
				return &marshallerMock.MarshalizerMock{}
			},
			HasherCalled: func() hashing.Hasher {
				return &hashingMocks.HasherMock{}
			},
			EnableEpochsHandlerCalled: func() common.EnableEpochsHandler {
				return &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
			},
			AddressPubKeyConverterCalled: func() core.PubkeyConverter {
				return &testscommon.PubkeyConverterStub{}
			},
		},
		CryptoComponents: &mockCoreComp.CryptoComponentsStub{
			TxKeyGen: &mockCoreComp.KeyGenMock{},
			BlockSig: &mock.SingleSignerMock{},
		},
		Configs: config.Configs{
			EconomicsConfig: &config.EconomicsConfig{
				GlobalSettings: config.GlobalSettings{
					GenesisTotalSupply:          "2",
					GenesisMintingSenderAddress: "erd17rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rcqqkhty3",
				},
			},
			GeneralConfig: &generalConfig,
		},
		InitialAccounts: []genesis.InitialAccountHandler{&acc1},
	}
}
