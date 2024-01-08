package nodesCoordinator

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
)

var _ NodesShuffler = (*randHashShuffler)(nil)

// NodesShufflerArgs defines the arguments required to create a nodes shuffler
type NodesShufflerArgs struct {
	NodesShard           uint32
	NodesMeta            uint32
	Hysteresis           float32
	Adaptivity           bool
	ShuffleBetweenShards bool
	MaxNodesEnableConfig []config.MaxNodesChangeConfig
	EnableEpochsHandler  common.EnableEpochsHandler
	EnableEpochs         config.EnableEpochs
}

type shuffleNodesArg struct {
	eligible                map[uint32][]Validator
	waiting                 map[uint32][]Validator
	unstakeLeaving          []Validator
	additionalLeaving       []Validator
	newNodes                []Validator
	auction                 []Validator
	randomness              []byte
	distributor             ValidatorsDistributor
	nodesMeta               uint32
	nodesPerShard           uint32
	nbShards                uint32
	maxNodesToSwapPerShard  uint32
	maxNumNodes             uint32
	flagBalanceWaitingLists bool
	flagStakingV4Step2      bool
	flagStakingV4Step3      bool
}

type shuffledNodesConfig struct {
	numShuffled        uint32
	numNewEligible     uint32
	numNewWaiting      uint32
	numSelectedAuction uint32
	maxNumNodes        uint32
	flagStakingV4Step2 bool
}

// TODO: Decide if transaction load statistics will be used for limiting the number of shards
type randHashShuffler struct {
	// TODO: remove the references to this constant and the distributor
	// when reinitialization of node in new shard is implemented
	shuffleBetweenShards bool

	adaptivity                bool
	nodesShard                uint32
	nodesMeta                 uint32
	shardHysteresis           uint32
	metaHysteresis            uint32
	activeNodesConfig         config.MaxNodesChangeConfig
	availableNodesConfigs     []config.MaxNodesChangeConfig
	mutShufflerParams         sync.RWMutex
	validatorDistributor      ValidatorsDistributor
	flagBalanceWaitingLists   atomic.Flag
	enableEpochsHandler       common.EnableEpochsHandler
	stakingV4Step2EnableEpoch uint32
	flagStakingV4Step2        atomic.Flag
	stakingV4Step3EnableEpoch uint32
	flagStakingV4Step3        atomic.Flag
}

// NewHashValidatorsShuffler creates a validator shuffler that uses a hash between validator key and a given
// random number to do the shuffling
func NewHashValidatorsShuffler(args *NodesShufflerArgs) (*randHashShuffler, error) {
	if args == nil {
		return nil, ErrNilNodeShufflerArguments
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, ErrNilEnableEpochsHandler
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.BalanceWaitingListsFlag,
	})
	if err != nil {
		return nil, err
	}

	var configs []config.MaxNodesChangeConfig

	log.Debug("hashValidatorShuffler: enable epoch for max nodes change", "epoch", args.MaxNodesEnableConfig)
	log.Debug("hashValidatorShuffler: enable epoch for staking v4 step 2", "epoch", args.EnableEpochs.StakingV4Step2EnableEpoch)
	log.Debug("hashValidatorShuffler: enable epoch for staking v4 step 3", "epoch", args.EnableEpochs.StakingV4Step3EnableEpoch)

	if args.MaxNodesEnableConfig != nil {
		configs = make([]config.MaxNodesChangeConfig, len(args.MaxNodesEnableConfig))
		copy(configs, args.MaxNodesEnableConfig)
	}

	log.Debug("Shuffler created", "shuffleBetweenShards", args.ShuffleBetweenShards)
	rxs := &randHashShuffler{
		shuffleBetweenShards:      args.ShuffleBetweenShards,
		availableNodesConfigs:     configs,
		enableEpochsHandler:       args.EnableEpochsHandler,
		stakingV4Step2EnableEpoch: args.EnableEpochs.StakingV4Step2EnableEpoch,
		stakingV4Step3EnableEpoch: args.EnableEpochs.StakingV4Step3EnableEpoch,
	}

	rxs.UpdateParams(args.NodesShard, args.NodesMeta, args.Hysteresis, args.Adaptivity)

	if rxs.shuffleBetweenShards {
		rxs.validatorDistributor = &CrossShardValidatorDistributor{}
	} else {
		rxs.validatorDistributor = &IntraShardValidatorDistributor{}
	}

	rxs.sortConfigs()

	return rxs, nil
}

