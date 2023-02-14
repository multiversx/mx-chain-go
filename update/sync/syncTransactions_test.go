package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	dataTransaction "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/mock"
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
		Marshaller:     &mock.MarshalizerFake{},
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
	args.Marshaller = nil

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
		require.NotNil(t, err)
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

func TestTransactionsSync_RequestTransactionsForPeerMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	svi1 := &state.ShardValidatorInfo{
		PublicKey: []byte("x"),
	}
	svi2 := &state.ShardValidatorInfo{
		PublicKey: []byte("y"),
	}

	args := createMockArgs()
	args.DataPools = getDataPoolsWithShardValidatorInfoAndTxHash(svi2, []byte("b"))
	transactionsSyncer, _ := NewTransactionsSyncer(args)

	miniBlock := &block.MiniBlock{
		TxHashes: [][]byte{
			[]byte("a"),
			[]byte("b"),
			[]byte("c"),
		},
	}
	transactionsSyncer.mutPendingTx.Lock()
	transactionsSyncer.mapValidatorsInfo["a"] = svi1
	transactionsSyncer.mapTxsToMiniBlocks["b"] = miniBlock
	numMissingValidatorsInfo := transactionsSyncer.requestTransactionsForPeerMiniBlock(miniBlock)

	assert.Equal(t, 1, numMissingValidatorsInfo)
	assert.Equal(t, 2, len(transactionsSyncer.mapValidatorsInfo))
	assert.Equal(t, 2, len(transactionsSyncer.mapTxsToMiniBlocks))
	assert.Equal(t, svi2, transactionsSyncer.mapValidatorsInfo["b"])
	assert.Equal(t, miniBlock, transactionsSyncer.mapTxsToMiniBlocks["c"])
	transactionsSyncer.mutPendingTx.Unlock()
}

func TestTransactionsSync_ReceivedValidatorInfo(t *testing.T) {
	t.Parallel()

	txHash := []byte("hash")
	svi := &state.ShardValidatorInfo{
		PublicKey: []byte("x"),
	}

	args := createMockArgs()
	transactionsSyncer, _ := NewTransactionsSyncer(args)

	// stop sync is true
	transactionsSyncer.receivedValidatorInfo(txHash, svi)
	transactionsSyncer.mutPendingTx.Lock()
	assert.Equal(t, 0, len(transactionsSyncer.mapValidatorsInfo))
	transactionsSyncer.mutPendingTx.Unlock()

	// txHash does not exist in mapTxsToMiniBlocks
	transactionsSyncer.stopSync = false
	transactionsSyncer.receivedValidatorInfo(txHash, svi)
	transactionsSyncer.mutPendingTx.Lock()
	assert.Equal(t, 0, len(transactionsSyncer.mapValidatorsInfo))
	transactionsSyncer.mutPendingTx.Unlock()

	miniBlock := &block.MiniBlock{
		TxHashes: [][]byte{
			[]byte("a"),
			[]byte("b"),
			[]byte("c"),
		},
	}
	transactionsSyncer.mutPendingTx.Lock()
	transactionsSyncer.mapTxsToMiniBlocks[string(txHash)] = miniBlock
	transactionsSyncer.mutPendingTx.Unlock()

	// value received is not of type *state.ShardValidatorInfo
	transactionsSyncer.receivedValidatorInfo(txHash, nil)
	transactionsSyncer.mutPendingTx.Lock()
	assert.Equal(t, 0, len(transactionsSyncer.mapValidatorsInfo))
	transactionsSyncer.mutPendingTx.Unlock()

	wasReceivedAll := atomic.Flag{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		select {
		case <-transactionsSyncer.chReceivedAll:
			wasReceivedAll.SetValue(true)
		case <-time.After(time.Second):
		}
		wg.Done()
	}()

	// received all missing validators info with success
	transactionsSyncer.receivedValidatorInfo(txHash, svi)
	wg.Wait()
	transactionsSyncer.mutPendingTx.Lock()
	assert.Equal(t, 1, len(transactionsSyncer.mapValidatorsInfo))
	transactionsSyncer.mutPendingTx.Unlock()
	assert.True(t, wasReceivedAll.IsSet())
}

