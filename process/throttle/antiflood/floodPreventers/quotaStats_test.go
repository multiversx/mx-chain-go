package floodPreventers

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestQuotaStats(printInterval time.Duration, currentTime *time.Time) *quotaStats {
	qs := newQuotaStats(printInterval)
	qs.getTimeHandler = func() time.Time {
		return *currentTime
	}
	qs.windowStart = *currentTime

	return qs
}

func TestNewQuotaStats(t *testing.T) {
	t.Parallel()

	qs := newQuotaStats(time.Minute)
	require.NotNil(t, qs)
	assert.Equal(t, time.Minute, qs.printInterval)
	assert.False(t, qs.windowStart.IsZero())
	assert.False(t, qs.shouldPrint())
}

func TestQuotaStats_AddRejectedMessage(t *testing.T) {
	t.Parallel()

	currentTime := time.Now()
	qs := createTestQuotaStats(time.Minute, &currentTime)

	qs.addRejectedMessage(true)
	qs.addRejectedMessage(false)
	qs.addRejectedMessage(false)
	qs.addRejectedMessage(true)

	assert.Equal(t, uint64(4), qs.numRejectedMessages)
	assert.Equal(t, uint32(2), qs.numPeersReachingQuota)
}

func TestQuotaStats_AddIntervalPeaksShouldKeepTheHighestValues(t *testing.T) {
	t.Parallel()

	currentTime := time.Now()
	qs := createTestQuotaStats(time.Minute, &currentTime)

	qs.addIntervalPeaks(3, peerPeak{numMessages: 10, pid: "pid1"}, peerPeak{size: 100, pid: "pid2"})
	qs.addIntervalPeaks(7, peerPeak{numMessages: 4, pid: "pid3"}, peerPeak{size: 900, pid: "pid4"})
	qs.addIntervalPeaks(1, peerPeak{numMessages: 40, pid: "pid5"}, peerPeak{size: 50, pid: "pid6"})

	assert.Equal(t, uint32(3), qs.numIntervals)
	assert.Equal(t, 7, qs.peakNumPeers)
	assert.Equal(t, uint32(40), qs.peakNumMessages.numMessages)
	assert.Equal(t, core.PeerID("pid5"), qs.peakNumMessages.pid)
	assert.Equal(t, uint64(900), qs.peakSize.size)
	assert.Equal(t, core.PeerID("pid4"), qs.peakSize.pid)
}

func TestQuotaStats_ShouldPrintAndWindow(t *testing.T) {
	t.Parallel()

	currentTime := time.Now()
	qs := createTestQuotaStats(time.Minute, &currentTime)
	assert.False(t, qs.shouldPrint())

	currentTime = currentTime.Add(time.Minute - time.Nanosecond)
	assert.False(t, qs.shouldPrint())

	currentTime = currentTime.Add(time.Nanosecond)
	assert.True(t, qs.shouldPrint())
	assert.Equal(t, time.Minute, qs.window())
}

func TestQuotaStats_ResetShouldClearValuesAndStartNewWindow(t *testing.T) {
	t.Parallel()

	currentTime := time.Now()
	qs := createTestQuotaStats(time.Minute, &currentTime)
	qs.addRejectedMessage(true)
	qs.addIntervalPeaks(5, peerPeak{numMessages: 10, pid: "pid1"}, peerPeak{size: 100, pid: "pid1"})

	currentTime = currentTime.Add(time.Minute)
	qs.reset()

	assert.Equal(t, uint64(0), qs.numRejectedMessages)
	assert.Equal(t, uint32(0), qs.numPeersReachingQuota)
	assert.Equal(t, uint32(0), qs.numIntervals)
	assert.Equal(t, 0, qs.peakNumPeers)
	assert.Equal(t, peerPeak{}, qs.peakNumMessages)
	assert.Equal(t, peerPeak{}, qs.peakSize)
	assert.Equal(t, currentTime, qs.windowStart)
	// the print interval and the time handler must survive the reset
	assert.Equal(t, time.Minute, qs.printInterval)
	assert.False(t, qs.shouldPrint())
}

func TestComputeUsagePercent(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "n/a", computeUsagePercent(10, 0))
	assert.Equal(t, "0.00%", computeUsagePercent(0, 100))
	assert.Equal(t, "64.29%", computeUsagePercent(180, 280))
	assert.Equal(t, "100.00%", computeUsagePercent(100, 100))
	assert.Equal(t, "150.00%", computeUsagePercent(150, 100))
}
