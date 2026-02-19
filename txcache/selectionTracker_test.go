package txcache

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
)

func proposeBlocks(t *testing.T, numOfBlocks int, selectionTracker *selectionTracker, accountsProvider common.AccountNonceAndBalanceProvider) {

	for i := 1; i < numOfBlocks+1; i++ {
		err := selectionTracker.OnProposedBlock(
			[]byte(fmt.Sprintf("hash%d", i)),
			&block.Body{},
			&block.Header{
				Nonce:    uint64(i),
				PrevHash: []byte(fmt.Sprintf("hash%d", i-1)),
				RootHash: []byte("rootHash0"),
			},
			accountsProvider,
			defaultLatestExecutedHash,
		)
		require.Nil(t, err)
	}
}

func executeBlocksConcurrently(t *testing.T, numOfBlocks int, selectionTracker *selectionTracker) {
	wg := sync.WaitGroup{}
	wg.Add(numOfBlocks)

	for i := 1; i <= numOfBlocks; i++ {
		go func(index int) {
			defer wg.Done()

			rootHash := []byte("rootHash0")

			blockHeader := &block.Header{
				Nonce:    uint64(index),
				PrevHash: []byte(fmt.Sprintf("prevHash%d", index-1)),
			}
			err := selectionTracker.OnExecutedBlock(blockHeader, rootHash)
			require.Nil(t, err)
		}(i)
	}

	wg.Wait()
}

func TestNewSelectionTracker(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		_, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)
	})

	t.Run("should fail because of nil TxCache", func(t *testing.T) {
		t.Parallel()

		tracker, err := NewSelectionTracker(nil, 0, maxTrackedBlocks)
		require.Equal(t, errNilTxCache, err)
		require.Nil(t, tracker)
	})

	t.Run("should fail because of maxTrackedBlocks", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, 0)
		require.Equal(t, errInvalidMaxTrackedBlocks, err)
		require.Nil(t, tracker)
	})
}

func TestSelectionTracker_OnProposedBlockShouldErr(t *testing.T) {
	t.Parallel()

	t.Run("should err nil block hash", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.OnProposedBlock(nil, &block.Body{}, nil, nil, nil)
		require.Equal(t, errNilBlockHash, err)
	})

	t.Run("should err nil block header", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		blockBody := block.Body{}
		err = tracker.OnProposedBlock([]byte("hash1"), &blockBody, nil, nil, nil)
		require.Equal(t, errNilBlockHeader, err)
	})

	t.Run("should return errWrongTypeAssertion", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)

		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.OnProposedBlock([]byte("hash1"), nil, &block.Header{}, nil, defaultLatestExecutedHash)

		require.Equal(t, errWrongTypeAssertion, err)
	})

	t.Run("should err nil accounts provider", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.OnProposedBlock([]byte("hash1"), &block.Body{}, &block.Header{}, nil, nil)
		require.Equal(t, errNilAccountNonceAndBalanceProvider, err)
	})

	t.Run("should return errNonceGap", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		txCache.txByHash.addTx(createTx([]byte("txHash1"), "alice", 1))
		txCache.txByHash.addTx(createTx([]byte("txHash2"), "alice", 5))

		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		blockBody := block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txHash1"),
						[]byte("txHash2"),
					},
				},
			},
		}

		accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 1, big.NewInt(20), true, nil
			},
		}

		err = tracker.OnProposedBlock([]byte("hash1"), &blockBody, &block.Header{
			Nonce:    uint64(0),
			PrevHash: []byte(fmt.Sprintf("prevHash%d", 0)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		}, accountsProvider, defaultLatestExecutedHash)

		require.Equal(t, errNonceGap, err)
	})

	t.Run("should return errDiscontinuousBreadcrumbs", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		txCache.txByHash.addTx(createTx([]byte("txHash1"), "alice", 1))
		txCache.txByHash.addTx(createTx([]byte("txHash2"), "alice", 2))
		txCache.txByHash.addTx(createTx([]byte("txHash3"), "alice", 4))
		txCache.txByHash.addTx(createTx([]byte("txHash4"), "alice", 5))

		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		blockBody1 := block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txHash1"),
						[]byte("txHash2"),
					},
				},
			},
		}

		blockBody2 := block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txHash3"),
						[]byte("txHash4"),
					},
				},
			},
		}

		accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 1, big.NewInt(20), true, nil
			},
		}

		err = tracker.OnProposedBlock([]byte("hash1"), &blockBody1, &block.Header{
			Nonce:    uint64(0),
			PrevHash: []byte(fmt.Sprintf("prevHash%d", 0)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		}, accountsProvider, defaultLatestExecutedHash)
		require.Nil(t, err)

		err = tracker.OnProposedBlock([]byte("hash2"), &blockBody2, &block.Header{
			Nonce:    uint64(1),
			PrevHash: []byte(fmt.Sprintf("hash%d", 1)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		}, accountsProvider, []byte("prevHash0"))
		require.Equal(t, errDiscontinuousBreadcrumbs, err)
	})

	t.Run("should return errExceededBalance because of fees", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		txCache.txByHash.addTx(createTx([]byte("txHash1"), "alice", 1).withTransferredValue(big.NewInt(5)))
		txCache.txByHash.addTx(createTx([]byte("txHash2"), "alice", 2).withTransferredValue(big.NewInt(5)))
		txCache.txByHash.addTx(createTx([]byte("txHash3"), "alice", 3).withTransferredValue(big.NewInt(5)))
		txCache.txByHash.addTx(createTx([]byte("txHash4"), "alice", 4).withTransferredValue(big.NewInt(6)))

		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		blockBody1 := block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txHash1"),
						[]byte("txHash2"),
					},
				},
			},
		}

		blockBody2 := block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txHash3"),
						[]byte("txHash4"),
					},
				},
			},
		}

		accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 1, big.NewInt(20), true, nil
			},
		}

		err = tracker.OnProposedBlock([]byte("hash1"), &blockBody1, &block.Header{
			Nonce:    uint64(0),
			PrevHash: []byte(fmt.Sprintf("prevHash%d", 0)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		}, accountsProvider, []byte("prevHash0"))
		require.Nil(t, err)

		err = tracker.OnProposedBlock([]byte("hash2"), &blockBody2, &block.Header{
			Nonce:    uint64(1),
			PrevHash: []byte(fmt.Sprintf("hash%d", 1)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		}, accountsProvider, []byte("prevHash0"))
		require.Equal(t, errExceededBalance, err)
	})

	t.Run("should return err from selection session", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("default err")

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		txCache.txByHash.addTx(createTx([]byte("txHash1"), "alice", 1))

		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		blockBody1 := block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txHash1"),
					},
				},
			},
		}

		accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 0, nil, false, expectedErr
			},
		}

		err = tracker.OnProposedBlock([]byte("hash1"), &blockBody1, &block.Header{
			Nonce:    uint64(0),
			PrevHash: []byte(fmt.Sprintf("prevHash%d", 0)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		}, accountsProvider, []byte("prevHash0"))
		require.Equal(t, expectedErr, err)
	})

	t.Run("should return errRootHashMismatch in case of different root hashes", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		txCache.txByHash.addTx(createTx([]byte("txHash1"), "alice", 1))

		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		blockBody1 := block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txHash1"),
					},
				},
			},
		}

		accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
			GetRootHashCalled: func() ([]byte, error) {
				return []byte("rootHash1"), nil
			},
		}

		tracker.latestRootHash = []byte("rootHashX")

		err = tracker.OnProposedBlock([]byte("hash1"), &blockBody1, &block.Header{
			Nonce:    uint64(0),
			PrevHash: []byte(fmt.Sprintf("prevHash%d", 0)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		}, accountsProvider, []byte("prevHash0"))
		require.Equal(t, errRootHashMismatch, err)
	})
}

func TestSelectionTracker_OnProposedBlockShouldWork(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	numOfBlocks := 20
	accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMock()

	proposeBlocks(t, numOfBlocks, tracker, accountsProvider)
	require.Equal(t, 20, len(tracker.blocks))
}

