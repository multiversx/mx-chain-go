package sharding

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ NodesCoordinator = (*indexHashedNodesCoordinator)(nil)
var _ PublicKeysSelector = (*indexHashedNodesCoordinator)(nil)

const (
	keyFormat               = "%s_%v_%v_%v"
	defaultSelectionChances = uint32(1)
)

// TODO: move this to config parameters
const nodeCoordinatorStoredEpochs = 3

type validatorWithShardID struct {
	validator Validator
	shardID   uint32
}

type validatorList []Validator

// Len will return the length of the validatorList
func (v validatorList) Len() int { return len(v) }

// Swap will interchange the objects on input indexes
func (v validatorList) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

// Less will return true if object on index i should appear before object in index j
// Sorting of validators should be by index and public key
func (v validatorList) Less(i, j int) bool {
	if v[i].Index() == v[j].Index() {
		return bytes.Compare(v[i].PubKey(), v[j].PubKey()) < 0
	}
	return v[i].Index() < v[j].Index()
}

// TODO: add a parameter for shardID  when acting as observer
type epochNodesConfig struct {
	nbShards     uint32
	shardID      uint32
	eligibleMap  map[uint32][]Validator
	waitingMap   map[uint32][]Validator
	selectors    map[uint32]RandomSelector
	leavingMap   map[uint32][]Validator
	newList      []Validator
	mutNodesMaps sync.RWMutex
}

type indexHashedNodesCoordinator struct {
	shardIDAsObserver             uint32
	currentEpoch                  uint32
	shardConsensusGroupSize       int
	metaConsensusGroupSize        int
	numTotalEligible              uint64
	selfPubKey                    []byte
	savedStateKey                 []byte
	marshalizer                   marshal.Marshalizer
	hasher                        hashing.Hasher
	shuffler                      NodesShuffler
	epochStartRegistrationHandler EpochStartEventNotifier
	bootStorer                    storage.Storer
	nodesConfig                   map[uint32]*epochNodesConfig
	mutNodesConfig                sync.RWMutex
	mutSavedStateKey              sync.RWMutex
	nodesCoordinatorHelper        NodesCoordinatorHelper
	consensusGroupCacher          Cacher
	loadingFromDisk               atomic.Value
	shuffledOutHandler            ShuffledOutHandler
	startEpoch                    uint32
	publicKeyToValidatorMap       map[string]*validatorWithShardID
}

// NewIndexHashedNodesCoordinator creates a new index hashed group selector
func NewIndexHashedNodesCoordinator(arguments ArgNodesCoordinator) (*indexHashedNodesCoordinator, error) {
	err := checkArguments(arguments)
	if err != nil {
		return nil, err
	}

	nodesConfig := make(map[uint32]*epochNodesConfig, nodeCoordinatorStoredEpochs)

	nodesConfig[arguments.Epoch] = &epochNodesConfig{
		nbShards:    arguments.NbShards,
		shardID:     arguments.ShardIDAsObserver,
		eligibleMap: make(map[uint32][]Validator),
		waitingMap:  make(map[uint32][]Validator),
		selectors:   make(map[uint32]RandomSelector),
		leavingMap:  make(map[uint32][]Validator),
		newList:     make([]Validator, 0),
	}

	savedKey := arguments.Hasher.Compute(string(arguments.SelfPublicKey))

	ihgs := &indexHashedNodesCoordinator{
		marshalizer:                   arguments.Marshalizer,
		hasher:                        arguments.Hasher,
		shuffler:                      arguments.Shuffler,
		epochStartRegistrationHandler: arguments.EpochStartNotifier,
		bootStorer:                    arguments.BootStorer,
		selfPubKey:                    arguments.SelfPublicKey,
		nodesConfig:                   nodesConfig,
		currentEpoch:                  arguments.Epoch,
		savedStateKey:                 savedKey,
		shardConsensusGroupSize:       arguments.ShardConsensusGroupSize,
		metaConsensusGroupSize:        arguments.MetaConsensusGroupSize,
		consensusGroupCacher:          arguments.ConsensusGroupCache,
		shardIDAsObserver:             arguments.ShardIDAsObserver,
		shuffledOutHandler:            arguments.ShuffledOutHandler,
		startEpoch:                    arguments.StartEpoch,
		publicKeyToValidatorMap:       make(map[string]*validatorWithShardID),
	}

	ihgs.loadingFromDisk.Store(false)

	ihgs.nodesCoordinatorHelper = ihgs
	err = ihgs.setNodesPerShards(arguments.EligibleNodes, arguments.WaitingNodes, nil, arguments.Epoch)
	if err != nil {
		return nil, err
	}

	ihgs.fillPublicKeyToValidatorMap()
	err = ihgs.saveState(ihgs.savedStateKey)
	if err != nil {
		log.Error("saving initial nodes coordinator config failed",
			"error", err.Error())
	}
	log.Info("new nodes config is set for epoch", "epoch", arguments.Epoch)
	currentNodesConfig := ihgs.nodesConfig[arguments.Epoch]
	if currentNodesConfig == nil {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, arguments.Epoch)
	}

	currentConfig := nodesConfig[arguments.Epoch]
	if currentConfig == nil {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, arguments.Epoch)
	}

	displayNodesConfiguration(
		currentConfig.eligibleMap,
		currentConfig.waitingMap,
		currentConfig.leavingMap,
		make(map[uint32][]Validator),
		currentConfig.nbShards)

	ihgs.epochStartRegistrationHandler.RegisterHandler(ihgs)

	return ihgs, nil
}

