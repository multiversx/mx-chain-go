package broadcast_test

import (
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// the block data propagation delays the broadcast tests are written against
const (
	testExtraDelayForBroadcastBlockInfo     = 30 * time.Millisecond
	testExtraDelayBetweenBroadcastMbsAndTxs = 50 * time.Millisecond
)

func createTestProcessConfigsHandler() common.ProcessConfigsHandler {
	return &testscommon.ProcessConfigsHandlerStub{
		GetExtraDelayForBroadcastBlockInfoCalled: func(_ uint64) time.Duration {
			return testExtraDelayForBroadcastBlockInfo
		},
		GetExtraDelayBetweenBroadcastMbsAndTxsCalled: func(_ uint64) time.Duration {
			return testExtraDelayBetweenBroadcastMbsAndTxs
		},
	}
}
