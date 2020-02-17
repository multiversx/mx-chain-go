package sync

import (
	"encoding/json"
	"errors"
	"github.com/ElrondNetwork/elrond-go/core"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

func createHeaderSyncHandler(retErr bool) update.HeaderSyncHandler {
	meta := &block.MetaBlock{
		Nonce: 1, Epoch: 1, RootHash: []byte("metaRootHash"),
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					ShardId:                 0,
					RootHash:                []byte("shardDataRootHash"),
					PendingMiniBlockHeaders: []block.ShardMiniBlockHeader{{Hash: []byte("hash")}},
					FirstPendingMetaBlock:   []byte("firstPending"),
				},
			},
		},
	}
	args := createMockHeadersSyncHandlerArgs()
	args.StorageService = &mock.ChainStorerMock{GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
		return &mock.StorerStub{
			GetCalled: func(key []byte) (bytes []byte, err error) {
				if retErr {
					return nil, errors.New("err")
				}

				return json.Marshal(meta)
			},
		}
	}}

	if !retErr {
		args.StorageService = initStore()
		byteArray := args.Uint64Converter.ToByteSlice(meta.Nonce)
		_ = args.StorageService.Put(dataRetriever.MetaHdrNonceHashDataUnit, byteArray, []byte("firstPending"))
		marshalledData, _ := json.Marshal(meta)
		_ = args.StorageService.Put(dataRetriever.MetaBlockUnit, []byte("firstPending"), marshalledData)

		_ = args.StorageService.Put(dataRetriever.MetaBlockUnit, []byte(core.EpochStartIdentifier(meta.Epoch)), marshalledData)
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
	require.False(t, syncState.IsInterfaceNil())

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

func TestSyncState_SyncAllStatePendingMiniBlocksErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("err")
	args := ArgsNewSyncState{
		Headers: &mock.HeaderSyncHandlerMock{
			SyncUnFinishedMetaHeadersCalled: func(epoch uint32) error {
				return nil
			},
			GetEpochStartMetaBlockCalled: func() (metaBlock *block.MetaBlock, err error) {
				return &block.MetaBlock{}, nil
			},
		},
		Tries: &mock.EpochStartTriesSyncHandlerMock{},
		MiniBlocks: &mock.EpochStartPendingMiniBlocksSyncHandlerMock{
			SyncPendingMiniBlocksFromMetaCalled: func(meta *block.MetaBlock, unFinished map[string]*block.MetaBlock, waitTime time.Duration) error {
				return localErr
			},
		},
		Transactions: &mock.PendingTransactionsSyncHandlerMock{},
	}

	syncState, err := NewSyncState(args)
	require.Nil(t, err)

	err = syncState.SyncAllState(0)
	require.Equal(t, localErr, err)
}

func TestSyncState_SyncAllStateGetMiniBlocksErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("err")
	args := ArgsNewSyncState{
		Headers: &mock.HeaderSyncHandlerMock{
			SyncUnFinishedMetaHeadersCalled: func(epoch uint32) error {
				return nil
			},
			GetEpochStartMetaBlockCalled: func() (metaBlock *block.MetaBlock, err error) {
				return &block.MetaBlock{}, nil
			},
		},
		Tries: &mock.EpochStartTriesSyncHandlerMock{},
		MiniBlocks: &mock.EpochStartPendingMiniBlocksSyncHandlerMock{
			GetMiniBlocksCalled: func() (m map[string]*block.MiniBlock, err error) {
				return nil, localErr
			},
		},
		Transactions: &mock.PendingTransactionsSyncHandlerMock{},
	}

	syncState, err := NewSyncState(args)
	require.Nil(t, err)

	err = syncState.SyncAllState(0)
	require.Equal(t, localErr, err)
}

func TestSyncState_SyncAllStateSyncTxsErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("err")
	args := ArgsNewSyncState{
		Headers: &mock.HeaderSyncHandlerMock{
			SyncUnFinishedMetaHeadersCalled: func(epoch uint32) error {
				return nil
			},
			GetEpochStartMetaBlockCalled: func() (metaBlock *block.MetaBlock, err error) {
				return &block.MetaBlock{}, nil
			},
		},
		Tries:      &mock.EpochStartTriesSyncHandlerMock{},
		MiniBlocks: &mock.EpochStartPendingMiniBlocksSyncHandlerMock{},
		Transactions: &mock.PendingTransactionsSyncHandlerMock{
			SyncPendingTransactionsForCalled: func(miniBlocks map[string]*block.MiniBlock, epoch uint32, waitTime time.Duration) error {
				return localErr
			},
		},
	}

	syncState, err := NewSyncState(args)
	require.Nil(t, err)

	err = syncState.SyncAllState(0)
	require.Equal(t, localErr, err)
}
