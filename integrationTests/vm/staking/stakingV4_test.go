package staking

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/stakingcommon"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
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

// remove will remove the item from slice without keeping the order of the original slice
func remove(slice [][]byte, elem []byte) [][]byte {
	ret := slice
	for i, e := range slice {
		if bytes.Equal(elem, e) {
			ret[i] = ret[len(slice)-1]
			return ret[:len(slice)-1]
		}
	}

	return ret
}

func getIntersection(slice1, slice2 [][]byte) [][]byte {
	ret := make([][]byte, 0)
	for _, value := range slice2 {
		if searchInSlice(slice1, value) {
			copiedVal := make([]byte, len(value))
			copy(copiedVal, value)
			ret = append(ret, copiedVal)
		}
	}

	return ret
}

func getAllPubKeysFromConfig(nodesCfg nodesConfig) [][]byte {
	allPubKeys := getAllPubKeys(nodesCfg.eligible)
	allPubKeys = append(allPubKeys, getAllPubKeys(nodesCfg.waiting)...)
	allPubKeys = append(allPubKeys, getAllPubKeys(nodesCfg.leaving)...)
	allPubKeys = append(allPubKeys, getAllPubKeys(nodesCfg.shuffledOut)...)
	allPubKeys = append(allPubKeys, nodesCfg.queue...)
	allPubKeys = append(allPubKeys, nodesCfg.auction...)
	allPubKeys = append(allPubKeys, nodesCfg.new...)

	return allPubKeys
}

func unStake(t *testing.T, owner []byte, accountsDB state.AccountsAdapter, marshaller marshal.Marshalizer, stake *big.Int) {
	validatorSC := stakingcommon.LoadUserAccount(accountsDB, vm.ValidatorSCAddress)
	ownerStoredData, _, err := validatorSC.RetrieveValue(owner)
	require.Nil(t, err)

	validatorData := &systemSmartContracts.ValidatorDataV2{}
	err = marshaller.Unmarshal(validatorData, ownerStoredData)
	require.Nil(t, err)

	validatorData.TotalStakeValue.Sub(validatorData.TotalStakeValue, stake)
	marshaledData, _ := marshaller.Marshal(validatorData)
	err = validatorSC.SaveKeyValue(owner, marshaledData)
	require.Nil(t, err)

	err = accountsDB.SaveAccount(validatorSC)
	require.Nil(t, err)
	_, err = accountsDB.Commit()
	require.Nil(t, err)
}

type configNum struct {
	eligible    map[uint32]int
	waiting     map[uint32]int
	leaving     map[uint32]int
	shuffledOut map[uint32]int
	queue       int
	auction     int
	new         int
}

func checkConfig(t *testing.T, expectedConfig *configNum, nodesConfig nodesConfig) {
	checkNumNodes(t, expectedConfig.eligible, nodesConfig.eligible)
	checkNumNodes(t, expectedConfig.waiting, nodesConfig.waiting)
	checkNumNodes(t, expectedConfig.leaving, nodesConfig.leaving)
	checkNumNodes(t, expectedConfig.shuffledOut, nodesConfig.shuffledOut)

	require.Equal(t, expectedConfig.queue, len(nodesConfig.queue))
	require.Equal(t, expectedConfig.auction, len(nodesConfig.auction))
	require.Equal(t, expectedConfig.new, len(nodesConfig.new))
}

func checkNumNodes(t *testing.T, expectedNumNodes map[uint32]int, actualNodes map[uint32][][]byte) {
	for shardID, numNodesInShard := range expectedNumNodes {
		require.Equal(t, numNodesInShard, len(actualNodes[shardID]))
	}
}

func checkShuffledOutNodes(t *testing.T, currNodesConfig, prevNodesConfig nodesConfig, numShuffledOutNodes int, numRemainingEligible int) {
	// Shuffled nodes from previous eligible are sent to waiting and previous waiting list nodes are replacing shuffled nodes
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesConfig.eligible), getAllPubKeys(prevNodesConfig.waiting), numShuffledOutNodes)
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesConfig.eligible), getAllPubKeys(prevNodesConfig.eligible), numRemainingEligible)
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesConfig.waiting), getAllPubKeys(prevNodesConfig.eligible), numShuffledOutNodes)
}

func checkStakingV4EpochChangeFlow(
	t *testing.T,
	currNodesConfig, prevNodesConfig nodesConfig,
	numOfShuffledOut, numOfUnselectedNodesFromAuction, numOfSelectedNodesFromAuction int) {

	// Nodes which are now in eligible are from previous waiting list
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesConfig.eligible), getAllPubKeys(prevNodesConfig.waiting), numOfShuffledOut)

	// New auction list also contains unselected nodes from previous auction list
	requireSliceContainsNumOfElements(t, currNodesConfig.auction, prevNodesConfig.auction, numOfUnselectedNodesFromAuction)

	// All shuffled out are from previous eligible config
	requireMapContains(t, prevNodesConfig.eligible, getAllPubKeys(currNodesConfig.shuffledOut))

	// All shuffled out are now in auction
	requireSliceContains(t, currNodesConfig.auction, getAllPubKeys(currNodesConfig.shuffledOut))

	// Nodes which have been selected from previous auction list are now in waiting
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesConfig.waiting), prevNodesConfig.auction, numOfSelectedNodesFromAuction)
}

func getAllOwnerNodesMap(nodeGroups ...[][]byte) map[string][][]byte {
	ret := make(map[string][][]byte)

	for _, nodes := range nodeGroups {
		addNodesToMap(nodes, ret)
	}

	return ret
}

func addNodesToMap(nodes [][]byte, allOwnerNodes map[string][][]byte) {
	for _, node := range nodes {
		allOwnerNodes[string(node)] = [][]byte{node}
	}
}

func TestStakingV4(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	nodesConfigStakingV4Step1 := node.NodesConfig
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Step1.eligible), totalEligible)
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Step1.waiting), totalWaiting)
	require.Empty(t, nodesConfigStakingV4Step1.queue)
	require.Empty(t, nodesConfigStakingV4Step1.shuffledOut)
	require.Empty(t, nodesConfigStakingV4Step1.auction) // the queue should be empty

	// 3. re-stake the node nodes that were in the queue
	node.ProcessReStake(t, initialNodes.queue)
	nodesConfigStakingV4Step1 = node.NodesConfig
	requireSameSliceDifferentOrder(t, initialNodes.queue, nodesConfigStakingV4Step1.auction)

	// 4. Check config after first staking v4 epoch, WITHOUT distribution from auction -> waiting
	node.Process(t, 6)
	nodesConfigStakingV4Step2 := node.NodesConfig
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Step2.eligible), totalEligible) // 1600

	numOfShuffledOut := int((numOfShards + 1) * numOfNodesToShufflePerShard) // 320
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Step2.shuffledOut), numOfShuffledOut)

	newWaiting := totalWaiting - numOfShuffledOut // 1280 (1600 - 320)
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Step2.waiting), newWaiting)

	// 380 (320 from shuffled out + 60 from initial staking queue -> auction from stakingV4 init)
	auctionListSize := numOfShuffledOut + len(nodesConfigStakingV4Step1.auction)
	require.Len(t, nodesConfigStakingV4Step2.auction, auctionListSize)
	requireSliceContains(t, nodesConfigStakingV4Step2.auction, nodesConfigStakingV4Step1.auction)

	require.Empty(t, nodesConfigStakingV4Step2.queue)
	require.Empty(t, nodesConfigStakingV4Step2.leaving)

	// 320 nodes which are now in eligible are from previous waiting list
	requireSliceContainsNumOfElements(t, getAllPubKeys(nodesConfigStakingV4Step2.eligible), getAllPubKeys(nodesConfigStakingV4Step1.waiting), numOfShuffledOut)

	// All shuffled out are from previous staking v4 init eligible
	requireMapContains(t, nodesConfigStakingV4Step1.eligible, getAllPubKeys(nodesConfigStakingV4Step2.shuffledOut))

	// All shuffled out are in auction
	requireSliceContains(t, nodesConfigStakingV4Step2.auction, getAllPubKeys(nodesConfigStakingV4Step2.shuffledOut))

	// No auction node from previous epoch has been moved to waiting
	requireMapDoesNotContain(t, nodesConfigStakingV4Step2.waiting, nodesConfigStakingV4Step1.auction)

	epochs := 0
	prevConfig := nodesConfigStakingV4Step2
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

		checkStakingV4EpochChangeFlow(t, newNodeConfig, prevConfig, numOfShuffledOut, numOfUnselectedNodesFromAuction, numOfSelectedNodesFromAuction)
		prevConfig = newNodeConfig
		epochs++
	}
}

