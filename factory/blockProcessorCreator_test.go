package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	factoryState "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	trieFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/txsimulator"
	"github.com/stretchr/testify/require"
)

func Test_newBlockProcessorCreatorForShard(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	pcf, _ := factory.NewProcessComponentsFactory(getProcessComponentsArgs(shardCoordinator))
	require.NotNil(t, pcf)

	_, err := pcf.Create()
	require.NoError(t, err)

	bp, err := pcf.NewBlockProcessor(
		&mock.RequestHandlerStub{},
		&mock.ForkDetectorStub{},
		&mock.EpochStartTriggerStub{},
		&mock.BoostrapStorerStub{},
		&mock.ValidatorStatisticsProcessorStub{},
		&mock.HeaderValidatorStub{},
		&mock.BlockTrackerStub{},
		&mock.PendingMiniBlocksHandlerStub{},
		&txsimulator.ArgsTxSimulator{},
	)

	require.NoError(t, err)
	require.NotNil(t, bp)
}

func Test_newBlockProcessorCreatorForMeta(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
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

	dataArgs := getDataArgs(coreComponents, shardC)
	dataComponentsFactory, _ := factory.NewDataComponentsFactory(dataArgs)
	dataComponents, _ := factory.NewManagedDataComponents(dataComponentsFactory)
	_ = dataComponents.Create()

	networkComponents := getNetworkComponents()
	cryptoComponents := getCryptoComponents(coreComponents)

	memDBMock := mock.NewMemDbMock()
	storageManager, _ := trie.NewTrieStorageManagerWithoutPruning(memDBMock)

	trieStorageManagers := make(map[string]data.StorageManager)
	trieStorageManagers[trieFactory.UserAccountTrie] = storageManager
	trieStorageManagers[trieFactory.PeerAccountTrie] = storageManager

	accounts, err := createAccountAdapter(
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		factoryState.NewAccountCreator(),
		trieStorageManagers[trieFactory.UserAccountTrie],
	)
	require.Nil(t, err)

	stateComp := &mock.StateComponentsHolderStub{
		PeerAccountsCalled: func() state.AccountsAdapter {
			return &mock.AccountsStub{
				RootHashCalled: func() ([]byte, error) {
					return make([]byte, 0), nil
				},
				CommitCalled: func() ([]byte, error) {
					return make([]byte, 0), nil
				},
				SaveAccountCalled: func(account state.AccountHandler) error {
					return nil
				},
				LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
					return state.NewEmptyPeerAccount(), nil
				},
			}
		},
		AccountsAdapterCalled: func() state.AccountsAdapter {
			return accounts
		},
		TriesContainerCalled: func() state.TriesHolder {
			return &mock.TriesHolderStub{}
		},
		TrieStorageManagersCalled: func() map[string]data.StorageManager {
			return trieStorageManagers
		},
	}
	args := getProcessArgs(
		shardC,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComp,
		networkComponents,
	)

	factory.SetShardCoordinator(shardC, args.BootstrapComponents)

	pcf, _ := factory.NewProcessComponentsFactory(args)
	require.NotNil(t, pcf)

	_, err = pcf.Create()
	require.NoError(t, err)

	bp, err := pcf.NewBlockProcessor(
		&mock.RequestHandlerStub{},
		&mock.ForkDetectorStub{},
		&mock.EpochStartTriggerStub{},
		&mock.BoostrapStorerStub{},
		&mock.ValidatorStatisticsProcessorStub{},
		&mock.HeaderValidatorStub{},
		&mock.BlockTrackerStub{},
		&mock.PendingMiniBlocksHandlerStub{},
		&txsimulator.ArgsTxSimulator{},
	)

	require.NoError(t, err)
	require.NotNil(t, bp)
}

func createAccountAdapter(
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	accountFactory state.AccountFactory,
	trieStorage data.StorageManager,
) (state.AccountsAdapter, error) {
	tr, err := trie.NewTrie(trieStorage, marshalizer, hasher, 5)
	if err != nil {
		return nil, err
	}

	adb, err := state.NewAccountsDB(tr, hasher, marshalizer, accountFactory)
	if err != nil {
		return nil, err
	}

	return adb, nil
}
