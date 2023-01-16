package blockInfoProviders

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewFinalBlockInfo(t *testing.T) {
	t.Parallel()

	t.Run("nil chain handler", func(t *testing.T) {
		t.Parallel()

		provider, err := NewFinalBlockInfo(nil)

		assert.Equal(t, ErrNilChainHandler, err)
		assert.True(t, check.IfNil(provider))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		provider, err := NewFinalBlockInfo(&testscommon.ChainHandlerStub{})

		assert.Nil(t, err)
		assert.False(t, check.IfNil(provider))
	})
}

func TestFinalBlockInfo_GetBlockInfo(t *testing.T) {
	t.Parallel()

	providedNonce := uint64(8837)
	providedHash := []byte("hash")
	providedRootHash := []byte("root hash")
	chainHandler := &testscommon.ChainHandlerStub{
		GetFinalBlockInfoCalled: func() (nonce uint64, blockHash []byte, rootHash []byte) {
			nonce = providedNonce
			blockHash = providedHash
			rootHash = providedRootHash

			return
		},
	}
	provider, _ := NewFinalBlockInfo(chainHandler)
	bi := provider.GetBlockInfo()

	expectedBi := holders.NewBlockInfo(providedHash, providedNonce, providedRootHash)

	assert.True(t, bi.Equal(expectedBi))
}
