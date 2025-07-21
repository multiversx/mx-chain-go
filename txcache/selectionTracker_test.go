package txcache

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func createMockedHeaders(numOfHeaders int) []*block.Header {
	headers := make([]*block.Header, numOfHeaders)

	for i := 0; i < numOfHeaders; i++ {
		headers[i] = &block.Header{
			Nonce:    uint64(i),
			PrevHash: []byte(fmt.Sprintf("prevHash%d", i)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", i)),
		}
	}

	return headers
}

func proposeBlocksConcurrently(t *testing.T, numOfBlocks int, selectionTracker *selectionTracker, headers []*block.Header) {
	wg := sync.WaitGroup{}
	wg.Add(numOfBlocks)

	for i := 0; i < numOfBlocks; i++ {
		go func(index int) {
			defer wg.Done()

			err := selectionTracker.OnProposedBlock(
				[]byte(fmt.Sprintf("blockHash%d", index)),
				&block.Body{}, headers[index], nil, defaultBlockchainInfo)
			require.Nil(t, err)
		}(i)
	}

	wg.Wait()
}

func executeBlocksConcurrently(t *testing.T, numOfBlocks int, selectionTracker *selectionTracker, headers []*block.Header) {
	wg := sync.WaitGroup{}
	wg.Add(numOfBlocks)

	for i := 0; i < numOfBlocks; i++ {
		go func(index int) {
			defer wg.Done()

			err := selectionTracker.OnExecutedBlock(headers[index])
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
		_, err := NewSelectionTracker(txCache)
		require.Nil(t, err)
	})

	t.Run("should fail", func(t *testing.T) {
		t.Parallel()

		tracker, err := NewSelectionTracker(nil)
		require.Equal(t, errNilTxCache, err)
		require.Nil(t, tracker)
	})
}

func TestSelectionTracker_OnProposedBlockShouldErr(t *testing.T) {
	t.Parallel()

	t.Run("should err nil block hash", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		err = tracker.OnProposedBlock(nil, nil, nil, nil, nil)
		require.Equal(t, err, errNilBlockHash)
	})

	t.Run("should err nil header", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		err = tracker.OnProposedBlock([]byte("hash1"), nil, nil, nil, nil)
		require.Equal(t, err, errNilBlockBody)
	})

	t.Run("should err nil header", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		blockBody := block.Body{}
		err = tracker.OnProposedBlock([]byte("hash1"), &blockBody, nil, nil, nil)
		require.Equal(t, err, errNilHeaderHandler)
	})
}

func TestSelectionTracker_OnProposedBlockShouldWork(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache)
	require.Nil(t, err)

	numOfBlocks := 20
	headers := createMockedHeaders(numOfBlocks)
	proposeBlocksConcurrently(t, numOfBlocks, tracker, headers)
	require.Equal(t, 20, len(tracker.blocks))
}

func TestSelectionTracker_OnExecutedBlockShouldError(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache)
	require.Nil(t, err)

	err = tracker.OnExecutedBlock(nil)
	require.Equal(t, errNilHeaderHandler, err)
}

func TestSelectionTracker_OnExecutedBlockShouldWork(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache)
	require.Nil(t, err)

	numOfBlocks := 20
	headers := createMockedHeaders(numOfBlocks)

	proposeBlocksConcurrently(t, numOfBlocks, tracker, headers)
	require.Equal(t, numOfBlocks, len(tracker.blocks))

	executeBlocksConcurrently(t, numOfBlocks, tracker, headers)
	require.Equal(t, 0, len(tracker.blocks))
	require.Equal(t, uint64(19), tracker.latestNonce)
	require.Equal(t, []byte("rootHash19"), tracker.latestRootHash)
}

func TestSelectionTracker_updateLatestRoothash(t *testing.T) {
	t.Parallel()

	t.Run("latest roothash is nil", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		tracker.updateLatestRootHashNoLock(1, []byte("rootHash1"))
		require.Equal(t, uint64(1), tracker.latestNonce)
		require.Equal(t, []byte("rootHash1"), tracker.latestRootHash)
	})

	t.Run("root hash of block N after root hash of block N+1", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache)
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
		tracker, err := NewSelectionTracker(txCache)
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
	tracker, err := NewSelectionTracker(txCache)
	require.Nil(t, err)

	expectedTrackedBlock, _ := newTrackedBlock(1, []byte("blockHash2"), []byte("rootHash2"), []byte("prevHash2"), nil)
	b1, err := newTrackedBlock(0, []byte("blockHash1"), []byte("rootHash1"), []byte("prevHash1"), nil)
	require.Nil(t, err)

	b2, err := newTrackedBlock(0, []byte("blockHash3"), []byte("rootHash3"), []byte("prevHash1"), nil)
	require.Nil(t, err)

	tracker.blocks = []*trackedBlock{
		b1,
		expectedTrackedBlock,
		b2,
	}

	require.Equal(t, 3, len(tracker.blocks))

	r, err := newTrackedBlock(0, nil, nil, []byte("prevHash1"), nil)
	require.Nil(t, err)

	tracker.removeFromTrackedBlocksNoLock(r)
	require.Equal(t, 1, len(tracker.blocks))

	require.Equal(t, expectedTrackedBlock, tracker.blocks[0])
}