func TestStakingV4MetaProcessor_ProcessMultipleNodesWithSameSetupExpectSameRootHash(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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

func TestStakingV4_UnStakeNodesWithNotEnoughFunds(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	pubKeys := generateAddresses(0, 20)

	// Owner1 has 8 nodes, but enough stake for just 7 nodes. At the end of the epoch(staking v4 init),
	// his last node from staking queue should be unStaked
	owner1 := "owner1"
	owner1Stats := &OwnerStats{
		EligibleBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[:3],
		},
		WaitingBlsKeys: map[uint32][][]byte{
			0: pubKeys[3:6],
		},
		StakingQueueKeys: pubKeys[6:8],
		TotalStake:       big.NewInt(7 * nodePrice),
	}

	// Owner2 has 6 nodes, but enough stake for just 5 nodes. At the end of the epoch(staking v4 init),
	// one node from waiting list should be unStaked
	owner2 := "owner2"
	owner2Stats := &OwnerStats{
		EligibleBlsKeys: map[uint32][][]byte{
			0: pubKeys[8:11],
		},
		WaitingBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[11:14],
		},
		TotalStake: big.NewInt(5 * nodePrice),
	}

	// Owner3 has 2 nodes in staking queue with topUp = nodePrice
	owner3 := "owner3"
	owner3Stats := &OwnerStats{
		StakingQueueKeys: pubKeys[14:16],
		TotalStake:       big.NewInt(3 * nodePrice),
	}

	// Owner4 has 1 node in staking queue with topUp = nodePrice
	owner4 := "owner4"
	owner4Stats := &OwnerStats{
		StakingQueueKeys: pubKeys[16:17],
		TotalStake:       big.NewInt(2 * nodePrice),
	}

	cfg := &InitialNodesConfig{
		MetaConsensusGroupSize:        2,
		ShardConsensusGroupSize:       2,
		MinNumberOfEligibleShardNodes: 3,
		MinNumberOfEligibleMetaNodes:  3,
		NumOfShards:                   1,
		Owners: map[string]*OwnerStats{
			owner1: owner1Stats,
			owner2: owner2Stats,
			owner3: owner3Stats,
			owner4: owner4Stats,
		},
		MaxNodesChangeConfig: []config.MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            12,
				NodesToShufflePerShard: 1,
			},
			{
				EpochEnable:            stakingV4Step3EnableEpoch,
				MaxNumNodes:            10,
				NodesToShufflePerShard: 1,
			},
		},
	}
	node := NewTestMetaProcessorWithCustomNodes(cfg)
	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	// 1. Check initial config is correct
	currNodesConfig := node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesConfig.eligible), 6)
	require.Len(t, getAllPubKeys(currNodesConfig.waiting), 6)
	require.Len(t, currNodesConfig.eligible[core.MetachainShardId], 3)
	require.Len(t, currNodesConfig.waiting[core.MetachainShardId], 3)
	require.Len(t, currNodesConfig.eligible[0], 3)
	require.Len(t, currNodesConfig.waiting[0], 3)

	requireSliceContainsNumOfElements(t, currNodesConfig.eligible[core.MetachainShardId], owner1Stats.EligibleBlsKeys[core.MetachainShardId], 3)
	requireSliceContainsNumOfElements(t, currNodesConfig.waiting[core.MetachainShardId], owner2Stats.WaitingBlsKeys[core.MetachainShardId], 3)
	requireSliceContainsNumOfElements(t, currNodesConfig.eligible[0], owner2Stats.EligibleBlsKeys[0], 3)
	requireSliceContainsNumOfElements(t, currNodesConfig.waiting[0], owner1Stats.WaitingBlsKeys[0], 3)

	owner1StakingQueue := owner1Stats.StakingQueueKeys
	owner3StakingQueue := owner3Stats.StakingQueueKeys
	owner4StakingQueue := owner4Stats.StakingQueueKeys
	queue := make([][]byte, 0)
	queue = append(queue, owner1StakingQueue...)
	queue = append(queue, owner3StakingQueue...)
	queue = append(queue, owner4StakingQueue...)
	require.Len(t, currNodesConfig.queue, 5)
	requireSameSliceDifferentOrder(t, currNodesConfig.queue, queue)

	require.Empty(t, currNodesConfig.shuffledOut)
	require.Empty(t, currNodesConfig.auction)

	// 2. Check config after staking v4 initialization
	node.Process(t, 5)
	currNodesConfig = node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesConfig.eligible), 6)
	require.Len(t, getAllPubKeys(currNodesConfig.waiting), 5)
	require.Len(t, currNodesConfig.eligible[core.MetachainShardId], 3)
	require.Len(t, currNodesConfig.waiting[core.MetachainShardId], 2)
	require.Len(t, currNodesConfig.eligible[0], 3)
	require.Len(t, currNodesConfig.waiting[0], 3)

	// Owner1 will have the second node from queue removed, before adding all the nodes to auction list
	queue = remove(queue, owner1StakingQueue[1])
	require.Empty(t, currNodesConfig.queue)
	require.Empty(t, currNodesConfig.auction) // all nodes from the queue should be unStaked and the auction list should be empty

	// Owner2 will have one of the nodes in waiting list removed
	require.Len(t, getAllPubKeys(currNodesConfig.leaving), 1)
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesConfig.leaving), getAllPubKeys(owner2Stats.WaitingBlsKeys), 1)

	// Owner1 will unStake some EGLD => at the end of next epoch, he should not be able to reStake all the nodes
	unStake(t, []byte(owner1), node.AccountsAdapter, node.Marshaller, big.NewInt(0.1*nodePrice))

	// 3. ReStake the nodes that were in the queue
	queue = remove(queue, owner1StakingQueue[0])
	node.ProcessReStake(t, queue)
	currNodesConfig = node.NodesConfig
	require.Len(t, currNodesConfig.auction, 3)
	requireSameSliceDifferentOrder(t, currNodesConfig.auction, queue)

	// 4. Check config in epoch = staking v4
	node.Process(t, 4)
	currNodesConfig = node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesConfig.eligible), 6)
	require.Len(t, getAllPubKeys(currNodesConfig.waiting), 3)
	require.Len(t, getAllPubKeys(currNodesConfig.shuffledOut), 2)

	require.Len(t, currNodesConfig.eligible[core.MetachainShardId], 3)
	require.Len(t, currNodesConfig.waiting[core.MetachainShardId], 1)
	require.Len(t, currNodesConfig.shuffledOut[core.MetachainShardId], 1)
	require.Len(t, currNodesConfig.eligible[0], 3)
	require.Len(t, currNodesConfig.waiting[0], 2)
	require.Len(t, currNodesConfig.shuffledOut[0], 1)

	require.Len(t, currNodesConfig.auction, 3)
	requireSameSliceDifferentOrder(t, currNodesConfig.auction, queue)
	require.Len(t, getAllPubKeys(currNodesConfig.leaving), 0)
	// There are no more unStaked nodes left from owner1 because of insufficient funds
	requireSliceContainsNumOfElements(t, getAllPubKeysFromConfig(currNodesConfig), owner1StakingQueue, 0)

	// Owner3 will unStake EGLD => he will have negative top-up at the selection time => one of his nodes will be unStaked.
	// His other node should not have been selected => remains in auction.
	// Meanwhile, owner4 had never unStaked EGLD => his node from auction list will be distributed to waiting
	unStake(t, []byte(owner3), node.AccountsAdapter, node.Marshaller, big.NewInt(2*nodePrice))

	// 5. Check config in epoch = staking v4 step3
	node.Process(t, 5)
	currNodesConfig = node.NodesConfig
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesConfig.leaving), owner3StakingQueue, 1)
	requireSliceContainsNumOfElements(t, currNodesConfig.auction, owner3StakingQueue, 1)
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesConfig.waiting), owner4StakingQueue, 1)
}