func TestSelectionTracker_OnProposedBlockWhenMaxTrackedBlocksIsReached(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 0, 3)
	require.Nil(t, err)

	numOfBlocks := 3
	accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMock()

	proposeBlocks(t, numOfBlocks, tracker, accountsProvider)

	// this one should not be added, it has new transactions, but it doesn't have execution results
	err = tracker.OnProposedBlock(
		[]byte("hashX"),
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txHash"),
					},
				},
			},
		},
		&block.Header{
			Nonce:    uint64(4),
			PrevHash: []byte(fmt.Sprintf("hash%d", 3)),
			RootHash: []byte("rootHash0"),
		},
		accountsProvider,
		defaultLatestExecutedHash,
	)
	require.Equal(t, errBadBlockWhileMaxTrackedBlocksReached, err)
	require.Equal(t, 3, len(tracker.blocks))

	// this one should be added because it has new execution results
	err = tracker.OnProposedBlock(
		[]byte(fmt.Sprintf("hash%d", 4)),
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{},
				},
			},
		},
		&block.HeaderV3{
			Nonce:    uint64(4),
			PrevHash: []byte(fmt.Sprintf("hash%d", 3)),
			ExecutionResults: []*block.ExecutionResult{
				{},
			},
		},
		accountsProvider,
		defaultLatestExecutedHash,
	)
	require.Nil(t, err)
	require.Equal(t, 4, len(tracker.blocks))

	// this one should be added because it's an empty block
	err = tracker.OnProposedBlock(
		[]byte(fmt.Sprintf("hash%d", 5)),
		&block.Body{},
		&block.HeaderV3{
			Nonce:    uint64(5),
			PrevHash: []byte(fmt.Sprintf("hash%d", 4)),
			ExecutionResults: []*block.ExecutionResult{
				{},
			},
		},
		accountsProvider,
		defaultLatestExecutedHash,
	)
	require.Nil(t, err)
	require.Equal(t, 5, len(tracker.blocks))
}

func Test_CompleteFlowShouldWork(t *testing.T) {
	t.Parallel()

	accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
		GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
			return 11, big.NewInt(8 * 100000 * oneBillion), true, nil
		},
	}

	config := ConfigSourceMe{
		Name:                        "test",
		NumChunks:                   16,
		NumBytesThreshold:           maxNumBytesUpperBound,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBoundTest,
		CountThreshold:              math.MaxUint32,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             true,
		NumItemsToPreemptivelyEvict: 1,
		TxCacheBoundsConfig:         createMockTxBoundsConfig(),
	}

	host := txcachemocks.NewMempoolHostMock()

	cache, err := NewTxCache(config, host, 0)
	require.Nil(t, err)

	txs := []*WrappedTransaction{
		// txs for first block
		createTx([]byte("txHash1"), "alice", 11).withValue(big.NewInt(0)).withRelayer([]byte("bob")).withGasLimit(100_000), // the fee is 100000000000000
		createTx([]byte("txHash2"), "alice", 12).withValue(big.NewInt(0)),                                                  // the fee is 50000000000000
		createTx([]byte("txHash3"), "alice", 13).withValue(big.NewInt(0)),
		createTx([]byte("txHash4"), "bob", 11).withValue(big.NewInt(0)),
		createTx([]byte("txHash5"), "carol", 11).withValue(big.NewInt(0)),
		createTx([]byte("txHash6"), "carol", 12).withValue(big.NewInt(0)).withRelayer([]byte("bob")).withGasLimit(100_000),
		createTx([]byte("txHash7"), "carol", 13).withValue(big.NewInt(0)).withRelayer([]byte("alice")).withGasLimit(100_000),
		createTx([]byte("txHash8"), "carol", 14).withValue(big.NewInt(0)).withRelayer([]byte("eve")).withGasLimit(100_000),

		// txs for second block
		createTx([]byte("txHash9"), "carol", 15).withValue(big.NewInt(0)),
		createTx([]byte("txHash10"), "eve", 11).withValue(big.NewInt(0)).withRelayer([]byte("bob")).withGasLimit(100_000),

		// tx to be selected
		createTx([]byte("txHash11"), "bob", 12).withValue(big.NewInt(0)),
		createTx([]byte("txHash12"), "carol", 13).withValue(big.NewInt(0)), // this one should not be selected
		createTx([]byte("txHash13"), "eve", 14).withValue(big.NewInt(0)),   // this one should not be selected
	}
	for _, tx := range txs {
		cache.AddTx(tx)
	}

	proposedBlock1 := [][]byte{
		[]byte("txHash1"),
		[]byte("txHash2"),
		[]byte("txHash3"),
		[]byte("txHash4"),
		[]byte("txHash5"),
		[]byte("txHash6"),
		[]byte("txHash7"),
		[]byte("txHash8"),
	}

	err = cache.OnProposedBlock(
		[]byte("hash1"),
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: proposedBlock1,
				},
			},
		},
		&block.Header{
			Nonce:    uint64(0),
			PrevHash: []byte("hash0"),
			RootHash: []byte("rootHash0"),
		},
		accountsProvider,
		defaultLatestExecutedHash,
	)
	require.Nil(t, err)

	expectedBreadcrumbs := map[string]*accountBreadcrumb{
		"alice": createExpectedBreadcrumb(true, 11, 13, big.NewInt(200000000000000)), // feeOf(txHash2) + feeOf(txHash3) + feeOf(txHash7)
		"bob":   createExpectedBreadcrumb(true, 11, 11, big.NewInt(250000000000000)), // feeOf(txHash1) + feeOf(txHash4) + feeOf(txHash6)
		"carol": createExpectedBreadcrumb(true, 11, 14, big.NewInt(50000000000000)),  // feeOf(txHash5)
		"eve":   createExpectedBreadcrumb(false, 0, 0, big.NewInt(100000000000000)),  // feeOf(txHash8)
	}

	require.Equal(t, 1, len(cache.tracker.blocks))
	tb, ok := cache.tracker.blocks["hash1"]
	require.True(t, ok)
	require.Equal(t, expectedBreadcrumbs, tb.breadcrumbsByAddress)

	// propose another block
	err = cache.OnProposedBlock(
		[]byte("hash2"),
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txHash9"),
						[]byte("txHash10"),
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
		defaultLatestExecutedHash,
	)
	require.Nil(t, err)

	expectedBreadcrumbs = map[string]*accountBreadcrumb{
		"bob":   createExpectedBreadcrumb(false, 0, 0, big.NewInt(100000000000000)),
		"carol": createExpectedBreadcrumb(true, 15, 15, big.NewInt(50000000000000)),
		"eve":   createExpectedBreadcrumb(true, 11, 11, big.NewInt(0)), // feeOf(txHash8)
	}
	require.Equal(t, 2, len(cache.tracker.blocks))
	tb, ok = cache.tracker.blocks["hash2"]
	require.True(t, ok)
	require.Equal(t, expectedBreadcrumbs, tb.breadcrumbsByAddress)

	selectionSession := &txcachemocks.SelectionSessionMock{
		GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
			return 11, big.NewInt(8 * 100000 * oneBillion), true, nil
		},
	}

	virtualSession, err := cache.tracker.deriveVirtualSelectionSession(selectionSession, 2, false)

	require.Nil(t, err)
	require.NotNil(t, virtualSession)

	expectedVirtualRecords := map[string]*virtualAccountRecord{
		"alice": createExpectedVirtualRecord(true, 14, big.NewInt(8*100000*oneBillion), big.NewInt(200000000000000)),
		"bob":   createExpectedVirtualRecord(true, 12, big.NewInt(8*100000*oneBillion), big.NewInt(350000000000000)),
		"carol": createExpectedVirtualRecord(true, 16, big.NewInt(8*100000*oneBillion), big.NewInt(100000000000000)),
		"eve":   createExpectedVirtualRecord(true, 12, big.NewInt(8*100000*oneBillion), big.NewInt(100000000000000)),
	}
	require.Equal(t, expectedVirtualRecords, virtualSession.virtualAccountsByAddress)

	// execute the first block
	err = cache.OnExecutedBlock(&block.Header{
		Nonce:    uint64(0),
		PrevHash: []byte("hash0"),
	}, []byte("rootHash0"))
	require.Nil(t, err)

	for _, txHash := range proposedBlock1 {
		cache.RemoveTxByHash(txHash)
	}

	// update the session nonce
	selectionSession = &txcachemocks.SelectionSessionMock{
		GetRootHashCalled: func() ([]byte, error) {
			return []byte("rootHash0"), nil
		},
		GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
			if string(address) == "alice" {
				return 14, big.NewInt(8 * 100000 * oneBillion), true, nil
			}
			if string(address) == "bob" {
				return 12, big.NewInt(8 * 100000 * oneBillion), true, nil
			}
			if string(address) == "carol" {
				return 15, big.NewInt(8 * 100000 * oneBillion), true, nil
			}

			return 11, big.NewInt(8 * 100000 * oneBillion), true, nil
		},
	}

	virtualSession, err = cache.tracker.deriveVirtualSelectionSession(selectionSession, 2, false)
	require.Nil(t, err)

	expectedVirtualRecords = map[string]*virtualAccountRecord{
		// bob was only relayer in the last proposed block (which is still tracked).
		// However, its initialNonce shouldn't remain uninitialized, so it's initialized with the session nonce.
		"bob":   createExpectedVirtualRecord(true, 12, big.NewInt(8*100000*oneBillion), big.NewInt(100000000000000)),
		"carol": createExpectedVirtualRecord(true, 16, big.NewInt(8*100000*oneBillion), big.NewInt(50000000000000)),
		"eve":   createExpectedVirtualRecord(true, 12, big.NewInt(8*100000*oneBillion), big.NewInt(0)),
	}
	require.Equal(t, expectedVirtualRecords, virtualSession.virtualAccountsByAddress)

	options, _ := holders.NewTxSelectionOptions(
		10_000_000_000,
		10,
		10,
		haveTimeTrueForSelection,
	)

	selectedTxs, _, err := cache.SelectTransactions(
		selectionSession,
		options,
		2,
	)
	require.Nil(t, err)
	require.Len(t, selectedTxs, 1)
	require.Equal(t, "txHash11", string(selectedTxs[0].TxHash))
}

