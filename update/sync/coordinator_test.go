package sync

import (
	"encoding/json"
	"errors"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func createHeaderSyncHandler(retErr bool) update.HeaderSyncHandler {
	meta := &block.MetaBlock{
		Nonce: 1, Epoch: 1, RootHash: []byte("metaRootHash"),
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardId: 0, RootHash: []byte("shardDataRootHash"),
					PendingMiniBlockHeaders: []block.ShardMiniBlockHeader{
						{Hash: []byte("hash")},
					},
				},
			},
		},
	}
	args := createMockHeadersSuncHandlerArgs()
	args.Storage = &mock.StorerStub{
		GetCalled: func(key []byte) (bytes []byte, err error) {
			if retErr {
				return nil, errors.New("err")
			}

			return json.Marshal(meta)
		},
	}

	headersSyncHandler, _ := NewHeadersSyncHandler(args)
	return headersSyncHandler
}

func createPendingMiniBlocksSyncHandler() update.EpochStartPendingMiniBlocksSyncHandler {
	txHash := []byte("txHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	args := ArgsNewPendingMiniBlocksSyncer{
		Storage: &mock.StorerStub{},
		Cache: &mock.CacherStub{
			RegisterHandlerCalled: func(f func(key []byte)) {},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return mb, true
			},
		},
		Marshalizer:    &mock.MarshalizerFake{},
		RequestHandler: &mock.RequestHandlerStub{},
	}

	pendingMiniBlocksSyncer, _ := NewPendingMiniBlocksSyncer(args)
	return pendingMiniBlocksSyncer
}

func createPendingTxSyncHandler() update.PendingTransactionsSyncHandler {
	args := createMockArgs()
	args.Storages = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, err error) {
					tx := &dataTransaction.Transaction{
						Nonce: 1, Value: big.NewInt(10), SndAddr: []byte("snd"), RcvAddr: []byte("rcv"),
					}
					return json.Marshal(tx)
				},
			}
		},
	}

	pendingTxsSyncer, _ := NewPendingTransactionsSyncer(args)
	return pendingTxsSyncer
}

func createSyncTrieState(retErr bool) update.EpochStartTriesSyncHandler {
	args := ArgsNewSyncTriesHandler{
		TrieSyncers: &mock.TrieSyncersStub{
			GetCalled: func(key string) (syncer update.TrieSyncer, err error) {
				if retErr {
					return nil, errors.New("err")
				}
				return &mock.TrieSyncersStub{}, nil
			},
		},
		ActiveTries: &mock.TriesHolderMock{
			GetCalled: func(bytes []byte) data.Trie {
				return &mock.TrieStub{
					RecreateCalled: func(root []byte) (trie data.Trie, err error) {
						return &mock.TrieStub{
							CommitCalled: func() error {
								if retErr {
									return errors.New("err")
								}
								return nil
							},
						}, nil
					},
				}
			},
		},
	}

	triesSyncHandler, _ := NewSyncTriesHandler(args)
	return triesSyncHandler
}

func TestNewSyncState_Ok(t *testing.T) {
	t.Parallel()

	args := ArgsNewSyncState{
		Headers:      createHeaderSyncHandler(false),
		Tries:        createSyncTrieState(false),
		MiniBlocks:   createPendingMiniBlocksSyncHandler(),
		Transactions: createPendingTxSyncHandler(),
	}

	syncState, err := NewSyncState(args)
	require.Nil(t, err)

	err = syncState.SyncAllState(1)
	require.Nil(t, err)
}

func TestNewSyncState_CannotSyncHeaderErr(t *testing.T) {
	t.Parallel()

	args := ArgsNewSyncState{
		Headers:      createHeaderSyncHandler(true),
		Tries:        createSyncTrieState(false),
		MiniBlocks:   createPendingMiniBlocksSyncHandler(),
		Transactions: createPendingTxSyncHandler(),
	}

	syncState, err := NewSyncState(args)
	require.Nil(t, err)

	err = syncState.SyncAllState(1)
	require.NotNil(t, err)
}

func TestNewSyncState_CannotSyncTriesErr(t *testing.T) {
	t.Parallel()

	args := ArgsNewSyncState{
		Headers:      createHeaderSyncHandler(false),
		Tries:        createSyncTrieState(true),
		MiniBlocks:   createPendingMiniBlocksSyncHandler(),
		Transactions: createPendingTxSyncHandler(),
	}

	syncState, err := NewSyncState(args)
	require.Nil(t, err)

	err = syncState.SyncAllState(1)
	require.NotNil(t, err)
}
