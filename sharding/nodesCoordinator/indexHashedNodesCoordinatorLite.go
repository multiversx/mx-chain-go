package nodesCoordinator

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	atomicFlags "github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
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
const NodesCoordinatorStoredEpochs = 4

type ValidatorList []Validator

// Len will return the length of the validatorList
func (v ValidatorList) Len() int { return len(v) }

// Swap will interchange the objects on input indexes
func (v ValidatorList) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

// Less will return true if object on index i should appear before object in index j
// Sorting of validators should be by index and public key
func (v ValidatorList) Less(i, j int) bool {
	if v[i].Index() == v[j].Index() {
		return bytes.Compare(v[i].PubKey(), v[j].PubKey()) < 0
	}
	return v[i].Index() < v[j].Index()
}

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
	shardIDAsObserver         uint32
	currentEpoch              uint32
	shardConsensusGroupSize   int
	metaConsensusGroupSize    int
	numTotalEligible          uint64
	selfPubKey                []byte
	hasher                    hashing.Hasher
	nodesConfig               map[uint32]*EpochNodesConfig
	mutNodesConfig            sync.RWMutex
	nodesCoordinatorHelper    NodesCoordinatorHelper
	consensusGroupCacher      Cacher
	publicKeyToValidatorMap   map[string]*ValidatorWithShardID
	waitingListFixEnableEpoch uint32
	isFullArchive             bool
	chanStopNode              chan endProcess.ArgEndProcess
	nodeTypeProvider          NodeTypeProviderHandler
	flagWaitingListFix        atomicFlags.Flag
	shuffler                  NodesShuffler
}

// NewIndexHashedNodesCoordinatorLite creates a new index hashed group selector
func NewIndexHashedNodesCoordinatorLite(arguments ArgNodesCoordinatorLite) (*IndexHashedNodesCoordinatorLite, error) {
	err := checkArguments(arguments)
	if err != nil {
		return nil, err
	}

	nodesConfig := make(map[uint32]*EpochNodesConfig, NodesCoordinatorStoredEpochs)

	nodesConfig[arguments.Epoch] = &EpochNodesConfig{
		NbShards:    arguments.NbShards,
		ShardID:     arguments.ShardIDAsObserver,
		EligibleMap: make(map[uint32][]Validator),
		WaitingMap:  make(map[uint32][]Validator),
		Selectors:   make(map[uint32]RandomSelector),
		LeavingMap:  make(map[uint32][]Validator),
		NewList:     make([]Validator, 0),
	}

	ihncl := &IndexHashedNodesCoordinatorLite{
		shardConsensusGroupSize:   arguments.ShardConsensusGroupSize,
		metaConsensusGroupSize:    arguments.MetaConsensusGroupSize,
		hasher:                    arguments.Hasher,
		shardIDAsObserver:         arguments.ShardIDAsObserver,
		selfPubKey:                arguments.SelfPublicKey,
		nodesConfig:               nodesConfig,
		currentEpoch:              arguments.Epoch,
		consensusGroupCacher:      arguments.ConsensusGroupCache,
		publicKeyToValidatorMap:   make(map[string]*ValidatorWithShardID),
		waitingListFixEnableEpoch: arguments.WaitingListFixEnabledEpoch,
		chanStopNode:              arguments.ChanStopNode,
		nodeTypeProvider:          arguments.NodeTypeProvider,
		isFullArchive:             arguments.IsFullArchive,
		shuffler:                  arguments.Shuffler,
	}
	log.Debug("indexHashedNodesCoordinator: enable epoch for waiting waiting list", "epoch", ihncl.waitingListFixEnableEpoch)

	ihncl.nodesCoordinatorHelper = ihncl
	err = ihncl.SetNodesPerShards(arguments.EligibleNodes, arguments.WaitingNodes, nil, arguments.Epoch)
	if err != nil {
		return nil, err
	}

	ihncl.FillPublicKeyToValidatorMap()

	log.Info("new nodes config is set for epoch", "epoch", arguments.Epoch)
	currentNodesConfig := ihncl.nodesConfig[arguments.Epoch]
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

	return ihncl, nil
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
	if check.IfNilReflect(arguments.ConsensusGroupCache) {
		return ErrNilCacher
	}
	if check.IfNil(arguments.NodeTypeProvider) {
		return ErrNilNodeTypeProvider
	}
	if nil == arguments.ChanStopNode {
		return ErrNilNodeStopChannel
	}
	if check.IfNil(arguments.Shuffler) {
		return ErrNilShuffler
	}

	return nil
}

