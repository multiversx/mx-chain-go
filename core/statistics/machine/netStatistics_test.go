package machine

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/shirou/gopsutil/net"
	"github.com/stretchr/testify/assert"
)

func TestNetStatistics(t *testing.T) {
	t.Parallel()

	ns := &NetStatistics{}

	ns.ComputeStatistics()
	bpsRecv := ns.BpsRecv()
	bpsSent := ns.BpsSent()
	bpsRecvPeak := ns.BpsRecvPeak()
	bpsSentPeak := ns.BpsSentPeak()
	bpsRecvPercent := ns.PercentSent()
	bpsSentPercent := ns.PercentRecv()

	fmt.Printf("Recv: %s/s, sent: %s/s\n", core.ConvertBytes(bpsRecv), core.ConvertBytes(bpsSent))
	fmt.Printf("Peak recv: %s/s, sent: %s/s\n", core.ConvertBytes(bpsRecvPeak), core.ConvertBytes(bpsSentPeak))
	fmt.Printf("Recv usage: %d%%, sent: %d%%\n", bpsRecvPercent, bpsSentPercent)
}

func TestNetStatistics_ResetShouldWork(t *testing.T) {
	t.Parallel()

	ns := &NetStatistics{}

	ns.setZeroStatsAndWait()

	bpsRecv := ns.BpsRecv()
	bpsSent := ns.BpsSent()
	bpsRecvPeak := ns.BpsRecvPeak()
	bpsSentPeak := ns.BpsSentPeak()
	bpsRecvPercent := ns.PercentSent()
	bpsSentPercent := ns.PercentRecv()

	assert.Equal(t, uint64(0), bpsRecv)
	assert.Equal(t, uint64(0), bpsSent)
	assert.Equal(t, uint64(0), bpsRecvPeak)
	assert.Equal(t, uint64(0), bpsSentPeak)
	assert.Equal(t, uint64(0), bpsRecvPercent)
	assert.Equal(t, uint64(0), bpsSentPercent)
}

func TestNetStatistics_processStatistics(t *testing.T) {
	ns := &NetStatistics{}

	start := net.IOCountersStat{
		BytesRecv: 1000,
		BytesSent: 2000,
	}

	end := net.IOCountersStat{
		BytesRecv: 3000,
		BytesSent: 5000,
	}

	ns.processStatistics([]net.IOCountersStat{start}, []net.IOCountersStat{end})
	assert.Equal(t, uint64(3000), atomic.LoadUint64(&ns.totalBytesSentInEpoch))
	assert.Equal(t, uint64(2000), atomic.LoadUint64(&ns.totalBytesReceivedInEpoch))
	assert.Equal(t, uint64(2000), atomic.LoadUint64(&ns.bpsRecvPeak))
	assert.Equal(t, uint64(3000), atomic.LoadUint64(&ns.bpsSentPeak))
	assert.Equal(t, uint64(100), atomic.LoadUint64(&ns.percentRecv))
	assert.Equal(t, uint64(100), atomic.LoadUint64(&ns.percentSent))
}

func TestNetStatistics_processStatisticsInvalidEndVsStart(t *testing.T) {
	ns := &NetStatistics{}

	start := net.IOCountersStat{
		BytesRecv: 1000,
		BytesSent: 2000,
	}

	end := net.IOCountersStat{
		BytesRecv: 999,
		BytesSent: 2000,
	}

	ns.processStatistics([]net.IOCountersStat{start}, []net.IOCountersStat{end})
	assert.Equal(t, uint64(0), atomic.LoadUint64(&ns.totalBytesSentInEpoch))
	assert.Equal(t, uint64(0), atomic.LoadUint64(&ns.totalBytesReceivedInEpoch))
	assert.Equal(t, uint64(0), atomic.LoadUint64(&ns.bpsRecvPeak))
	assert.Equal(t, uint64(0), atomic.LoadUint64(&ns.bpsSentPeak))
	assert.Equal(t, uint64(0), atomic.LoadUint64(&ns.percentRecv))
	assert.Equal(t, uint64(0), atomic.LoadUint64(&ns.percentSent))
}
