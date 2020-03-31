package trie

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func defaultPruningBuffer() *pruningBuffer {
	sb := newPruningBuffer()
	sb.add([]byte("0"))
	sb.add([]byte("1"))

	return sb
}

func TestPruningBuffer_NewPruningBuffer(t *testing.T) {
	assert.NotNil(t, newPruningBuffer())
}

func TestSnapshotsBuffer_Add(t *testing.T) {
	sb := defaultPruningBuffer()

	assert.Equal(t, 2, len(sb.buffer))
}

func TestSnapshotsBuffer_RemoveFirstValue(t *testing.T) {
	sb := defaultPruningBuffer()

	sb.remove([]byte("0"))
	assert.Equal(t, 1, len(sb.buffer))
}

func TestSnapshotsBuffer_RemoveLastValue(t *testing.T) {
	sb := defaultPruningBuffer()

	newVal := []byte("2")
	sb.add(newVal)
	assert.Equal(t, 3, len(sb.buffer))

	sb.remove(newVal)
	assert.Equal(t, 2, len(sb.buffer))
}

func TestSnapshotsBuffer_RemoveAll(t *testing.T) {
	sb := defaultPruningBuffer()

	buffer := sb.removeAll()
	_, ok := buffer[string([]byte("0"))]
	assert.True(t, ok)
	_, ok = buffer[string([]byte("1"))]
	assert.True(t, ok)

	assert.Equal(t, 0, len(sb.buffer))
}
