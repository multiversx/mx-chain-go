package nodesCoordinator

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	atomicFlags "github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
)

var _ NodesCoordinatorLite = (*IndexHashedNodesCoordinatorLite)(nil)
var _ PublicKeysSelector = (*IndexHashedNodesCoordinatorLite)(nil)

const (
	keyFormat               = "%s_%v_%v_%v"
	DefaultSelectionChances = uint32(1)
)

// TODO: move this to config parameters
const nodesCoordinatorStoredEpochs = 4

type ValidatorWithShardID struct {
	Validator Validator
	ShardID   uint32
}

// TODO: add a parameter for shardID  when acting as observer
type EpochNodesConfig struct {
	NbShards     uint32
	ShardID      uint32
	EligibleMap  map[uint32][]Validator
	WaitingMap   map[uint32][]Validator
	Selectors    map[uint32]RandomSelector
	LeavingMap   map[uint32][]Validator
	NewList      []Validator
	MutNodesMaps sync.RWMutex
}

type IndexHashedNodesCoordinatorLite struct {
	shardIDAsObserver             uint32
	currentEpoch                  uint32
	shardConsensusGroupSize       int
	metaConsensusGroupSize        int
	numTotalEligible              uint64
	selfPubKey                    []byte
	hasher                        hashing.Hasher
	epochStartRegistrationHandler EpochStartEventNotifier
	nodesConfig                   map[uint32]*EpochNodesConfig
	mutNodesConfig                sync.RWMutex
	nodesCoordinatorHelper        NodesCoordinatorHelper
	consensusGroupCacher          Cacher
	startEpoch                    uint32
	publicKeyToValidatorMap       map[string]*ValidatorWithShardID
	waitingListFixEnableEpoch     uint32
	isFullArchive                 bool
	chanStopNode                  chan endProcess.ArgEndProcess
	flagWaitingListFix            atomicFlags.Flag
	nodeTypeProvider              NodeTypeProviderHandler
}

// NewIndexHashedNodesCoordinatorLite creates a new index hashed group selector
func NewIndexHashedNodesCoordinatorLite(arguments ArgNodesCoordinatorLite) (*IndexHashedNodesCoordinatorLite, error) {
	err := checkArguments(arguments)
	if err != nil {
		return nil, err
	}

	nodesConfig := make(map[uint32]*EpochNodesConfig, nodesCoordinatorStoredEpochs)

	nodesConfig[arguments.Epoch] = &EpochNodesConfig{
		NbShards:    arguments.NbShards,
		ShardID:     arguments.ShardIDAsObserver,
		EligibleMap: make(map[uint32][]Validator),
		WaitingMap:  make(map[uint32][]Validator),
		Selectors:   make(map[uint32]RandomSelector),
		LeavingMap:  make(map[uint32][]Validator),
		NewList:     make([]Validator, 0),
	}

	ihgs := &IndexHashedNodesCoordinatorLite{
		hasher:                        arguments.Hasher,
		epochStartRegistrationHandler: arguments.EpochStartNotifier,
		selfPubKey:                    arguments.SelfPublicKey,
		nodesConfig:                   nodesConfig,
		currentEpoch:                  arguments.Epoch,
		shardConsensusGroupSize:       arguments.ShardConsensusGroupSize,
		metaConsensusGroupSize:        arguments.MetaConsensusGroupSize,
		consensusGroupCacher:          arguments.ConsensusGroupCache,
		shardIDAsObserver:             arguments.ShardIDAsObserver,
		startEpoch:                    arguments.StartEpoch,
		publicKeyToValidatorMap:       make(map[string]*ValidatorWithShardID),
		waitingListFixEnableEpoch:     arguments.WaitingListFixEnabledEpoch,
		chanStopNode:                  arguments.ChanStopNode,
		nodeTypeProvider:              arguments.NodeTypeProvider,
		isFullArchive:                 arguments.IsFullArchive,
	}
	log.Debug("indexHashedNodesCoordinator: enable epoch for waiting waiting list", "epoch", ihgs.waitingListFixEnableEpoch)

	ihgs.nodesCoordinatorHelper = ihgs
	err = ihgs.SetNodesPerShards(arguments.EligibleNodes, arguments.WaitingNodes, nil, arguments.Epoch)
	if err != nil {
		return nil, err
	}

	ihgs.fillPublicKeyToValidatorMap()
	// err = ihgs.saveState(ihgs.savedStateKey)
	// if err != nil {
	// 	log.Error("saving initial nodes coordinator config failed",
	// 		"error", err.Error())
	// }
	log.Info("new nodes config is set for epoch", "epoch", arguments.Epoch)
	currentNodesConfig := ihgs.nodesConfig[arguments.Epoch]
	if currentNodesConfig == nil {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, arguments.Epoch)
	}

	currentConfig := nodesConfig[arguments.Epoch]
	if currentConfig == nil {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, arguments.Epoch)
	}

	DisplayNodesConfiguration(
		currentConfig.EligibleMap,
		currentConfig.WaitingMap,
		currentConfig.LeavingMap,
		make(map[uint32][]Validator),
		currentConfig.NbShards)

	ihgs.epochStartRegistrationHandler.RegisterHandler(ihgs)

	return ihgs, nil
}

