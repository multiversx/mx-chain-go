package sharding

import (
	"bytes"
	"sort"
	"sync"
)

var _ NodesShuffler = (*randXORShuffler)(nil)

// TODO: Decide if transaction load statistics will be used for limiting the number of shards
type randXORShuffler struct {
	// TODO: remove the references to this constant when reinitialization of node in new shard is implemented
	shuffleBetweenShards bool

	adaptivity        bool
	nodesShard        uint32
	nodesMeta         uint32
	shardHysteresis   uint32
	metaHysteresis    uint32
	mutShufflerParams sync.RWMutex
}

// NewXorValidatorsShuffler creates a validator shuffler that uses a XOR between validator key and a given
// random number to do the shuffling
func NewXorValidatorsShuffler(
	nodesShard uint32,
	nodesMeta uint32,
	hysteresis float32,
	adaptivity bool,
	shuffleBetweenShards bool,
) *randXORShuffler {
	log.Debug("Shuffler created", "shuffleBetweenShards", shuffleBetweenShards)
	rxs := &randXORShuffler{shuffleBetweenShards: shuffleBetweenShards}

	rxs.UpdateParams(nodesShard, nodesMeta, hysteresis, adaptivity)

	return rxs
}

// UpdateParams updates the shuffler parameters
// Should be called when new params are agreed through governance
func (rxs *randXORShuffler) UpdateParams(
	nodesShard uint32,
	nodesMeta uint32,
	hysteresis float32,
	adaptivity bool,
) {
	// TODO: are there constraints we want to enforce? e.g min/max hysteresis
	shardHysteresis := uint32(float32(nodesShard) * hysteresis)
	metaHysteresis := uint32(float32(nodesMeta) * hysteresis)

	rxs.mutShufflerParams.Lock()
	rxs.shardHysteresis = shardHysteresis
	rxs.metaHysteresis = metaHysteresis
	rxs.nodesShard = nodesShard
	rxs.nodesMeta = nodesMeta
	rxs.adaptivity = adaptivity
	rxs.mutShufflerParams.Unlock()
}

// UpdateNodeLists shuffles the nodes and returns the lists with the new nodes configuration
// The function needs to ensure that:
//      1.  Old eligible nodes list will have up to shuffleOutThreshold percent nodes shuffled out from each shard
//      2.  The leaving nodes are checked against the eligible nodes and waiting nodes and removed if present from the
//          pools and leaving nodes list (if remaining nodes can still sustain the shard)
//      3.  shuffledOutNodes = oldEligibleNodes + waitingListNodes - minNbNodesPerShard (for each shard)
//      4.  Old waiting nodes list for each shard will be added to the remaining eligible nodes list
//      5.  The new nodes are equally distributed among the existing shards into waiting lists
//      6.  The shuffled out nodes are distributed among the existing shards into waiting lists.
//          We may have three situations:
//          a)  In case (shuffled out nodes + new nodes) > (nbShards * perShardHysteresis + minNodesPerShard) then
//              we need to prepare for a split event, so a higher percentage of nodes need to be directed to the shard
//              that will be split.
//          b)  In case (shuffled out nodes + new nodes) < (nbShards * perShardHysteresis) then we can immediately
//              execute the shard merge
//          c)  No change in the number of shards then nothing extra needs to be done
func (rxs *randXORShuffler) UpdateNodeLists(args ArgsUpdateNodes) ResUpdateNodes {
	eligibleAfterReshard := copyValidatorMap(args.Eligible)
	waitingAfterReshard := copyValidatorMap(args.Waiting)

	removeDupplicates(args.UnstakeLeaving, args.AdditionalLeaving)
	totalLeavingNum := len(args.AdditionalLeaving) + len(args.UnstakeLeaving)

	newNbShards := rxs.computeNewShards(
		args.Eligible,
		args.Waiting,
		len(args.NewNodes),
		totalLeavingNum,
		args.NbShards,
	)

	rxs.mutShufflerParams.RLock()
	canSplit := rxs.adaptivity && newNbShards > args.NbShards
	canMerge := rxs.adaptivity && newNbShards < args.NbShards
	rxs.mutShufflerParams.RUnlock()

	if canSplit {
		eligibleAfterReshard, waitingAfterReshard = rxs.splitShards(args.Eligible, args.Waiting, newNbShards)
	}
	if canMerge {
		eligibleAfterReshard, waitingAfterReshard = rxs.mergeShards(args.Eligible, args.Waiting, newNbShards)
	}

	// TODO: remove the conditional and the else branch when reinitialization of node in new shard is implemented
	var validatorsDistributor ValidatorsDistributor = nil
	if rxs.shuffleBetweenShards {
		validatorsDistributor = &CrossShardValidatorDistributor{}
	} else {
		validatorsDistributor = &IntraShardValidatorDistributor{}
	}

	return shuffleNodes(
		eligibleAfterReshard,
		waitingAfterReshard,
		args.UnstakeLeaving,
		args.AdditionalLeaving,
		args.NewNodes,
		args.Rand,
		validatorsDistributor,
	)
}

