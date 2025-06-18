package txcache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"
)

func createMockedHeaders(numOfHeaders int) []*block.Header {
	headers := make([]*block.Header, 0)

	for i := 0; i < numOfHeaders; i++ {
		headers = append(headers, &block.Header{
			Nonce:    uint64(i),
			PrevHash: []byte(fmt.Sprintf("prevHash%d", i)),
			RootHash: []byte(fmt.Sprintf("rootHash%d", i)),
		})
	}

	return headers
}

func proposeBlocks(t *testing.T, numOfBlocks int, selectionTracker *selectionTracker, headers []*block.Header) {
	wg := sync.WaitGroup{}
	wg.Add(numOfBlocks)

	for i := 0; i < numOfBlocks; i++ {
		go func(index int) {
			err := selectionTracker.OnProposedBlock([]byte(fmt.Sprintf("blockHash%d", i)), &block.Body{}, headers[index])
			require.Nil(t, err)
			defer wg.Done()
		}(i)
	}

	wg.Wait()
}

func executeBlocks(t *testing.T, numOfBlocks int, selectionTracker *selectionTracker, headers []*block.Header) {
	wg := sync.WaitGroup{}
	wg.Add(numOfBlocks)

	for i := 0; i < numOfBlocks; i++ {
		go func(index int) {
			err := selectionTracker.OnExecutedBlock(headers[index])
			require.Nil(t, err)
			defer wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestNewSelectionTracker(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		selTracker, err := NewSelectionTracker(nil)
		require.Nil(t, selTracker)
		require.Error(t, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		_, err := NewSelectionTracker(txCache)
		require.Nil(t, err)
	})
}

func TestSelectionTracker_OnProposedBlockShouldErr(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	selTracker, err := NewSelectionTracker(txCache)
	require.Nil(t, err)

	t.Run("should err nil block hash", func(t *testing.T) {
		t.Parallel()

		err = selTracker.OnProposedBlock(nil, nil, nil)
		require.Equal(t, err, errNilBlockHash)
	})

	t.Run("should err nil header", func(t *testing.T) {
		t.Parallel()

		err = selTracker.OnProposedBlock([]byte("hash1"), nil, nil)
		require.Equal(t, err, errNilBlockBody)
	})

	t.Run("should err nil header", func(t *testing.T) {
		t.Parallel()

		blockBody := block.Body{}
		err = selTracker.OnProposedBlock([]byte("hash1"), &blockBody, nil)
		require.Equal(t, err, errNilHeaderHandler)
	})
}

func TestSelectionTracker_OnProposedBlockShouldWork(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	selTracker, err := NewSelectionTracker(txCache)
	require.Nil(t, err)

	numOfBlocks := 20
	headers := createMockedHeaders(numOfBlocks)
	proposeBlocks(t, numOfBlocks, selTracker, headers)
}

func TestSelectionTracker_OnExecutedBlockShouldError(t *testing.T) {
	t.Parallel()

	t.Run("should error nil block hash", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		selectionTracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		err = selectionTracker.OnExecutedBlock(nil)
		require.Equal(t, errNilHeaderHandler, err)
	})
}

func TestSelectionTracker_OnExecutedBlockShouldWork(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	selTracker, err := NewSelectionTracker(txCache)
	require.Nil(t, err)

	numOfBlocks := 20
	headers := createMockedHeaders(numOfBlocks)

	proposeBlocks(t, numOfBlocks, selTracker, headers)
	require.Equal(t, numOfBlocks, len(selTracker.state))

	executeBlocks(t, numOfBlocks, selTracker, headers)
	require.Equal(t, 0, len(selTracker.state))
	require.Equal(t, uint64(19), selTracker.latestNonce)
	require.Equal(t, []byte("rootHash19"), selTracker.latestRootHash)
}

func TestSelectionTracker_equalBlocks(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	selTracker, err := NewSelectionTracker(txCache)
	require.Nil(t, err)

	t.Run("same nonce and same prev hash", func(t *testing.T) {
		t.Parallel()

		tracedBlock1 := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		tracedBlock2 := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash2"), []byte("blockPrevHash1"))
		equalBlocks := selTracker.equalBlocks(tracedBlock1, tracedBlock2)
		require.True(t, equalBlocks)
	})

	t.Run("different nonce", func(t *testing.T) {
		t.Parallel()

		tracedBlock1 := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		tracedBlock2 := newTrackedBlock(1, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		equalBlocks := selTracker.equalBlocks(tracedBlock1, tracedBlock2)
		require.False(t, equalBlocks)
	})
}

func TestSelectionTracker_updateLatestRoothash(t *testing.T) {
	t.Parallel()

	t.Run("latest roothash is nil", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		selTracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		selTracker.updateLatestRootHash(1, []byte("rootHash1"))
		require.Equal(t, uint64(1), selTracker.latestNonce)
		require.Equal(t, []byte("rootHash1"), selTracker.latestRootHash)
	})

	t.Run("duplicated roothash", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		selTracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		selTracker.updateLatestRootHash(1, []byte("rootHash1"))
		require.Equal(t, uint64(1), selTracker.latestNonce)
		require.Equal(t, []byte("rootHash1"), selTracker.latestRootHash)

		selTracker.updateLatestRootHash(2, []byte("rootHash1"))
		require.Equal(t, uint64(1), selTracker.latestNonce)
		require.Equal(t, []byte("rootHash1"), selTracker.latestRootHash)
	})

	t.Run("root hash of block N after root hash of block N+1", func(t *testing.T) {
		t.Parallel()

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		selTracker, err := NewSelectionTracker(txCache)
		require.Nil(t, err)

		selTracker.updateLatestRootHash(2, []byte("rootHash2"))
		require.Equal(t, uint64(2), selTracker.latestNonce)
		require.Equal(t, []byte("rootHash2"), selTracker.latestRootHash)

		selTracker.updateLatestRootHash(1, []byte("rootHash1"))
		require.Equal(t, uint64(2), selTracker.latestNonce)
		require.Equal(t, []byte("rootHash2"), selTracker.latestRootHash)
	})
}