func TestStakingV4_StakeNewNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	pubKeys := generateAddresses(0, 20)

	// Owner1 has 6 nodes, zero top up
	owner1 := "owner1"
	owner1Stats := &OwnerStats{
		EligibleBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[:2],
		},
		WaitingBlsKeys: map[uint32][][]byte{
			0: pubKeys[2:4],
		},
		StakingQueueKeys: pubKeys[4:6],
		TotalStake:       big.NewInt(6 * nodePrice),
	}

	// Owner2 has 4 nodes, zero top up
	owner2 := "owner2"
	owner2Stats := &OwnerStats{
		EligibleBlsKeys: map[uint32][][]byte{
			0: pubKeys[6:8],
		},
		WaitingBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[8:10],
		},
		TotalStake: big.NewInt(4 * nodePrice),
	}
	// Owner3 has 1 node in staking queue with topUp = nodePrice
	owner3 := "owner3"
	owner3Stats := &OwnerStats{
		StakingQueueKeys: pubKeys[10:11],
		TotalStake:       big.NewInt(2 * nodePrice),
	}

	cfg := &InitialNodesConfig{
		MetaConsensusGroupSize:        1,
		ShardConsensusGroupSize:       1,
		MinNumberOfEligibleShardNodes: 1,
		MinNumberOfEligibleMetaNodes:  1,
		NumOfShards:                   1,
		Owners: map[string]*OwnerStats{
			owner1: owner1Stats,
			owner2: owner2Stats,
			owner3: owner3Stats,
		},
		MaxNodesChangeConfig: []config.MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            8,
				NodesToShufflePerShard: 1,
			},
		},
	}
	node := NewTestMetaProcessorWithCustomNodes(cfg)
	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	// 1.1 Check initial config is correct
	currNodesConfig := node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesConfig.eligible), 4)
	require.Len(t, getAllPubKeys(currNodesConfig.waiting), 4)
	require.Len(t, currNodesConfig.eligible[core.MetachainShardId], 2)
	require.Len(t, currNodesConfig.waiting[core.MetachainShardId], 2)
	require.Len(t, currNodesConfig.eligible[0], 2)
	require.Len(t, currNodesConfig.waiting[0], 2)

	owner1StakingQueue := owner1Stats.StakingQueueKeys
	owner3StakingQueue := owner3Stats.StakingQueueKeys
	queue := make([][]byte, 0)
	queue = append(queue, owner1StakingQueue...)
	queue = append(queue, owner3StakingQueue...)
	require.Len(t, currNodesConfig.queue, 3)
	requireSameSliceDifferentOrder(t, currNodesConfig.queue, queue)

	require.Empty(t, currNodesConfig.shuffledOut)
	require.Empty(t, currNodesConfig.auction)

	// NewOwner0 stakes 1 node with top up = 0 before staking v4; should be sent to staking queue
	newOwner0 := "newOwner0"
	newNodes0 := map[string]*NodesRegisterData{
		newOwner0: {
			BLSKeys:    [][]byte{generateAddress(333)},
			TotalStake: big.NewInt(nodePrice),
		},
	}

	// 1.2 Check staked node before staking v4 is sent to staking queue
	node.ProcessStake(t, newNodes0)
	queue = append(queue, newNodes0[newOwner0].BLSKeys...)
	currNodesConfig = node.NodesConfig
	require.Len(t, currNodesConfig.queue, 4)
	requireSameSliceDifferentOrder(t, currNodesConfig.queue, queue)

	// NewOwner1 stakes 1 node with top up = 2*node price; should be sent to auction list
	newOwner1 := "newOwner1"
	newNodes1 := map[string]*NodesRegisterData{
		newOwner1: {
			BLSKeys:    [][]byte{generateAddress(444)},
			TotalStake: big.NewInt(3 * nodePrice),
		},
	}
	// 2. Check config after staking v4 init when a new node is staked
	node.Process(t, 4)
	node.ProcessStake(t, newNodes1)
	node.ProcessReStake(t, queue)
	currNodesConfig = node.NodesConfig
	queue = append(queue, newNodes1[newOwner1].BLSKeys...)
	require.Empty(t, currNodesConfig.queue)
	require.Empty(t, currNodesConfig.leaving)
	require.Len(t, currNodesConfig.auction, 5)
	requireSameSliceDifferentOrder(t, currNodesConfig.auction, queue)

	// NewOwner2 stakes 2 node with top up = 2*node price; should be sent to auction list
	newOwner2 := "newOwner2"
	newNodes2 := map[string]*NodesRegisterData{
		newOwner2: {
			BLSKeys:    [][]byte{generateAddress(555), generateAddress(666)},
			TotalStake: big.NewInt(4 * nodePrice),
		},
	}
	// 2. Check in epoch = staking v4 step2 when 2 new nodes are staked
	node.Process(t, 4)
	node.ProcessStake(t, newNodes2)
	currNodesConfig = node.NodesConfig
	queue = append(queue, newNodes2[newOwner2].BLSKeys...)
	require.Empty(t, currNodesConfig.queue)
	requireSliceContainsNumOfElements(t, currNodesConfig.auction, queue, 7)

	// 3. Epoch =  staking v4 step3
	// Only the new 2 owners + owner3 had enough top up to be distributed to waiting.
	// Meanwhile, owner1 which had 0 top up, still has his bls keys in auction, along with newOwner0
	node.Process(t, 5)
	currNodesConfig = node.NodesConfig
	require.Empty(t, currNodesConfig.queue)
	requireMapContains(t, currNodesConfig.waiting, newNodes1[newOwner1].BLSKeys)
	requireMapContains(t, currNodesConfig.waiting, newNodes2[newOwner2].BLSKeys)
	requireMapContains(t, currNodesConfig.waiting, owner3StakingQueue)
	requireSliceContains(t, currNodesConfig.auction, owner1StakingQueue)
	requireSliceContains(t, currNodesConfig.auction, newNodes0[newOwner0].BLSKeys)
}

