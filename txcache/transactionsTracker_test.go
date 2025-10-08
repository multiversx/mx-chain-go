package txcache

import (
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func Test_updateRangeWithBreadcrumb(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
	require.Nil(t, err)

	txTracker := newTransactionsTracker(tracker, nil)

	t.Run("range of sender not set yet", func(t *testing.T) {
		t.Parallel()

		rangeOfSender := &accountRange{
			minNonce: core.OptionalUint64{Value: math.MaxUint64, HasValue: false},
			maxNonce: core.OptionalUint64{Value: 0, HasValue: false},
		}

		senderBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{
			Value:    10,
			HasValue: true,
		})
		err := senderBreadcrumb.updateNonceRange(core.OptionalUint64{Value: 12, HasValue: true})
		require.NoError(t, err)

		txTracker.updateRangeWithBreadcrumb(rangeOfSender, senderBreadcrumb)
		require.Equal(t, uint64(10), rangeOfSender.minNonce.Value)
		require.Equal(t, uint64(12), rangeOfSender.maxNonce.Value)
	})

	t.Run("the sender breadcrumb is a breadcrumb of fee payer", func(t *testing.T) {
		t.Parallel()

		rangeOfSender := &accountRange{
			minNonce: core.OptionalUint64{Value: 10, HasValue: true},
			maxNonce: core.OptionalUint64{Value: 12, HasValue: true},
		}

		senderBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{
			Value:    0,
			HasValue: false,
		})

		txTracker.updateRangeWithBreadcrumb(rangeOfSender, senderBreadcrumb)
		require.Equal(t, uint64(10), rangeOfSender.minNonce.Value)
	})
}

func Test_IsTransactionTracked(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 6)
	tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
	require.Nil(t, err)

	txCache.tracker = tracker

	accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
		GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
			return 11, big.NewInt(6 * 100000 * oneBillion), true, nil
		},
	}

	txs := []*WrappedTransaction{
		createTx([]byte("txHash1"), "alice", 11).withRelayer([]byte("bob")).withGasLimit(100_000),
		createTx([]byte("txHash2"), "alice", 12),
		createTx([]byte("txHash3"), "alice", 13),
		createTx([]byte("txHash4"), "alice", 14),
		createTx([]byte("txHash5"), "alice", 15).withRelayer([]byte("bob")).withGasLimit(100_000),
		createTx([]byte("txHash6"), "eve", 11).withRelayer([]byte("alice")).withGasLimit(100_000),
		// This one is not proposed. However, will be detected as "tracked" because it has the same nonce with as a tracked one.
		// This is not critical. It is ok that a sender has a specific nonce "protected".
		createTx([]byte("txHash7"), "eve", 11).withRelayer([]byte("alice")).withGasLimit(100_000),
	}

	for _, tx := range txs {
		txCache.AddTx(tx)
	}

	err = txCache.OnProposedBlock(
		[]byte("hash1"),
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txHash1"),
						[]byte("txHash2"),
						[]byte("txHash3"),
					},
				},
			},
		},
		&block.Header{
			Nonce:    uint64(0),
			PrevHash: []byte("hash0"),
			RootHash: []byte("rootHash0"),
		},
		accountsProvider,
		holders.NewBlockchainInfo([]byte("hash0"), []byte("hash0"), 0),
	)
	require.Nil(t, err)

	err = txCache.OnProposedBlock(
		[]byte("hash2"),
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txHash4"),
						[]byte("txHash5"),
					},
				},
			},
		},
		&block.Header{
			Nonce:    uint64(1),
			PrevHash: []byte("hash1"),
			RootHash: []byte("rootHash0"),
		},
		accountsProvider,
		holders.NewBlockchainInfo([]byte("hash0"), []byte("hash0"), 1),
	)
	require.Nil(t, err)

	err = txCache.OnProposedBlock(
		[]byte("hash3"),
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txHash6"),
					},
				},
			},
		},
		&block.Header{
			Nonce:    uint64(2),
			PrevHash: []byte("hash2"),
			RootHash: []byte("rootHash0"),
		},
		accountsProvider,
		holders.NewBlockchainInfo([]byte("hash0"), []byte("hash0"), 2),
	)
	require.Nil(t, err)

	t.Run("should return true", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHash1"), "alice", 11)
		tx2 := createTx([]byte("txHash6"), "eve", 11)
		txTracker := newTransactionsTracker(tracker, []*WrappedTransaction{
			tx1,
			tx2,
		})

		require.True(t, txTracker.isTransactionTracked(tx1))
		require.True(t, txTracker.isTransactionTracked(tx2))
	})

	t.Run("should return false because out of range", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHashX"), "alice", 16)
		tx2 := createTx([]byte("txHashX"), "eve", 12)
		txTracker := newTransactionsTracker(tracker, []*WrappedTransaction{
			tx1,
			tx2,
		})

		require.False(t, txTracker.isTransactionTracked(tx1))
		require.False(t, txTracker.isTransactionTracked(tx2))

	})

	t.Run("should return false because account is only relayer", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHashX"), "alice", 16)
		txTracker := newTransactionsTracker(tracker, []*WrappedTransaction{
			tx1,
		})

		require.False(t, txTracker.isTransactionTracked(tx1))
	})

	t.Run("should return false because account is not tracked at all", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHash2"), "carol", 12)
		txTracker := newTransactionsTracker(tracker, []*WrappedTransaction{
			tx1,
		})

		require.False(t, txTracker.isTransactionTracked(tx1))
	})

	t.Run("should return true for any transaction of sender with a tracked nonce", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHash7"), "eve", 12)
		txTracker := newTransactionsTracker(tracker, []*WrappedTransaction{
			tx1,
		})

		require.False(t, txTracker.isTransactionTracked(tx1))
	})
}

func Test_RemoveTxs(t *testing.T) {
	t.Parallel()

	txByHash := newTxByHashMap(2)
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 10)
	txCache.txByHash = txByHash

	accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
		GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
			return 0, big.NewInt(6 * 100000 * oneBillion), true, nil
		},
	}

	txHashes := createMockTxHashes(10)
	wrappedTxs := createSliceMockWrappedTxsWithSameSender(txHashes, "alice")

	for _, tx := range wrappedTxs {
		txCache.AddTx(tx)
	}

	err := txCache.OnProposedBlock(
		[]byte("hash1"),
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: txHashes[0:4],
				},
			},
		},
		&block.Header{
			Nonce:    uint64(0),
			PrevHash: []byte("hash0"),
			RootHash: []byte("rootHash0"),
		},
		accountsProvider,
		holders.NewBlockchainInfo([]byte("hash0"), []byte("hash0"), 0),
	)
	require.Nil(t, err)

	txsTracker := newTransactionsTracker(txCache.tracker, wrappedTxs)

	untrackedTxs := txsTracker.GetBulkOfUntrackedTransactions(wrappedTxs)
	noOfRemovedTxs := txByHash.RemoveTxsBulk(untrackedTxs)

	require.Equal(t, uint32(6), noOfRemovedTxs)
}