func TestSelectionTracker_removeFromTrackedBlocks(t *testing.T) {
	t.Parallel()

	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
	selTracker, err := NewSelectionTracker(txCache)
	require.Nil(t, err)

	selTracker.state = []*trackedBlock{
		{
			nonce:                0,
			hash:                 []byte("blockHash1"),
			rootHash:             []byte("rootHash1"),
			prevHash:             []byte("prevHash1"),
			breadcrumbsByAddress: make(map[string]*accountBreadcrumb),
		},
		{
			nonce:                1,
			hash:                 []byte("blockHash2"),
			rootHash:             []byte("rootHash2"),
			prevHash:             []byte("prevHash2"),
			breadcrumbsByAddress: make(map[string]*accountBreadcrumb),
		},
		{
			nonce:                0,
			hash:                 []byte("blockHash3"),
			rootHash:             []byte("rootHash3"),
			prevHash:             []byte("prevHash1"),
			breadcrumbsByAddress: make(map[string]*accountBreadcrumb),
		},
	}

	require.Equal(t, 3, len(selTracker.state))
	selTracker.removeFromTrackedBlocks(&trackedBlock{
		nonce:                0,
		hash:                 nil,
		rootHash:             nil,
		prevHash:             []byte("prevHash1"),
		breadcrumbsByAddress: make(map[string]*accountBreadcrumb),
	})
	require.Equal(t, 1, len(selTracker.state))
	require.Equal(t, &trackedBlock{
		nonce:                1,
		hash:                 []byte("blockHash2"),
		rootHash:             []byte("rootHash2"),
		prevHash:             []byte("prevHash2"),
		breadcrumbsByAddress: make(map[string]*accountBreadcrumb),
	}, selTracker.state[0])
}
