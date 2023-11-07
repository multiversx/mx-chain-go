package goroutines

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSnapshot(t *testing.T) {
	t.Parallel()

	buff, err := os.ReadFile("testdata/test1.data")
	require.Nil(t, err)
	s, err := newSnapshot(string(buff), AllPassFilter)
	assert.Nil(t, err)
	assert.Equal(t, 6, len(s.routines))
}

func TestNewSnapshot_WithFilter(t *testing.T) {
	t.Parallel()

	buff, err := os.ReadFile("testdata/test1.data")
	require.Nil(t, err)
	s, err := newSnapshot(string(buff), TestsRelevantGoRoutines)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(s.routines))
}

func TestSnapshot_Diff(t *testing.T) {
	t.Parallel()

	buffFile2, err := os.ReadFile("testdata/test2.data")
	require.Nil(t, err)
	snapshot2, err := newSnapshot(string(buffFile2), AllPassFilter)
	require.Nil(t, err)

	buffFile3, err := os.ReadFile("testdata/test3.data")
	require.Nil(t, err)
	snapshot3, err := newSnapshot(string(buffFile3), AllPassFilter)
	require.Nil(t, err)

	diff := snapshot2.diff(nil)
	assert.Equal(t, 5, len(diff))

	diff = snapshot2.diff(snapshot3)
	assert.Equal(t, 2, len(diff))
	assert.Equal(t, "15", diff[0].ID())
	assert.Equal(t, "14", diff[1].ID())
}
