package machine

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/shirou/gopsutil/net"
	"github.com/stretchr/testify/assert"
)

func TestNewNetStatistics_ShouldWorkAndNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not paniced %v", r))
		}
	}()

	ns := NewNetStatistics()
	assert.NotNil(t, ns)

	ns.ComputeStatistics()
}

func TestNetStatistics_ComputeStatisticsGetStatisticsErrorsFirstTime(t *testing.T) {
	t.Parallel()

	numCalls := 0
	expectedErr := errors.New("expected error")
	ns := &netStatistics{
		getStatisticsHandler: func() ([]net.IOCountersStat, error) {
			numCalls++
			if numCalls == 1 {
				return nil, expectedErr
			}

			return []net.IOCountersStat{
				{
					BytesSent: 1,
					BytesRecv: 2,
				},
			}, nil
		},
	}
	populateFields(ns)

	ns.ComputeStatistics()
	checkResetFieldsAreZero(t, ns)
}

func TestNetStatistics_ComputeStatisticsGetStatisticsReturnsEmptyFirstTime(t *testing.T) {
	t.Parallel()

	numCalls := 0
	ns := &netStatistics{
		getStatisticsHandler: func() ([]net.IOCountersStat, error) {
			numCalls++
			if numCalls == 1 {
				return make([]net.IOCountersStat, 0), nil
			}

			return []net.IOCountersStat{
				{
					BytesSent: 1,
					BytesRecv: 2,
				},
			}, nil
		},
	}
	populateFields(ns)

	ns.ComputeStatistics()
	checkResetFieldsAreZero(t, ns)
}

func TestNetStatistics_ComputeStatisticsGetStatisticsErrorsSecondTime(t *testing.T) {
	t.Parallel()

	numCalls := 0
	expectedErr := errors.New("expected error")
	ns := &netStatistics{
		getStatisticsHandler: func() ([]net.IOCountersStat, error) {
			numCalls++
			if numCalls == 2 {
				return nil, expectedErr
			}

			return []net.IOCountersStat{
				{
					BytesSent: 1,
					BytesRecv: 2,
				},
			}, nil
		},
	}
	populateFields(ns)

	ns.ComputeStatistics()
	checkResetFieldsAreZero(t, ns)
}

func TestNetStatistics_ComputeStatisticsGetStatisticsReturnsEmptySecondTime(t *testing.T) {
	t.Parallel()

	numCalls := 0
	ns := &netStatistics{
		getStatisticsHandler: func() ([]net.IOCountersStat, error) {
			numCalls++
			if numCalls == 2 {
				return make([]net.IOCountersStat, 0), nil
			}

			return []net.IOCountersStat{
				{
					BytesSent: 1,
					BytesRecv: 2,
				},
			}, nil
		},
	}
	populateFields(ns)

	ns.ComputeStatistics()
	checkResetFieldsAreZero(t, ns)
}

func populateFields(ns *netStatistics) {
	ns.bpsSent = 1
	ns.bpsRecv = 2
	ns.bpsSentPeak = 3
	ns.bpsRecvPeak = 4
	ns.percentSent = 5
	ns.percentRecv = 6
	ns.totalBytesSentInEpoch = 7
	ns.totalBytesReceivedInEpoch = 8
}

func checkResetFieldsAreZero(tb testing.TB, ns *netStatistics) {
	assert.Equal(tb, uint64(0), ns.BpsSent())
	assert.Equal(tb, uint64(0), ns.BpsRecv())
	assert.NotEqual(tb, uint64(0), ns.BpsSentPeak())
	assert.NotEqual(tb, uint64(0), ns.BpsRecvPeak())
	assert.Equal(tb, uint64(0), ns.PercentSent())
	assert.Equal(tb, uint64(0), ns.PercentRecv())
	assert.NotEqual(tb, uint64(0), ns.TotalBytesSentInCurrentEpoch())
	assert.NotEqual(tb, uint64(0), ns.TotalBytesReceivedInCurrentEpoch())
}

func TestNetStatistics_ResetShouldWork(t *testing.T) {
	t.Parallel()

	ns := NewNetStatistics()

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

	fmt.Printf("Recv: %s/s, sent: %s/s\n", core.ConvertBytes(bpsRecv), core.ConvertBytes(bpsSent))
	fmt.Printf("Peak recv: %s/s, sent: %s/s\n", core.ConvertBytes(bpsRecvPeak), core.ConvertBytes(bpsSentPeak))
	fmt.Printf("Recv usage: %d%%, sent: %d%%\n", bpsRecvPercent, bpsSentPercent)
}

func TestNetStatistics_processStatistics(t *testing.T) {
	ns := NewNetStatistics()

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
	ns := NewNetStatistics()

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

func TestNetStatistics_EpochStartEventHandler(t *testing.T) {
	ns := NewNetStatistics()
	populateFields(ns)

	handler := ns.EpochStartEventHandler()
	handler.EpochStartAction(nil)

	assert.Equal(t, uint64(0), ns.TotalBytesSentInCurrentEpoch())
	assert.Equal(t, uint64(0), ns.TotalBytesReceivedInCurrentEpoch())
}