func checkArguments(arguments ArgNodesCoordinator) error {
	if arguments.ShardConsensusGroupSize < 1 || arguments.MetaConsensusGroupSize < 1 {
		return ErrInvalidConsensusGroupSize
	}
	if arguments.NbShards < 1 {
		return ErrInvalidNumberOfShards
	}
	if arguments.ShardIDAsObserver >= arguments.NbShards && arguments.ShardIDAsObserver != core.MetachainShardId {
		return ErrInvalidShardId
	}
	if check.IfNil(arguments.Hasher) {
		return ErrNilHasher
	}
	if len(arguments.SelfPublicKey) == 0 {
		return ErrNilPubKey
	}
	if check.IfNil(arguments.Shuffler) {
		return ErrNilShuffler
	}
	if check.IfNil(arguments.BootStorer) {
		return ErrNilBootStorer
	}
	if check.IfNilReflect(arguments.ConsensusGroupCache) {
		return ErrNilCacher
	}
	if check.IfNil(arguments.Marshalizer) {
		return ErrNilMarshalizer
	}
	if check.IfNil(arguments.ShuffledOutHandler) {
		return ErrNilShuffledOutHandler
	}

	return nil
}

// setNodesPerShards loads the distribution of nodes per shard into the nodes management component
func (ihgs *indexHashedNodesCoordinator) setNodesPerShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	leaving map[uint32][]Validator,
	epoch uint32,
) error {
	ihgs.mutNodesConfig.Lock()
	defer ihgs.mutNodesConfig.Unlock()

	nodesConfig, ok := ihgs.nodesConfig[epoch]
	if !ok {
		log.Warn("Did not find nodesConfig", "epoch", epoch)
		nodesConfig = &epochNodesConfig{}
	}

	nodesConfig.mutNodesMaps.Lock()
	defer nodesConfig.mutNodesMaps.Unlock()

	if eligible == nil || waiting == nil {
		return ErrNilInputNodesMap
	}

	nodesList := eligible[core.MetachainShardId]
	if len(nodesList) < ihgs.metaConsensusGroupSize {
		return ErrSmallMetachainEligibleListSize
	}

	numTotalEligible := uint64(len(nodesList))
	for shardId := uint32(0); shardId < uint32(len(eligible)-1); shardId++ {
		nbNodesShard := len(eligible[shardId])
		if nbNodesShard < ihgs.shardConsensusGroupSize {
			return ErrSmallShardEligibleListSize
		}
		numTotalEligible += uint64(nbNodesShard)
	}

	var err error
	// nbShards holds number of shards without meta
	nodesConfig.nbShards = uint32(len(eligible) - 1)
	nodesConfig.eligibleMap = eligible
	nodesConfig.waitingMap = waiting
	nodesConfig.leavingMap = leaving
	nodesConfig.shardID = ihgs.computeShardForSelfPublicKey(nodesConfig)
	nodesConfig.selectors, err = ihgs.createSelectors(nodesConfig)
	if err != nil {
		return err
	}

	ihgs.nodesConfig[epoch] = nodesConfig
	ihgs.numTotalEligible = numTotalEligible

	return nil
}

// ComputeAdditionalLeaving - computes extra leaving validators based on computation at the start of epoch
func (ihgs *indexHashedNodesCoordinator) ComputeAdditionalLeaving(_ []*state.ShardValidatorInfo) (map[uint32][]Validator, error) {
	return make(map[uint32][]Validator), nil
}