func removeDupplicates(unstake []Validator, additionalLeaving []Validator) {
	for _, validator := range unstake {
		for i, additionalLeavingValidator := range additionalLeaving {
			if bytes.Equal(validator.PubKey(), additionalLeavingValidator.PubKey()) {
				removeValidatorFromList(additionalLeaving, i)
			}
		}
	}
}

func removeNodesFromMap(
	existingNodes map[uint32][]Validator,
	leavingNodes []Validator,
	numToRemove map[uint32]int,
) (map[uint32][]Validator, []Validator) {
	sortedShardIds := sortKeys(existingNodes)
	numRemoved := 0

	for _, shardId := range sortedShardIds {
		numToRemoveOnShard := numToRemove[shardId]
		leavingNodes, numRemoved = removeNodesFromShard(existingNodes, leavingNodes, shardId, numToRemoveOnShard)
		numToRemove[shardId] -= numRemoved
	}

	return existingNodes, leavingNodes
}

func removeNodesFromShard(existingNodes map[uint32][]Validator, leavingNodes []Validator, shard uint32, nbToRemove int) ([]Validator, int) {
	vList := existingNodes[shard]
	if len(leavingNodes) < nbToRemove {
		nbToRemove = len(leavingNodes)
	}

	vList, removedNodes := removeValidatorsFromList(existingNodes[shard], leavingNodes, nbToRemove)
	leavingNodes, _ = removeValidatorsFromList(leavingNodes, removedNodes, len(removedNodes))
	existingNodes[shard] = vList
	return leavingNodes, len(removedNodes)
}

// IsInterfaceNil verifies if the underlying object is nil
func (rxs *randXORShuffler) IsInterfaceNil() bool {
	return rxs == nil
}

func shuffleNodes(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	unstakeLeaving []Validator,
	additionalLeaving []Validator,
	newNodes []Validator,
	randomness []byte,
	distributor ValidatorsDistributor,
) ResUpdateNodes {
	shuffledOutMap := make(map[uint32][]Validator)

	allLeaving := append(unstakeLeaving, additionalLeaving...)

	waitingCopy := copyValidatorMap(waiting)
	eligibleCopy := copyValidatorMap(eligible)

	numToRemove := make(map[uint32]int)
	for shardId := range waiting {
		numToRemove[shardId] = len(waiting[shardId])
	}

	removeLeavingNodesNotExistingInEligibleOrWaiting(unstakeLeaving, additionalLeaving, waitingCopy, eligibleCopy)

	newEligible, newWaiting, stillRemainingInLeaving := removeLeavingNodesFromValidatorMaps(eligibleCopy, waitingCopy, numToRemove, unstakeLeaving)
	newEligible, newWaiting, stillRemainingInLeaving = removeLeavingNodesFromValidatorMaps(newEligible, newWaiting, numToRemove, additionalLeaving)

	sortedShardIds := sortKeys(eligible)
	for _, shardId := range sortedShardIds {
		validators := newEligible[shardId]
		numToShuffle := numToRemove[shardId]
		shuffledOutMap[shardId], newEligible[shardId] = shuffleOutShard(
			validators,
			numToShuffle,
			randomness,
		)
	}

	moveNodesToMap(newEligible, newWaiting)
	err := distributeValidators(newNodes, newWaiting, randomness)
	if err != nil {
		log.Warn("distributeValidators newNodes failed", "error", err)
	}

	err = distributor.DistributeValidators(newWaiting, shuffledOutMap, randomness)
	if err != nil {
		log.Warn("distributeValidators newNodes failed", "error", err)
	}

	actualLeaving, _ := removeValidatorsFromList(allLeaving, stillRemainingInLeaving, len(stillRemainingInLeaving))

	return ResUpdateNodes{
		Eligible:       newEligible,
		Waiting:        newWaiting,
		Leaving:        actualLeaving,
		StillRemaining: stillRemainingInLeaving,
	}
}

