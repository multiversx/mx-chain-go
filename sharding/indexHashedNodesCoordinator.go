package sharding

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	atomicFlags "github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/storage"

	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
)

var _ NodesCoordinator = (*indexHashedNodesCoordinator)(nil)

// TODO: move this to config parameters
const nodesCoordinatorStoredEpochs = 4

type validatorList []nodesCoordinator.Validator

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

// type indexHashedNodesCoordinator struct {
// 	shardIDAsObserver             uint32
// 	currentEpoch                  uint32
// 	shardConsensusGroupSize       int
// 	metaConsensusGroupSize        int
// 	numTotalEligible              uint64
// 	selfPubKey                    []byte
// 	savedStateKey                 []byte
// 	marshalizer                   marshal.Marshalizer
// 	hasher                        hashing.Hasher
// 	shuffler                      NodesShuffler
// 	epochStartRegistrationHandler EpochStartEventNotifier
// 	bootStorer                    storage.Storer
// 	nodesConfig                   map[uint32]*epochNodesConfig
// 	mutNodesConfig                sync.RWMutex
// 	mutSavedStateKey              sync.RWMutex
// 	nodesCoordinatorHelper        NodesCoordinatorHelper
// 	consensusGroupCacher          Cacher
// 	loadingFromDisk               atomic.Value
// 	shuffledOutHandler            ShuffledOutHandler
// 	startEpoch                    uint32
// 	publicKeyToValidatorMap       map[string]*validatorWithShardID
// 	waitingListFixEnableEpoch     uint32
// 	isFullArchive                 bool
// 	chanStopNode                  chan endProcess.ArgEndProcess
// 	flagWaitingListFix            atomicFlags.Flag
// 	nodeTypeProvider              NodeTypeProviderHandler
// }

type indexHashedNodesCoordinator struct {
	*nodesCoordinator.IndexHashedNodesCoordinatorLite

	epochStartRegistrationHandler nodesCoordinator.EpochStartEventNotifier
	flagWaitingListFix            atomicFlags.Flag
	publicKeyToValidatorMap       map[string]*nodesCoordinator.ValidatorWithShardID
	loadingFromDisk               atomic.Value
	bootStorer                    storage.Storer
	savedStateKey                 []byte
	marshalizer                   marshal.Marshalizer
	shuffler                      NodesShuffler
	shuffledOutHandler            ShuffledOutHandler
	mutNodesConfig                sync.RWMutex
	nodesConfig                   map[uint32]*nodesCoordinator.EpochNodesConfig
	mutSavedStateKey              sync.RWMutex
}

// NewIndexHashedNodesCoordinator creates a new index hashed group selector
func NewIndexHashedNodesCoordinator(arguments ArgNodesCoordinator) (*indexHashedNodesCoordinator, error) {

	argumentsNodesCoordinatorLite := nodesCoordinator.ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    arguments.ShardConsensusGroupSize,
		MetaConsensusGroupSize:     arguments.MetaConsensusGroupSize,
		Hasher:                     arguments.Hasher,
		EpochStartNotifier:         arguments.EpochStartNotifier,
		BootStorer:                 arguments.BootStorer,
		ShardIDAsObserver:          arguments.ShardIDAsObserver,
		NbShards:                   arguments.NbShards,
		EligibleNodes:              arguments.EligibleNodes,
		WaitingNodes:               arguments.WaitingNodes,
		SelfPublicKey:              arguments.SelfPublicKey,
		ConsensusGroupCache:        arguments.ConsensusGroupCache,
		Epoch:                      arguments.Epoch,
		StartEpoch:                 arguments.StartEpoch,
		WaitingListFixEnabledEpoch: arguments.WaitingListFixEnabledEpoch,
		ChanStopNode:               arguments.ChanStopNode,
		NodeTypeProvider:           arguments.NodeTypeProvider,
		IsFullArchive:              arguments.IsFullArchive,
	}

	err := checkArguments(arguments)
	if err != nil {
		return nil, err
	}

	ihgsLite, err := nodesCoordinator.NewIndexHashedNodesCoordinatorLite(argumentsNodesCoordinatorLite)
	if err != nil {
		return nil, err
	}

	savedKey := arguments.Hasher.Compute(string(arguments.SelfPublicKey))

	ihgs := &indexHashedNodesCoordinator{
		IndexHashedNodesCoordinatorLite: ihgsLite,
		epochStartRegistrationHandler:   ihgsLite.GetEpochRegistrationHandler(),
		flagWaitingListFix:              ihgsLite.GetFalWaitingListFix(),
		publicKeyToValidatorMap:         ihgsLite.GetPublicKeyToValidatorsMap(),
		bootStorer:                      arguments.BootStorer,
		savedStateKey:                   savedKey,
		marshalizer:                     arguments.Marshalizer,
		shuffler:                        arguments.Shuffler,
		shuffledOutHandler:              arguments.ShuffledOutHandler,
		mutNodesConfig:                  ihgsLite.GetMutNodesConfig(),
		nodesConfig:                     ihgsLite.GetNodesConfig(),
	}

	ihgs.loadingFromDisk.Store(false)

	err = ihgs.saveState(ihgs.savedStateKey)
	if err != nil {
		log.Error("saving initial nodes coordinator config failed",
			"error", err.Error())
	}

	ihgs.epochStartRegistrationHandler.UnregisterHandler(ihgsLite)
	ihgs.epochStartRegistrationHandler.RegisterHandler(ihgs)

	return ihgs, nil
}