func TestTransactionsSync_GetValidatorInfoFromPoolShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("get validator info from pool when tx hash does not exist in mapTxsToMiniBlocks", func(t *testing.T) {
		t.Parallel()

		txHash := []byte("hash")

		args := createMockArgs()
		transactionsSyncer, _ := NewTransactionsSyncer(args)

		shardValidatorInfo, bFound := transactionsSyncer.getValidatorInfoFromPool(txHash)
		assert.Nil(t, shardValidatorInfo)
		assert.False(t, bFound)
	})

	t.Run("get validator info from pool when shard data store is missing", func(t *testing.T) {
		t.Parallel()

		txHash := []byte("hash")

		args := createMockArgs()
		args.DataPools = &dataRetrieverMock.PoolsHolderStub{
			ValidatorsInfoCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{
					ShardDataStoreCalled: func(cacheID string) storage.Cacher {
						return nil
					},
				}
			},
		}
		transactionsSyncer, _ := NewTransactionsSyncer(args)

		miniBlock := &block.MiniBlock{
			TxHashes: [][]byte{
				[]byte("a"),
				[]byte("b"),
				[]byte("c"),
			},
		}
		transactionsSyncer.mutPendingTx.Lock()
		transactionsSyncer.mapTxsToMiniBlocks[string(txHash)] = miniBlock
		transactionsSyncer.mutPendingTx.Unlock()

		shardValidatorInfo, bFound := transactionsSyncer.getValidatorInfoFromPool(txHash)
		assert.Nil(t, shardValidatorInfo)
		assert.False(t, bFound)
	})

	t.Run("get validator info from pool when tx hash is not found in shard data store", func(t *testing.T) {
		t.Parallel()

		txHash := []byte("hash")

		args := createMockArgs()
		args.DataPools = getDataPoolsWithShardValidatorInfoAndTxHash(nil, nil)
		transactionsSyncer, _ := NewTransactionsSyncer(args)

		miniBlock := &block.MiniBlock{
			TxHashes: [][]byte{
				[]byte("a"),
				[]byte("b"),
				[]byte("c"),
			},
		}
		transactionsSyncer.mutPendingTx.Lock()
		transactionsSyncer.mapTxsToMiniBlocks[string(txHash)] = miniBlock
		transactionsSyncer.mutPendingTx.Unlock()

		shardValidatorInfo, bFound := transactionsSyncer.getValidatorInfoFromPool(txHash)
		assert.Nil(t, shardValidatorInfo)
		assert.False(t, bFound)
	})

	t.Run("get validator info from pool when value received from pool is not of type *state.ShardValidatorInfo", func(t *testing.T) {
		t.Parallel()

		txHash := []byte("hash")

		args := createMockArgs()
		args.DataPools = &dataRetrieverMock.PoolsHolderStub{
			ValidatorsInfoCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{
					ShardDataStoreCalled: func(cacheID string) storage.Cacher {
						return &testscommon.CacherStub{
							PeekCalled: func(key []byte) (value interface{}, ok bool) {
								if bytes.Equal(key, txHash) {
									return nil, true
								}
								return nil, false
							},
						}
					},
				}
			},
		}
		transactionsSyncer, _ := NewTransactionsSyncer(args)

		miniBlock := &block.MiniBlock{
			TxHashes: [][]byte{
				[]byte("a"),
				[]byte("b"),
				[]byte("c"),
			},
		}
		transactionsSyncer.mutPendingTx.Lock()
		transactionsSyncer.mapTxsToMiniBlocks[string(txHash)] = miniBlock
		transactionsSyncer.mutPendingTx.Unlock()

		shardValidatorInfo, bFound := transactionsSyncer.getValidatorInfoFromPool(txHash)
		assert.Nil(t, shardValidatorInfo)
		assert.False(t, bFound)
	})

	t.Run("get validator info from pool should work", func(t *testing.T) {
		t.Parallel()

		txHash := []byte("hash")
		svi := &state.ShardValidatorInfo{
			PublicKey: []byte("x"),
		}

		args := createMockArgs()
		args.DataPools = getDataPoolsWithShardValidatorInfoAndTxHash(svi, txHash)
		transactionsSyncer, _ := NewTransactionsSyncer(args)

		miniBlock := &block.MiniBlock{
			TxHashes: [][]byte{
				[]byte("a"),
				[]byte("b"),
				[]byte("c"),
			},
		}
		transactionsSyncer.mutPendingTx.Lock()
		transactionsSyncer.mapTxsToMiniBlocks[string(txHash)] = miniBlock
		transactionsSyncer.mutPendingTx.Unlock()

		shardValidatorInfo, bFound := transactionsSyncer.getValidatorInfoFromPool(txHash)
		assert.Equal(t, svi, shardValidatorInfo)
		assert.True(t, bFound)
	})
}