// UpdateParams updates the shuffler parameters
// Should be called when new params are agreed through governance
func (rhs *randHashShuffler) UpdateParams(
	nodesShard uint32,
	nodesMeta uint32,
	hysteresis float32,
	adaptivity bool,
) {
	// TODO: are there constraints we want to enforce? e.g min/max hysteresis
	shardHysteresis := uint32(float32(nodesShard) * hysteresis)
	metaHysteresis := uint32(float32(nodesMeta) * hysteresis)

	rhs.mutShufflerParams.Lock()
	rhs.shardHysteresis = shardHysteresis
	rhs.metaHysteresis = metaHysteresis
	rhs.nodesShard = nodesShard
	rhs.nodesMeta = nodesMeta
	rhs.adaptivity = adaptivity
	rhs.mutShufflerParams.Unlock()
}

// UpdateNodeLists shuffles the nodes and returns the lists with the new nodes configuration
// The function needs to ensure that:
//  1. Old eligible nodes list will have up to shuffleOutThreshold percent nodes shuffled out from each shard
//  2. The leaving nodes are checked against the eligible nodes and waiting nodes and removed if present from the
//     pools and leaving nodes list (if remaining nodes can still sustain the shard)
//  3. shuffledOutNodes = oldEligibleNodes + waitingListNodes - minNbNodesPerShard (for each shard)
//  4. Old waiting nodes list for each shard will be added to the remaining eligible nodes list
//  5. The new nodes are equally distributed among the existing shards into waiting lists
//  6. The shuffled out nodes are distributed among the existing shards into waiting lists.
//     We may have three situations:
//     a)  In case (shuffled out nodes + new nodes) > (nbShards * perShardHysteresis + minNodesPerShard) then
//     we need to prepare for a split event, so a higher percentage of nodes need to be directed to the shard
//     that will be split.
//     b)  In case (shuffled out nodes + new nodes) < (nbShards * perShardHysteresis) then we can immediately
//     execute the shard merge
//     c)  No change in the number of shards then nothing extra needs to be done
func (rhs *randHashShuffler) UpdateNodeLists(args ArgsUpdateNodes) (*ResUpdateNodes, error) {
	rhs.updateShufflerConfig(args.Epoch)
	eligibleAfterReshard := copyValidatorMap(args.Eligible)
	waitingAfterReshard := copyValidatorMap(args.Waiting)

	args.AdditionalLeaving = removeDupplicates(args.UnStakeLeaving, args.AdditionalLeaving)
	totalLeavingNum := len(args.AdditionalLeaving) + len(args.UnStakeLeaving)

	newNbShards := rhs.computeNewShards(
		args.Eligible,
		args.Waiting,
		len(args.NewNodes),
		totalLeavingNum,
		args.NbShards,
	)

	rhs.mutShufflerParams.RLock()
	canSplit := rhs.adaptivity && newNbShards > args.NbShards
	canMerge := rhs.adaptivity && newNbShards < args.NbShards
	nodesPerShard := rhs.nodesShard
	nodesMeta := rhs.nodesMeta
	rhs.mutShufflerParams.RUnlock()

	if canSplit {
		eligibleAfterReshard, waitingAfterReshard = rhs.splitShards(args.Eligible, args.Waiting, newNbShards)
	}
	if canMerge {
		eligibleAfterReshard, waitingAfterReshard = rhs.mergeShards(args.Eligible, args.Waiting, newNbShards)
	}

	return shuffleNodes(shuffleNodesArg{
		eligible:                eligibleAfterReshard,
		waiting:                 waitingAfterReshard,
		unstakeLeaving:          args.UnStakeLeaving,
		additionalLeaving:       args.AdditionalLeaving,
		newNodes:                args.NewNodes,
		auction:                 args.Auction,
		randomness:              args.Rand,
		nodesMeta:               nodesMeta,
		nodesPerShard:           nodesPerShard,
		nbShards:                args.NbShards,
		distributor:             rhs.validatorDistributor,
		maxNodesToSwapPerShard:  rhs.activeNodesConfig.NodesToShufflePerShard,
		flagBalanceWaitingLists: rhs.enableEpochsHandler.IsFlagEnabledInEpoch(common.BalanceWaitingListsFlag, args.Epoch),
		flagStakingV4Step2:      rhs.flagStakingV4Step2.IsSet(),
		flagStakingV4Step3:      rhs.flagStakingV4Step3.IsSet(),
		maxNumNodes:             rhs.activeNodesConfig.MaxNumNodes,
	})
}