func checkArguments(arguments ArgNodesCoordinatorLite) error {
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
	if check.IfNil(arguments.BootStorer) {
		return ErrNilBootStorer
	}
	if check.IfNilReflect(arguments.ConsensusGroupCache) {
		return ErrNilCacher
	}
	if check.IfNil(arguments.NodeTypeProvider) {
		return ErrNilNodeTypeProvider
	}
	if nil == arguments.ChanStopNode {
		return ErrNilNodeStopChannel
	}

	return nil
}

// SetNodesPerShards loads the distribution of nodes per shard into the nodes management component
func (ihgs *IndexHashedNodesCoordinatorLite) SetNodesPerShards(
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
		nodesConfig = &EpochNodesConfig{}
	}

	nodesConfig.MutNodesMaps.Lock()
	defer nodesConfig.MutNodesMaps.Unlock()

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
	var isValidator bool
	// nbShards holds number of shards without meta
	nodesConfig.NbShards = uint32(len(eligible) - 1)
	nodesConfig.EligibleMap = eligible
	nodesConfig.WaitingMap = waiting
	nodesConfig.LeavingMap = leaving
	nodesConfig.ShardID, isValidator = ihgs.computeShardForSelfPublicKey(nodesConfig)
	nodesConfig.Selectors, err = ihgs.CreateSelectors(nodesConfig)
	if err != nil {
		return err
	}

	ihgs.nodesConfig[epoch] = nodesConfig
	ihgs.numTotalEligible = numTotalEligible
	ihgs.setNodeType(isValidator)

	if ihgs.isFullArchive && isValidator {
		ihgs.chanStopNode <- endProcess.ArgEndProcess{
			Reason:      common.WrongConfiguration,
			Description: ErrValidatorCannotBeFullArchive.Error(),
		}

		return nil
	}

	return nil
}

func (ihgs *IndexHashedNodesCoordinatorLite) setNodeType(isValidator bool) {
	if isValidator {
		ihgs.nodeTypeProvider.SetType(core.NodeTypeValidator)
		return
	}

	ihgs.nodeTypeProvider.SetType(core.NodeTypeObserver)
}

// ComputeAdditionalLeaving - computes extra leaving validators based on computation at the start of epoch
func (ihgs *IndexHashedNodesCoordinatorLite) ComputeAdditionalLeaving(_ []*state.ShardValidatorInfo) (map[uint32][]Validator, error) {
	return make(map[uint32][]Validator), nil
}

