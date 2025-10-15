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
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func proposeBlocks(t *testing.T, numOfBlocks int, selectionTracker *selectionTracker, accountsProvider common.AccountNonceAndBalanceProvider) {
	blockchainInfo := holders.NewBlockchainInfo([]byte("hash0"), nil, 20)

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
			blockchainInfo,
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

			blockHeader := &block.Header{
				Nonce:    uint64(index),
				PrevHash: []byte(fmt.Sprintf("prevHash%d", index-1)),
				RootHash: []byte("rootHash0"),
			}
			err := selectionTracker.OnExecutedBlock(blockHeader)
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
		_, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)
	})

	t.Run("should fail because of nil TxCache", func(t *testing.T) {
		t.Parallel()

		tracker, err := NewSelectionTracker(nil, maxTrackedBlocks)
		require.Equal(t, errNilTxCache, err)
		require.Nil(t, tracker)
	})

	t.Run("should fail because of maxTrackedBlocks", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, 0)
		require.Equal(t, errInvalidMaxTrackedBlocks, err)
		require.Nil(t, tracker)
	})
}

func TestSelectionTracker_OnProposedBlockShouldErr(t *testing.T) {
	t.Parallel()

	t.Run("should err nil block hash", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.OnProposedBlock(nil, &block.Body{}, nil, nil, nil)
		require.Equal(t, errNilBlockHash, err)
	})

	t.Run("should err nil block header", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)

		blockBody := block.Body{}
		err = tracker.OnProposedBlock([]byte("hash1"), &blockBody, nil, nil, nil)
		require.Equal(t, errNilBlockHeader, err)
	})

	t.Run("should return errWrongTypeAssertion", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)

		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.OnProposedBlock([]byte("hash1"), nil, &block.Header{}, nil, defaultBlockchainInfo)

		require.Equal(t, errWrongTypeAssertion, err)
	})

	t.Run("should err nil accounts provider", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.OnProposedBlock([]byte("hash1"), &block.Body{}, &block.Header{}, nil, nil)
		require.Equal(t, errNilAccountNonceAndBalanceProvider, err)
	})

	t.Run("should return errNonceGap", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		txCache.txByHash.addTx(createTx([]byte("txHash1"), "alice", 1))
		txCache.txByHash.addTx(createTx([]byte("txHash2"), "alice", 5))

		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
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
		}, accountsProvider, defaultBlockchainInfo)

		require.Equal(t, errNonceGap, err)
	})

	t.Run("should return errDiscontinuousBreadcrumbs", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		txCache.txByHash.addTx(createTx([]byte("txHash1"), "alice", 1))
		txCache.txByHash.addTx(createTx([]byte("txHash2"), "alice", 2))
		txCache.txByHash.addTx(createTx([]byte("txHash3"), "alice", 4))
		txCache.txByHash.addTx(createTx([]byte("txHash4"), "alice", 5))

		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
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
		}, accountsProvider, holders.NewBlockchainInfo(
			[]byte(fmt.Sprintf("prevHash%d", 0)),
			nil,
			1,
		))
		require.Nil(t, err)

		err = tracker.OnProposedBlock([]byte("hash2"), &blockBody2, &block.Header{
			Nonce:    uint64(1),
			PrevHash: []byte(fmt.Sprintf("hash%d", 1)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		}, accountsProvider, holders.NewBlockchainInfo(
			[]byte("prevHash0"),
			[]byte("hash1"),
			2,
		))
		require.Equal(t, errDiscontinuousBreadcrumbs, err)
	})

	t.Run("should return errExceededBalance because of fees", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		txCache.txByHash.addTx(createTx([]byte("txHash1"), "alice", 1).withTransferredValue(big.NewInt(5)))
		txCache.txByHash.addTx(createTx([]byte("txHash2"), "alice", 2).withTransferredValue(big.NewInt(5)))
		txCache.txByHash.addTx(createTx([]byte("txHash3"), "alice", 3).withTransferredValue(big.NewInt(5)))
		txCache.txByHash.addTx(createTx([]byte("txHash4"), "alice", 4).withTransferredValue(big.NewInt(6)))

		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
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
		}, accountsProvider, holders.NewBlockchainInfo(
			[]byte(fmt.Sprintf("prevHash%d", 0)),
			nil,
			2,
		))
		require.Nil(t, err)

		err = tracker.OnProposedBlock([]byte("hash2"), &blockBody2, &block.Header{
			Nonce:    uint64(1),
			PrevHash: []byte(fmt.Sprintf("hash%d", 1)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		}, accountsProvider, holders.NewBlockchainInfo(
			[]byte(fmt.Sprintf("prevHash%d", 0)),
			nil,
			2))
		require.Equal(t, errExceededBalance, err)
	})

	t.Run("should return err from selection session", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("default err")

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		txCache.txByHash.addTx(createTx([]byte("txHash1"), "alice", 1))

		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
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
		}, accountsProvider, holders.NewBlockchainInfo([]byte("prevHash0"), nil, 1))
		require.Equal(t, expectedErr, err)
	})
}