func checkArguments(arguments ArgNodesCoordinator) error {
	if check.IfNil(arguments.Shuffler) {
		return ErrNilShuffler
	}
	if check.IfNil(arguments.Marshalizer) {
		return ErrNilMarshalizer
	}
	if check.IfNil(arguments.ShuffledOutHandler) {
		return ErrNilShuffledOutHandler
	}

	return nil
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
func (ihgs *indexHashedNodesCoordinator) GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
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

	nodesConfig.MutNodesMaps.RLock()
	defer nodesConfig.MutNodesMaps.RUnlock()

	for shardID, shardLeaving := range nodesConfig.LeavingMap {
		for i := 0; i < len(shardLeaving); i++ {
			validatorsPubKeys[shardID] = append(validatorsPubKeys[shardID], shardLeaving[i].PubKey())
		}
	}

	return validatorsPubKeys, nil
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

	if check.IfNil(body) && newEpoch == ihgs.GetCurrentEpoch() {
		log.Debug("nil body provided for epoch start prepare, it is normal in case of revertStateToBlock")
		return
	}

	allValidatorInfo, err := createValidatorInfoFromBody(body, ihgs.marshalizer, ihgs.GetNumTotalEligible())
	if err != nil {
		log.Error("could not create validator info from body - do nothing on nodesCoordinator epochStartPrepare")
		return
	}

	ihgs.updateEpochFlags(newEpoch)

	ihgs.mutNodesConfig.RLock()
	previousConfig := ihgs.nodesConfig[ihgs.GetCurrentEpoch()]
	if previousConfig == nil {
		log.Error("previous nodes config is nil")
		ihgs.mutNodesConfig.RUnlock()
		return
	}

	//TODO: remove the copy if no changes are done to the maps
	copiedPrevious := &nodesCoordinator.EpochNodesConfig{}
	copiedPrevious.EligibleMap = copyValidatorMap(previousConfig.EligibleMap)
	copiedPrevious.WaitingMap = copyValidatorMap(previousConfig.WaitingMap)
	copiedPrevious.NbShards = previousConfig.NbShards

	ihgs.mutNodesConfig.RUnlock()

	// TODO: compare with previous nodesConfig if exists
	newNodesConfig, err := ihgs.computeNodesConfigFromList(copiedPrevious, allValidatorInfo)
	if err != nil {
		log.Error("could not compute nodes config from list - do nothing on nodesCoordinator epochStartPrepare")
		return
	}

	if copiedPrevious.NbShards != newNodesConfig.NbShards {
		log.Warn("number of shards does not match",
			"previous epoch", ihgs.GetCurrentEpoch(),
			"previous number of shards", copiedPrevious.NbShards,
			"new epoch", newEpoch,
			"new number of shards", newNodesConfig.NbShards)
	}

	additionalLeavingMap, err := ihgs.GetNodesCoordinatorHelper().ComputeAdditionalLeaving(allValidatorInfo)
	if err != nil {
		log.Error("could not compute additionalLeaving Nodes  - do nothing on nodesCoordinator epochStartPrepare")
		return
	}

	unStakeLeavingList := ihgs.createSortedListFromMap(newNodesConfig.LeavingMap)
	additionalLeavingList := ihgs.createSortedListFromMap(additionalLeavingMap)

	shufflerArgs := ArgsUpdateNodes{
		Eligible:          newNodesConfig.EligibleMap,
		Waiting:           newNodesConfig.WaitingMap,
		NewNodes:          newNodesConfig.NewList,
		UnStakeLeaving:    unStakeLeavingList,
		AdditionalLeaving: additionalLeavingList,
		Rand:              randomness,
		NbShards:          newNodesConfig.NbShards,
		Epoch:             newEpoch,
	}

	resUpdateNodes, err := ihgs.shuffler.UpdateNodeLists(shufflerArgs)
	if err != nil {
		log.Error("could not compute UpdateNodeLists - do nothing on nodesCoordinator epochStartPrepare", "err", err.Error())
		return
	}

	leavingNodesMap, stillRemainingNodesMap := createActuallyLeavingPerShards(
		newNodesConfig.LeavingMap,
		additionalLeavingMap,
		resUpdateNodes.Leaving,
	)

	err = ihgs.SetNodesPerShards(resUpdateNodes.Eligible, resUpdateNodes.Waiting, leavingNodesMap, newEpoch)
	if err != nil {
		log.Error("set nodes per shard failed", "error", err.Error())
	}

	ihgs.fillPublicKeyToValidatorMap()
	err = ihgs.saveState(randomness)
	if err != nil {
		log.Error("saving nodes coordinator config failed", "error", err.Error())
	}

	nodesCoordinator.DisplayNodesConfiguration(
		resUpdateNodes.Eligible,
		resUpdateNodes.Waiting,
		leavingNodesMap,
		stillRemainingNodesMap,
		newNodesConfig.NbShards)

	ihgs.mutSavedStateKey.Lock()
	ihgs.savedStateKey = randomness
	ihgs.mutSavedStateKey.Unlock()

	ihgs.GetConsensusGroupHasher().Clear()
}

