package staking

import (
	"testing"
)

func TestNewTestMetaProcessor(t *testing.T) {
	node := NewTestMetaProcessor(3, 3, 3, 2, 2)
	node.DisplayNodesConfig(0, 4)

	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	node.Process(t, 1, 7)
}
