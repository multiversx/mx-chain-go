package floodPreventers

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
)

// quotaStatsPrintInterval represents the period of time after which the gathered quota statistics are
// printed and reset. It is intentionally larger than the reset interval of any flood preventer, so a
// fast reacting preventer (that resets each second) will not spam the log with one line per second
var quotaStatsPrintInterval = 5 * time.Minute

const percentMultiplier = 100.0

// peerPeak holds the highest value recorded for a single peer along with the peer that produced it
type peerPeak struct {
	numMessages uint32
	size        uint64
	pid         core.PeerID
}

// quotaStats aggregates, over multiple flood preventer intervals, the peak load produced by a single peer.
// It provides the data needed to tune the antiflood configuration: how close the peers actually get to the
// configured maximum values and how often those values are exceeded.
// All its methods are called while the quotaFloodPreventer's mutOperation is locked, so it holds no mutex.
type quotaStats struct {
	windowStart           time.Time
	numIntervals          uint32
	peakNumPeers          int
	peakNumMessages       peerPeak
	peakSize              peerPeak
	numRejectedMessages   uint64
	numPeersReachingQuota uint32
	printInterval         time.Duration
	getTimeHandler        func() time.Time
}

func newQuotaStats(printInterval time.Duration) *quotaStats {
	qs := &quotaStats{
		printInterval:  printInterval,
		getTimeHandler: time.Now,
	}
	qs.windowStart = qs.getTimeHandler()

	return qs
}

// addRejectedMessage records the fact that a message was rejected because the peer's quota was reached.
// firstForPeerInInterval should be true only the first time a peer reaches its quota in the current interval
func (qs *quotaStats) addRejectedMessage(firstForPeerInInterval bool) {
	qs.numRejectedMessages++
	if firstForPeerInInterval {
		qs.numPeersReachingQuota++
	}
}

// addIntervalPeaks records the peak values measured over a completed flood preventer interval
func (qs *quotaStats) addIntervalPeaks(numPeers int, peakNumMessages peerPeak, peakSize peerPeak) {
	qs.numIntervals++

	if numPeers > qs.peakNumPeers {
		qs.peakNumPeers = numPeers
	}
	if peakNumMessages.numMessages > qs.peakNumMessages.numMessages {
		qs.peakNumMessages = peakNumMessages
	}
	if peakSize.size > qs.peakSize.size {
		qs.peakSize = peakSize
	}
}

// shouldPrint returns true if the print interval elapsed
func (qs *quotaStats) shouldPrint() bool {
	return qs.getTimeHandler().Sub(qs.windowStart) >= qs.printInterval
}

// window returns the time elapsed since the last reset of these statistics
func (qs *quotaStats) window() time.Duration {
	return qs.getTimeHandler().Sub(qs.windowStart).Truncate(time.Second)
}

// reset clears the gathered values and starts a new window
func (qs *quotaStats) reset() {
	now := qs.getTimeHandler()
	*qs = quotaStats{
		windowStart:    now,
		printInterval:  qs.printInterval,
		getTimeHandler: qs.getTimeHandler,
	}
}

func computeUsagePercent(counted uint64, max uint64) string {
	if max == 0 {
		return "n/a"
	}

	return fmt.Sprintf("%.2f%%", float64(counted)*percentMultiplier/float64(max))
}
