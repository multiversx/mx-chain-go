package block

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	commonStorage "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

var (
	errExpected = errors.New("expected error")
	headerHash  = []byte("headerHash")
)

func TestBaseProcessor_createMiniBlockHeaderHandlers(t *testing.T) {
	t.Parallel()

	t.Run("not header v3 should skip", func(t *testing.T) {
		t.Parallel()

		bp := &baseProcessor{
			marshalizer:                &marshallerMock.MarshalizerStub{},
			hasher:                     &testscommon.HasherStub{},
			shardCoordinator:           &testscommon.ShardsCoordinatorMock{},
			enableEpochsHandler:        &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			processedMiniBlocksTracker: &testscommon.ProcessedMiniBlocksTrackerStub{},
			dataPool: &dataRetrieverMock.PoolsHolderStub{
				ExecutedMiniBlocksCalled: func() storage.Cacher {
					return &cache.CacherStub{
						PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
							require.Fail(t, "should not be called")
							return false
						},
					}
				},
			},
		}

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{},
			},
		}
		_, _, err := bp.createMiniBlockHeaderHandlers(body, nil, false)
		require.NoError(t, err)
	})
	t.Run("cross incoming should skip", func(t *testing.T) {
		t.Parallel()

		bp := &baseProcessor{
			marshalizer:                &marshallerMock.MarshalizerStub{},
			hasher:                     &testscommon.HasherStub{},
			shardCoordinator:           &testscommon.ShardsCoordinatorMock{},
			enableEpochsHandler:        &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			processedMiniBlocksTracker: &testscommon.ProcessedMiniBlocksTrackerStub{},
			dataPool: &dataRetrieverMock.PoolsHolderStub{
				ExecutedMiniBlocksCalled: func() storage.Cacher {
					return &cache.CacherStub{
						PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
							require.Fail(t, "should not be called")
							return false
						},
					}
				},
			},
		}

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					ReceiverShardID: bp.shardCoordinator.SelfId(),
					SenderShardID:   bp.shardCoordinator.SelfId() + 1,
				},
			},
		}
		_, _, err := bp.createMiniBlockHeaderHandlers(body, nil, true)
		require.NoError(t, err)
	})
	t.Run("marshal error", func(t *testing.T) {
		t.Parallel()

		cntMarshal := 0
		bp := &baseProcessor{
			marshalizer: &marshallerMock.MarshalizerStub{
				MarshalCalled: func(obj interface{}) ([]byte, error) {
					cntMarshal++
					if cntMarshal == 2 {
						return nil, errExpected
					}
					return []byte(""), nil
				},
			},
			hasher:                     &testscommon.HasherStub{},
			shardCoordinator:           &testscommon.ShardsCoordinatorMock{},
			enableEpochsHandler:        &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			processedMiniBlocksTracker: &testscommon.ProcessedMiniBlocksTrackerStub{},
			dataPool: &dataRetrieverMock.PoolsHolderStub{
				ExecutedMiniBlocksCalled: func() storage.Cacher {
					return &cache.CacherStub{
						PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
							require.Fail(t, "should not be called")
							return false
						},
					}
				},
			},
		}

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					ReceiverShardID: bp.shardCoordinator.SelfId() + 1,
					SenderShardID:   bp.shardCoordinator.SelfId(),
				},
			},
		}
		_, _, err := bp.createMiniBlockHeaderHandlers(body, nil, true)
		require.Equal(t, errExpected, err)
	})
	t.Run("should add executed mini blocks", func(t *testing.T) {
		t.Parallel()

		wasPutCalled := false
		bp := &baseProcessor{
			marshalizer:                &marshallerMock.MarshalizerStub{},
			hasher:                     &testscommon.HasherStub{},
			shardCoordinator:           &testscommon.ShardsCoordinatorMock{},
			enableEpochsHandler:        &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			processedMiniBlocksTracker: &testscommon.ProcessedMiniBlocksTrackerStub{},
			dataPool: &dataRetrieverMock.PoolsHolderStub{
				ExecutedMiniBlocksCalled: func() storage.Cacher {
					return &cache.CacherStub{
						PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
							wasPutCalled = true
							return false
						},
					}
				},
			},
		}

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					ReceiverShardID: bp.shardCoordinator.SelfId() + 1,
					SenderShardID:   bp.shardCoordinator.SelfId(),
				},
			},
		}
		_, _, err := bp.createMiniBlockHeaderHandlers(body, nil, true)
		require.NoError(t, err)
		require.True(t, wasPutCalled)
	})
}