// ComputeConsensusGroup will generate a list of validators based on the the eligible list
// and each eligible validator weight/chance
func (ihgs *indexHashedNodesCoordinator) ComputeConsensusGroup(
	randomness []byte,
	round uint64,
	shardID uint32,
	epoch uint32,
) (validatorsGroup []Validator, err error) {
	var selector RandomSelector
	var eligibleList []Validator

	log.Trace("computing consensus group for",
		"epoch", epoch,
		"shardID", shardID,
		"randomness", randomness,
		"round", round)

	if len(randomness) == 0 {
		return nil, ErrNilRandomness
	}

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	if ok {
		if shardID >= nodesConfig.nbShards && shardID != core.MetachainShardId {
			log.Warn("shardID is not ok", "shardID", shardID, "nbShards", nodesConfig.nbShards)
			ihgs.mutNodesConfig.RUnlock()
			return nil, ErrInvalidShardId
		}
		selector = nodesConfig.selectors[shardID]
		eligibleList = nodesConfig.eligibleMap[shardID]
	}
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	key := []byte(fmt.Sprintf(keyFormat, string(randomness), round, shardID, epoch))
	validators := ihgs.searchConsensusForKey(key)
	if validators != nil {
		return validators, nil
	}

	consensusSize := ihgs.ConsensusGroupSize(shardID)
	randomness = []byte(fmt.Sprintf("%d-%s", round, randomness))

	log.Debug("computeValidatorsGroup",
		"randomness", randomness,
		"consensus size", consensusSize,
		"eligible list length", len(eligibleList),
		"epoch", epoch,
		"round", round,
		"shardID", shardID)

	tempList, err := selectValidators(selector, randomness, uint32(consensusSize), eligibleList)
	if err != nil {
		return nil, err
	}

	size := 0
	for _, v := range tempList {
		size += v.Size()
	}

	ihgs.consensusGroupCacher.Put(key, tempList, size)

	return tempList, nil
}

func (ihgs *indexHashedNodesCoordinator) searchConsensusForKey(key []byte) []Validator {
	value, ok := ihgs.consensusGroupCacher.Get(key)
	if ok {
		consensusGroup, typeOk := value.([]Validator)
		if typeOk {
			return consensusGroup
		}
	}
	return nil
}

// GetValidatorWithPublicKey gets the validator with the given public key
func (ihgs *indexHashedNodesCoordinator) GetValidatorWithPublicKey(publicKey []byte) (Validator, uint32, error) {
	if len(publicKey) == 0 {
		return nil, 0, ErrNilPubKey
	}
	ihgs.mutNodesConfig.RLock()
	v, ok := ihgs.publicKeyToValidatorMap[string(publicKey)]
	ihgs.mutNodesConfig.RUnlock()
	if ok {
		return v.validator, v.shardID, nil
	}

	return nil, 0, ErrValidatorNotFound
}

// GetConsensusValidatorsPublicKeys calculates the validators consensus group for a specific shard, randomness and round number,
// returning their public keys
func (ihgs *indexHashedNodesCoordinator) GetConsensusValidatorsPublicKeys(
	randomness []byte,
	round uint64,
	shardID uint32,
	epoch uint32,
) ([]string, error) {
	consensusNodes, err := ihgs.ComputeConsensusGroup(randomness, round, shardID, epoch)
	if err != nil {
		return nil, err
	}

	pubKeys := make([]string, 0)

	for _, v := range consensusNodes {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

// GetAllEligibleValidatorsPublicKeys will return all validators public keys for all shards
func (ihgs *indexHashedNodesCoordinator) GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	for shardID, shardEligible := range nodesConfig.eligibleMap {
		for i := 0; i < len(shardEligible); i++ {
			validatorsPubKeys[shardID] = append(validatorsPubKeys[shardID], shardEligible[i].PubKey())
		}
	}

	return validatorsPubKeys, nil
}

// GetAllWaitingValidatorsPublicKeys will return all validators public keys for all shards
func (ihgs *indexHashedNodesCoordinator) GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	for shardID, shardWaiting := range nodesConfig.waitingMap {
		for i := 0; i < len(shardWaiting); i++ {
			validatorsPubKeys[shardID] = append(validatorsPubKeys[shardID], shardWaiting[i].PubKey())
		}
	}

	return validatorsPubKeys, nil
}

// GetAllLeavingValidatorsPublicKeys will return all leaving validators public keys for all shards
func (ihgs *indexHashedNodesCoordinator) GetAllLeavingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	for shardID, shardLeaving := range nodesConfig.leavingMap {
		for i := 0; i < len(shardLeaving); i++ {
			validatorsPubKeys[shardID] = append(validatorsPubKeys[shardID], shardLeaving[i].PubKey())
		}
	}

	return validatorsPubKeys, nil
}

