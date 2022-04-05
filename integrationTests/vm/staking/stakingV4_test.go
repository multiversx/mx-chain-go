package staking

import (
	"testing"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

func TestNewTestMetaProcessor(t *testing.T) {
	node := NewTestMetaProcessor(3, 3, 3, 2, 2)
	node.DisplayNodesConfig(0, 4)

	//logger.SetLogLevel("*:DEBUG,process:TRACE")
	logger.SetLogLevel("*:DEBUG")
	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	node.Process(t, 1, 27)
}
