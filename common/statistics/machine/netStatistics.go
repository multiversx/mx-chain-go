package machine

import (
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/shirou/gopsutil/net"
)

// netStatistics can compute the network statistics
type netStatistics struct {
	bpsSent                   uint64
	bpsRecv                   uint64
	bpsSentPeak               uint64
	bpsRecvPeak               uint64
	percentSent               uint64
	percentRecv               uint64
	totalBytesSentInEpoch     uint64
	totalBytesReceivedInEpoch uint64
	getStatisticsHandler      func() ([]net.IOCountersStat, error)
}

// NewNetStatistics returns a new instance of the netStatistics
func NewNetStatistics() *netStatistics {
	return &netStatistics{
		getStatisticsHandler: getStatistics,
	}
}

// ComputeStatistics computes the current network statistics usage.
// It should be called on a go routine as it is a blocking call for a bounded time (1 second)
func (ns *netStatistics) ComputeStatistics() {
	nStart, err := ns.getStatisticsHandler()
	if err != nil {
		ns.setZeroStatsAndWait()
		return
	}
	if len(nStart) == 0 {
		ns.setZeroStatsAndWait()
		return
	}

	time.Sleep(durationSecond)

	nEnd, err := ns.getStatisticsHandler()
	if err != nil {
		ns.setZeroStatsAndWait()
		return
	}
	if len(nEnd) == 0 {
		ns.setZeroStatsAndWait()
		return
	}

	ns.processStatistics(nStart, nEnd)
}

func getStatistics() ([]net.IOCountersStat, error) {
	return net.IOCounters(false)
}

func (ns *netStatistics) processStatistics(nStart []net.IOCountersStat, nEnd []net.IOCountersStat) {
	isLessRecv := nEnd[0].BytesRecv < nStart[0].BytesRecv
	isLessSent := nEnd[0].BytesSent < nStart[0].BytesSent
	if isLessRecv || isLessSent {
		ns.setZeroStatsAndWait()
		return
	}

	bpsRecv := nEnd[0].BytesRecv - nStart[0].BytesRecv
	bpsSent := nEnd[0].BytesSent - nStart[0].BytesSent

	atomic.AddUint64(&ns.totalBytesSentInEpoch, bpsSent)
	atomic.AddUint64(&ns.totalBytesReceivedInEpoch, bpsRecv)

	atomic.StoreUint64(&ns.bpsRecv, bpsRecv)
	atomic.StoreUint64(&ns.bpsSent, bpsSent)

	recvPeak := atomic.LoadUint64(&ns.bpsRecvPeak)
	if recvPeak < bpsRecv {
		atomic.StoreUint64(&ns.bpsRecvPeak, bpsRecv)
		recvPeak = bpsRecv
	}

	sentPeak := atomic.LoadUint64(&ns.bpsSentPeak)
	if sentPeak < bpsSent {
		atomic.StoreUint64(&ns.bpsSentPeak, bpsSent)
		sentPeak = bpsSent
	}

	recvPercent := uint64(0)
	if recvPeak != 0 {
		recvPercent = bpsRecv * 100 / recvPeak
	}
	atomic.StoreUint64(&ns.percentRecv, recvPercent)

	sentPercent := uint64(0)
	if sentPeak != 0 {
		sentPercent = bpsSent * 100 / sentPeak
	}
	atomic.StoreUint64(&ns.percentSent, sentPercent)

	time.Sleep(durationSecond)
}

// EpochStartEventHandler -
func (ns *netStatistics) EpochStartEventHandler() epochStart.ActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {
		atomic.StoreUint64(&ns.totalBytesSentInEpoch, 0)
		atomic.StoreUint64(&ns.totalBytesReceivedInEpoch, 0)
	}, func(_ data.HeaderHandler) {}, common.NetStatisticsOrder)

	return subscribeHandler
}

func (ns *netStatistics) setZeroStatsAndWait() {
	atomic.StoreUint64(&ns.bpsSent, 0)
	atomic.StoreUint64(&ns.bpsRecv, 0)
	atomic.StoreUint64(&ns.percentSent, 0)
	atomic.StoreUint64(&ns.percentRecv, 0)
	time.Sleep(durationSecond)
}

// BpsSent bytes sent per second on all interfaces
func (ns *netStatistics) BpsSent() uint64 {
	return atomic.LoadUint64(&ns.bpsSent)
}

// BpsRecv bytes received per second on all interfaces
func (ns *netStatistics) BpsRecv() uint64 {
	return atomic.LoadUint64(&ns.bpsRecv)
}

// BpsSentPeak peak bytes sent per second on all interfaces
func (ns *netStatistics) BpsSentPeak() uint64 {
	return atomic.LoadUint64(&ns.bpsSentPeak)
}

// BpsRecvPeak peak bytes received per second on all interfaces
func (ns *netStatistics) BpsRecvPeak() uint64 {
	return atomic.LoadUint64(&ns.bpsRecvPeak)
}

// PercentSent BpsSent / BpsSentPeak * 100
func (ns *netStatistics) PercentSent() uint64 {
	return atomic.LoadUint64(&ns.percentSent)
}

// PercentRecv BpsRecv / BpsRecvPeak * 100
func (ns *netStatistics) PercentRecv() uint64 {
	return atomic.LoadUint64(&ns.percentRecv)
}

// TotalBytesSentInCurrentEpoch returns the number of bytes sent in current epoch
func (ns *netStatistics) TotalBytesSentInCurrentEpoch() uint64 {
	return atomic.LoadUint64(&ns.totalBytesSentInEpoch)
}

// TotalBytesReceivedInCurrentEpoch returns the number of bytes received in current epoch
func (ns *netStatistics) TotalBytesReceivedInCurrentEpoch() uint64 {
	return atomic.LoadUint64(&ns.totalBytesReceivedInEpoch)
}