func TestSelectionTracker_OnProposedBlockShouldWork(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
	require.Nil(t, err)

	numOfBlocks := 20
	accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMock()

	proposeBlocks(t, numOfBlocks, tracker, accountsProvider)
	require.Equal(t, 20, len(tracker.blocks))
}

func TestSelectionTracker_OnProposedBlockWhenMaxTrackedBlocksIsReached(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 3)
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
		holders.NewBlockchainInfo([]byte("hash0"), nil, 20),
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
		holders.NewBlockchainInfo([]byte("hash0"), nil, 20),
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
		holders.NewBlockchainInfo([]byte("hash0"), nil, 20),
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

	cache, err := NewTxCache(config, host)
	require.Nil(t, err)

	txs := []*WrappedTransaction{
		// txs for first block
		createTx([]byte("txHash1"), "alice", 11).withRelayer([]byte("bob")).withGasLimit(100_000), // the fee is 100000000000000
		createTx([]byte("txHash2"), "alice", 12),                                                  // the fee is 50000000000000
		createTx([]byte("txHash3"), "alice", 13),
		createTx([]byte("txHash4"), "bob", 11),
		createTx([]byte("txHash5"), "carol", 11),
		createTx([]byte("txHash6"), "carol", 12).withRelayer([]byte("bob")).withGasLimit(100_000),
		createTx([]byte("txHash7"), "carol", 13).withRelayer([]byte("alice")).withGasLimit(100_000),
		createTx([]byte("txHash8"), "carol", 14).withRelayer([]byte("eve")).withGasLimit(100_000),

		// txs for second block
		createTx([]byte("txHash9"), "carol", 15),
		createTx([]byte("txHash10"), "eve", 11).withRelayer([]byte("bob")).withGasLimit(100_000),

		// tx to be selected
		createTx([]byte("txHash11"), "bob", 12),
		createTx([]byte("txHash12"), "carol", 13), // this one should not be selected
		createTx([]byte("txHash13"), "eve", 14),   // this one should not be selected
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
		holders.NewBlockchainInfo([]byte("hash0"), []byte("hash0"), 0),
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
		holders.NewBlockchainInfo([]byte("hash0"), []byte("hash1"), 1),
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

	virtualSession, err := cache.tracker.deriveVirtualSelectionSession(selectionSession, 2)

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
		RootHash: []byte("rootHash0"),
	})
	require.Nil(t, err)

	for _, txHash := range proposedBlock1 {
		cache.RemoveTxByHash(txHash)
	}

	// update the session nonce
	selectionSession = &txcachemocks.SelectionSessionMock{
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

	virtualSession, err = cache.tracker.deriveVirtualSelectionSession(selectionSession, 2)
	require.Nil(t, err)

	expectedVirtualRecords = map[string]*virtualAccountRecord{
		// bob was only relayer in the last proposed block (which is still tracked).
		// However, its initialNonce shouldn't remain uninitialized, so it's initialized with the session nonce.
		"bob":   createExpectedVirtualRecord(true, 12, big.NewInt(8*100000*oneBillion), big.NewInt(100000000000000)),
		"carol": createExpectedVirtualRecord(true, 16, big.NewInt(8*100000*oneBillion), big.NewInt(50000000000000)),
		"eve":   createExpectedVirtualRecord(true, 12, big.NewInt(8*100000*oneBillion), big.NewInt(0)),
	}
	require.Equal(t, expectedVirtualRecords, virtualSession.virtualAccountsByAddress)

	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		10,
		selectionLoopMaximumDuration,
		10,
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
	tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
	require.Nil(t, err)

	err = tracker.OnExecutedBlock(nil)
	require.Equal(t, errNilBlockHeader, err)
}