func removeLeavingNodesNotExistingInEligibleOrWaiting(unstakeLeaving []Validator, additionalLeaving []Validator, waitingCopy map[uint32][]Validator, eligibleCopy map[uint32][]Validator) {
	removeValidatorsNotInMap(unstakeLeaving, waitingCopy)
	removeValidatorsNotInMap(unstakeLeaving, eligibleCopy)
	removeValidatorsNotInMap(additionalLeaving, waitingCopy)
	removeValidatorsNotInMap(additionalLeaving, eligibleCopy)
}

func removeLeavingNodesFromValidatorMaps(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	numToRemove map[uint32]int,
	leaving []Validator,
) (map[uint32][]Validator, map[uint32][]Validator, []Validator) {

	stillRemainingInLeaving := make([]Validator, len(leaving))
	copy(stillRemainingInLeaving, leaving)

	waiting, stillRemainingInLeaving = removeNodesFromMap(waiting, stillRemainingInLeaving, numToRemove)
	eligible, stillRemainingInLeaving = removeNodesFromMap(eligible, stillRemainingInLeaving, numToRemove)
	return eligible, waiting, stillRemainingInLeaving
}

func removeValidatorsNotInMap(validators []Validator, validatorsMap map[uint32][]Validator) {
	for i, v := range validators {
		if !foundInMap(validatorsMap, v) {
			removeValidatorFromList(validators, i)
		}
	}
}

func foundInMap(nodesMap map[uint32][]Validator, validator Validator) bool {
	for _, shardValidators := range nodesMap {
		for _, v := range shardValidators {
			if bytes.Equal(validator.PubKey(), v.PubKey()) {
				return true
			}
		}
	}

	return false
}

// computeNewShards determines the new number of shards based on the number of nodes in the network
func (rxs *randXORShuffler) computeNewShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	numNewNodes int,
	numLeavingNodes int,
	nbShards uint32,
) uint32 {

	nbEligible := 0
	nbWaiting := 0
	for shard := range eligible {
		nbEligible += len(eligible[shard])
		nbWaiting += len(waiting[shard])
	}

	nodesNewEpoch := uint32(nbEligible + nbWaiting + numNewNodes - numLeavingNodes)

	rxs.mutShufflerParams.RLock()
	maxNodesMeta := rxs.nodesMeta + rxs.metaHysteresis
	maxNodesShard := rxs.nodesShard + rxs.shardHysteresis
	nodesForSplit := (nbShards+1)*maxNodesShard + maxNodesMeta
	nodesForMerge := nbShards*rxs.nodesShard + rxs.nodesMeta
	rxs.mutShufflerParams.RUnlock()

	nbShardsNew := nbShards
	if nodesNewEpoch > nodesForSplit {
		nbNodesWithoutMaxMeta := nodesNewEpoch - maxNodesMeta
		nbShardsNew = nbNodesWithoutMaxMeta / maxNodesShard

		return nbShardsNew
	}

	if nodesNewEpoch < nodesForMerge {
		return nbShardsNew - 1
	}

	return nbShardsNew
}

// shuffleOutNodes shuffles the list of eligible validators in each shard and returns the array of shuffled out
// validators
func shuffleOutNodes(
	eligible map[uint32][]Validator,
	numToShuffle map[uint32]int,
	randomness []byte,
) ([]Validator, map[uint32][]Validator) {
	shuffledOut := make([]Validator, 0)
	newEligible := make(map[uint32][]Validator)
	var shardShuffledOut []Validator

	sortedShardIds := sortKeys(eligible)
	for _, shardId := range sortedShardIds {
		validators := eligible[shardId]
		shardShuffledOut, validators = shuffleOutShard(validators, numToShuffle[shardId], randomness)
		shuffledOut = append(shuffledOut, shardShuffledOut...)

		newEligible[shardId], _ = removeValidatorsFromList(validators, shardShuffledOut, len(shardShuffledOut))
	}

	return shuffledOut, newEligible
}

// shuffleOutShard selects the validators to be shuffled out from a shard
func shuffleOutShard(
	validators []Validator,
	validatorsToSelect int,
	randomness []byte,
) ([]Validator, []Validator) {

	shardShuffledEligible := shuffleList(validators, randomness)
	shardShuffledOut := shardShuffledEligible[:validatorsToSelect]
	remainingEligible := shardShuffledEligible[validatorsToSelect:]

	return shardShuffledOut, remainingEligible
}

// shuffleList returns a shuffled list of validators.
// The shuffling is done based by xor-ing the randomness with the
// public keys of validators and sorting the validators depending on
// the xor result.
func shuffleList(validators []Validator, randomness []byte) []Validator {
	keys := make([]string, len(validators))
	mapValidators := make(map[string]Validator)

	for i, v := range validators {
		keys[i] = string(xorBytes(v.PubKey(), randomness))
		mapValidators[keys[i]] = v
	}

	sort.Strings(keys)

	result := make([]Validator, len(validators))
	for i := 0; i < len(validators); i++ {
		result[i] = mapValidators[keys[i]]
	}

	return result
}

