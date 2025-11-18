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

func TestBaseBlockchain_SetAndGetLastExecutedBlockHeaderAndRootHash(t *testing.T) {
	t.Parallel()

	t.Run("should set nil if header nil provided", func(t *testing.T) {
		t.Parallel()

		base := &baseBlockChain{
			appStatusHandler:      &mock.AppStatusHandlerStub{},
			finalBlockInfo:        &blockInfo{},
			lastExecutedBlockInfo: &blockInfo{},
		}

		hash := []byte("hash")
		rootHash := []byte("root-hash")

		base.SetLastExecutedBlockHeaderAndRootHash(nil, hash, rootHash)

		retHeader := base.GetLastExecutedBlockHeader()

		require.Nil(t, retHeader)

		retNonce, retHeaderHash, retRootHash := base.GetLastExecutedBlockInfo()
		require.Zero(t, retNonce)
		require.Nil(t, retHeaderHash)
		require.Nil(t, retRootHash)
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

		header1 := &block.HeaderV3{
			Nonce: nonce,
		}

		base.SetLastExecutedBlockHeaderAndRootHash(header1, hash, rootHash)

		retHeader := base.GetLastExecutedBlockHeader()

		require.Equal(t, nonce, retHeader.GetNonce())

		retNonce, retHeaderHash, retRootHash := base.GetLastExecutedBlockInfo()
		require.Equal(t, nonce, retNonce)
		require.Equal(t, hash, retHeaderHash)
		require.Equal(t, rootHash, retRootHash)
	})
}

func TestBaseBlockchain_SetAndGetLastExecutionResult(t *testing.T) {
	t.Parallel()

	base := &baseBlockChain{
		appStatusHandler:      &mock.AppStatusHandlerStub{},
		finalBlockInfo:        &blockInfo{},
		lastExecutedBlockInfo: &blockInfo{},
	}

	nonce := uint64(10)
	hash := []byte("hash")
	rootHash := []byte("root-hash")

	execResult := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			RootHash:    rootHash,
			HeaderHash:  hash,
			HeaderNonce: nonce,
		},
	}

	base.SetLastExecutionResult(execResult)

	retExecResult := base.GetLastExecutionResult()

	require.Equal(t, execResult, retExecResult)
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

	execResult := &block.ExecutionResult{}

	var wg sync.WaitGroup
	wg.Add(numCalls)

	for i := range numCalls {
		go func(i int) {
			switch i % 10 {
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
				bc.SetLastExecutedBlockHeaderAndRootHash(header, headerHash, rootHash)
			case 7:
				bc.SetCurrentBlockHeaderHash(headerHash)
			case 8:
				bc.SetFinalBlockInfo(0, headerHash, rootHash)
			case 9:
				bc.SetGenesisHeaderHash(headerHash)
			case 10:
				bc.SetLastExecutionResult(execResult)
			case 11:
				_ = bc.GetLastExecutionResult()
			default:
				require.Fail(t, "should have not been called")
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}
