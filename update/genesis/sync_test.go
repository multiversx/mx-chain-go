package genesis_test

import (
	"encoding/json"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update/genesis"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func createMockArguments(metaBlock *block.MetaBlock) genesis.ArgsNewSyncState {
	return genesis.ArgsNewSyncState{
		Hasher:           &mock.HasherStub{},
		Marshalizer:      &mock.MarshalizerStub{},
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		TrieSyncers:      &mock.TrieSyncersStub{},
		EpochHandler:     &mock.EpochStartTriggerStub{},
		Storages: &mock.ChainStorerMock{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
				switch unitType {
				case dataRetriever.MetaBlockUnit:
					return &mock.StorerStub{
						GetCalled: func(key []byte) (bytes []byte, err error) {
							return json.Marshal(&metaBlock)
						},
					}
				case dataRetriever.MiniBlockUnit:
					return &mock.StorerStub{
						GetCalled: func(key []byte) (bytes []byte, err error) {
							miniBlock := &block.MiniBlock{
								TxHashes: [][]byte{[]byte("txHash")},
							}
							return json.Marshal(&miniBlock)
						},
					}
				case dataRetriever.TransactionUnit:
					return &mock.StorerStub{
						GetCalled: func(key []byte) (bytes []byte, err error) {
							tx := &transaction.Transaction{Nonce: 1, Value: big.NewInt(0), SndAddr: []byte("addr"), RcvAddr: []byte("addr")}
							return json.Marshal(&tx)
						},
					}
				default:
					return &mock.StorerStub{}
				}

			},
		},
		DataPools:       mock.NewPoolsHolderMock(),
		RequestHandler:  &mock.RequestHandlerStub{},
		HeaderValidator: &mock.HeaderValidatorStub{},
		AccountHandlers: &mock.AccountsHandlerMock{
			GetCalled: func(key string) (adapter state.AccountsAdapter, err error) {
				return &mock.AccountsStub{
					CopyRecreateTrieCalled: func(rootHash []byte) (trie data.Trie, err error) {
						return &mock.TrieStub{
							CommitCalled: func() error {
								return nil
							},
						}, nil
					},
				}, nil
			},
		},
	}
}

func TestNewSyncState(t *testing.T) {
	t.Parallel()

	args := createMockArguments(nil)

	syncState, err := genesis.NewSyncState(args)
	require.NotNil(t, syncState)
	require.Nil(t, err)
}

func TestSyncState_SyncAllState(t *testing.T) {
	t.Parallel()

	metablock := &block.MetaBlock{
		Epoch: 1,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					ShardId:  0,
					RootHash: []byte("rootHash"),
					PendingMiniBlockHeaders: []block.ShardMiniBlockHeader{
						{
							Hash: []byte("mbHash"),
						},
					},
				},
			},
		},
	}
	args := createMockArguments(metablock)
	args.Marshalizer = &mock.MarshalizerFake{}

	syncState, err := genesis.NewSyncState(args)
	require.NotNil(t, syncState)
	require.Nil(t, err)

	err = syncState.SyncAllState(1)
	require.Nil(t, err)
}