// ComputeConsensusGroup will generate a list of validators based on the the eligible list
// and each eligible validator weight/chance
func (ihgs *IndexHashedNodesCoordinatorLite) ComputeConsensusGroup(
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
		if shardID >= nodesConfig.NbShards && shardID != core.MetachainShardId {
			log.Warn("shardID is not ok", "shardID", shardID, "nbShards", nodesConfig.NbShards)
			ihgs.mutNodesConfig.RUnlock()
			return nil, ErrInvalidShardId
		}
		selector = nodesConfig.Selectors[shardID]
		eligibleList = nodesConfig.EligibleMap[shardID]
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

func (ihgs *IndexHashedNodesCoordinatorLite) searchConsensusForKey(key []byte) []Validator {
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
func (ihgs *IndexHashedNodesCoordinatorLite) GetValidatorWithPublicKey(publicKey []byte) (Validator, uint32, error) {
	if len(publicKey) == 0 {
		return nil, 0, ErrNilPubKey
	}
	ihgs.mutNodesConfig.RLock()
	v, ok := ihgs.publicKeyToValidatorMap[string(publicKey)]
	ihgs.mutNodesConfig.RUnlock()
	if ok {
		return v.Validator, v.ShardID, nil
	}

	return nil, 0, ErrValidatorNotFound
}

// GetConsensusValidatorsPublicKeys calculates the validators consensus group for a specific shard, randomness and round number,
// returning their public keys
func (ihgs *IndexHashedNodesCoordinatorLite) GetConsensusValidatorsPublicKeys(
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
func (ihgs *IndexHashedNodesCoordinatorLite) GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	nodesConfig.MutNodesMaps.RLock()
	defer nodesConfig.MutNodesMaps.RUnlock()

	for shardID, shardEligible := range nodesConfig.EligibleMap {
		for i := 0; i < len(shardEligible); i++ {
			validatorsPubKeys[shardID] = append(validatorsPubKeys[shardID], shardEligible[i].PubKey())
		}
	}

	return validatorsPubKeys, nil
}

// GetAllWaitingValidatorsPublicKeys will return all validators public keys for all shards
// no need
func (ihgs *IndexHashedNodesCoordinatorLite) GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	nodesConfig.MutNodesMaps.RLock()
	defer nodesConfig.MutNodesMaps.RUnlock()

	for shardID, shardWaiting := range nodesConfig.WaitingMap {
		for i := 0; i < len(shardWaiting); i++ {
			validatorsPubKeys[shardID] = append(validatorsPubKeys[shardID], shardWaiting[i].PubKey())
		}
	}

	return validatorsPubKeys, nil

	// 	panic("method not implemented in this lite version")
}

// GetAllLeavingValidatorsPublicKeys will return all leaving validators public keys for all shards
// no need
func (ihgs *IndexHashedNodesCoordinatorLite) GetAllLeavingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	nodesConfig.MutNodesMaps.RLock()
	defer nodesConfig.MutNodesMaps.RUnlock()

	for shardID, shardLeaving := range nodesConfig.LeavingMap {
		for i := 0; i < len(shardLeaving); i++ {
			validatorsPubKeys[shardID] = append(validatorsPubKeys[shardID], shardLeaving[i].PubKey())
		}
	}

	return validatorsPubKeys, nil

	// 	panic("method not implemented in this lite version")
}

