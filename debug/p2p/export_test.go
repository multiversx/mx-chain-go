package p2p

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core"
)

func newTestP2PDebugger(
	selfPeerId core.PeerID,
	shouldProcessDataFn func() bool,
	printStringFn func(string),
) *p2pDebugger {
	pd := &p2pDebugger{
		selfPeerId: selfPeerId,
		data:       make(map[string]*metric),
	}
	pd.shouldProcessDataFn = shouldProcessDataFn
	pd.printStringFn = printStringFn

	ctx, cancelFunc := context.WithCancel(context.Background())
	pd.cancelFunc = cancelFunc

	go pd.continuouslyPrintStatistics(ctx)

	return pd
}

func (pd *p2pDebugger) GetClonedMetric(topic string) *metric {
	pd.mut.Lock()
	defer pd.mut.Unlock()

	m := pd.data[topic]
	if m == nil {
		return nil
	}

	clonedMetric := *m

	return &clonedMetric
}
