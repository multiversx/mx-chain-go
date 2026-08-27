package floodPreventers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestQuotaStats(currentTime *time.Time) *quotaStats {
	qs := newQuotaStats()
	qs.getTimeHandler = func() time.Time {
		return *currentTime
	}
	qs.windowStart = *currentTime

	return qs
}

func TestNewQuotaStats(t *testing.T) {
	t.Parallel()

	qs := newQuotaStats()
	require.NotNil(t, qs)
	assert.False(t, qs.windowStart.IsZero())
	assert.Equal(t, uint64(0), qs.numRejectedMessages)
	assert.Equal(t, uint32(0), qs.numPeersReachingQuota)
}

func TestQuotaStats_AddRejectedMessage(t *testing.T) {
	t.Parallel()

	currentTime := time.Now()
	qs := createTestQuotaStats(&currentTime)

	qs.addRejectedMessage(true)
	qs.addRejectedMessage(false)
	qs.addRejectedMessage(false)
	qs.addRejectedMessage(true)

	assert.Equal(t, uint64(4), qs.numRejectedMessages)
	assert.Equal(t, uint32(2), qs.numPeersReachingQuota)
}

func TestQuotaStats_Window(t *testing.T) {
	t.Parallel()

	currentTime := time.Now()
	qs := createTestQuotaStats(&currentTime)
	assert.Equal(t, time.Duration(0), qs.window())

	currentTime = currentTime.Add(time.Second + time.Millisecond)
	assert.Equal(t, time.Second+time.Millisecond, qs.window())
}

func TestQuotaStats_ResetShouldClearValuesAndStartNewWindow(t *testing.T) {
	t.Parallel()

	currentTime := time.Now()
	qs := createTestQuotaStats(&currentTime)
	qs.addRejectedMessage(true)
	qs.addRejectedMessage(false)

	currentTime = currentTime.Add(time.Second)
	qs.reset()

	assert.Equal(t, uint64(0), qs.numRejectedMessages)
	assert.Equal(t, uint32(0), qs.numPeersReachingQuota)
	assert.Equal(t, currentTime, qs.windowStart)
	assert.Equal(t, time.Duration(0), qs.window())
	// the time handler must survive the reset
	require.NotNil(t, qs.getTimeHandler)
}

func TestComputeUsagePercent(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "n/a", computeUsagePercent(10, 0))
	assert.Equal(t, "0.00%", computeUsagePercent(0, 100))
	assert.Equal(t, "64.29%", computeUsagePercent(180, 280))
	assert.Equal(t, "100.00%", computeUsagePercent(100, 100))
	assert.Equal(t, "150.00%", computeUsagePercent(150, 100))
}
