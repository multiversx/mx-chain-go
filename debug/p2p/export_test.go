package p2p

import (
	"context"

	"github.com/ElrondNetwork/elrond-go/core"
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
