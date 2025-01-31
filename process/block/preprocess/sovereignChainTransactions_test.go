package preprocess

import (
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	state2 "github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/state"
)

func TestTxsPreprocessor_NewSovereignChainTransactionPreprocessorShouldErrNilPreProcessor(t *testing.T) {
	t.Parallel()

	sctp, err := NewSovereignChainTransactionPreprocessor(nil)
	assert.Nil(t, sctp)
	assert.Equal(t, process.ErrNilPreProcessor, err)
}

func TestTxsPreprocessor_NewSovereignChainTransactionPreprocessorShouldWork(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()

	tp, err := NewTransactionPreprocessor(args)
	require.Nil(t, err)
	require.NotNil(t, tp)

	sctp, err := NewSovereignChainTransactionPreprocessor(tp)
	require.Nil(t, err)
	require.NotNil(t, sctp)
}

func TestTxsPreprocessor_ProcessBlockTransactionsShouldWork(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()

	tp, _ := NewTransactionPreprocessor(args)
	sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

	tx1 := transaction.Transaction{
		Nonce: 1,
	}
	tx2 := transaction.Transaction{
		Nonce: 2,
	}
	tx3 := transaction.Transaction{
		Nonce: 3,
	}

	txHash1 := []byte("a")
	txHash2 := []byte("b")
	txHash3 := []byte("c")

	tp.txsForCurrBlock.txHashAndInfo[string(txHash1)] = &txInfo{
		tx: &tx1,
	}
	tp.txsForCurrBlock.txHashAndInfo[string(txHash2)] = &txInfo{
		tx: &tx2,
	}
	tp.txsForCurrBlock.txHashAndInfo[string(txHash3)] = &txInfo{
		tx: &tx3,
	}

	header := &block.Header{
		PrevRandSeed: []byte("X"),
	}
	body := &block.Body{
		MiniBlocks: block.MiniBlockSlice{
			&block.MiniBlock{
				TxHashes: [][]byte{
					txHash1, txHash2, txHash3,
				},
			},
		},
	}

	mbs, err := sctp.ProcessBlockTransactions(header, body, haveTimeTrue)
	require.Nil(t, err)
	require.NotNil(t, mbs)
	require.Equal(t, 1, len(mbs))
	require.Equal(t, 3, len(mbs[0].TxHashes))
	assert.Equal(t, txHash1, mbs[0].TxHashes[0])
	assert.Equal(t, txHash2, mbs[0].TxHashes[1])
	assert.Equal(t, txHash3, mbs[0].TxHashes[2])
}

