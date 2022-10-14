package machine

import (
	"context"

	"github.com/shirou/gopsutil/net"
)

// NewNetTestStatistics -
func NewNetTestStatistics(getTestStatistics func() ([]net.IOCountersStat, error)) *netStatistics {
	stats := newNetStatistics(getTestStatistics)

	var ctx context.Context
	ctx, stats.cancel = context.WithCancel(context.Background())

	go stats.processLoop(ctx)

	return stats
}
