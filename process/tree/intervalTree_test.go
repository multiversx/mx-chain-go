package tree

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/stretchr/testify/assert"
)

func TestNewIntervalTree(t *testing.T) {
	t.Parallel()

	cfg := config.HardforkV2Config{
		BlocksExceptionsByRound: []config.BlocksExceptionInterval{
			{
				Low:  15,
				High: 20,
			},
			{
				Low:  10,
				High: 30,
			},
			{
				Low:  17,
				High: 19,
			},
			{
				Low:  5,
				High: 20,
			},
			{
				Low:  12,
				High: 15,
			},
			{
				Low:  30,
				High: 40,
			},
		},
	}
	//             [15, 20]
	//              /     \
	//       [10, 30]     [17, 19]
	//        /   \           \
	//  [5, 20]   [12, 15]    [30, 40]

	tree := NewIntervalTree(cfg)
	assert.False(t, check.IfNil(tree))
	assert.Equal(t, uint64(40), tree.root.max)
	assert.True(t, tree.Contains(16))
	assert.True(t, tree.Contains(11))
	assert.True(t, tree.Contains(35))
	assert.True(t, tree.Contains(6))
	assert.False(t, tree.Contains(4))
	assert.False(t, tree.Contains(41))
}