func TestSelectionTracker_OnExecutedBlockShouldError(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	err = tracker.OnExecutedBlock(nil, []byte{})
	require.Equal(t, errNilBlockHeader, err)
}

func TestSelectionTracker_OnExecutedBlockShouldWork(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	numOfBlocks := 20
	accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMock()

	proposeBlocks(t, numOfBlocks, tracker, accountsProvider)
	require.Equal(t, numOfBlocks, len(tracker.blocks))

	executeBlocksConcurrently(t, numOfBlocks, tracker)
	require.Equal(t, 0, len(tracker.blocks))
	require.Equal(t, uint64(20), tracker.latestNonce)
	require.Equal(t, []byte("rootHash0"), tracker.latestRootHash)
}

func TestSelectionTracker_OnExecutedBlockShouldDeleteAllBlocksBelowSpecificNonce(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMock()
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	err = tracker.OnProposedBlock(
		[]byte(fmt.Sprintf("blockHash%d", 0)),
		&block.Body{},
		&block.Header{
			Nonce:    0,
			PrevHash: nil,
			RootHash: nil,
		},
		accountsProvider,
		defaultLatestExecutedHash,
	)
	require.Nil(t, err)

	err = tracker.OnProposedBlock(
		[]byte(fmt.Sprintf("blockHash%d", 1)),
		&block.Body{},
		&block.Header{
			Nonce:    1,
			PrevHash: []byte(fmt.Sprintf("blockHash%d", 0)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		},
		accountsProvider,
		defaultLatestExecutedHash,
	)
	require.Nil(t, err)

	err = tracker.OnProposedBlock(
		[]byte(fmt.Sprintf("blockHash%d", 2)),
		&block.Body{},
		&block.Header{
			Nonce:    2,
			PrevHash: []byte(fmt.Sprintf("blockHash%d", 1)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		},
		accountsProvider,
		defaultLatestExecutedHash,
	)
	require.Nil(t, err)

	err = tracker.OnProposedBlock(
		[]byte(fmt.Sprintf("blockHash%d", 3)),
		&block.Body{},
		&block.Header{
			Nonce:    3,
			PrevHash: []byte(fmt.Sprintf("blockHash%d", 2)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		},
		accountsProvider,
		defaultLatestExecutedHash,
	)
	require.Nil(t, err)

	err = tracker.OnExecutedBlock(&block.Header{
		Nonce:    2,
		PrevHash: []byte(fmt.Sprintf("blockHash%d", 1)),
	}, []byte(fmt.Sprintf("rootHash%d", 0)))
	require.Nil(t, err)
	require.Equal(t, 1, len(tracker.blocks))

	err = tracker.OnExecutedBlock(&block.Header{
		Nonce:    3,
		PrevHash: []byte(fmt.Sprintf("blockHash%d", 2)),
	}, []byte(fmt.Sprintf("rootHash%d", 0)))
	require.Nil(t, err)
	require.Equal(t, 0, len(tracker.blocks))
}

func TestSelectionTracker_updateLatestRoothash(t *testing.T) {
	t.Parallel()

	t.Run("latest roothash is nil", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		tracker.updateLatestRootHashNoLock(1, []byte("rootHash1"))
		require.Equal(t, uint64(1), tracker.latestNonce)
		require.Equal(t, []byte("rootHash1"), tracker.latestRootHash)
	})

	t.Run("root hash of block N after root hash of block N+1", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		tracker.updateLatestRootHashNoLock(2, []byte("rootHash2"))
		require.Equal(t, uint64(2), tracker.latestNonce)
		require.Equal(t, []byte("rootHash2"), tracker.latestRootHash)

		tracker.updateLatestRootHashNoLock(1, []byte("rootHash1"))
		require.Equal(t, uint64(2), tracker.latestNonce)
		require.Equal(t, []byte("rootHash2"), tracker.latestRootHash)
	})

	t.Run("root hash of block N + 1 after root hash of block N", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		tracker.updateLatestRootHashNoLock(1, []byte("rootHash1"))
		require.Equal(t, uint64(1), tracker.latestNonce)
		require.Equal(t, []byte("rootHash1"), tracker.latestRootHash)

		tracker.updateLatestRootHashNoLock(2, []byte("rootHash2"))
		require.Equal(t, uint64(2), tracker.latestNonce)
		require.Equal(t, []byte("rootHash2"), tracker.latestRootHash)
	})
}

func TestSelectionTracker_removeFromTrackedBlocks(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	expectedTrackedBlock := newTrackedBlock(1, []byte("blockHash2"), []byte("prevHash2"))
	b1 := newTrackedBlock(0, []byte("blockHash1"), []byte("prevHash1"))
	b2 := newTrackedBlock(0, []byte("blockHash3"), []byte("prevHash1"))

	tracker.blocks = map[string]*trackedBlock{
		string(b1.hash):                   b1,
		string(expectedTrackedBlock.hash): expectedTrackedBlock,
		string(b2.hash):                   b2,
	}

	require.Equal(t, 3, len(tracker.blocks))

	r := newTrackedBlock(0, nil, []byte("prevHash1"))

	err = tracker.removeUpToBlockNoLock(r)
	require.Nil(t, err)
	require.Equal(t, 1, len(tracker.blocks))

	_, ok := tracker.blocks[string(expectedTrackedBlock.hash)]
	require.True(t, ok)

	_, ok = tracker.blocks[string(b1.hash)]
	require.False(t, ok)

	_, ok = tracker.blocks[string(b2.hash)]
	require.False(t, ok)
}

func TestSelectionTracker_getChainOfTrackedBlocks(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	// create a slice of tracked block which aren't ordered
	tracker.blocks = make(map[string]*trackedBlock)
	b := newTrackedBlock(7, []byte("blockHash8"), []byte("blockHash7"))
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(5, []byte("blockHash6"), []byte("blockHash5"))
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(1, []byte("blockHash2"), []byte("blockHash1"))
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(0, []byte("blockHash1"), []byte("prevHash1"))
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(3, []byte("blockHash4"), []byte("blockHash3"))
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(2, []byte("blockHash3"), []byte("blockHash2"))
	tracker.blocks[string(b.hash)] = b

	// create a block with a wrong previous hash
	b = newTrackedBlock(4, []byte("blockHash5"), []byte("blockHashY"))
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(6, []byte("blockHash7"), []byte("blockHash6"))
	tracker.blocks[string(b.hash)] = b

	t.Run("check order to be from head to tail", func(t *testing.T) {
		t.Parallel()

		expectedTrackedBlockHashes := [][]byte{
			[]byte("blockHash2"),
			[]byte("blockHash3"),
			[]byte("blockHash4"),
		}

		actualChain, err := tracker.getChainOfTrackedPendingBlocks([]byte("blockHash1"), []byte("blockHash4"), 4)
		require.Nil(t, err)
		for i, returnedBlock := range actualChain {
			require.Equal(t, returnedBlock.hash, expectedTrackedBlockHashes[i])
		}
	})

	t.Run("should return expected tracked blocks and stop before nonce", func(t *testing.T) {
		t.Parallel()

		expectedTrackedBlockHashes := [][]byte{
			[]byte("blockHash6"),
			[]byte("blockHash7"),
		}

		actualChain, err := tracker.getChainOfTrackedPendingBlocks([]byte("blockHash5"), []byte("blockHash7"), 7)
		require.Nil(t, err)
		for i, returnedBlock := range actualChain {
			require.Equal(t, returnedBlock.hash, expectedTrackedBlockHashes[i])
		}
	})

	t.Run("should return errBlockNotFound because of prevHash not found", func(t *testing.T) {
		t.Parallel()

		actualChain, err := tracker.getChainOfTrackedPendingBlocks([]byte("blockHash4"), []byte("blockHash7"), 7)
		require.Equal(t, err, errBlockNotFound)
		require.Nil(t, actualChain)
	})

	t.Run("should return errDiscontinuousSequenceOfBlocks because of nonce", func(t *testing.T) {
		t.Parallel()

		actualChain, err := tracker.getChainOfTrackedPendingBlocks([]byte("blockHash5"), []byte("blockHash7"), 6)
		require.Equal(t, errDiscontinuousSequenceOfBlocks, err)
		require.Equal(t, 0, len(actualChain))
	})

}

