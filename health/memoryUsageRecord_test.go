package health

import (
	"runtime"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/require"
)

func TestMemoryUsageRecord_GetFilename(t *testing.T) {
	timestamp, err := time.Parse("20060102150405", "20200621000000")
	require.Nil(t, err)

	record := newMemoryUsageRecord(runtime.MemStats{HeapInuse: 42 * core.MegabyteSize}, timestamp, ".")
	require.Equal(t, "mem__20200621000000__42_00_MB.pprof", record.getFilename())
}

func TestMemoryUsageRecord_SaveThenDelete(t *testing.T) {
	record := newMemoryUsageRecord(runtime.MemStats{HeapInuse: 42 * core.MegabyteSize}, time.Now(), ".")
	filename := record.getFilename()

	err := record.save()
	require.Nil(t, err)
	require.FileExists(t, filename)

	err = record.delete()
	require.Nil(t, err)
	require.NoFileExists(t, filename)
}

func TestMemoryUsageRecord_IsMoreImportantThan(t *testing.T) {
	a := newMemoryUsageRecord(runtime.MemStats{HeapInuse: 42}, time.Now(), ".")
	b := newMemoryUsageRecord(runtime.MemStats{HeapInuse: 41}, time.Now(), ".")
	c := newMemoryUsageRecord(runtime.MemStats{HeapInuse: 42}, time.Now(), ".")

	require.True(t, a.isMoreImportantThan(b))
	require.True(t, c.isMoreImportantThan(b))

	require.False(t, b.isMoreImportantThan(a))
	require.False(t, b.isMoreImportantThan(c))

	require.False(t, a.isMoreImportantThan(c))
	require.False(t, c.isMoreImportantThan(a))

	// Same record
	require.False(t, a.isMoreImportantThan(a))

	// Different type of record
	require.False(t, a.isMoreImportantThan(newDummyRecord(12345)))
}