func TestStakingV4_UnStakeNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	pubKeys := generateAddresses(0, 20)

	owner1 := "owner1"
	owner1Stats := &OwnerStats{
		EligibleBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[:2],
		},
		WaitingBlsKeys: map[uint32][][]byte{
			0: pubKeys[2:4],
		},
		StakingQueueKeys: pubKeys[4:6],
		TotalStake:       big.NewInt(10 * nodePrice),
	}

	owner2 := "owner2"
	owner2Stats := &OwnerStats{
		EligibleBlsKeys: map[uint32][][]byte{
			0: pubKeys[6:8],
		},
		WaitingBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[8:12],
		},
		StakingQueueKeys: pubKeys[12:15],
		TotalStake:       big.NewInt(10 * nodePrice),
	}

	owner3 := "owner3"
	owner3Stats := &OwnerStats{
		StakingQueueKeys: pubKeys[15:17],
		TotalStake:       big.NewInt(6 * nodePrice),
	}

	cfg := &InitialNodesConfig{
		MetaConsensusGroupSize:        1,
		ShardConsensusGroupSize:       1,
		MinNumberOfEligibleShardNodes: 2,
		MinNumberOfEligibleMetaNodes:  2,
		NumOfShards:                   1,
		Owners: map[string]*OwnerStats{
			owner1: owner1Stats,
			owner2: owner2Stats,
			owner3: owner3Stats,
		},
		MaxNodesChangeConfig: []config.MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            10,
				NodesToShufflePerShard: 1,
			},
		},
	}
	node := NewTestMetaProcessorWithCustomNodes(cfg)
	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	// 1. Check initial config is correct
	currNodesConfig := node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesConfig.eligible), 4)
	require.Len(t, getAllPubKeys(currNodesConfig.waiting), 6)
	require.Len(t, currNodesConfig.eligible[core.MetachainShardId], 2)
	require.Len(t, currNodesConfig.waiting[core.MetachainShardId], 4)
	require.Len(t, currNodesConfig.eligible[0], 2)
	require.Len(t, currNodesConfig.waiting[0], 2)
	require.Empty(t, currNodesConfig.shuffledOut)
	require.Empty(t, currNodesConfig.auction)

	owner1StakingQueue := owner1Stats.StakingQueueKeys
	owner2StakingQueue := owner2Stats.StakingQueueKeys
	owner3StakingQueue := owner3Stats.StakingQueueKeys
	queue := make([][]byte, 0)
	queue = append(queue, owner1StakingQueue...)
	queue = append(queue, owner2StakingQueue...)
	queue = append(queue, owner3StakingQueue...)
	require.Len(t, currNodesConfig.queue, 7)
	requireSameSliceDifferentOrder(t, currNodesConfig.queue, queue)

	// 1.1 Owner2 unStakes one of his staking queue nodes. Node should be removed from staking queue list
	node.ProcessUnStake(t, map[string][][]byte{
		owner2: {owner2Stats.StakingQueueKeys[0]},
	})
	currNodesConfig = node.NodesConfig
	queue = remove(queue, owner2Stats.StakingQueueKeys[0])
	require.Len(t, currNodesConfig.queue, 6)
	requireSameSliceDifferentOrder(t, currNodesConfig.queue, queue)
	require.Empty(t, currNodesConfig.new)
	require.Empty(t, currNodesConfig.auction)

	// 1.2 Owner2 unStakes one of his waiting list keys. First node from staking queue should be added to fill its place.
	copy(queue, currNodesConfig.queue) // copy queue to local variable so we have the queue in same order
	node.ProcessUnStake(t, map[string][][]byte{
		owner2: {owner2Stats.WaitingBlsKeys[core.MetachainShardId][0]},
	})
	currNodesConfig = node.NodesConfig
	require.Len(t, currNodesConfig.new, 1)
	requireSliceContains(t, queue, currNodesConfig.new)
	require.Empty(t, currNodesConfig.auction)
	queue = remove(queue, currNodesConfig.new[0])
	require.Len(t, currNodesConfig.queue, 5)
	requireSameSliceDifferentOrder(t, queue, currNodesConfig.queue)

	// 2. Check config after staking v4 step1
	node.Process(t, 3)
	currNodesConfig = node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesConfig.eligible), 4)
	require.Len(t, getAllPubKeys(currNodesConfig.waiting), 6)
	// Owner2's node from waiting list which was unStaked in previous epoch is now leaving
	require.Len(t, currNodesConfig.leaving, 1)
	require.Equal(t, owner2Stats.WaitingBlsKeys[core.MetachainShardId][0], currNodesConfig.leaving[core.MetachainShardId][0])
	require.Empty(t, currNodesConfig.auction) // all nodes from queue have been unStaked, the auction list is empty

	// 2.1 restake the nodes that were on the queue
	node.ProcessReStake(t, queue)
	currNodesConfig = node.NodesConfig
	requireSameSliceDifferentOrder(t, queue, currNodesConfig.auction)

	// 2.2 Owner3 unStakes one of his nodes from auction
	node.ProcessUnStake(t, map[string][][]byte{
		owner3: {owner3StakingQueue[1]},
	})
	unStakedNodesInStakingV4Step1Epoch := make([][]byte, 0)
	unStakedNodesInStakingV4Step1Epoch = append(unStakedNodesInStakingV4Step1Epoch, owner3StakingQueue[1])
	currNodesConfig = node.NodesConfig
	queue = remove(queue, owner3StakingQueue[1])
	require.Len(t, currNodesConfig.auction, 4)
	requireSameSliceDifferentOrder(t, queue, currNodesConfig.auction)
	require.Empty(t, currNodesConfig.queue)
	require.Empty(t, currNodesConfig.new)

	// 2.3 Owner1 unStakes 2 nodes: one from auction + one active
	node.ProcessUnStake(t, map[string][][]byte{
		owner1: {owner1StakingQueue[1], owner1Stats.WaitingBlsKeys[0][0]},
	})
	unStakedNodesInStakingV4Step1Epoch = append(unStakedNodesInStakingV4Step1Epoch, owner1StakingQueue[1])
	unStakedNodesInStakingV4Step1Epoch = append(unStakedNodesInStakingV4Step1Epoch, owner1Stats.WaitingBlsKeys[0][0])
	currNodesConfig = node.NodesConfig
	queue = remove(queue, owner1StakingQueue[1])
	require.Len(t, currNodesConfig.auction, 3)
	requireSameSliceDifferentOrder(t, queue, currNodesConfig.auction)
	require.Empty(t, currNodesConfig.queue)
	require.Empty(t, currNodesConfig.new)

	// 3. Check config in epoch = staking v4 step2
	node.Process(t, 3)
	currNodesConfig = node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesConfig.eligible), 4)
	require.Len(t, getAllPubKeys(currNodesConfig.waiting), 4)
	require.Len(t, getAllPubKeys(currNodesConfig.leaving), 3)
	// All unStaked nodes in previous epoch are now leaving
	requireMapContains(t, currNodesConfig.leaving, unStakedNodesInStakingV4Step1Epoch)
	// 3.1 Owner2 unStakes one of his nodes from auction
	node.ProcessUnStake(t, map[string][][]byte{
		owner2: {owner2StakingQueue[1]},
	})
	currNodesConfig = node.NodesConfig
	queue = remove(queue, owner2StakingQueue[1])
	shuffledOutNodes := getAllPubKeys(currNodesConfig.shuffledOut)
	require.Len(t, currNodesConfig.auction, len(shuffledOutNodes)+len(queue))
	requireSliceContains(t, currNodesConfig.auction, shuffledOutNodes)
	requireSliceContains(t, currNodesConfig.auction, queue)

	// 4. Check config after whole staking v4 chain is ready, when one of the owners unStakes a node
	node.Process(t, 4)
	currNodesConfig = node.NodesConfig
	node.ProcessUnStake(t, map[string][][]byte{
		owner2: {owner2Stats.EligibleBlsKeys[0][0]},
	})
	node.Process(t, 4)
	currNodesConfig = node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesConfig.leaving), 1)
	requireMapContains(t, currNodesConfig.leaving, [][]byte{owner2Stats.EligibleBlsKeys[0][0]})
	require.Empty(t, currNodesConfig.new)
	require.Empty(t, currNodesConfig.queue)

	// 4.1 NewOwner stakes 1 node, should be sent to auction
	newOwner := "newOwner1"
	newNode := map[string]*NodesRegisterData{
		newOwner: {
			BLSKeys:    [][]byte{generateAddress(444)},
			TotalStake: big.NewInt(2 * nodePrice),
		},
	}
	node.ProcessStake(t, newNode)
	currNodesConfig = node.NodesConfig
	requireSliceContains(t, currNodesConfig.auction, newNode[newOwner].BLSKeys)

	// 4.2 NewOwner unStakes his node, he should not be in auction anymore + set to leaving
	node.ProcessUnStake(t, map[string][][]byte{
		newOwner: {newNode[newOwner].BLSKeys[0]},
	})
	currNodesConfig = node.NodesConfig
	requireSliceContainsNumOfElements(t, currNodesConfig.auction, newNode[newOwner].BLSKeys, 0)
	node.Process(t, 3)
	currNodesConfig = node.NodesConfig
	requireMapContains(t, currNodesConfig.leaving, newNode[newOwner].BLSKeys)
}

func TestStakingV4_JailAndUnJailNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	pubKeys := generateAddresses(0, 20)

	owner1 := "owner1"
	owner1Stats := &OwnerStats{
		EligibleBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[:2],
		},
		WaitingBlsKeys: map[uint32][][]byte{
			0: pubKeys[2:4],
		},
		StakingQueueKeys: pubKeys[4:6],
		TotalStake:       big.NewInt(10 * nodePrice),
	}

	owner2 := "owner2"
	owner2Stats := &OwnerStats{
		EligibleBlsKeys: map[uint32][][]byte{
			0: pubKeys[6:8],
		},
		WaitingBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[8:12],
		},
		StakingQueueKeys: pubKeys[12:15],
		TotalStake:       big.NewInt(10 * nodePrice),
	}

	cfg := &InitialNodesConfig{
		MetaConsensusGroupSize:        1,
		ShardConsensusGroupSize:       1,
		MinNumberOfEligibleShardNodes: 2,
		MinNumberOfEligibleMetaNodes:  2,
		NumOfShards:                   1,
		Owners: map[string]*OwnerStats{
			owner1: owner1Stats,
			owner2: owner2Stats,
		},
		MaxNodesChangeConfig: []config.MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            10,
				NodesToShufflePerShard: 1,
			},
			{
				EpochEnable:            stakingV4Step3EnableEpoch,
				MaxNumNodes:            4,
				NodesToShufflePerShard: 1,
			},
		},
	}
	node := NewTestMetaProcessorWithCustomNodes(cfg)
	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	// 1. Check initial config is correct
	currNodesConfig := node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesConfig.eligible), 4)
	require.Len(t, getAllPubKeys(currNodesConfig.waiting), 6)
	require.Len(t, currNodesConfig.eligible[core.MetachainShardId], 2)
	require.Len(t, currNodesConfig.waiting[core.MetachainShardId], 4)
	require.Len(t, currNodesConfig.eligible[0], 2)
	require.Len(t, currNodesConfig.waiting[0], 2)
	require.Empty(t, currNodesConfig.shuffledOut)
	require.Empty(t, currNodesConfig.auction)

	owner1StakingQueue := owner1Stats.StakingQueueKeys
	owner2StakingQueue := owner2Stats.StakingQueueKeys
	queue := make([][]byte, 0)
	queue = append(queue, owner1StakingQueue...)
	queue = append(queue, owner2StakingQueue...)
	require.Len(t, currNodesConfig.queue, 5)
	requireSameSliceDifferentOrder(t, currNodesConfig.queue, queue)

	// 1.1 Jail 4 nodes:
	// - 2 nodes from waiting list shard = 0
	// - 2 nodes from waiting list shard = meta chain
	jailedNodes := make([][]byte, 0)
	jailedNodes = append(jailedNodes, owner1Stats.WaitingBlsKeys[0]...)
	jailedNodes = append(jailedNodes, owner2Stats.WaitingBlsKeys[core.MetachainShardId][:2]...)
	node.ProcessJail(t, jailedNodes)

	// 1.2 UnJail 2 nodes from initial jailed nodes:
	// - 1 node from waiting list shard = 0
	// - 1 node from waiting list shard = meta chain
	unJailedNodes := make([][]byte, 0)
	unJailedNodes = append(unJailedNodes, owner1Stats.WaitingBlsKeys[0][0])
	unJailedNodes = append(unJailedNodes, owner2Stats.WaitingBlsKeys[core.MetachainShardId][0])
	jailedNodes = remove(jailedNodes, unJailedNodes[0])
	jailedNodes = remove(jailedNodes, unJailedNodes[1])
	node.ProcessUnJail(t, unJailedNodes)

	// 2. Two jailed nodes are now leaving; the other two unJailed nodes are re-staked and distributed on waiting list
	node.Process(t, 3)
	currNodesConfig = node.NodesConfig
	requireMapContains(t, currNodesConfig.leaving, jailedNodes)
	requireMapContains(t, currNodesConfig.waiting, unJailedNodes)
	requireSameSliceDifferentOrder(t, currNodesConfig.auction, make([][]byte, 0))
	require.Len(t, getAllPubKeys(currNodesConfig.eligible), 4)
	require.Len(t, getAllPubKeys(currNodesConfig.waiting), 4)
	require.Empty(t, currNodesConfig.queue)

	// 2.1 ReStake the nodes that were in the queue
	// but first, we need to unJail the nodes
	node.ProcessUnJail(t, jailedNodes)
	node.ProcessReStake(t, queue)
	currNodesConfig = node.NodesConfig
	queue = append(queue, jailedNodes...)
	require.Empty(t, currNodesConfig.queue)
	requireSameSliceDifferentOrder(t, currNodesConfig.auction, queue)

	// 3. Epoch = stakingV4Step2
	node.Process(t, 1)
	currNodesConfig = node.NodesConfig
	queue = append(queue, getAllPubKeys(currNodesConfig.shuffledOut)...)
	require.Empty(t, currNodesConfig.queue)
	requireSameSliceDifferentOrder(t, currNodesConfig.auction, queue)

	// 3.1 Jail a random node from waiting list
	newJailed := getAllPubKeys(currNodesConfig.waiting)[:1]
	node.ProcessJail(t, newJailed)

	// 4. Epoch = stakingV4Step3;
	// 4.1 Expect jailed node from waiting list is now leaving
	node.Process(t, 4)
	currNodesConfig = node.NodesConfig
	requireMapContains(t, currNodesConfig.leaving, newJailed)
	requireSliceContainsNumOfElements(t, currNodesConfig.auction, newJailed, 0)
	require.Empty(t, currNodesConfig.queue)

	// 4.2 	UnJail previous node and expect it is sent to auction
	node.ProcessUnJail(t, newJailed)
	currNodesConfig = node.NodesConfig
	requireSliceContains(t, currNodesConfig.auction, newJailed)
	require.Empty(t, currNodesConfig.queue)

	// 5. Epoch is now after whole staking v4 chain is activated
	node.Process(t, 3)
	currNodesConfig = node.NodesConfig
	queue = currNodesConfig.auction
	newJailed = queue[:1]
	newUnJailed := newJailed[0]

	// 5.1 Take a random node from auction and jail it; expect it is removed from auction list
	node.ProcessJail(t, newJailed)
	queue = remove(queue, newJailed[0])
	currNodesConfig = node.NodesConfig
	requireSameSliceDifferentOrder(t, queue, currNodesConfig.auction)

	// 5.2 UnJail previous node; expect it is sent back to auction
	node.ProcessUnJail(t, [][]byte{newUnJailed})
	queue = append(queue, newUnJailed)
	currNodesConfig = node.NodesConfig
	requireSameSliceDifferentOrder(t, queue, currNodesConfig.auction)
	require.Empty(t, node.NodesConfig.queue)
}

