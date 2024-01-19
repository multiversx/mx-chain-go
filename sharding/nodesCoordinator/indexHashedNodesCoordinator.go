package nodesCoordinator

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/multiversx/mx-chain-core-go/core"
	atomicFlags "github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ NodesCoordinator = (*indexHashedNodesCoordinator)(nil)
var _ PublicKeysSelector = (*indexHashedNodesCoordinator)(nil)

const (
	keyFormat               = "%s_%v_%v_%v"
	defaultSelectionChances = uint32(1)
	minStoredEpochs         = uint32(1)
	minEpochsToWait         = uint32(1)
)

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
	isFullArchive                 bool
	chanStopNode                  chan endProcess.ArgEndProcess
	flagWaitingListFix            atomicFlags.Flag
	nodeTypeProvider              NodeTypeProviderHandler
	enableEpochsHandler           common.EnableEpochsHandler
	validatorInfoCacher           epochStart.ValidatorInfoCacher
	numStoredEpochs               uint32
	nodesConfigCacher             Cacher
	epochStartStaticStorer        storage.Storer
	genesisNodesSetupHandler      GenesisNodesSetupHandler
}

// NewIndexHashedNodesCoordinator creates a new index hashed group selector
func NewIndexHashedNodesCoordinator(args ArgNodesCoordinator) (*indexHashedNodesCoordinator, error) {
	err := checkArguments(args)
	if err != nil {
		return nil, err
	}

	nodesConfig := make(map[uint32]*epochNodesConfig, args.NumStoredEpochs)

	nodesConfig[args.Epoch] = &epochNodesConfig{
		nbShards:    args.NbShards,
		shardID:     args.ShardIDAsObserver,
		eligibleMap: make(map[uint32][]Validator),
		waitingMap:  make(map[uint32][]Validator),
		selectors:   make(map[uint32]RandomSelector),
		leavingMap:  make(map[uint32][]Validator),
		newList:     make([]Validator, 0),
	}

	savedKey := args.Hasher.Compute(string(args.SelfPublicKey))

	ihnc := &indexHashedNodesCoordinator{
		marshalizer:                   args.Marshalizer,
		hasher:                        args.Hasher,
		shuffler:                      args.Shuffler,
		epochStartRegistrationHandler: args.EpochStartNotifier,
		bootStorer:                    args.BootStorer,
		selfPubKey:                    args.SelfPublicKey,
		nodesConfig:                   nodesConfig,
		currentEpoch:                  args.Epoch,
		savedStateKey:                 savedKey,
		shardConsensusGroupSize:       args.ShardConsensusGroupSize,
		metaConsensusGroupSize:        args.MetaConsensusGroupSize,
		consensusGroupCacher:          args.ConsensusGroupCache,
		shardIDAsObserver:             args.ShardIDAsObserver,
		shuffledOutHandler:            args.ShuffledOutHandler,
		startEpoch:                    args.StartEpoch,
		publicKeyToValidatorMap:       make(map[string]*validatorWithShardID),
		chanStopNode:                  args.ChanStopNode,
		nodeTypeProvider:              args.NodeTypeProvider,
		isFullArchive:                 args.IsFullArchive,
		enableEpochsHandler:           args.EnableEpochsHandler,
		validatorInfoCacher:           args.ValidatorInfoCacher,
		numStoredEpochs:               args.NumStoredEpochs,
		nodesConfigCacher:             args.NodesConfigCache,
		epochStartStaticStorer:        args.EpochStartStaticStorer,
		genesisNodesSetupHandler:      args.GenesisNodesSetupHandler,
	}

	ihnc.loadingFromDisk.Store(false)

	ihnc.nodesCoordinatorHelper = ihnc
	err = ihnc.setNodesPerShards(args.EligibleNodes, args.WaitingNodes, args.LeavingNodes, args.Epoch)
	if err != nil {
		return nil, err
	}

	ihnc.fillPublicKeyToValidatorMap()
	err = ihnc.saveState(ihnc.savedStateKey)
	if err != nil {
		log.Error("saving initial nodes coordinator config failed",
			"error", err.Error())
	}

	log.Info("new nodes config is set for epoch", "epoch", args.Epoch)
	currentConfig, ok := ihnc.getNodesConfig(args.Epoch)
	if !ok {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, args.Epoch)
	}

	displayNodesConfiguration(
		currentConfig.eligibleMap,
		currentConfig.waitingMap,
		currentConfig.leavingMap,
		make(map[uint32][]Validator),
		currentConfig.nbShards)

	ihnc.epochStartRegistrationHandler.RegisterHandler(ihnc)

	return ihnc, nil
}

