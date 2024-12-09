package factory

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/require"
)

func TestNewInterceptedSovereignShardHeaderDataFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil input in args", func(t *testing.T) {
		coreComp, cryptoComp := createMockComponentHolders()
		coreComp.IntMarsh = nil
		args := createMockArgument(coreComp, cryptoComp)
		mbDataFactory, err := NewInterceptedSovereignMiniBlockDataFactory(args)
		require.Nil(t, mbDataFactory)
		require.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("should work", func(t *testing.T) {
		coreComp, cryptoComp := createMockComponentHolders()
		args := createMockArgument(coreComp, cryptoComp)
		mbDataFactory, err := NewInterceptedSovereignMiniBlockDataFactory(args)
		require.Nil(t, err)
		require.False(t, mbDataFactory.IsInterfaceNil())
	})
}

func TestInterceptedSovereignMiniBlockDataFactory_Create(t *testing.T) {
	t.Parallel()

	miniBlock := &block.MiniBlock{}

	coreComp, cryptoComp := createMockComponentHolders()
	args := createMockArgument(coreComp, cryptoComp)
	mbDataFactory, _ := NewInterceptedSovereignMiniBlockDataFactory(args)

	buff, err := coreComp.IntMarsh.Marshal(miniBlock)
	require.Nil(t, err)

	interceptedData, err := mbDataFactory.Create(buff)
	require.Nil(t, err)
	require.Equal(t, "*interceptedBlocks.interceptedSovereignMiniBlock", fmt.Sprintf("%T", interceptedData))
}
