package floodPreventers

import (
	"math"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createArgumentWithLimits(maxMessages uint32, maxSize uint64) ArgQuotaFloodPreventer {
	arg := createDefaultArgument()
	arg.Cacher = cache.NewCacherMock()
	arg.AntifloodConfigs = &testscommon.AntifloodConfigsHandlerStub{
		GetFloodPreventerConfigByTypeCalled: func(configType common.FloodPreventerType) config.FloodPreventerConfig {
			return config.FloodPreventerConfig{
				IntervalInSeconds: 1,
				ReservedPercent:   0,
				PeerMaxInput: config.AntifloodLimitsConfig{
					BaseMessagesPerInterval: maxMessages,
					TotalSizePerInterval:    maxSize,
				},
			}
		},
	}

	return arg
}

// setDebugLogLevel raises the log level of this package so the statistics are gathered, returning the
// function that restores the previous level
func setDebugLogLevel() func() {
	oldLevel := log.GetLevel()
	log.SetLevel(logger.LogDebug)

	return func() {
		log.SetLevel(oldLevel)
	}
}

func TestQuotaFloodPreventer_QuotaReachedShouldBeCountedOncePerPeerPerInterval(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(createArgumentWithLimits(2, math.MaxUint64))
	require.Nil(t, err)

	pid1 := core.PeerID("pid1")
	pid2 := core.PeerID("pid2")

	// first 2 messages of each peer are within the quota
	assert.Nil(t, qfp.IncreaseLoad(pid1, 1))
	assert.Nil(t, qfp.IncreaseLoad(pid1, 1))
	assert.Nil(t, qfp.IncreaseLoad(pid2, 1))
	assert.Nil(t, qfp.IncreaseLoad(pid2, 1))

	assert.Equal(t, uint64(0), qfp.stats.numRejectedMessages)
	assert.Equal(t, uint32(0), qfp.stats.numPeersReachingQuota)

	// each of the following ones is rejected, but each peer is counted a single time
	for i := 0; i < 5; i++ {
		assert.NotNil(t, qfp.IncreaseLoad(pid1, 1))
	}
	assert.NotNil(t, qfp.IncreaseLoad(pid2, 1))

	assert.Equal(t, uint64(6), qfp.stats.numRejectedMessages)
	assert.Equal(t, uint32(2), qfp.stats.numPeersReachingQuota)
}

func TestQuotaFloodPreventer_QuotaReachedFlagShouldBeClearedOnReset(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(createArgumentWithLimits(1, math.MaxUint64))
	require.Nil(t, err)

	pid := core.PeerID("pid")
	assert.Nil(t, qfp.IncreaseLoad(pid, 1))
	assert.NotNil(t, qfp.IncreaseLoad(pid, 1))
	assert.Equal(t, uint32(1), qfp.stats.numPeersReachingQuota)

	qfp.Reset()
	assert.Equal(t, uint32(0), qfp.stats.numPeersReachingQuota)

	// after the reset the peer gets a fresh quota, so it will be counted again when reaching it
	assert.Nil(t, qfp.IncreaseLoad(pid, 1))
	assert.NotNil(t, qfp.IncreaseLoad(pid, 1))
	assert.Equal(t, uint32(1), qfp.stats.numPeersReachingQuota)
}

func TestQuotaFloodPreventer_SizeQuotaReachedShouldBeRecorded(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(createArgumentWithLimits(math.MaxUint32, 100))
	require.Nil(t, err)

	pid := core.PeerID("pid")
	assert.Nil(t, qfp.IncreaseLoad(pid, 60))
	assert.NotNil(t, qfp.IncreaseLoad(pid, 60))

	assert.Equal(t, uint64(1), qfp.stats.numRejectedMessages)
	assert.Equal(t, uint32(1), qfp.stats.numPeersReachingQuota)
}

func TestQuotaFloodPreventer_CreateStatisticsShouldReturnPeaks(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(createArgumentWithLimits(math.MaxUint32, math.MaxUint64))
	require.Nil(t, err)

	// pid1 sends the most messages, pid2 sends the most bytes
	for i := 0; i < 5; i++ {
		assert.Nil(t, qfp.IncreaseLoad("pid1", 10))
	}
	assert.Nil(t, qfp.IncreaseLoad("pid2", 1000))
	assert.Nil(t, qfp.IncreaseLoad("pid3", 5))

	numPeers, peakNumMessages, peakSize := qfp.createStatistics()

	assert.Equal(t, 3, numPeers)
	assert.Equal(t, uint32(5), peakNumMessages.numMessages)
	assert.Equal(t, core.PeerID("pid1"), peakNumMessages.pid)
	assert.Equal(t, uint64(1000), peakSize.size)
	assert.Equal(t, core.PeerID("pid2"), peakSize.pid)
}

func TestQuotaFloodPreventer_ResetShouldReportTheElapsedIntervalAndStartANewOne(t *testing.T) {
	defer setDebugLogLevel()()

	qfp, err := NewQuotaFloodPreventer(createArgumentWithLimits(2, math.MaxUint64))
	require.Nil(t, err)

	currentTime := time.Now()
	qfp.stats.getTimeHandler = func() time.Time {
		return currentTime
	}
	qfp.stats.windowStart = currentTime

	assert.Nil(t, qfp.IncreaseLoad("pid1", 10))
	assert.Nil(t, qfp.IncreaseLoad("pid1", 10))
	assert.NotNil(t, qfp.IncreaseLoad("pid1", 10))
	assert.Nil(t, qfp.IncreaseLoad("pid2", 7))

	currentTime = currentTime.Add(time.Second)

	shouldPrint, args := qfp.resetAndGatherStatistics()
	require.True(t, shouldPrint)
	assertLogArg(t, args, "window", time.Second)
	assertLogArg(t, args, "num peers", 2)
	assertLogArg(t, args, "peak num messages/peer", uint32(3))
	assertLogArg(t, args, "max num messages/peer", uint64(2))
	assertLogArg(t, args, "num messages usage", "150.00%")
	assertLogArg(t, args, "num rejected messages", uint64(1))
	assertLogArg(t, args, "num peers reaching quota", uint32(1))

	// a new window starts, holding no data from the previous one
	assert.Equal(t, currentTime, qfp.stats.windowStart)
	assert.Equal(t, uint64(0), qfp.stats.numRejectedMessages)
	assert.Equal(t, uint32(0), qfp.stats.numPeersReachingQuota)

	shouldPrint, args = qfp.resetAndGatherStatistics()
	assert.False(t, shouldPrint)
	assert.Nil(t, args)
}

func TestQuotaFloodPreventer_ResetShouldNotPrintOnNoTraffic(t *testing.T) {
	defer setDebugLogLevel()()

	qfp, err := NewQuotaFloodPreventer(createArgumentWithLimits(2, math.MaxUint64))
	require.Nil(t, err)

	shouldPrint, args := qfp.resetAndGatherStatistics()
	assert.False(t, shouldPrint)
	assert.Nil(t, args)
}

func TestQuotaFloodPreventer_ResetShouldNotPrintOnHigherLogLevel(t *testing.T) {
	oldLevel := log.GetLevel()
	log.SetLevel(logger.LogInfo)
	defer log.SetLevel(oldLevel)

	qfp, err := NewQuotaFloodPreventer(createArgumentWithLimits(2, math.MaxUint64))
	require.Nil(t, err)

	assert.Nil(t, qfp.IncreaseLoad("pid1", 10))

	shouldPrint, args := qfp.resetAndGatherStatistics()
	assert.False(t, shouldPrint)
	assert.Nil(t, args)
}

func assertLogArg(tb testing.TB, args []interface{}, key string, expectedValue interface{}) {
	for i := 0; i < len(args)-1; i += 2 {
		if args[i] == key {
			assert.Equal(tb, expectedValue, args[i+1], "for log key %s", key)
			return
		}
	}

	assert.Fail(tb, "log key not found", key)
}
