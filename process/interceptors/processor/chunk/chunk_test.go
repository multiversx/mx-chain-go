package chunk

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChunk(t *testing.T) {
	t.Parallel()

	c := NewChunk(0)
	assert.False(t, check.IfNil(c))
}

func TestChunk_Put(t *testing.T) {
	t.Parallel()

	c := NewChunk(2)
	val1 := []byte("val1")
	c.Put(1, val1)
	require.Equal(t, 1, len(c.data))
	assert.Equal(t, val1, c.data[1])
	assert.Equal(t, 4, c.Size())

	val2 := []byte("val222222")
	c.Put(1, val2)
	require.Equal(t, 1, len(c.data))
	assert.Equal(t, val2, c.data[1])
	assert.Equal(t, 9, c.Size())
}

func TestChunk_PutOutOfBounds(t *testing.T) {
	t.Parallel()

	c := NewChunk(2)
	val1 := []byte("val1")
	c.Put(2, val1)
	require.Equal(t, 0, len(c.data))
	assert.Equal(t, 0, c.Size())
}

func TestChunk_TryAssembleAllChunks(t *testing.T) {
	t.Parallel()

	c := NewChunk(3)
	completeBuff := c.TryAssembleAllChunks()
	assert.Nil(t, completeBuff)

	buff4 := []byte("buff4")
	c.Put(4, buff4)
	completeBuff = c.TryAssembleAllChunks()
	assert.Nil(t, completeBuff)

	buff2 := []byte("buff2")
	c.Put(1, buff2)
	completeBuff = c.TryAssembleAllChunks()
	assert.Nil(t, completeBuff)

	buff1 := []byte("buff1")
	c.Put(0, buff1)
	completeBuff = c.TryAssembleAllChunks()
	assert.Nil(t, completeBuff)

	buff3 := []byte("buff3")
	c.Put(2, buff3)
	completeBuff = c.TryAssembleAllChunks()
	expectedBuff := append(append(buff1, buff2...), buff3...)
	assert.Equal(t, expectedBuff, completeBuff)
}

func TestChunk_GetAllMissingChunkIndexes(t *testing.T) {
	t.Parallel()

	c := NewChunk(3)
	missing := c.GetAllMissingChunkIndexes()
	assert.Equal(t, []uint32{0, 1, 2}, missing)

	buff4 := []byte("buff4")
	c.Put(4, buff4)
	missing = c.GetAllMissingChunkIndexes()
	assert.Equal(t, []uint32{0, 1, 2}, missing)

	buff2 := []byte("buff2")
	c.Put(1, buff2)
	missing = c.GetAllMissingChunkIndexes()
	assert.Equal(t, []uint32{0, 2}, missing)

	buff1 := []byte("buff1")
	c.Put(0, buff1)
	missing = c.GetAllMissingChunkIndexes()
	assert.Equal(t, []uint32{2}, missing)

	buff3 := []byte("buff3")
	c.Put(2, buff3)
	missing = c.GetAllMissingChunkIndexes()
	assert.Equal(t, 0, len(missing))
}
