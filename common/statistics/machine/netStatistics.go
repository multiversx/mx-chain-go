package machine

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
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
	cancel                    func()
}

// NewNetStatistics returns a new instance of the netStatistics
func NewNetStatistics() *netStatistics {
	stats := newNetStatistics(getStatistics)

	var ctx context.Context
	ctx, stats.cancel = context.WithCancel(context.Background())

	go stats.processLoop(ctx)

	return stats
}

func newNetStatistics(getStatisticsHandler func() ([]net.IOCountersStat, error)) *netStatistics {
	stats := &netStatistics{
		getStatisticsHandler: getStatisticsHandler,
	}

	return stats
}

func (ns *netStatistics) processLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("netStatistics.processLoop go routine is stopping...")
			return
		default:
		}

		ns.computeStatistics(ctx)
	}
}

// computeStatistics computes the current network statistics usage.
func (ns *netStatistics) computeStatistics(ctx context.Context) {
	nStart, err := ns.getStatisticsHandler()
	if err != nil {
		ns.setZeroStatsAndWait()
		return
	}
	if len(nStart) == 0 {
		ns.setZeroStatsAndWait()
		return
	}

	timer := time.NewTimer(durationSecond)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-ctx.Done():
		return
	}

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

	timer.Reset(durationSecond)
	select {
	case <-timer.C:
	case <-ctx.Done():
		return
	}
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
}

// EpochConfirmed is called whenever a new epoch is starting
func (ns *netStatistics) EpochConfirmed(_ uint32, _ uint64) {
	atomic.StoreUint64(&ns.totalBytesSentInEpoch, 0)
	atomic.StoreUint64(&ns.totalBytesReceivedInEpoch, 0)
	atomic.StoreUint64(&ns.bpsSentPeak, 0)
	atomic.StoreUint64(&ns.bpsRecvPeak, 0)
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

// TotalSentInCurrentEpoch will convert the TotalBytesSentInCurrentEpoch using byte multipliers
func (ns *netStatistics) TotalSentInCurrentEpoch() string {
	return core.ConvertBytes(ns.TotalBytesSentInCurrentEpoch())
}

// TotalReceivedInCurrentEpoch will convert the TotalBytesReceivedInCurrentEpoch using byte multipliers
func (ns *netStatistics) TotalReceivedInCurrentEpoch() string {
	return core.ConvertBytes(ns.TotalBytesReceivedInCurrentEpoch())
}

// Close stops the associated go routine to this struct
func (ns *netStatistics) Close() error {
	ns.cancel()

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (ns *netStatistics) IsInterfaceNil() bool {
	return ns == nil
}