func TestStakingV4_DifferentEdgeCasesWithNotEnoughNodesInWaitingShouldSendShuffledToToWaiting(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	pubKeys := generateAddresses(0, 20)

	owner1 := "owner1"
	owner1Stats := &OwnerStats{
		EligibleBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[:4],
			0:                     pubKeys[4:8],
		},
		WaitingBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[8:9],
			0:                     pubKeys[9:10],
		},
		TotalStake: big.NewInt(20 * nodePrice),
	}

	cfg := &InitialNodesConfig{
		MetaConsensusGroupSize:        2,
		ShardConsensusGroupSize:       2,
		MinNumberOfEligibleShardNodes: 4,
		MinNumberOfEligibleMetaNodes:  4,
		NumOfShards:                   1,
		Owners: map[string]*OwnerStats{
			owner1: owner1Stats,
		},
		MaxNodesChangeConfig: []config.MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            12,
				NodesToShufflePerShard: 1,
			},
			{
				EpochEnable:            stakingV4Step3EnableEpoch, // epoch 3
				MaxNumNodes:            10,
				NodesToShufflePerShard: 1,
			},
			{
				EpochEnable:            6,
				MaxNumNodes:            12,
				NodesToShufflePerShard: 1,
			},
		},
	}
	node := NewTestMetaProcessorWithCustomNodes(cfg)
	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	// 1. Check initial config is correct
	expectedNodesNum := &configNum{
		eligible: map[uint32]int{
			core.MetachainShardId: 4,
			0:                     4,
		},
		waiting: map[uint32]int{
			core.MetachainShardId: 1,
			0:                     1,
		},
	}
	currNodesConfig := node.NodesConfig
	checkConfig(t, expectedNodesNum, currNodesConfig)

	// During these 9 epochs, we will always have:
	// - 10 activeNodes (8 eligible + 2 waiting)
	// - 1 node to shuffle out per shard
	// Meanwhile, maxNumNodes changes from 12-10-12
	// Since activeNodes <= maxNumNodes, shuffled out nodes will always be sent directly to waiting list,
	// instead of auction(there is no reason to send them to auction, they will be selected anyway)
	epoch := uint32(0)
	numOfShuffledOut := 2
	numRemainingEligible := 6
	prevNodesConfig := currNodesConfig
	for epoch < 9 {
		node.Process(t, 5)

		currNodesConfig = node.NodesConfig
		checkConfig(t, expectedNodesNum, currNodesConfig)
		checkShuffledOutNodes(t, currNodesConfig, prevNodesConfig, numOfShuffledOut, numRemainingEligible)

		prevNodesConfig = currNodesConfig
		epoch++
	}

	require.Equal(t, epoch, node.EpochStartTrigger.Epoch())

	// Epoch = 9 with:
	// - activeNodes = 10
	// - maxNumNodes = 12
	// Owner2 stakes 2 nodes, which should be initially sent to auction list
	owner2Nodes := pubKeys[10:12]
	node.ProcessStake(t, map[string]*NodesRegisterData{
		"owner2": {
			BLSKeys:    owner2Nodes,
			TotalStake: big.NewInt(5 * nodePrice),
		},
	})
	currNodesConfig = node.NodesConfig
	expectedNodesNum.auction = 2
	checkConfig(t, expectedNodesNum, currNodesConfig)
	requireSameSliceDifferentOrder(t, currNodesConfig.auction, owner2Nodes)

	// Epoch = 10 with:
	// - activeNodes = 12
	// - maxNumNodes = 12
	// Owner2's new nodes are selected from auction and distributed to waiting list
	node.Process(t, 5)
	currNodesConfig = node.NodesConfig
	expectedNodesNum.waiting[core.MetachainShardId]++
	expectedNodesNum.waiting[0]++
	expectedNodesNum.auction = 0
	checkConfig(t, expectedNodesNum, currNodesConfig)
	checkShuffledOutNodes(t, currNodesConfig, prevNodesConfig, numOfShuffledOut, numRemainingEligible)
	requireSliceContains(t, getAllPubKeys(currNodesConfig.waiting), owner2Nodes)

	// During epochs 10-13, we will have:
	// - activeNodes = 12
	// - maxNumNodes = 12
	// Since activeNodes == maxNumNodes, shuffled out nodes will always be sent directly to waiting list, instead of auction
	epoch = 10
	require.Equal(t, epoch, node.EpochStartTrigger.Epoch())
	prevNodesConfig = currNodesConfig
	for epoch < 13 {
		node.Process(t, 5)

		currNodesConfig = node.NodesConfig
		checkConfig(t, expectedNodesNum, currNodesConfig)
		checkShuffledOutNodes(t, currNodesConfig, prevNodesConfig, numOfShuffledOut, numRemainingEligible)

		prevNodesConfig = currNodesConfig
		epoch++
	}

	// Epoch = 13 with:
	// - activeNodes = 12
	// - maxNumNodes = 12
	// Owner3 stakes 2 nodes, which should be initially sent to auction list
	owner3Nodes := pubKeys[12:14]
	node.ProcessStake(t, map[string]*NodesRegisterData{
		"owner3": {
			BLSKeys:    owner3Nodes,
			TotalStake: big.NewInt(5 * nodePrice),
		},
	})
	currNodesConfig = node.NodesConfig
	expectedNodesNum.auction = 2
	checkConfig(t, expectedNodesNum, currNodesConfig)
	requireSameSliceDifferentOrder(t, currNodesConfig.auction, owner3Nodes)

	// During epochs 14-18, we will have:
	// - activeNodes = 14
	// - maxNumNodes = 12
	// Since activeNodes > maxNumNodes, shuffled out nodes (2) will be sent to auction list
	node.Process(t, 5)
	prevNodesConfig = node.NodesConfig
	epoch = 14
	require.Equal(t, epoch, node.EpochStartTrigger.Epoch())

	numOfUnselectedNodesFromAuction := 0
	numOfSelectedNodesFromAuction := 2
	for epoch < 18 {
		checkConfig(t, expectedNodesNum, currNodesConfig)

		node.Process(t, 5)
		currNodesConfig = node.NodesConfig
		checkStakingV4EpochChangeFlow(t, currNodesConfig, prevNodesConfig, numOfShuffledOut, numOfUnselectedNodesFromAuction, numOfSelectedNodesFromAuction)

		prevNodesConfig = currNodesConfig
		epoch++
	}

	// Epoch = 18, with:
	// - activeNodes = 14
	// - maxNumNodes = 12
	// Owner3 unStakes one of his nodes
	node.ProcessUnStake(t, map[string][][]byte{
		"owner3": {owner3Nodes[0]},
	})

	// Epoch = 19, with:
	// - activeNodes = 13
	// - maxNumNodes = 12
	// Owner3's unStaked node is now leaving
	node.Process(t, 5)
	currNodesConfig = node.NodesConfig
	require.Len(t, currNodesConfig.leaving, 1)
	requireMapContains(t, currNodesConfig.leaving, [][]byte{owner3Nodes[0]})

	epoch = 19
	require.Equal(t, epoch, node.EpochStartTrigger.Epoch())

	// During epochs 19-23, we will have:
	// - activeNodes = 13
	// - maxNumNodes = 12
	// Since activeNodes > maxNumNodes:
	// - shuffled out nodes (2) will be sent to auction list
	// - waiting lists will be unbalanced (3 in total: 1 + 2 per shard)
	// - no node will spend extra epochs in eligible/waiting, since waiting lists will always be refilled
	prevNodesConfig = node.NodesConfig
	for epoch < 23 {
		require.Len(t, getAllPubKeys(currNodesConfig.eligible), 8)
		require.Len(t, getAllPubKeys(currNodesConfig.waiting), 3)
		require.Len(t, currNodesConfig.eligible[core.MetachainShardId], 4)
		require.Len(t, currNodesConfig.eligible[0], 4)
		require.Len(t, currNodesConfig.auction, 2)

		node.Process(t, 5)

		currNodesConfig = node.NodesConfig
		checkStakingV4EpochChangeFlow(t, currNodesConfig, prevNodesConfig, numOfShuffledOut, numOfUnselectedNodesFromAuction, numOfSelectedNodesFromAuction)

		prevNodesConfig = currNodesConfig
		epoch++
	}
}