func TestSelectionTracker_nextBlock(t *testing.T) {
	t.Parallel()

	t.Run("should return next block", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		expectedNextBlock, err := newTrackedBlock(0, []byte("blockHash2"), []byte("rootHash2"), []byte("blockHash1"), nil)
		require.Nil(t, err)
		b1, err := newTrackedBlock(0, []byte("blockHash1"), []byte("rootHash1"), []byte("prevHash1"), nil)
		require.Nil(t, err)

		tracker.blocks = []*trackedBlock{
			b1,
			expectedNextBlock,
		}

		receivedNextBlock := tracker.findNextBlock([]byte("blockHash1"))
		require.Equal(t, expectedNextBlock, receivedNextBlock)
	})

	t.Run("should return nil", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		expectedNextBlock, err := newTrackedBlock(0, []byte("blockHash2"), []byte("rootHash2"), []byte("blockHash1"), nil)
		require.Nil(t, err)
		b1, err := newTrackedBlock(0, []byte("blockHash1"), []byte("rootHash1"), []byte("prevHash1"), nil)
		require.Nil(t, err)
		tracker.blocks = []*trackedBlock{
			b1,
			expectedNextBlock,
		}

		receivedNextBlock := tracker.findNextBlock([]byte("notExistingBlockHash"))
		require.Nil(t, receivedNextBlock)
	})
}

func TestSelectionTracker_getChainOfTrackedBlocks(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache)
	require.Nil(t, err)

	// create a slice of tracked block which aren't ordered
	tracker.blocks = make([]*trackedBlock, 0)
	b, err := newTrackedBlock(7, []byte("blockHash8"), []byte("rootHash8"), []byte("blockHash7"), nil)
	require.Nil(t, err)
	tracker.blocks = append(tracker.blocks, b)

	b, err = newTrackedBlock(5, []byte("blockHash6"), []byte("rootHash6"), []byte("blockHash5"), nil)
	require.Nil(t, err)
	tracker.blocks = append(tracker.blocks, b)

	b, err = newTrackedBlock(1, []byte("blockHash2"), []byte("rootHash2"), []byte("blockHash1"), nil)
	require.Nil(t, err)
	tracker.blocks = append(tracker.blocks, b)

	b, err = newTrackedBlock(0, []byte("blockHash1"), []byte("rootHash1"), []byte("prevHash1"), nil)
	require.Nil(t, err)
	tracker.blocks = append(tracker.blocks, b)

	b, err = newTrackedBlock(3, []byte("blockHash4"), []byte("rootHash4"), []byte("blockHash3"), nil)
	require.Nil(t, err)
	tracker.blocks = append(tracker.blocks, b)

	b, err = newTrackedBlock(2, []byte("blockHash3"), []byte("rootHash3"), []byte("blockHash2"), nil)
	require.Nil(t, err)
	tracker.blocks = append(tracker.blocks, b)

	b, err = newTrackedBlock(4, []byte("blockHash5"), []byte("rootHash5"), []byte("blockHash4"), nil)
	require.Nil(t, err)
	tracker.blocks = append(tracker.blocks, b)

	b, err = newTrackedBlock(6, []byte("blockHash7"), []byte("rootHash7"), []byte("blockHash6"), nil)
	require.Nil(t, err)
	tracker.blocks = append(tracker.blocks, b)

	t.Run("should return expected tracked blocks and stop before nonce", func(t *testing.T) {
		t.Parallel()

		expectedTrackedBlockHashes := [][]byte{
			[]byte("blockHash5"),
			[]byte("blockHash6"),
			[]byte("blockHash7"),
		}

		actualChain := tracker.getChainOfTrackedBlocks([]byte("blockHash4"), 7)
		for i, returnedBlock := range actualChain {
			require.Equal(t, returnedBlock.hash, expectedTrackedBlockHashes[i])
		}
	})

	t.Run("should return expected tracked blocks and stop because of nil block encountered", func(t *testing.T) {
		t.Parallel()

		expectedTrackedBlockHashes := [][]byte{
			[]byte("blockHash5"),
			[]byte("blockHash6"),
			[]byte("blockHash7"),
			[]byte("blockHash8"),
		}

		actualChain := tracker.getChainOfTrackedBlocks([]byte("blockHash4"), 12)
		for i, returnedBlock := range actualChain {
			require.Equal(t, returnedBlock.hash, expectedTrackedBlockHashes[i])
		}
	})

	t.Run("should return 0 blocks because prevHash not found", func(t *testing.T) {
		t.Parallel()

		actualChain := tracker.getChainOfTrackedBlocks([]byte("blockHashX"), 12)
		require.Equal(t, 0, len(actualChain))
	})

	t.Run("should return 0 blocks because of greater or equal nonce encountered", func(t *testing.T) {
		t.Parallel()

		actualChain := tracker.getChainOfTrackedBlocks([]byte("blockHash6"), 6)
		require.Equal(t, 0, len(actualChain))
	})

	t.Run("should return 1 block because of greater or equal once encountered", func(t *testing.T) {
		t.Parallel()

		actualChain := tracker.getChainOfTrackedBlocks([]byte("blockHash6"), 7)
		require.Equal(t, 1, len(actualChain))
	})
}

func TestSelectionTracker_deriveVirtualSelectionSessionShouldErr(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	tracker, err := NewSelectionTracker(txCache)
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
		tracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		actualResult := tracker.computeNumberOfTxsInMiniBlocks(blockBody.MiniBlocks)
		require.Equal(t, 6, actualResult)
	})
}

func TestSelectionTracker_getTransactionsFromBlock(t *testing.T) {
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

		selTracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		txs, err := selTracker.getTransactionsFromBlock(&blockBody)
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

		selTracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		txs, err := selTracker.getTransactionsFromBlock(&blockBody)
		require.Nil(t, txs)
		require.Equal(t, errNotFoundTx, err)
	})

}
