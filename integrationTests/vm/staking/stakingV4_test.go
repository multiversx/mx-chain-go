package staking

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/stretchr/testify/require"
)

func requireSliceContains(t *testing.T, s1, s2 [][]byte) {
	for _, elemInS2 := range s2 {
		require.Contains(t, s1, elemInS2)
	}
}

func requireSliceContainsNumOfElements(t *testing.T, s1, s2 [][]byte, numOfElements int) {
	foundCt := 0
	for _, elemInS2 := range s2 {
		if searchInSlice(s1, elemInS2) {
			foundCt++
		}
	}

	require.Equal(t, numOfElements, foundCt)
}

func requireSameSliceDifferentOrder(t *testing.T, s1, s2 [][]byte) {
	require.Equal(t, len(s1), len(s2))

	for _, elemInS1 := range s1 {
		require.Contains(t, s2, elemInS1)
	}
}

func searchInSlice(s1 [][]byte, s2 []byte) bool {
	for _, elemInS1 := range s1 {
		if bytes.Equal(elemInS1, s2) {
			return true
		}
	}

	return false
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
		require.True(t, searchInMap(m, elemInSlice))
	}
}

func requireMapDoesNotContain(t *testing.T, m map[uint32][][]byte, s [][]byte) {
	for _, elemInSlice := range s {
		require.False(t, searchInMap(m, elemInSlice))
	}
}

// TODO: Staking v4: more tests to check exactly which nodes have been selected/unselected from previous nodes config auction

func TestStakingV4(t *testing.T) {
	numOfMetaNodes := uint32(400)
	numOfShards := uint32(3)
	numOfEligibleNodesPerShard := uint32(400)
	numOfWaitingNodesPerShard := uint32(400)
	numOfNodesToShufflePerShard := uint32(80)
	shardConsensusGroupSize := 266
	metaConsensusGroupSize := 266
	numOfNodesInStakingQueue := uint32(60)

	totalEligible := int(numOfEligibleNodesPerShard*numOfShards) + int(numOfMetaNodes) // 1600
	totalWaiting := int(numOfWaitingNodesPerShard*numOfShards) + int(numOfMetaNodes)   // 1600

	node := NewTestMetaProcessor(
		numOfMetaNodes,
		numOfShards,
		numOfEligibleNodesPerShard,
		numOfWaitingNodesPerShard,
		numOfNodesToShufflePerShard,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
		numOfNodesInStakingQueue,
	)
	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	// 1. Check initial config is correct
	initialNodes := node.NodesConfig
	require.Len(t, getAllPubKeys(initialNodes.eligible), totalEligible)
	require.Len(t, getAllPubKeys(initialNodes.waiting), totalWaiting)
	require.Len(t, initialNodes.queue, int(numOfNodesInStakingQueue))
	require.Empty(t, initialNodes.shuffledOut)
	require.Empty(t, initialNodes.auction)

	// 2. Check config after staking v4 initialization
	node.Process(t, 5)
	nodesConfigStakingV4Init := node.NodesConfig
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Init.eligible), totalEligible)
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Init.waiting), totalWaiting)
	require.Empty(t, nodesConfigStakingV4Init.queue)
	require.Empty(t, nodesConfigStakingV4Init.shuffledOut)
	requireSameSliceDifferentOrder(t, initialNodes.queue, nodesConfigStakingV4Init.auction)

	// 3. Check config after first staking v4 epoch, WITHOUT distribution from auction -> waiting
	node.Process(t, 6)
	nodesConfigStakingV4 := node.NodesConfig
	require.Len(t, getAllPubKeys(nodesConfigStakingV4.eligible), totalEligible) // 1600

	numOfShuffledOut := int((numOfShards + 1) * numOfNodesToShufflePerShard) // 320
	require.Len(t, getAllPubKeys(nodesConfigStakingV4.shuffledOut), numOfShuffledOut)

	newWaiting := totalWaiting - numOfShuffledOut // 1280 (1600 - 320)
	require.Len(t, getAllPubKeys(nodesConfigStakingV4.waiting), newWaiting)

	// 380 (320 from shuffled out + 60 from initial staking queue -> auction from stakingV4 init)
	auctionListSize := numOfShuffledOut + len(nodesConfigStakingV4Init.auction)
	require.Len(t, nodesConfigStakingV4.auction, auctionListSize)
	requireSliceContains(t, nodesConfigStakingV4.auction, nodesConfigStakingV4Init.auction)

	require.Empty(t, nodesConfigStakingV4.queue)
	require.Empty(t, nodesConfigStakingV4.leaving)

	// 320 nodes which are now in eligible are from previous waiting list
	requireSliceContainsNumOfElements(t, getAllPubKeys(nodesConfigStakingV4.eligible), getAllPubKeys(nodesConfigStakingV4Init.waiting), numOfShuffledOut)

	// All shuffled out are from previous staking v4 init eligible
	requireMapContains(t, nodesConfigStakingV4Init.eligible, getAllPubKeys(nodesConfigStakingV4.shuffledOut))

	// All shuffled out are in auction
	requireSliceContains(t, nodesConfigStakingV4.auction, getAllPubKeys(nodesConfigStakingV4.shuffledOut))

	// No auction node from previous epoch has been moved to waiting
	requireMapDoesNotContain(t, nodesConfigStakingV4.waiting, nodesConfigStakingV4Init.auction)

	epochs := 0
	prevConfig := nodesConfigStakingV4
	numOfSelectedNodesFromAuction := numOfShuffledOut                     // 320, since we will always fill shuffled out nodes with this config
	numOfUnselectedNodesFromAuction := auctionListSize - numOfShuffledOut // 60 = 380 - 320
	for epochs < 10 {
		node.Process(t, 5)
		newNodeConfig := node.NodesConfig

		require.Len(t, getAllPubKeys(newNodeConfig.eligible), totalEligible)       // 1600
		require.Len(t, getAllPubKeys(newNodeConfig.waiting), newWaiting)           // 1280
		require.Len(t, getAllPubKeys(newNodeConfig.shuffledOut), numOfShuffledOut) // 320
		require.Len(t, newNodeConfig.auction, auctionListSize)                     // 380
		require.Empty(t, newNodeConfig.queue)
		require.Empty(t, newNodeConfig.leaving)

		// 320 nodes which are now in eligible are from previous waiting list
		requireSliceContainsNumOfElements(t, getAllPubKeys(newNodeConfig.eligible), getAllPubKeys(prevConfig.waiting), numOfShuffledOut)

		// New auction list also contains unselected nodes from previous auction list
		requireSliceContainsNumOfElements(t, newNodeConfig.auction, prevConfig.auction, numOfUnselectedNodesFromAuction)

		// All shuffled out are from previous eligible config
		requireMapContains(t, prevConfig.eligible, getAllPubKeys(newNodeConfig.shuffledOut))

		// All shuffled out are now in auction
		requireSliceContains(t, newNodeConfig.auction, getAllPubKeys(newNodeConfig.shuffledOut))

		// 320 nodes which have been selected from previous auction list are now in waiting
		requireSliceContainsNumOfElements(t, getAllPubKeys(newNodeConfig.waiting), prevConfig.auction, numOfSelectedNodesFromAuction)

		prevConfig = newNodeConfig
		epochs++
	}
}