func TestStakingV4_NewlyStakedNodesInStakingV4Step2ShouldBeSentToWaitingIfListIsTooLow(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	pubKeys := generateAddresses(0, 20)

	owner1 := "owner1"
	owner1Stats := &OwnerStats{
		EligibleBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[:4],
			0:                     pubKeys[4:8],
		},
		WaitingBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[8:9],
			0:                     pubKeys[9:10],
		},
		TotalStake: big.NewInt(20 * nodePrice),
	}

	cfg := &InitialNodesConfig{
		MetaConsensusGroupSize:        2,
		ShardConsensusGroupSize:       2,
		MinNumberOfEligibleShardNodes: 4,
		MinNumberOfEligibleMetaNodes:  4,
		NumOfShards:                   1,
		Owners: map[string]*OwnerStats{
			owner1: owner1Stats,
		},
		MaxNodesChangeConfig: []config.MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            20,
				NodesToShufflePerShard: 1,
			},
			{
				EpochEnable:            stakingV4Step3EnableEpoch,
				MaxNumNodes:            18,
				NodesToShufflePerShard: 1,
			},
		},
	}
	node := NewTestMetaProcessorWithCustomNodes(cfg)
	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	// 1. Check initial config is correct
	expectedNodesNum := &configNum{
		eligible: map[uint32]int{
			core.MetachainShardId: 4,
			0:                     4,
		},
		waiting: map[uint32]int{
			core.MetachainShardId: 1,
			0:                     1,
		},
	}
	currNodesConfig := node.NodesConfig
	checkConfig(t, expectedNodesNum, currNodesConfig)

	// Epoch = 0, before staking v4, owner2 stakes 2 nodes
	// - maxNumNodes    = 20
	// - activeNumNodes = 10
	// Newly staked nodes should be sent to new list
	owner2Nodes := pubKeys[12:14]
	node.ProcessStake(t, map[string]*NodesRegisterData{
		"owner2": {
			BLSKeys:    owner2Nodes,
			TotalStake: big.NewInt(2 * nodePrice),
		},
	})
	currNodesConfig = node.NodesConfig
	expectedNodesNum.new = 2
	checkConfig(t, expectedNodesNum, currNodesConfig)
	requireSameSliceDifferentOrder(t, currNodesConfig.new, owner2Nodes)

	// Epoch = 1, staking v4 step 1
	// - maxNumNodes    = 20
	// - activeNumNodes = 12
	// Owner2's new nodes should have been sent to waiting
	node.Process(t, 5)
	currNodesConfig = node.NodesConfig
	expectedNodesNum.new = 0
	expectedNodesNum.waiting[0]++
	expectedNodesNum.waiting[core.MetachainShardId]++
	checkConfig(t, expectedNodesNum, currNodesConfig)
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesConfig.waiting), owner2Nodes, 2)

	// Epoch = 1, before staking v4, owner3 stakes 2 nodes
	// - maxNumNodes    = 20
	// - activeNumNodes = 12
	// Newly staked nodes should be sent to auction list
	owner3Nodes := pubKeys[15:17]
	node.ProcessStake(t, map[string]*NodesRegisterData{
		"owner3": {
			BLSKeys:    owner3Nodes,
			TotalStake: big.NewInt(2 * nodePrice),
		},
	})
	currNodesConfig = node.NodesConfig
	expectedNodesNum.auction = 2
	checkConfig(t, expectedNodesNum, currNodesConfig)
	requireSameSliceDifferentOrder(t, currNodesConfig.auction, owner3Nodes)

	// Epoch = 2, staking v4 step 2
	// - maxNumNodes    = 20
	// - activeNumNodes = 14
	// Owner3's auction nodes should have been sent to waiting
	node.Process(t, 5)
	currNodesConfig = node.NodesConfig
	expectedNodesNum.auction = 0
	expectedNodesNum.waiting[0]++
	expectedNodesNum.waiting[core.MetachainShardId]++
	checkConfig(t, expectedNodesNum, currNodesConfig)
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesConfig.waiting), owner3Nodes, 2)

	// During epochs 2-6, we will have:
	// - activeNodes = 14
	// - maxNumNodes = 18-20
	// Since activeNodes < maxNumNodes, shuffled out nodes will always be sent directly to waiting list, instead of auction
	epoch := uint32(2)
	require.Equal(t, epoch, node.EpochStartTrigger.Epoch())

	numOfShuffledOut := 2
	numRemainingEligible := 6
	numOfUnselectedNodesFromAuction := 0
	numOfSelectedNodesFromAuction := 0

	prevNodesConfig := currNodesConfig
	for epoch < 6 {
		node.Process(t, 5)

		currNodesConfig = node.NodesConfig
		checkConfig(t, expectedNodesNum, currNodesConfig)
		checkShuffledOutNodes(t, currNodesConfig, prevNodesConfig, numOfShuffledOut, numRemainingEligible)
		checkStakingV4EpochChangeFlow(t, currNodesConfig, prevNodesConfig, numOfShuffledOut, numOfUnselectedNodesFromAuction, numOfSelectedNodesFromAuction)

		prevNodesConfig = currNodesConfig
		epoch++
	}
}

