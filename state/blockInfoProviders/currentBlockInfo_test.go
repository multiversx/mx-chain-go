package blockInfoProviders

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewCurrentBlockInfo(t *testing.T) {
	t.Parallel()

	t.Run("nil chain handler", func(t *testing.T) {
		t.Parallel()

		provider, err := NewCurrentBlockInfo(nil)

		assert.Equal(t, ErrNilChainHandler, err)
		assert.True(t, check.IfNil(provider))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		provider, err := NewCurrentBlockInfo(&testscommon.ChainHandlerStub{})

		assert.Nil(t, err)
		assert.False(t, check.IfNil(provider))
	})
}

func TestCurrentBlockInfo_GetBlockInfo(t *testing.T) {
	t.Parallel()

	t.Run("nil block header", func(t *testing.T) {
		t.Parallel()

		chainHandler := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return nil
			},
		}

		provider, _ := NewCurrentBlockInfo(chainHandler)
		bi := provider.GetBlockInfo()

		assert.True(t, bi.Equal(holders.NewBlockInfo(nil, 0, nil)))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		nonce := uint64(8837)
		hash := []byte("hash")
		rootHash := []byte("root hash")
		chainHandler := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					Nonce: nonce,
				}
			},
			GetCurrentBlockRootHashCalled: func() []byte {
				return rootHash
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return hash
			},
		}

		provider, _ := NewCurrentBlockInfo(chainHandler)
		bi := provider.GetBlockInfo()

		expectedBi := holders.NewBlockInfo(hash, nonce, rootHash)

		assert.True(t, bi.Equal(expectedBi))
	})
}
