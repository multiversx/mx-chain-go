package sharding

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"

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

type indexHashedNodesCoordinator struct {
	*nodesCoordinator.IndexHashedNodesCoordinatorLite

	epochStartRegistrationHandler nodesCoordinator.EpochStartEventNotifier
	loadingFromDisk               atomic.Value
	bootStorer                    storage.Storer
	startEpoch                    uint32
	savedStateKey                 []byte
	marshalizer                   marshal.Marshalizer
	shuffledOutHandler            ShuffledOutHandler
	mutNodesConfig                sync.RWMutex
	mutSavedStateKey              sync.RWMutex
}

// NewIndexHashedNodesCoordinator creates a new index hashed group selector
func NewIndexHashedNodesCoordinator(arguments ArgNodesCoordinator) (*indexHashedNodesCoordinator, error) {
	argumentsNodesCoordinatorLite := nodesCoordinator.ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    arguments.ShardConsensusGroupSize,
		MetaConsensusGroupSize:     arguments.MetaConsensusGroupSize,
		Hasher:                     arguments.Hasher,
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
		Shuffler:                   arguments.Shuffler,
	}

	ihgsLite, err := nodesCoordinator.NewIndexHashedNodesCoordinatorLite(argumentsNodesCoordinatorLite)
	if err != nil {
		return nil, err
	}

	err = checkArguments(arguments)
	if err != nil {
		return nil, err
	}

	savedKey := arguments.Hasher.Compute(string(arguments.SelfPublicKey))

	ihgs := &indexHashedNodesCoordinator{
		IndexHashedNodesCoordinatorLite: ihgsLite,
		epochStartRegistrationHandler:   arguments.EpochStartNotifier,
		bootStorer:                      arguments.BootStorer,
		startEpoch:                      arguments.StartEpoch,
		savedStateKey:                   savedKey,
		marshalizer:                     arguments.Marshalizer,
		shuffledOutHandler:              arguments.ShuffledOutHandler,
	}

	ihgs.SetNodesCoordinatorHelper(ihgs)

	ihgs.loadingFromDisk.Store(false)

	err = ihgs.saveState(ihgs.savedStateKey)
	if err != nil {
		log.Error("saving initial nodes coordinator config failed",
			"error", err.Error())
	}

	ihgs.epochStartRegistrationHandler.RegisterHandler(ihgs)

	return ihgs, nil
}

func checkArguments(arguments ArgNodesCoordinator) error {
	if check.IfNil(arguments.BootStorer) {
		return ErrNilBootStorer
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

	nodesConfig, ok := ihgs.GetNodesConfigPerEpoch(epoch)

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

	nodesConfig, ok := ihgs.GetNodesConfigPerEpoch(epoch)

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

	nodesConfig, ok := ihgs.GetNodesConfigPerEpoch(epoch)

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

	ihgs.UpdateEpochFlags(newEpoch)

	//TODO: remove the copy if no changes are done to the maps
	copiedPrevious := ihgs.GetPreviousConfigCopy()
	if copiedPrevious == nil {
		log.Error("previous nodes config is nil")
		return
	}

	// TODO: compare with previous nodesConfig if exists
	newNodesConfig, err := ihgs.ComputeNodesConfigFromList(copiedPrevious, allValidatorInfo)
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

	unStakeLeavingList := ihgs.CreateSortedListFromMap(newNodesConfig.LeavingMap)
	additionalLeavingList := ihgs.CreateSortedListFromMap(additionalLeavingMap)

	shufflerArgs := nodesCoordinator.ArgsUpdateNodes{
		Eligible:          newNodesConfig.EligibleMap,
		Waiting:           newNodesConfig.WaitingMap,
		NewNodes:          newNodesConfig.NewList,
		UnStakeLeaving:    unStakeLeavingList,
		AdditionalLeaving: additionalLeavingList,
		Rand:              randomness,
		NbShards:          newNodesConfig.NbShards,
		Epoch:             newEpoch,
	}

	resUpdateNodes, err := ihgs.UpdateNodeLists(shufflerArgs)
	if err != nil {
		log.Error("could not compute UpdateNodeLists - do nothing on nodesCoordinator epochStartPrepare", "err", err.Error())
		return
	}

	leavingNodesMap, stillRemainingNodesMap := nodesCoordinator.CreateActuallyLeavingPerShards(
		newNodesConfig.LeavingMap,
		additionalLeavingMap,
		resUpdateNodes.Leaving,
	)

	err = ihgs.SetNodesPerShards(resUpdateNodes.Eligible, resUpdateNodes.Waiting, leavingNodesMap, newEpoch)
	if err != nil {
		log.Error("set nodes per shard failed", "error", err.Error())
	}

	ihgs.FillPublicKeyToValidatorMap()
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

	ihgs.ClearConsensusGroupCacher()
}

// EpochStartAction is called upon a start of epoch event.
// NodeCoordinator has to get the nodes assignment to shards using the shuffler.
func (ihgs *indexHashedNodesCoordinator) EpochStartAction(hdr data.HeaderHandler) {
	newEpoch := hdr.GetEpoch()
	epochToRemove := int32(newEpoch) - nodesCoordinator.NodesCoordinatorStoredEpochs
	needToRemove := epochToRemove >= 0
	ihgs.SetCurrentEpoch(newEpoch)

	err := ihgs.saveState(ihgs.savedStateKey)
	if err != nil {
		log.Error("saving nodes coordinator config failed", "error", err.Error())
	}

	if needToRemove {
		ihgs.RemoveNodesConfigEpochs(epochToRemove)
	}
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

	nodesConfig, _ := ihgs.GetNodesConfigPerEpoch(epoch)

	if nodesConfig == nil {
		log.Warn("shuffleOutForEpoch failed",
			"epoch", epoch,
			"error", ErrEpochNodesConfigDoesNotExist)
		return
	}

	if isValidator(nodesConfig, ihgs.GetOwnPublicKey()) {
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

func searchInMap(validatorMap map[uint32][]validator, pk []byte) (bool, uint32) {
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

func (ihgs *indexHashedNodesCoordinator) computeShardForSelfPublicKey(nodesConfig *nodesCoordinator.EpochNodesConfig) (uint32, bool) {
	pubKey := ihgs.GetOwnPublicKey()
	selfShard := ihgs.ShardIDAsObserver()
	epNodesConfig, ok := ihgs.GetNodesConfigPerEpoch(ihgs.GetCurrentEpoch())
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

// IsInterfaceNil returns true if there is no value under the interface
func (ihgs *indexHashedNodesCoordinator) IsInterfaceNil() bool {
	return ihgs == nil
}