func TestStakingV4_LeavingNodesEdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	pubKeys := generateAddresses(0, 20)

	owner1 := "owner1"
	owner1Stats := &OwnerStats{
		EligibleBlsKeys: map[uint32][][]byte{
			core.MetachainShardId: pubKeys[:3],
			0:                     pubKeys[3:6],
			1:                     pubKeys[6:9],
			2:                     pubKeys[9:12],
		},
		TotalStake: big.NewInt(12 * nodePrice),
	}

	cfg := &InitialNodesConfig{
		MetaConsensusGroupSize:        3,
		ShardConsensusGroupSize:       3,
		MinNumberOfEligibleShardNodes: 3,
		MinNumberOfEligibleMetaNodes:  3,
		NumOfShards:                   3,
		Owners: map[string]*OwnerStats{
			owner1: owner1Stats,
		},
		MaxNodesChangeConfig: []config.MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            16,
				NodesToShufflePerShard: 4,
			},
			{
				EpochEnable:            1,
				MaxNumNodes:            18,
				NodesToShufflePerShard: 2,
			},
			{
				EpochEnable:            stakingV4Step3EnableEpoch,
				MaxNumNodes:            12,
				NodesToShufflePerShard: 2,
			},
		},
	}
	node := NewTestMetaProcessorWithCustomNodes(cfg)
	node.EpochStartTrigger.SetRoundsPerEpoch(5)

	// 1. Check initial config is correct
	currNodesConfig := node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesConfig.eligible), 12)
	require.Len(t, getAllPubKeys(currNodesConfig.waiting), 0)
	require.Len(t, currNodesConfig.eligible[core.MetachainShardId], 3)
	require.Len(t, currNodesConfig.waiting[core.MetachainShardId], 0)
	require.Len(t, currNodesConfig.eligible[0], 3)
	require.Len(t, currNodesConfig.waiting[0], 0)
	require.Len(t, currNodesConfig.eligible[1], 3)
	require.Len(t, currNodesConfig.waiting[1], 0)
	require.Len(t, currNodesConfig.eligible[2], 3)
	require.Len(t, currNodesConfig.waiting[2], 0)
	require.Empty(t, currNodesConfig.shuffledOut)
	require.Empty(t, currNodesConfig.auction)

	// NewOwner0 stakes 1 node with top up = 0 before staking v4; should be sent to new nodes, since there are enough slots
	newOwner0 := "newOwner0"
	newOwner0BlsKeys := [][]byte{generateAddress(101)}
	node.ProcessStake(t, map[string]*NodesRegisterData{
		newOwner0: {
			BLSKeys:    newOwner0BlsKeys,
			TotalStake: big.NewInt(nodePrice),
		},
	})
	currNodesConfig = node.NodesConfig
	requireSameSliceDifferentOrder(t, currNodesConfig.new, newOwner0BlsKeys)

	// UnStake one of the initial nodes
	node.ProcessUnStake(t, map[string][][]byte{
		owner1: {owner1Stats.EligibleBlsKeys[core.MetachainShardId][0]},
	})

	// Fast-forward few epochs such that the whole staking v4 is activated.
	// We should have same 12 initial nodes + 1 extra node (because of legacy code where all leaving nodes were
	// considered to be eligible and the unStaked node was forced to remain eligible)
	node.Process(t, 49)
	currNodesConfig = node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesConfig.eligible), 12)
	require.Len(t, getAllPubKeys(currNodesConfig.waiting), 1)

	// Stake 10 extra nodes and check that they are sent to auction
	newOwner1 := "newOwner1"
	newOwner1BlsKeys := generateAddresses(303, 10)
	node.ProcessStake(t, map[string]*NodesRegisterData{
		newOwner1: {
			BLSKeys:    newOwner1BlsKeys,
			TotalStake: big.NewInt(nodePrice * 10),
		},
	})
	currNodesConfig = node.NodesConfig
	requireSameSliceDifferentOrder(t, currNodesConfig.auction, newOwner1BlsKeys)

	// After 2 epochs, unStake all previously staked keys. Some of them have been already sent to eligible/waiting, but most
	// of them are still in auction. UnStaked nodes' status from auction should be: leaving now, but their previous list was auction.
	// We should not force his auction nodes as being eligible in the next epoch. We should only force his existing active
	// nodes to remain in the system.
	node.Process(t, 10)
	currNodesConfig = node.NodesConfig
	newOwner1AuctionNodes := getIntersection(currNodesConfig.auction, newOwner1BlsKeys)
	newOwner1EligibleNodes := getIntersection(getAllPubKeys(currNodesConfig.eligible), newOwner1BlsKeys)
	newOwner1WaitingNodes := getIntersection(getAllPubKeys(currNodesConfig.waiting), newOwner1BlsKeys)
	newOwner1ActiveNodes := append(newOwner1EligibleNodes, newOwner1WaitingNodes...)
	require.Equal(t, len(newOwner1AuctionNodes)+len(newOwner1ActiveNodes), len(newOwner1BlsKeys)) // sanity check

	node.ClearStoredMbs()
	node.ProcessUnStake(t, map[string][][]byte{
		newOwner1: newOwner1BlsKeys,
	})

	node.Process(t, 5)
	currNodesConfig = node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesConfig.eligible), 12)
	requireMapContains(t, currNodesConfig.leaving, newOwner1AuctionNodes)
	requireMapDoesNotContain(t, currNodesConfig.eligible, newOwner1AuctionNodes)
	requireMapDoesNotContain(t, currNodesConfig.waiting, newOwner1AuctionNodes)

	allCurrentActiveNodes := append(getAllPubKeys(currNodesConfig.eligible), getAllPubKeys(currNodesConfig.waiting)...)
	owner1NodesThatAreStillForcedToRemain := getIntersection(allCurrentActiveNodes, newOwner1ActiveNodes)
	require.NotZero(t, len(owner1NodesThatAreStillForcedToRemain))

	// Fast-forward some epochs, no error should occur, and we should have our initial config of:
	// - 12 eligible nodes
	// - 1 waiting list
	// - some forced nodes to remain from newOwner1
	node.Process(t, 10)
	currNodesConfig = node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesConfig.eligible), 12)
	require.Len(t, getAllPubKeys(currNodesConfig.waiting), 1)
	allCurrentActiveNodes = append(getAllPubKeys(currNodesConfig.eligible), getAllPubKeys(currNodesConfig.waiting)...)
	owner1NodesThatAreStillForcedToRemain = getIntersection(allCurrentActiveNodes, newOwner1ActiveNodes)
	require.NotZero(t, len(owner1NodesThatAreStillForcedToRemain))

	// Stake 10 extra nodes such that the forced eligible nodes from previous newOwner1 can leave the system
	// and are replaced by new nodes
	newOwner2 := "newOwner2"
	newOwner2BlsKeys := generateAddresses(403, 10)
	node.ProcessStake(t, map[string]*NodesRegisterData{
		newOwner2: {
			BLSKeys:    newOwner2BlsKeys,
			TotalStake: big.NewInt(nodePrice * 10),
		},
	})
	currNodesConfig = node.NodesConfig
	requireSliceContains(t, currNodesConfig.auction, newOwner2BlsKeys)

	// Fast-forward multiple epochs and check that newOwner1's forced nodes from previous epochs left
	node.Process(t, 20)
	currNodesConfig = node.NodesConfig
	allCurrentNodesInSystem := getAllPubKeysFromConfig(currNodesConfig)
	owner1LeftNodes := getIntersection(owner1NodesThatAreStillForcedToRemain, allCurrentNodesInSystem)
	require.Zero(t, len(owner1LeftNodes))
}