func TestSelectionTracker_OnExecutedBlockShouldWork(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
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
	tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
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
		defaultBlockchainInfo,
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
		defaultBlockchainInfo,
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
		defaultBlockchainInfo,
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
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	err = tracker.OnExecutedBlock(&block.Header{
		Nonce:    2,
		PrevHash: []byte(fmt.Sprintf("blockHash%d", 1)),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(tracker.blocks))

	err = tracker.OnExecutedBlock(&block.Header{
		Nonce:    3,
		PrevHash: []byte(fmt.Sprintf("blockHash%d", 2)),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	})
	require.Nil(t, err)
	require.Equal(t, 0, len(tracker.blocks))
}

func TestSelectionTracker_updateLatestRoothash(t *testing.T) {
	t.Parallel()

	t.Run("latest roothash is nil", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)

		tracker.updateLatestRootHashNoLock(1, []byte("rootHash1"))
		require.Equal(t, uint64(1), tracker.latestNonce)
		require.Equal(t, []byte("rootHash1"), tracker.latestRootHash)
	})

	t.Run("root hash of block N after root hash of block N+1", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
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
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
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
	tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
	require.Nil(t, err)

	expectedTrackedBlock := newTrackedBlock(1, []byte("blockHash2"), []byte("rootHash2"), []byte("prevHash2"))
	b1 := newTrackedBlock(0, []byte("blockHash1"), []byte("rootHash1"), []byte("prevHash1"))
	b2 := newTrackedBlock(0, []byte("blockHash3"), []byte("rootHash3"), []byte("prevHash1"))

	tracker.blocks = map[string]*trackedBlock{
		string(b1.hash):                   b1,
		string(expectedTrackedBlock.hash): expectedTrackedBlock,
		string(b2.hash):                   b2,
	}

	require.Equal(t, 3, len(tracker.blocks))

	r := newTrackedBlock(0, nil, nil, []byte("prevHash1"))

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
	tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
	require.Nil(t, err)

	// create a slice of tracked block which aren't ordered
	tracker.blocks = make(map[string]*trackedBlock)
	b := newTrackedBlock(7, []byte("blockHash8"), []byte("rootHash8"), []byte("blockHash7"))
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(5, []byte("blockHash6"), []byte("rootHash6"), []byte("blockHash5"))
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(1, []byte("blockHash2"), []byte("rootHash2"), []byte("blockHash1"))
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(0, []byte("blockHash1"), []byte("rootHash1"), []byte("prevHash1"))
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(3, []byte("blockHash4"), []byte("rootHash4"), []byte("blockHash3"))
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(2, []byte("blockHash3"), []byte("rootHash3"), []byte("blockHash2"))
	tracker.blocks[string(b.hash)] = b

	// create a block with a wrong previous hash
	b = newTrackedBlock(4, []byte("blockHash5"), []byte("rootHash5"), []byte("blockHashY"))
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(6, []byte("blockHash7"), []byte("rootHash7"), []byte("blockHash6"))
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
	tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
	require.Nil(t, err)

	expectedErr := errors.New("expected err")

	session := txcachemocks.SelectionSessionMock{}
	session.GetRootHashCalled = func() ([]byte, error) {
		return nil, expectedErr
	}
	virtualSession, actualErr := tracker.deriveVirtualSelectionSession(&session, 0)
	require.Nil(t, virtualSession)
	require.Equal(t, expectedErr, actualErr)
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
				rootHash: []byte("rootHash1"),
				prevHash: []byte("prevHash1"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlice1,
				},
			},
			{
				nonce:    0,
				hash:     []byte("hash2"),
				rootHash: []byte("rootHash2"),
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
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
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
				rootHash: []byte("rootHash1"),
				prevHash: []byte("prevHash1"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlice1,
				},
			},
			{
				nonce:    0,
				hash:     []byte("hash2"),
				rootHash: []byte("rootHash2"),
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
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
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
				rootHash: []byte("rootHash1"),
				prevHash: []byte("prevHash1"),
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": breadcrumbAlice1,
				},
			},
			{
				nonce:    0,
				hash:     []byte("hash2"),
				rootHash: []byte("rootHash2"),
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
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.validateBreadcrumbsOfTrackedBlocks(trackedBlocks, &mockSelectionSession)
		require.Nil(t, err)
	})
}

