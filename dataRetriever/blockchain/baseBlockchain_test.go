package blockchain

import (
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/mock"
	"github.com/stretchr/testify/require"
)

func TestBaseBlockchain_SetAndGetSetFinalBlockInfo(t *testing.T) {
	t.Parallel()

	base := &baseBlockChain{
		appStatusHandler: &mock.AppStatusHandlerStub{},
		finalBlockInfo:   &blockInfo{},
	}

	nonce := uint64(42)
	hash := []byte("hash")
	rootHash := []byte("root-hash")

	base.SetFinalBlockInfo(nonce, hash, rootHash)
	actualNonce, actualHash, actualRootHash := base.GetFinalBlockInfo()

	require.Equal(t, nonce, actualNonce)
	require.Equal(t, hash, actualHash)
	require.Equal(t, rootHash, actualRootHash)
}

func TestBaseBlockchain_SetAndGetSetFinalBlockInfoWorksWithNilValues(t *testing.T) {
	t.Parallel()

	base := &baseBlockChain{
		appStatusHandler: &mock.AppStatusHandlerStub{},
		finalBlockInfo:   &blockInfo{},
	}

	actualNonce, actualHash, actualRootHash := base.GetFinalBlockInfo()
	require.Equal(t, uint64(0), actualNonce)
	require.Nil(t, actualHash)
	require.Nil(t, actualRootHash)

	base.SetFinalBlockInfo(0, nil, nil)

	actualNonce, actualHash, actualRootHash = base.GetFinalBlockInfo()
	require.Equal(t, uint64(0), actualNonce)
	require.Nil(t, actualHash)
	require.Nil(t, actualRootHash)
}

func TestBaseBlockchain_SetAndGetLastExecutedBlockInfo(t *testing.T) {
	t.Parallel()

	base := &baseBlockChain{
		appStatusHandler:      &mock.AppStatusHandlerStub{},
		finalBlockInfo:        &blockInfo{},
		lastExecutedBlockInfo: &blockInfo{},
	}

	nonce := uint64(10)
	hash := []byte("hash")
	rootHash := []byte("root-hash")

	base.SetLastExecutedBlockInfo(nonce, hash, rootHash)
	actualNonce, actualHash, actualRootHash := base.GetLastExecutedBlockInfo()

	require.Equal(t, nonce, actualNonce)
	require.Equal(t, hash, actualHash)
	require.Equal(t, rootHash, actualRootHash)
}

func TestBaseBlockchain_SetAndGetLastExecutedBlockHeader(t *testing.T) {
	t.Parallel()

	t.Run("should fail if nonce does not match with last executed info", func(t *testing.T) {
		t.Parallel()

		base := &baseBlockChain{
			appStatusHandler:      &mock.AppStatusHandlerStub{},
			finalBlockInfo:        &blockInfo{},
			lastExecutedBlockInfo: &blockInfo{},
			lastExecutedBlockHeader: &block.HeaderV3{
				Nonce: uint64(1),
			},
		}

		nonce := uint64(10)
		hash := []byte("hash")
		rootHash := []byte("root-hash")

		base.SetLastExecutedBlockInfo(nonce, hash, rootHash)

		header1 := &block.HeaderV3{
			Nonce: uint64(11),
		}

		err := base.SetLastExecutedBlockHeader(header1)
		require.Equal(t, ErrNonceDoesNotMatch, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		base := &baseBlockChain{
			appStatusHandler:      &mock.AppStatusHandlerStub{},
			finalBlockInfo:        &blockInfo{},
			lastExecutedBlockInfo: &blockInfo{},
		}

		nonce := uint64(10)
		hash := []byte("hash")
		rootHash := []byte("root-hash")

		base.SetLastExecutedBlockInfo(nonce, hash, rootHash)

		// should set nil header if nil provided
		err := base.SetLastExecutedBlockHeader(nil)
		require.Nil(t, err)

		// should return nil if not set
		retHeader := base.GetLastExecutedBlockHeader()
		require.Nil(t, retHeader)

		header1 := &block.HeaderV3{
			Nonce: nonce,
		}

		err = base.SetLastExecutedBlockHeader(header1)
		require.Nil(t, err)

		retHeader = base.GetLastExecutedBlockHeader()

		require.Equal(t, nonce, retHeader.GetNonce())
	})
}

func TestBaseBlockchain_Concurrency(t *testing.T) {
	t.Parallel()

	bc := &baseBlockChain{
		appStatusHandler:      &mock.AppStatusHandlerStub{},
		finalBlockInfo:        &blockInfo{},
		lastExecutedBlockInfo: &blockInfo{},
	}

	numCalls := 100

	header := &block.HeaderV3{}
	headerHash := []byte("headerHash")
	rootHash := []byte("rootHash")

	var wg sync.WaitGroup
	wg.Add(numCalls)

	for i := range numCalls {
		go func(i int) {
			switch i % 11 {
			case 0:
				_ = bc.GetCurrentBlockHeaderHash()
			case 1:
				_ = bc.GetGenesisHeader()
			case 2:
				_ = bc.GetGenesisHeaderHash()
			case 3:
				_ = bc.GetLastExecutedBlockHeader()
			case 4:
				_, _, _ = bc.GetFinalBlockInfo()
			case 5:
				_, _, _ = bc.GetLastExecutedBlockInfo()
			case 6:
				_ = bc.SetLastExecutedBlockHeader(header)
			case 7:
				bc.SetCurrentBlockHeaderHash(headerHash)
			case 8:
				bc.SetFinalBlockInfo(0, headerHash, rootHash)
			case 9:
				bc.SetGenesisHeaderHash(headerHash)
			case 10:
				bc.SetLastExecutedBlockInfo(0, headerHash, rootHash)
			default:
				require.Fail(t, "should have not been called")
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}
