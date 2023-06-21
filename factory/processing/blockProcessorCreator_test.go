package processing_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	customErrors "github.com/multiversx/mx-chain-go/errors"
	dataComp "github.com/multiversx/mx-chain-go/factory/data"
	"github.com/multiversx/mx-chain-go/factory/mock"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/state"
	factoryState "github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/disabled"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageManager "github.com/multiversx/mx-chain-go/testscommon/storage"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func Test_newBlockProcessorCreatorForShard(t *testing.T) {
	t.Parallel()

	t.Run("new block processor creator for shard in regular chain should work", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		pcf, err := processComp.NewProcessComponentsFactory(componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator))
		require.NoError(t, err)
		require.NotNil(t, pcf)

		_, err = pcf.Create()
		require.NoError(t, err)

		bp, err := pcf.NewBlockProcessor(
			&testscommon.RequestHandlerStub{},
			&mock.ForkDetectorStub{},
			&mock.EpochStartTriggerStub{},
			&mock.BoostrapStorerStub{},
			&mock.ValidatorStatisticsProcessorStub{},
			&mock.HeaderValidatorStub{},
			&mock.BlockTrackerStub{},
			&mock.PendingMiniBlocksHandlerStub{},
			&sync.RWMutex{},
			&testscommon.ScheduledTxsExecutionStub{},
			&testscommon.ProcessedMiniBlocksTrackerStub{},
			&testscommon.ReceiptsRepositoryStub{},
			&testscommon.BlockProcessingCutoffStub{},
			&testscommon.MissingTrieNodesNotifierStub{})

		require.NoError(t, err)
		require.Equal(t, "*block.shardProcessor", fmt.Sprintf("%T", bp))
	})

	t.Run("new block processor creator for shard in sovereign chain should work", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		args := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.NoError(t, err)
		require.NotNil(t, pcf)

		_, err = pcf.Create()
		require.NoError(t, err)

		bp, err := pcf.NewBlockProcessor(
			&testscommon.ExtendedShardHeaderRequestHandlerStub{},
			&mock.ForkDetectorStub{},
			&mock.EpochStartTriggerStub{},
			&mock.BoostrapStorerStub{},
			&mock.ValidatorStatisticsProcessorStub{},
			&mock.HeaderValidatorStub{},
			&mock.ExtendedShardHeaderTrackerStub{},
			&mock.PendingMiniBlocksHandlerStub{},
			&sync.RWMutex{},
			&testscommon.ScheduledTxsExecutionStub{},
			&testscommon.ProcessedMiniBlocksTrackerStub{},
			&testscommon.ReceiptsRepositoryStub{},
			&testscommon.BlockProcessingCutoffStub{},
			&testscommon.MissingTrieNodesNotifierStub{})

		require.NoError(t, err)
		require.Equal(t, "*block.sovereignChainBlockProcessor", fmt.Sprintf("%T", bp))
	})

	t.Run("invalid chain id, should return error", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		args := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
		pcf, err := processComp.NewProcessComponentsFactory(args)
		require.NoError(t, err)
		require.NotNil(t, pcf)

		_, err = pcf.Create()
		require.NoError(t, err)

		bp, err := pcf.NewBlockProcessor(
			&testscommon.ExtendedShardHeaderRequestHandlerStub{},
			&mock.ForkDetectorStub{},
			&mock.EpochStartTriggerStub{},
			&mock.BoostrapStorerStub{},
			&mock.ValidatorStatisticsProcessorStub{},
			&mock.HeaderValidatorStub{},
			&mock.ExtendedShardHeaderTrackerStub{},
			&mock.PendingMiniBlocksHandlerStub{},
			&sync.RWMutex{},
			&testscommon.ScheduledTxsExecutionStub{},
			&testscommon.ProcessedMiniBlocksTrackerStub{},
			&testscommon.ReceiptsRepositoryStub{},
			&testscommon.BlockProcessingCutoffStub{},
			&testscommon.MissingTrieNodesNotifierStub{})

		require.NotNil(t, err)
		require.Nil(t, bp)
		require.True(t, strings.Contains(err.Error(), customErrors.ErrUnimplementedChainRunType.Error()))
		require.True(t, strings.Contains(err.Error(), "invalid"))
	})
}