// GetValidatorsIndexes will return validators indexes for a block
func (ihgs *indexHashedNodesCoordinator) GetValidatorsIndexes(
	publicKeys []string,
	epoch uint32,
) ([]uint64, error) {
	signersIndexes := make([]uint64, 0)

	validatorsPubKeys, err := ihgs.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		return nil, err
	}

	ihgs.mutNodesConfig.RLock()
	nodesConfig := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	for _, pubKey := range publicKeys {
		for index, value := range validatorsPubKeys[nodesConfig.shardID] {
			if bytes.Equal([]byte(pubKey), value) {
				signersIndexes = append(signersIndexes, uint64(index))
			}
		}
	}

	if len(publicKeys) != len(signersIndexes) {
		strHaving := "having the following keys: \n"
		for index, value := range validatorsPubKeys[nodesConfig.shardID] {
			strHaving += fmt.Sprintf(" index %d  key %s\n", index, logger.DisplayByteSlice(value))
		}

		strNeeded := "needed the following keys: \n"
		for _, pubKey := range publicKeys {
			strNeeded += fmt.Sprintf(" key %s\n", logger.DisplayByteSlice([]byte(pubKey)))
		}

		log.Error("public keys not found\n"+strHaving+"\n"+strNeeded+"\n",
			"len pubKeys", len(publicKeys),
			"len signers", len(signersIndexes),
		)

		return nil, ErrInvalidNumberPubKeys
	}

	return signersIndexes, nil
}

// EpochStartPrepare is called when an epoch start event is observed, but not yet confirmed/committed.
// Some components may need to do some initialisation on this event
func (ihgs *indexHashedNodesCoordinator) EpochStartPrepare(metaHdr data.HeaderHandler, body data.BodyHandler) {
	if !metaHdr.IsStartOfEpochBlock() {
		log.Error("could not process EpochStartPrepare on nodesCoordinator - not epoch start block")
		return
	}

	if _, ok := metaHdr.(*block.MetaBlock); !ok {
		log.Error("could not process EpochStartPrepare on nodesCoordinator - not metaBlock")
		return
	}

	randomness := metaHdr.GetPrevRandSeed()
	newEpoch := metaHdr.GetEpoch()

	if check.IfNil(body) && newEpoch == ihgs.currentEpoch {
		log.Debug("nil body provided for epoch start prepare, it is normal in case of revertStateToBlock")
		return
	}

	allValidatorInfo, err := createValidatorInfoFromBody(body, ihgs.marshalizer, ihgs.numTotalEligible)
	if err != nil {
		log.Error("could not create validator info from body - do nothing on nodesCoordinator epochStartPrepare")
		return
	}

	// TODO: compare with previous nodesConfig if exists
	newNodesConfig, err := ihgs.computeNodesConfigFromList(allValidatorInfo)
	if err != nil {
		log.Error("could not compute nodes config from list - do nothing on nodesCoordinator epochStartPrepare")
		return
	}

	ihgs.mutNodesConfig.RLock()
	previousConfig := ihgs.nodesConfig[ihgs.currentEpoch]
	if previousConfig != nil && previousConfig.nbShards != newNodesConfig.nbShards {
		log.Warn("number of shards does not match",
			"previous epoch", ihgs.currentEpoch,
			"previous number of shards", previousConfig.nbShards,
			"new epoch", newEpoch,
			"new number of shards", newNodesConfig.nbShards)
	}
	ihgs.mutNodesConfig.RUnlock()

	additionalLeavingMap, err := ihgs.nodesCoordinatorHelper.ComputeAdditionalLeaving(allValidatorInfo)
	if err != nil {
		log.Error("could not compute additionalLeaving Nodes  - do nothing on nodesCoordinator epochStartPrepare")
		return
	}

	unStakeLeavingList := ihgs.createSortedListFromMap(newNodesConfig.leavingMap)
	additionalLeavingList := ihgs.createSortedListFromMap(additionalLeavingMap)

	shufflerArgs := ArgsUpdateNodes{
		Eligible:          newNodesConfig.eligibleMap,
		Waiting:           newNodesConfig.waitingMap,
		NewNodes:          newNodesConfig.newList,
		UnStakeLeaving:    unStakeLeavingList,
		AdditionalLeaving: additionalLeavingList,
		Rand:              randomness,
		NbShards:          newNodesConfig.nbShards,
		Epoch:             newEpoch,
	}

	resUpdateNodes, err := ihgs.shuffler.UpdateNodeLists(shufflerArgs)
	if err != nil {
		log.Error("could not compute UpdateNodeLists - do nothing on nodesCoordinator epochStartPrepare", "err", err.Error())
		return
	}

	leavingNodesMap, stillRemainingNodesMap := createActuallyLeavingPerShards(
		newNodesConfig.leavingMap,
		additionalLeavingMap,
		resUpdateNodes.Leaving,
	)

	err = ihgs.setNodesPerShards(resUpdateNodes.Eligible, resUpdateNodes.Waiting, leavingNodesMap, newEpoch)
	if err != nil {
		log.Error("set nodes per shard failed", "error", err.Error())
	}

	ihgs.fillPublicKeyToValidatorMap()
	err = ihgs.saveState(randomness)
	if err != nil {
		log.Error("saving nodes coordinator config failed", "error", err.Error())
	}

	displayNodesConfiguration(
		resUpdateNodes.Eligible,
		resUpdateNodes.Waiting,
		leavingNodesMap,
		stillRemainingNodesMap,
		newNodesConfig.nbShards)

	ihgs.mutSavedStateKey.Lock()
	ihgs.savedStateKey = randomness
	ihgs.mutSavedStateKey.Unlock()

	ihgs.consensusGroupCacher.Clear()
}

