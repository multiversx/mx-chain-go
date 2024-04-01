package nodesCoordinator

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
)

type sovereignIndexHashedNodesCoordinator struct {
	*indexHashedNodesCoordinator
}

func NewSovereignIndexHashedNodesCoordinator(arguments ArgNodesCoordinator) (*sovereignIndexHashedNodesCoordinator, error) {
	err := checkSovereignArguments(arguments)
	if err != nil {
		return nil, err
	}

	nodesConfig := make(map[uint32]*epochNodesConfig, nodesCoordinatorStoredEpochs)

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

	ihnc := &sovereignIndexHashedNodesCoordinator{
		indexHashedNodesCoordinator: &indexHashedNodesCoordinator{
			marshalizer:                     arguments.Marshalizer,
			hasher:                          arguments.Hasher,
			shuffler:                        arguments.Shuffler,
			epochStartRegistrationHandler:   arguments.EpochStartNotifier,
			bootStorer:                      arguments.BootStorer,
			selfPubKey:                      arguments.SelfPublicKey,
			nodesConfig:                     nodesConfig,
			currentEpoch:                    arguments.Epoch,
			savedStateKey:                   savedKey,
			shardConsensusGroupSize:         arguments.ShardConsensusGroupSize,
			metaConsensusGroupSize:          arguments.MetaConsensusGroupSize,
			consensusGroupCacher:            arguments.ConsensusGroupCache,
			shardIDAsObserver:               arguments.ShardIDAsObserver,
			shuffledOutHandler:              arguments.ShuffledOutHandler,
			startEpoch:                      arguments.StartEpoch,
			publicKeyToValidatorMap:         make(map[string]*validatorWithShardID),
			chanStopNode:                    arguments.ChanStopNode,
			nodeTypeProvider:                arguments.NodeTypeProvider,
			isFullArchive:                   arguments.IsFullArchive,
			enableEpochsHandler:             arguments.EnableEpochsHandler,
			validatorInfoCacher:             arguments.ValidatorInfoCacher,
			genesisNodesSetupHandler:        arguments.GenesisNodesSetupHandler,
			nodesCoordinatorRegistryFactory: arguments.NodesCoordinatorRegistryFactory,
		},
	}

	ihnc.loadingFromDisk.Store(false)

	ihnc.nodesCoordinatorHelper = ihnc
	err = ihnc.setNodesPerShards(arguments.EligibleNodes, arguments.WaitingNodes, nil, nil, arguments.Epoch)
	if err != nil {
		return nil, err
	}

	ihnc.fillPublicKeyToValidatorMap()
	err = ihnc.saveState(ihnc.savedStateKey, arguments.Epoch)
	if err != nil {
		log.Error("saving initial nodes coordinator config failed",
			"error", err.Error())
	}
	log.Info("new nodes config is set for epoch", "epoch", arguments.Epoch)
	currentNodesConfig := ihnc.nodesConfig[arguments.Epoch]
	if currentNodesConfig == nil {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, arguments.Epoch)
	}

	currentConfig := nodesConfig[arguments.Epoch]
	if currentConfig == nil {
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, arguments.Epoch)
	}

	displaySovereignNodesConfiguration(
		currentConfig.eligibleMap,
		currentConfig.waitingMap,
		currentConfig.leavingMap,
		make(map[uint32][]Validator))

	ihnc.epochStartRegistrationHandler.RegisterHandler(ihnc)
	return ihnc, nil
}

func checkSovereignArguments(arguments ArgNodesCoordinator) error {
	if arguments.ShardConsensusGroupSize < 1 {
		return ErrInvalidConsensusGroupSize
	}
	if arguments.NbShards != 1 {
		return ErrInvalidNumberOfShards
	}

	return checkNilArguments(arguments)
}

func (ihnc *sovereignIndexHashedNodesCoordinator) setNodesPerShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	leaving map[uint32][]Validator,
	shuffledOut map[uint32][]Validator,
	epoch uint32,
) error {
	ihnc.mutNodesConfig.Lock()
	defer ihnc.mutNodesConfig.Unlock()

	nodesConfig, ok := ihnc.nodesConfig[epoch]
	if !ok {
		log.Debug("Did not find nodesConfig", "epoch", epoch)
		nodesConfig = &epochNodesConfig{}
	}

	nodesConfig.mutNodesMaps.Lock()
	defer nodesConfig.mutNodesMaps.Unlock()

	if eligible == nil || waiting == nil {
		return ErrNilInputNodesMap
	}

	nbNodesShard := len(eligible[core.SovereignChainShardId])
	if nbNodesShard < ihnc.shardConsensusGroupSize {
		return ErrSmallShardEligibleListSize
	}
	numTotalEligible := uint64(nbNodesShard)

	err := ihnc.baseSetNodesPerShard(nodesConfig, numTotalEligible, eligible, waiting, leaving, shuffledOut, epoch)
	if err != nil {
		return err
	}

	nodesConfig.nbShards = 1
	return nil
}

// ComputeConsensusGroup will generate a list of validators based on the eligible list
// and each eligible validator weight/chance
func (ihnc *sovereignIndexHashedNodesCoordinator) ComputeConsensusGroup(
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
	nodesConfig, ok := ihnc.nodesConfig[epoch]
	if !ok {
		ihnc.mutNodesConfig.RUnlock()
		return nil, fmt.Errorf("%w epoch=%v", ErrEpochNodesConfigDoesNotExist, epoch)
	}

	if shardID != core.SovereignChainShardId {
		log.Warn("shardID is not ok, expected a sovereign chain id", "shardID", shardID, "nbShards", nodesConfig.nbShards)
		ihnc.mutNodesConfig.RUnlock()
		return nil, ErrInvalidShardId
	}
	selector = nodesConfig.selectors[shardID]
	eligibleList = nodesConfig.eligibleMap[shardID]
	ihnc.mutNodesConfig.RUnlock()

	return ihnc.baseComputeConsensusGroup(randomness, round, shardID, epoch, selector, eligibleList)
}

// GetConsensusValidatorsPublicKeys calculates the validators consensus group for a specific shard, randomness and round number,
// returning their public keys
func (ihnc *sovereignIndexHashedNodesCoordinator) GetConsensusValidatorsPublicKeys(
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

// EpochStartPrepare is not implemented for sovereign
func (ihnc *sovereignIndexHashedNodesCoordinator) EpochStartPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
	log.Error("sovereignIndexHashedNodesCoordinator.EpochStartPrepare was called, not implemented in sovereign")
}

func displaySovereignNodesConfiguration(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	leaving map[uint32][]Validator,
	actualRemaining map[uint32][]Validator,
) {
	shardID := core.SovereignChainShardId
	for _, v := range eligible[shardID] {
		pk := v.PubKey()
		log.Debug("eligible", "pk", pk, "shardID", shardID)
	}
	for _, v := range waiting[shardID] {
		pk := v.PubKey()
		log.Debug("waiting", "pk", pk, "shardID", shardID)
	}
	for _, v := range leaving[shardID] {
		pk := v.PubKey()
		log.Debug("leaving", "pk", pk, "shardID", shardID)
	}
	for _, v := range actualRemaining[shardID] {
		pk := v.PubKey()
		log.Debug("actual remaining", "pk", pk, "shardID", shardID)
	}

}
