package holders

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestNewReceiptsHolder(t *testing.T) {
	holder := NewReceiptsHolder(nil)
	require.Nil(t, holder.GetMiniblocks())

	holder = NewReceiptsHolder([]*block.MiniBlock{})
	require.Equal(t, []*block.MiniBlock{}, holder.GetMiniblocks())

	holder = NewReceiptsHolder([]*block.MiniBlock{{SenderShardID: 42}, {SenderShardID: 43}})
	require.Equal(t, []*block.MiniBlock{{SenderShardID: 42}, {SenderShardID: 43}}, holder.GetMiniblocks())
}

func TestReceiptsHolder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var holder *receiptsHolder
	require.True(t, holder.IsInterfaceNil())

	holder = NewReceiptsHolder([]*block.MiniBlock{})
	require.False(t, holder.IsInterfaceNil())
}
