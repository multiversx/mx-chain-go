package blockchain

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/mock"
	"github.com/stretchr/testify/require"
)

func TestBaseBlockchain_SetAndGetSetPreviousToFinalBlockInfo(t *testing.T) {
	base := &baseBlockChain{
		appStatusHandler:         &mock.AppStatusHandlerStub{},
		previousToFinalBlockInfo: &blockInfo{},
	}

	nonce := uint64(42)
	hash := []byte("hash")
	rootHash := []byte("root-hash")

	base.SetPreviousToFinalBlockInfo(nonce, hash, rootHash)
	actualNonce, actualHash, actualRootHash := base.GetPreviousToFinalBlockInfo()

	require.Equal(t, nonce, actualNonce)
	require.Equal(t, hash, actualHash)
	require.Equal(t, rootHash, actualRootHash)
}

func TestBaseBlockchain_SetAndGetSetPreviousToFinalBlockInfoWorksWithNilValues(t *testing.T) {
	base := &baseBlockChain{
		appStatusHandler:         &mock.AppStatusHandlerStub{},
		previousToFinalBlockInfo: &blockInfo{},
	}

	actualNonce, actualHash, actualRootHash := base.GetPreviousToFinalBlockInfo()
	require.Equal(t, uint64(0), actualNonce)
	require.Nil(t, actualHash)
	require.Nil(t, actualRootHash)

	base.SetPreviousToFinalBlockInfo(0, nil, nil)

	actualNonce, actualHash, actualRootHash = base.GetPreviousToFinalBlockInfo()
	require.Equal(t, uint64(0), actualNonce)
	require.Nil(t, actualHash)
	require.Nil(t, actualRootHash)
}
