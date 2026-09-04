package floodPreventers

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
)

const percentMultiplier = 100.0

// peerPeak holds the highest value recorded for a single peer along with the peer that produced it
type peerPeak struct {
	numMessages uint32
	size        uint64
	pid         core.PeerID
}

// quotaStats holds the statistics accumulated by a quotaFloodPreventer during one of its intervals. The values
// are gathered while the interval is running and are printed and cleared when the flood preventer resets its
// data, so the reported window always matches the interval the configured antiflood limits apply to.
// All its methods are called while the quotaFloodPreventer's mutOperation is locked, so it holds no mutex.
type quotaStats struct {
	windowStart           time.Time
	numRejectedMessages   uint64
	numPeersReachingQuota uint32
	getTimeHandler        func() time.Time
}

func newQuotaStats() *quotaStats {
	qs := &quotaStats{
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

// window returns the time elapsed since these statistics started being accumulated
func (qs *quotaStats) window() time.Duration {
	return qs.getTimeHandler().Sub(qs.windowStart)
}

// reset clears the gathered values and starts a new window
func (qs *quotaStats) reset() {
	*qs = quotaStats{
		windowStart:    qs.getTimeHandler(),
		getTimeHandler: qs.getTimeHandler,
	}
}

func computeUsagePercent(counted uint64, max uint64) string {
	if max == 0 {
		return "n/a"
	}

	return fmt.Sprintf("%.2f%%", float64(counted)*percentMultiplier/float64(max))
}