func Test_newBlockProcessorCreatorForMeta(t *testing.T) {
	t.Parallel()

	coreComponents := componentsMock.GetCoreComponents()
	shardC := mock.NewMultiShardsCoordinatorMock(1)
	shardC.SelfIDCalled = func() uint32 {
		return core.MetachainShardId
	}
	shardC.ComputeIdCalled = func(address []byte) uint32 {
		if core.IsSmartContractOnMetachain(address[len(address)-1:], address) {
			return core.MetachainShardId
		}

		return 0
	}
	shardC.CurrentShard = core.MetachainShardId

	dataArgs := componentsMock.GetDataArgs(coreComponents, shardC)
	dataComponentsFactory, _ := dataComp.NewDataComponentsFactory(dataArgs)
	dataComponents, _ := dataComp.NewManagedDataComponents(dataComponentsFactory)
	_ = dataComponents.Create()

	cryptoComponents := componentsMock.GetCryptoComponents(coreComponents)
	networkComponents := componentsMock.GetNetworkComponents(cryptoComponents)

	storageManagerArgs := storageManager.GetStorageManagerArgs()
	storageManagerArgs.Marshalizer = coreComponents.InternalMarshalizer()
	storageManagerArgs.Hasher = coreComponents.Hasher()
	storageManagerUser, _ := trie.CreateTrieStorageManager(storageManagerArgs, storageManager.GetStorageManagerOptions())

	storageManagerArgs.MainStorer = mock.NewMemDbMock()
	storageManagerArgs.CheckpointsStorer = mock.NewMemDbMock()
	storageManagerPeer, _ := trie.CreateTrieStorageManager(storageManagerArgs, storageManager.GetStorageManagerOptions())

	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[dataRetriever.UserAccountsUnit.String()] = storageManagerUser
	trieStorageManagers[dataRetriever.PeerAccountsUnit.String()] = storageManagerPeer

	accounts, err := createAccountAdapter(
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		factoryState.NewAccountCreator(),
		trieStorageManagers[dataRetriever.UserAccountsUnit.String()],
	)
	require.Nil(t, err)

	stateComp := &mock.StateComponentsHolderStub{
		PeerAccountsCalled: func() state.AccountsAdapter {
			return &stateMock.AccountsStub{
				RootHashCalled: func() ([]byte, error) {
					return make([]byte, 0), nil
				},
				CommitCalled: func() ([]byte, error) {
					return make([]byte, 0), nil
				},
				SaveAccountCalled: func(account vmcommon.AccountHandler) error {
					return nil
				},
				LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
					return state.NewEmptyPeerAccount(), nil
				},
			}
		},
		AccountsAdapterCalled: func() state.AccountsAdapter {
			return accounts
		},
		AccountsAdapterAPICalled: func() state.AccountsAdapter {
			return accounts
		},
		TriesContainerCalled: func() common.TriesHolder {
			return &trieMock.TriesHolderStub{
				GetCalled: func(bytes []byte) common.Trie {
					return &trieMock.TrieStub{}
				},
			}
		},
		TrieStorageManagersCalled: func() map[string]common.StorageManager {
			return trieStorageManagers
		},
	}
	args := componentsMock.GetProcessArgs(
		shardC,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComp,
		networkComponents,
	)

	componentsMock.SetShardCoordinator(t, args.BootstrapComponents, shardC)

	pcf, _ := processComp.NewProcessComponentsFactory(args)
	require.NotNil(t, pcf)

	_, err = pcf.Create()
	require.NoError(t, err)

	bp, err := pcf.NewBlockProcessor(
		&testscommon.RequestHandlerStub{},
		&mock.ForkDetectorStub{},
		&mock.EpochStartTriggerStub{},
		&mock.BoostrapStorerStub{},
		&mock.ValidatorStatisticsProcessorStub{},
		&mock.HeaderValidatorStub{},
		&mock.BlockTrackerStub{},
		&mock.PendingMiniBlocksHandlerStub{},
		&sync.RWMutex{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&testscommon.ReceiptsRepositoryStub{},
		&testscommon.BlockProcessingCutoffStub{},
		&testscommon.MissingTrieNodesNotifierStub{},
	)

	require.NoError(t, err)
	require.NotNil(t, bp)
}

func createAccountAdapter(
	marshaller marshal.Marshalizer,
	hasher hashing.Hasher,
	accountFactory state.AccountFactory,
	trieStorage common.StorageManager,
) (state.AccountsAdapter, error) {
	tr, err := trie.NewTrie(trieStorage, marshaller, hasher, 5)
	if err != nil {
		return nil, err
	}

	args := state.ArgsAccountsDB{
		Trie:                  tr,
		Hasher:                hasher,
		Marshaller:            marshaller,
		AccountFactory:        accountFactory,
		StoragePruningManager: disabled.NewDisabledStoragePruningManager(),
		ProcessingMode:        common.Normal,
		ProcessStatusHandler:  &testscommon.ProcessStatusHandlerStub{},
		AppStatusHandler:      &statusHandler.AppStatusHandlerStub{},
		AddressConverter:      &testscommon.PubkeyConverterMock{},
	}
	adb, err := state.NewAccountsDB(args)
	if err != nil {
		return nil, err
	}

	return adb, nil
}
