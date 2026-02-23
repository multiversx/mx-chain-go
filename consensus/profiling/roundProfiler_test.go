package profiling

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRoundProfiler_EmptyFolderPath(t *testing.T) {
	t.Parallel()

	rp, err := NewRoundProfiler(ArgRoundProfiler{
		FolderPath:    "",
		ShardID:       0,
		RoundsPerFile: 1,
	})

	assert.Nil(t, rp)
	assert.Equal(t, errEmptyFolderPath, err)
}

func TestNewRoundProfiler_InvalidRoundsPerFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	rp, err := NewRoundProfiler(ArgRoundProfiler{
		FolderPath:    dir,
		ShardID:       0,
		RoundsPerFile: 0,
	})

	assert.Nil(t, rp)
	assert.Equal(t, errInvalidRoundsPerFile, err)
}

func TestNewRoundProfiler_ShouldWork(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	folderPath := filepath.Join(dir, "profiles")

	rp, err := NewRoundProfiler(ArgRoundProfiler{
		FolderPath:    folderPath,
		ShardID:       1,
		RoundsPerFile: 1,
	})

	require.Nil(t, err)
	require.NotNil(t, rp)

	// directory should have been created
	info, err := os.Stat(folderPath)
	require.Nil(t, err)
	assert.True(t, info.IsDir())
}

// NOTE: tests below do NOT use t.Parallel() because pprof.StartCPUProfile is process-global;
// only one CPU profile can be active at a time.

func TestRoundProfiler_OnRoundStartCreatesFile(t *testing.T) {
	dir := t.TempDir()
	folderPath := filepath.Join(dir, "profiles")

	rp, err := NewRoundProfiler(ArgRoundProfiler{
		FolderPath:    folderPath,
		ShardID:       2,
		RoundsPerFile: 1,
	})
	require.Nil(t, err)

	ts := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	rp.OnRoundStart(42, ts)

	// stop the profile so the file is flushed
	err = rp.Close()
	require.Nil(t, err)

	entries, err := os.ReadDir(folderPath)
	require.Nil(t, err)
	require.Equal(t, 1, len(entries))

	name := entries[0].Name()
	assert.True(t, strings.HasPrefix(name, "round_42_"))
	assert.True(t, strings.HasSuffix(name, "_2.pprof"))
	assert.True(t, strings.Contains(name, "20260115103000"))

	// file should have content
	info, err := entries[0].Info()
	require.Nil(t, err)
	assert.True(t, info.Size() > 0)
}

func TestRoundProfiler_MultipleRoundsCreateMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	folderPath := filepath.Join(dir, "profiles")

	rp, err := NewRoundProfiler(ArgRoundProfiler{
		FolderPath:    folderPath,
		ShardID:       0,
		RoundsPerFile: 1,
	})
	require.Nil(t, err)

	ts1 := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	ts2 := time.Date(2026, 1, 15, 10, 30, 6, 0, time.UTC)

	rp.OnRoundStart(100, ts1)
	rp.OnRoundStart(101, ts2)

	err = rp.Close()
	require.Nil(t, err)

	entries, err := os.ReadDir(folderPath)
	require.Nil(t, err)
	assert.Equal(t, 2, len(entries))
}

func TestRoundProfiler_GroupedRounds(t *testing.T) {
	dir := t.TempDir()
	folderPath := filepath.Join(dir, "profiles")

	rp, err := NewRoundProfiler(ArgRoundProfiler{
		FolderPath:    folderPath,
		ShardID:       0,
		RoundsPerFile: 3,
	})
	require.Nil(t, err)

	baseTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	// rounds 100-102 should all go into one file
	rp.OnRoundStart(100, baseTime)
	rp.OnRoundStart(101, baseTime.Add(6*time.Second))
	rp.OnRoundStart(102, baseTime.Add(12*time.Second))

	// round 103 should trigger a new file
	rp.OnRoundStart(103, baseTime.Add(18*time.Second))

	err = rp.Close()
	require.Nil(t, err)

	entries, err := os.ReadDir(folderPath)
	require.Nil(t, err)
	assert.Equal(t, 2, len(entries))

	// first file is named after round 100, second after round 103
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	assert.True(t, strings.HasPrefix(names[0], "round_100_"))
	assert.True(t, strings.HasPrefix(names[1], "round_103_"))
}

func TestRoundProfiler_GroupedRoundsExactBoundary(t *testing.T) {
	dir := t.TempDir()
	folderPath := filepath.Join(dir, "profiles")

	rp, err := NewRoundProfiler(ArgRoundProfiler{
		FolderPath:    folderPath,
		ShardID:       0,
		RoundsPerFile: 2,
	})
	require.Nil(t, err)

	baseTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	// 4 rounds with grouping of 2 should produce 2 files
	rp.OnRoundStart(10, baseTime)
	rp.OnRoundStart(11, baseTime.Add(6*time.Second))
	rp.OnRoundStart(12, baseTime.Add(12*time.Second))
	rp.OnRoundStart(13, baseTime.Add(18*time.Second))

	err = rp.Close()
	require.Nil(t, err)

	entries, err := os.ReadDir(folderPath)
	require.Nil(t, err)
	assert.Equal(t, 2, len(entries))
}

func TestRoundProfiler_CloseWithoutStart(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	folderPath := filepath.Join(dir, "profiles")

	rp, err := NewRoundProfiler(ArgRoundProfiler{
		FolderPath:    folderPath,
		ShardID:       0,
		RoundsPerFile: 1,
	})
	require.Nil(t, err)

	// Close without any round started should not panic
	err = rp.Close()
	assert.Nil(t, err)
}

func TestRoundProfiler_DoubleClose(t *testing.T) {
	dir := t.TempDir()
	folderPath := filepath.Join(dir, "profiles")

	rp, err := NewRoundProfiler(ArgRoundProfiler{
		FolderPath:    folderPath,
		ShardID:       0,
		RoundsPerFile: 1,
	})
	require.Nil(t, err)

	ts := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	rp.OnRoundStart(1, ts)

	err = rp.Close()
	assert.Nil(t, err)

	// second close should also be safe
	err = rp.Close()
	assert.Nil(t, err)
}

func TestDisabledRoundProfiler(t *testing.T) {
	t.Parallel()

	dp := NewDisabledRoundProfiler()
	require.NotNil(t, dp)
	assert.False(t, dp.IsInterfaceNil())

	// these should not panic
	dp.OnRoundStart(1, time.Now())
	err := dp.Close()
	assert.Nil(t, err)
}

func TestRoundProfiler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var rp *roundProfiler
	assert.True(t, rp.IsInterfaceNil())

	dir := t.TempDir()
	rp, _ = NewRoundProfiler(ArgRoundProfiler{
		FolderPath:    dir,
		ShardID:       0,
		RoundsPerFile: 1,
	})
	assert.False(t, rp.IsInterfaceNil())
}