func (ihgs *indexHashedNodesCoordinator) fillPublicKeyToValidatorMap() {
	ihgs.mutNodesConfig.Lock()
	defer ihgs.mutNodesConfig.Unlock()

	index := 0
	epochList := make([]uint32, len(ihgs.nodesConfig))
	mapAllValidators := make(map[uint32]map[string]*nodesCoordinator.ValidatorWithShardID)
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

	ihgs.publicKeyToValidatorMap = make(map[string]*nodesCoordinator.ValidatorWithShardID)
	for _, epoch := range epochList {
		validatorsForEpoch := mapAllValidators[epoch]
		for pubKey, vInfo := range validatorsForEpoch {
			ihgs.publicKeyToValidatorMap[pubKey] = vInfo
		}
	}
}

func (ihgs *indexHashedNodesCoordinator) createSortedListFromMap(validatorsMap map[uint32][]nodesCoordinator.Validator) []nodesCoordinator.Validator {
	sortedList := make([]nodesCoordinator.Validator, 0)
	for _, validators := range validatorsMap {
		sortedList = append(sortedList, validators...)
	}
	sort.Sort(validatorList(sortedList))
	return sortedList
}

func (ihgs *indexHashedNodesCoordinator) computeNodesConfigFromList(
	previousEpochConfig *nodesCoordinator.EpochNodesConfig,
	validatorInfos []*state.ShardValidatorInfo,
) (*nodesCoordinator.EpochNodesConfig, error) {
	eligibleMap := make(map[uint32][]nodesCoordinator.Validator)
	waitingMap := make(map[uint32][]nodesCoordinator.Validator)
	leavingMap := make(map[uint32][]nodesCoordinator.Validator)
	newNodesList := make([]nodesCoordinator.Validator, 0)

	if ihgs.flagWaitingListFix.IsSet() && previousEpochConfig == nil {
		return nil, ErrNilPreviousEpochConfig
	}

	if len(validatorInfos) == 0 {
		log.Warn("computeNodesConfigFromList - validatorInfos len is 0")
	}

	for _, validatorInfo := range validatorInfos {
		chance := ihgs.GetNodesCoordinatorHelper().GetChance(validatorInfo.TempRating)
		currentValidator, err := nodesCoordinator.NewValidator(validatorInfo.PublicKey, chance, validatorInfo.Index)
		if err != nil {
			return nil, err
		}

		switch validatorInfo.List {
		case string(common.WaitingList):
			waitingMap[validatorInfo.ShardId] = append(waitingMap[validatorInfo.ShardId], currentValidator)
		case string(common.EligibleList):
			eligibleMap[validatorInfo.ShardId] = append(eligibleMap[validatorInfo.ShardId], currentValidator)
		case string(common.LeavingList):
			log.Debug("leaving node validatorInfo", "pk", validatorInfo.PublicKey)
			leavingMap[validatorInfo.ShardId] = append(leavingMap[validatorInfo.ShardId], currentValidator)
			ihgs.addValidatorToPreviousMap(
				previousEpochConfig,
				eligibleMap,
				waitingMap,
				currentValidator,
				validatorInfo.ShardId)
		case string(common.NewList):
			log.Debug("new node registered", "pk", validatorInfo.PublicKey)
			newNodesList = append(newNodesList, currentValidator)
		case string(common.InactiveList):
			log.Debug("inactive validator", "pk", validatorInfo.PublicKey)
		case string(common.JailedList):
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

	newNodesConfig := &nodesCoordinator.EpochNodesConfig{
		EligibleMap: eligibleMap,
		WaitingMap:  waitingMap,
		LeavingMap:  leavingMap,
		NewList:     newNodesList,
		NbShards:    uint32(nbShards),
	}

	return newNodesConfig, nil
}

func (ihgs *indexHashedNodesCoordinator) addValidatorToPreviousMap(
	previousEpochConfig *nodesCoordinator.EpochNodesConfig,
	eligibleMap map[uint32][]nodesCoordinator.Validator,
	waitingMap map[uint32][]nodesCoordinator.Validator,
	currentValidator nodesCoordinator.Validator,
	currentValidatorShardId uint32) {

	if !ihgs.flagWaitingListFix.IsSet() {
		eligibleMap[currentValidatorShardId] = append(eligibleMap[currentValidatorShardId], currentValidator)
		return
	}

	found, shardId := searchInMap(previousEpochConfig.EligibleMap, currentValidator.PubKey())
	if found {
		log.Debug("leaving node found in", "list", "eligible", "shardId", shardId)
		eligibleMap[shardId] = append(eligibleMap[currentValidatorShardId], currentValidator)
		return
	}

	found, shardId = searchInMap(previousEpochConfig.WaitingMap, currentValidator.PubKey())
	if found {
		log.Debug("leaving node found in", "list", "waiting", "shardId", shardId)
		waitingMap[shardId] = append(waitingMap[currentValidatorShardId], currentValidator)
		return
	}
}

// EpochStartAction is called upon a start of epoch event.
// NodeCoordinator has to get the nodes assignment to shards using the shuffler.
func (ihgs *indexHashedNodesCoordinator) EpochStartAction(hdr data.HeaderHandler) {
	newEpoch := hdr.GetEpoch()
	epochToRemove := int32(newEpoch) - nodesCoordinatorStoredEpochs
	needToRemove := epochToRemove >= 0
	ihgs.SetCurrentEpoch(newEpoch)

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
	return common.NodesCoordinatorOrder
}

// GetSavedStateKey returns the key for the last nodes coordinator saved state
func (ihgs *indexHashedNodesCoordinator) GetSavedStateKey() []byte {
	ihgs.mutSavedStateKey.RLock()
	key := ihgs.savedStateKey
	ihgs.mutSavedStateKey.RUnlock()

	return key
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

	if isValidator(nodesConfig, ihgs.GetSelfPubKey()) {
		err := ihgs.shuffledOutHandler.Process(nodesConfig.ShardID)
		if err != nil {
			log.Warn("shuffle out process failed", "err", err)
		}
	}
}

func isValidator(config *nodesCoordinator.EpochNodesConfig, pk []byte) bool {
	if config == nil {
		return false
	}

	config.MutNodesMaps.RLock()
	defer config.MutNodesMaps.RUnlock()

	found := false
	found, _ = searchInMap(config.EligibleMap, pk)
	if found {
		return true
	}

	found, _ = searchInMap(config.WaitingMap, pk)
	return found
}

func searchInMap(validatorMap map[uint32][]nodesCoordinator.Validator, pk []byte) (bool, uint32) {
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

	if epoch > ihgs.GetStartEpoch() {
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
	eligible map[uint32][]nodesCoordinator.Validator,
	waiting map[uint32][]nodesCoordinator.Validator,
) map[string]*nodesCoordinator.ValidatorWithShardID {
	publicKeyToValidatorMap := make(map[string]*nodesCoordinator.ValidatorWithShardID)
	for shardId, shardEligible := range eligible {
		for i := 0; i < len(shardEligible); i++ {
			publicKeyToValidatorMap[string(shardEligible[i].PubKey())] = &nodesCoordinator.ValidatorWithShardID{
				Validator: shardEligible[i],
				ShardID:   shardId,
			}
		}
	}
	for shardId, shardWaiting := range waiting {
		for i := 0; i < len(shardWaiting); i++ {
			publicKeyToValidatorMap[string(shardWaiting[i].PubKey())] = &nodesCoordinator.ValidatorWithShardID{
				Validator: shardWaiting[i],
				ShardID:   shardId,
			}
		}
	}

	return publicKeyToValidatorMap
}

func (ihgs *indexHashedNodesCoordinator) computeShardForSelfPublicKey(nodesConfig *nodesCoordinator.EpochNodesConfig) (uint32, bool) {
	pubKey := ihgs.GetSelfPubKey()
	selfShard := ihgs.GetShardIDAsObserver()
	epNodesConfig, ok := ihgs.nodesConfig[ihgs.GetCurrentEpoch()]
	if ok {
		log.Trace("computeShardForSelfPublicKey found existing config",
			"shard", epNodesConfig.ShardID,
		)
		selfShard = epNodesConfig.ShardID
	}

	found, shardId := searchInMap(nodesConfig.EligibleMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in eligible",
			"epoch", ihgs.GetCurrentEpoch(),
			"shard", shardId,
			"validator PK", pubKey,
		)
		return shardId, true
	}

	found, shardId = searchInMap(nodesConfig.WaitingMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in waiting",
			"epoch", ihgs.GetCurrentEpoch(),
			"shard", shardId,
			"validator PK", pubKey,
		)
		return shardId, true
	}

	found, shardId = searchInMap(nodesConfig.LeavingMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in leaving",
			"epoch", ihgs.GetCurrentEpoch(),
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

func createActuallyLeavingPerShards(
	unstakeLeaving map[uint32][]nodesCoordinator.Validator,
	additionalLeaving map[uint32][]nodesCoordinator.Validator,
	leaving []nodesCoordinator.Validator,
) (map[uint32][]nodesCoordinator.Validator, map[uint32][]nodesCoordinator.Validator) {
	actuallyLeaving := make(map[uint32][]nodesCoordinator.Validator)
	actuallyRemaining := make(map[uint32][]nodesCoordinator.Validator)
	processedValidatorsMap := make(map[string]bool)

	computeActuallyLeaving(unstakeLeaving, leaving, actuallyLeaving, actuallyRemaining, processedValidatorsMap)
	computeActuallyLeaving(additionalLeaving, leaving, actuallyLeaving, actuallyRemaining, processedValidatorsMap)

	return actuallyLeaving, actuallyRemaining
}

func computeActuallyLeaving(
	unstakeLeaving map[uint32][]nodesCoordinator.Validator,
	leaving []nodesCoordinator.Validator,
	actuallyLeaving map[uint32][]nodesCoordinator.Validator,
	actuallyRemaining map[uint32][]nodesCoordinator.Validator,
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

func (ihgs *indexHashedNodesCoordinator) updateEpochFlags(epoch uint32) {
	ihgs.flagWaitingListFix.Toggle(epoch >= ihgs.GetWaitingListFixEnableEpoch())
	log.Debug("indexHashedNodesCoordinator: waiting list fix", "enabled", ihgs.flagWaitingListFix.IsSet())
}

// IsInterfaceNil returns true if there is no value under the interface
func (ihgs *indexHashedNodesCoordinator) IsInterfaceNil() bool {
	return ihgs == nil
}