func (ihgs *indexHashedNodesCoordinator) fillPublicKeyToValidatorMap() {
	ihgs.mutNodesConfig.Lock()
	defer ihgs.mutNodesConfig.Unlock()

	index := 0
	epochList := make([]uint32, len(ihgs.nodesConfig))
	mapAllValidators := make(map[uint32]map[string]*validatorWithShardID)
	for epoch, epochConfig := range ihgs.nodesConfig {
		epochConfig.mutNodesMaps.RLock()
		mapAllValidators[epoch] = ihgs.createPublicKeyToValidatorMap(epochConfig.eligibleMap, epochConfig.waitingMap)
		epochConfig.mutNodesMaps.RUnlock()

		epochList[index] = epoch
		index++
	}

	sort.Slice(epochList, func(i, j int) bool {
		return epochList[i] < epochList[j]
	})

	ihgs.publicKeyToValidatorMap = make(map[string]*validatorWithShardID)
	for _, epoch := range epochList {
		validatorsForEpoch := mapAllValidators[epoch]
		for pubKey, vInfo := range validatorsForEpoch {
			ihgs.publicKeyToValidatorMap[pubKey] = vInfo
		}
	}
}

func (ihgs *indexHashedNodesCoordinator) createSortedListFromMap(validatorsMap map[uint32][]Validator) []Validator {
	sortedList := make([]Validator, 0)
	for _, validators := range validatorsMap {
		sortedList = append(sortedList, validators...)
	}
	sort.Sort(validatorList(sortedList))
	return sortedList
}

// GetChance will return default chance
func (ihgs *indexHashedNodesCoordinator) GetChance(_ uint32) uint32 {
	return defaultSelectionChances
}

func (ihgs *indexHashedNodesCoordinator) computeNodesConfigFromList(
	validatorInfos []*state.ShardValidatorInfo,
) (*epochNodesConfig, error) {
	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	leavingMap := make(map[uint32][]Validator)
	newNodesList := make([]Validator, 0)

	for _, validatorInfo := range validatorInfos {
		chance := ihgs.nodesCoordinatorHelper.GetChance(validatorInfo.TempRating)
		currentValidator, err := NewValidator(validatorInfo.PublicKey, chance, validatorInfo.Index)
		if err != nil {
			return nil, err
		}

		switch validatorInfo.List {
		case string(core.WaitingList):
			waitingMap[validatorInfo.ShardId] = append(waitingMap[validatorInfo.ShardId], currentValidator)
		case string(core.EligibleList):
			eligibleMap[validatorInfo.ShardId] = append(eligibleMap[validatorInfo.ShardId], currentValidator)
		case string(core.LeavingList):
			log.Debug("leaving node trie", "pk", validatorInfo.PublicKey)
			eligibleMap[validatorInfo.ShardId] = append(eligibleMap[validatorInfo.ShardId], currentValidator)
			leavingMap[validatorInfo.ShardId] = append(leavingMap[validatorInfo.ShardId], currentValidator)
		case string(core.NewList):
			log.Debug("new node registered", "pk", validatorInfo.PublicKey)
			newNodesList = append(newNodesList, currentValidator)
		case string(core.InactiveList):
			log.Debug("inactive validator", "pk", validatorInfo.PublicKey)
		case string(core.JailedList):
			log.Debug("jailed validator", "pk", validatorInfo.PublicKey)
		}
	}

	sort.Sort(validatorList(newNodesList))
	for _, eligibleList := range eligibleMap {
		sort.Sort(validatorList(eligibleList))
	}
	for _, waitingList := range waitingMap {
		sort.Sort(validatorList(waitingList))
	}
	for _, leavingList := range leavingMap {
		sort.Sort(validatorList(leavingList))
	}

	if len(eligibleMap) == 0 {
		return nil, fmt.Errorf("%w eligible map size is zero. No validators found", ErrMapSizeZero)
	}

	nbShards := len(eligibleMap) - 1

	newNodesConfig := &epochNodesConfig{
		eligibleMap: eligibleMap,
		waitingMap:  waitingMap,
		leavingMap:  leavingMap,
		newList:     newNodesList,
		nbShards:    uint32(nbShards),
	}

	return newNodesConfig, nil
}

