package staking

import (
	"testing"
)

func TestNewTestMetaProcessor(t *testing.T) {
	node := NewTestMetaProcessor(3, 3, 3, 2, 2, 10, t)
	node.DisplayNodesConfig(0, 4)

	//logger.SetLogLevel("*:DEBUG,process:TRACE")
	//logger.SetLogLevel("*:DEBUG")
	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	node.Process(t, 1, 56)
}
