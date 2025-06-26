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

	_, err := NewSelectionTracker()
	require.Nil(t, err)
}

func TestSelectionTracker_OnProposedBlockShouldErr(t *testing.T) {
	t.Parallel()

	t.Run("should err nil block hash", func(t *testing.T) {
		t.Parallel()

		tracker, err := NewSelectionTracker()
		require.Nil(t, err)

		err = tracker.OnProposedBlock(nil, nil, nil)
		require.Equal(t, err, errNilBlockHash)
	})

	t.Run("should err nil header", func(t *testing.T) {
		t.Parallel()

		tracker, err := NewSelectionTracker()
		require.Nil(t, err)

		err = tracker.OnProposedBlock([]byte("hash1"), nil, nil)
		require.Equal(t, err, errNilBlockBody)
	})

	t.Run("should err nil header", func(t *testing.T) {
		t.Parallel()

		tracker, err := NewSelectionTracker()
		require.Nil(t, err)

		blockBody := block.Body{}
		err = tracker.OnProposedBlock([]byte("hash1"), &blockBody, nil)
		require.Equal(t, err, errNilHeaderHandler)
	})
}

func TestSelectionTracker_OnProposedBlockShouldWork(t *testing.T) {
	t.Parallel()

	tracker, err := NewSelectionTracker()
	require.Nil(t, err)

	numOfBlocks := 20
	headers := createMockedHeaders(numOfBlocks)
	proposeBlocksConcurrently(t, numOfBlocks, tracker, headers)
	require.Equal(t, 20, len(tracker.blocks))
}

func TestSelectionTracker_OnExecutedBlockShouldError(t *testing.T) {
	t.Parallel()

	tracker, err := NewSelectionTracker()
	require.Nil(t, err)

	err = tracker.OnExecutedBlock(nil)
	require.Equal(t, errNilHeaderHandler, err)
}