func TestSelectionTracker_addNewBlockNoLock(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
	require.Nil(t, err)

	tb0 := newTrackedBlock(0, []byte("blockHash0"), []byte("rootHash0"), []byte("blockHash"))
	tb1 := newTrackedBlock(1, []byte("blockHash1"), []byte("rootHash0"), []byte("blockHash0"))
	tb2 := newTrackedBlock(1, []byte("blockHash2"), []byte("rootHash0"), []byte("blockHash0"))

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
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)

		tracker.blocks["hash2"] = newTrackedBlock(0, []byte("hash2"), []byte("rootHash0"), []byte("hash1"))

		_, _, err = tracker.getVirtualNonceOfAccountWithRootHash([]byte("alice"))
		require.Equal(t, errGlobalBreadcrumbDoesNotExist, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)

		breadcrumb := newAccountBreadcrumb(core.OptionalUint64{HasValue: true, Value: 10})
		err = breadcrumb.updateNonceRange(core.OptionalUint64{
			HasValue: true,
			Value:    20,
		})
		require.Nil(t, err)

		tb := newTrackedBlock(0, []byte("hash2"), []byte("rootHash0"), []byte("hash1"))
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

		require.True(t, txCache.tracker.isTransactionTracked(tx1))
		require.True(t, txCache.tracker.isTransactionTracked(tx2))
	})

	t.Run("should return false because out of range", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHashX"), "alice", 16)
		tx2 := createTx([]byte("txHashX"), "eve", 12)

		require.False(t, txCache.tracker.isTransactionTracked(tx1))
		require.False(t, txCache.tracker.isTransactionTracked(tx2))

	})

	t.Run("should return false because account is only relayer", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHashX"), "alice", 16)

		require.False(t, txCache.tracker.isTransactionTracked(tx1))
	})

	t.Run("should return false because account is not tracked at all", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHash2"), "carol", 12)

		require.False(t, txCache.tracker.isTransactionTracked(tx1))
	})

	t.Run("should return true for any transaction of sender with a tracked nonce", func(t *testing.T) {
		t.Parallel()

		tx1 := createTx([]byte("txHash7"), "eve", 12)

		require.False(t, txCache.tracker.isTransactionTracked(tx1))
	})
}

func TestSelectionTracker_GetBulkOfUntrackedTransactions(t *testing.T) {
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

	bulk := tracker.GetBulkOfUntrackedTransactions(txs)
	require.Len(t, bulk, 4)
}

func TestSelectionTracker_ResetTracker(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 6)
	tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
	require.Nil(t, err)

	txCache.tracker = tracker

	tracker.blocks = map[string]*trackedBlock{
		"hash1": {},
		"hash2": {},
	}

	tracker.globalBreadcrumbsCompiler.globalAccountBreadcrumbs = map[string]*globalAccountBreadcrumb{
		"alice": {},
		"bob":   {},
		"carol": {},
	}

	tracker.latestNonce = 10
	tracker.latestRootHash = []byte("rootHash0")

	require.Equal(t, []byte("rootHash0"), tracker.latestRootHash)
	require.Equal(t, uint64(10), tracker.latestNonce)

	require.Equal(t, 2, len(tracker.blocks))
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
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)
		txCache.tracker = tracker

		tracker.blocks = map[string]*trackedBlock{}
		numBlocks, numBreadcrumbs := tracker.getTrackerDiagnosis()
		require.Equal(t, uint64(0), numBlocks)
		require.Equal(t, uint64(0), numBreadcrumbs)

		tracker.blocks = map[string]*trackedBlock{
			"hash1": {},
			"hash2": {},
			"hash3": {},
		}
		tracker.globalBreadcrumbsCompiler.globalAccountBreadcrumbs = map[string]*globalAccountBreadcrumb{
			"alice": {},
			"bob":   {},
		}

		numBlocks, numBreadcrumbs = tracker.getTrackerDiagnosis()
		require.Equal(t, uint64(3), numBlocks)
		require.Equal(t, uint64(2), numBreadcrumbs)
	})
}