// SetNodesPerShards loads the distribution of nodes per shard into the nodes management component
func (ihncl *IndexHashedNodesCoordinatorLite) SetNodesPerShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	leaving map[uint32][]Validator,
	epoch uint32,
) error {
	ihncl.mutNodesConfig.Lock()
	defer ihncl.mutNodesConfig.Unlock()

	nodesConfig, ok := ihncl.nodesConfig[epoch]
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
	if len(nodesList) < ihncl.metaConsensusGroupSize {
		return ErrSmallMetachainEligibleListSize
	}

	numTotalEligible := uint64(len(nodesList))
	for shardId := uint32(0); shardId < uint32(len(eligible)-1); shardId++ {
		nbNodesShard := len(eligible[shardId])
		if nbNodesShard < ihncl.shardConsensusGroupSize {
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
	nodesConfig.ShardID, isValidator = ihncl.computeShardForSelfPublicKey(nodesConfig)
	nodesConfig.Selectors, err = ihncl.CreateSelectors(nodesConfig)
	if err != nil {
		return err
	}

	ihncl.nodesConfig[epoch] = nodesConfig
	ihncl.numTotalEligible = numTotalEligible
	ihncl.setNodeType(isValidator)

	if ihncl.isFullArchive && isValidator {
		ihncl.chanStopNode <- endProcess.ArgEndProcess{
			Reason:      common.WrongConfiguration,
			Description: ErrValidatorCannotBeFullArchive.Error(),
		}

		return nil
	}

	return nil
}

func (ihncl *IndexHashedNodesCoordinatorLite) setNodeType(isValidator bool) {
	if isValidator {
		ihncl.nodeTypeProvider.SetType(core.NodeTypeValidator)
		return
	}

	ihncl.nodeTypeProvider.SetType(core.NodeTypeObserver)
}

func (ihncl *IndexHashedNodesCoordinatorLite) SetNodesConfigOnNewEpochStart(newEpoch uint32, randomness []byte, validatorsInfo []*state.ShardValidatorInfo) {
	copiedPrevious := ihncl.GetPreviousConfigCopy()
	if copiedPrevious == nil {
		log.Error("previous nodes config is nil")
		return
	}

	newNodesConfig, err := ihncl.ComputeNodesConfigFromList(copiedPrevious, validatorsInfo)
	if err != nil {
		log.Error("could not compute nodes config from list - do nothing on nodesCoordinator epochStartPrepare")
		return
	}

	additionalLeavingMap, err := ihncl.GetNodesCoordinatorHelper().ComputeAdditionalLeaving(validatorsInfo)
	if err != nil {
		log.Error("could not compute additionalLeaving Nodes  - do nothing on nodesCoordinator epochStartPrepare")
		return
	}

	unStakeLeavingList := ihncl.CreateSortedListFromMap(newNodesConfig.LeavingMap)
	additionalLeavingList := ihncl.CreateSortedListFromMap(additionalLeavingMap)

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

	resUpdateNodes, err := ihncl.shuffler.UpdateNodeLists(shufflerArgs)
	if err != nil {
		log.Error("could not compute UpdateNodeLists - do nothing on nodesCoordinator epochStartPrepare", "err", err.Error())
		return
	}

	leavingNodesMap, _ := CreateActuallyLeavingPerShards(
		newNodesConfig.LeavingMap,
		additionalLeavingMap,
		resUpdateNodes.Leaving,
	)

	err = ihncl.SetNodesPerShards(resUpdateNodes.Eligible, resUpdateNodes.Waiting, leavingNodesMap, newEpoch)
	if err != nil {
		log.Error("set nodes per shard failed", "error", err.Error())
	}
}

func (ihncl *IndexHashedNodesCoordinatorLite) CreateSortedListFromMap(validatorsMap map[uint32][]Validator) []Validator {
	sortedList := make([]Validator, 0)
	for _, validators := range validatorsMap {
		sortedList = append(sortedList, validators...)
	}
	sort.Sort(ValidatorList(sortedList))
	return sortedList
}

func CreateActuallyLeavingPerShards(
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

func (ihncl *IndexHashedNodesCoordinatorLite) UpdateNodeLists(args ArgsUpdateNodes) (*ResUpdateNodes, error) {
	return ihncl.shuffler.UpdateNodeLists(args)
}

// ComputeAdditionalLeaving - computes extra leaving validators based on computation at the start of epoch
func (ihncl *IndexHashedNodesCoordinatorLite) ComputeAdditionalLeaving(_ []*state.ShardValidatorInfo) (map[uint32][]Validator, error) {
	return make(map[uint32][]Validator), nil
}

// ComputeConsensusGroup will generate a list of validators based on the the eligible list
// and each eligible validator weight/chance
func (ihncl *IndexHashedNodesCoordinatorLite) ComputeConsensusGroup(
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

	ihncl.mutNodesConfig.RLock()
	nodesConfig, ok := ihncl.nodesConfig[epoch]
	if ok {
		if shardID >= nodesConfig.NbShards && shardID != core.MetachainShardId {
			log.Warn("shardID is not ok", "shardID", shardID, "nbShards", nodesConfig.NbShards)
			ihncl.mutNodesConfig.RUnlock()
			return nil, ErrInvalidShardId
		}
		selector = nodesConfig.Selectors[shardID]
		eligibleList = nodesConfig.EligibleMap[shardID]
	}
	ihncl.mutNodesConfig.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	key := []byte(fmt.Sprintf(keyFormat, string(randomness), round, shardID, epoch))
	validators := ihncl.searchConsensusForKey(key)
	if validators != nil {
		return validators, nil
	}

	consensusSize := ihncl.ConsensusGroupSize(shardID)
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

	ihncl.consensusGroupCacher.Put(key, tempList, size)

	return tempList, nil
}

func (ihncl *IndexHashedNodesCoordinatorLite) searchConsensusForKey(key []byte) []Validator {
	value, ok := ihncl.consensusGroupCacher.Get(key)
	if ok {
		consensusGroup, typeOk := value.([]Validator)
		if typeOk {
			return consensusGroup
		}
	}
	return nil
}

// GetValidatorWithPublicKey gets the validator with the given public key
func (ihncl *IndexHashedNodesCoordinatorLite) GetValidatorWithPublicKey(publicKey []byte) (Validator, uint32, error) {
	if len(publicKey) == 0 {
		return nil, 0, ErrNilPubKey
	}
	ihncl.mutNodesConfig.RLock()
	v, ok := ihncl.publicKeyToValidatorMap[string(publicKey)]
	ihncl.mutNodesConfig.RUnlock()
	if ok {
		return v.Validator, v.ShardID, nil
	}

	return nil, 0, ErrValidatorNotFound
}

// GetConsensusValidatorsPublicKeys calculates the validators consensus group for a specific shard, randomness and round number,
// returning their public keys
func (ihncl *IndexHashedNodesCoordinatorLite) GetConsensusValidatorsPublicKeys(
	randomness []byte,
	round uint64,
	shardID uint32,
	epoch uint32,
) ([]string, error) {
	consensusNodes, err := ihncl.ComputeConsensusGroup(randomness, round, shardID, epoch)
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
func (ihncl *IndexHashedNodesCoordinatorLite) GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihncl.mutNodesConfig.RLock()
	nodesConfig, ok := ihncl.nodesConfig[epoch]
	ihncl.mutNodesConfig.RUnlock()

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
func (ihncl *IndexHashedNodesCoordinatorLite) GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihncl.mutNodesConfig.RLock()
	nodesConfig, ok := ihncl.nodesConfig[epoch]
	ihncl.mutNodesConfig.RUnlock()

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
func (ihncl *IndexHashedNodesCoordinatorLite) GetAllLeavingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihncl.mutNodesConfig.RLock()
	nodesConfig, ok := ihncl.nodesConfig[epoch]
	ihncl.mutNodesConfig.RUnlock()

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

// GetValidatorsIndexes will return validators indexes for a block
func (ihncl *IndexHashedNodesCoordinatorLite) GetValidatorsIndexes(
	publicKeys []string,
	epoch uint32,
) ([]uint64, error) {
	signersIndexes := make([]uint64, 0)

	validatorsPubKeys, err := ihncl.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		return nil, err
	}

	ihncl.mutNodesConfig.RLock()
	nodesConfig := ihncl.nodesConfig[epoch]
	ihncl.mutNodesConfig.RUnlock()

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

func (ihncl *IndexHashedNodesCoordinatorLite) FillPublicKeyToValidatorMap() {
	ihncl.mutNodesConfig.Lock()
	defer ihncl.mutNodesConfig.Unlock()

	index := 0
	epochList := make([]uint32, len(ihncl.nodesConfig))
	mapAllValidators := make(map[uint32]map[string]*ValidatorWithShardID)
	for epoch, epochConfig := range ihncl.nodesConfig {
		epochConfig.MutNodesMaps.RLock()
		mapAllValidators[epoch] = ihncl.createPublicKeyToValidatorMap(epochConfig.EligibleMap, epochConfig.WaitingMap)
		epochConfig.MutNodesMaps.RUnlock()

		epochList[index] = epoch
		index++
	}

	sort.Slice(epochList, func(i, j int) bool {
		return epochList[i] < epochList[j]
	})

	ihncl.publicKeyToValidatorMap = make(map[string]*ValidatorWithShardID)
	for _, epoch := range epochList {
		validatorsForEpoch := mapAllValidators[epoch]
		for pubKey, vInfo := range validatorsForEpoch {
			ihncl.publicKeyToValidatorMap[pubKey] = vInfo
		}
	}
}

// GetChance will return default chance
func (ihncl *IndexHashedNodesCoordinatorLite) GetChance(_ uint32) uint32 {
	return DefaultSelectionChances
}

// ShardIdForEpoch returns the nodesCoordinator configured ShardId for specified epoch if epoch configuration exists,
// otherwise error
func (ihncl *IndexHashedNodesCoordinatorLite) ShardIdForEpoch(epoch uint32) (uint32, error) {
	ihncl.mutNodesConfig.RLock()
	nodesConfig, ok := ihncl.nodesConfig[epoch]
	ihncl.mutNodesConfig.RUnlock()

	if !ok {
		return 0, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	return nodesConfig.ShardID, nil
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

func (ihncl *IndexHashedNodesCoordinatorLite) createPublicKeyToValidatorMap(
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

func (ihncl *IndexHashedNodesCoordinatorLite) computeShardForSelfPublicKey(nodesConfig *EpochNodesConfig) (uint32, bool) {
	pubKey := ihncl.selfPubKey
	selfShard := ihncl.shardIDAsObserver
	epNodesConfig, ok := ihncl.nodesConfig[ihncl.currentEpoch]
	if ok {
		log.Trace("computeShardForSelfPublicKey found existing config",
			"shard", epNodesConfig.ShardID,
		)
		selfShard = epNodesConfig.ShardID
	}

	found, shardId := searchInMap(nodesConfig.EligibleMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in eligible",
			"epoch", ihncl.currentEpoch,
			"shard", shardId,
			"validator PK", pubKey,
		)
		return shardId, true
	}

	found, shardId = searchInMap(nodesConfig.WaitingMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in waiting",
			"epoch", ihncl.currentEpoch,
			"shard", shardId,
			"validator PK", pubKey,
		)
		return shardId, true
	}

	found, shardId = searchInMap(nodesConfig.LeavingMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in leaving",
			"epoch", ihncl.currentEpoch,
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

func (ihncl *IndexHashedNodesCoordinatorLite) ComputeNodesConfigFromList(
	previousEpochConfig *EpochNodesConfig,
	validatorInfos []*state.ShardValidatorInfo,
) (*EpochNodesConfig, error) {
	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	leavingMap := make(map[uint32][]Validator)
	newNodesList := make([]Validator, 0)

	if ihncl.flagWaitingListFix.IsSet() && previousEpochConfig == nil {
		return nil, ErrNilPreviousEpochConfig
	}

	if len(validatorInfos) == 0 {
		log.Warn("computeNodesConfigFromList - validatorInfos len is 0")
	}

	for _, validatorInfo := range validatorInfos {
		chance := ihncl.GetNodesCoordinatorHelper().GetChance(validatorInfo.TempRating)
		currentValidator, err := NewValidator(validatorInfo.PublicKey, chance, validatorInfo.Index)
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
			ihncl.addValidatorToPreviousMap(
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

	sort.Sort(ValidatorList(newNodesList))
	for _, eligibleList := range eligibleMap {
		sort.Sort(ValidatorList(eligibleList))
	}
	for _, waitingList := range waitingMap {
		sort.Sort(ValidatorList(waitingList))
	}
	for _, leavingList := range leavingMap {
		sort.Sort(ValidatorList(leavingList))
	}

	if len(eligibleMap) == 0 {
		return nil, fmt.Errorf("%w eligible map size is zero. No validators found", ErrMapSizeZero)
	}

	nbShards := len(eligibleMap) - 1

	newNodesConfig := &EpochNodesConfig{
		EligibleMap: eligibleMap,
		WaitingMap:  waitingMap,
		LeavingMap:  leavingMap,
		NewList:     newNodesList,
		NbShards:    uint32(nbShards),
	}

	return newNodesConfig, nil
}

func (ihncl *IndexHashedNodesCoordinatorLite) addValidatorToPreviousMap(
	previousEpochConfig *EpochNodesConfig,
	eligibleMap map[uint32][]Validator,
	waitingMap map[uint32][]Validator,
	currentValidator Validator,
	currentValidatorShardId uint32) {

	if !ihncl.flagWaitingListFix.IsSet() {
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

func (ihncl *IndexHashedNodesCoordinatorLite) UpdateEpochFlags(epoch uint32) {
	ihncl.flagWaitingListFix.Toggle(epoch >= ihncl.waitingListFixEnableEpoch)
	log.Debug("indexHashedNodesCoordinator: waiting list fix", "enabled", ihncl.flagWaitingListFix.IsSet())
}

// ConsensusGroupSize returns the consensus group size for a specific shard
func (ihncl *IndexHashedNodesCoordinatorLite) ConsensusGroupSize(
	shardID uint32,
) int {
	if shardID == core.MetachainShardId {
		return ihncl.metaConsensusGroupSize
	}

	return ihncl.shardConsensusGroupSize
}

// CreateSelectors creates the consensus group selectors for each shard
// Not concurrent safe, needs to be called under mutex
func (ihncl *IndexHashedNodesCoordinatorLite) CreateSelectors(
	nodesConfig *EpochNodesConfig,
) (map[uint32]RandomSelector, error) {
	var err error
	var weights []uint32

	selectors := make(map[uint32]RandomSelector)
	// weights for validators are computed according to each validator rating
	for shard, vList := range nodesConfig.EligibleMap {
		log.Debug("create selectors", "shard", shard)
		weights, err = ihncl.nodesCoordinatorHelper.ValidatorsWeights(vList)
		if err != nil {
			return nil, err
		}

		selectors[shard], err = NewSelectorExpandedList(weights, ihncl.hasher)
		if err != nil {
			return nil, err
		}
	}

	return selectors, nil
}

// ValidatorsWeights returns the weights/chances for each of the given validators
func (ihncl *IndexHashedNodesCoordinatorLite) ValidatorsWeights(validators []Validator) ([]uint32, error) {
	weights := make([]uint32, len(validators))
	for i := range validators {
		weights[i] = DefaultSelectionChances
	}

	return weights, nil
}

func (ihncl *IndexHashedNodesCoordinatorLite) GetPreviousConfigCopy() *EpochNodesConfig {
	ihncl.mutNodesConfig.RLock()
	defer ihncl.mutNodesConfig.RUnlock()

	previousConfig := ihncl.nodesConfig[ihncl.GetCurrentEpoch()]
	if previousConfig == nil {
		return nil
	}

	copiedPrevious := &EpochNodesConfig{}
	copiedPrevious.EligibleMap = CopyValidatorMap(previousConfig.EligibleMap)
	copiedPrevious.WaitingMap = CopyValidatorMap(previousConfig.WaitingMap)
	copiedPrevious.NbShards = previousConfig.NbShards

	return copiedPrevious
}

func (ihncl *IndexHashedNodesCoordinatorLite) RemoveNodesConfigEpochs(epochToRemove int32) {
	ihncl.mutNodesConfig.RLock()
	for epoch := range ihncl.nodesConfig {
		if epoch <= uint32(epochToRemove) {
			delete(ihncl.nodesConfig, epoch)
		}
	}
	ihncl.mutNodesConfig.RUnlock()
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

// GetLastEpochConfig returns the last epoch from nodes config
func (ihncl *IndexHashedNodesCoordinatorLite) GetLastEpochConfig() uint32 {
	ihncl.mutNodesConfig.Lock()
	defer ihncl.mutNodesConfig.Unlock()

	lastEpoch := uint32(0)
	for epoch := range ihncl.nodesConfig {
		if lastEpoch < epoch {
			lastEpoch = epoch
		}
	}

	return lastEpoch
}

// GetNumTotalEligible returns the number of total eligible accross all shards from current setup
func (ihncl *IndexHashedNodesCoordinatorLite) GetNumTotalEligible() uint64 {
	return ihncl.numTotalEligible
}

// GetCurrentEpoch returns current epoch
func (ihncl *IndexHashedNodesCoordinatorLite) GetCurrentEpoch() uint32 {
	return ihncl.currentEpoch
}

// SetCurrentEpoch updates current epoch
func (ihncl *IndexHashedNodesCoordinatorLite) SetCurrentEpoch(epoch uint32) {
	ihncl.currentEpoch = epoch
}

// SetNodesConfig updates nodes config in a concurrent safe way
func (ihncl *IndexHashedNodesCoordinatorLite) SetNodesConfig(nodesConfig map[uint32]*EpochNodesConfig) {
	ihncl.mutNodesConfig.Lock()
	ihncl.nodesConfig = nodesConfig
	ihncl.mutNodesConfig.Unlock()
}

// GetNodesConfig returns nodes config for all epochs
func (ihncl *IndexHashedNodesCoordinatorLite) GetNodesConfig() map[uint32]*EpochNodesConfig {
	ihncl.mutNodesConfig.RLock()
	defer ihncl.mutNodesConfig.RUnlock()

	return ihncl.nodesConfig
}

// GetNodesConfigPerEpoch returns nodes config for the specified epoch
func (ihncl *IndexHashedNodesCoordinatorLite) GetNodesConfigPerEpoch(epoch uint32) (*EpochNodesConfig, bool) {
	ihncl.mutNodesConfig.RLock()
	nodesConfig, ok := ihncl.nodesConfig[epoch]
	ihncl.mutNodesConfig.RUnlock()

	return nodesConfig, ok
}

// SetNodesConfigPerEpoch sets nodes config for the specified epoch
func (ihncl *IndexHashedNodesCoordinatorLite) SetNodesConfigPerEpoch(epoch uint32, nodeConfig *EpochNodesConfig) {
	ihncl.mutNodesConfig.RLock()
	ihncl.nodesConfig[epoch] = nodeConfig
	ihncl.mutNodesConfig.RUnlock()
}

// GetNodesCoordinatorHelper
func (ihncl *IndexHashedNodesCoordinatorLite) GetNodesCoordinatorHelper() NodesCoordinatorHelper {
	return ihncl.nodesCoordinatorHelper
}

// SetNodesCoordinatorHelper
func (ihncl *IndexHashedNodesCoordinatorLite) SetNodesCoordinatorHelper(nch NodesCoordinatorHelper) {
	ihncl.nodesCoordinatorHelper = nch
}

// ShardIDAsObserver
func (ihncl *IndexHashedNodesCoordinatorLite) ShardIDAsObserver() uint32 {
	return ihncl.shardIDAsObserver
}

// ClearConsensusGroupCacher will clear the consensus group cacher
func (ihncl *IndexHashedNodesCoordinatorLite) ClearConsensusGroupCacher() {
	ihncl.consensusGroupCacher.Clear()
}

// GetOwnPublicKey will return current node public key  for block sign
func (ihncl *IndexHashedNodesCoordinatorLite) GetOwnPublicKey() []byte {
	return ihncl.selfPubKey
}

// IsInterfaceNil returns true if there is no value under the interface
func (ihncl *IndexHashedNodesCoordinatorLite) IsInterfaceNil() bool {
	return ihncl == nil
}
