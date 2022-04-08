package staking

import (
	"testing"
)

func TestNewTestMetaProcessor(t *testing.T) {
	node := NewTestMetaProcessor(3, 3, 3, 3, 2, 2, 2, 10, t)

	//logger.SetLogLevel("*:DEBUG,process:TRACE")
	//logger.SetLogLevel("*:DEBUG")
	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	node.Process(t, 1, 56)
}
