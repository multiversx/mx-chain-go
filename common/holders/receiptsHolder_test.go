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