func TestSelectionTracker_deriveVirtualSelectionSessionShouldErr(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	// should do the tests sequentially
	t.Run("get roothash returns error, should error", func(t *testing.T) {
		expectedErr := errors.New("expected err")

		session := txcachemocks.SelectionSessionMock{}
		session.GetRootHashCalled = func() ([]byte, error) {
			return nil, expectedErr
		}
		virtualSession, actualErr := tracker.deriveVirtualSelectionSession(&session, 0, false)
		require.Nil(t, virtualSession)
		require.Equal(t, expectedErr, actualErr)
	})
	t.Run("cannot do simulation error on wrong nonce, returns error", func(t *testing.T) {
		session := txcachemocks.SelectionSessionMock{}
		session.GetRootHashCalled = func() ([]byte, error) {
			return []byte("root hash"), nil
		}
		virtualSession, actualErr := tracker.deriveVirtualSelectionSession(&session, 10, true)
		require.Nil(t, virtualSession)
		require.Equal(t, errSimulateSelectionContextInvalid, actualErr)
	})
}

func TestSelectionTracker_canDoSimulateSelection(t *testing.T) {
	t.Parallel()

	t.Run("OK - matching nonce", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		tracker.latestNonce = 9
		require.Nil(t, err)
		require.True(t, tracker.canDoSimulateSelection(10))

		// matching on top of tracked blocks
		tracker.blocks["blockHash1"] = newTrackedBlock(10, []byte("blockHash1"), []byte("prevHash1"))
		require.True(t, tracker.canDoSimulateSelection(11))
	})
	t.Run("simulation trial with lower nonce than last nonce should return false", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		tracker.latestNonce = 9
		require.False(t, tracker.canDoSimulateSelection(8))
	})
	t.Run("simulation trial when having tracked blocks", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)
		tracker.latestNonce = 9
		tracker.blocks["blockHash1"] = newTrackedBlock(10, []byte("blockHash1"), []byte("prevHash1"))
		// same as last tracked block
		require.False(t, tracker.canDoSimulateSelection(10))

		// previous nonce to last tracked block
		tracker.blocks["blockHash2"] = newTrackedBlock(11, []byte("blockHash2"), []byte("blockHash1"))
		require.False(t, tracker.canDoSimulateSelection(10))

		// above and with gap compared to last tracked block
		require.False(t, tracker.canDoSimulateSelection(13))
	})
}

func TestSelectionTracker_deriveVirtualSelectionSessionShouldDeleteProposedBlocks(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	tracker.blocks = createDummyTrackedBlocks()
	require.Equal(t, 3, len(tracker.blocks))

	session := txcachemocks.SelectionSessionMock{}
	session.GetRootHashCalled = func() ([]byte, error) {
		return nil, nil
	}

	_, err = tracker.deriveVirtualSelectionSession(&session, 0, false)
	require.Nil(t, err)
	require.Equal(t, 0, len(tracker.blocks))
}

func TestSelectionTracker_deriveVirtualSelectionSessionShouldNotDeleteProposedBlocks(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)
	tracker.blocks = createDummyTrackedBlocks()
	require.Nil(t, err)
	require.Equal(t, 3, len(tracker.blocks))

	session := txcachemocks.SelectionSessionMock{}
	session.GetRootHashCalled = func() ([]byte, error) {
		return nil, nil
	}

	_, err = tracker.deriveVirtualSelectionSession(&session, 0, true)
	require.Nil(t, err)
	require.Equal(t, 3, len(tracker.blocks))
}

func TestSelectionTracker_validateTrackedBlocks(t *testing.T) {
	t.Parallel()

	t.Run("should return discontinuous breadcrumbs", func(t *testing.T) {
		t.Parallel()

		breadcrumbAlice1 := newAccountBreadcrumb(core.OptionalUint64{
			Value:    0,
			HasValue: true,
		})
		breadcrumbAlice1.lastNonce = core.OptionalUint64{
			Value:    4,
			HasValue: true,
		}

		breadcrumbAlice2 := newAccountBreadcrumb(core.OptionalUint64{
			Value:    6,
			HasValue: true,
		})
		breadcrumbAlice2.lastNonce = core.OptionalUint64{
			Value:    7,
			HasValue: true,
		}

		trackedBlocks := []*trackedBlock{
			{
				nonce:    0,
				hash:     []byte("hash1"),
				prevHash: []byte("prevHash1"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlice1,
				},
			},
			{
				nonce:    0,
				hash:     []byte("hash2"),
				prevHash: []byte("prevHash2"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlice2,
				},
			},
		}

		mockSelectionSession := txcachemocks.SelectionSessionMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 0, big.NewInt(20), true, nil
			},
		}

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.validateBreadcrumbsOfTrackedBlocks(trackedBlocks, &mockSelectionSession)
		require.Equal(t, errDiscontinuousBreadcrumbs, err)
	})

	t.Run("should return balance exceeded", func(t *testing.T) {
		t.Parallel()

		breadcrumbAlice1 := newAccountBreadcrumb(core.OptionalUint64{
			Value:    0,
			HasValue: true,
		})
		breadcrumbAlice1.accumulateConsumedBalance(big.NewInt(3))
		breadcrumbAlice1.lastNonce = core.OptionalUint64{
			Value:    4,
			HasValue: true,
		}

		breadcrumbAlice2 := newAccountBreadcrumb(core.OptionalUint64{
			Value:    5,
			HasValue: true,
		})
		breadcrumbAlice2.accumulateConsumedBalance(big.NewInt(3))
		breadcrumbAlice2.lastNonce = core.OptionalUint64{
			Value:    7,
			HasValue: true,
		}

		trackedBlocks := []*trackedBlock{
			{
				nonce:    0,
				hash:     []byte("hash1"),
				prevHash: []byte("prevHash1"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlice1,
				},
			},
			{
				nonce:    0,
				hash:     []byte("hash2"),
				prevHash: []byte("prevHash2"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlice2,
				},
			},
		}

		mockSelectionSession := txcachemocks.SelectionSessionMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 0, big.NewInt(5), true, nil
			},
		}

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.validateBreadcrumbsOfTrackedBlocks(trackedBlocks, &mockSelectionSession)
		require.Equal(t, errExceededBalance, err)
	})

	t.Run("should return nil", func(t *testing.T) {
		t.Parallel()

		breadcrumbAlice1 := newAccountBreadcrumb(core.OptionalUint64{
			Value:    0,
			HasValue: true,
		})
		breadcrumbAlice1.lastNonce = core.OptionalUint64{
			Value:    4,
			HasValue: true,
		}

		breadcrumbAlice2 := newAccountBreadcrumb(core.OptionalUint64{
			Value:    5,
			HasValue: true,
		})
		breadcrumbAlice2.accumulateConsumedBalance(big.NewInt(1))

		breadcrumbAlice2.lastNonce = core.OptionalUint64{
			Value:    7,
			HasValue: true,
		}

		trackedBlocks := []*trackedBlock{
			{
				nonce:    0,
				hash:     []byte("hash1"),
				prevHash: []byte("prevHash1"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlice1,
				},
			},
			{
				nonce:    0,
				hash:     []byte("hash2"),
				prevHash: []byte("prevHash2"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlice2,
				},
			},
		}

		mockSelectionSession := txcachemocks.SelectionSessionMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 0, big.NewInt(2), true, nil
			},
		}

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.validateBreadcrumbsOfTrackedBlocks(trackedBlocks, &mockSelectionSession)
		require.Nil(t, err)
	})
}

