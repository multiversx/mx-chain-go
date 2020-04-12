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

const (
	keyFormat               = "%s_%v_%v_%v"
	defaultSelectionChances = uint32(1)
)

// TODO: move this to config parameters
const nodeCoordinatorStoredEpochs = 2

type validatorWithShardID struct {
	validator Validator
	shardID   uint32
}

// TODO: add a parameter for shardID  when acting as observer
type epochNodesConfig struct {
	nbShards                uint32
	shardID                 uint32
	eligibleMap             map[uint32][]Validator
	waitingMap              map[uint32][]Validator
	selectors               map[uint32]RandomSelector
	publicKeyToValidatorMap map[string]*validatorWithShardID
	leavingList             []Validator
	newList                 []Validator
	mutNodesMaps            sync.RWMutex
}

type indexHashedNodesCoordinator struct {
	marshalizer                   marshal.Marshalizer
	hasher                        hashing.Hasher
	shuffler                      NodesShuffler
	epochStartRegistrationHandler EpochStartEventNotifier
	bootStorer                    storage.Storer
	selfPubKey                    []byte
	nodesConfig                   map[uint32]*epochNodesConfig
	mutNodesConfig                sync.RWMutex
	currentEpoch                  uint32
	savedStateKey                 []byte
	mutSavedStateKey              sync.RWMutex
	numTotalEligible              uint64
	shardConsensusGroupSize       int
	metaConsensusGroupSize        int
	nodesCoordinatorHelper        NodesCoordinatorHelper
	consensusGroupCacher          Cacher
	shardIDAsObserver             uint32
	loadingFromDisk               atomic.Value
	shuffledOutHandler            ShuffledOutHandler
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
	}

	ihgs.loadingFromDisk.Store(false)

	ihgs.nodesCoordinatorHelper = ihgs
	err = ihgs.setNodesPerShards(arguments.EligibleNodes, arguments.WaitingNodes, nil, arguments.Epoch)
	if err != nil {
		return nil, err
	}

	err = ihgs.saveState(ihgs.savedStateKey)
	if err != nil {
		log.Error("saving initial nodes coordinator config failed",
			"error", err.Error())
	}

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
	leaving []Validator,
	epoch uint32,
) error {
	ihgs.mutNodesConfig.Lock()
	defer ihgs.mutNodesConfig.Unlock()

	nodesConfig, ok := ihgs.nodesConfig[epoch]
	if !ok {
		nodesConfig = &epochNodesConfig{}
	}

	nodesConfig.mutNodesMaps.Lock()
	defer nodesConfig.mutNodesMaps.Unlock()

	if eligible == nil || waiting == nil {
		return ErrNilInputNodesMap
	}

	nodesConfig.leavingList = make([]Validator, 0, len(leaving))
	nodesConfig.leavingList = append(nodesConfig.leavingList, leaving...)

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
	nodesConfig.publicKeyToValidatorMap = ihgs.createPublicKeyToValidatorMap(eligible, waiting)
	nodesConfig.shardID = ihgs.computeShardForSelfPublicKey(nodesConfig)
	nodesConfig.selectors, err = ihgs.createSelectors(nodesConfig)
	if err != nil {
		return err
	}

	shardIDForSelfPublicKey := ihgs.computeShardForSelfPublicKey(nodesConfig)
	nodesConfig.shardID = shardIDForSelfPublicKey
	ihgs.nodesConfig[epoch] = nodesConfig
	ihgs.numTotalEligible = numTotalEligible

	return ihgs.shuffledOutHandler.Process(shardIDForSelfPublicKey)
}

