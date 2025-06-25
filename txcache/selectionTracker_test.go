package txcache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
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

			err := selectionTracker.OnProposedBlock([]byte(fmt.Sprintf("blockHash%d", index)), &block.Body{}, headers[index])
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

		err = tracker.OnProposedBlock(nil, nil, nil)
		require.Equal(t, err, errNilBlockHash)
	})

	t.Run("should err nil header", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		err = tracker.OnProposedBlock([]byte("hash1"), nil, nil)
		require.Equal(t, err, errNilBlockBody)
	})

	t.Run("should err nil header", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		tracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		blockBody := block.Body{}
		err = tracker.OnProposedBlock([]byte("hash1"), &blockBody, nil)
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
	selTracker, err := NewSelectionTracker(txCache)
	require.Nil(t, err)

	numOfBlocks := 20
	headers := createMockedHeaders(numOfBlocks)

	proposeBlocksConcurrently(t, numOfBlocks, selTracker, headers)
	require.Equal(t, numOfBlocks, len(selTracker.blocks))

	executeBlocksConcurrently(t, numOfBlocks, selTracker, headers)
	require.Equal(t, 0, len(selTracker.blocks))
	require.Equal(t, uint64(19), selTracker.latestNonce)
	require.Equal(t, []byte("rootHash19"), selTracker.latestRootHash)
}

func TestSelectionTracker_updateLatestRoothash(t *testing.T) {
	t.Parallel()

	t.Run("latest roothash is nil", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		selTracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		selTracker.updateLatestRootHashNoLock(1, []byte("rootHash1"))
		require.Equal(t, uint64(1), selTracker.latestNonce)
		require.Equal(t, []byte("rootHash1"), selTracker.latestRootHash)
	})

	t.Run("root hash of block N after root hash of block N+1", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		selTracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		selTracker.updateLatestRootHashNoLock(2, []byte("rootHash2"))
		require.Equal(t, uint64(2), selTracker.latestNonce)
		require.Equal(t, []byte("rootHash2"), selTracker.latestRootHash)

		selTracker.updateLatestRootHashNoLock(1, []byte("rootHash1"))
		require.Equal(t, uint64(2), selTracker.latestNonce)
		require.Equal(t, []byte("rootHash2"), selTracker.latestRootHash)
	})

	t.Run("root hash of block N + 1 after root hash of block N", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		selTracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		selTracker.updateLatestRootHashNoLock(1, []byte("rootHash1"))
		require.Equal(t, uint64(1), selTracker.latestNonce)
		require.Equal(t, []byte("rootHash1"), selTracker.latestRootHash)

		selTracker.updateLatestRootHashNoLock(2, []byte("rootHash2"))
		require.Equal(t, uint64(2), selTracker.latestNonce)
		require.Equal(t, []byte("rootHash2"), selTracker.latestRootHash)
	})
}

func TestSelectionTracker_removeFromTrackedBlocks(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	selTracker, err := NewSelectionTracker(txCache)
	require.Nil(t, err)

	expectedTrackedBlock := newTrackedBlock(1, []byte("blockHash2"), []byte("rootHash2"), []byte("prevHash2"))

	selTracker.blocks = []*trackedBlock{
		newTrackedBlock(0, []byte("blockHash1"), []byte("rootHash1"), []byte("prevHash1")),
		expectedTrackedBlock,
		newTrackedBlock(0, []byte("blockHash3"), []byte("rootHash3"), []byte("prevHash1")),
	}

	require.Equal(t, 3, len(selTracker.blocks))
	selTracker.removeFromTrackedBlocksNoLock(newTrackedBlock(0, nil, nil, []byte("prevHash1")))
	require.Equal(t, 1, len(selTracker.blocks))

	require.Equal(t, expectedTrackedBlock, selTracker.blocks[0])
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
				ReceiverShardID: 0,
				SenderShardID:   0,
				Type:            0,
				Reserved:        nil,
			},
			{
				TxHashes: [][]byte{
					[]byte("txHash3"),
				},
				ReceiverShardID: 0,
				SenderShardID:   0,
				Type:            0,
				Reserved:        nil,
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
				ReceiverShardID: 0,
				SenderShardID:   0,
				Type:            0,
				Reserved:        nil,
			},
			{
				TxHashes: [][]byte{
					[]byte("txHash3"),
				},
				ReceiverShardID: 0,
				SenderShardID:   0,
				Type:            0,
				Reserved:        nil,
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