func removeDupplicates(unstake []Validator, additionalLeaving []Validator) []Validator {
	additionalCopy := make([]Validator, 0, len(additionalLeaving))
	additionalCopy = append(additionalCopy, additionalLeaving...)

	for _, unstakeValidator := range unstake {
		for i := len(additionalCopy) - 1; i >= 0; i-- {
			if bytes.Equal(unstakeValidator.PubKey(), additionalCopy[i].PubKey()) {
				additionalCopy = removeValidatorFromList(additionalCopy, i)
			}
		}
	}

	return additionalCopy
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
	if len(leavingNodes) < nbToRemove {
		nbToRemove = len(leavingNodes)
	}

	vList, removedNodes := removeValidatorsFromList(existingNodes[shard], leavingNodes, nbToRemove)
	leavingNodes, _ = removeValidatorsFromList(leavingNodes, removedNodes, len(removedNodes))
	existingNodes[shard] = vList
	return leavingNodes, len(removedNodes)
}

// IsInterfaceNil verifies if the underlying object is nil
func (rhs *randHashShuffler) IsInterfaceNil() bool {
	return rhs == nil
}

func shuffleNodes(arg shuffleNodesArg) (*ResUpdateNodes, error) {
	allLeaving := append(arg.unstakeLeaving, arg.additionalLeaving...)

	waitingCopy := copyValidatorMap(arg.waiting)
	eligibleCopy := copyValidatorMap(arg.eligible)

	createListsForAllShards(waitingCopy, arg.nbShards)

	numToRemove, err := computeNumToRemove(arg)
	if err != nil {
		return nil, err
	}

	for i, toRemove := range numToRemove {
		if toRemove > int(arg.maxNodesToSwapPerShard) {
			numToRemove[i] = int(arg.maxNodesToSwapPerShard)
		}
	}

	remainingUnstakeLeaving, _ := removeLeavingNodesNotExistingInEligibleOrWaiting(arg.unstakeLeaving, waitingCopy, eligibleCopy)
	remainingAdditionalLeaving, _ := removeLeavingNodesNotExistingInEligibleOrWaiting(arg.additionalLeaving, waitingCopy, eligibleCopy)

	newEligible, newWaiting, stillRemainingUnstakeLeaving := removeLeavingNodesFromValidatorMaps(
		eligibleCopy,
		waitingCopy,
		numToRemove,
		remainingUnstakeLeaving)
	newEligible, newWaiting, stillRemainingAdditionalLeaving := removeLeavingNodesFromValidatorMaps(
		newEligible,
		newWaiting,
		numToRemove,
		remainingAdditionalLeaving)

	stillRemainingInLeaving := append(stillRemainingUnstakeLeaving, stillRemainingAdditionalLeaving...)

	shuffledOutMap, newEligible := shuffleOutNodes(newEligible, numToRemove, arg.randomness)

	err = moveMaxNumNodesToMap(newEligible, newWaiting, arg.nodesMeta, arg.nodesPerShard)
	if err != nil {
		return nil, fmt.Errorf("moveNodesToMap failed, error: %w", err)
	}

	err = checkAndDistributeNewNodes(newWaiting, arg.newNodes, arg.randomness, arg.flagStakingV4Step3)
	if err != nil {
		return nil, fmt.Errorf("distributeValidators newNodes failed, error: %w", err)
	}

	shuffledNodesCfg := &shuffledNodesConfig{
		numShuffled:        getNumPubKeys(shuffledOutMap),
		numNewEligible:     getNumPubKeys(newEligible),
		numNewWaiting:      getNumPubKeys(newWaiting),
		numSelectedAuction: uint32(len(arg.auction)),
		maxNumNodes:        arg.maxNumNodes,
		flagStakingV4Step2: arg.flagStakingV4Step2,
	}

	lowWaitingList := shouldDistributeShuffledToWaitingInStakingV4(shuffledNodesCfg)
	if arg.flagStakingV4Step3 || lowWaitingList {
		log.Debug("distributing selected nodes from auction to waiting",
			"num auction nodes", len(arg.auction), "num waiting nodes", shuffledNodesCfg.numNewWaiting)

		// Distribute selected validators from AUCTION -> WAITING
		err = distributeValidators(newWaiting, arg.auction, arg.randomness, arg.flagBalanceWaitingLists)
		if err != nil {
			return nil, fmt.Errorf("distributeValidators auction list failed, error: %w", err)
		}
	}

	if !arg.flagStakingV4Step2 || lowWaitingList {
		log.Debug("distributing shuffled out nodes to waiting",
			"num shuffled nodes", shuffledNodesCfg.numShuffled, "num waiting nodes", shuffledNodesCfg.numNewWaiting)

		// Distribute validators from SHUFFLED OUT -> WAITING
		err = arg.distributor.DistributeValidators(newWaiting, shuffledOutMap, arg.randomness, arg.flagBalanceWaitingLists)
		if err != nil {
			return nil, fmt.Errorf("distributeValidators shuffled out failed, error: %w", err)
		}
	}

	actualLeaving, _ := removeValidatorsFromList(allLeaving, stillRemainingInLeaving, len(stillRemainingInLeaving))

	return &ResUpdateNodes{
		Eligible:       newEligible,
		Waiting:        newWaiting,
		ShuffledOut:    shuffledOutMap,
		Leaving:        actualLeaving,
		StillRemaining: stillRemainingInLeaving,
	}, nil
}

