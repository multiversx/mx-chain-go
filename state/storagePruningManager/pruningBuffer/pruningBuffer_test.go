package pruningBuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func defaultPruningBuffer() *pruningBuffer {
	pb := NewPruningBuffer(100)
	pb.Add([]byte("0"))
	pb.Add([]byte("1"))

	return pb
}

func TestPruningBuffer_NewPruningBuffer(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, NewPruningBuffer(100))
}

func TestSnapshotsBuffer_Add(t *testing.T) {
	t.Parallel()

	pb := defaultPruningBuffer()

	assert.Equal(t, 2, len(pb.buffer))
}

func TestSnapshotBuffer_Len(t *testing.T) {
	t.Parallel()

	pb := defaultPruningBuffer()

	assert.Equal(t, 2, pb.Len())
}

func TestSnapshotsBuffer_RemoveAll(t *testing.T) {
	t.Parallel()

	pb := defaultPruningBuffer()

	buffer := pb.RemoveAll()
	assert.Equal(t, []byte("0"), buffer[0])
	assert.Equal(t, []byte("1"), buffer[1])

	assert.Equal(t, 0, len(pb.buffer))
}

func TestSnapshotsBuffer_AddWhileFull(t *testing.T) {
	t.Parallel()

	pb := NewPruningBuffer(2)
	pb.Add([]byte("0"))
	pb.Add([]byte("1"))
	pb.Add([]byte("2"))

	assert.Equal(t, 2, pb.Len())
	buffer := pb.RemoveAll()
	assert.Equal(t, []byte("1"), buffer[0])
	assert.Equal(t, []byte("2"), buffer[1])
}

func TestSnapshotsBuffer_AddWhileSizeIsZero(t *testing.T) {
	t.Parallel()

	pb := NewPruningBuffer(0)
	pb.Add([]byte("0"))
	assert.Equal(t, 0, pb.Len())
	pb.Add([]byte("1"))
	assert.Equal(t, 0, pb.Len())
}
