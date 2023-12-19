package blockchain

import (
	"testing"

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

func TestBaseBlockChain_SetCurrentAggregatedSignatureAndBitmap(t *testing.T) {
	t.Parallel()

	base := &baseBlockChain{}
	sig, bitmap := base.GetCurrentAggregatedSignatureAndBitmap()
	require.Nil(t, sig)
	require.Nil(t, bitmap)

	providedSig := []byte("provided sig")
	providedBitmap := []byte("provided bitmap")
	base.SetCurrentAggregatedSignatureAndBitmap(providedSig, providedBitmap)
	sig, bitmap = base.GetCurrentAggregatedSignatureAndBitmap()
	require.Equal(t, providedSig, sig)
	require.Equal(t, providedBitmap, bitmap)
}