func TestStakingV4MetaProcessor_ProcessMultipleNodesWithSameSetupExpectSameRootHash(t *testing.T) {
	numOfMetaNodes := uint32(6)
	numOfShards := uint32(3)
	numOfEligibleNodesPerShard := uint32(6)
	numOfWaitingNodesPerShard := uint32(6)
	numOfNodesToShufflePerShard := uint32(2)
	shardConsensusGroupSize := 2
	metaConsensusGroupSize := 2
	numOfNodesInStakingQueue := uint32(2)

	nodes := make([]*TestMetaProcessor, 0, numOfMetaNodes)
	for i := uint32(0); i < numOfMetaNodes; i++ {
		nodes = append(nodes, NewTestMetaProcessor(
			numOfMetaNodes,
			numOfShards,
			numOfEligibleNodesPerShard,
			numOfWaitingNodesPerShard,
			numOfNodesToShufflePerShard,
			shardConsensusGroupSize,
			metaConsensusGroupSize,
			numOfNodesInStakingQueue,
		))
		nodes[i].EpochStartTrigger.SetRoundsPerEpoch(4)
	}

	numOfEpochs := uint32(15)
	rootHashes := make(map[uint32][][]byte)
	for currEpoch := uint32(1); currEpoch <= numOfEpochs; currEpoch++ {
		for _, node := range nodes {
			rootHash, _ := node.ValidatorStatistics.RootHash()
			rootHashes[currEpoch] = append(rootHashes[currEpoch], rootHash)

			node.Process(t, 5)
			require.Equal(t, currEpoch, node.EpochStartTrigger.Epoch())
		}
	}

	for _, rootHashesInEpoch := range rootHashes {
		firstNodeRootHashInEpoch := rootHashesInEpoch[0]
		for _, rootHash := range rootHashesInEpoch {
			require.Equal(t, firstNodeRootHashInEpoch, rootHash)
		}
	}
}

func TestStakingV4_CustomScenario(t *testing.T) {
	pubKeys := generateAddresses(0, 20)

	//_ = logger.SetLogLevel("*:DEBUG")

	owner1 := "owner1"
	owner1Stats := &OwnerStats{
		EligibleBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[:3],
			0:                     pubKeys[3:6],
		},
		StakingQueueKeys: pubKeys[6:9],
		TotalStake:       big.NewInt(5000),
	}

	owner2 := "owner2"
	owner2StakingQueueKeys := [][]byte{pubKeys[12], pubKeys[13], pubKeys[14]}
	owner2Stats := &OwnerStats{
		StakingQueueKeys: owner2StakingQueueKeys,
		TotalStake:       big.NewInt(5000),
	}

	cfg := &InitialNodesConfig{
		MetaConsensusGroupSize:        2,
		ShardConsensusGroupSize:       2,
		MinNumberOfEligibleShardNodes: 2,
		MinNumberOfEligibleMetaNodes:  2,
		NumOfShards:                   2,
		Owners: map[string]*OwnerStats{
			owner1: owner1Stats,
			owner2: owner2Stats,
		},
		MaxNodesChangeConfig: []config.MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            4,
				NodesToShufflePerShard: 2,
			},
		},
	}
	//todo; check that in epoch = staking v4 nodes with not enough stake will be unstaked
	node := NewTestMetaProcessorWithCustomNodes(cfg)
	node.EpochStartTrigger.SetRoundsPerEpoch(5)

	node.Process(t, 16)

}
