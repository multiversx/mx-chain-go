package factory

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/require"
)

func TestNewInterceptedSovereignMiniBlockDataFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil input in args", func(t *testing.T) {
		coreComp, cryptoComp := createMockComponentHolders()
		coreComp.IntMarsh = nil
		args := createMockArgument(coreComp, cryptoComp)
		hdrDataFactory, err := NewInterceptedSovereignShardHeaderDataFactory(args)
		require.Nil(t, hdrDataFactory)
		require.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("should work", func(t *testing.T) {
		coreComp, cryptoComp := createMockComponentHolders()
		args := createMockArgument(coreComp, cryptoComp)
		hdrDataFactory, err := NewInterceptedSovereignShardHeaderDataFactory(args)
		require.Nil(t, err)
		require.False(t, hdrDataFactory.IsInterfaceNil())
	})
}

func TestInterceptedSovereignShardHeaderDataFactory_Create(t *testing.T) {
	t.Parallel()

	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{},
	}

	coreComp, cryptoComp := createMockComponentHolders()
	args := createMockArgument(coreComp, cryptoComp)
	hdrDataFactory, _ := NewInterceptedSovereignShardHeaderDataFactory(args)

	buff, err := coreComp.IntMarsh.Marshal(sovHdr)
	require.Nil(t, err)

	interceptedData, err := hdrDataFactory.Create(buff)
	require.Nil(t, err)
	require.Equal(t, "*interceptedBlocks.interceptedSovereignBlockHeader", fmt.Sprintf("%T", interceptedData))
}
