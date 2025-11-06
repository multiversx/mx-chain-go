package blockchain

import (
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