// ComputeLeaving - computes leaving validators
func (ihgs *indexHashedNodesCoordinator) ComputeLeaving(allValidators []*state.ShardValidatorInfo) ([]Validator, error) {
	leavingList := make([]Validator, 0)
	for _, vInfo := range allValidators {
		if vInfo.List == string(core.LeavingList) {
			val, err := NewValidator(vInfo.PublicKey, ihgs.GetChance(vInfo.TempRating), vInfo.Index)
			if err != nil {
				return nil, err
			}

			leavingList = append(leavingList, val)
		}
	}
	return leavingList, nil
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
			return nil, ErrInvalidShardId
		}
		selector = nodesConfig.selectors[shardID]
		eligibleList = nodesConfig.eligibleMap[shardID]
	}
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, ErrEpochNodesConfigDoesNotExist
	}

	key := []byte(fmt.Sprintf(keyFormat, string(randomness), round, shardID, epoch))
	validators := ihgs.searchConsensusForKey(key)
	if validators != nil {
		return validators, nil
	}

	consensusSize := ihgs.ConsensusGroupSize(shardID)
	randomness = []byte(fmt.Sprintf("%d-%s", round, randomness))

	log.Debug("ComputeValidatorsGroup",
		"randomness", randomness,
		"consensus size", consensusSize,
		"eligible list length", len(eligibleList))

	tempList, err := selectValidators(selector, randomness, uint32(consensusSize), eligibleList)
	if err != nil {
		return nil, err
	}

	ihgs.consensusGroupCacher.Put(key, tempList)

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
func (ihgs *indexHashedNodesCoordinator) GetValidatorWithPublicKey(
	publicKey []byte,
	epoch uint32,
) (Validator, uint32, error) {
	if len(publicKey) == 0 {
		return nil, 0, ErrNilPubKey
	}
	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, 0, ErrEpochNodesConfigDoesNotExist
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	v, ok := nodesConfig.publicKeyToValidatorMap[string(publicKey)]
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
		return nil, ErrEpochNodesConfigDoesNotExist
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	for shardID, shardEligible := range nodesConfig.eligibleMap {
		for i := 0; i < len(shardEligible); i++ {
			validatorsPubKeys[shardID] = append(validatorsPubKeys[shardID], nodesConfig.eligibleMap[shardID][i].PubKey())
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
		return nil, ErrEpochNodesConfigDoesNotExist
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	for shardID, shardEligible := range nodesConfig.waitingMap {
		for i := 0; i < len(shardEligible); i++ {
			validatorsPubKeys[shardID] = append(validatorsPubKeys[shardID], nodesConfig.waitingMap[shardID][i].PubKey())
		}
	}

	return validatorsPubKeys, nil
}

// GetAllLeavingValidatorsPublicKeys will return all leaving validators public keys for all shards
func (ihgs *indexHashedNodesCoordinator) GetAllLeavingValidatorsPublicKeys(epoch uint32) ([][]byte, error) {
	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, ErrEpochNodesConfigDoesNotExist
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	leavingPubKeys := make([][]byte, 0, len(nodesConfig.leavingList))

	for _, leaving := range nodesConfig.leavingList {
		leavingPubKeys = append(leavingPubKeys, leaving.PubKey())
	}

	return leavingPubKeys, nil
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

	allValidatorInfo, err := createValidatorInfoFromBody(body, ihgs.marshalizer, ihgs.numTotalEligible)
	if err != nil {
		log.Error("could not create validator info from body - do nothing on nodesCoordinator epochStartPrepare")
		return
	}

	newNodesConfig, err := ihgs.computeNodesConfigFromList(allValidatorInfo)
	if err != nil {
		log.Error("could not compute nodes config from list - do nothing on nodesCoordinator epochStartPrepare")
		return
	}

	shufflerArgs := ArgsUpdateNodes{
		Eligible: newNodesConfig.eligibleMap,
		Waiting:  newNodesConfig.waitingMap,
		NewNodes: newNodesConfig.newList,
		Leaving:  newNodesConfig.leavingList,
		Rand:     randomness,
		NbShards: newNodesConfig.nbShards,
	}

	eligibleMap, waitingMap, stillRemaining := ihgs.shuffler.UpdateNodeLists(shufflerArgs)

	actualLeaving := ComputeActuallyLeaving(newNodesConfig.leavingList, stillRemaining)
	err = ihgs.setNodesPerShards(eligibleMap, waitingMap, actualLeaving, newEpoch)
	if err != nil {
		log.Error("set nodes per shard failed", "error", err.Error())
	}

	err = ihgs.saveState(randomness)
	if err != nil {
		log.Error("saving nodes coordinator config failed", "error", err.Error())
	}

	displayNodesConfiguration(eligibleMap, waitingMap, newNodesConfig.leavingList, stillRemaining, newNodesConfig.nbShards)

	ihgs.mutSavedStateKey.Lock()
	ihgs.savedStateKey = randomness
	ihgs.mutSavedStateKey.Unlock()
}

// GetChance will return default chance
func (ihgs *indexHashedNodesCoordinator) GetChance(_ uint32) uint32 {
	return defaultSelectionChances
}

func (ihgs *indexHashedNodesCoordinator) computeNodesConfigFromList(
	validatorInfos []*state.ShardValidatorInfo,
) (*epochNodesConfig, error) {

	leaving, err := ihgs.nodesCoordinatorHelper.ComputeLeaving(validatorInfos)
	if err != nil {
		return nil, err
	}

	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	newNodesList := make([]Validator, 0)

	for _, validatorInfo := range validatorInfos {
		chance := ihgs.nodesCoordinatorHelper.GetChance(validatorInfo.TempRating)
		validator, err := NewValidator(validatorInfo.PublicKey, chance, validatorInfo.Index)
		if err != nil {
			return nil, err
		}

		switch validatorInfo.List {
		case string(core.WaitingList):
			waitingMap[validatorInfo.ShardId] = append(waitingMap[validatorInfo.ShardId], validator)
		case string(core.EligibleList):
			eligibleMap[validatorInfo.ShardId] = append(eligibleMap[validatorInfo.ShardId], validator)
		case string(core.NewList):
			newNodesList = append(newNodesList, validator)
		}
	}

	sort.Slice(leaving, func(i, j int) bool {
		return leaving[i].Index() > leaving[j].Index()
	})

	sort.Slice(newNodesList, func(i, j int) bool {
		return newNodesList[i].Index() > newNodesList[j].Index()
	})

	for _, eligibleList := range eligibleMap {
		sort.Slice(eligibleList, func(i, j int) bool {
			return eligibleList[i].Index() > eligibleList[j].Index()
		})
	}

	for _, waitingList := range waitingMap {
		sort.Slice(waitingList, func(i, j int) bool {
			return waitingList[i].Index() > waitingList[j].Index()
		})
	}

	newNodesConfig := &epochNodesConfig{
		eligibleMap: eligibleMap,
		waitingMap:  waitingMap,
		leavingList: leaving,
		newList:     newNodesList,
		nbShards:    uint32(len(eligibleMap)),
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
		return 0, ErrEpochNodesConfigDoesNotExist
	}

	return nodesConfig.shardID, nil
}

// GetConsensusWhitelistedNodes return the whitelisted nodes allowed to send consensus messages, for each of the shards
func (ihgs *indexHashedNodesCoordinator) GetConsensusWhitelistedNodes(
	epoch uint32,
) (map[string]struct{}, error) {
	var err error
	shardEligible := make(map[string]struct{})
	publicKeysPrevEpoch := make(map[uint32][][]byte)
	prevEpochConfigExists := false

	if epoch > 0 {
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

	for shard, validators := range nodesConfig.eligibleMap {
		for _, v := range validators {
			if bytes.Equal(v.PubKey(), pubKey) {
				log.Trace("computeShardForSelfPublicKey found validator in eligible",
					"shard", shard,
					"validator PK", v,
				)

				return shard
			}
		}
	}

	for shard, validators := range nodesConfig.waitingMap {
		for _, v := range validators {
			if bytes.Equal(v.PubKey(), pubKey) {
				log.Trace("computeShardForSelfPublicKey found validator in waiting",
					"shard", shard,
					"validator PK", v,
				)

				return shard
			}
		}
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
