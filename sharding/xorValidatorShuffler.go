package sharding

import (
	"sort"
	"sync"
)

// TODO: Decide if transaction load statistics will be used for limiting the number of shards
type randXORShuffler struct {
	nodesShard        uint32
	nodesMeta         uint32
	shardHysteresis   uint32
	metaHysteresis    uint32
	adaptivity        bool
	mutShufflerParams sync.Mutex
}

// NewXorValidatorsShuffler creates a validator shuffler that uses a XOR between validator key and a given
// random number to do the shuffling
func NewXorValidatorsShuffler(
	nodesShard uint32,
	nodesMeta uint32,
	hysteresis float32,
	adaptivity bool,
) *randXORShuffler {
	rxs := &randXORShuffler{}

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
//      2.  shuffledOutNodes = oldEligibleNodes + waitingListNodes - minNbNodesPerShard (for each shard)
//      3.  Old waiting nodes list for each shard will be added to the remaining eligible nodes list
//      4.  The leaving nodes are checked against the new nodes and shuffled out nodes and removed if present from the
//          pools and leaving nodes list
//      5.  The remaining new nodes are equally distributed among the existing shards into waiting lists
//      6.  The remaining shuffled out nodes are distributed among the existing shards into waiting lists.
//          We may have three situations:
//          a)  In case (shuffled out nodes + new nodes) > (nbShards * perShardHysteresis + minNodesPerShard) then
//              we need to prepare for a split event, so a higher percentage of nodes need to be directed to the shard
//              that will be split.
//          b)  In case (shuffled out nodes + new nodes) < (nbShards * perShardHysteresis) then we need to prepare for a
//              shard merge event so
//          c)  No change in the number of shards then nothing extra needs to be done
func (rxs *randXORShuffler) UpdateNodeLists(args UpdateNodesArgs) (map[uint32][]Validator, map[uint32][]Validator) {
	var shuffledOutNodes []Validator
	var newEligible map[uint32][]Validator
	var newWaiting map[uint32][]Validator

	newNbShards := rxs.computeNewShards(args.eligible, args.waiting, args.new, args.leaving, args.nbShards)
	canSplit := rxs.adaptivity && newNbShards > args.nbShards

	if canSplit {
		newEligible, newWaiting = rxs.splitShards(args.eligible, args.waiting, newNbShards)
	}

	canMerge := rxs.adaptivity && newNbShards < args.nbShards
	if canMerge {
		newEligible, newWaiting = rxs.mergeShards(args.eligible, args.waiting, newNbShards)
	}

	for shard, vList := range newWaiting {
		nbToRemove := len(vList)
		if len(args.leaving) < nbToRemove {
			nbToRemove = len(args.leaving)
		}

		vList, args.leaving = removeValidatorsFromList(vList, args.leaving, nbToRemove)
		newWaiting[shard] = vList
	}

	shuffledOutNodes, newEligible, args.leaving = shuffleOutNodes(args.eligible, args.waiting, args.leaving, args.rand)
	promoteWaitingToEligible(newEligible, args.waiting)
	distributeValidators(args.new, newWaiting, args.rand, newNbShards+1)
	distributeValidators(shuffledOutNodes, newWaiting, args.rand, newNbShards+1)

	return newEligible, newWaiting
}

// computeNewShards determines the new number of shards based on the number of nodes in the network
func (rxs *randXORShuffler) computeNewShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	newNodes []Validator,
	leavingNodes []Validator,
	nbShards uint32,
) uint32 {

	nbEligible := 0
	nbWaiting := 0
	for shard := range eligible {
		nbEligible += len(eligible[shard])
		nbWaiting += len(waiting[shard])
	}

	nodesNewEpoch := uint32(nbEligible + nbWaiting + len(newNodes) - len(leavingNodes))
	maxNodesMeta := rxs.nodesMeta + rxs.metaHysteresis
	maxNodesShard := rxs.nodesShard + rxs.shardHysteresis

	nodesForSplit := (nbShards+1)*maxNodesShard + maxNodesMeta
	nodesForMerge := nbShards*rxs.nodesShard + rxs.nodesMeta
	nbShardsNew := nbShards

	if nodesNewEpoch > nodesForSplit {
		nbNodesWithoutMaxMeta := nodesNewEpoch - maxNodesMeta
		nbShardsNew = nbNodesWithoutMaxMeta / maxNodesShard

		return nbShardsNew
	}

	if nodesNewEpoch == nodesForMerge {
		return nbShardsNew - 1
	}

	return nbShardsNew
}

// shuffleOutNodes shuffles the list of eligible validators in each shard and returns the array of shuffled out
// validators
func shuffleOutNodes(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	leaving []Validator,
	randomness []byte,
) ([]Validator, map[uint32][]Validator, []Validator) {
	shuffledOut := make([]Validator, 0)
	newEligible := make(map[uint32][]Validator)
	var removed []Validator

	for shard, validators := range eligible {

		nodesToSelect := len(waiting[shard])

		if len(validators) < nodesToSelect {
			nodesToSelect = len(validators)
		}

		validators, removed = removeValidatorsFromList(validators, leaving, nodesToSelect)
		leaving, _ = removeValidatorsFromList(leaving, removed, len(removed))

		nodesToSelect -= len(removed)
		shardShuffledEligible := shuffleList(validators, randomness)
		shardShuffledOut := shardShuffledEligible[:nodesToSelect]
		shuffledOut = append(shuffledOut, shardShuffledOut...)

		newEligible[shard], _ = removeValidatorsFromList(validators, shardShuffledOut, len(shardShuffledOut))
	}

	return shuffledOut, newEligible, leaving
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

	for _, v2 := range validatorsToRemove {
		for i, v1 := range resultedList {
			if v1 == v2 {
				resultedList = removeValidatorFromList(resultedList, i)
				removed = append(removed, v1)
				break
			}
		}

		if len(removed) == maxToRemove {
			break
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
	newNbShards uint32,
) (map[uint32][]Validator, map[uint32][]Validator) {
	log.Error(ErrNotImplemented.Error())

	// TODO: do the split
	return copyValidatorMap(eligible), copyValidatorMap(waiting)
}

// mergeShards merges the required shards, returning the resulting shards configuration for eligible and waiting lists
func (rxs *randXORShuffler) mergeShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	newNbShards uint32,
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

// promoteWaitingToEligible moves the validators in the waiting list to corresponding eligible list
func promoteWaitingToEligible(eligible map[uint32][]Validator, waiting map[uint32][]Validator) {
	for k, v := range waiting {
		eligible[k] = append(eligible[k], v...)
		waiting[k] = make([]Validator, 0)
	}
}

// distributeNewNodes distributes the newly registered nodes to the existing shards waiting lists
func distributeValidators(
	validators []Validator,
	destLists map[uint32][]Validator,
	randomness []byte,
	nbShardsPlusMeta uint32,
) {
	// if there was a split or a merge, eligible map should already have a different nb of keys (shards)
	shuffledValidators := shuffleList(validators, randomness)
	var shardId uint32

	if len(destLists) == 0 {
		destLists = make(map[uint32][]Validator)
	}

	for i, v := range shuffledValidators {
		shardId = uint32(i) % nbShardsPlusMeta
		if shardId == nbShardsPlusMeta-1 {
			shardId = MetachainShardId
		}
		destLists[shardId] = append(destLists[shardId], v)
	}
}
