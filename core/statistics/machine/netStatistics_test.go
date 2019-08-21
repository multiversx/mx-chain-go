package machine

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
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
