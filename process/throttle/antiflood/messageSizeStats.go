package antiflood

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const newLine = "\n"
const statsHeader = "p2p incoming messages size statistics"

// messageSizeStatsInterval represents the period of time after which the gathered message size statistics
// are printed and reset
var messageSizeStatsInterval = 5 * time.Minute

type topicSizeStats struct {
	numMessages uint64
	totalSize   uint64
	maxSize     uint64
	maxSizePid  core.PeerID
}

func (tss *topicSizeStats) averageSize() uint64 {
	if tss.numMessages == 0 {
		return 0
	}

	return tss.totalSize / tss.numMessages
}

// messageSizeStatsCollector gathers the size of all incoming p2p messages, regardless of the fact that the
// messages were or were not rejected by the flood preventers. On each elapsed interval, the gathered values
// are printed and reset, providing the data needed to tune the antiflood configuration.
// The collection is done only if the log level for this package allows the printing of the statistics.
type messageSizeStatsCollector struct {
	mut             sync.Mutex
	interval        time.Duration
	intervalStart   time.Time
	stats           map[string]*topicSizeStats
	allTimeMaxSizes map[string]uint64
	getTimeHandler  func() time.Time
}

func newMessageSizeStatsCollector(interval time.Duration) *messageSizeStatsCollector {
	collector := &messageSizeStatsCollector{
		interval:        interval,
		stats:           make(map[string]*topicSizeStats),
		allTimeMaxSizes: make(map[string]uint64),
		getTimeHandler:  time.Now,
	}
	collector.intervalStart = collector.getTimeHandler()

	return collector
}

// addMessage records the size of a received message on the provided topic. If the current interval elapsed,
// the accumulated statistics are printed and a new interval is started.
func (msc *messageSizeStatsCollector) addMessage(topic string, size uint64, pid core.PeerID) {
	if log.GetLevel() > logger.LogDebug {
		return
	}
	if len(topic) == 0 {
		topic = unidentifiedTopic
	}

	msc.mut.Lock()
	msc.accumulate(topic, size, pid)
	report := msc.reportIfIntervalElapsed()
	msc.mut.Unlock()

	if len(report) > 0 {
		log.Debug(report)
	}
}

// accumulate must be called with the mutex locked
func (msc *messageSizeStatsCollector) accumulate(topic string, size uint64, pid core.PeerID) {
	stats, found := msc.stats[topic]
	if !found {
		stats = &topicSizeStats{}
		msc.stats[topic] = stats
	}

	stats.numMessages++
	stats.totalSize += size
	if size > stats.maxSize {
		stats.maxSize = size
		stats.maxSizePid = pid
	}
	if size > msc.allTimeMaxSizes[topic] {
		msc.allTimeMaxSizes[topic] = size
	}
}

// reportIfIntervalElapsed must be called with the mutex locked
func (msc *messageSizeStatsCollector) reportIfIntervalElapsed() string {
	now := msc.getTimeHandler()
	elapsed := now.Sub(msc.intervalStart)
	if elapsed < msc.interval {
		return ""
	}

	report := msc.createReport(elapsed)
	msc.intervalStart = now
	msc.stats = make(map[string]*topicSizeStats)

	return report
}

// createReport must be called with the mutex locked
func (msc *messageSizeStatsCollector) createReport(elapsed time.Duration) string {
	if len(msc.stats) == 0 {
		return ""
	}

	topics := make([]string, 0, len(msc.stats))
	for topic := range msc.stats {
		topics = append(topics, topic)
	}

	sort.Slice(topics, func(i, j int) bool {
		statsI, statsJ := msc.stats[topics[i]], msc.stats[topics[j]]
		if statsI.totalSize == statsJ.totalSize {
			return topics[i] < topics[j]
		}

		return statsI.totalSize > statsJ.totalSize
	})

	total := &topicSizeStats{}
	lines := make([]string, 0, len(topics)+2)
	lines = append(lines, fmt.Sprintf("%s gathered in %v:", statsHeader, elapsed.Truncate(time.Second)))
	for _, topic := range topics {
		stats := msc.stats[topic]
		lines = append(lines, msc.createTopicLine(topic, stats))

		total.numMessages += stats.numMessages
		total.totalSize += stats.totalSize
		if stats.maxSize > total.maxSize {
			total.maxSize = stats.maxSize
			total.maxSizePid = stats.maxSizePid
		}
	}

	lines = append(lines, fmt.Sprintf("all topics; num messages: %d; total size: %s; average size: %s; max size: %s",
		total.numMessages,
		core.ConvertBytes(total.totalSize),
		core.ConvertBytes(total.averageSize()),
		core.ConvertBytes(total.maxSize),
	))

	return strings.Join(lines, newLine)
}

func (msc *messageSizeStatsCollector) createTopicLine(topic string, stats *topicSizeStats) string {
	return fmt.Sprintf("topic: %s; num messages: %d; total size: %s; average size: %s; max size: %s from pid: %s; all time max size: %s",
		topic,
		stats.numMessages,
		core.ConvertBytes(stats.totalSize),
		core.ConvertBytes(stats.averageSize()),
		core.ConvertBytes(stats.maxSize),
		stats.maxSizePid.Pretty(),
		core.ConvertBytes(msc.allTimeMaxSizes[topic]),
	)
}
