package antiflood

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestCollector(interval time.Duration, currentTime *time.Time) *messageSizeStatsCollector {
	collector := newMessageSizeStatsCollector(interval)
	collector.getTimeHandler = func() time.Time {
		return *currentTime
	}
	collector.intervalStart = *currentTime

	return collector
}

func TestNewMessageSizeStatsCollector(t *testing.T) {
	t.Parallel()

	collector := newMessageSizeStatsCollector(time.Second)
	require.NotNil(t, collector)
	assert.Equal(t, time.Second, collector.interval)
	assert.Equal(t, 0, len(collector.stats))
	assert.Equal(t, 0, len(collector.allTimeMaxSizes))
	assert.False(t, collector.intervalStart.IsZero())
}

func TestMessageSizeStatsCollector_AddMessageShouldNotCollectOnHigherLogLevel(t *testing.T) {
	oldLevel := log.GetLevel()
	log.SetLevel(logger.LogInfo)
	defer log.SetLevel(oldLevel)

	currentTime := time.Now()
	collector := createTestCollector(time.Minute, &currentTime)

	collector.addMessage("topic", 10, "pid")

	assert.Equal(t, 0, len(collector.stats))
}

func TestMessageSizeStatsCollector_AddMessageShouldAccumulate(t *testing.T) {
	oldLevel := log.GetLevel()
	log.SetLevel(logger.LogDebug)
	defer log.SetLevel(oldLevel)

	currentTime := time.Now()
	collector := createTestCollector(time.Minute, &currentTime)

	collector.addMessage("topic1", 10, "pid1")
	collector.addMessage("topic1", 40, "pid2")
	collector.addMessage("topic1", 25, "pid3")
	collector.addMessage("topic2", 7, "pid1")

	require.Equal(t, 2, len(collector.stats))

	stats := collector.stats["topic1"]
	assert.Equal(t, uint64(3), stats.numMessages)
	assert.Equal(t, uint64(75), stats.totalSize)
	assert.Equal(t, uint64(40), stats.maxSize)
	assert.Equal(t, core.PeerID("pid2"), stats.maxSizePid)
	assert.Equal(t, uint64(25), stats.averageSize())

	stats = collector.stats["topic2"]
	assert.Equal(t, uint64(1), stats.numMessages)
	assert.Equal(t, uint64(7), stats.totalSize)
	assert.Equal(t, uint64(7), stats.maxSize)
}

func TestMessageSizeStatsCollector_AddMessageEmptyTopicShouldUseUnidentifiedTopic(t *testing.T) {
	oldLevel := log.GetLevel()
	log.SetLevel(logger.LogDebug)
	defer log.SetLevel(oldLevel)

	currentTime := time.Now()
	collector := createTestCollector(time.Minute, &currentTime)

	collector.addMessage("", 10, "pid")

	require.Equal(t, 1, len(collector.stats))
	assert.Equal(t, uint64(10), collector.stats[unidentifiedTopic].totalSize)
}

func TestMessageSizeStatsCollector_AddMessageShouldResetOnElapsedInterval(t *testing.T) {
	oldLevel := log.GetLevel()
	log.SetLevel(logger.LogDebug)
	defer log.SetLevel(oldLevel)

	currentTime := time.Now()
	collector := createTestCollector(time.Minute, &currentTime)

	collector.addMessage("topic", 100, "pid")
	assert.Equal(t, uint64(100), collector.stats["topic"].totalSize)

	currentTime = currentTime.Add(time.Minute)
	collector.addMessage("topic", 10, "pid")

	// the message that triggered the print is accounted in the closed interval, the new interval starts empty
	assert.Equal(t, 0, len(collector.stats))
	assert.Equal(t, currentTime, collector.intervalStart)
	// the all time max size survives the reset
	assert.Equal(t, uint64(100), collector.allTimeMaxSizes["topic"])

	collector.addMessage("topic", 10, "pid")
	assert.Equal(t, uint64(10), collector.stats["topic"].totalSize)
	assert.Equal(t, uint64(100), collector.allTimeMaxSizes["topic"])
}

func TestMessageSizeStatsCollector_ReportIfIntervalElapsed(t *testing.T) {
	t.Parallel()

	t.Run("interval not elapsed should return empty", func(t *testing.T) {
		t.Parallel()

		currentTime := time.Now()
		collector := createTestCollector(time.Minute, &currentTime)
		collector.accumulate("topic", 10, "pid")

		currentTime = currentTime.Add(time.Minute - time.Nanosecond)
		assert.Empty(t, collector.reportIfIntervalElapsed())
		assert.Equal(t, 1, len(collector.stats))
	})
	t.Run("interval elapsed with no data should return empty but restart the interval", func(t *testing.T) {
		t.Parallel()

		currentTime := time.Now()
		collector := createTestCollector(time.Minute, &currentTime)

		currentTime = currentTime.Add(time.Minute)
		assert.Empty(t, collector.reportIfIntervalElapsed())
		assert.Equal(t, currentTime, collector.intervalStart)
	})
	t.Run("interval elapsed should return the report", func(t *testing.T) {
		t.Parallel()

		currentTime := time.Now()
		collector := createTestCollector(time.Minute, &currentTime)
		collector.accumulate("topic", 10, "pid")

		currentTime = currentTime.Add(time.Minute)
		report := collector.reportIfIntervalElapsed()
		assert.True(t, strings.Contains(report, statsHeader))
		assert.True(t, strings.Contains(report, "topic"))
		assert.Equal(t, 0, len(collector.stats))
	})
}

func TestMessageSizeStatsCollector_CreateReport(t *testing.T) {
	t.Parallel()

	t.Run("no data should return empty", func(t *testing.T) {
		t.Parallel()

		currentTime := time.Now()
		collector := createTestCollector(time.Minute, &currentTime)

		assert.Empty(t, collector.createReport(time.Minute))
	})
	t.Run("should sort the topics descending by total size and print the total", func(t *testing.T) {
		t.Parallel()

		currentTime := time.Now()
		collector := createTestCollector(time.Minute, &currentTime)
		collector.accumulate("smallTopic", 10, "pid1")
		collector.accumulate("largeTopic", 2048, "pid2")
		collector.accumulate("largeTopic", 1024, "pid3")

		report := collector.createReport(time.Minute)
		lines := strings.Split(report, newLine)
		require.Equal(t, 4, len(lines))

		assert.True(t, strings.Contains(lines[0], "gathered in 1m0s"))
		assert.True(t, strings.Contains(lines[1], "topic: largeTopic; num messages: 2; total size: 3.00 KB; average size: 1.50 KB; max size: 2.00 KB"))
		assert.True(t, strings.Contains(lines[1], core.PeerID("pid2").Pretty()))
		assert.True(t, strings.Contains(lines[2], "topic: smallTopic; num messages: 1; total size: 10 B"))
		assert.True(t, strings.Contains(lines[3], "all topics; num messages: 3; total size: 3.01 KB"))
	})
}

func TestMessageSizeStatsCollector_ConcurrentOperationsShouldNotPanic(t *testing.T) {
	oldLevel := log.GetLevel()
	log.SetLevel(logger.LogDebug)
	defer log.SetLevel(oldLevel)

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should have not panicked", r)
		}
	}()

	collector := newMessageSizeStatsCollector(time.Millisecond)

	numGoRoutines := 100
	wg := sync.WaitGroup{}
	wg.Add(numGoRoutines)
	for i := 0; i < numGoRoutines; i++ {
		go func(idx int) {
			defer wg.Done()

			collector.addMessage("topic", uint64(idx), core.PeerID("pid"))
		}(i)
	}

	wg.Wait()
}