func checkArguments(args ArgNodesCoordinator) error {
	if args.ShardConsensusGroupSize < 1 || args.MetaConsensusGroupSize < 1 {
		return ErrInvalidConsensusGroupSize
	}
	if args.NbShards < 1 {
		return ErrInvalidNumberOfShards
	}
	if args.ShardIDAsObserver >= args.NbShards && args.ShardIDAsObserver != core.MetachainShardId {
		return ErrInvalidShardId
	}
	if check.IfNil(args.Hasher) {
		return ErrNilHasher
	}
	if len(args.SelfPublicKey) == 0 {
		return ErrNilPubKey
	}
	if check.IfNil(args.Shuffler) {
		return ErrNilShuffler
	}
	if check.IfNil(args.BootStorer) {
		return ErrNilBootStorer
	}
	if check.IfNilReflect(args.ConsensusGroupCache) {
		return ErrNilCacher
	}
	if check.IfNil(args.Marshalizer) {
		return ErrNilMarshalizer
	}
	if check.IfNil(args.ShuffledOutHandler) {
		return ErrNilShuffledOutHandler
	}
	if check.IfNil(args.NodeTypeProvider) {
		return ErrNilNodeTypeProvider
	}
	if nil == args.ChanStopNode {
		return ErrNilNodeStopChannel
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return ErrNilEnableEpochsHandler
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.RefactorPeersMiniBlocksFlag,
		common.WaitingListFixFlag,
	})
	if err != nil {
		return err
	}
	if check.IfNil(args.ValidatorInfoCacher) {
		return ErrNilValidatorInfoCacher
	}
	if args.NumStoredEpochs < minStoredEpochs {
		return ErrInvalidNumberOfStoredEpochs
	}
	if check.IfNil(args.NodesConfigCache) {
		return ErrNilNodesConfigCacher
	}
	if check.IfNil(args.EpochStartStaticStorer) {
		return ErrNilEpochStartStaticStorer
	}
	if check.IfNil(args.GenesisNodesSetupHandler) {
		return ErrNilGenesisNodesSetupHandler
	}

	return nil
}

// getNodesConfig will try to get nodesConfig from map, if it doesn't succeed, it will try to get it from nodes config cache
// it has to be used under mutex
func (ihnc *indexHashedNodesCoordinator) getNodesConfig(epoch uint32) (*epochNodesConfig, bool) {
	nc, ok := ihnc.nodesConfig[epoch]
	if ok {
		return nc, ok
	}

	value, ok := ihnc.nodesConfigCacher.Get([]byte(fmt.Sprint(epoch)))
	if ok {
		enc, ok := value.(*epochNodesConfig)
		if ok {
			return enc, ok
		}
	}

	nodesConfig, err := ihnc.nodesConfigFromStaticStorer(epoch)
	if err != nil {
		return nil, false
	}

	ihnc.nodesConfigCacher.Put([]byte(fmt.Sprint(epoch)), nodesConfig, 0)

	return nodesConfig, true
}

