package bootstrap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/trie/factory"
)

func (e *epochStartBootstrap) initializeFromLocalStorage() {
	latestData, errNotCritical := e.latestStorageDataProvider.Get()
	if errNotCritical != nil {
		e.baseData.storageExists = false
		log.Debug("no epoch db found in storage", "error", errNotCritical.Error())
		return
	}

	if latestData.Epoch < e.startEpoch {
		e.baseData.storageExists = false
		log.Warn("data is older then start epoch", "latestEpoch", latestData.Epoch, "startEpoch", e.startEpoch)
		return
	}

	e.baseData.storageExists = true
	e.baseData.lastEpoch = latestData.Epoch
	e.baseData.shardId = latestData.ShardID
	e.baseData.lastRound = latestData.LastRound
	e.baseData.epochStartRound = latestData.EpochStartRound
	log.Debug("got last data from storage",
		"epoch", e.baseData.lastEpoch,
		"last round", e.baseData.lastRound,
		"last shard ID", e.baseData.shardId,
		"epoch start Round", e.baseData.epochStartRound)
}

func (e *epochStartBootstrap) getShardIDForLatestEpoch() (uint32, bool, error) {
	storer, err := e.storageOpenerHandler.GetMostRecentStorageUnit(e.generalConfig.BootstrapStorage.DB)
	defer func() {
		if check.IfNil(storer) {
			return
		}

		errClose := storer.Close()
		log.LogIfError(errClose)
	}()

	if err != nil {
		return 0, false, err
	}

	_, e.nodesConfig, err = e.getLastBootstrapData(storer)
	if err != nil {
		return 0, false, err
	}

	pubKey, err := e.cryptoComponentsHolder.PublicKey().ToByteArray()
	if err != nil {
		return 0, false, err
	}

	e.epochStartMeta, err = e.getEpochStartMetaFromStorage(storer)
	if err != nil {
		return 0, false, err
	}

	e.baseData.numberOfShards = uint32(len(e.epochStartMeta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers()))
	if e.baseData.numberOfShards == 0 {
		e.baseData.numberOfShards = e.genesisShardCoordinator.NumberOfShards()
	}

	newShardId, isShuffledOut := e.checkIfShuffledOut(pubKey, e.nodesConfig)
	modifiedShardId := e.applyShardIDAsObserverIfNeeded(newShardId)
	if newShardId != modifiedShardId {
		isShuffledOut = true
	}

	return modifiedShardId, isShuffledOut, nil
}

func (e *epochStartBootstrap) prepareEpochFromStorage() (Parameters, error) {
	newShardId, isShuffledOut, err := e.getShardIDForLatestEpoch()
	if err != nil {
		return Parameters{}, err
	}

	if !isShuffledOut {
		parameters := Parameters{
			Epoch:       e.baseData.lastEpoch,
			SelfShardId: e.baseData.shardId,
			NumOfShards: e.baseData.numberOfShards,
			NodesConfig: e.nodesConfig,
		}
		e.setEpochStartMetrics()

		return parameters, nil
	}

	e.closeTrieComponents()
	e.storageService = disabled.NewChainStorer()
	triesContainer, trieStorageManagers, err := factory.CreateTriesComponentsForShardId(
		e.generalConfig,
		e.coreComponentsHolder,
		e.storageService,
		e.stateStatsHandler,
	)
	if err != nil {
		return Parameters{}, err
	}

	e.trieContainer = triesContainer
	e.trieStorageManagers = trieStorageManagers

	e.shuffledOut = isShuffledOut
	log.Debug("prepareEpochFromStorage for shuffled out", "initial shard id", e.baseData.shardId, "new shard id", newShardId)
	e.baseData.shardId = newShardId

	err = e.createRequestHandler()
	if err != nil {
		return Parameters{}, err
	}

	err = e.createSyncers()
	if err != nil {
		return Parameters{}, err
	}

	defer func() {
		errClose := e.mainInterceptorContainer.Close()
		if errClose != nil {
			log.Warn("prepareEpochFromStorage mainInterceptorContainer.Close()", "error", errClose)
		}
		errClose = e.fullArchiveInterceptorContainer.Close()
		if errClose != nil {
			log.Warn("prepareEpochFromStorage fullArchiveInterceptorContainer.Close()", "error", errClose)
		}
	}()

	e.syncedHeaders, err = e.syncHeadersFrom(e.epochStartMeta)
	if err != nil {
		return Parameters{}, err
	}

	prevEpochStartMetaHash := e.epochStartMeta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash()
	prevEpochStartMeta, ok := e.syncedHeaders[string(prevEpochStartMetaHash)].(*block.MetaBlock)
	if !ok {
		return Parameters{}, epochStart.ErrWrongTypeAssertion
	}
	e.prevEpochStartMeta = prevEpochStartMeta

	e.shardCoordinator, err = e.shardCoordinatorFactory.CreateShardCoordinator(e.baseData.numberOfShards, e.baseData.shardId)
	if err != nil {
		return Parameters{}, err
	}

	consensusTopic := common.ConsensusTopic + e.shardCoordinator.CommunicationIdentifier(e.shardCoordinator.SelfId())
	err = e.mainMessenger.CreateTopic(consensusTopic, true)
	if err != nil {
		return Parameters{}, err
	}

	emptyPeerMiniBlocksSlice := make([]*block.MiniBlock, 0) // empty slice since we have bootstrapped from storage
	if e.shardCoordinator.SelfId() == core.MetachainShardId {
		err = e.requestAndProcessForMeta(emptyPeerMiniBlocksSlice)
		if err != nil {
			return Parameters{}, err
		}
	} else {
		err = e.requestAndProcessForShard(emptyPeerMiniBlocksSlice)
		if err != nil {
			return Parameters{}, err
		}
	}

	shardIDToReturn := e.applyShardIDAsObserverIfNeeded(e.shardCoordinator.SelfId())
	parameters := Parameters{
		Epoch:       e.baseData.lastEpoch,
		SelfShardId: shardIDToReturn,
		NumOfShards: e.shardCoordinator.NumberOfShards(),
		NodesConfig: e.nodesConfig,
	}
	return parameters, nil
}

