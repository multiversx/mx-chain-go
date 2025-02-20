package interceptedBlocks_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/interceptedBlocks"
	"github.com/stretchr/testify/require"
)

func createSovMBInterceptorWithMBInShard(shardID uint32) process.InterceptedData {
	mb := createMockMiniblock()
	mb.SenderShardID = shardID
	buff, _ := testMarshalizer.Marshal(mb)

	args := createDefaultMiniblockArgument()
	args.MiniblockBuff = buff

	sovMBInterceptor, _ := interceptedBlocks.NewInterceptedSovereignMiniBlock(args)
	return sovMBInterceptor
}

func TestNewInterceptedSovereignMiniBlock(t *testing.T) {
	t.Parallel()

	t.Run("nil input in args", func(t *testing.T) {
		args := createDefaultMiniblockArgument()
		args.Marshalizer = nil
		sovMBInterceptor, err := interceptedBlocks.NewInterceptedSovereignMiniBlock(args)
		require.Nil(t, sovMBInterceptor)
		require.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createDefaultMiniblockArgument()
		sovMBInterceptor, err := interceptedBlocks.NewInterceptedSovereignMiniBlock(args)
		require.Nil(t, err)
		require.False(t, sovMBInterceptor.IsInterfaceNil())
	})
}

func TestInterceptedSovereignMiniBlock_CheckValidity(t *testing.T) {
	t.Parallel()

	sovMBInterceptor := createSovMBInterceptorWithMBInShard(core.MetachainShardId)
	err := sovMBInterceptor.CheckValidity()
	require.Equal(t, process.ErrInvalidShardId, err)

	sovMBInterceptor = createSovMBInterceptorWithMBInShard(core.SovereignChainShardId)
	err = sovMBInterceptor.CheckValidity()
	require.Nil(t, err)

	sovMBInterceptor = createSovMBInterceptorWithMBInShard(core.MainChainShardId)
	err = sovMBInterceptor.CheckValidity()
	require.Nil(t, err)
}