// setNodesPerShards loads the distribution of nodes per shard into the nodes management component
func (ihnc *indexHashedNodesCoordinator) setNodesPerShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	leaving map[uint32][]Validator,
	epoch uint32,
) error {
	ihnc.mutNodesConfig.Lock()
	defer ihnc.mutNodesConfig.Unlock()

	nodesConfig, ok := ihnc.getNodesConfig(epoch)
	if !ok {
		log.Debug("Did not find nodesConfig", "epoch", epoch)
		nodesConfig = &epochNodesConfig{}
	}

	nodesConfig.mutNodesMaps.Lock()
	defer nodesConfig.mutNodesMaps.Unlock()

	if eligible == nil || waiting == nil {
		return ErrNilInputNodesMap
	}

	nodesList := eligible[core.MetachainShardId]
	if len(nodesList) < ihnc.metaConsensusGroupSize {
		return ErrSmallMetachainEligibleListSize
	}

	numTotalEligible := uint64(len(nodesList))
	for shardId := uint32(0); shardId < uint32(len(eligible)-1); shardId++ {
		nbNodesShard := len(eligible[shardId])
		if nbNodesShard < ihnc.shardConsensusGroupSize {
			return ErrSmallShardEligibleListSize
		}
		numTotalEligible += uint64(nbNodesShard)
	}

	var err error
	var isCurrentNodeValidator bool
	// nbShards holds number of shards without meta
	nodesConfig.nbShards = uint32(len(eligible) - 1)
	nodesConfig.eligibleMap = eligible
	nodesConfig.waitingMap = waiting
	nodesConfig.leavingMap = leaving
	nodesConfig.shardID, isCurrentNodeValidator = ihnc.computeShardForSelfPublicKey(nodesConfig)
	nodesConfig.selectors, err = ihnc.createSelectors(nodesConfig)
	if err != nil {
		return err
	}

	ihnc.nodesConfig[epoch] = nodesConfig

	ihnc.numTotalEligible = numTotalEligible
	ihnc.setNodeType(isCurrentNodeValidator)

	if ihnc.isFullArchive && isCurrentNodeValidator {
		ihnc.chanStopNode <- endProcess.ArgEndProcess{
			Reason:      common.WrongConfiguration,
			Description: ErrValidatorCannotBeFullArchive.Error(),
		}

		return nil
	}

	return nil
}

func (ihnc *indexHashedNodesCoordinator) setNodeType(isValidator bool) {
	if isValidator {
		ihnc.nodeTypeProvider.SetType(core.NodeTypeValidator)
		return
	}

	ihnc.nodeTypeProvider.SetType(core.NodeTypeObserver)
}

// ComputeAdditionalLeaving - computes extra leaving validators based on computation at the start of epoch
func (ihnc *indexHashedNodesCoordinator) ComputeAdditionalLeaving(_ []*state.ShardValidatorInfo) (map[uint32][]Validator, error) {
	return make(map[uint32][]Validator), nil
}

// ComputeConsensusGroup will generate a list of validators based on the eligible list
// and each eligible validator weight/chance
func (ihnc *indexHashedNodesCoordinator) ComputeConsensusGroup(
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

	ihnc.mutNodesConfig.RLock()
	nodesConfig, ok := ihnc.getNodesConfig(epoch)
	if ok {
		if shardID >= nodesConfig.nbShards && shardID != core.MetachainShardId {
			log.Warn("shardID is not ok", "shardID", shardID, "nbShards", nodesConfig.nbShards)
			ihnc.mutNodesConfig.RUnlock()
			return nil, ErrInvalidShardId
		}
		selector = nodesConfig.selectors[shardID]
		eligibleList = nodesConfig.eligibleMap[shardID]
	}
	ihnc.mutNodesConfig.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	key := []byte(fmt.Sprintf(keyFormat, string(randomness), round, shardID, epoch))
	validators := ihnc.searchConsensusForKey(key)
	if validators != nil {
		return validators, nil
	}

	consensusSize := ihnc.ConsensusGroupSize(shardID)
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

	ihnc.consensusGroupCacher.Put(key, tempList, size)

	return tempList, nil
}

func (ihnc *indexHashedNodesCoordinator) searchConsensusForKey(key []byte) []Validator {
	value, ok := ihnc.consensusGroupCacher.Get(key)
	if ok {
		consensusGroup, typeOk := value.([]Validator)
		if typeOk {
			return consensusGroup
		}
	}
	return nil
}

// GetValidatorWithPublicKey gets the validator with the given public key
func (ihnc *indexHashedNodesCoordinator) GetValidatorWithPublicKey(publicKey []byte) (Validator, uint32, error) {
	if len(publicKey) == 0 {
		return nil, 0, ErrNilPubKey
	}
	ihnc.mutNodesConfig.RLock()
	v, ok := ihnc.publicKeyToValidatorMap[string(publicKey)]
	ihnc.mutNodesConfig.RUnlock()
	if ok {
		return v.validator, v.shardID, nil
	}

	return nil, 0, ErrValidatorNotFound
}

