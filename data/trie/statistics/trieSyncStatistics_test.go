package statistics

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
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

	tss.SetNumMissing([]byte("rh1"), 2)
	assert.Equal(t, 2, tss.NumMissing())

	tss.SetNumMissing([]byte("rh1"), 4)
	assert.Equal(t, 4, tss.NumMissing())

	tss.SetNumMissing([]byte("rh2"), 6)
	assert.Equal(t, 10, tss.NumMissing())

	tss.SetNumMissing([]byte("rh3"), 0)
	assert.Equal(t, 10, tss.NumMissing())

	tss.SetNumMissing([]byte("rh1"), 0)
	assert.Equal(t, 6, tss.NumMissing())

	tss.SetNumMissing([]byte("rh2"), 0)
	assert.Equal(t, 0, tss.NumMissing())

	tss.SetNumMissing([]byte("rh1"), 67)
	assert.Equal(t, 67, tss.NumMissing())

	tss.Reset()
	assert.Equal(t, 0, tss.NumMissing())
}
