package txcache

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func proposeBlocks(t *testing.T, numOfBlocks int, selectionTracker *selectionTracker, accountsProvider AccountNonceAndBalanceProvider) {
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

		err = tracker.OnProposedBlock(nil, nil, nil, nil, nil)
		require.Equal(t, errNilBlockHash, err)
	})

	t.Run("should err nil header", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)

		err = tracker.OnProposedBlock([]byte("hash1"), nil, nil, nil, nil)
		require.Equal(t, errNilBlockBody, err)
	})

	t.Run("should err nil block header", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)

		blockBody := block.Body{}
		err = tracker.OnProposedBlock([]byte("hash1"), &blockBody, nil, nil, nil)
		require.Equal(t, errNilHeaderHandler, err)
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

func TestSelectionTracker_OnProposedBlockShouldCleanWhenMaxTrackedBlocksIsReached(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, 3)
	require.Nil(t, err)

	numOfBlocks := 3
	accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMock()

	proposeBlocks(t, numOfBlocks, tracker, accountsProvider)
	require.Equal(t, 0, len(tracker.blocks))
}

func TestSelectionTracker_OnExecutedBlockShouldError(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
	require.Nil(t, err)

	err = tracker.OnExecutedBlock(nil)
	require.Equal(t, errNilHeaderHandler, err)
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
	require.Nil(t, err)

	tracker.blocks = map[string]*trackedBlock{
		string(b1.hash):                   b1,
		string(expectedTrackedBlock.hash): expectedTrackedBlock,
		string(b2.hash):                   b2,
	}

	require.Equal(t, 3, len(tracker.blocks))

	r := newTrackedBlock(0, nil, nil, []byte("prevHash1"))
	require.Nil(t, err)

	tracker.removeFromTrackedBlocksNoLock(r)
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
	require.Nil(t, err)
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(5, []byte("blockHash6"), []byte("rootHash6"), []byte("blockHash5"))
	require.Nil(t, err)
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(1, []byte("blockHash2"), []byte("rootHash2"), []byte("blockHash1"))
	require.Nil(t, err)
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(0, []byte("blockHash1"), []byte("rootHash1"), []byte("prevHash1"))
	require.Nil(t, err)
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(3, []byte("blockHash4"), []byte("rootHash4"), []byte("blockHash3"))
	require.Nil(t, err)
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(2, []byte("blockHash3"), []byte("rootHash3"), []byte("blockHash2"))
	require.Nil(t, err)
	tracker.blocks[string(b.hash)] = b

	// create a block with a wrong previous hash
	b = newTrackedBlock(4, []byte("blockHash5"), []byte("rootHash5"), []byte("blockHashY"))
	require.Nil(t, err)
	tracker.blocks[string(b.hash)] = b

	b = newTrackedBlock(6, []byte("blockHash7"), []byte("rootHash7"), []byte("blockHash6"))
	require.Nil(t, err)
	tracker.blocks[string(b.hash)] = b

	t.Run("check order to be from head to tail", func(t *testing.T) {
		t.Parallel()

		expectedTrackedBlockHashes := [][]byte{
			[]byte("blockHash2"),
			[]byte("blockHash3"),
			[]byte("blockHash4"),
		}

		actualChain, err := tracker.getChainOfTrackedBlocks([]byte("blockHash1"), []byte("blockHash4"), 4)
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

		actualChain, err := tracker.getChainOfTrackedBlocks([]byte("blockHash5"), []byte("blockHash7"), 7)
		require.Nil(t, err)
		for i, returnedBlock := range actualChain {
			require.Equal(t, returnedBlock.hash, expectedTrackedBlockHashes[i])
		}
	})

	t.Run("should return errPreviousBlockNotFound because of prevHash not found", func(t *testing.T) {
		t.Parallel()

		actualChain, err := tracker.getChainOfTrackedBlocks([]byte("blockHash4"), []byte("blockHash7"), 7)
		require.Equal(t, err, errPreviousBlockNotFound)
		require.Nil(t, actualChain)
	})

	t.Run("should return errDiscontinuousSequenceOfBlocks because of nonce", func(t *testing.T) {
		t.Parallel()

		actualChain, err := tracker.getChainOfTrackedBlocks([]byte("blockHash5"), []byte("blockHash7"), 6)
		require.Equal(t, errDiscontinuousSequenceOfBlocks, err)
		require.Equal(t, 0, len(actualChain))
	})

}