func TestSelectionTracker_addNewBlockNoLock(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	tb0 := newTrackedBlock(0, []byte("blockHash0"), []byte("blockHash"))
	tb1 := newTrackedBlock(1, []byte("blockHash1"), []byte("blockHash0"))
	tb2 := newTrackedBlock(1, []byte("blockHash2"), []byte("blockHash0"))

	err = tracker.addNewTrackedBlockNoLock([]byte("blockHash0"), tb0)
	require.Nil(t, err)
	require.Equal(t, len(tracker.blocks), 1)

	err = tracker.addNewTrackedBlockNoLock([]byte("blockHash1"), tb1)
	require.Nil(t, err)
	require.Equal(t, len(tracker.blocks), 2)

	err = tracker.addNewTrackedBlockNoLock([]byte("blockHash2"), tb2)
	require.Nil(t, err)
	require.Equal(t, len(tracker.blocks), 2)
}

func Test_getVirtualNonceOfAccount(t *testing.T) {
	t.Parallel()

	t.Run("should return errGlobalBreadcrumbDoesNotExist error", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		tracker.blocks["hash2"] = newTrackedBlock(0, []byte("hash2"), []byte("hash1"))

		_, _, err = tracker.getVirtualNonceOfAccountWithRootHash([]byte("alice"))
		require.Equal(t, errGlobalBreadcrumbDoesNotExist, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		breadcrumb := newAccountBreadcrumb(core.OptionalUint64{HasValue: true, Value: 10})
		err = breadcrumb.updateNonceRange(core.OptionalUint64{
			HasValue: true,
			Value:    20,
		})
		require.Nil(t, err)

		tb := newTrackedBlock(0, []byte("hash2"), []byte("hash1"))
		tb.breadcrumbsByAddress["alice"] = breadcrumb

		tracker.globalBreadcrumbsCompiler.updateOnAddedBlock(tb)

		nonce, _, err := tracker.getVirtualNonceOfAccountWithRootHash([]byte("alice"))
		require.Nil(t, err)
		require.Equal(t, uint64(21), nonce)
	})
}

func Test_isTransactionTracked(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 6)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
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
		defaultLatestExecutedHash,
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
		defaultLatestExecutedHash,
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
		defaultLatestExecutedHash,
	)
	require.Nil(t, err)

	t.Run("should return true", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHash1"), "alice", 11)
		tx2 := createTx([]byte("txHash6"), "eve", 11)

		require.True(t, txCache.tracker.IsTransactionTracked(tx1))
		require.True(t, txCache.tracker.IsTransactionTracked(tx2))
	})

	t.Run("should return false because out of range", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHashX"), "alice", 16)
		tx2 := createTx([]byte("txHashX"), "eve", 12)

		require.False(t, txCache.tracker.IsTransactionTracked(tx1))
		require.False(t, txCache.tracker.IsTransactionTracked(tx2))

	})

	t.Run("should return false because account is only relayer", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHashX"), "alice", 16)

		require.False(t, txCache.tracker.IsTransactionTracked(tx1))
	})

	t.Run("should return false because account is not tracked at all", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHash2"), "carol", 12)

		require.False(t, txCache.tracker.IsTransactionTracked(tx1))
	})

	t.Run("should return true for any transaction of sender with a tracked nonce", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHash7"), "eve", 12)

		require.False(t, txCache.tracker.IsTransactionTracked(tx1))
	})
}

func TestSelectionTracker_IsTransactionTracked(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 6)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
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
						[]byte("txHash6"),
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
		defaultLatestExecutedHash,
	)
	require.Nil(t, err)

	require.True(t, txCache.tracker.IsTransactionTracked(txs[0]))
	require.True(t, txCache.tracker.IsTransactionTracked(txs[1]))
	require.True(t, txCache.tracker.IsTransactionTracked(txs[2]))
	require.False(t, txCache.tracker.IsTransactionTracked(txs[3]))
	require.False(t, txCache.tracker.IsTransactionTracked(txs[4]))
	require.True(t, txCache.tracker.IsTransactionTracked(txs[5]))
	require.True(t, txCache.tracker.IsTransactionTracked(txs[6]))
}

func TestSelectionTracker_ResetTracker(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 6)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	txCache.tracker = tracker

	tracker.blocks = createDummyTrackedBlocks()
	tracker.globalBreadcrumbsCompiler.globalAccountBreadcrumbs = map[string]*globalAccountBreadcrumb{
		"alice": {},
		"bob":   {},
		"carol": {},
	}

	tracker.latestNonce = 10
	tracker.latestRootHash = []byte("rootHash0")

	require.Equal(t, []byte("rootHash0"), tracker.latestRootHash)
	require.Equal(t, uint64(10), tracker.latestNonce)

	require.Equal(t, 3, len(tracker.blocks))
	require.Equal(t, 3, len(tracker.globalBreadcrumbsCompiler.globalAccountBreadcrumbs))

	tracker.ResetTrackedBlocks()
	require.Equal(t, 0, len(tracker.blocks))
	require.Equal(t, 0, len(tracker.globalBreadcrumbsCompiler.globalAccountBreadcrumbs))

	require.Nil(t, tracker.latestRootHash)
	require.Equal(t, uint64(0), tracker.latestNonce)
}

func Test_getDimensionOfTrackedBlocks(t *testing.T) {
	t.Parallel()

	t.Run("should return the number of tracked blocks", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)
		txCache.tracker = tracker

		tracker.blocks = map[string]*trackedBlock{}
		trackerDiagnosis := tracker.getTrackerDiagnosis()
		require.Equal(t, uint64(0), trackerDiagnosis.GetNumTrackedBlocks())
		require.Equal(t, uint64(0), trackerDiagnosis.GetNumTrackedAccounts())

		tracker.blocks = createDummyTrackedBlocks()
		tracker.globalBreadcrumbsCompiler.globalAccountBreadcrumbs = map[string]*globalAccountBreadcrumb{
			"alice": {},
			"bob":   {},
		}

		trackerDiagnosis = tracker.getTrackerDiagnosis()
		require.Equal(t, uint64(3), trackerDiagnosis.GetNumTrackedBlocks())
		require.Equal(t, uint64(2), trackerDiagnosis.GetNumTrackedAccounts())
	})
}

func TestSelectionTracker_addNewTrackedBlockNoLock(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 6)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	txCache.tracker = tracker
	tracker.blocks = createDummyTrackedBlocks()

	require.Equal(t, 3, len(txCache.tracker.blocks))
	err = tracker.addNewTrackedBlockNoLock([]byte("hashX"), &trackedBlock{
		nonce: 1,
		hash:  []byte("hashX"),
	})
	require.Nil(t, err)

	require.Equal(t, 1, len(txCache.tracker.blocks))
	_, ok := txCache.tracker.blocks["hashX"]
	require.True(t, ok)
}

func TestSelectionTracker_removeBlocksAboveNonce(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 6)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	txCache.tracker = tracker

	tracker.blocks = createDummyTrackedBlocks()

	require.Equal(t, 3, len(txCache.tracker.blocks))
	err = tracker.removeBlocksAboveOrEqualToNonceNoLock(1)
	require.Nil(t, err)

	require.Equal(t, 0, len(txCache.tracker.blocks))
}

func TestSelectionTracker_MaxUniqueAccounts(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	st, err := NewSelectionTracker(txCache, 0, 100)
	require.Nil(t, err)

	count := maxAccountsPerBlock + 1
	txs := make([][]byte, count)

	for i := 0; i < count; i++ {
		sender := fmt.Sprintf("sender%d", i)
		hash := []byte(fmt.Sprintf("tx%d", i))

		wTx := &WrappedTransaction{
			Tx: &transaction.Transaction{
				SndAddr: []byte(sender),
				RcvAddr: []byte("receiver"),
				Nonce:   uint64(i),
			},
			TxHash: hash,
		}
		txCache.txByHash.addTx(wTx)
		txs[i] = hash
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{TxHashes: txs, Type: block.TxBlock},
		},
	}
	header := &block.Header{
		Nonce: 10,
	}

	accProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
		GetRootHashCalled: func() ([]byte, error) {
			return defaultLatestExecutedHash, nil
		},
	}

	err = st.OnProposedBlock([]byte("blockHash"), body, header, accProvider, defaultLatestExecutedHash)
	require.Equal(t, errToManyUniqueAccountsInBlock, err)
}

type twoBlockTrackerSetup struct {
	tracker       *selectionTracker
	blockNonce1   uint64
	blockNonce2   uint64
	blockHashPrev string
}