// GetValidatorsIndexes will return validators indexes for a block
func (ihgs *IndexHashedNodesCoordinatorLite) GetValidatorsIndexes(
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
		for index, value := range validatorsPubKeys[nodesConfig.ShardID] {
			if bytes.Equal([]byte(pubKey), value) {
				signersIndexes = append(signersIndexes, uint64(index))
			}
		}
	}

	if len(publicKeys) != len(signersIndexes) {
		strHaving := "having the following keys: \n"
		for index, value := range validatorsPubKeys[nodesConfig.ShardID] {
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
// no need - set new epoch validators instead
func (ihgs *IndexHashedNodesCoordinatorLite) EpochStartPrepare(metaHdr data.HeaderHandler, body data.BodyHandler) {
	//if !metaHdr.IsStartOfEpochBlock() {
	//	log.Error("could not process EpochStartPrepare on nodesCoordinator - not epoch start block")
	//	return
	//}

	//if _, ok := metaHdr.(*block.MetaBlock); !ok {
	//	log.Error("could not process EpochStartPrepare on nodesCoordinator - not metaBlock")
	//	return
	//}

	//randomness := metaHdr.GetPrevRandSeed()
	//newEpoch := metaHdr.GetEpoch()

	//if check.IfNil(body) && newEpoch == ihgs.currentEpoch {
	//	log.Debug("nil body provided for epoch start prepare, it is normal in case of revertStateToBlock")
	//	return
	//}

	//allValidatorInfo, err := createValidatorInfoFromBody(body, ihgs.marshalizer, ihgs.numTotalEligible)
	//if err != nil {
	//	log.Error("could not create validator info from body - do nothing on nodesCoordinator epochStartPrepare")
	//	return
	//}

	//ihgs.updateEpochFlags(newEpoch)

	//ihgs.mutNodesConfig.RLock()
	//previousConfig := ihgs.nodesConfig[ihgs.currentEpoch]
	//if previousConfig == nil {
	//	log.Error("previous nodes config is nil")
	//	ihgs.mutNodesConfig.RUnlock()
	//	return
	//}

	////TODO: remove the copy if no changes are done to the maps
	//copiedPrevious := &epochNodesConfig{}
	//copiedPrevious.eligibleMap = copyValidatorMap(previousConfig.eligibleMap)
	//copiedPrevious.waitingMap = copyValidatorMap(previousConfig.waitingMap)
	//copiedPrevious.nbShards = previousConfig.nbShards

	//ihgs.mutNodesConfig.RUnlock()

	//// TODO: compare with previous nodesConfig if exists
	//newNodesConfig, err := ihgs.computeNodesConfigFromList(copiedPrevious, allValidatorInfo)
	//if err != nil {
	//	log.Error("could not compute nodes config from list - do nothing on nodesCoordinator epochStartPrepare")
	//	return
	//}

	//if copiedPrevious.nbShards != newNodesConfig.nbShards {
	//	log.Warn("number of shards does not match",
	//		"previous epoch", ihgs.currentEpoch,
	//		"previous number of shards", copiedPrevious.nbShards,
	//		"new epoch", newEpoch,
	//		"new number of shards", newNodesConfig.nbShards)
	//}

	//additionalLeavingMap, err := ihgs.nodesCoordinatorHelper.ComputeAdditionalLeaving(allValidatorInfo)
	//if err != nil {
	//	log.Error("could not compute additionalLeaving Nodes  - do nothing on nodesCoordinator epochStartPrepare")
	//	return
	//}

	//unStakeLeavingList := ihgs.createSortedListFromMap(newNodesConfig.leavingMap)
	//additionalLeavingList := ihgs.createSortedListFromMap(additionalLeavingMap)

	//shufflerArgs := ArgsUpdateNodes{
	//	Eligible:          newNodesConfig.eligibleMap,
	//	Waiting:           newNodesConfig.waitingMap,
	//	NewNodes:          newNodesConfig.newList,
	//	UnStakeLeaving:    unStakeLeavingList,
	//	AdditionalLeaving: additionalLeavingList,
	//	Rand:              randomness,
	//	NbShards:          newNodesConfig.nbShards,
	//	Epoch:             newEpoch,
	//}

	//resUpdateNodes, err := ihgs.shuffler.UpdateNodeLists(shufflerArgs)
	//if err != nil {
	//	log.Error("could not compute UpdateNodeLists - do nothing on nodesCoordinator epochStartPrepare", "err", err.Error())
	//	return
	//}

	//leavingNodesMap, stillRemainingNodesMap := createActuallyLeavingPerShards(
	//	newNodesConfig.leavingMap,
	//	additionalLeavingMap,
	//	resUpdateNodes.Leaving,
	//)

	//err = ihgs.setNodesPerShards(resUpdateNodes.Eligible, resUpdateNodes.Waiting, leavingNodesMap, newEpoch)
	//if err != nil {
	//	log.Error("set nodes per shard failed", "error", err.Error())
	//}

	//ihgs.fillPublicKeyToValidatorMap()
	//err = ihgs.saveState(randomness)
	//if err != nil {
	//	log.Error("saving nodes coordinator config failed", "error", err.Error())
	//}

	//displayNodesConfiguration(
	//	resUpdateNodes.Eligible,
	//	resUpdateNodes.Waiting,
	//	leavingNodesMap,
	//	stillRemainingNodesMap,
	//	newNodesConfig.nbShards)

	//ihgs.mutSavedStateKey.Lock()
	//ihgs.savedStateKey = randomness
	//ihgs.mutSavedStateKey.Unlock()

	//ihgs.consensusGroupCacher.Clear()

	panic("method not implemented in this lite version")
}

func (ihgs *IndexHashedNodesCoordinatorLite) fillPublicKeyToValidatorMap() {
	ihgs.mutNodesConfig.Lock()
	defer ihgs.mutNodesConfig.Unlock()

	index := 0
	epochList := make([]uint32, len(ihgs.nodesConfig))
	mapAllValidators := make(map[uint32]map[string]*ValidatorWithShardID)
	for epoch, epochConfig := range ihgs.nodesConfig {
		epochConfig.MutNodesMaps.RLock()
		mapAllValidators[epoch] = ihgs.createPublicKeyToValidatorMap(epochConfig.EligibleMap, epochConfig.WaitingMap)
		epochConfig.MutNodesMaps.RUnlock()

		epochList[index] = epoch
		index++
	}

	sort.Slice(epochList, func(i, j int) bool {
		return epochList[i] < epochList[j]
	})

	ihgs.publicKeyToValidatorMap = make(map[string]*ValidatorWithShardID)
	for _, epoch := range epochList {
		validatorsForEpoch := mapAllValidators[epoch]
		for pubKey, vInfo := range validatorsForEpoch {
			ihgs.publicKeyToValidatorMap[pubKey] = vInfo
		}
	}
}

// GetChance will return default chance
func (ihgs *IndexHashedNodesCoordinatorLite) GetChance(_ uint32) uint32 {
	return DefaultSelectionChances
}

// func (ihgs *IndexHashedNodesCoordinatorLite) computeNodesConfigFromList(
// 	previousEpochConfig *EpochNodesConfig,
// 	validatorInfos []*state.ShardValidatorInfo,
// ) (*EpochNodesConfig, error) {
// 	eligibleMap := make(map[uint32][]Validator)
// 	waitingMap := make(map[uint32][]Validator)
// 	leavingMap := make(map[uint32][]Validator)
// 	newNodesList := make([]Validator, 0)

// 	if ihgs.flagWaitingListFix.IsSet() && previousEpochConfig == nil {
// 		return nil, ErrNilPreviousEpochConfig
// 	}

// 	if len(validatorInfos) == 0 {
// 		log.Warn("computeNodesConfigFromList - validatorInfos len is 0")
// 	}

// 	for _, validatorInfo := range validatorInfos {
// 		chance := ihgs.nodesCoordinatorHelper.GetChance(validatorInfo.TempRating)
// 		currentValidator, err := NewValidator(validatorInfo.PublicKey, chance, validatorInfo.Index)
// 		if err != nil {
// 			return nil, err
// 		}

// 		switch validatorInfo.List {
// 		case string(common.WaitingList):
// 			waitingMap[validatorInfo.ShardId] = append(waitingMap[validatorInfo.ShardId], currentValidator)
// 		case string(common.EligibleList):
// 			eligibleMap[validatorInfo.ShardId] = append(eligibleMap[validatorInfo.ShardId], currentValidator)
// 		case string(common.LeavingList):
// 			log.Debug("leaving node validatorInfo", "pk", validatorInfo.PublicKey)
// 			leavingMap[validatorInfo.ShardId] = append(leavingMap[validatorInfo.ShardId], currentValidator)
// 			ihgs.addValidatorToPreviousMap(
// 				previousEpochConfig,
// 				eligibleMap,
// 				waitingMap,
// 				currentValidator,
// 				validatorInfo.ShardId)
// 		case string(common.NewList):
// 			log.Debug("new node registered", "pk", validatorInfo.PublicKey)
// 			newNodesList = append(newNodesList, currentValidator)
// 		case string(common.InactiveList):
// 			log.Debug("inactive validator", "pk", validatorInfo.PublicKey)
// 		case string(common.JailedList):
// 			log.Debug("jailed validator", "pk", validatorInfo.PublicKey)
// 		}
// 	}

// 	sort.Sort(validatorList(newNodesList))
// 	for _, eligibleList := range eligibleMap {
// 		sort.Sort(validatorList(eligibleList))
// 	}
// 	for _, waitingList := range waitingMap {
// 		sort.Sort(validatorList(waitingList))
// 	}
// 	for _, leavingList := range leavingMap {
// 		sort.Sort(validatorList(leavingList))
// 	}

// 	if len(eligibleMap) == 0 {
// 		return nil, fmt.Errorf("%w eligible map size is zero. No validators found", ErrMapSizeZero)
// 	}

// 	nbShards := len(eligibleMap) - 1

// 	newNodesConfig := &EpochNodesConfig{
// 		EligibleMap: eligibleMap,
// 		WaitingMap:  waitingMap,
// 		LeavingMap:  leavingMap,
// 		NewList:     newNodesList,
// 		NbShards:    uint32(nbShards),
// 	}

// 	return newNodesConfig, nil
// }

// func (ihgs *IndexHashedNodesCoordinatorLite) addValidatorToPreviousMap(
// 	previousEpochConfig *EpochNodesConfig,
// 	eligibleMap map[uint32][]Validator,
// 	waitingMap map[uint32][]Validator,
// 	currentValidator Validator,
// 	currentValidatorShardId uint32) {

// 	if !ihgs.flagWaitingListFix.IsSet() {
// 		eligibleMap[currentValidatorShardId] = append(eligibleMap[currentValidatorShardId], currentValidator)
// 		return
// 	}

// 	found, shardId := searchInMap(previousEpochConfig.EligibleMap, currentValidator.PubKey())
// 	if found {
// 		log.Debug("leaving node found in", "list", "eligible", "shardId", shardId)
// 		eligibleMap[shardId] = append(eligibleMap[currentValidatorShardId], currentValidator)
// 		return
// 	}

// 	found, shardId = searchInMap(previousEpochConfig.WaitingMap, currentValidator.PubKey())
// 	if found {
// 		log.Debug("leaving node found in", "list", "waiting", "shardId", shardId)
// 		waitingMap[shardId] = append(waitingMap[currentValidatorShardId], currentValidator)
// 		return
// 	}
// }

// EpochStartAction is called upon a start of epoch event.
// NodeCoordinator has to get the nodes assignment to shards using the shuffler.
// no need
func (ihgs *IndexHashedNodesCoordinatorLite) EpochStartAction(hdr data.HeaderHandler) {
	// newEpoch := hdr.GetEpoch()
	// epochToRemove := int32(newEpoch) - nodesCoordinatorStoredEpochs
	// needToRemove := epochToRemove >= 0
	// ihgs.currentEpoch = newEpoch

	// err := ihgs.saveState(ihgs.savedStateKey)
	// if err != nil {
	// 	log.Error("saving nodes coordinator config failed", "error", err.Error())
	// }

	// ihgs.mutNodesConfig.Lock()
	// if needToRemove {
	// 	for epoch := range ihgs.nodesConfig {
	// 		if epoch <= uint32(epochToRemove) {
	// 			delete(ihgs.nodesConfig, epoch)
	// 		}
	// 	}
	// }
	// ihgs.mutNodesConfig.Unlock()

	panic("method not implemented in this lite version")
}

// NotifyOrder returns the notification order for a start of epoch event
// no need
func (ihgs *IndexHashedNodesCoordinatorLite) NotifyOrder() uint32 {
	// return common.NodesCoordinatorOrder

	panic("method not implemented in this lite version")
}

// GetSavedStateKey returns the key for the last nodes coordinator saved state
// no
func (ihgs *IndexHashedNodesCoordinatorLite) GetSavedStateKey() []byte {
	// ihgs.mutSavedStateKey.RLock()
	// key := ihgs.savedStateKey
	// ihgs.mutSavedStateKey.RUnlock()

	// return key

	panic("method not implemented in this lite version")
}

// ShardIdForEpoch returns the nodesCoordinator configured ShardId for specified epoch if epoch configuration exists,
// otherwise error
func (ihgs *IndexHashedNodesCoordinatorLite) ShardIdForEpoch(epoch uint32) (uint32, error) {
	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return 0, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	return nodesConfig.ShardID, nil
}

// ShuffleOutForEpoch verifies if the shards changed in the new epoch and calls the shuffleOutHandler
// no
func (ihgs *IndexHashedNodesCoordinatorLite) ShuffleOutForEpoch(epoch uint32) {
	// log.Debug("shuffle out called for", "epoch", epoch)

	// ihgs.mutNodesConfig.Lock()
	// nodesConfig := ihgs.nodesConfig[epoch]
	// ihgs.mutNodesConfig.Unlock()

	// if nodesConfig == nil {
	// 	log.Warn("shuffleOutForEpoch failed",
	// 		"epoch", epoch,
	// 		"error", ErrEpochNodesConfigDoesNotExist)
	// 	return
	// }

	// if isValidator(nodesConfig, ihgs.selfPubKey) {
	// 	err := ihgs.shuffledOutHandler.Process(nodesConfig.shardID)
	// 	if err != nil {
	// 		log.Warn("shuffle out process failed", "err", err)
	// 	}
	// }

	panic("method not implemented in this lite version")
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
// no
func (ihgs *IndexHashedNodesCoordinatorLite) GetConsensusWhitelistedNodes(
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

func (ihgs *IndexHashedNodesCoordinatorLite) createPublicKeyToValidatorMap(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
) map[string]*ValidatorWithShardID {
	publicKeyToValidatorMap := make(map[string]*ValidatorWithShardID)
	for shardId, shardEligible := range eligible {
		for i := 0; i < len(shardEligible); i++ {
			publicKeyToValidatorMap[string(shardEligible[i].PubKey())] = &ValidatorWithShardID{
				Validator: shardEligible[i],
				ShardID:   shardId,
			}
		}
	}
	for shardId, shardWaiting := range waiting {
		for i := 0; i < len(shardWaiting); i++ {
			publicKeyToValidatorMap[string(shardWaiting[i].PubKey())] = &ValidatorWithShardID{
				Validator: shardWaiting[i],
				ShardID:   shardId,
			}
		}
	}

	return publicKeyToValidatorMap
}

func (ihgs *IndexHashedNodesCoordinatorLite) computeShardForSelfPublicKey(nodesConfig *EpochNodesConfig) (uint32, bool) {
	pubKey := ihgs.selfPubKey
	selfShard := ihgs.shardIDAsObserver
	epNodesConfig, ok := ihgs.nodesConfig[ihgs.currentEpoch]
	if ok {
		log.Trace("computeShardForSelfPublicKey found existing config",
			"shard", epNodesConfig.ShardID,
		)
		selfShard = epNodesConfig.ShardID
	}

	found, shardId := searchInMap(nodesConfig.EligibleMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in eligible",
			"epoch", ihgs.currentEpoch,
			"shard", shardId,
			"validator PK", pubKey,
		)
		return shardId, true
	}

	found, shardId = searchInMap(nodesConfig.WaitingMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in waiting",
			"epoch", ihgs.currentEpoch,
			"shard", shardId,
			"validator PK", pubKey,
		)
		return shardId, true
	}

	found, shardId = searchInMap(nodesConfig.LeavingMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in leaving",
			"epoch", ihgs.currentEpoch,
			"shard", shardId,
			"validator PK", pubKey,
		)
		return shardId, true
	}

	log.Trace("computeShardForSelfPublicKey returned default",
		"shard", selfShard,
	)
	return selfShard, false
}

// ConsensusGroupSize returns the consensus group size for a specific shard
func (ihgs *IndexHashedNodesCoordinatorLite) ConsensusGroupSize(
	shardID uint32,
) int {
	if shardID == core.MetachainShardId {
		return ihgs.metaConsensusGroupSize
	}

	return ihgs.shardConsensusGroupSize
}

// GetNumTotalEligible returns the number of total eligible accross all shards from current setup
func (ihgs *IndexHashedNodesCoordinatorLite) GetNumTotalEligible() uint64 {
	return ihgs.numTotalEligible
}

// GetCurrentEpoch
func (ihgs *IndexHashedNodesCoordinatorLite) GetCurrentEpoch() uint32 {
	return ihgs.currentEpoch
}

func (ihgs *IndexHashedNodesCoordinatorLite) GetStartEpoch() uint32 {
	return ihgs.startEpoch
}

func (ihgs *IndexHashedNodesCoordinatorLite) GetSelfPubKey() []byte {
	return ihgs.selfPubKey
}

func (ihgs *IndexHashedNodesCoordinatorLite) SetCurrentEpoch(epoch uint32) {
	ihgs.currentEpoch = epoch
}

func (ihgs *IndexHashedNodesCoordinatorLite) GetMutNodesConfig() sync.RWMutex {
	return ihgs.mutNodesConfig
}

func (ihgs *IndexHashedNodesCoordinatorLite) GetNodesConfig() map[uint32]*EpochNodesConfig {
	return ihgs.nodesConfig
}

func (ihgs *IndexHashedNodesCoordinatorLite) GetNodesCoordinatorHelper() NodesCoordinatorHelper {
	return ihgs.nodesCoordinatorHelper
}
func (ihgs *IndexHashedNodesCoordinatorLite) SetNodesCoordinatorHelper(nch NodesCoordinatorHelper) {
	ihgs.nodesCoordinatorHelper = nch
}

func (ihgs *IndexHashedNodesCoordinatorLite) GetShardIDAsObserver() uint32 {
	return ihgs.shardIDAsObserver
}

func (ihgs *IndexHashedNodesCoordinatorLite) GetConsensusGroupHasher() Cacher {
	return ihgs.consensusGroupCacher
}

func (ihgs *IndexHashedNodesCoordinatorLite) GetPublicKeyToValidatorsMap() map[string]*ValidatorWithShardID {
	return ihgs.publicKeyToValidatorMap
}

func (ihgs *IndexHashedNodesCoordinatorLite) GetFalWaitingListFix() atomicFlags.Flag {
	return ihgs.flagWaitingListFix
}

func (ihgs *IndexHashedNodesCoordinatorLite) GetWaitingListFixEnableEpoch() uint32 {
	return ihgs.waitingListFixEnableEpoch
}

// GetOwnPublicKey will return current node public key  for block sign
// p no
func (ihgs *IndexHashedNodesCoordinatorLite) GetOwnPublicKey() []byte {
	return ihgs.selfPubKey
}

// IsInterfaceNil returns true if there is no value under the interface
func (ihgs *IndexHashedNodesCoordinatorLite) IsInterfaceNil() bool {
	return ihgs == nil
}

// CreateSelectors creates the consensus group selectors for each shard
// Not concurrent safe, needs to be called under mutex
func (ihgs *IndexHashedNodesCoordinatorLite) CreateSelectors(
	nodesConfig *EpochNodesConfig,
) (map[uint32]RandomSelector, error) {
	var err error
	var weights []uint32

	selectors := make(map[uint32]RandomSelector)
	// weights for validators are computed according to each validator rating
	for shard, vList := range nodesConfig.EligibleMap {
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
func (ihgs *IndexHashedNodesCoordinatorLite) ValidatorsWeights(validators []Validator) ([]uint32, error) {
	weights := make([]uint32, len(validators))
	for i := range validators {
		weights[i] = DefaultSelectionChances
	}

	return weights, nil
}

// func (ihgs *IndexHashedNodesCoordinatorLite) saveState(key []byte) error {
// 	registry := ihgs.NodesCoordinatorToRegistry()
// 	data, err := json.Marshal(registry)
// 	if err != nil {
// 		return err
// 	}

// 	ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), key...)

// 	log.Debug("saving nodes coordinator config", "key", ncInternalkey)

// 	return ihgs.bootStorer.Put(ncInternalkey, data)
// }

// // LoadState loads the nodes coordinator state from the used boot storage
// func (ihgs *IndexHashedNodesCoordinatorLite) LoadState(key []byte) error {
// 	return ihgs.baseLoadState(key)
// }

// func (ihgs *IndexHashedNodesCoordinatorLite) baseLoadState(key []byte) error {
// 	ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), key...)

// 	log.Debug("getting nodes coordinator config", "key", ncInternalkey)

// 	ihgs.loadingFromDisk.Store(true)
// 	defer ihgs.loadingFromDisk.Store(false)

// 	data, err := ihgs.bootStorer.Get(ncInternalkey)
// 	if err != nil {
// 		return err
// 	}

// 	config := &NodesCoordinatorRegistry{}
// 	err = json.Unmarshal(data, config)
// 	if err != nil {
// 		return err
// 	}

// 	ihgs.mutSavedStateKey.Lock()
// 	ihgs.savedStateKey = key
// 	ihgs.mutSavedStateKey.Unlock()

// 	ihgs.currentEpoch = config.CurrentEpoch
// 	log.Debug("loaded nodes config", "current epoch", config.CurrentEpoch)

// 	nodesConfig, err := ihgs.registryToNodesCoordinator(config)
// 	if err != nil {
// 		return err
// 	}

// 	displayNodesConfigInfo(nodesConfig)
// 	ihgs.mutNodesConfig.Lock()
// 	ihgs.nodesConfig = nodesConfig
// 	ihgs.mutNodesConfig.Unlock()

// 	return nil
// }

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