func createListsForAllShards(shardMap map[uint32][]Validator, shards uint32) {
	for shardId := uint32(0); shardId < shards; shardId++ {
		if shardMap[shardId] == nil {
			shardMap[shardId] = make([]Validator, 0)
		}
	}

	if shardMap[core.MetachainShardId] == nil {
		shardMap[core.MetachainShardId] = make([]Validator, 0)
	}
}

func computeNumToRemove(arg shuffleNodesArg) (map[uint32]int, error) {
	numToRemove := make(map[uint32]int)
	if arg.nbShards == 0 {
		return numToRemove, nil
	}

	for shardId := uint32(0); shardId < arg.nbShards; shardId++ {
		maxToRemove, err := computeNumToRemovePerShard(
			len(arg.eligible[shardId]),
			len(arg.waiting[shardId]),
			int(arg.nodesPerShard))
		if err != nil {
			return nil, fmt.Errorf("%w shard=%v", err, shardId)
		}
		numToRemove[shardId] = maxToRemove
	}

	maxToRemove, err := computeNumToRemovePerShard(
		len(arg.eligible[core.MetachainShardId]),
		len(arg.waiting[core.MetachainShardId]),
		int(arg.nodesMeta))
	if err != nil {
		return nil, fmt.Errorf("%w shard=%v", err, core.MetachainShardId)
	}
	numToRemove[core.MetachainShardId] = maxToRemove

	return numToRemove, nil
}

func computeNumToRemovePerShard(numEligible int, numWaiting int, nodesPerShard int) (int, error) {
	notEnoughValidatorsInShard := numEligible+numWaiting < nodesPerShard
	if notEnoughValidatorsInShard {
		return 0, ErrSmallShardEligibleListSize
	}
	return numEligible + numWaiting - nodesPerShard, nil
}

func removeLeavingNodesNotExistingInEligibleOrWaiting(
	leavingValidators []Validator,
	waiting map[uint32][]Validator,
	eligible map[uint32][]Validator,
) ([]Validator, []Validator) {
	notFoundValidators := make([]Validator, 0)

	for _, v := range leavingValidators {
		found, _ := searchInMap(waiting, v.PubKey())
		if found {
			continue
		}
		found, _ = searchInMap(eligible, v.PubKey())
		if !found {
			log.Debug("Leaving validator not found in waiting or eligible", "pk", v.PubKey())
			notFoundValidators = append(notFoundValidators, v)
		}
	}

	return removeValidatorsFromList(leavingValidators, notFoundValidators, len(notFoundValidators))
}

