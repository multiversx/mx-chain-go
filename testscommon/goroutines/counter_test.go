package goroutines

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoRoutines_SnapshotAll(t *testing.T) {
	t.Parallel()

	handler := func() string {
		buffFile1, err := os.ReadFile("testdata/test1.data")
		require.Nil(t, err)

		return string(buffFile1)
	}

	gc := NewGoCounterWithHandler(handler, AllPassFilter)
	diff := gc.DiffGoRoutines(-1, 1)
	assert.Equal(t, 0, len(diff))

	idx, err := gc.Snapshot()
	require.Nil(t, err)
	assert.Equal(t, 0, idx)

	diff = gc.DiffGoRoutines(idx, -1)
	assert.Equal(t, 6, len(diff))
}

func TestGoRoutines_SnapshotAllDifferentFiles(t *testing.T) {
	t.Parallel()

	filename := "testdata/test2.data"
	handler := func() string {
		buffFile, err := os.ReadFile(filename)
		require.Nil(t, err)

		return string(buffFile)
	}

	gc := NewGoCounterWithHandler(handler, AllPassFilter)
	idx1, err := gc.Snapshot()
	require.Nil(t, err)
	assert.Equal(t, 0, idx1)

	filename = "testdata/test3.data"

	idx2, err := gc.Snapshot()
	require.Nil(t, err)
	assert.Equal(t, 1, idx2)

	idx3, err := gc.Snapshot()
	require.Nil(t, err)
	assert.Equal(t, 2, idx3)

	diff := gc.DiffGoRoutines(idx1, idx2)
	assert.Equal(t, 2, len(diff))

	diff = gc.DiffGoRoutines(idx2, idx3)
	assert.Equal(t, 0, len(diff))
}

func TestGoRoutines_Reset(t *testing.T) {
	t.Parallel()

	filename := "testdata/test2.data"
	handler := func() string {
		buffFile, err := os.ReadFile(filename)
		require.Nil(t, err)

		return string(buffFile)
	}

	gc := NewGoCounterWithHandler(handler, AllPassFilter)
	idx1, err := gc.Snapshot()
	require.Nil(t, err)
	assert.Equal(t, 0, idx1)
	assert.Equal(t, 1, len(gc.snapshots))

	gc.Reset()
	assert.Equal(t, 0, len(gc.snapshots))
}