func TestTxsPreprocessor_CreateAndProcessMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("CreateAndProcessMiniBlocks should return empty mini blocks slice when computeSortedTxs fails", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()
		args.TxDataPool = &testscommon.ShardedDataStub{
			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
				return nil
			},
		}

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		mbs, err := sctp.CreateAndProcessMiniBlocks(haveTimeTrue, []byte("X"))
		assert.Nil(t, err)
		assert.Equal(t, 0, len(mbs))
	})

	t.Run("CreateAndProcessMiniBlocks should return empty mini blocks slice when there are no sorted txs", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()
		args.TxDataPool = &testscommon.ShardedDataStub{
			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
				return &testscommon.CacherStub{}
			},
		}

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		mbs, err := sctp.CreateAndProcessMiniBlocks(haveTimeTrue, []byte("X"))
		assert.Nil(t, err)
		assert.Equal(t, 0, len(mbs))
	})

	t.Run("CreateAndProcessMiniBlocks should return empty mini blocks slice when have no time", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()
		args.TxDataPool = &testscommon.ShardedDataStub{
			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
				return &testscommon.TxCacherStub{
					SelectTransactionsCalled: func(_ txcache.SelectionSession, _ uint64, _ int, _ time.Duration) ([]*txcache.WrappedTransaction, uint64) {
						return []*txcache.WrappedTransaction{
							{Tx: &transaction.Transaction{Nonce: 1}},
						}, 0
					},
				}
			},
		}

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		haveTimeFalse := func() bool { return false }
		mbs, err := sctp.CreateAndProcessMiniBlocks(haveTimeFalse, []byte("X"))
		assert.Nil(t, err)
		assert.Equal(t, 0, len(mbs))
	})

	t.Run("CreateAndProcessMiniBlocks should return empty mini blocks slice when scheduled is not activated", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()
		args.TxDataPool = &testscommon.ShardedDataStub{
			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
				return &testscommon.TxCacherStub{
					SelectTransactionsCalled: func(_ txcache.SelectionSession, _ uint64, _ int, _ time.Duration) ([]*txcache.WrappedTransaction, uint64) {
						return []*txcache.WrappedTransaction{
							{Tx: &transaction.Transaction{Nonce: 1}},
						}, 0
					},
				}
			},
		}

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		mbs, err := sctp.CreateAndProcessMiniBlocks(haveTimeTrue, []byte("X"))
		assert.Nil(t, err)
		assert.Equal(t, 0, len(mbs))
	})

	t.Run("CreateAndProcessMiniBlocks should work", func(t *testing.T) {
		t.Parallel()
		args := createDefaultTransactionsProcessorArgs()
		args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		args.TxProcessor = &testscommon.TxProcessorMock{
			VerifyTransactionCalled: func(tx *transaction.Transaction) error {
				return nil
			},
			GetSenderAndReceiverAccountsCalled: func(tx *transaction.Transaction) (state2.UserAccountHandler, state2.UserAccountHandler, error) {
				senderAccount := &state.UserAccountStub{
					Nonce:   tx.Nonce,
					Balance: big.NewInt(10),
				}
				return senderAccount, nil, nil
			},
		}

		tx1 := &transaction.Transaction{
			Nonce: 1,
			Value: big.NewInt(1),
		}
		tx2 := &transaction.Transaction{
			Nonce: 2,
			Value: big.NewInt(2),
			Data:  []byte("X"),
		}
		tx3 := &transaction.Transaction{
			Nonce: 3,
			Value: big.NewInt(3),
		}

		txHash1 := []byte("1")
		txHash2 := []byte("2")
		txHash3 := []byte("3")

		args.TxDataPool = &testscommon.ShardedDataStub{
			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
				return &testscommon.TxCacherStub{
					SelectTransactionsCalled: func(_ txcache.SelectionSession, _ uint64, _ int, _ time.Duration) ([]*txcache.WrappedTransaction, uint64) {
						return []*txcache.WrappedTransaction{
							{Tx: tx1, TxHash: txHash1},
							{Tx: tx2, TxHash: txHash2},
							{Tx: tx3, TxHash: txHash3},
						}, 0
					},
				}
			},
		}

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		mbs, err := sctp.CreateAndProcessMiniBlocks(haveTimeTrue, []byte("X"))
		assert.Nil(t, err)
		require.Equal(t, 1, len(mbs))
		require.Equal(t, 3, len(mbs[0].TxHashes))
	})
}

func TestTxsPreprocessor_ComputeSortedTxsShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("computeSortedTxs should return error when tx data pool is nil", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()
		args.TxDataPool = &testscommon.ShardedDataStub{
			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
				return nil
			},
		}

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		wtxs, err := sctp.computeSortedTxs(0, 0, []byte("X"))
		assert.Nil(t, wtxs)
		assert.Equal(t, process.ErrNilTxDataPool, err)
	})

	t.Run("computeSortedTxs should work", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()
		tx := &transaction.Transaction{Nonce: 1}
		txHash := []byte("x")
		args.TxDataPool = &testscommon.ShardedDataStub{
			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
				return &testscommon.TxCacherStub{
					SelectTransactionsCalled: func(_ txcache.SelectionSession, _ uint64, _ int, _ time.Duration) ([]*txcache.WrappedTransaction, uint64) {
						return []*txcache.WrappedTransaction{
							{Tx: tx, TxHash: txHash},
						}, 0
					},
				}
			},
		}

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		wtxs, err := sctp.computeSortedTxs(0, 0, []byte("X"))
		require.Nil(t, err)
		require.Equal(t, 1, len(wtxs))
		assert.Equal(t, tx, wtxs[0].Tx)
		assert.Equal(t, txHash, wtxs[0].TxHash)
	})
}