// GetConsensusValidatorsPublicKeys calculates the validators consensus group for a specific shard, randomness and round number,
// returning their public keys
func (ihnc *indexHashedNodesCoordinator) GetConsensusValidatorsPublicKeys(
	randomness []byte,
	round uint64,
	shardID uint32,
	epoch uint32,
) ([]string, error) {
	consensusNodes, err := ihnc.ComputeConsensusGroup(randomness, round, shardID, epoch)
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
func (ihnc *indexHashedNodesCoordinator) GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihnc.mutNodesConfig.RLock()
	nodesConfig, ok := ihnc.getNodesConfig(epoch)
	ihnc.mutNodesConfig.RUnlock()

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
func (ihnc *indexHashedNodesCoordinator) GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihnc.mutNodesConfig.RLock()
	nodesConfig, ok := ihnc.getNodesConfig(epoch)
	ihnc.mutNodesConfig.RUnlock()

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
func (ihnc *indexHashedNodesCoordinator) GetAllLeavingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihnc.mutNodesConfig.RLock()
	nodesConfig, ok := ihnc.getNodesConfig(epoch)
	ihnc.mutNodesConfig.RUnlock()

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
func (ihnc *indexHashedNodesCoordinator) GetValidatorsIndexes(
	publicKeys []string,
	epoch uint32,
) ([]uint64, error) {
	signersIndexes := make([]uint64, 0)

	validatorsPubKeys, err := ihnc.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		return nil, err
	}

	ihnc.mutNodesConfig.RLock()
	nodesConfig, _ := ihnc.getNodesConfig(epoch)
	ihnc.mutNodesConfig.RUnlock()

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
func (ihnc *indexHashedNodesCoordinator) EpochStartPrepare(metaHdr data.HeaderHandler, body data.BodyHandler) {
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

	if check.IfNil(body) && newEpoch == ihnc.currentEpoch {
		log.Debug("nil body provided for epoch start prepare, it is normal in case of revertStateToBlock")
		return
	}

	ihnc.updateEpochFlags(newEpoch)

	allValidatorInfo, err := ihnc.createValidatorInfoFromBody(body, ihnc.numTotalEligible, newEpoch)
	if err != nil {
		log.Error("could not create validator info from body - do nothing on nodesCoordinator epochStartPrepare", "error", err.Error())
		return
	}

	ihnc.mutNodesConfig.RLock()
	previousConfig, _ := ihnc.getNodesConfig(ihnc.currentEpoch)
	if previousConfig == nil {
		log.Error("previous nodes config is nil")
		ihnc.mutNodesConfig.RUnlock()
		return
	}

	// TODO: remove the copy if no changes are done to the maps
	copiedPrevious := &epochNodesConfig{}
	copiedPrevious.eligibleMap = copyValidatorMap(previousConfig.eligibleMap)
	copiedPrevious.waitingMap = copyValidatorMap(previousConfig.waitingMap)
	copiedPrevious.nbShards = previousConfig.nbShards

	ihnc.mutNodesConfig.RUnlock()

	// TODO: compare with previous nodesConfig if exists
	newNodesConfig, err := ihnc.computeNodesConfigFromList(copiedPrevious, allValidatorInfo)
	if err != nil {
		log.Error("could not compute nodes config from list - do nothing on nodesCoordinator epochStartPrepare")
		return
	}

	if copiedPrevious.nbShards != newNodesConfig.nbShards {
		log.Warn("number of shards does not match",
			"previous epoch", ihnc.currentEpoch,
			"previous number of shards", copiedPrevious.nbShards,
			"new epoch", newEpoch,
			"new number of shards", newNodesConfig.nbShards)
	}

	additionalLeavingMap, err := ihnc.nodesCoordinatorHelper.ComputeAdditionalLeaving(allValidatorInfo)
	if err != nil {
		log.Error("could not compute additionalLeaving Nodes  - do nothing on nodesCoordinator epochStartPrepare")
		return
	}

	unStakeLeavingList := ihnc.createSortedListFromMap(newNodesConfig.leavingMap)
	additionalLeavingList := ihnc.createSortedListFromMap(additionalLeavingMap)

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

	resUpdateNodes, err := ihnc.shuffler.UpdateNodeLists(shufflerArgs)
	if err != nil {
		log.Error("could not compute UpdateNodeLists - do nothing on nodesCoordinator epochStartPrepare", "err", err.Error())
		return
	}

	leavingNodesMap, stillRemainingNodesMap := createActuallyLeavingPerShards(
		newNodesConfig.leavingMap,
		additionalLeavingMap,
		resUpdateNodes.Leaving,
	)

	err = ihnc.setNodesPerShards(resUpdateNodes.Eligible, resUpdateNodes.Waiting, leavingNodesMap, newEpoch)
	if err != nil {
		log.Error("set nodes per shard failed", "error", err.Error())
	}

	ihnc.fillPublicKeyToValidatorMap()
	err = ihnc.saveState(randomness)
	ihnc.handleErrorLog(err, "saving nodes coordinator config failed")

	displayNodesConfiguration(
		resUpdateNodes.Eligible,
		resUpdateNodes.Waiting,
		leavingNodesMap,
		stillRemainingNodesMap,
		newNodesConfig.nbShards)

	ihnc.mutSavedStateKey.Lock()
	ihnc.savedStateKey = randomness
	ihnc.mutSavedStateKey.Unlock()

	ihnc.consensusGroupCacher.Clear()
}

func (ihnc *indexHashedNodesCoordinator) fillPublicKeyToValidatorMap() {
	ihnc.mutNodesConfig.Lock()
	defer ihnc.mutNodesConfig.Unlock()

	index := 0
	epochList := make([]uint32, len(ihnc.nodesConfig))
	mapAllValidators := make(map[uint32]map[string]*validatorWithShardID)
	for epoch, epochConfig := range ihnc.nodesConfig {
		epochConfig.mutNodesMaps.RLock()
		mapAllValidators[epoch] = ihnc.createPublicKeyToValidatorMap(epochConfig.eligibleMap, epochConfig.waitingMap)
		epochConfig.mutNodesMaps.RUnlock()

		epochList[index] = epoch
		index++
	}

	sort.Slice(epochList, func(i, j int) bool {
		return epochList[i] < epochList[j]
	})

	ihnc.publicKeyToValidatorMap = make(map[string]*validatorWithShardID)
	for _, epoch := range epochList {
		validatorsForEpoch := mapAllValidators[epoch]
		for pubKey, vInfo := range validatorsForEpoch {
			ihnc.publicKeyToValidatorMap[pubKey] = vInfo
		}
	}
}

func (ihnc *indexHashedNodesCoordinator) createSortedListFromMap(validatorsMap map[uint32][]Validator) []Validator {
	sortedList := make([]Validator, 0)
	for _, validators := range validatorsMap {
		sortedList = append(sortedList, validators...)
	}
	sort.Sort(validatorList(sortedList))
	return sortedList
}

// GetChance will return default chance
func (ihnc *indexHashedNodesCoordinator) GetChance(_ uint32) uint32 {
	return defaultSelectionChances
}

func (ihnc *indexHashedNodesCoordinator) computeNodesConfigFromList(
	previousEpochConfig *epochNodesConfig,
	validatorInfos []*state.ShardValidatorInfo,
) (*epochNodesConfig, error) {
	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	leavingMap := make(map[uint32][]Validator)
	newNodesList := make([]Validator, 0)

	if ihnc.flagWaitingListFix.IsSet() && previousEpochConfig == nil {
		return nil, ErrNilPreviousEpochConfig
	}

	if len(validatorInfos) == 0 {
		log.Warn("computeNodesConfigFromList - validatorInfos len is 0")
	}

	for _, validatorInfo := range validatorInfos {
		chance := ihnc.nodesCoordinatorHelper.GetChance(validatorInfo.TempRating)
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
			ihnc.addValidatorToPreviousMap(
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

	newNodesConfig := &epochNodesConfig{
		eligibleMap: eligibleMap,
		waitingMap:  waitingMap,
		leavingMap:  leavingMap,
		newList:     newNodesList,
		nbShards:    uint32(nbShards),
	}

	return newNodesConfig, nil
}

func (ihnc *indexHashedNodesCoordinator) addValidatorToPreviousMap(
	previousEpochConfig *epochNodesConfig,
	eligibleMap map[uint32][]Validator,
	waitingMap map[uint32][]Validator,
	currentValidator *validator,
	currentValidatorShardId uint32) {

	if !ihnc.flagWaitingListFix.IsSet() {
		eligibleMap[currentValidatorShardId] = append(eligibleMap[currentValidatorShardId], currentValidator)
		return
	}

	found, shardId := searchInMap(previousEpochConfig.eligibleMap, currentValidator.PubKey())
	if found {
		log.Debug("leaving node found in", "list", "eligible", "shardId", shardId)
		eligibleMap[shardId] = append(eligibleMap[currentValidatorShardId], currentValidator)
		return
	}

	found, shardId = searchInMap(previousEpochConfig.waitingMap, currentValidator.PubKey())
	if found {
		log.Debug("leaving node found in", "list", "waiting", "shardId", shardId)
		waitingMap[shardId] = append(waitingMap[currentValidatorShardId], currentValidator)
		return
	}
}

func (ihnc *indexHashedNodesCoordinator) handleErrorLog(err error, message string) {
	if err == nil {
		return
	}

	logLevel := logger.LogError
	if core.IsClosingError(err) {
		logLevel = logger.LogDebug
	}

	log.Log(logLevel, message, "error", err.Error())
}

// EpochStartAction is called upon a start of epoch event.
// NodeCoordinator has to get the nodes assignment to shards using the shuffler.
func (ihnc *indexHashedNodesCoordinator) EpochStartAction(hdr data.HeaderHandler) {
	newEpoch := hdr.GetEpoch()
	epochToRemove := int32(newEpoch) - int32(ihnc.numStoredEpochs)
	needToRemove := epochToRemove >= 0
	ihnc.currentEpoch = newEpoch

	err := ihnc.saveState(ihnc.savedStateKey)
	ihnc.handleErrorLog(err, "saving nodes coordinator config failed")

	ihnc.mutNodesConfig.Lock()
	if needToRemove {
		for epoch := range ihnc.nodesConfig {
			if epoch <= uint32(epochToRemove) {
				delete(ihnc.nodesConfig, epoch)
			}
		}
	}
	ihnc.mutNodesConfig.Unlock()
}

// NotifyOrder returns the notification order for a start of epoch event
func (ihnc *indexHashedNodesCoordinator) NotifyOrder() uint32 {
	return common.NodesCoordinatorOrder
}

// GetSavedStateKey returns the key for the last nodes coordinator saved state
func (ihnc *indexHashedNodesCoordinator) GetSavedStateKey() []byte {
	ihnc.mutSavedStateKey.RLock()
	key := ihnc.savedStateKey
	ihnc.mutSavedStateKey.RUnlock()

	return key
}

// ShardIdForEpoch returns the nodesCoordinator configured ShardId for specified epoch if epoch configuration exists,
// otherwise error
func (ihnc *indexHashedNodesCoordinator) ShardIdForEpoch(epoch uint32) (uint32, error) {
	ihnc.mutNodesConfig.RLock()
	nodesConfig, ok := ihnc.getNodesConfig(epoch)
	ihnc.mutNodesConfig.RUnlock()

	if !ok {
		return 0, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	return nodesConfig.shardID, nil
}

// ShuffleOutForEpoch verifies if the shards changed in the new epoch and calls the shuffleOutHandler
func (ihnc *indexHashedNodesCoordinator) ShuffleOutForEpoch(epoch uint32) {
	log.Debug("shuffle out called for", "epoch", epoch)

	ihnc.mutNodesConfig.Lock()
	nodesConfig, _ := ihnc.getNodesConfig(epoch)
	ihnc.mutNodesConfig.Unlock()

	if nodesConfig == nil {
		log.Warn("shuffleOutForEpoch failed",
			"epoch", epoch,
			"error", ErrEpochNodesConfigDoesNotExist)
		return
	}

	if isValidator(nodesConfig, ihnc.selfPubKey) {
		err := ihnc.shuffledOutHandler.Process(nodesConfig.shardID)
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
func (ihnc *indexHashedNodesCoordinator) GetConsensusWhitelistedNodes(
	epoch uint32,
) (map[string]struct{}, error) {
	var err error
	shardEligible := make(map[string]struct{})
	publicKeysPrevEpoch := make(map[uint32][][]byte)
	prevEpochConfigExists := false

	if epoch > ihnc.startEpoch {
		publicKeysPrevEpoch, err = ihnc.GetAllEligibleValidatorsPublicKeys(epoch - 1)
		if err == nil {
			prevEpochConfigExists = true
		} else {
			log.Warn("get consensus whitelisted nodes", "error", err.Error())
		}
	}

	var prevEpochShardId uint32
	if prevEpochConfigExists {
		prevEpochShardId, err = ihnc.ShardIdForEpoch(epoch - 1)
		if err == nil {
			for _, pubKey := range publicKeysPrevEpoch[prevEpochShardId] {
				shardEligible[string(pubKey)] = struct{}{}
			}
		} else {
			log.Trace("not critical error getting shardID for epoch", "epoch", epoch-1, "error", err)
		}
	}

	publicKeysNewEpoch, errGetEligible := ihnc.GetAllEligibleValidatorsPublicKeys(epoch)
	if errGetEligible != nil {
		return nil, errGetEligible
	}

	epochShardId, errShardIdForEpoch := ihnc.ShardIdForEpoch(epoch)
	if errShardIdForEpoch != nil {
		return nil, errShardIdForEpoch
	}

	for _, pubKey := range publicKeysNewEpoch[epochShardId] {
		shardEligible[string(pubKey)] = struct{}{}
	}

	return shardEligible, nil
}

func (ihnc *indexHashedNodesCoordinator) createPublicKeyToValidatorMap(
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

func (ihnc *indexHashedNodesCoordinator) computeShardForSelfPublicKey(nodesConfig *epochNodesConfig) (uint32, bool) {
	pubKey := ihnc.selfPubKey
	selfShard := ihnc.shardIDAsObserver
	epNodesConfig, ok := ihnc.nodesConfig[ihnc.currentEpoch]
	if ok {
		log.Trace("computeShardForSelfPublicKey found existing config",
			"shard", epNodesConfig.shardID,
		)
		selfShard = epNodesConfig.shardID
	}

	found, shardId := searchInMap(nodesConfig.eligibleMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in eligible",
			"epoch", ihnc.currentEpoch,
			"shard", shardId,
			"validator PK", pubKey,
		)
		return shardId, true
	}

	found, shardId = searchInMap(nodesConfig.waitingMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in waiting",
			"epoch", ihnc.currentEpoch,
			"shard", shardId,
			"validator PK", pubKey,
		)
		return shardId, true
	}

	found, shardId = searchInMap(nodesConfig.leavingMap, pubKey)
	if found {
		log.Trace("computeShardForSelfPublicKey found validator in leaving",
			"epoch", ihnc.currentEpoch,
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
func (ihnc *indexHashedNodesCoordinator) ConsensusGroupSize(
	shardID uint32,
) int {
	if shardID == core.MetachainShardId {
		return ihnc.metaConsensusGroupSize
	}

	return ihnc.shardConsensusGroupSize
}

// GetNumTotalEligible returns the number of total eligible accross all shards from current setup
func (ihnc *indexHashedNodesCoordinator) GetNumTotalEligible() uint64 {
	return ihnc.numTotalEligible
}

// GetOwnPublicKey will return current node public key  for block sign
func (ihnc *indexHashedNodesCoordinator) GetOwnPublicKey() []byte {
	return ihnc.selfPubKey
}

// IsInterfaceNil returns true if there is no value under the interface
func (ihnc *indexHashedNodesCoordinator) IsInterfaceNil() bool {
	return ihnc == nil
}

// createSelectors creates the consensus group selectors for each shard
// Not concurrent safe, needs to be called under mutex
func (ihnc *indexHashedNodesCoordinator) createSelectors(
	nodesConfig *epochNodesConfig,
) (map[uint32]RandomSelector, error) {
	var err error
	var weights []uint32

	selectors := make(map[uint32]RandomSelector)
	// weights for validators are computed according to each validator rating
	for shard, vList := range nodesConfig.eligibleMap {
		log.Debug("create selectors", "shard", shard)
		weights, err = ihnc.nodesCoordinatorHelper.ValidatorsWeights(vList)
		if err != nil {
			return nil, err
		}

		selectors[shard], err = NewSelectorExpandedList(weights, ihnc.hasher)
		if err != nil {
			return nil, err
		}
	}

	return selectors, nil
}

// ValidatorsWeights returns the weights/chances for each of the given validators
func (ihnc *indexHashedNodesCoordinator) ValidatorsWeights(validators []Validator) ([]uint32, error) {
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
func (ihnc *indexHashedNodesCoordinator) createValidatorInfoFromBody(
	body data.BodyHandler,
	previousTotal uint64,
	epoch uint32,
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
			shardValidatorInfo, err := ihnc.getShardValidatorInfoData(txHash, epoch)
			if err != nil {
				return nil, err
			}

			allValidatorInfo = append(allValidatorInfo, shardValidatorInfo)
		}
	}

	return allValidatorInfo, nil
}

func (ihnc *indexHashedNodesCoordinator) getShardValidatorInfoData(txHash []byte, epoch uint32) (*state.ShardValidatorInfo, error) {
	if ihnc.enableEpochsHandler.IsFlagEnabledInEpoch(common.RefactorPeersMiniBlocksFlag, epoch) {
		shardValidatorInfo, err := ihnc.validatorInfoCacher.GetValidatorInfo(txHash)
		if err != nil {
			return nil, err
		}

		return shardValidatorInfo, nil
	}

	shardValidatorInfo := &state.ShardValidatorInfo{}
	err := ihnc.marshalizer.Unmarshal(shardValidatorInfo, txHash)
	if err != nil {
		return nil, err
	}

	return shardValidatorInfo, nil
}

func (ihnc *indexHashedNodesCoordinator) updateEpochFlags(epoch uint32) {
	ihnc.flagWaitingListFix.SetValue(epoch >= ihnc.enableEpochsHandler.GetActivationEpoch(common.WaitingListFixFlag))
	log.Debug("indexHashedNodesCoordinator: waiting list fix", "enabled", ihnc.flagWaitingListFix.IsSet())
}

// GetWaitingEpochsLeftForPublicKey returns the number of epochs left for the public key until it becomes eligible
func (ihnc *indexHashedNodesCoordinator) GetWaitingEpochsLeftForPublicKey(publicKey []byte) (uint32, error) {
	if len(publicKey) == 0 {
		return 0, ErrNilPubKey
	}

	currentEpoch := ihnc.enableEpochsHandler.GetCurrentEpoch()

	ihnc.mutNodesConfig.RLock()
	nodesConfig, ok := ihnc.nodesConfig[currentEpoch]
	ihnc.mutNodesConfig.RUnlock()

	if !ok {
		return 0, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, currentEpoch)
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	for shardId, shardWaiting := range nodesConfig.waitingMap {
		epochsLeft, err := ihnc.searchWaitingEpochsLeftForPublicKeyInShard(publicKey, shardId, shardWaiting)
		if err != nil {
			continue
		}

		return epochsLeft, err
	}

	return 0, ErrKeyNotFoundInWaitingList
}

func (ihnc *indexHashedNodesCoordinator) searchWaitingEpochsLeftForPublicKeyInShard(publicKey []byte, shardId uint32, shardWaiting []Validator) (uint32, error) {
	for idx, val := range shardWaiting {
		if !bytes.Equal(val.PubKey(), publicKey) {
			continue
		}

		minHysteresisNodes := ihnc.getMinHysteresisNodes(shardId)
		if minHysteresisNodes == 0 {
			return minEpochsToWait, nil
		}

		return uint32(idx)/minHysteresisNodes + minEpochsToWait, nil
	}

	return 0, ErrKeyNotFoundInWaitingList
}

func (ihnc *indexHashedNodesCoordinator) getMinHysteresisNodes(shardId uint32) uint32 {
	if shardId == common.MetachainShardId {
		return ihnc.genesisNodesSetupHandler.MinMetaHysteresisNodes()
	}

	return ihnc.genesisNodesSetupHandler.MinShardHysteresisNodes()
}
