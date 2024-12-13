package processing_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/factory/mock"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	disabledState "github.com/multiversx/mx-chain-go/state/disabled"
	factoryState "github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/disabled"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
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
		processArgs := createProcessFactoryArgs(t, shardCoordinator)
		pcf, err := processComp.NewProcessComponentsFactory(processArgs)
		require.NoError(t, err)
		require.NotNil(t, pcf)

		_, err = pcf.Create()
		require.NoError(t, err)

		bp, epochStartSCProc, err := pcf.NewBlockProcessor(
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
		require.Equal(t, "*disabled.epochStartSystemSCProcessor", fmt.Sprintf("%T", epochStartSCProc))
	})

	t.Run("new block processor creator for shard in sovereign chain should work", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := sharding.NewSovereignShardCoordinator()
		processArgs := createSovereignProcessFactoryArgs(t, shardCoordinator)
		pcf, err := processComp.NewProcessComponentsFactory(processArgs)
		require.NoError(t, err)
		require.NotNil(t, pcf)

		_, err = pcf.Create()
		require.NoError(t, err)

		bp, epochStartSCProc, err := pcf.NewBlockProcessor(
			&testscommon.ExtendedShardHeaderRequestHandlerStub{},
			&processMocks.ForkDetectorStub{},
			&mock.EpochStartTriggerStub{},
			&mock.BoostrapStorerStub{},
			&testscommon.ValidatorStatisticsProcessorStub{},
			&mock.HeaderValidatorStub{},
			&testscommon.ExtendedShardHeaderTrackerStub{},
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
		require.Equal(t, "*block.sovereignChainBlockProcessor", fmt.Sprintf("%T", bp))
		require.Equal(t, "*disabled.epochStartSystemSCProcessor", fmt.Sprintf("%T", epochStartSCProc))
	})
}

func Test_newBlockProcessorCreatorForMeta(t *testing.T) {
	t.Parallel()

	cfg := testscommon.GetGeneralConfig()
	coreComponents := componentsMock.GetCoreComponents(cfg, componentsMock.GetRunTypeCoreComponents())
	cryptoComponents := componentsMock.GetCryptoComponents(coreComponents)
	runTypeComponents := componentsMock.GetRunTypeComponents(coreComponents, cryptoComponents)
	networkComponents := componentsMock.GetNetworkComponents(cryptoComponents)
	statusCoreComponents := componentsMock.GetStatusCoreComponents(cfg, coreComponents)

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

	stateComponents := &mock.StateComponentsHolderStub{
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

	bootstrapComponents := componentsMock.GetBootstrapComponents(cfg, statusCoreComponents, coreComponents, cryptoComponents, networkComponents, runTypeComponents)
	componentsMock.SetShardCoordinator(t, bootstrapComponents, shardC)
	statusComponents := componentsMock.GetStatusComponents(cfg, statusCoreComponents, coreComponents, networkComponents, bootstrapComponents, stateComponents, &shardingMocks.NodesCoordinatorMock{}, cryptoComponents)
	dataComponents := componentsMock.GetDataComponents(cfg, statusCoreComponents, coreComponents, bootstrapComponents, cryptoComponents, runTypeComponents)

	processArgs := componentsMock.GetProcessFactoryArgs(cfg, runTypeComponents, coreComponents, cryptoComponents, networkComponents, bootstrapComponents, stateComponents, dataComponents, statusComponents, statusCoreComponents)
	pcf, _ := processComp.NewProcessComponentsFactory(processArgs)
	require.NotNil(t, pcf)

	_, err = pcf.Create()
	require.NoError(t, err)

	bp, epochStartSCProc, err := pcf.NewBlockProcessor(
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
	require.Equal(t, "*metachain.systemSCProcessor", fmt.Sprintf("%T", epochStartSCProc))
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