func TestTxsPreprocessor_ProcessMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	args := createDefaultTransactionsProcessorArgs()

	tp, _ := NewTransactionPreprocessor(args)
	sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

	txsToBeReverted, indexOfLastTxProcessed, shouldRevert, err := sctp.ProcessMiniBlock(
		&block.MiniBlock{},
		haveTimeTrue,
		haveAdditionalTimeFalse,
		false,
		false,
		-1,
		&testscommon.PreProcessorExecutionInfoHandlerMock{},
	)

	assert.Nil(t, txsToBeReverted)
	assert.Equal(t, 0, indexOfLastTxProcessed)
	assert.False(t, shouldRevert)
	assert.Nil(t, err)
}

func TestTxsPreprocessor_ShouldContinueProcessingScheduledTxShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("shouldContinueProcessingScheduledTx should return false when assertion fails", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		wrappedTx := &txcache.WrappedTransaction{}
		mapSCTxs := make(map[string]struct{})
		mbi := &createScheduledMiniBlocksInfo{}

		tx, mb, shouldContinue := sctp.shouldContinueProcessingScheduledTx(isShardStuckFalse, wrappedTx, mapSCTxs, mbi)
		assert.Nil(t, tx)
		assert.Nil(t, mb)
		assert.False(t, shouldContinue)
	})

	t.Run("shouldContinueProcessingScheduledTx should return false when receiver's mini block does not exist", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		wrappedTx := &txcache.WrappedTransaction{
			Tx: &transaction.Transaction{},
		}

		mapSCTxs := make(map[string]struct{})
		mbi := &createScheduledMiniBlocksInfo{}

		tx, mb, shouldContinue := sctp.shouldContinueProcessingScheduledTx(isShardStuckFalse, wrappedTx, mapSCTxs, mbi)
		assert.Nil(t, tx)
		assert.Nil(t, mb)
		assert.False(t, shouldContinue)
	})

	t.Run("shouldContinueProcessingScheduledTx should return true", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()
		args.BalanceComputation = &testscommon.BalanceComputationStub{
			IsAddressSetCalled: func(address []byte) bool {
				return true
			},
			AddressHasEnoughBalanceCalled: func(address []byte, value *big.Int) bool {
				return true
			},
		}

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		wrappedTx := &txcache.WrappedTransaction{
			Tx: &transaction.Transaction{},
		}

		mapSCTxs := make(map[string]struct{})
		mbi := &createScheduledMiniBlocksInfo{
			mapMiniBlocks: make(map[uint32]*block.MiniBlock),
		}

		mbi.mapMiniBlocks[0] = &block.MiniBlock{}

		tx, mb, shouldContinue := sctp.shouldContinueProcessingScheduledTx(isShardStuckFalse, wrappedTx, mapSCTxs, mbi)
		assert.Equal(t, wrappedTx.Tx, tx)
		assert.Equal(t, mbi.mapMiniBlocks[0], mb)
		assert.True(t, shouldContinue)
	})
}