func TestBaseProcessor_cacheIntermediateTxsForHeader(t *testing.T) {
	t.Parallel()

	t.Run("marshal error", func(t *testing.T) {
		t.Parallel()

		bp := &baseProcessor{
			marshalizer: &marshallerMock.MarshalizerStub{
				MarshalCalled: func(obj interface{}) ([]byte, error) {
					return nil, errExpected
				},
			},
			txCoordinator: &testscommon.TransactionCoordinatorMock{},
		}

		err := bp.cacheIntermediateTxsForHeader(headerHash)
		require.Equal(t, errExpected, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasPutCalled := false
		bp := &baseProcessor{
			marshalizer:   &marshallerMock.MarshalizerStub{},
			txCoordinator: &testscommon.TransactionCoordinatorMock{},
			dataPool: &dataRetrieverMock.PoolsHolderStub{
				PostProcessTransactionsCalled: func() storage.Cacher {
					return &cache.CacherStub{
						PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
							wasPutCalled = true
							return false
						},
					}
				},
			},
		}

		err := bp.cacheIntermediateTxsForHeader(headerHash)
		require.NoError(t, err)
		require.True(t, wasPutCalled)
	})
}

func TestBaseProcessor_saveExecutedData(t *testing.T) {
	t.Parallel()

	t.Run("not header v3", func(t *testing.T) {
		t.Parallel()

		bp := &baseProcessor{}
		header := &testscommon.HeaderHandlerStub{
			IsHeaderV3Called: func() bool {
				return false
			},
		}

		err := bp.saveExecutedData(header, headerHash)
		require.Nil(t, err)
	})
	t.Run("header v3 with no execution results", func(t *testing.T) {
		t.Parallel()

		bp := &baseProcessor{
			txCoordinator: &testscommon.TransactionCoordinatorMock{
				GetAllIntermediateTxsCalled: func() map[block.Type]map[string]data.TransactionHandler {
					return make(map[block.Type]map[string]data.TransactionHandler)
				},
			},
			marshalizer: &marshallerMock.MarshalizerStub{},
			dataPool: &dataRetrieverMock.PoolsHolderStub{
				PostProcessTransactionsCalled: func() storage.Cacher {
					return &cache.CacherStub{
						GetCalled: func(key []byte) (value interface{}, ok bool) {
							return []byte("marshalled map"), true
						},
					}
				},
			},
		}
		header := &testscommon.HeaderHandlerStub{
			IsHeaderV3Called: func() bool {
				return true
			},
			GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
				return []data.BaseExecutionResultHandler{}
			},
		}

		err := bp.saveExecutedData(header, headerHash)
		require.NoError(t, err)
	})
	t.Run("saveMiniBlocksFromExecutionResults path", func(t *testing.T) {
		t.Run("extractMiniBlocksHeaderHandlersFromExecResult cast failure for meta execution result", func(t *testing.T) {
			t.Parallel()

			bp := &baseProcessor{}
			header := &testscommon.HeaderHandlerStub{
				IsHeaderV3Called: func() bool {
					return true
				},
				GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
					return []data.BaseExecutionResultHandler{
						&block.ExecutionResult{}, // shard execution result
					}
				},
				GetShardIDCalled: func() uint32 {
					return common.MetachainShardId // meta
				},
			}

			err := bp.saveExecutedData(header, headerHash)
			require.Equal(t, process.ErrWrongTypeAssertion, err)
		})
		t.Run("putMiniBlocksIntoStorage early exit, empty mini block handlers", func(t *testing.T) {
			t.Parallel()

			bp := &baseProcessor{
				store: &commonStorage.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						require.Fail(t, "should not be called")
						return nil, nil
					},
				},
				dataPool: &dataRetrieverMock.PoolsHolderStub{
					PostProcessTransactionsCalled: func() storage.Cacher {
						return &cache.CacherStub{
							GetCalled: func(key []byte) (value interface{}, ok bool) {
								return []byte("marshalled map"), true
							},
						}
					},
				},
				marshalizer: &marshallerMock.MarshalizerStub{},
			}
			header := &testscommon.HeaderHandlerStub{
				IsHeaderV3Called: func() bool {
					return true
				},
				GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
					return []data.BaseExecutionResultHandler{
						&block.ExecutionResult{
							MiniBlockHeaders: []block.MiniBlockHeader{},
						},
					}
				},
			}

			err := bp.saveExecutedData(header, headerHash)
			require.NoError(t, err)
		})
		t.Run("putMiniBlocksIntoStorage returns error on GetStorer", func(t *testing.T) {
			t.Parallel()

			bp := &baseProcessor{
				store: &commonStorage.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return nil, errExpected
					},
				},
			}
			header := &testscommon.HeaderHandlerStub{
				IsHeaderV3Called: func() bool {
					return true
				},
				GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
					return []data.BaseExecutionResultHandler{
						&block.ExecutionResult{
							MiniBlockHeaders: []block.MiniBlockHeader{
								{},
							},
						},
					}
				},
			}

			err := bp.saveExecutedData(header, headerHash)
			require.Equal(t, errExpected, err)
		})
		t.Run("putMiniBlocksIntoStorage does not find a mini block in cache", func(t *testing.T) {
			t.Parallel()

			bp := &baseProcessor{
				store: &commonStorage.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return &commonStorage.StorerStub{}, nil
					},
				},
				dataPool: &dataRetrieverMock.PoolsHolderStub{
					ExecutedMiniBlocksCalled: func() storage.Cacher {
						return &cache.CacherStub{
							GetCalled: func(key []byte) (value interface{}, ok bool) {
								return nil, false
							},
						}
					},
				},
			}
			header := &testscommon.HeaderHandlerStub{
				IsHeaderV3Called: func() bool {
					return true
				},
				GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
					return []data.BaseExecutionResultHandler{
						&block.ExecutionResult{
							MiniBlockHeaders: []block.MiniBlockHeader{
								{},
							},
						},
					}
				},
			}

			err := bp.saveExecutedData(header, headerHash)
			require.Equal(t, process.ErrMissingMiniBlock, err)
		})
		t.Run("putMiniBlocksIntoStorage fails tu add into storer", func(t *testing.T) {
			t.Parallel()

			bp := &baseProcessor{
				store: &commonStorage.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return &commonStorage.StorerStub{
							PutCalled: func(key, data []byte) error {
								return errExpected
							},
						}, nil
					},
				},
				dataPool: &dataRetrieverMock.PoolsHolderStub{
					ExecutedMiniBlocksCalled: func() storage.Cacher {
						return &cache.CacherStub{
							GetCalled: func(key []byte) (value interface{}, ok bool) {
								return []byte("marshalled mb"), true
							},
						}
					},
				},
			}
			header := &testscommon.HeaderHandlerStub{
				IsHeaderV3Called: func() bool {
					return true
				},
				GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
					return []data.BaseExecutionResultHandler{
						&block.MetaExecutionResult{
							MiniBlockHeaders: []block.MiniBlockHeader{
								{},
							},
						},
					}
				},
			}

			err := bp.saveExecutedData(header, headerHash)
			require.Equal(t, errExpected, err)
		})
	})
	t.Run("saveIntermediateTxs path", func(t *testing.T) {
		t.Run("header not found in the cache", func(t *testing.T) {
			t.Parallel()

			bp := &baseProcessor{
				store: &commonStorage.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						require.Fail(t, "should not be called")
						return nil, nil
					},
				},
				dataPool: &dataRetrieverMock.PoolsHolderStub{
					PostProcessTransactionsCalled: func() storage.Cacher {
						return &cache.CacherStub{
							GetCalled: func(key []byte) (value interface{}, ok bool) {
								return nil, false
							},
						}
					},
				},
				marshalizer: &marshallerMock.MarshalizerStub{},
			}
			header := &testscommon.HeaderHandlerStub{
				IsHeaderV3Called: func() bool {
					return true
				},
				GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
					return []data.BaseExecutionResultHandler{
						&block.ExecutionResult{
							MiniBlockHeaders: []block.MiniBlockHeader{},
						},
					}
				},
			}

			err := bp.saveExecutedData(header, headerHash)
			require.Equal(t, process.ErrMissingHeader, err)
		})
		t.Run("unmarshall error", func(t *testing.T) {
			t.Parallel()

			bp := &baseProcessor{
				store: &commonStorage.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						require.Fail(t, "should not be called")
						return nil, nil
					},
				},
				dataPool: &dataRetrieverMock.PoolsHolderStub{
					PostProcessTransactionsCalled: func() storage.Cacher {
						return &cache.CacherStub{
							GetCalled: func(key []byte) (value interface{}, ok bool) {
								return []byte("marshalled map"), true
							},
						}
					},
				},
				marshalizer: &marshallerMock.MarshalizerStub{
					UnmarshalCalled: func(obj interface{}, buff []byte) error {
						return errExpected
					},
				},
			}
			header := &testscommon.HeaderHandlerStub{
				IsHeaderV3Called: func() bool {
					return true
				},
				GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
					return []data.BaseExecutionResultHandler{
						&block.ExecutionResult{
							MiniBlockHeaders: []block.MiniBlockHeader{},
						},
					}
				},
			}

			err := bp.saveExecutedData(header, headerHash)
			require.Equal(t, errExpected, err)
		})
		t.Run("putTransactionsIntoStorage fails due to invalid block type", func(t *testing.T) {
			t.Parallel()

			bp := &baseProcessor{
				store: &commonStorage.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						require.Fail(t, "should not be called")
						return nil, nil
					},
				},
				dataPool: &dataRetrieverMock.PoolsHolderStub{
					PostProcessTransactionsCalled: func() storage.Cacher {
						return &cache.CacherStub{
							GetCalled: func(key []byte) (value interface{}, ok bool) {
								return []byte("marshalled map"), true
							},
						}
					},
				},
				marshalizer: &marshallerMock.MarshalizerStub{
					UnmarshalCalled: func(obj interface{}, buff []byte) error {
						txsMap := obj.(map[block.Type]map[string]data.TransactionHandler)
						txsMap[block.PeerBlock] = map[string]data.TransactionHandler{} // should never have PeerBlock
						return nil
					},
				},
			}
			header := &testscommon.HeaderHandlerStub{
				IsHeaderV3Called: func() bool {
					return true
				},
				GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
					return []data.BaseExecutionResultHandler{
						&block.ExecutionResult{
							MiniBlockHeaders: []block.MiniBlockHeader{},
						},
					}
				},
			}

			err := bp.saveExecutedData(header, headerHash)
			require.Equal(t, process.ErrInvalidBlockType, err)
		})
		t.Run("putTransactionsIntoStorage fails due to GetStorer issue", func(t *testing.T) {
			t.Parallel()

			bp := &baseProcessor{
				store: &commonStorage.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						if unitType == dataRetriever.TransactionUnit {
							return nil, errExpected
						}

						return &commonStorage.StorerStub{}, nil
					},
				},
				dataPool: &dataRetrieverMock.PoolsHolderStub{
					PostProcessTransactionsCalled: func() storage.Cacher {
						return &cache.CacherStub{
							GetCalled: func(key []byte) (value interface{}, ok bool) {
								return []byte("marshalled map"), true
							},
						}
					},
				},
				marshalizer: &marshallerMock.MarshalizerStub{
					UnmarshalCalled: func(obj interface{}, buff []byte) error {
						txsMap := obj.(map[block.Type]map[string]data.TransactionHandler)
						txsMap[block.TxBlock] = map[string]data.TransactionHandler{} // force TransactionUnit
						return nil
					},
				},
			}
			header := &testscommon.HeaderHandlerStub{
				IsHeaderV3Called: func() bool {
					return true
				},
				GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
					return []data.BaseExecutionResultHandler{
						&block.ExecutionResult{
							MiniBlockHeaders: []block.MiniBlockHeader{},
						},
					}
				},
			}

			err := bp.saveExecutedData(header, headerHash)
			require.Equal(t, errExpected, err)
		})
		t.Run("putOneTransactionIntoStorage fails due to nil transaction", func(t *testing.T) {
			t.Parallel()

			bp := &baseProcessor{
				store: &commonStorage.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return &commonStorage.StorerStub{}, nil
					},
				},
				dataPool: &dataRetrieverMock.PoolsHolderStub{
					PostProcessTransactionsCalled: func() storage.Cacher {
						return &cache.CacherStub{
							GetCalled: func(key []byte) (value interface{}, ok bool) {
								return []byte("marshalled map"), true
							},
						}
					},
				},
				marshalizer: &marshallerMock.MarshalizerStub{
					UnmarshalCalled: func(obj interface{}, buff []byte) error {
						txsMap := obj.(map[block.Type]map[string]data.TransactionHandler)
						txsMap[block.TxBlock] = map[string]data.TransactionHandler{
							"hash": nil,
						}
						return nil
					},
				},
			}
			header := &testscommon.HeaderHandlerStub{
				IsHeaderV3Called: func() bool {
					return true
				},
				GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
					return []data.BaseExecutionResultHandler{
						&block.ExecutionResult{
							MiniBlockHeaders: []block.MiniBlockHeader{},
						},
					}
				},
			}

			err := bp.saveExecutedData(header, headerHash)
			require.Equal(t, process.ErrNilTransaction, err)
		})
		t.Run("putOneTransactionIntoStorage fails due to marshal error", func(t *testing.T) {
			t.Parallel()

			bp := &baseProcessor{
				store: &commonStorage.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return &commonStorage.StorerStub{}, nil
					},
				},
				dataPool: &dataRetrieverMock.PoolsHolderStub{
					PostProcessTransactionsCalled: func() storage.Cacher {
						return &cache.CacherStub{
							GetCalled: func(key []byte) (value interface{}, ok bool) {
								return []byte("marshalled map"), true
							},
						}
					},
				},
				marshalizer: &marshallerMock.MarshalizerStub{
					UnmarshalCalled: func(obj interface{}, buff []byte) error {
						txsMap := obj.(map[block.Type]map[string]data.TransactionHandler)
						txsMap[block.TxBlock] = map[string]data.TransactionHandler{
							"hash": &transaction.Transaction{},
						}
						return nil
					},
					MarshalCalled: func(obj interface{}) ([]byte, error) {
						return nil, errExpected
					},
				},
			}
			header := &testscommon.HeaderHandlerStub{
				IsHeaderV3Called: func() bool {
					return true
				},
				GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
					return []data.BaseExecutionResultHandler{
						&block.ExecutionResult{
							MiniBlockHeaders: []block.MiniBlockHeader{},
						},
					}
				},
			}

			err := bp.saveExecutedData(header, headerHash)
			require.Equal(t, errExpected, err)
		})
	})
	t.Run("should work and move all", func(t *testing.T) {
		t.Parallel()

		cntPutCalled := 0
		wasRemoveCalledForTxs := false
		wasRemoveCalledForMbs := false
		bp := &baseProcessor{
			store: &commonStorage.ChainStorerStub{
				GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
					return &commonStorage.StorerStub{
						PutCalled: func(key, data []byte) error {
							cntPutCalled++
							return nil
						},
					}, nil
				},
			},
			dataPool: &dataRetrieverMock.PoolsHolderStub{
				PostProcessTransactionsCalled: func() storage.Cacher {
					return &cache.CacherStub{
						GetCalled: func(key []byte) (value interface{}, ok bool) {
							return []byte("marshalled map"), true
						},
						RemoveCalled: func(key []byte) {
							wasRemoveCalledForTxs = true
						},
					}
				},
				ExecutedMiniBlocksCalled: func() storage.Cacher {
					return &cache.CacherStub{
						GetCalled: func(key []byte) (value interface{}, ok bool) {
							return []byte("marshalled mb"), true
						},
						RemoveCalled: func(key []byte) {
							wasRemoveCalledForMbs = true
						},
					}
				},
			},
			marshalizer: &marshallerMock.MarshalizerStub{
				UnmarshalCalled: func(obj interface{}, buff []byte) error {
					txsMap := obj.(map[block.Type]map[string]data.TransactionHandler)
					txsMap[block.SmartContractResultBlock] = map[string]data.TransactionHandler{
						"hashSCR": &transaction.Transaction{},
					}
					txsMap[block.RewardsBlock] = map[string]data.TransactionHandler{
						"hashReward": &transaction.Transaction{}, // for coverage
					}
					txsMap[block.ReceiptBlock] = map[string]data.TransactionHandler{
						"hashReward": &transaction.Transaction{}, // for coverage
					}
					return nil
				},
			},
		}
		header := &testscommon.HeaderHandlerStub{
			IsHeaderV3Called: func() bool {
				return true
			},
			GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
				return []data.BaseExecutionResultHandler{
					&block.MetaExecutionResult{
						MiniBlockHeaders: []block.MiniBlockHeader{
							{},
						},
					},
				}
			},
			GetShardIDCalled: func() uint32 {
				return common.MetachainShardId
			},
		}

		err := bp.saveExecutedData(header, headerHash)
		require.NoError(t, err)
		require.True(t, wasRemoveCalledForTxs)
		require.True(t, wasRemoveCalledForMbs)
		require.Equal(t, 4, cntPutCalled) // 3 types of tx blocks + one for mbs
	})
}