func (e *epochStartBootstrap) checkIfShuffledOut(
	pubKey []byte,
	nodesConfig *nodesCoordinator.NodesCoordinatorRegistry,
) (uint32, bool) {
	epochIDasString := fmt.Sprint(e.baseData.lastEpoch)
	epochConfig := nodesConfig.EpochsConfig[epochIDasString]
	if epochConfig == nil {
		return e.baseData.shardId, false
	}

	newShardId, isWaitingForShard := checkIfPubkeyIsInMap(pubKey, epochConfig.WaitingValidators)
	if isWaitingForShard {
		isShuffledOut := newShardId != e.baseData.shardId
		e.nodeType = core.NodeTypeValidator
		return newShardId, isShuffledOut
	}

	newShardId, isEligibleForShard := checkIfPubkeyIsInMap(pubKey, epochConfig.EligibleValidators)
	if isEligibleForShard {
		isShuffledOut := newShardId != e.baseData.shardId
		e.nodeType = core.NodeTypeValidator
		return newShardId, isShuffledOut
	}

	return e.baseData.shardId, false
}

func checkIfPubkeyIsInMap(
	pubKey []byte,
	allShardList map[string][]*nodesCoordinator.SerializableValidator,
) (uint32, bool) {
	for shardIdStr, validatorList := range allShardList {
		isValidatorInList := checkIfValidatorIsInList(pubKey, validatorList)
		if isValidatorInList {
			shardId, err := strconv.ParseInt(shardIdStr, 10, 64)
			if err != nil {
				log.Error("checkIfIsValidatorForEpoch parsing string to int error should not happen", "err", err)
				return 0, false
			}

			return uint32(shardId), true
		}
	}
	return 0, false
}

func checkIfValidatorIsInList(
	pubKey []byte,
	validatorList []*nodesCoordinator.SerializableValidator,
) bool {
	for _, validator := range validatorList {
		if bytes.Equal(pubKey, validator.PubKey) {
			return true
		}
	}
	return false
}

func (e *epochStartBootstrap) getLastBootstrapData(storer storage.Storer) (*bootstrapStorage.BootstrapData, *nodesCoordinator.NodesCoordinatorRegistry, error) {
	bootStorer, err := bootstrapStorage.NewBootstrapStorer(e.coreComponentsHolder.InternalMarshalizer(), storer)
	if err != nil {
		return nil, nil, err
	}

	highestRound := bootStorer.GetHighestRound()
	bootstrapData, err := bootStorer.Get(highestRound)
	if err != nil {
		return nil, nil, err
	}

	ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), bootstrapData.NodesCoordinatorConfigKey...)
	d, err := storer.SearchFirst(ncInternalkey)
	if err != nil {
		log.Debug("getLastBootstrapData", "key", ncInternalkey, "error", err)
		return nil, nil, err
	}

	config := &nodesCoordinator.NodesCoordinatorRegistry{}
	err = json.Unmarshal(d, config)
	if err != nil {
		return nil, nil, err
	}

	return &bootstrapData, config, nil
}

func (e *epochStartBootstrap) getEpochStartMetaFromStorage(storer storage.Storer) (data.MetaHeaderHandler, error) {
	epochIdentifier := core.EpochStartIdentifier(e.baseData.lastEpoch)
	epochStartMetaBlock, err := storer.SearchFirst([]byte(epochIdentifier))
	if err != nil {
		log.Debug("getEpochStartMetaFromStorage", "key", epochIdentifier, "error", err)
		return nil, err
	}

	metaBlock := &block.MetaBlock{}
	err = e.coreComponentsHolder.InternalMarshalizer().Unmarshal(metaBlock, epochStartMetaBlock)
	if err != nil {
		return nil, err
	}

	return metaBlock, nil
}
