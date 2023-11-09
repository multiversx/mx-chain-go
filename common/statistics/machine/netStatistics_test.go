package machine

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
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
	assert.False(t, check.IfNil(ns))

	ns.computeStatistics(context.Background())
}

func TestNetStatistics_ComputeStatisticsGetStatisticsErrorsFirstTime(t *testing.T) {
	t.Parallel()

	numCalls := 0
	expectedErr := errors.New("expected error")
	testGetStats := func() ([]net.IOCountersStat, error) {
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
	}

	ns := newNetStatistics(testGetStats)

	populateFields(ns)

	ns.computeStatistics(context.Background())

	checkResetFieldsAreZero(t, ns)
}

func TestNetStatistics_ComputeStatisticsGetStatisticsReturnsEmptyFirstTime(t *testing.T) {
	t.Parallel()

	numCalls := 0
	testGetStats := func() ([]net.IOCountersStat, error) {
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
	}

	ns := newNetStatistics(testGetStats)
	populateFields(ns)

	ns.computeStatistics(context.Background())
	checkResetFieldsAreZero(t, ns)
}

func TestNetStatistics_ComputeStatisticsGetStatisticsErrorsSecondTime(t *testing.T) {
	t.Parallel()

	numCalls := 0
	expectedErr := errors.New("expected error")
	testGetStats := func() ([]net.IOCountersStat, error) {
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
	}

	ns := newNetStatistics(testGetStats)
	populateFields(ns)

	ns.computeStatistics(context.Background())
	checkResetFieldsAreZero(t, ns)
}

func TestNetStatistics_ComputeStatisticsGetStatisticsReturnsEmptySecondTime(t *testing.T) {
	t.Parallel()

	numCalls := 0
	testGetStats := func() ([]net.IOCountersStat, error) {
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
	}

	ns := newNetStatistics(testGetStats)
	populateFields(ns)

	ns.computeStatistics(context.Background())
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

	numBytes := uint64(0)
	testGetStats := func() ([]net.IOCountersStat, error) {
		atomic.AddUint64(&numBytes, 1)

		return []net.IOCountersStat{
			{
				BytesSent: numBytes,
				BytesRecv: numBytes * 2,
			},
		}, nil
	}

	ns := newNetStatistics(testGetStats)
	ns.computeStatistics(context.Background())

	ns.setZeroStatsAndWait()

	bpsRecv := ns.BpsRecv()
	bpsSent := ns.BpsSent()
	bpsRecvPeak := ns.BpsRecvPeak()
	bpsSentPeak := ns.BpsSentPeak()
	bpsRecvPercent := ns.PercentSent()
	bpsSentPercent := ns.PercentRecv()

	assert.Equal(t, uint64(0), bpsRecv)
	assert.Equal(t, uint64(0), bpsSent)
	assert.Equal(t, uint64(2), bpsRecvPeak) // these do not reset
	assert.Equal(t, uint64(1), bpsSentPeak) // these do not reset
	assert.Equal(t, uint64(0), bpsRecvPercent)
	assert.Equal(t, uint64(0), bpsSentPercent)

	fmt.Printf("Recv: %s/s, sent: %s/s\n", core.ConvertBytes(bpsRecv), core.ConvertBytes(bpsSent))
	fmt.Printf("Peak recv: %s/s, sent: %s/s\n", core.ConvertBytes(bpsRecvPeak), core.ConvertBytes(bpsSentPeak))
	fmt.Printf("Recv usage: %d%%, sent: %d%%\n", bpsRecvPercent, bpsSentPercent)
}

func TestNetStatistics_processStatistics(t *testing.T) {
	t.Parallel()

	testGetStats := func() ([]net.IOCountersStat, error) {
		return make([]net.IOCountersStat, 0), nil
	}
	ns := newNetStatistics(testGetStats)

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
	t.Parallel()

	testGetStats := func() ([]net.IOCountersStat, error) {
		return make([]net.IOCountersStat, 0), nil
	}
	ns := newNetStatistics(testGetStats)

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

func TestNetStatistics_EpochConfirmed(t *testing.T) {
	t.Parallel()

	testGetStats := func() ([]net.IOCountersStat, error) {
		return make([]net.IOCountersStat, 0), nil
	}
	ns := newNetStatistics(testGetStats)

	populateFields(ns)

	ns.EpochConfirmed(0, 0)

	assert.Equal(t, uint64(0), ns.TotalBytesSentInCurrentEpoch())
	assert.Equal(t, "0 B", ns.TotalSentInCurrentEpoch())
	assert.Equal(t, uint64(0), ns.TotalBytesReceivedInCurrentEpoch())
	assert.Equal(t, "0 B", ns.TotalReceivedInCurrentEpoch())
	assert.Equal(t, uint64(0), ns.BpsRecvPeak())
	assert.Equal(t, uint64(0), ns.BpsSentPeak())
}

func TestNetStatistics_Close(t *testing.T) {
	t.Parallel()

	numCalls := uint64(0)
	testGetStats := func() ([]net.IOCountersStat, error) {
		atomic.AddUint64(&numCalls, 1)
		return make([]net.IOCountersStat, 0), nil
	}
	ns := NewNetTestStatistics(testGetStats)

	time.Sleep(time.Second*3 + time.Millisecond*500)
	assert.True(t, atomic.LoadUint64(&numCalls) > 0) // go routine is running

	err := ns.Close()
	assert.Nil(t, err)
	time.Sleep(time.Second) // wait for the go routine finish
	atomic.StoreUint64(&numCalls, 0)

	time.Sleep(time.Second * 3)
	assert.Zero(t, atomic.LoadUint64(&numCalls))
}