func TestTxsPreprocessor_IsTransactionEligibleForExecutionShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("isTransactionEligibleForExecution should return false when error is not nil and is not related to higher nonce in transaction", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		expectedError := errors.New("error")
		err, value := sctp.isTransactionEligibleForExecution(nil, expectedError)

		require.Equal(t, expectedError, err)
		require.False(t, value)
	})

	t.Run("isTransactionEligibleForExecution should return false when sender account is nil", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		err, value := sctp.isTransactionEligibleForExecution(nil, nil)

		require.Nil(t, err)
		require.False(t, value)
	})

	t.Run("isTransactionEligibleForExecution should return false when transaction has a higher nonce", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()
		args.TxProcessor = &testscommon.TxProcessorMock{
			GetSenderAndReceiverAccountsCalled: func(tx *transaction.Transaction) (state2.UserAccountHandler, state2.UserAccountHandler, error) {
				senderAccount := &state.UserAccountStub{
					Nonce:   0,
					Balance: big.NewInt(10),
				}
				return senderAccount, nil, nil
			},
		}

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		tx := &transaction.Transaction{
			SndAddr: []byte("X"),
			Nonce:   1,
		}
		err, value := sctp.isTransactionEligibleForExecution(tx, nil)

		require.ErrorIs(t, err, process.ErrHigherNonceInTransaction)
		require.False(t, value)
	})

	t.Run("isTransactionEligibleForExecution should return false when transaction has a lower nonce", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()
		args.TxProcessor = &testscommon.TxProcessorMock{
			GetSenderAndReceiverAccountsCalled: func(tx *transaction.Transaction) (state2.UserAccountHandler, state2.UserAccountHandler, error) {
				senderAccount := &state.UserAccountStub{
					Nonce:   5,
					Balance: big.NewInt(10),
				}
				return senderAccount, nil, nil
			},
		}

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		tx := &transaction.Transaction{
			SndAddr: []byte("X"),
			Nonce:   4,
		}
		err, value := sctp.isTransactionEligibleForExecution(tx, nil)

		require.ErrorIs(t, err, process.ErrLowerNonceInTransaction)
		require.False(t, value)
	})

	t.Run("isTransactionEligibleForExecution should return false when account has insufficient balance for fees", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()
		args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
			ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
				return big.NewInt(1)
			},
		}
		args.TxProcessor = &testscommon.TxProcessorMock{
			GetSenderAndReceiverAccountsCalled: func(tx *transaction.Transaction) (state2.UserAccountHandler, state2.UserAccountHandler, error) {
				senderAccount := &state.UserAccountStub{
					Nonce:   1,
					Balance: big.NewInt(0),
				}
				return senderAccount, nil, nil
			},
		}

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		tx := &transaction.Transaction{
			SndAddr: []byte("X"),
			Nonce:   1,
		}
		err, value := sctp.isTransactionEligibleForExecution(tx, nil)

		require.ErrorIs(t, err, process.ErrInsufficientFee)
		require.False(t, value)
	})

	t.Run("isTransactionEligibleForExecution should return true if account has sufficient funds for fee but not sufficient for transfer.", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()
		args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
			ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
				return big.NewInt(1)
			},
		}
		args.TxProcessor = &testscommon.TxProcessorMock{
			GetSenderAndReceiverAccountsCalled: func(tx *transaction.Transaction) (state2.UserAccountHandler, state2.UserAccountHandler, error) {
				senderAccount := &state.UserAccountStub{
					Nonce:   1,
					Balance: big.NewInt(2),
				}
				return senderAccount, nil, nil
			},
		}

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		tx := &transaction.Transaction{
			SndAddr: []byte("X"),
			Nonce:   1,
			Value:   big.NewInt(2),
		}
		err, value := sctp.isTransactionEligibleForExecution(tx, nil)

		require.Nil(t, err)
		require.True(t, value)
	})

	t.Run("isTransactionEligibleForExecution should return true", func(t *testing.T) {
		t.Parallel()

		args := createDefaultTransactionsProcessorArgs()
		args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
			ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
				return big.NewInt(1)
			},
		}
		args.TxProcessor = &testscommon.TxProcessorMock{
			GetSenderAndReceiverAccountsCalled: func(tx *transaction.Transaction) (state2.UserAccountHandler, state2.UserAccountHandler, error) {
				senderAccount := &state.UserAccountStub{
					Nonce:   1,
					Balance: big.NewInt(10),
				}
				return senderAccount, nil, nil
			},
		}

		tp, _ := NewTransactionPreprocessor(args)
		sctp, _ := NewSovereignChainTransactionPreprocessor(tp)

		tx := &transaction.Transaction{
			SndAddr: []byte("X"),
			Nonce:   1,
			Value:   big.NewInt(2),
		}
		err, value := sctp.isTransactionEligibleForExecution(tx, nil)

		require.Nil(t, err)
		assert.True(t, value)

		accntInfo, found := sctp.accntsTracker.getAccountInfo(tx.GetSndAddr())
		require.True(t, found)

		nonce := accntInfo.nonce
		balance := accntInfo.balance

		assert.Equal(t, uint64(2), nonce)
		assert.Equal(t, big.NewInt(7), balance)

		tx2 := &transaction.Transaction{
			SndAddr: []byte("X"),
			Nonce:   2,
			Value:   big.NewInt(5),
		}
		err, value = sctp.isTransactionEligibleForExecution(tx2, nil)

		require.Nil(t, err)
		assert.True(t, value)

		accntInfo, found = sctp.accntsTracker.getAccountInfo(tx2.GetSndAddr())
		require.True(t, found)

		nonce = accntInfo.nonce
		balance = accntInfo.balance

		assert.Equal(t, uint64(3), nonce)
		assert.Equal(t, big.NewInt(1), balance)
	})
}
