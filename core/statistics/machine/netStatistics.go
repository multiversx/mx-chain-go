package machine

import (
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/net"
)

// NetStatistics can compute the network statistics
type NetStatistics struct {
	bpsSent     uint64
	bpsRecv     uint64
	bpsSentPeak uint64
	bpsRecvPeak uint64
	percentSent uint64
	percentRecv uint64
}

// ComputeStatistics computes the current network statistics usage.
// It should be called on a go routine as it is a blocking call for a bounded time (1 second)
func (ns *NetStatistics) ComputeStatistics() {
	nStart, err := net.IOCounters(false)
	if err != nil {
		ns.setZeroStatsAndWait()
		return
	}
	if len(nStart) == 0 {
		ns.setZeroStatsAndWait()
		return
	}

	time.Sleep(durationSecond)

	nEnd, err := net.IOCounters(false)
	if err != nil {
		ns.setZeroStatsAndWait()
		return
	}
	if len(nEnd) == 0 {
		ns.setZeroStatsAndWait()
		return
	}

	isLessRecv := nEnd[0].BytesRecv < nStart[0].BytesRecv
	isLessSent := nEnd[0].BytesSent < nStart[0].BytesSent
	if isLessRecv || isLessSent {
		ns.setZeroStatsAndWait()
		return
	}

	bpsRecv := nEnd[0].BytesRecv - nStart[0].BytesRecv
	bpsSent := nEnd[0].BytesSent - nStart[0].BytesSent

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

func (ns *NetStatistics) setZeroStatsAndWait() {
	atomic.StoreUint64(&ns.bpsSent, 0)
	atomic.StoreUint64(&ns.bpsRecv, 0)
	atomic.StoreUint64(&ns.percentSent, 0)
	atomic.StoreUint64(&ns.percentRecv, 0)
	time.Sleep(durationSecond)
}

// BpsSent bytes sent per second on all interfaces
func (ns *NetStatistics) BpsSent() uint64 {
	return atomic.LoadUint64(&ns.bpsSent)
}

// BpsRecv bytes received per second on all interfaces
func (ns *NetStatistics) BpsRecv() uint64 {
	return atomic.LoadUint64(&ns.bpsRecv)
}

// BpsSentPeak peak bytes sent per second on all interfaces
func (ns *NetStatistics) BpsSentPeak() uint64 {
	return atomic.LoadUint64(&ns.bpsSentPeak)
}

// BpsRecvPeak peak bytes received per second on all interfaces
func (ns *NetStatistics) BpsRecvPeak() uint64 {
	return atomic.LoadUint64(&ns.bpsRecvPeak)
}

// PercentSent BpsSent / BpsSentPeak * 100
func (ns *NetStatistics) PercentSent() uint64 {
	return atomic.LoadUint64(&ns.percentSent)
}

// PercentRecv BpsRecv / BpsRecvPeak * 100
func (ns *NetStatistics) PercentRecv() uint64 {
	return atomic.LoadUint64(&ns.percentRecv)
}