func TestStakingV4LeavingNodesShouldDistributeToWaitingOnlyNecessaryNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfMetaNodes := uint32(400)
	numOfShards := uint32(3)
	numOfEligibleNodesPerShard := uint32(400)
	numOfWaitingNodesPerShard := uint32(400)
	numOfNodesToShufflePerShard := uint32(80)
	shardConsensusGroupSize := 266
	metaConsensusGroupSize := 266
	numOfNodesInStakingQueue := uint32(80)

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
	nodesConfigStakingV4Step1 := node.NodesConfig
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Step1.eligible), totalEligible)
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Step1.waiting), totalWaiting)
	require.Empty(t, nodesConfigStakingV4Step1.queue)
	require.Empty(t, nodesConfigStakingV4Step1.shuffledOut)
	require.Empty(t, nodesConfigStakingV4Step1.auction) // the queue should be empty

	// 3. re-stake the node nodes that were in the queue
	node.ProcessReStake(t, initialNodes.queue)
	nodesConfigStakingV4Step1 = node.NodesConfig
	requireSameSliceDifferentOrder(t, initialNodes.queue, nodesConfigStakingV4Step1.auction)

	// Reach step 3 and check normal flow
	node.Process(t, 10)
	epochs := 0
	prevConfig := node.NodesConfig
	numOfSelectedNodesFromAuction := 320  // 320, since we will always fill shuffled out nodes with this config
	numOfUnselectedNodesFromAuction := 80 // 80 = 400 from queue - 320
	numOfShuffledOut := 80 * 4            // 80 per shard + meta
	for epochs < 3 {
		node.Process(t, 5)
		newNodeConfig := node.NodesConfig

		require.Len(t, getAllPubKeys(newNodeConfig.eligible), totalEligible) // 1600
		require.Len(t, getAllPubKeys(newNodeConfig.waiting), 1280)           // 1280
		require.Len(t, getAllPubKeys(newNodeConfig.shuffledOut), 320)        // 320
		require.Len(t, newNodeConfig.auction, 400)                           // 400
		require.Empty(t, newNodeConfig.queue)
		require.Empty(t, newNodeConfig.leaving)

		checkStakingV4EpochChangeFlow(t, newNodeConfig, prevConfig, numOfShuffledOut, numOfUnselectedNodesFromAuction, numOfSelectedNodesFromAuction)
		prevConfig = newNodeConfig
		epochs++
	}

	// UnStake:
	// - 46 from waiting + eligible ( 13 waiting + 36 eligible)
	// - 11 from auction
	currNodesCfg := node.NodesConfig
	nodesToUnStakeFromAuction := currNodesCfg.auction[:11]

	nodesToUnStakeFromWaiting := append(currNodesCfg.waiting[0][:3], currNodesCfg.waiting[1][:3]...)
	nodesToUnStakeFromWaiting = append(nodesToUnStakeFromWaiting, currNodesCfg.waiting[2][:3]...)
	nodesToUnStakeFromWaiting = append(nodesToUnStakeFromWaiting, currNodesCfg.waiting[core.MetachainShardId][:4]...)

	nodesToUnStakeFromEligible := append(currNodesCfg.eligible[0][:8], currNodesCfg.eligible[1][:8]...)
	nodesToUnStakeFromEligible = append(nodesToUnStakeFromEligible, currNodesCfg.eligible[2][:8]...)
	nodesToUnStakeFromEligible = append(nodesToUnStakeFromEligible, currNodesCfg.eligible[core.MetachainShardId][:9]...)

	nodesToUnStake := getAllOwnerNodesMap(nodesToUnStakeFromAuction, nodesToUnStakeFromWaiting, nodesToUnStakeFromEligible)

	prevConfig = currNodesCfg
	node.ProcessUnStake(t, nodesToUnStake)
	node.Process(t, 5)
	currNodesCfg = node.NodesConfig

	require.Len(t, getAllPubKeys(currNodesCfg.leaving), 57)                                            // 11 auction + 46 active (13 waiting + 36 eligible)
	require.Len(t, getAllPubKeys(currNodesCfg.shuffledOut), 274)                                       // 320 - 46 active
	require.Len(t, currNodesCfg.auction, 343)                                                          // 400 initial - 57 leaving
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesCfg.waiting), prevConfig.auction, 320) // 320 selected
	requireSliceContainsNumOfElements(t, currNodesCfg.auction, prevConfig.auction, 69)                 // 69 unselected

	nodesToUnStakeFromAuction = make([][]byte, 0)
	nodesToUnStakeFromWaiting = make([][]byte, 0)
	nodesToUnStakeFromEligible = make([][]byte, 0)

	prevConfig = currNodesCfg
	// UnStake:
	// - 224 from waiting + eligible ( 13 waiting + 36 eligible), but unbalanced:
	//    -> unStake 100 from waiting shard=meta
	//	  -> unStake 90 from eligible shard=2
	// - 11 from auction
	nodesToUnStakeFromAuction = currNodesCfg.auction[:11]
	nodesToUnStakeFromWaiting = append(currNodesCfg.waiting[0][:3], currNodesCfg.waiting[1][:3]...)
	nodesToUnStakeFromWaiting = append(nodesToUnStakeFromWaiting, currNodesCfg.waiting[2][:3]...)
	nodesToUnStakeFromWaiting = append(nodesToUnStakeFromWaiting, currNodesCfg.waiting[core.MetachainShardId][:100]...)

	nodesToUnStakeFromEligible = append(currNodesCfg.eligible[0][:8], currNodesCfg.eligible[1][:8]...)
	nodesToUnStakeFromEligible = append(nodesToUnStakeFromEligible, currNodesCfg.eligible[2][:90]...)
	nodesToUnStakeFromEligible = append(nodesToUnStakeFromEligible, currNodesCfg.eligible[core.MetachainShardId][:9]...)

	nodesToUnStake = getAllOwnerNodesMap(nodesToUnStakeFromAuction, nodesToUnStakeFromWaiting, nodesToUnStakeFromEligible)
	node.ProcessUnStake(t, nodesToUnStake)
	node.Process(t, 4)
	currNodesCfg = node.NodesConfig

	// Leaving:
	// - 11 auction
	// - shard 0 = 11
	// - shard 1 = 11
	// - shard 2 = 80 (there were 93 unStakes, but only 80 will be leaving, rest 13 will be forced to stay)
	// - shard meta = 80 (there were 109 unStakes, but only 80 will be leaving, rest 29 will be forced to stay)
	// Therefore we will have in total actually leaving = 193 (11 + 11 + 11 + 80 + 80)
	// We should see a log in selector like this:
	// auctionListSelector.SelectNodesFromAuctionList max nodes = 2880 current number of validators = 2656 num of nodes which will be shuffled out = 138 num forced to stay = 42 num of validators after shuffling = 2518 auction list size = 332 available slots (2880 - 2560) = 320
	require.Len(t, getAllPubKeys(currNodesCfg.leaving), 193)
	require.Len(t, getAllPubKeys(currNodesCfg.shuffledOut), 138)                                       //    69 from shard0 + shard from shard1, rest will not be shuffled
	require.Len(t, currNodesCfg.auction, 150)                                                          // 138 shuffled out + 12 unselected
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesCfg.waiting), prevConfig.auction, 320) // 320 selected
	requireSliceContainsNumOfElements(t, currNodesCfg.auction, prevConfig.auction, 12)                 // 12 unselected
}

func TestStakingV4MoreLeavingNodesThanToShufflePerShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfMetaNodes := uint32(400)
	numOfShards := uint32(3)
	numOfEligibleNodesPerShard := uint32(400)
	numOfWaitingNodesPerShard := uint32(400)
	numOfNodesToShufflePerShard := uint32(80)
	shardConsensusGroupSize := 266
	metaConsensusGroupSize := 266
	numOfNodesInStakingQueue := uint32(80)

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
	nodesConfigStakingV4Step1 := node.NodesConfig
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Step1.eligible), totalEligible)
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Step1.waiting), totalWaiting)
	require.Empty(t, nodesConfigStakingV4Step1.queue)
	require.Empty(t, nodesConfigStakingV4Step1.shuffledOut)
	require.Empty(t, nodesConfigStakingV4Step1.auction) // the queue should be empty

	// 3. re-stake the node nodes that were in the queue
	node.ProcessReStake(t, initialNodes.queue)
	nodesConfigStakingV4Step1 = node.NodesConfig
	requireSameSliceDifferentOrder(t, initialNodes.queue, nodesConfigStakingV4Step1.auction)

	// Reach step 3
	node.Process(t, 10)

	// UnStake 100 nodes from each shard:
	// - shard 0: 100 waiting
	// - shard 1: 50 waiting + 50 eligible
	// - shard 2: 20 waiting + 80 eligible
	// - shard meta: 100 eligible
	currNodesCfg := node.NodesConfig

	nodesToUnStakeFromWaiting := currNodesCfg.waiting[0][:100]
	nodesToUnStakeFromWaiting = append(nodesToUnStakeFromWaiting, currNodesCfg.waiting[1][:50]...)
	nodesToUnStakeFromWaiting = append(nodesToUnStakeFromWaiting, currNodesCfg.waiting[2][:20]...)

	nodesToUnStakeFromEligible := currNodesCfg.eligible[1][:50]
	nodesToUnStakeFromEligible = append(nodesToUnStakeFromEligible, currNodesCfg.eligible[2][:80]...)
	nodesToUnStakeFromEligible = append(nodesToUnStakeFromEligible, currNodesCfg.eligible[core.MetachainShardId][:100]...)

	nodesToUnStake := getAllOwnerNodesMap(nodesToUnStakeFromWaiting, nodesToUnStakeFromEligible)

	prevConfig := currNodesCfg
	node.ProcessUnStake(t, nodesToUnStake)
	node.Process(t, 4)
	currNodesCfg = node.NodesConfig

	require.Len(t, getAllPubKeys(currNodesCfg.leaving), 320)                                           // we unStaked 400, but only allowed 320 to leave
	require.Len(t, getAllPubKeys(currNodesCfg.shuffledOut), 0)                                         // no shuffled out, since 80 per shard were leaving
	require.Len(t, currNodesCfg.auction, 80)                                                           // 400 initial - 320 selected
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesCfg.waiting), prevConfig.auction, 320) // 320 selected
	requireSliceContainsNumOfElements(t, currNodesCfg.auction, prevConfig.auction, 80)                 // 80 unselected

	// Add 400 new nodes in the system and fast-forward
	node.ProcessStake(t, map[string]*NodesRegisterData{
		"ownerX": {
			BLSKeys:    generateAddresses(99999, 400),
			TotalStake: big.NewInt(nodePrice * 400),
		},
	})
	node.Process(t, 10)

	// UnStake exactly 80 nodes
	prevConfig = node.NodesConfig
	nodesToUnStake = getAllOwnerNodesMap(node.NodesConfig.eligible[1][:80])
	node.ProcessUnStake(t, nodesToUnStake)
	node.Process(t, 4)

	currNodesCfg = node.NodesConfig
	require.Len(t, getAllPubKeys(currNodesCfg.leaving), 80)                                            // 320 - 80 leaving
	require.Len(t, getAllPubKeys(currNodesCfg.shuffledOut), 240)                                       // 240 shuffled out
	requireSliceContainsNumOfElements(t, getAllPubKeys(currNodesCfg.waiting), prevConfig.auction, 320) // 320 selected
	requireSliceContainsNumOfElements(t, currNodesCfg.auction, prevConfig.auction, 80)                 // 80 unselected
}
