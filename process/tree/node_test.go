package tree

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNode(t *testing.T) {
	t.Parallel()

	iLow := uint64(100)
	iHigh := uint64(200)
	i := newInterval(iLow, iHigh)

	n := newNode(i)
	assert.False(t, check.IfNil(n))
	assert.Equal(t, iLow, n.low())
	assert.Equal(t, iHigh, n.high())
	assert.True(t, n.contains(100))
	assert.False(t, n.contains(99))
	assert.True(t, n.isLeaf())
	assert.Nil(t, n.left)
	assert.Nil(t, n.right)

	i2 := newInterval(20, 50)
	n2 := newNode(i2)
	i3 := newInterval(30, 60)
	n3 := newNode(i3)
	n.left = n2
	n.right = n3
	assert.False(t, n.isLeaf())
	assert.True(t, n2.isLeaf())
	assert.True(t, n3.isLeaf())
	assert.Equal(t, n2, n.left)
	assert.Equal(t, n2.low(), n.left.low())
	assert.Equal(t, n2.high(), n.left.high())
	assert.Equal(t, n3, n.right)
	assert.Equal(t, n3.low(), n.right.low())
	assert.Equal(t, n3.high(), n.right.high())
}