func TestTransactionsSync_GetValidatorInfoFromPoolWithSearchFirstShouldWork(t *testing.T) {
	t.Parallel()

	txHash := []byte("hash")
	svi := &state.ShardValidatorInfo{
		PublicKey: []byte("x"),
	}

	args := createMockArgs()
	transactionsSyncer, _ := NewTransactionsSyncer(args)

	// txHash is not found in validatorInfoPool
	validatorsInfoPool := &testscommon.ShardedDataStub{
		SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
	}
	shardValidatorInfo, bFound := transactionsSyncer.getValidatorInfoFromPoolWithSearchFirst(txHash, validatorsInfoPool)
	assert.Nil(t, shardValidatorInfo)
	assert.False(t, bFound)

	// value received from validatorInfoPool is not of type *state.ShardValidatorInfo
	validatorsInfoPool = &testscommon.ShardedDataStub{
		SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, true
		},
	}
	shardValidatorInfo, bFound = transactionsSyncer.getValidatorInfoFromPoolWithSearchFirst(txHash, validatorsInfoPool)
	assert.Nil(t, shardValidatorInfo)
	assert.False(t, bFound)

	// get validator info from pool with search first should work
	validatorsInfoPool = &testscommon.ShardedDataStub{
		SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(key, txHash) {
				return svi, true
			}
			return nil, false
		},
	}
	shardValidatorInfo, bFound = transactionsSyncer.getValidatorInfoFromPoolWithSearchFirst(txHash, validatorsInfoPool)
	assert.Equal(t, svi, shardValidatorInfo)
	assert.True(t, bFound)
}

func TestTransactionsSync_GetValidatorInfoFromPoolOrStorage(t *testing.T) {
	t.Parallel()

	t.Run("get validator info from pool or storage should work from pool", func(t *testing.T) {
		t.Parallel()

		txHash := []byte("hash")
		svi := &state.ShardValidatorInfo{
			PublicKey: []byte("x"),
		}

		args := createMockArgs()
		args.DataPools = getDataPoolsWithShardValidatorInfoAndTxHash(svi, txHash)
		transactionsSyncer, _ := NewTransactionsSyncer(args)

		miniBlock := &block.MiniBlock{
			TxHashes: [][]byte{
				[]byte("a"),
				[]byte("b"),
				[]byte("c"),
			},
		}
		transactionsSyncer.mutPendingTx.Lock()
		transactionsSyncer.mapTxsToMiniBlocks[string(txHash)] = miniBlock
		transactionsSyncer.mutPendingTx.Unlock()

		shardValidatorInfo, bFound := transactionsSyncer.getValidatorInfoFromPoolOrStorage(txHash)
		assert.Equal(t, svi, shardValidatorInfo)
		assert.True(t, bFound)
	})

	t.Run("get validator info from pool or storage when txHash does not exist in mapTxsToMiniBlocks", func(t *testing.T) {
		t.Parallel()

		txHash := []byte("hash")

		args := createMockArgs()
		transactionsSyncer, _ := NewTransactionsSyncer(args)

		shardValidatorInfo, bFound := transactionsSyncer.getValidatorInfoFromPoolOrStorage(txHash)
		assert.Nil(t, shardValidatorInfo)
		assert.False(t, bFound)
	})

	t.Run("get validator info from pool or storage should work using search first", func(t *testing.T) {
		t.Parallel()

		txHash := []byte("hash")
		svi := &state.ShardValidatorInfo{
			PublicKey: []byte("x"),
		}

		args := createMockArgs()
		args.DataPools = &dataRetrieverMock.PoolsHolderStub{
			ValidatorsInfoCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{
					ShardDataStoreCalled: func(cacheID string) storage.Cacher {
						return &testscommon.CacherStub{
							PeekCalled: func(key []byte) (value interface{}, ok bool) {
								return nil, false
							},
						}
					},
					SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
						return svi, true
					},
				}
			},
		}
		transactionsSyncer, _ := NewTransactionsSyncer(args)

		miniBlock := &block.MiniBlock{
			TxHashes: [][]byte{
				[]byte("a"),
				[]byte("b"),
				[]byte("c"),
			},
		}
		transactionsSyncer.mutPendingTx.Lock()
		transactionsSyncer.mapTxsToMiniBlocks[string(txHash)] = miniBlock
		transactionsSyncer.mutPendingTx.Unlock()

		shardValidatorInfo, bFound := transactionsSyncer.getValidatorInfoFromPoolOrStorage(txHash)
		assert.Equal(t, svi, shardValidatorInfo)
		assert.True(t, bFound)
	})

	t.Run("get validator info from pool or storage when txHash does not exist in storage", func(t *testing.T) {
		t.Parallel()

		txHash := []byte("hash")

		args := createMockArgs()
		args.Storages = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						return nil, errors.New("error")
					},
				}, nil
			},
		}
		args.DataPools = getDataPoolsWithShardValidatorInfoAndTxHash(nil, nil)
		transactionsSyncer, _ := NewTransactionsSyncer(args)

		miniBlock := &block.MiniBlock{
			TxHashes: [][]byte{
				[]byte("a"),
				[]byte("b"),
				[]byte("c"),
			},
		}
		transactionsSyncer.mutPendingTx.Lock()
		transactionsSyncer.mapTxsToMiniBlocks[string(txHash)] = miniBlock
		transactionsSyncer.mutPendingTx.Unlock()

		shardValidatorInfo, bFound := transactionsSyncer.getValidatorInfoFromPoolOrStorage(txHash)
		assert.Nil(t, shardValidatorInfo)
		assert.False(t, bFound)
	})

	t.Run("get validator info from pool or storage should work from storage", func(t *testing.T) {
		t.Parallel()

		txHash := []byte("hash")
		svi := &state.ShardValidatorInfo{
			PublicKey: []byte("x"),
		}

		args := createMockArgs()
		marshalledSVI, _ := args.Marshaller.Marshal(svi)
		args.Storages = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						if bytes.Equal(key, txHash) {
							return marshalledSVI, nil
						}
						return nil, errors.New("error")
					},
				}, nil
			},
		}
		args.DataPools = getDataPoolsWithShardValidatorInfoAndTxHash(nil, nil)
		transactionsSyncer, _ := NewTransactionsSyncer(args)

		miniBlock := &block.MiniBlock{
			TxHashes: [][]byte{
				[]byte("a"),
				[]byte("b"),
				[]byte("c"),
			},
		}
		transactionsSyncer.mutPendingTx.Lock()
		transactionsSyncer.mapTxsToMiniBlocks[string(txHash)] = miniBlock
		transactionsSyncer.mutPendingTx.Unlock()

		shardValidatorInfo, bFound := transactionsSyncer.getValidatorInfoFromPoolOrStorage(txHash)
		assert.Equal(t, svi, shardValidatorInfo)
		assert.True(t, bFound)
	})
}

