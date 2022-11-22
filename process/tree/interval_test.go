package tree

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestInterval(t *testing.T) {
	t.Parallel()

	i1Low := uint64(100)
	i1High := uint64(200)
	i1 := newInterval(i1Low, i1High)
	assert.False(t, check.IfNil(i1))
	assert.Equal(t, i1Low, i1.low)
	assert.Equal(t, i1High, i1.high)
	assert.True(t, i1.contains(100))
	assert.True(t, i1.contains(200))
	assert.True(t, i1.contains(150))
	assert.False(t, i1.contains(99))
	assert.False(t, i1.contains(201))

	i2 := newInterval(i1High, i1Low)
	assert.False(t, check.IfNil(i2))
	assert.Equal(t, i1Low, i2.low)
	assert.Equal(t, i1High, i2.high)
}