func removeValidatorsFromList(
	validatorList []Validator,
	validatorsToRemove []Validator,
	maxToRemove int,
) ([]Validator, []Validator) {
	resultedList := make([]Validator, 0)
	resultedList = append(resultedList, validatorList...)
	removed := make([]Validator, 0)

	for _, valToRemove := range validatorsToRemove {
		if len(removed) == maxToRemove {
			break
		}

		for index, val := range resultedList {
			if bytes.Equal(val.PubKey(), valToRemove.PubKey()) {
				resultedList = removeValidatorFromList(resultedList, index)
				removed = append(removed, val)
				break
			}
		}
	}

	return resultedList, removed
}

// removeValidatorFromList replaces the element at given index with the last element in the slice and returns a slice
// with a decremented length.The order in the list is important as long as it is kept the same for all validators,
// so not critical to maintain the original order inside the list, as that would be slower.
//
// Attention: The slice given as parameter will have its element on position index swapped with the last element
func removeValidatorFromList(validatorList []Validator, index int) []Validator {
	indexNotOK := index > len(validatorList)-1 || index < 0

	if indexNotOK {
		return validatorList
	}

	validatorList[index] = validatorList[len(validatorList)-1]
	return validatorList[:len(validatorList)-1]
}

func removeValidatorFromListKeepOrder(validatorList []Validator, index int) []Validator {
	indexNotOK := index > len(validatorList)-1 || index < 0

	if indexNotOK {
		return validatorList
	}

	return append(validatorList[:index], validatorList[index+1:]...)
}

// xorBytes XORs two byte arrays up to the shortest length of the two, and returns the resulted XORed bytes.
func xorBytes(a []byte, b []byte) []byte {
	lenA := len(a)
	lenB := len(b)
	minLen := lenA

	if lenB < minLen {
		minLen = lenB
	}

	result := make([]byte, minLen)
	for i := 0; i < minLen; i++ {
		result[i] = a[i] ^ b[i]
	}

	return result
}

// splitShards prepares for the shards split, or if already prepared does the split returning the resulting
// shards configuration for eligible and waiting lists
func (rxs *randXORShuffler) splitShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	_ uint32,
) (map[uint32][]Validator, map[uint32][]Validator) {
	log.Error(ErrNotImplemented.Error())

	// TODO: do the split
	return copyValidatorMap(eligible), copyValidatorMap(waiting)
}

// mergeShards merges the required shards, returning the resulting shards configuration for eligible and waiting lists
func (rxs *randXORShuffler) mergeShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	_ uint32,
) (map[uint32][]Validator, map[uint32][]Validator) {
	log.Error(ErrNotImplemented.Error())

	// TODO: do the merge
	return copyValidatorMap(eligible), copyValidatorMap(waiting)
}

// copyValidatorMap creates a copy for the Validators map, creating copies for each of the lists for each shard
func copyValidatorMap(validators map[uint32][]Validator) map[uint32][]Validator {
	result := make(map[uint32][]Validator)

	for k, v := range validators {
		elems := make([]Validator, 0)
		result[k] = append(elems, v...)
	}

	return result
}

// moveNodesToMap moves the validators in the waiting list to corresponding eligible list
func moveNodesToMap(destination map[uint32][]Validator, source map[uint32][]Validator) {
	for k, v := range source {
		destination[k] = append(destination[k], v...)
		source[k] = make([]Validator, 0)
	}
}

// distributeNewNodes distributes a list of validators to the given validators map
func distributeValidators(validators []Validator, destLists map[uint32][]Validator, randomness []byte) error {
	if destLists == nil {
		return ErrNilDestinationForDistribute
	}

	// if there was a split or a merge, eligible map should already have a different nb of keys (shards)
	shuffledValidators := shuffleList(validators, randomness)
	var shardId uint32

	sortedShardIds := sortKeys(destLists)
	destLength := uint32(len(sortedShardIds))

	for i, v := range shuffledValidators {
		shardId = sortedShardIds[uint32(i)%destLength]
		destLists[shardId] = append(destLists[shardId], v)
	}

	return nil
}

func sortKeys(nodes map[uint32][]Validator) []uint32 {
	keys := make([]uint32, 0, len(nodes))
	for k := range nodes {
		keys = append(keys, k)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}
