package statistics

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieSyncStatistics_ShouldWork(t *testing.T) {
	t.Parallel()

	tss := NewTrieSyncStatistics()

	assert.False(t, check.IfNil(tss))
}

func TestTrieSyncStatistics_Received(t *testing.T) {
	t.Parallel()

	tss := NewTrieSyncStatistics()

	assert.Equal(t, 0, tss.NumReceived())

	tss.AddNumReceived(2)
	assert.Equal(t, 2, tss.NumReceived())

	tss.AddNumReceived(4)
	assert.Equal(t, 6, tss.NumReceived())

	tss.Reset()
	assert.Equal(t, 0, tss.NumReceived())
}

func TestTrieSyncStatistics_Missing(t *testing.T) {
	t.Parallel()

	tss := NewTrieSyncStatistics()

	assert.Equal(t, 0, tss.NumMissing())
	assert.Equal(t, 0, tss.NumTries())

	tss.SetNumMissing([]byte("rh1"), 2)
	assert.Equal(t, 2, tss.NumMissing())
	assert.Equal(t, 1, tss.NumTries())

	tss.SetNumMissing([]byte("rh1"), 4)
	assert.Equal(t, 4, tss.NumMissing())
	assert.Equal(t, 1, tss.NumTries())

	tss.SetNumMissing([]byte("rh2"), 6)
	assert.Equal(t, 10, tss.NumMissing())
	assert.Equal(t, 2, tss.NumTries())

	tss.SetNumMissing([]byte("rh3"), 0)
	assert.Equal(t, 10, tss.NumMissing())
	assert.Equal(t, 2, tss.NumTries())

	tss.SetNumMissing([]byte("rh1"), 0)
	assert.Equal(t, 6, tss.NumMissing())
	assert.Equal(t, 1, tss.NumTries())

	tss.SetNumMissing([]byte("rh2"), 0)
	assert.Equal(t, 0, tss.NumMissing())
	assert.Equal(t, 0, tss.NumTries())

	tss.SetNumMissing([]byte("rh1"), 67)
	assert.Equal(t, 67, tss.NumMissing())
	assert.Equal(t, 1, tss.NumTries())

	tss.Reset()
	assert.Equal(t, 0, tss.NumMissing())
	assert.Equal(t, 0, tss.NumTries())
}

func TestTrieSyncStatistics_Large(t *testing.T) {
	t.Parallel()

	tss := NewTrieSyncStatistics()

	assert.Equal(t, 0, tss.NumLarge())

	tss.AddNumLarge(2)
	assert.Equal(t, 2, tss.NumLarge())

	tss.AddNumLarge(4)
	assert.Equal(t, 6, tss.NumLarge())

	tss.Reset()
	assert.Equal(t, 0, tss.NumLarge())
}

func TestTrieSyncStatistics_BytesReceived(t *testing.T) {
	t.Parallel()

	tss := NewTrieSyncStatistics()

	assert.Equal(t, uint64(0), tss.NumBytesReceived())

	tss.AddNumBytesReceived(2)
	assert.Equal(t, uint64(2), tss.NumBytesReceived())

	tss.AddNumBytesReceived(4)
	assert.Equal(t, uint64(6), tss.NumBytesReceived())

	tss.Reset()
	assert.Equal(t, uint64(0), tss.NumBytesReceived())
}
