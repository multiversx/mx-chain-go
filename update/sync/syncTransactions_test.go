package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	dataTransaction "github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgs() ArgsNewTransactionsSyncer {
	return ArgsNewTransactionsSyncer{
		DataPools: dataRetrieverMock.NewPoolsHolderMock(),
		Storages: &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{}, nil
			},
		},
		Marshalizer:    &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{},
	}
}

func TestNewPendingTransactionsSyncer(t *testing.T) {
	t.Parallel()

	args := createMockArgs()

	pendingTxsSyncer, err := NewTransactionsSyncer(args)
	require.Nil(t, err)
	require.NotNil(t, pendingTxsSyncer)
	require.False(t, pendingTxsSyncer.IsInterfaceNil())
}

func TestNewPendingTransactionsSyncer_NilStorages(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.Storages = nil

	pendingTxsSyncer, err := NewTransactionsSyncer(args)
	require.Nil(t, pendingTxsSyncer)
	require.NotNil(t, dataRetriever.ErrNilHeadersStorage, err)
}

func TestNewPendingTransactionsSyncer_NilDataPools(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.DataPools = nil

	pendingTxsSyncer, err := NewTransactionsSyncer(args)
	require.Nil(t, pendingTxsSyncer)
	require.NotNil(t, dataRetriever.ErrNilDataPoolHolder, err)
}

func TestNewPendingTransactionsSyncer_NilMarshalizer(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.Marshalizer = nil

	pendingTxsSyncer, err := NewTransactionsSyncer(args)
	require.Nil(t, pendingTxsSyncer)
	require.NotNil(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewPendingTransactionsSyncer_NilRequestHandler(t *testing.T) {
	t.Parallel()

	t.Run("TransactionUnit not found", testWithMissingStorer(dataRetriever.TransactionUnit))
	t.Run("UnsignedTransactionUnit not found", testWithMissingStorer(dataRetriever.UnsignedTransactionUnit))
	t.Run("RewardTransactionUnit not found", testWithMissingStorer(dataRetriever.RewardTransactionUnit))
}

func testWithMissingStorer(missingUnit dataRetriever.UnitType) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Storages = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				if unitType == missingUnit {
					return nil, fmt.Errorf("%w for %s", storage.ErrKeyNotFound, missingUnit.String())
				}
				return &storageStubs.StorerStub{}, nil
			},
		}

		pendingTxsSyncer, err := NewTransactionsSyncer(args)
		require.True(t, strings.Contains(err.Error(), storage.ErrKeyNotFound.Error()))
		require.True(t, strings.Contains(err.Error(), missingUnit.String()))
		require.True(t, check.IfNil(pendingTxsSyncer))
	}
}

func TestNewPendingTransactionsSyncer_GetStorerReturnsError(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.RequestHandler = nil

	pendingTxsSyncer, err := NewTransactionsSyncer(args)
	require.Nil(t, pendingTxsSyncer)
	require.NotNil(t, process.ErrNilRequestHandler, err)
}

func TestSyncPendingTransactionsFor(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.Storages = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, err error) {
					tx := &dataTransaction.Transaction{
						Nonce: 1, Value: big.NewInt(10), SndAddr: []byte("snd"), RcvAddr: []byte("rcv"),
					}
					return json.Marshal(tx)
				},
			}, nil
		},
	}

	pendingTxsSyncer, err := NewTransactionsSyncer(args)
	require.Nil(t, err)

	miniBlocks := make(map[string]*block.MiniBlock)
	mb := &block.MiniBlock{TxHashes: [][]byte{[]byte("txHash")}}
	miniBlocks["key"] = mb
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err = pendingTxsSyncer.SyncTransactionsFor(miniBlocks, 1, ctx)
	cancel()
	require.Nil(t, err)
}

func TestSyncPendingTransactionsFor_MissingTxFromPool(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.Storages = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, err error) {
					dummy := 10
					return json.Marshal(dummy)
				},
			}, nil
		},
	}

	pendingTxsSyncer, err := NewTransactionsSyncer(args)
	require.Nil(t, err)

	miniBlocks := make(map[string]*block.MiniBlock)
	mb := &block.MiniBlock{TxHashes: [][]byte{[]byte("txHash")}}
	miniBlocks["key"] = mb

	// we need a value larger than the request interval as to also test what happens after the normal request interval has expired
	timeout := time.Second + time.Millisecond*500
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err = pendingTxsSyncer.SyncTransactionsFor(miniBlocks, 1, ctx)
	cancel()
	require.Equal(t, process.ErrTimeIsOut, err)
}

