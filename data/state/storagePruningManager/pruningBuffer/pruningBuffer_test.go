package pruningBuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func defaultPruningBuffer() *pruningBuffer {
	sb := NewPruningBuffer(100)
	sb.Add([]byte("0"))
	sb.Add([]byte("1"))

	return sb
}

func TestPruningBuffer_NewPruningBuffer(t *testing.T) {
	assert.NotNil(t, NewPruningBuffer(100))
}

func TestSnapshotsBuffer_Add(t *testing.T) {
	sb := defaultPruningBuffer()

	assert.Equal(t, 2, len(sb.buffer))
}

func TestSnapshotBuffer_Len(t *testing.T) {
	sb := defaultPruningBuffer()

	assert.Equal(t, 2, sb.Len())
}

func TestSnapshotsBuffer_RemoveAll(t *testing.T) {
	sb := defaultPruningBuffer()

	buffer := sb.RemoveAll()
	assert.Equal(t, []byte("0"), buffer[0])
	assert.Equal(t, []byte("1"), buffer[1])

	assert.Equal(t, 0, len(sb.buffer))
}