func TestSelectionTracker_reverseOrderOfBlocks(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
	require.Nil(t, err)

	chainOfTrackedBlocks := []*trackedBlock{
		{
			nonce: 0,
			hash:  []byte("hash0"),
		},
		{
			nonce: 1,
			hash:  []byte("hash1"),
		},
		{
			nonce: 2,
			hash:  []byte("hash2"),
		},
	}

	reversedChainOfTrackedBlocks := tracker.reverseOrderOfBlocks(chainOfTrackedBlocks)
	require.Equal(t, len(chainOfTrackedBlocks), len(reversedChainOfTrackedBlocks))

	for i := 0; i < len(reversedChainOfTrackedBlocks); i++ {
		require.Equal(t, chainOfTrackedBlocks[len(chainOfTrackedBlocks)-i-1], reversedChainOfTrackedBlocks[i])
	}
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
	virtualSession, actualErr := tracker.deriveVirtualSelectionSession(&session, defaultBlockchainInfo)
	require.Nil(t, virtualSession)
	require.Equal(t, expectedErr, actualErr)
}

func TestSelectionTracker_computeNumberOfTxsInMiniBlocks(t *testing.T) {
	t.Parallel()

	t.Run("should return the right number of txs", func(t *testing.T) {
		blockBody := block.Body{MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{
					[]byte("txHash1"),
					[]byte("txHash2"),
				},
			},
			{
				TxHashes: [][]byte{
					[]byte("txHash3"),
				},
			},
			{
				TxHashes: [][]byte{
					[]byte("txHash4"),
					[]byte("txHash5"),
					[]byte("txHash6"),
				},
			},
		}}

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)

		actualResult := tracker.computeNumberOfTxsInMiniBlocks(blockBody.MiniBlocks)
		require.Equal(t, 6, actualResult)
	})
}

func TestSelectionTracker_getTransactionsInBlock(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		blockBody := block.Body{MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{
					[]byte("txHash1"),
					[]byte("txHash2"),
				},
			},
			{
				TxHashes: [][]byte{
					[]byte("txHash3"),
				},
			},
		}}

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		txCache.txByHash = newTxByHashMap(1)

		txCache.txByHash.addTx(createTx([]byte("txHash1"), "alice", 1))
		txCache.txByHash.addTx(createTx([]byte("txHash2"), "alice", 2))
		txCache.txByHash.addTx(createTx([]byte("txHash3"), "alice", 3))

		selTracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)

		txs, err := selTracker.getTransactionsInBlock(&blockBody)
		require.Nil(t, err)
		require.Equal(t, 3, len(txs))
	})

	t.Run("should fail", func(t *testing.T) {
		blockBody := block.Body{MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{
					[]byte("txHash1"),
					[]byte("txHash2"),
				},
			},
			{
				TxHashes: [][]byte{
					[]byte("txHash3"),
				},
			},
		}}

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		txCache.txByHash = newTxByHashMap(1)

		txCache.txByHash.addTx(createTx([]byte("txHash1"), "alice", 1))
		txCache.txByHash.addTx(createTx([]byte("txHash2"), "alice", 2))

		selTracker, err := NewSelectionTracker(txCache, maxTrackedBlocks)
		require.Nil(t, err)

		txs, err := selTracker.getTransactionsInBlock(&blockBody)
		require.Nil(t, txs)
		require.Equal(t, errNotFoundTx, err)
	})
}

func TestSelectionTracker_validateTrackedBlocks(t *testing.T) {
	t.Parallel()

	t.Run("should return discontinuous breadcrumbs", func(t *testing.T) {
		t.Parallel()

		breadcrumbAlice1 := newAccountBreadcrumb(core.OptionalUint64{
			Value:    0,
			HasValue: true,
		}, nil)
		breadcrumbAlice1.lastNonce = core.OptionalUint64{
			Value:    4,
			HasValue: true,
		}

		breadcrumbAlice2 := newAccountBreadcrumb(core.OptionalUint64{
			Value:    6,
			HasValue: true,
		}, nil)
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
		}, big.NewInt(3))
		breadcrumbAlice1.lastNonce = core.OptionalUint64{
			Value:    4,
			HasValue: true,
		}

		breadcrumbAlice2 := newAccountBreadcrumb(core.OptionalUint64{
			Value:    5,
			HasValue: true,
		}, big.NewInt(3))
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
		}, nil)
		breadcrumbAlice1.lastNonce = core.OptionalUint64{
			Value:    4,
			HasValue: true,
		}

		breadcrumbAlice2 := newAccountBreadcrumb(core.OptionalUint64{
			Value:    5,
			HasValue: true,
		}, big.NewInt(1))
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

	tb1 := newTrackedBlock(0, []byte("blockHash1"), []byte("rootHash0"), []byte("blockHash0"))

	tb2 := newTrackedBlock(0, []byte("blockHash2"), []byte("rootHash0"), []byte("blockHash0"))

	tracker.addNewTrackedBlockNoLock([]byte("blockHash1"), tb1)
	require.Equal(t, len(tracker.blocks), 1)

	tracker.addNewTrackedBlockNoLock([]byte("blockHash1"), tb2)
	require.Equal(t, len(tracker.blocks), 1)
}