// buildTrackerWithTwoSharedSenderBlocks creates a tracker with two consecutive tracked blocks
// that share a common sender (alice) with contiguous nonce ranges and breadcrumbs.
// Block1 (nonce=100): alice nonces 5-7, balance=10
// Block2 (nonce=101): alice nonces 8-10, balance=10
func buildTrackerWithTwoSharedSenderBlocks(t *testing.T) *twoBlockTrackerSetup {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	aliceFirstNonceBlock1 := uint64(5)
	aliceLastNonceBlock1 := uint64(7)
	aliceFirstNonceBlock2 := uint64(8)
	aliceLastNonceBlock2 := uint64(10)
	balancePerBlock := big.NewInt(10)

	blockNonce1 := uint64(100)
	blockNonce2 := uint64(101)
	blockHash1 := "blockHash100"
	blockHash2 := "blockHash101"
	blockHashPrev := "blockHash99"
	alice := "alice"

	breadcrumbAlice1 := &accountBreadcrumb{
		firstNonce:      core.OptionalUint64{Value: aliceFirstNonceBlock1, HasValue: true},
		lastNonce:       core.OptionalUint64{Value: aliceLastNonceBlock1, HasValue: true},
		consumedBalance: new(big.Int).Set(balancePerBlock),
	}
	breadcrumbAlice2 := &accountBreadcrumb{
		firstNonce:      core.OptionalUint64{Value: aliceFirstNonceBlock2, HasValue: true},
		lastNonce:       core.OptionalUint64{Value: aliceLastNonceBlock2, HasValue: true},
		consumedBalance: new(big.Int).Set(balancePerBlock),
	}

	block1 := &trackedBlock{
		nonce:    blockNonce1,
		hash:     []byte(blockHash1),
		prevHash: []byte(blockHashPrev),
		breadcrumbsByAddress: map[string]*accountBreadcrumb{
			alice: breadcrumbAlice1,
		},
	}
	block2 := &trackedBlock{
		nonce:    blockNonce2,
		hash:     []byte(blockHash2),
		prevHash: []byte(blockHash1),
		breadcrumbsByAddress: map[string]*accountBreadcrumb{
			alice: breadcrumbAlice2,
		},
	}

	tracker.blocks[blockHash1] = block1
	tracker.blocks[blockHash2] = block2

	tracker.globalBreadcrumbsCompiler.updateOnAddedBlock(block1)
	tracker.globalBreadcrumbsCompiler.updateOnAddedBlock(block2)

	// Verify the setup is correct
	globalBreadcrumbs := tracker.globalBreadcrumbsCompiler.getGlobalBreadcrumbs()
	require.Equal(t, aliceFirstNonceBlock1, globalBreadcrumbs[alice].firstNonce.Value)
	require.Equal(t, aliceLastNonceBlock2, globalBreadcrumbs[alice].lastNonce.Value)
	expectedTotalBalance := new(big.Int).Mul(balancePerBlock, big.NewInt(2))
	require.Equal(t, expectedTotalBalance, globalBreadcrumbs[alice].consumedBalance)
	require.Equal(t, 2, len(tracker.blocks))
	require.Equal(t, uint64(1), tracker.globalBreadcrumbsCompiler.getNumGlobalBreadcrumbs())

	return &twoBlockTrackerSetup{
		tracker:       tracker,
		blockNonce1:   blockNonce1,
		blockNonce2:   blockNonce2,
		blockHashPrev: blockHashPrev,
	}
}

func TestSelectionTracker_removeUpToBlockNoLock_orderedRemoval(t *testing.T) {
	t.Parallel()

	setup := buildTrackerWithTwoSharedSenderBlocks(t)

	// Remove both blocks at once (nonce <= blockNonce2)
	searchedBlock := newTrackedBlock(setup.blockNonce2, nil, nil)
	err := setup.tracker.removeUpToBlockNoLock(searchedBlock)
	require.Nil(t, err)

	require.Equal(t, 0, len(setup.tracker.blocks))
	require.Equal(t, uint64(0), setup.tracker.globalBreadcrumbsCompiler.getNumGlobalBreadcrumbs())
}

func TestSelectionTracker_removeBlocksAboveOrEqualToNonceNoLock_orderedRemoval(t *testing.T) {
	t.Parallel()

	setup := buildTrackerWithTwoSharedSenderBlocks(t)

	// Same scenario as removeBlockEqualOrAboveNoLock but triggered from deriveVirtualSelectionSession path
	err := setup.tracker.removeBlocksAboveOrEqualToNonceNoLock(setup.blockNonce1)
	require.Nil(t, err)

	require.Equal(t, 0, len(setup.tracker.blocks))
	require.Equal(t, uint64(0), setup.tracker.globalBreadcrumbsCompiler.getNumGlobalBreadcrumbs())
}

func TestSelectionTracker_OnExecutedBlock_multipleBlocksWithSharedSender(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 6)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)
	txCache.tracker = tracker

	aliceInitialNonce := uint64(1)
	accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
		GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
			return aliceInitialNonce, big.NewInt(8 * 100000 * oneBillion), true, nil
		},
	}

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	txHash3 := []byte("txHash3")
	txHash4 := []byte("txHash4")
	blockHash1 := []byte("hash1")
	blockHash2 := []byte("hash2")
	rootHash0 := []byte("rootHash0")
	rootHash1 := []byte("rootHash1")
	alice := "alice"

	// Add transactions for alice across two blocks
	txCache.txByHash.addTx(createTx(txHash1, alice, 1))
	txCache.txByHash.addTx(createTx(txHash2, alice, 2))
	txCache.txByHash.addTx(createTx(txHash3, alice, 3))
	txCache.txByHash.addTx(createTx(txHash4, alice, 4))

	// Propose block 1 with alice's tx nonces 1-2
	err = tracker.OnProposedBlock(
		blockHash1,
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				{TxHashes: [][]byte{txHash1, txHash2}},
			},
		},
		&block.Header{
			Nonce:    1,
			PrevHash: defaultLatestExecutedHash,
			RootHash: rootHash0,
		},
		accountsProvider,
		defaultLatestExecutedHash,
	)
	require.Nil(t, err)

	// Propose block 2 with alice's tx nonces 3-4
	err = tracker.OnProposedBlock(
		blockHash2,
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				{TxHashes: [][]byte{txHash3, txHash4}},
			},
		},
		&block.Header{
			Nonce:    2,
			PrevHash: blockHash1,
			RootHash: rootHash0,
		},
		accountsProvider,
		defaultLatestExecutedHash,
	)
	require.Nil(t, err)

	require.Equal(t, 2, len(tracker.blocks))
	require.Equal(t, uint64(1), tracker.globalBreadcrumbsCompiler.getNumGlobalBreadcrumbs())

	// Execute block 2, which removes both blocks at once (nonce <= 2)
	err = tracker.OnExecutedBlock(&block.Header{
		Nonce:    2,
		PrevHash: blockHash1,
	}, rootHash1)
	require.Nil(t, err)

	require.Equal(t, 0, len(tracker.blocks))
	require.Equal(t, uint64(0), tracker.globalBreadcrumbsCompiler.getNumGlobalBreadcrumbs())
}

func TestSelectionTracker_removeUpToBlockNoLock_orderedRemovalWithRelayer(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
	require.Nil(t, err)

	// Test with an account that appears as both sender and relayer across blocks.
	// Bob is a sender in block1 and a relayer in block2.
	bobFirstNonce := uint64(5)
	bobLastNonce := uint64(7)
	bobSenderBalance := big.NewInt(10)
	bobRelayerBalance := big.NewInt(5)
	aliceFirstNonce := uint64(10)
	aliceLastNonce := uint64(12)
	aliceBalance := big.NewInt(10)

	blockNonce1 := uint64(100)
	blockNonce2 := uint64(101)
	blockHash1 := "blockHash100"
	blockHash2 := "blockHash101"
	blockHashPrev := "blockHash99"
	alice := "alice"
	bob := "bob"

	breadcrumbBobAsSender := &accountBreadcrumb{
		firstNonce:      core.OptionalUint64{Value: bobFirstNonce, HasValue: true},
		lastNonce:       core.OptionalUint64{Value: bobLastNonce, HasValue: true},
		consumedBalance: new(big.Int).Set(bobSenderBalance),
	}
	breadcrumbBobAsRelayer := &accountBreadcrumb{
		firstNonce:      core.OptionalUint64{Value: 0, HasValue: false},
		lastNonce:       core.OptionalUint64{Value: 0, HasValue: false},
		consumedBalance: new(big.Int).Set(bobRelayerBalance),
	}
	breadcrumbAlice := &accountBreadcrumb{
		firstNonce:      core.OptionalUint64{Value: aliceFirstNonce, HasValue: true},
		lastNonce:       core.OptionalUint64{Value: aliceLastNonce, HasValue: true},
		consumedBalance: new(big.Int).Set(aliceBalance),
	}

	block1 := &trackedBlock{
		nonce:    blockNonce1,
		hash:     []byte(blockHash1),
		prevHash: []byte(blockHashPrev),
		breadcrumbsByAddress: map[string]*accountBreadcrumb{
			bob: breadcrumbBobAsSender,
		},
	}
	block2 := &trackedBlock{
		nonce:    blockNonce2,
		hash:     []byte(blockHash2),
		prevHash: []byte(blockHash1),
		breadcrumbsByAddress: map[string]*accountBreadcrumb{
			bob:   breadcrumbBobAsRelayer,
			alice: breadcrumbAlice,
		},
	}

	tracker.blocks[blockHash1] = block1
	tracker.blocks[blockHash2] = block2

	tracker.globalBreadcrumbsCompiler.updateOnAddedBlock(block1)
	tracker.globalBreadcrumbsCompiler.updateOnAddedBlock(block2)

	// Remove both blocks
	searchedBlock := newTrackedBlock(blockNonce2, nil, nil)
	err = tracker.removeUpToBlockNoLock(searchedBlock)
	require.Nil(t, err)

	require.Equal(t, 0, len(tracker.blocks))
	require.Equal(t, uint64(0), tracker.globalBreadcrumbsCompiler.getNumGlobalBreadcrumbs())
}

