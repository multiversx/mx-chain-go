package sync

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/tree"
	"github.com/stretchr/testify/assert"
)

func TestNewHardforkExclusionHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil tree should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewHardforkExclusionHandler(nil)
		assert.Equal(t, ErrNilExclusionTree, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cfg := []tree.BlocksExceptionInterval{
			{
				Low:  15,
				High: 18,
			},
			{
				Low:  10,
				High: 19,
			},
			{
				Low:  17,
				High: 22,
			},
		}
		//             [15, 18]
		//              /     \
		//       [10, 19]     [17, 22]

		providedTree := tree.NewIntervalTree(cfg)
		assert.False(t, check.IfNil(providedTree))
		handler, err := NewHardforkExclusionHandler(providedTree)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
		assert.True(t, handler.IsRoundExcluded(16))
		assert.True(t, handler.IsRoundExcluded(19))
		assert.True(t, handler.IsRoundExcluded(22))
		assert.False(t, handler.IsRoundExcluded(1))
	})
}
