package machine

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestNetStatistics(t *testing.T) {
	t.Parallel()

	net := &NetStatistics{}

	net.ComputeStatistics()
	bpsRecv := net.BpsRecv()
	bpsSent := net.BpsSent()
	bpsRecvPeak := net.BpsRecvPeak()
	bpsSentPeak := net.BpsSentPeak()
	bpsRecvPercent := net.PercentSent()
	bpsSentPercent := net.PercentRecv()

	fmt.Printf("Recv: %s/s, sent: %s/s\n", core.ConvertBytes(bpsRecv), core.ConvertBytes(bpsSent))
	fmt.Printf("Peak recv: %s/s, sent: %s/s\n", core.ConvertBytes(bpsRecvPeak), core.ConvertBytes(bpsSentPeak))
	fmt.Printf("Recv usage: %d%%, sent: %d%%\n", bpsRecvPercent, bpsSentPercent)
}

func TestNetStatistics_ResetShouldWork(t *testing.T) {
	t.Parallel()

	net := &NetStatistics{}

	net.setZeroStatsAndWait()

	bpsRecv := net.BpsRecv()
	bpsSent := net.BpsSent()
	bpsRecvPeak := net.BpsRecvPeak()
	bpsSentPeak := net.BpsSentPeak()
	bpsRecvPercent := net.PercentSent()
	bpsSentPercent := net.PercentRecv()

	assert.Equal(t, uint64(0), bpsRecv)
	assert.Equal(t, uint64(0), bpsSent)
	assert.Equal(t, uint64(0), bpsRecvPeak)
	assert.Equal(t, uint64(0), bpsSentPeak)
	assert.Equal(t, uint64(0), bpsRecvPercent)
	assert.Equal(t, uint64(0), bpsSentPercent)
}
