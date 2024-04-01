package processing_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	dataComp "github.com/multiversx/mx-chain-go/factory/data"
	"github.com/multiversx/mx-chain-go/factory/mock"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	disabledState "github.com/multiversx/mx-chain-go/state/disabled"
	factoryState "github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/disabled"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageManager "github.com/multiversx/mx-chain-go/testscommon/storage"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func Test_newBlockProcessorCreatorForShard(t *testing.T) {
	t.Parallel()

	t.Run("new block processor creator for shard in regular chain should work", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
		pcf, err := processComp.NewProcessComponentsFactory(processArgs)
		require.NoError(t, err)
		require.NotNil(t, pcf)

		_, err = pcf.Create()
		require.NoError(t, err)

		bp, err := pcf.NewBlockProcessor(
			&testscommon.RequestHandlerStub{},
			&processMocks.ForkDetectorStub{},
			&mock.EpochStartTriggerStub{},
			&mock.BoostrapStorerStub{},
			&testscommon.ValidatorStatisticsProcessorStub{},
			&mock.HeaderValidatorStub{},
			&mock.BlockTrackerStub{},
			&mock.PendingMiniBlocksHandlerStub{},
			&sync.RWMutex{},
			&testscommon.ScheduledTxsExecutionStub{},
			&testscommon.ProcessedMiniBlocksTrackerStub{},
			&testscommon.ReceiptsRepositoryStub{},
			&testscommon.BlockProcessingCutoffStub{},
			&testscommon.MissingTrieNodesNotifierStub{},
			&testscommon.SentSignatureTrackerStub{},
		)

		require.NoError(t, err)
		require.Equal(t, "*block.shardProcessor", fmt.Sprintf("%T", bp))
	})

	//t.Run("new block processor creator for shard in sovereign chain should work", func(t *testing.T) {
	//	t.Parallel()
	//
	//	shardCoordinator := sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
	//	args := componentsMock.GetSovereignProcessComponentsFactoryArgs(shardCoordinator)
	//	pcf, err := processComp.NewProcessComponentsFactory(args)
	//	require.NoError(t, err)
	//	require.NotNil(t, pcf)
	//
	//	_, err = pcf.Create()
	//	require.NoError(t, err)
	//
	//	bp, err := pcf.NewBlockProcessor(
	//		&testscommon.ExtendedShardHeaderRequestHandlerStub{},
	//		&processMocks.ForkDetectorStub{},
	//		&mock.EpochStartTriggerStub{},
	//		&mock.BoostrapStorerStub{},
	//		&testscommon.ValidatorStatisticsProcessorStub{},
	//		&mock.HeaderValidatorStub{},
	//		&mock.ExtendedShardHeaderTrackerStub{},
	//		&mock.PendingMiniBlocksHandlerStub{},
	//		&sync.RWMutex{},
	//		&testscommon.ScheduledTxsExecutionStub{},
	//		&testscommon.ProcessedMiniBlocksTrackerStub{},
	//		&testscommon.ReceiptsRepositoryStub{},
	//		&testscommon.BlockProcessingCutoffStub{},
	//		&testscommon.MissingTrieNodesNotifierStub{},
			&testscommon.SentSignatureTrackerStub{},
		)
	//
	//	require.NoError(t, err)
	//	require.Equal(t, "*block.sovereignChainBlockProcessor", fmt.Sprintf("%T", bp))
	//})
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
	storageManagerPeer, _ := trie.CreateTrieStorageManager(storageManagerArgs, storageManager.GetStorageManagerOptions())

	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[dataRetriever.UserAccountsUnit.String()] = storageManagerUser
	trieStorageManagers[dataRetriever.PeerAccountsUnit.String()] = storageManagerPeer

	argsAccCreator := factoryState.ArgsAccountCreator{
		Hasher:              coreComponents.Hasher(),
		Marshaller:          coreComponents.InternalMarshalizer(),
		EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
	}
	accCreator, _ := factoryState.NewAccountCreator(argsAccCreator)

	adb, err := createAccountAdapter(
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		accCreator,
		trieStorageManagers[dataRetriever.UserAccountsUnit.String()],
		coreComponents.EnableEpochsHandler(),
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
					return accounts.NewPeerAccount(address)
				},
			}
		},
		AccountsAdapterCalled: func() state.AccountsAdapter {
			return adb
		},
		AccountsAdapterAPICalled: func() state.AccountsAdapter {
			return adb
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
		&processMocks.ForkDetectorStub{},
		&mock.EpochStartTriggerStub{},
		&mock.BoostrapStorerStub{},
		&testscommon.ValidatorStatisticsProcessorStub{},
		&mock.HeaderValidatorStub{},
		&mock.BlockTrackerStub{},
		&mock.PendingMiniBlocksHandlerStub{},
		&sync.RWMutex{},
		&testscommon.ScheduledTxsExecutionStub{},
		&testscommon.ProcessedMiniBlocksTrackerStub{},
		&testscommon.ReceiptsRepositoryStub{},
		&testscommon.BlockProcessingCutoffStub{},
		&testscommon.MissingTrieNodesNotifierStub{},
		&testscommon.SentSignatureTrackerStub{},
	)

	require.NoError(t, err)
	require.NotNil(t, bp)
}

func createAccountAdapter(
	marshaller marshal.Marshalizer,
	hasher hashing.Hasher,
	accountFactory state.AccountFactory,
	trieStorage common.StorageManager,
	handler common.EnableEpochsHandler,
) (state.AccountsAdapter, error) {
	tr, err := trie.NewTrie(trieStorage, marshaller, hasher, handler, 5)
	if err != nil {
		return nil, err
	}

	args := state.ArgsAccountsDB{
		Trie:                  tr,
		Hasher:                hasher,
		Marshaller:            marshaller,
		AccountFactory:        accountFactory,
		StoragePruningManager: disabled.NewDisabledStoragePruningManager(),
		AddressConverter:      &testscommon.PubkeyConverterMock{},
		SnapshotsManager:      disabledState.NewDisabledSnapshotsManager(),
	}
	adb, err := state.NewAccountsDB(args)
	if err != nil {
		return nil, err
	}

	return adb, nil
}