// EpochStartAction is called upon a start of epoch event.
// NodeCoordinator has to get the nodes assignment to shards using the shuffler.
func (ihgs *indexHashedNodesCoordinator) EpochStartAction(hdr data.HeaderHandler) {
	newEpoch := hdr.GetEpoch()
	epochToRemove := int32(newEpoch) - nodeCoordinatorStoredEpochs
	needToRemove := epochToRemove >= 0
	ihgs.currentEpoch = newEpoch

	err := ihgs.saveState(ihgs.savedStateKey)
	if err != nil {
		log.Error("saving nodes coordinator config failed", "error", err.Error())
	}

	ihgs.mutNodesConfig.Lock()
	if needToRemove {
		for epoch := range ihgs.nodesConfig {
			if epoch <= uint32(epochToRemove) {
				delete(ihgs.nodesConfig, epoch)
			}
		}
	}
	ihgs.mutNodesConfig.Unlock()
}

// NotifyOrder returns the notification order for a start of epoch event
func (ihgs *indexHashedNodesCoordinator) NotifyOrder() uint32 {
	return core.NodesCoordinatorOrder
}

// GetSavedStateKey returns the key for the last nodes coordinator saved state
func (ihgs *indexHashedNodesCoordinator) GetSavedStateKey() []byte {
	ihgs.mutSavedStateKey.RLock()
	key := ihgs.savedStateKey
	ihgs.mutSavedStateKey.RUnlock()

	return key
}

// ShardIdForEpoch returns the nodesCoordinator configured ShardId for specified epoch if epoch configuration exists,
// otherwise error
func (ihgs *indexHashedNodesCoordinator) ShardIdForEpoch(epoch uint32) (uint32, error) {
	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return 0, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	return nodesConfig.shardID, nil
}

// ShuffleOutForEpoch verifies if the shards changed in the new epoch and calls the shuffleOutHandler
func (ihgs *indexHashedNodesCoordinator) ShuffleOutForEpoch(epoch uint32) {
	log.Debug("shuffle out called for", "epoch", epoch)

	ihgs.mutNodesConfig.Lock()
	nodesConfig := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.Unlock()

	if nodesConfig == nil {
		log.Warn("shuffleOutForEpoch failed",
			"epoch", epoch,
			"error", ErrEpochNodesConfigDoesNotExist)
		return
	}

	if isValidator(nodesConfig, ihgs.selfPubKey) {
		err := ihgs.shuffledOutHandler.Process(nodesConfig.shardID)
		if err != nil {
			log.Warn("shuffle out process failed", "err", err)
		}
	}
}

func isValidator(config *epochNodesConfig, pk []byte) bool {
	if config == nil {
		return false
	}

	config.mutNodesMaps.RLock()
	defer config.mutNodesMaps.RUnlock()

	found := false
	found, _ = searchInMap(config.eligibleMap, pk)
	if found {
		return true
	}

	found, _ = searchInMap(config.waitingMap, pk)
	return found
}

func searchInMap(validatorMap map[uint32][]Validator, pk []byte) (bool, uint32) {
	for shardId, validatorsInShard := range validatorMap {
		for _, val := range validatorsInShard {
			if bytes.Equal(val.PubKey(), pk) {
				return true, shardId
			}
		}
	}
	return false, 0
}