func TestSelectionTracker_validateBreadcrumbsToleratesPredecessorDiscontinuity(t *testing.T) {
	t.Parallel()

	t.Run("should tolerate discontinuous breadcrumbs in predecessor block when new block has no txs for that account", func(t *testing.T) {
		t.Parallel()

		// Predecessor block has discontinuous breadcrumbs for alice (firstNonce=52, but sessionNonce=0)
		// New block has no breadcrumbs for alice \u2014 should be tolerated
		breadcrumbAlicePredecessor := newAccountBreadcrumb(core.OptionalUint64{
			Value:    52,
			HasValue: true,
		})
		breadcrumbAlicePredecessor.lastNonce = core.OptionalUint64{
			Value:    55,
			HasValue: true,
		}

		// Bob is continuous in both blocks
		breadcrumbBobPredecessor := newAccountBreadcrumb(core.OptionalUint64{
			Value:    0,
			HasValue: true,
		})
		breadcrumbBobPredecessor.lastNonce = core.OptionalUint64{
			Value:    2,
			HasValue: true,
		}

		breadcrumbBobNewBlock := newAccountBreadcrumb(core.OptionalUint64{
			Value:    3,
			HasValue: true,
		})
		breadcrumbBobNewBlock.lastNonce = core.OptionalUint64{
			Value:    5,
			HasValue: true,
		}

		trackedBlocks := []*trackedBlock{
			{
				nonce:    100,
				hash:     []byte("predecessorHash"),
				prevHash: []byte("prevHash"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlicePredecessor,
					"bob":   breadcrumbBobPredecessor,
				},
			},
			{
				nonce:    101,
				hash:     []byte("newBlockHash"),
				prevHash: []byte("predecessorHash"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"bob": breadcrumbBobNewBlock,
				},
			},
		}

		accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 0, big.NewInt(1000), true, nil
			},
		}

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.validateBreadcrumbsOfTrackedBlocks(trackedBlocks, accountsProvider)
		require.Nil(t, err)
	})

	t.Run("should reject when new block has breadcrumbs for an address that is discontinuous in predecessor", func(t *testing.T) {
		t.Parallel()

		// Predecessor block has discontinuous breadcrumbs for alice (firstNonce=52, but sessionNonce=0)
		breadcrumbAlicePredecessor := newAccountBreadcrumb(core.OptionalUint64{
			Value:    52,
			HasValue: true,
		})
		breadcrumbAlicePredecessor.lastNonce = core.OptionalUint64{
			Value:    55,
			HasValue: true,
		}

		// New block also has breadcrumbs for alice \u2014 should be rejected
		breadcrumbAliceNewBlock := newAccountBreadcrumb(core.OptionalUint64{
			Value:    56,
			HasValue: true,
		})
		breadcrumbAliceNewBlock.lastNonce = core.OptionalUint64{
			Value:    58,
			HasValue: true,
		}

		trackedBlocks := []*trackedBlock{
			{
				nonce:    100,
				hash:     []byte("predecessorHash"),
				prevHash: []byte("prevHash"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlicePredecessor,
				},
			},
			{
				nonce:    101,
				hash:     []byte("newBlockHash"),
				prevHash: []byte("predecessorHash"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAliceNewBlock,
				},
			},
		}

		accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 0, big.NewInt(1000), true, nil
			},
		}

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.validateBreadcrumbsOfTrackedBlocks(trackedBlocks, accountsProvider)
		require.Equal(t, errDiscontinuousBreadcrumbs, err)
	})

	t.Run("should still reject when new block itself has discontinuous breadcrumbs (no predecessor)", func(t *testing.T) {
		t.Parallel()

		// Single block (the new block) has discontinuous breadcrumbs for alice
		breadcrumbAlice := newAccountBreadcrumb(core.OptionalUint64{
			Value:    52,
			HasValue: true,
		})
		breadcrumbAlice.lastNonce = core.OptionalUint64{
			Value:    55,
			HasValue: true,
		}

		trackedBlocks := []*trackedBlock{
			{
				nonce:    100,
				hash:     []byte("newBlockHash"),
				prevHash: []byte("prevHash"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlice,
				},
			},
		}

		accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 0, big.NewInt(1000), true, nil
			},
		}

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.validateBreadcrumbsOfTrackedBlocks(trackedBlocks, accountsProvider)
		require.Equal(t, errDiscontinuousBreadcrumbs, err)
	})

	t.Run("should tolerate discontinuous breadcrumbs in multiple predecessors", func(t *testing.T) {
		t.Parallel()

		// Two predecessor blocks with discontinuous breadcrumbs for alice
		// New block has no breadcrumbs for alice \u2014 should be tolerated
		breadcrumbAlice1 := newAccountBreadcrumb(core.OptionalUint64{
			Value:    52,
			HasValue: true,
		})
		breadcrumbAlice1.lastNonce = core.OptionalUint64{
			Value:    55,
			HasValue: true,
		}

		breadcrumbAlice2 := newAccountBreadcrumb(core.OptionalUint64{
			Value:    56,
			HasValue: true,
		})
		breadcrumbAlice2.lastNonce = core.OptionalUint64{
			Value:    58,
			HasValue: true,
		}

		// Bob is continuous everywhere
		breadcrumbBob := newAccountBreadcrumb(core.OptionalUint64{
			Value:    0,
			HasValue: true,
		})
		breadcrumbBob.lastNonce = core.OptionalUint64{
			Value:    1,
			HasValue: true,
		}

		trackedBlocks := []*trackedBlock{
			{
				nonce:    100,
				hash:     []byte("predecessorHash1"),
				prevHash: []byte("prevHash"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlice1,
					"bob":   breadcrumbBob,
				},
			},
			{
				nonce:    101,
				hash:     []byte("predecessorHash2"),
				prevHash: []byte("predecessorHash1"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlice2,
				},
			},
			{
				nonce:                102,
				hash:                 []byte("newBlockHash"),
				prevHash:             []byte("predecessorHash2"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{},
			},
		}

		accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 0, big.NewInt(1000), true, nil
			},
		}

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.validateBreadcrumbsOfTrackedBlocks(trackedBlocks, accountsProvider)
		require.Nil(t, err)
	})
}

