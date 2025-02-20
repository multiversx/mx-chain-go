package block

import (
	"fmt"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestNewSovereignShardHeaderFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil input", func(t *testing.T) {
		sovHdrFactory, err := NewSovereignShardHeaderFactory(nil)
		require.Nil(t, sovHdrFactory)
		require.Equal(t, ErrNilHeaderVersionHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		sovHdrFactory, err := NewSovereignShardHeaderFactory(&testscommon.HeaderVersionHandlerStub{})
		require.Nil(t, err)
		require.False(t, sovHdrFactory.IsInterfaceNil())
	})
}

func TestSovereignShardHeaderFactory_Create(t *testing.T) {
	t.Parallel()

	wasGetVersionCalled := false
	sovHdrFactory, _ := NewSovereignShardHeaderFactory(&testscommon.HeaderVersionHandlerStub{
		GetVersionCalled: func(epoch uint32) string {
			wasGetVersionCalled = true
			return fmt.Sprintf("%d", epoch)
		},
	})

	createdHdr := sovHdrFactory.Create(0)
	require.IsType(t, &block.SovereignChainHeader{}, createdHdr)

	createdHdr = sovHdrFactory.Create(math.MaxUint32)
	require.IsType(t, &block.SovereignChainHeader{}, createdHdr)

	require.True(t, wasGetVersionCalled)
}