// GetConsensusWhitelistedNodes return the whitelisted nodes allowed to send consensus messages, for each of the shards
func (ihgs *indexHashedNodesCoordinator) GetConsensusWhitelistedNodes(
	epoch uint32,
) (map[string]struct{}, error) {
	var err error
	shardEligible := make(map[string]struct{})
	publicKeysPrevEpoch := make(map[uint32][][]byte)
	prevEpochConfigExists := false

	if epoch > ihgs.startEpoch {
		publicKeysPrevEpoch, err = ihgs.GetAllEligibleValidatorsPublicKeys(epoch - 1)
		if err == nil {
			prevEpochConfigExists = true
		} else {
			log.Warn("get consensus whitelisted nodes", "error", err.Error())
		}
	}

	var prevEpochShardId uint32
	if prevEpochConfigExists {
		prevEpochShardId, err = ihgs.ShardIdForEpoch(epoch - 1)
		if err == nil {
			for _, pubKey := range publicKeysPrevEpoch[prevEpochShardId] {
				shardEligible[string(pubKey)] = struct{}{}
			}
		} else {
			log.Trace("not critical error getting shardID for epoch", "epoch", epoch-1, "error", err)
		}
	}

	publicKeysNewEpoch, errGetEligible := ihgs.GetAllEligibleValidatorsPublicKeys(epoch)
	if errGetEligible != nil {
		return nil, errGetEligible
	}

	epochShardId, errShardIdForEpoch := ihgs.ShardIdForEpoch(epoch)
	if errShardIdForEpoch != nil {
		return nil, errShardIdForEpoch
	}

	for _, pubKey := range publicKeysNewEpoch[epochShardId] {
		shardEligible[string(pubKey)] = struct{}{}
	}

	return shardEligible, nil
}

func (ihgs *indexHashedNodesCoordinator) createPublicKeyToValidatorMap(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
) map[string]*validatorWithShardID {
	publicKeyToValidatorMap := make(map[string]*validatorWithShardID)
	for shardId, shardEligible := range eligible {
		for i := 0; i < len(shardEligible); i++ {
			publicKeyToValidatorMap[string(shardEligible[i].PubKey())] = &validatorWithShardID{
				validator: shardEligible[i],
				shardID:   shardId,
			}
		}
	}
	for shardId, shardWaiting := range waiting {
		for i := 0; i < len(shardWaiting); i++ {
			publicKeyToValidatorMap[string(shardWaiting[i].PubKey())] = &validatorWithShardID{
				validator: shardWaiting[i],
				shardID:   shardId,
			}
		}
	}

	return publicKeyToValidatorMap
}

func (ihgs *indexHashedNodesCoordinator) computeShardForSelfPublicKey(nodesConfig *epochNodesConfig) uint32 {
	pubKey := ihgs.selfPubKey
	selfShard := ihgs.shardIDAsObserver
	epNodesConfig, ok := ihgs.nodesConfig[ihgs.currentEpoch]
	if ok {
		log.Trace("computeShardForSelfPublicKey found existing config",
			"shard", epNodesConfig.shardID,
		)
		selfShard = epNodesConfig.shardID
	}

	found, shardId := searchInMap(nodesConfig.eligibleMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in eligible",
			"epoch", ihgs.currentEpoch,
			"shard", shardId,
			"validator PK", pubKey,
		)
		return shardId
	}

	found, shardId = searchInMap(nodesConfig.waitingMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in waiting",
			"epoch", ihgs.currentEpoch,
			"shard", shardId,
			"validator PK", pubKey,
		)
		return shardId
	}

	found, shardId = searchInMap(nodesConfig.leavingMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in leaving",
			"epoch", ihgs.currentEpoch,
			"shard", shardId,
			"validator PK", pubKey,
		)
		return shardId
	}

	log.Trace("computeShardForSelfPublicKey returned default",
		"shard", selfShard,
	)
	return selfShard
}

// ConsensusGroupSize returns the consensus group size for a specific shard
func (ihgs *indexHashedNodesCoordinator) ConsensusGroupSize(
	shardID uint32,
) int {
	if shardID == core.MetachainShardId {
		return ihgs.metaConsensusGroupSize
	}

	return ihgs.shardConsensusGroupSize
}

// GetNumTotalEligible returns the number of total eligible accross all shards from current setup
func (ihgs *indexHashedNodesCoordinator) GetNumTotalEligible() uint64 {
	return ihgs.numTotalEligible
}

// GetOwnPublicKey will return current node public key  for block sign
func (ihgs *indexHashedNodesCoordinator) GetOwnPublicKey() []byte {
	return ihgs.selfPubKey
}

// IsInterfaceNil returns true if there is no value under the interface
func (ihgs *indexHashedNodesCoordinator) IsInterfaceNil() bool {
	return ihgs == nil
}