func TestSelectionTracker_OnExecutedBlockShouldWork(t *testing.T) {
	t.Parallel()

	selTracker, err := NewSelectionTracker()
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

		selTracker, err := NewSelectionTracker()
		require.Nil(t, err)

		selTracker.updateLatestRootHashNoLock(1, []byte("rootHash1"))
		require.Equal(t, uint64(1), selTracker.latestNonce)
		require.Equal(t, []byte("rootHash1"), selTracker.latestRootHash)
	})

	t.Run("root hash of block N after root hash of block N+1", func(t *testing.T) {
		t.Parallel()

		selTracker, err := NewSelectionTracker()
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

		selTracker, err := NewSelectionTracker()
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

	selTracker, err := NewSelectionTracker()
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

func TestSelectionTracker_nextBlock(t *testing.T) {
	t.Parallel()

	t.Run("should return next block", func(t *testing.T) {
		t.Parallel()

		selTracker, err := NewSelectionTracker()
		require.Nil(t, err)

		expectedNextBlock := newTrackedBlock(0, []byte("blockHash2"), []byte("rootHash2"), []byte("blockHash1"))
		selTracker.blocks = []*trackedBlock{
			newTrackedBlock(0, []byte("blockHash1"), []byte("rootHash1"), []byte("prevHash1")),
			expectedNextBlock,
		}

		receivedNextBlock := selTracker.nextBlock([]byte("blockHash1"))
		require.Equal(t, expectedNextBlock, receivedNextBlock)
	})

	t.Run("should return nil", func(t *testing.T) {
		t.Parallel()

		selTracker, err := NewSelectionTracker()
		require.Nil(t, err)

		expectedNextBlock := newTrackedBlock(0, []byte("blockHash2"), []byte("rootHash2"), []byte("blockHash1"))
		selTracker.blocks = []*trackedBlock{
			newTrackedBlock(0, []byte("blockHash1"), []byte("rootHash1"), []byte("prevHash1")),
			expectedNextBlock,
		}

		receivedNextBlock := selTracker.nextBlock([]byte("notExistingBlockHash"))
		require.Nil(t, receivedNextBlock)
	})
}

func TestSelectionTracker_getChainOfTrackedBlocks(t *testing.T) {
	t.Parallel()

	selTracker, err := NewSelectionTracker()
	require.Nil(t, err)

	// create a slice of tracked block which aren't ordered
	selTracker.blocks = []*trackedBlock{
		newTrackedBlock(7, []byte("blockHash8"), []byte("rootHash8"), []byte("blockHash7")),
		newTrackedBlock(5, []byte("blockHash6"), []byte("rootHash6"), []byte("blockHash5")),
		newTrackedBlock(1, []byte("blockHash2"), []byte("rootHash2"), []byte("blockHash1")),
		newTrackedBlock(0, []byte("blockHash1"), []byte("rootHash1"), []byte("prevHash1")),
		newTrackedBlock(3, []byte("blockHash4"), []byte("rootHash4"), []byte("blockHash3")),
		newTrackedBlock(2, []byte("blockHash3"), []byte("rootHash3"), []byte("blockHash2")),
		newTrackedBlock(4, []byte("blockHash5"), []byte("rootHash5"), []byte("blockHash4")),
		newTrackedBlock(6, []byte("blockHash7"), []byte("rootHash7"), []byte("blockHash6")),
	}

	t.Run("should return expected tracked blocks and stop before nonce", func(t *testing.T) {
		t.Parallel()

		expectedTrackedBlockHashes := [][]byte{
			[]byte("blockHash5"),
			[]byte("blockHash6"),
			[]byte("blockHash7"),
		}

		returnedChain := selTracker.getChainOfTrackedBlocks([]byte("blockHash4"), 7)
		for i, returnedBlock := range returnedChain {
			require.Equal(t, returnedBlock.hash, expectedTrackedBlockHashes[i])
		}
	})

	t.Run("should return expected tracked blocks and stop because nil block", func(t *testing.T) {
		t.Parallel()

		expectedTrackedBlockHashes := [][]byte{
			[]byte("blockHash5"),
			[]byte("blockHash6"),
			[]byte("blockHash7"),
			[]byte("blockHash8"),
		}

		returnedChain := selTracker.getChainOfTrackedBlocks([]byte("blockHash4"), 12)
		for i, returnedBlock := range returnedChain {
			require.Equal(t, returnedBlock.hash, expectedTrackedBlockHashes[i])
		}
	})

	t.Run("should return 0 blocks because prevHash not found", func(t *testing.T) {
		t.Parallel()

		returnedChain := selTracker.getChainOfTrackedBlocks([]byte("blockHashX"), 12)
		require.Equal(t, 0, len(returnedChain))
	})

	t.Run("should return 0 blocks because nonce", func(t *testing.T) {
		t.Parallel()

		returnedChain := selTracker.getChainOfTrackedBlocks([]byte("blockHash6"), 6)
		require.Equal(t, 0, len(returnedChain))
	})

	t.Run("should return 1 blocks", func(t *testing.T) {
		t.Parallel()

		returnedChain := selTracker.getChainOfTrackedBlocks([]byte("blockHash6"), 7)
		require.Equal(t, 1, len(returnedChain))
	})
}

func TestSelectionTracker_deriveVirtualSelectionSessionShouldErr(t *testing.T) {
	t.Parallel()

	selTracker, err := NewSelectionTracker()
	require.Nil(t, err)

	err = errors.New("expected err")

	mockSelectionSession := txcachemocks.SelectionSessionMock{}
	mockSelectionSession.GetRootHashCalled = func() ([]byte, error) {
		return nil, err
	}
	virtualSession, returnedErr := selTracker.deriveVirtualSelectionSession(&mockSelectionSession, nil, 0)
	require.Nil(t, virtualSession)
	require.Equal(t, err, returnedErr)
}
