package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	factory2 "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/process/txsimulator"
	"github.com/stretchr/testify/require"
)

func Test_newBlockProcessorCreatorForShard(t *testing.T) {
	t.Parallel()

	pcf, _ := factory.NewProcessComponentsFactory(getProcessComponentsArgs())
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
	shardC.ComputeIdCalled = func(_ []byte) uint32 {
		return 0
	}

	dataArgs := getDataArgs(coreComponents)
	dataArgs.ShardCoordinator = shardC
	dataComponentsFactory, _ := factory.NewDataComponentsFactory(dataArgs)
	dataComponents, _ := factory.NewManagedDataComponents(dataComponentsFactory)
	_ = dataComponents.Create()

	networkComponents := getNetworkComponents()
	cryptoComponents := getCryptoComponents(coreComponents)

	stateComp := &mock.StateComponentsHolderStub{
		PeerAccountsCalled: func() state.AccountsAdapter {
			return &mock.AccountsStub{
				LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
					return state.NewPeerAccount([]byte("address"))
				},
				CommitCalled: func() ([]byte, error) {
					return nil, nil
				},
				RootHashCalled: func() ([]byte, error) {
					return []byte("roothash"), nil
				},
			}
		},
		AccountsAdapterCalled: func() state.AccountsAdapter {
			return &mock.AccountsStub{
				LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
					return state.NewUserAccount([]byte("address"))
				},
				CommitCalled: func() ([]byte, error) {
					return nil, nil
				},
				RootHashCalled: func() ([]byte, error) {
					return []byte("roothash"), nil
				},
			}
		},
		TriesContainerCalled: func() state.TriesHolder {
			return &mock.TriesHolderStub{}
		},
		TrieStorageManagersCalled: func() map[string]data.StorageManager {
			mapToRet := map[string]data.StorageManager{
				factory2.UserAccountTrie: &mock.StorageManagerStub{},
				factory2.PeerAccountTrie: &mock.StorageManagerStub{},
			}

			return mapToRet
		},
	}
	args := getProcessArgs(
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComp,
		networkComponents,
	)

	args.ShardCoordinator = shardC

	pcf, _ := factory.NewProcessComponentsFactory(args)
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