func TestSyncPendingTransactionsFor_requestTransactionsFor(t *testing.T) {
	t.Parallel()

	mbTxHashes := [][]byte{[]byte("txHash1"), []byte("txHash2")}

	t.Run("request transactions", func(t *testing.T) {
		args := createMockArgs()
		requests := make(map[uint32]int)
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestTransactionHandlerCalled: func(destShardID uint32, txHashes [][]byte) {
				assert.Equal(t, mbTxHashes, txHashes)
				requests[destShardID]++
			},
		}
		pendingTxsSyncer, _ := NewTransactionsSyncer(args)
		mb := &block.MiniBlock{
			Type:            block.TxBlock,
			TxHashes:        mbTxHashes,
			SenderShardID:   0,
			ReceiverShardID: 0,
		}
		numMissing := pendingTxsSyncer.requestTransactionsFor(mb)
		require.Equal(t, 2, numMissing)
		require.Equal(t, 1, len(requests))
		require.Equal(t, 2, requests[0])

		requests = make(map[uint32]int)
		mb.SenderShardID = 1
		numMissing = pendingTxsSyncer.requestTransactionsFor(mb)
		require.Equal(t, 2, numMissing)
		require.Equal(t, 2, len(requests))
		require.Equal(t, 1, requests[0])
		require.Equal(t, 1, requests[1])
	})
	t.Run("request invalid transactions", func(t *testing.T) {
		args := createMockArgs()
		requests := make(map[uint32]int)
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestTransactionHandlerCalled: func(destShardID uint32, txHashes [][]byte) {
				assert.Equal(t, mbTxHashes, txHashes)
				requests[destShardID]++
			},
		}
		pendingTxsSyncer, _ := NewTransactionsSyncer(args)
		mb := &block.MiniBlock{
			Type:            block.InvalidBlock,
			TxHashes:        mbTxHashes,
			SenderShardID:   0,
			ReceiverShardID: 0,
		}
		numMissing := pendingTxsSyncer.requestTransactionsFor(mb)
		require.Equal(t, 2, numMissing)
		require.Equal(t, 1, len(requests))
		require.Equal(t, 2, requests[0])

		requests = make(map[uint32]int)
		mb.SenderShardID = 1
		numMissing = pendingTxsSyncer.requestTransactionsFor(mb)
		require.Equal(t, 2, numMissing)
		require.Equal(t, 2, len(requests))
		require.Equal(t, 1, requests[0])
		require.Equal(t, 1, requests[1])
	})
	t.Run("request unsigned txs", func(t *testing.T) {
		args := createMockArgs()
		requests := make(map[uint32]int)
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestScrHandlerCalled: func(destShardID uint32, txHashes [][]byte) {
				assert.Equal(t, mbTxHashes, txHashes)
				requests[destShardID]++
			},
		}
		pendingTxsSyncer, _ := NewTransactionsSyncer(args)
		mb := &block.MiniBlock{
			Type:            block.SmartContractResultBlock,
			TxHashes:        mbTxHashes,
			SenderShardID:   0,
			ReceiverShardID: 0,
		}
		numMissing := pendingTxsSyncer.requestTransactionsFor(mb)
		require.Equal(t, 2, numMissing)
		require.Equal(t, 1, len(requests))
		require.Equal(t, 2, requests[0])

		requests = make(map[uint32]int)
		mb.SenderShardID = 1
		numMissing = pendingTxsSyncer.requestTransactionsFor(mb)
		require.Equal(t, 2, numMissing)
		require.Equal(t, 2, len(requests))
		require.Equal(t, 1, requests[0])
		require.Equal(t, 1, requests[1])
	})
	t.Run("request rewards txs", func(t *testing.T) {
		args := createMockArgs()
		requests := make(map[uint32]int)
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestRewardTxHandlerCalled: func(destShardID uint32, txHashes [][]byte) {
				assert.Equal(t, mbTxHashes, txHashes)
				requests[destShardID]++
			},
		}
		pendingTxsSyncer, _ := NewTransactionsSyncer(args)
		mb := &block.MiniBlock{
			Type:            block.RewardsBlock,
			TxHashes:        mbTxHashes,
			SenderShardID:   0,
			ReceiverShardID: 0,
		}
		numMissing := pendingTxsSyncer.requestTransactionsFor(mb)
		require.Equal(t, 2, numMissing)
		require.Equal(t, 1, len(requests))
		require.Equal(t, 2, requests[0])

		requests = make(map[uint32]int)
		mb.SenderShardID = 1
		numMissing = pendingTxsSyncer.requestTransactionsFor(mb)
		require.Equal(t, 2, numMissing)
		require.Equal(t, 2, len(requests))
		require.Equal(t, 1, requests[0])
		require.Equal(t, 1, requests[1])
	})
}

func TestSyncPendingTransactionsFor_ReceiveMissingTx(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	args := createMockArgs()
	args.Storages = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, err error) {
					dummy := 10
					return json.Marshal(dummy)
				},
			}, nil
		},
	}

	pendingTxsSyncer, err := NewTransactionsSyncer(args)
	require.Nil(t, err)

	miniBlocks := make(map[string]*block.MiniBlock)
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	miniBlocks["key"] = mb

	go func() {
		time.Sleep(500 * time.Millisecond)

		tx := &dataTransaction.Transaction{
			Nonce: 1, Value: big.NewInt(10), SndAddr: []byte("snd"), RcvAddr: []byte("rcv"),
		}
		pendingTxsSyncer.txPools[block.TxBlock].AddData(txHash, tx, tx.Size(), "0")

		pendingTxsSyncer.receivedTransaction(txHash, tx)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err = pendingTxsSyncer.SyncTransactionsFor(miniBlocks, 1, ctx)
	cancel()
	require.Nil(t, err)
}