func TestTransactionsSync_GetValidatorsInfoShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	transactionsSyncer, _ := NewTransactionsSyncer(args)

	transactionsSyncer.syncedAll = false
	mapShardValidatorInfo, err := transactionsSyncer.GetValidatorsInfo()
	assert.Nil(t, mapShardValidatorInfo)
	assert.Equal(t, update.ErrNotSynced, err)

	txHash1 := []byte("hash1")
	svi1 := &state.ShardValidatorInfo{
		PublicKey: []byte("x"),
	}
	txHash2 := []byte("hash2")
	svi2 := &state.ShardValidatorInfo{
		PublicKey: []byte("y"),
	}
	transactionsSyncer.mapValidatorsInfo[string(txHash1)] = svi1
	transactionsSyncer.mapValidatorsInfo[string(txHash2)] = svi2

	transactionsSyncer.syncedAll = true
	mapShardValidatorInfo, err = transactionsSyncer.GetValidatorsInfo()
	assert.Equal(t, 2, len(mapShardValidatorInfo))
	assert.Equal(t, svi1, mapShardValidatorInfo[string(txHash1)])
	assert.Equal(t, svi2, mapShardValidatorInfo[string(txHash2)])
	assert.Nil(t, err)
}

func TestTransactionsSync_ClearFieldsShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	transactionsSyncer, _ := NewTransactionsSyncer(args)

	transactionsSyncer.mapTransactions["a"] = &dataTransaction.Transaction{}
	transactionsSyncer.mapTxsToMiniBlocks["b"] = &block.MiniBlock{}
	transactionsSyncer.mapValidatorsInfo["c"] = &state.ShardValidatorInfo{}

	assert.Equal(t, 1, len(transactionsSyncer.mapTransactions))
	assert.Equal(t, 1, len(transactionsSyncer.mapTxsToMiniBlocks))
	assert.Equal(t, 1, len(transactionsSyncer.mapValidatorsInfo))

	transactionsSyncer.ClearFields()

	assert.Equal(t, 0, len(transactionsSyncer.mapTransactions))
	assert.Equal(t, 0, len(transactionsSyncer.mapTxsToMiniBlocks))
	assert.Equal(t, 0, len(transactionsSyncer.mapValidatorsInfo))
}

func getDataPoolsWithShardValidatorInfoAndTxHash(svi *state.ShardValidatorInfo, txHash []byte) dataRetriever.PoolsHolder {
	return &dataRetrieverMock.PoolsHolderStub{
		ValidatorsInfoCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheID string) storage.Cacher {
					return &testscommon.CacherStub{
						PeekCalled: func(key []byte) (value interface{}, ok bool) {
							if bytes.Equal(key, txHash) {
								return svi, true
							}
							return nil, false
						},
					}
				},
				SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, false
				},
			}
		},
	}
}
