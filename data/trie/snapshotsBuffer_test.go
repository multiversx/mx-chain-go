package trie

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func defaultSnapshotsBuffer() *snapshotsBuffer {
	sb := newSnapshotsBuffer()
	sb.add([]byte("1"))
	sb.add([]byte("2"))

	return sb
}

func TestSnapshotsBuffer_NewSnapshotBuffer(t *testing.T) {
	assert.NotNil(t, newSnapshotsBuffer())
}

func TestSnapshotsBuffer_Add(t *testing.T) {
	sb := defaultSnapshotsBuffer()

	_, ok := sb.buffer[string([]byte("1"))]
	assert.True(t, ok)

	_, ok = sb.buffer[string([]byte("2"))]
	assert.True(t, ok)
}

func TestSnapshotsBuffer_Contains(t *testing.T) {
	sb := defaultSnapshotsBuffer()

	assert.True(t, sb.contains([]byte("1")))
	assert.False(t, sb.contains([]byte("3")))
}

func TestSnapshotsBuffer_Remove(t *testing.T) {
	sb := defaultSnapshotsBuffer()
	sb.remove([]byte("1"))

	_, ok := sb.buffer[string([]byte("1"))]
	assert.False(t, ok)

	_, ok = sb.buffer[string([]byte("2"))]
	assert.True(t, ok)
}

func TestSnapshotsBuffer_Len(t *testing.T) {
	sb := defaultSnapshotsBuffer()

	assert.Equal(t, 2, sb.len())
}