func TestSelectionTracker_SelectionSkipsDiscontinuousAccounts(t *testing.T) {
	t.Parallel()

	t.Run("selection should skip sender with discontinuous breadcrumbs and select other senders", func(t *testing.T) {
		t.Parallel()

		// Setup:
		// - alice has discontinuous breadcrumbs in a tracked block (firstNonce=52, sessionNonce=0)
		// - bob has continuous breadcrumbs (firstNonce=0, sessionNonce=0)
		// - Both have transactions in the pool
		// Expected: selection should skip alice's transactions, select bob's

		config := ConfigSourceMe{
			Name:                        "test",
			NumChunks:                   16,
			NumBytesThreshold:           maxNumBytesUpperBound,
			NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBoundTest,
			CountThreshold:              math.MaxUint32,
			CountPerSenderThreshold:     math.MaxUint32,
			EvictionEnabled:             false,
			NumItemsToPreemptivelyEvict: 1,
			TxCacheBoundsConfig:         createMockTxBoundsConfig(),
		}

		host := txcachemocks.NewMempoolHostMock()
		cache, err := NewTxCache(config, host, 0)
		require.Nil(t, err)

		// Add transactions for both alice and bob
		// Alice: nonces 0, 1, 2 (these would be selected if not blocked)
		// Bob: nonces 0, 1, 2
		txs := []*WrappedTransaction{
			createTx([]byte("txAlice0"), "alice", 0).withValue(big.NewInt(0)),
			createTx([]byte("txAlice1"), "alice", 1).withValue(big.NewInt(0)),
			createTx([]byte("txAlice2"), "alice", 2).withValue(big.NewInt(0)),
			createTx([]byte("txBob0"), "bob", 0).withValue(big.NewInt(0)),
			createTx([]byte("txBob1"), "bob", 1).withValue(big.NewInt(0)),
			createTx([]byte("txBob2"), "bob", 2).withValue(big.NewInt(0)),
		}

		for _, tx := range txs {
			cache.AddTx(tx)
		}

		// Manually set up a tracked block with discontinuous breadcrumbs for alice
		// alice: firstNonce=52 (but sessionNonce=0, so discontinuous)
		// bob: firstNonce=0 (continuous)
		breadcrumbAlice := &accountBreadcrumb{
			firstNonce:      core.OptionalUint64{Value: 52, HasValue: true},
			lastNonce:       core.OptionalUint64{Value: 55, HasValue: true},
			consumedBalance: big.NewInt(0),
		}
		breadcrumbBob := &accountBreadcrumb{
			firstNonce:      core.OptionalUint64{Value: 0, HasValue: true},
			lastNonce:       core.OptionalUint64{Value: 2, HasValue: true},
			consumedBalance: big.NewInt(0),
		}

		tb := &trackedBlock{
			nonce:    100,
			hash:     []byte("trackedBlockHash"),
			prevHash: []byte("prevBlockHash"),
			breadcrumbsByAddress: map[string]*accountBreadcrumb{
				"alice": breadcrumbAlice,
				"bob":   breadcrumbBob,
			},
		}

		cache.tracker.blocks["trackedBlockHash"] = tb
		cache.tracker.globalBreadcrumbsCompiler.updateOnAddedBlock(tb)
		cache.tracker.latestRootHash = []byte("rootHash0")
		cache.tracker.latestNonce = 99

		selectionSession := &txcachemocks.SelectionSessionMock{
			GetRootHashCalled: func() ([]byte, error) {
				return []byte("rootHash0"), nil
			},
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 0, big.NewInt(8 * 100000 * oneBillion), true, nil
			},
		}

		options, _ := holders.NewTxSelectionOptions(
			10_000_000_000,
			100,
			10,
			haveTimeTrueForSelection,
		)

		selectedTxs, _, err := cache.SelectTransactions(selectionSession, options, 101)
		require.Nil(t, err)

		// Alice should be skipped entirely (discontinuous breadcrumbs \u2192 blocked virtual record)
		// Bob's nonces 0-2 are in the tracked block, so selection starts from nonce 3+, but bob has no nonce 3
		// So no transactions should be selected for bob either (they're already proposed)
		// Let's verify alice's txs are NOT selected
		for _, tx := range selectedTxs {
			sender := string(tx.Tx.GetSndAddr())
			require.NotEqual(t, "alice", sender, "alice's transactions should not be selected due to discontinuous breadcrumbs")
		}
	})
}

func TestSelectionTracker_RecoveryFromDiscontinuousBreadcrumbs(t *testing.T) {
	t.Parallel()

	// End-to-end scenario:
	// 1. A tracked block has discontinuous breadcrumbs for alice
	// 2. Selection skips alice (creates blocked virtual record)
	// 3. OnProposedBlock succeeds (new block has no txs for alice, predecessor discontinuity is tolerated)
	// 4. OnExecutedBlock removes the stale predecessor block
	// 5. After cleanup, alice's transactions can be selected again

	accountNonces := map[string]uint64{
		"alice": 0,
		"bob":   0,
	}

	accountsProvider := &txcachemocks.AccountNonceAndBalanceProviderMock{
		GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
			nonce := accountNonces[string(address)]
			return nonce, big.NewInt(8 * 100000 * oneBillion), true, nil
		},
		GetRootHashCalled: func() ([]byte, error) {
			return []byte("rootHash0"), nil
		},
	}

	config := ConfigSourceMe{
		Name:                        "test",
		NumChunks:                   16,
		NumBytesThreshold:           maxNumBytesUpperBound,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBoundTest,
		CountThreshold:              math.MaxUint32,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             false,
		NumItemsToPreemptivelyEvict: 1,
		TxCacheBoundsConfig:         createMockTxBoundsConfig(),
	}

	host := txcachemocks.NewMempoolHostMock()
	cache, err := NewTxCache(config, host, 0)
	require.Nil(t, err)

	// Add alice's and bob's transactions
	txs := []*WrappedTransaction{
		createTx([]byte("txAlice0"), "alice", 0).withValue(big.NewInt(0)),
		createTx([]byte("txAlice1"), "alice", 1).withValue(big.NewInt(0)),
		createTx([]byte("txBob0"), "bob", 0).withValue(big.NewInt(0)),
		createTx([]byte("txBob1"), "bob", 1).withValue(big.NewInt(0)),
	}
	for _, tx := range txs {
		cache.AddTx(tx)
	}

	// Step 1: Manually create a stale tracked block with discontinuous breadcrumbs for alice
	// (simulates the effect of a block replacement that made alice's breadcrumbs stale)
	staleBlock := &trackedBlock{
		nonce:    100,
		hash:     []byte("staleBlockHash"),
		prevHash: []byte("prevBlockHash"),
		breadcrumbsByAddress: map[string]*accountBreadcrumb{
			"alice": {
				firstNonce:      core.OptionalUint64{Value: 52, HasValue: true},
				lastNonce:       core.OptionalUint64{Value: 55, HasValue: true},
				consumedBalance: big.NewInt(0),
			},
		},
	}

	cache.tracker.blocks["staleBlockHash"] = staleBlock
	cache.tracker.globalBreadcrumbsCompiler.updateOnAddedBlock(staleBlock)
	cache.tracker.latestRootHash = []byte("rootHash0")
	cache.tracker.latestNonce = 99

	// Step 2: Propose a new block with only bob's transactions
	// This should succeed because:
	// - alice's discontinuous breadcrumbs in predecessor are tolerated
	// - new block has no txs for alice
	cache.txByHash.addTx(createTx([]byte("txBob0_proposed"), "bob", 0).withValue(big.NewInt(0)))

	err = cache.OnProposedBlock(
		[]byte("newBlockHash"),
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txBob0_proposed"),
					},
				},
			},
		},
		&block.Header{
			Nonce:    101,
			PrevHash: []byte("staleBlockHash"),
			RootHash: []byte("rootHash0"),
		},
		accountsProvider,
		[]byte("prevBlockHash"),
	)
	require.Nil(t, err)
	require.Equal(t, 2, len(cache.tracker.blocks))

	// Step 3: Execute the stale block (simulates OnExecutedBlock removing it)
	err = cache.OnExecutedBlock(&block.Header{
		Nonce:    100,
		PrevHash: []byte("prevBlockHash"),
	}, []byte("rootHash1"))
	require.Nil(t, err)
	require.Equal(t, 1, len(cache.tracker.blocks))

	// Step 4: After the stale block is removed, alice's breadcrumbs are no longer in any tracked block
	// Now verify alice can be selected in the next selection
	selectionSession := &txcachemocks.SelectionSessionMock{
		GetRootHashCalled: func() ([]byte, error) {
			return []byte("rootHash1"), nil
		},
		GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
			nonce := accountNonces[string(address)]
			return nonce, big.NewInt(8 * 100000 * oneBillion), true, nil
		},
	}

	options, _ := holders.NewTxSelectionOptions(
		10_000_000_000,
		100,
		10,
		haveTimeTrueForSelection,
	)

	selectedTxs, _, err := cache.SelectTransactions(selectionSession, options, 102)
	require.Nil(t, err)

	// Alice should now be selectable since the stale block was removed
	hasAliceTx := false
	for _, tx := range selectedTxs {
		if string(tx.Tx.GetSndAddr()) == "alice" {
			hasAliceTx = true
			break
		}
	}
	require.True(t, hasAliceTx, "alice should be selectable after stale tracked block is removed")
}

func createDummyTrackedBlocks() map[string]*trackedBlock {
	return map[string]*trackedBlock{
		"hash1": {
			nonce: 1,
			hash:  []byte("hash1"),
		},
		"hash2": {
			nonce: 2,
			hash:  []byte("hash2"),
		},
		"hash3": {
			nonce: 3,
			hash:  []byte("hash3"),
		},
	}
}