func removeLeavingNodesFromValidatorMaps(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	numToRemove map[uint32]int,
	leaving []Validator,
) (map[uint32][]Validator, map[uint32][]Validator, []Validator) {

	stillRemainingInLeaving := make([]Validator, len(leaving))
	copy(stillRemainingInLeaving, leaving)

	newWaiting, stillRemainingInLeaving := removeNodesFromMap(waiting, stillRemainingInLeaving, numToRemove)
	newEligible, stillRemainingInLeaving := removeNodesFromMap(eligible, stillRemainingInLeaving, numToRemove)
	return newEligible, newWaiting, stillRemainingInLeaving
}

// computeNewShards determines the new number of shards based on the number of nodes in the network
func (rhs *randHashShuffler) computeNewShards(
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

	rhs.mutShufflerParams.RLock()
	maxNodesMeta := rhs.nodesMeta + rhs.metaHysteresis
	maxNodesShard := rhs.nodesShard + rhs.shardHysteresis
	nodesForSplit := (nbShards+1)*maxNodesShard + maxNodesMeta
	nodesForMerge := nbShards*rhs.nodesShard + rhs.nodesMeta
	rhs.mutShufflerParams.RUnlock()

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

// shuffleOutNodes shuffles the list of eligible validators in each shard and returns the map of shuffled out
// validators
func shuffleOutNodes(
	eligible map[uint32][]Validator,
	numToShuffle map[uint32]int,
	randomness []byte,
) (map[uint32][]Validator, map[uint32][]Validator) {
	shuffledOutMap := make(map[uint32][]Validator)
	newEligible := make(map[uint32][]Validator)
	var shardShuffledOut []Validator

	sortedShardIds := sortKeys(eligible)
	for _, shardId := range sortedShardIds {
		validators := eligible[shardId]
		shardShuffledOut, validators = shuffleOutShard(validators, numToShuffle[shardId], randomness)
		shuffledOutMap[shardId] = shardShuffledOut
		newEligible[shardId], _ = removeValidatorsFromList(validators, shardShuffledOut, len(shardShuffledOut))
	}

	return shuffledOutMap, newEligible
}

// shuffleOutShard selects the validators to be shuffled out from a shard
func shuffleOutShard(
	validators []Validator,
	validatorsToSelect int,
	randomness []byte,
) ([]Validator, []Validator) {
	if len(validators) < validatorsToSelect {
		validatorsToSelect = len(validators)
	}

	shardShuffledEligible := shuffleList(validators, randomness)
	shardShuffledOut := shardShuffledEligible[:validatorsToSelect]
	remainingEligible := shardShuffledEligible[validatorsToSelect:]

	return shardShuffledOut, remainingEligible
}

// shuffleList returns a shuffled list of validators.
// The shuffling is done by hash-ing the randomness concatenated with the
// public keys of validators and sorting the validators depending on
// the hash result.
func shuffleList(validators []Validator, randomness []byte) []Validator {
	keys := make([]string, len(validators))
	mapValidators := make(map[string]Validator)
	var concat []byte

	hasher := sha256.NewSha256()
	for i, v := range validators {
		concat = append(v.PubKey(), randomness...)

		keys[i] = string(hasher.Compute(string(concat)))
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

		for i := len(resultedList) - 1; i >= 0; i-- {
			val := resultedList[i]
			if bytes.Equal(val.PubKey(), valToRemove.PubKey()) {
				resultedList = removeValidatorFromList(resultedList, i)

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

func checkAndDistributeNewNodes(
	waiting map[uint32][]Validator,
	newNodes []Validator,
	randomness []byte,
	flagStakingV4Step3 bool,
) error {
	if !flagStakingV4Step3 {
		return distributeValidators(waiting, newNodes, randomness, false)
	}

	if len(newNodes) > 0 {
		return epochStart.ErrReceivedNewListNodeInStakingV4
	}

	return nil
}

func shouldDistributeShuffledToWaitingInStakingV4(shuffledNodesCfg *shuffledNodesConfig) bool {
	if !shuffledNodesCfg.flagStakingV4Step2 {
		return false
	}

	totalNewWaiting := shuffledNodesCfg.numNewWaiting + shuffledNodesCfg.numSelectedAuction
	totalNodes := totalNewWaiting + shuffledNodesCfg.numNewEligible + shuffledNodesCfg.numShuffled

	log.Debug("checking if should distribute shuffled out nodes to waiting in staking v4",
		"numShuffled", shuffledNodesCfg.numShuffled,
		"numNewEligible", shuffledNodesCfg.numNewEligible,
		"numSelectedAuction", shuffledNodesCfg.numSelectedAuction,
		"totalNewWaiting", totalNewWaiting,
		"totalNodes", totalNodes,
		"maxNumNodes", shuffledNodesCfg.maxNumNodes,
	)

	distributeShuffledToWaitingInStakingV4 := false
	if totalNodes <= shuffledNodesCfg.maxNumNodes {
		log.Warn("num of total nodes in waiting is too low after shuffling; will distribute " +
			"shuffled out nodes directly to waiting and skip sending them to auction")

		distributeShuffledToWaitingInStakingV4 = true
	}

	return distributeShuffledToWaitingInStakingV4
}

func removeValidatorFromListKeepOrder(validatorList []Validator, index int) []Validator {
	indexNotOK := index > len(validatorList)-1 || index < 0
	if indexNotOK {
		return validatorList
	}

	return append(validatorList[:index], validatorList[index+1:]...)
}

// splitShards prepares for the shards split, or if already prepared does the split returning the resulting
// shards configuration for eligible and waiting lists
func (rhs *randHashShuffler) splitShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	_ uint32,
) (map[uint32][]Validator, map[uint32][]Validator) {
	log.Error(ErrNotImplemented.Error())

	// TODO: do the split
	return copyValidatorMap(eligible), copyValidatorMap(waiting)
}

// mergeShards merges the required shards, returning the resulting shards configuration for eligible and waiting lists
func (rhs *randHashShuffler) mergeShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	_ uint32,
) (map[uint32][]Validator, map[uint32][]Validator) {
	log.Error(ErrNotImplemented.Error())

	// TODO: do the merge
	return copyValidatorMap(eligible), copyValidatorMap(waiting)
}

// copyValidatorMap creates a copy for the Validators map, creating copies for each of the lists for each shard
func copyValidatorMap(validatorsMap map[uint32][]Validator) map[uint32][]Validator {
	result := make(map[uint32][]Validator)

	for shardId, validators := range validatorsMap {
		elems := make([]Validator, 0)
		result[shardId] = append(elems, validators...)
	}

	return result
}

// moveNodesToMap moves the validators in the source list to the corresponding destination list
func moveNodesToMap(destination map[uint32][]Validator, source map[uint32][]Validator) error {
	if destination == nil {
		return ErrNilOrEmptyDestinationForDistribute
	}

	for shardId, validators := range source {
		destination[shardId] = append(destination[shardId], validators...)
		source[shardId] = make([]Validator, 0)
	}

	return nil
}

func getNumPubKeys(shardValidatorsMap map[uint32][]Validator) uint32 {
	numPubKeys := uint32(0)

	for _, validatorsInShard := range shardValidatorsMap {
		numPubKeys += uint32(len(validatorsInShard))
	}

	return numPubKeys
}

// moveMaxNumNodesToMap moves the validators in the source list to the corresponding destination list
// but adding just enough nodes so that at most the number of nodes is kept in the destination list
// The parameter maxNodesToMove is a limiting factor and should limit the number of nodes
func moveMaxNumNodesToMap(
	destination map[uint32][]Validator,
	source map[uint32][]Validator,
	numMeta uint32,
	numShard uint32,
) error {
	if destination == nil {
		return ErrNilOrEmptyDestinationForDistribute
	}

	for shardId, validators := range source {
		maxNodes := numShard
		if shardId == core.MetachainShardId {
			maxNodes = numMeta
		}

		numNeededNodes := computeNeededNodes(destination[shardId], source[shardId], maxNodes)
		destination[shardId] = append(destination[shardId], validators[0:numNeededNodes]...)
		source[shardId] = validators[numNeededNodes:]
	}

	return nil
}

func computeNeededNodes(destination []Validator, source []Validator, maxNumNodes uint32) uint32 {
	numNeededNodes := uint32(0)
	numCurrentNodes := uint32(len(destination))
	numSourceNodes := uint32(len(source))
	if maxNumNodes > numCurrentNodes {
		numNeededNodes = maxNumNodes - numCurrentNodes
	}
	if numSourceNodes < numNeededNodes {
		return numSourceNodes
	}

	return numNeededNodes
}

// distributeNewNodes distributes a list of validators to the given validators map
func distributeValidators(destLists map[uint32][]Validator, validators []Validator, randomness []byte, balanced bool) error {
	if len(destLists) == 0 {
		return ErrNilOrEmptyDestinationForDistribute
	}

	// if there was a split, or a merge, eligible map should already have a different nb of keys (shards)
	shuffledValidators := shuffleList(validators, randomness)
	var shardId uint32
	sortedShardIds := sortKeys(destLists)
	destLength := uint32(len(sortedShardIds))

	if balanced {
		shuffledValidators = equalizeValidatorsLists(destLists, shuffledValidators)
	}

	for i, v := range shuffledValidators {
		shardId = sortedShardIds[uint32(i)%destLength]
		destLists[shardId] = append(destLists[shardId], v)
	}

	return nil
}

func equalizeValidatorsLists(destLists map[uint32][]Validator, validators []Validator) []Validator {
	log.Debug("equalizeValidatorsLists")

	maxListSize := getMaxListSize(destLists)
	indexValidators := 0
	remainingValidatorsNumber := len(validators)

	sortedShardIds := sortKeys(destLists)

	for _, shardID := range sortedShardIds {
		shardList := destLists[shardID]
		if len(shardList) < maxListSize {
			toMove := maxListSize - len(shardList)
			if toMove > remainingValidatorsNumber {
				toMove = remainingValidatorsNumber
			}

			destLists[shardID] = append(destLists[shardID], validators[indexValidators:indexValidators+toMove]...)
			remainingValidatorsNumber -= toMove
			indexValidators += toMove
		}
	}

	return validators[indexValidators : indexValidators+remainingValidatorsNumber]
}

func getMaxListSize(lists map[uint32][]Validator) int {
	var maxSize int

	for _, list := range lists {
		if maxSize < len(list) {
			maxSize = len(list)
		}
	}
	return maxSize
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

// updateShufflerConfig updates the shuffler config according to the current epoch.
func (rhs *randHashShuffler) updateShufflerConfig(epoch uint32) {
	rhs.mutShufflerParams.Lock()
	defer rhs.mutShufflerParams.Unlock()
	rhs.activeNodesConfig.NodesToShufflePerShard = rhs.nodesShard
	for _, maxNodesConfig := range rhs.availableNodesConfigs {
		if epoch >= maxNodesConfig.EpochEnable {
			rhs.activeNodesConfig = maxNodesConfig
		}
	}

	log.Debug("randHashShuffler: UpdateShufflerConfig",
		"epoch", epoch,
		"maxNumNodes", rhs.activeNodesConfig.MaxNumNodes,
		"epochEnable", rhs.activeNodesConfig.EpochEnable,
		"maxNodesToShufflePerShard", rhs.activeNodesConfig.NodesToShufflePerShard,
	)

	rhs.flagStakingV4Step3.SetValue(epoch >= rhs.stakingV4Step3EnableEpoch)
	log.Debug("staking v4 step3", "enabled", rhs.flagStakingV4Step3.IsSet())

	rhs.flagStakingV4Step2.SetValue(epoch >= rhs.stakingV4Step2EnableEpoch)
	log.Debug("staking v4 step2", "enabled", rhs.flagStakingV4Step2.IsSet())
}

func (rhs *randHashShuffler) sortConfigs() {
	rhs.mutShufflerParams.Lock()
	sort.Slice(rhs.availableNodesConfigs, func(i, j int) bool {
		return rhs.availableNodesConfigs[i].EpochEnable < rhs.availableNodesConfigs[j].EpochEnable
	})
	rhs.mutShufflerParams.Unlock()
}