// createSelectors creates the consensus group selectors for each shard
// Not concurrent safe, needs to be called under mutex
func (ihgs *indexHashedNodesCoordinator) createSelectors(
	nodesConfig *epochNodesConfig,
) (map[uint32]RandomSelector, error) {
	var err error
	var weights []uint32

	selectors := make(map[uint32]RandomSelector)
	// weights for validators are computed according to each validator rating
	for shard, vList := range nodesConfig.eligibleMap {
		log.Debug("create selectors", "shard", shard)
		weights, err = ihgs.nodesCoordinatorHelper.ValidatorsWeights(vList)
		if err != nil {
			return nil, err
		}

		selectors[shard], err = NewSelectorExpandedList(weights, ihgs.hasher)
		if err != nil {
			return nil, err
		}
	}

	return selectors, nil
}

// ValidatorsWeights returns the weights/chances for each of the given validators
func (ihgs *indexHashedNodesCoordinator) ValidatorsWeights(validators []Validator) ([]uint32, error) {
	weights := make([]uint32, len(validators))
	for i := range validators {
		weights[i] = defaultSelectionChances
	}

	return weights, nil
}

func createActuallyLeavingPerShards(
	unstakeLeaving map[uint32][]Validator,
	additionalLeaving map[uint32][]Validator,
	leaving []Validator,
) (map[uint32][]Validator, map[uint32][]Validator) {
	actuallyLeaving := make(map[uint32][]Validator)
	actuallyRemaining := make(map[uint32][]Validator)
	processedValidatorsMap := make(map[string]bool)

	computeActuallyLeaving(unstakeLeaving, leaving, actuallyLeaving, actuallyRemaining, processedValidatorsMap)
	computeActuallyLeaving(additionalLeaving, leaving, actuallyLeaving, actuallyRemaining, processedValidatorsMap)

	return actuallyLeaving, actuallyRemaining
}

func computeActuallyLeaving(
	unstakeLeaving map[uint32][]Validator,
	leaving []Validator,
	actuallyLeaving map[uint32][]Validator,
	actuallyRemaining map[uint32][]Validator,
	processedValidatorsMap map[string]bool,
) {
	sortedShardIds := sortKeys(unstakeLeaving)
	for _, shardId := range sortedShardIds {
		leavingValidatorsPerShard := unstakeLeaving[shardId]
		for _, v := range leavingValidatorsPerShard {
			if processedValidatorsMap[string(v.PubKey())] {
				continue
			}
			processedValidatorsMap[string(v.PubKey())] = true
			found := false
			for _, leavingValidator := range leaving {
				if bytes.Equal(v.PubKey(), leavingValidator.PubKey()) {
					found = true
					break
				}
			}
			if found {
				actuallyLeaving[shardId] = append(actuallyLeaving[shardId], v)
			} else {
				actuallyRemaining[shardId] = append(actuallyRemaining[shardId], v)
			}
		}
	}
}

func selectValidators(
	selector RandomSelector,
	randomness []byte,
	consensusSize uint32,
	eligibleList []Validator,
) ([]Validator, error) {
	if check.IfNil(selector) {
		return nil, ErrNilRandomSelector
	}
	if len(randomness) == 0 {
		return nil, ErrNilRandomness
	}

	// todo: checks for indexes
	selectedIndexes, err := selector.Select(randomness, consensusSize)
	if err != nil {
		return nil, err
	}

	consensusGroup := make([]Validator, consensusSize)
	for i := range consensusGroup {
		consensusGroup[i] = eligibleList[selectedIndexes[i]]
	}

	displayValidatorsForRandomness(consensusGroup, randomness)

	return consensusGroup, nil
}

// createValidatorInfoFromBody unmarshalls body data to create validator info
func createValidatorInfoFromBody(
	body data.BodyHandler,
	marshalizer marshal.Marshalizer,
	previousTotal uint64,
) ([]*state.ShardValidatorInfo, error) {
	if check.IfNil(body) {
		return nil, ErrNilBlockBody
	}

	blockBody, ok := body.(*block.Body)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	allValidatorInfo := make([]*state.ShardValidatorInfo, 0, previousTotal)
	for _, peerMiniBlock := range blockBody.MiniBlocks {
		if peerMiniBlock.Type != block.PeerBlock {
			continue
		}

		for _, txHash := range peerMiniBlock.TxHashes {
			vid := &state.ShardValidatorInfo{}
			err := marshalizer.Unmarshal(vid, txHash)
			if err != nil {
				return nil, err
			}

			allValidatorInfo = append(allValidatorInfo, vid)
		}
	}

	return allValidatorInfo, nil
}
