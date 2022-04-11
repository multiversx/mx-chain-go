package staking

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func requireSameSliceDifferentOrder(t *testing.T, s1, s2 [][]byte) {
	require.Equal(t, len(s1), len(s2))

	for _, elemInS1 := range s1 {
		require.Contains(t, s2, elemInS1)
	}
}

func searchInMap(validatorMap map[uint32][][]byte, pk []byte) bool {
	for _, validatorsInShard := range validatorMap {
		for _, val := range validatorsInShard {
			if bytes.Equal(val, pk) {
				return true
			}
		}
	}
	return false
}

func requireMapContains(t *testing.T, m map[uint32][][]byte, s [][]byte) {
	for _, elemInSlice := range s {
		found := searchInMap(m, elemInSlice)
		require.True(t, found)
	}
}

func getAllPubKeys(validatorsMap map[uint32][][]byte) [][]byte {
	allValidators := make([][]byte, 0)
	for _, validatorsInShard := range validatorsMap {
		allValidators = append(allValidators, validatorsInShard...)
	}

	return allValidators
}

func TestNewTestMetaProcessor(t *testing.T) {
	node := NewTestMetaProcessor(3, 3, 3, 3, 2, 2, 2, 2)
	initialNodes := node.NodesConfig
	//logger.SetLogLevel("*:DEBUG,process:TRACE")
	//logger.SetLogLevel("*:DEBUG")
	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	node.Process(t, 5)

	eligibleAfterStakingV4Init := node.NodesConfig.eligible
	require.Empty(t, node.NodesConfig.queue)
	requireSameSliceDifferentOrder(t, initialNodes.queue, node.NodesConfig.auction)

	node.Process(t, 6)
	requireMapContains(t, node.NodesConfig.shuffledOut, node.NodesConfig.auction)
	requireMapContains(t, node.NodesConfig.waiting, initialNodes.queue)
	requireMapContains(t, eligibleAfterStakingV4Init, node.NodesConfig.auction) //todo: check size
}
