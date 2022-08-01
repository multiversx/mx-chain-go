package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	factoryDataPool "github.com/ElrondNetwork/elrond-go/dataRetriever/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	storageResolversContainers "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/storageResolversContainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/sharding"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/trie/factory"
)

// ArgsStorageEpochStartBootstrap holds the arguments needed for creating an epoch start data provider component
// from storage
type ArgsStorageEpochStartBootstrap struct {
	ArgsEpochStartBootstrap
	ImportDbConfig             config.ImportDbConfig
	ChanGracefullyClose        chan endProcess.ArgEndProcess
	TimeToWaitForRequestedData time.Duration
}

type storageEpochStartBootstrap struct {
	*epochStartBootstrap
	resolvers                  dataRetriever.ResolversContainer
	store                      dataRetriever.StorageService
	importDbConfig             config.ImportDbConfig
	chanGracefullyClose        chan endProcess.ArgEndProcess
	chainID                    string
	timeToWaitForRequestedData time.Duration
}

// NewStorageEpochStartBootstrap will return a new instance of storageEpochStartBootstrap that can bootstrap
// the node with the help of storage resolvers through the import-db process
func NewStorageEpochStartBootstrap(args ArgsStorageEpochStartBootstrap) (*storageEpochStartBootstrap, error) {
	esb, err := NewEpochStartBootstrap(args.ArgsEpochStartBootstrap)
	if err != nil {
		return nil, err
	}

	if args.ChanGracefullyClose == nil {
		return nil, dataRetriever.ErrNilGracefullyCloseChannel
	}

	sesb := &storageEpochStartBootstrap{
		epochStartBootstrap:        esb,
		importDbConfig:             args.ImportDbConfig,
		chanGracefullyClose:        args.ChanGracefullyClose,
		chainID:                    args.CoreComponentsHolder.ChainID(),
		timeToWaitForRequestedData: args.TimeToWaitForRequestedData,
	}

	return sesb, nil
}

// Bootstrap runs the fast bootstrap method from local storage or from import-db directory
func (sesb *storageEpochStartBootstrap) Bootstrap() (Parameters, error) {
	defer sesb.closeTrieComponents()

	if !sesb.generalConfig.GeneralSettings.StartInEpochEnabled {
		return sesb.bootstrapFromLocalStorage()
	}

	defer func() {
		sesb.cleanupOnBootstrapFinish()

		if !check.IfNil(sesb.resolvers) {
			err := sesb.resolvers.Close()
			if err != nil {
				log.Debug("non critical error closing resolvers", "error", err)
			}
		}

		if !check.IfNil(sesb.store) {
			err := sesb.store.CloseAll()
			if err != nil {
				log.Debug("non critical error closing storage service", "error", err)
			}
		}
	}()

	var err error
	sesb.shardCoordinator, err = sharding.NewMultiShardCoordinator(sesb.genesisShardCoordinator.NumberOfShards(), core.MetachainShardId)
	if err != nil {
		return Parameters{}, err
	}

	sesb.dataPool, err = factoryDataPool.NewDataPoolFromConfig(
		factoryDataPool.ArgsDataPool{
			Config:           &sesb.generalConfig,
			EconomicsData:    sesb.economicsData,
			ShardCoordinator: sesb.shardCoordinator,
			Marshalizer:      sesb.coreComponentsHolder.InternalMarshalizer(),
			PathManager:      sesb.coreComponentsHolder.PathHandler(),
		},
	)
	if err != nil {
		return Parameters{}, err
	}

	params, shouldContinue, err := sesb.startFromSavedEpoch()
	if !shouldContinue {
		return params, err
	}

	err = sesb.prepareComponentsToSync()
	if err != nil {
		return Parameters{}, err
	}

	sesb.epochStartMeta, err = sesb.epochStartMetaBlockSyncer.SyncEpochStartMeta(sesb.timeToWaitForRequestedData)
	if err != nil {
		return Parameters{}, err
	}
	log.Debug("start in epoch bootstrap: got epoch start meta header from storage", "epoch", sesb.epochStartMeta.GetEpoch(), "nonce", sesb.epochStartMeta.GetNonce())
	sesb.setEpochStartMetrics()

	err = sesb.createSyncers()
	if err != nil {
		return Parameters{}, err
	}

	params, err = sesb.requestAndProcessFromStorage()
	if err != nil {
		return Parameters{}, err
	}

	return params, nil
}

func (sesb *storageEpochStartBootstrap) prepareComponentsToSync() error {
	sesb.closeTrieComponents()
	sesb.storageService = disabled.NewChainStorer()
	triesContainer, trieStorageManagers, err := factory.CreateTriesComponentsForShardId(
		sesb.generalConfig,
		sesb.coreComponentsHolder,
		sesb.storageService,
	)
	if err != nil {
		return err
	}

	sesb.trieContainer = triesContainer
	sesb.trieStorageManagers = trieStorageManagers

	err = sesb.createStorageRequestHandler()
	if err != nil {
		return err
	}

	metablockProcessor, err := NewStorageEpochStartMetaBlockProcessor(
		sesb.messenger,
		sesb.requestHandler,
		sesb.coreComponentsHolder.InternalMarshalizer(),
		sesb.coreComponentsHolder.Hasher(),
	)
	if err != nil {
		return err
	}

	argsEpochStartSyncer := ArgsNewEpochStartMetaSyncer{
		CoreComponentsHolder:    sesb.coreComponentsHolder,
		CryptoComponentsHolder:  sesb.cryptoComponentsHolder,
		RequestHandler:          sesb.requestHandler,
		Messenger:               sesb.messenger,
		ShardCoordinator:        sesb.shardCoordinator,
		EconomicsData:           sesb.economicsData,
		WhitelistHandler:        sesb.whiteListHandler,
		StartInEpochConfig:      sesb.generalConfig.EpochStartConfig,
		HeaderIntegrityVerifier: sesb.headerIntegrityVerifier,
		MetaBlockProcessor:      metablockProcessor,
	}

	sesb.epochStartMetaBlockSyncer, err = NewEpochStartMetaSyncer(argsEpochStartSyncer)
	if err != nil {
		return err
	}

	return nil
}

func (sesb *storageEpochStartBootstrap) createStorageRequestHandler() error {
	err := sesb.createStorageResolvers()
	if err != nil {
		return err
	}

	finder, err := containers.NewResolversFinder(sesb.resolvers, sesb.shardCoordinator)
	if err != nil {
		return err
	}

	requestedItemsHandler := timecache.NewTimeCache(timeBetweenRequests)
	sesb.requestHandler, err = requestHandlers.NewResolverRequestHandler(
		finder,
		requestedItemsHandler,
		sesb.whiteListHandler,
		maxToRequest,
		core.MetachainShardId,
		timeBetweenRequests,
	)
	return err
}

func (sesb *storageEpochStartBootstrap) createStorageResolvers() error {
	dataPacker, err := partitioning.NewSimpleDataPacker(sesb.coreComponentsHolder.InternalMarshalizer())
	if err != nil {
		return err
	}

	shardCoordinator, err := sharding.NewMultiShardCoordinator(sesb.genesisShardCoordinator.NumberOfShards(), sesb.genesisShardCoordinator.SelfId())
	if err != nil {
		return err
	}

	mesn := notifier.NewManualEpochStartNotifier()
	mesn.NewEpoch(sesb.importDbConfig.ImportDBStartInEpoch + 1)
	sesb.store, err = sesb.createStoreForStorageResolvers(shardCoordinator, mesn)
	if err != nil {
		return err
	}

	resolversContainerFactoryArgs := storageResolversContainers.FactoryArgs{
		GeneralConfig:            sesb.generalConfig,
		ShardIDForTries:          sesb.importDbConfig.ImportDBTargetShardID,
		ChainID:                  sesb.chainID,
		WorkingDirectory:         sesb.importDbConfig.ImportDBWorkingDir,
		Hasher:                   sesb.coreComponentsHolder.Hasher(),
		ShardCoordinator:         shardCoordinator,
		Messenger:                sesb.messenger,
		Store:                    sesb.store,
		Marshalizer:              sesb.coreComponentsHolder.InternalMarshalizer(),
		Uint64ByteSliceConverter: sesb.coreComponentsHolder.Uint64ByteSliceConverter(),
		DataPacker:               dataPacker,
		ManualEpochStartNotifier: mesn,
		ChanGracefullyClose:      sesb.chanGracefullyClose,
	}

	var resolversContainerFactory dataRetriever.ResolversContainerFactory
	if sesb.importDbConfig.ImportDBTargetShardID == core.MetachainShardId {
		resolversContainerFactory, err = storageResolversContainers.NewMetaResolversContainerFactory(resolversContainerFactoryArgs)
	} else {
		resolversContainerFactory, err = storageResolversContainers.NewShardResolversContainerFactory(resolversContainerFactoryArgs)
	}

	if err != nil {
		return err
	}

	sesb.resolvers, err = resolversContainerFactory.Create()

	return err
}

func (sesb *storageEpochStartBootstrap) createStoreForStorageResolvers(shardCoordinator sharding.Coordinator, mesn epochStart.ManualEpochStartNotifier) (dataRetriever.StorageService, error) {
	pathManager, err := storageFactory.CreatePathManager(
		storageFactory.ArgCreatePathManager{
			WorkingDir: sesb.importDbConfig.ImportDBWorkingDir,
			ChainID:    sesb.chainID,
		},
	)
	if err != nil {
		return nil, err
	}

	return sesb.createStorageService(
		shardCoordinator,
		pathManager,
		mesn,
		sesb.importDbConfig.ImportDBStartInEpoch,
		sesb.importDbConfig.ImportDbSaveTrieEpochRootHash,
		sesb.importDbConfig.ImportDBTargetShardID,
	)
}

func (sesb *storageEpochStartBootstrap) requestAndProcessFromStorage() (Parameters, error) {
	var err error
	sesb.baseData.numberOfShards = uint32(len(sesb.epochStartMeta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers()))
	sesb.baseData.lastEpoch = sesb.epochStartMeta.GetEpoch()

	sesb.syncedHeaders, err = sesb.syncHeadersFromStorage(sesb.epochStartMeta, sesb.destinationShardAsObserver)
	if err != nil {
		return Parameters{}, err
	}
	log.Debug("start in epoch bootstrap: got shard header and previous epoch start meta block")

	prevEpochStartMetaHash := sesb.epochStartMeta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash()
	prevEpochStartMeta, ok := sesb.syncedHeaders[string(prevEpochStartMetaHash)].(*block.MetaBlock)
	if !ok {
		return Parameters{}, epochStart.ErrWrongTypeAssertion
	}
	sesb.prevEpochStartMeta = prevEpochStartMeta

	pubKeyBytes, err := sesb.cryptoComponentsHolder.PublicKey().ToByteArray()
	if err != nil {
		return Parameters{}, err
	}

	err = sesb.processNodesConfig(pubKeyBytes)
	if err != nil {
		return Parameters{}, err
	}
	log.Debug("start in epoch bootstrap: processNodesConfig")

	sesb.saveSelfShardId()
	sesb.shardCoordinator, err = sharding.NewMultiShardCoordinator(sesb.baseData.numberOfShards, sesb.baseData.shardId)
	if err != nil {
		return Parameters{}, fmt.Errorf("%w numberOfShards=%v shardId=%v", err, sesb.baseData.numberOfShards, sesb.baseData.shardId)
	}
	log.Debug("start in epoch bootstrap: shardCoordinator", "numOfShards", sesb.baseData.numberOfShards, "shardId", sesb.baseData.shardId)

	err = sesb.messenger.CreateTopic(common.ConsensusTopic+sesb.shardCoordinator.CommunicationIdentifier(sesb.shardCoordinator.SelfId()), true)
	if err != nil {
		return Parameters{}, err
	}

	emptyPeerMiniBlocksSlice := make([]*block.MiniBlock, 0) // empty slice since we have bootstrapped from storage
	if sesb.shardCoordinator.SelfId() == core.MetachainShardId {
		err = sesb.requestAndProcessForMeta(emptyPeerMiniBlocksSlice)
		if err != nil {
			return Parameters{}, err
		}
	} else {
		err = sesb.requestAndProcessForShard(emptyPeerMiniBlocksSlice)
		if err != nil {
			return Parameters{}, err
		}
	}

	parameters := Parameters{
		Epoch:       sesb.baseData.lastEpoch,
		SelfShardId: sesb.baseData.shardId,
		NumOfShards: sesb.baseData.numberOfShards,
		NodesConfig: sesb.nodesConfig,
	}

	return parameters, nil
}

func (sesb *storageEpochStartBootstrap) syncHeadersFromStorage(meta data.MetaHeaderHandler, syncingShardID uint32) (map[string]data.HeaderHandler, error) {
	hashesToRequest := make([][]byte, 0, len(meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers())+1)
	shardIds := make([]uint32, 0, len(meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers())+1)

	for _, epochStartData := range meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers() {
		shouldSkipHeaderFetch := epochStartData.GetShardID() != syncingShardID &&
			sesb.importDbConfig.ImportDBTargetShardID != core.MetachainShardId
		if shouldSkipHeaderFetch {
			continue
		}

		hashesToRequest = append(hashesToRequest, epochStartData.GetHeaderHash())
		shardIds = append(shardIds, epochStartData.GetShardID())
	}

	if meta.GetEpoch() > sesb.startEpoch+1 { // no need to request genesis block
		hashesToRequest = append(hashesToRequest, meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())
		shardIds = append(shardIds, core.MetachainShardId)
	}

	ctx, cancel := context.WithTimeout(context.Background(), sesb.timeToWaitForRequestedData)
	err := sesb.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	syncedHeaders, err := sesb.headersSyncer.GetHeaders()
	if err != nil {
		return nil, err
	}

	if meta.GetEpoch() == sesb.startEpoch+1 {
		syncedHeaders[string(meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())] = &block.MetaBlock{}
	}

	return syncedHeaders, nil
}

func (sesb *storageEpochStartBootstrap) processNodesConfig(pubKey []byte) error {
	var err error
	shardId := sesb.destinationShardAsObserver
	if shardId > sesb.baseData.numberOfShards && shardId != core.MetachainShardId {
		shardId = sesb.genesisShardCoordinator.SelfId()
	}
	argsNewValidatorStatusSyncers := ArgsNewSyncValidatorStatus{
		DataPool:           sesb.dataPool,
		Marshalizer:        sesb.coreComponentsHolder.InternalMarshalizer(),
		RequestHandler:     sesb.requestHandler,
		ChanceComputer:     sesb.rater,
		GenesisNodesConfig: sesb.genesisNodesConfig,
		NodeShuffler:       sesb.nodeShuffler,
		Hasher:             sesb.coreComponentsHolder.Hasher(),
		PubKey:             pubKey,
		ShardIdAsObserver:  shardId,
		ChanNodeStop:       sesb.coreComponentsHolder.ChanStopNodeProcess(),
		NodeTypeProvider:   sesb.coreComponentsHolder.NodeTypeProvider(),
		IsFullArchive:      sesb.prefsConfig.FullArchive,
	}
	sesb.nodesConfigHandler, err = NewSyncValidatorStatus(argsNewValidatorStatusSyncers)
	if err != nil {
		return err
	}

	clonedHeader := sesb.epochStartMeta.ShallowClone()
	clonedEpochStartMeta, ok := clonedHeader.(*block.MetaBlock)
	if !ok {
		return fmt.Errorf("%w while trying to assert clonedHeader to *block.MetaBlock", epochStart.ErrWrongTypeAssertion)
	}
	err = sesb.applyCurrentShardIDOnMiniblocksCopy(clonedEpochStartMeta)
	if err != nil {
		return err
	}

	clonedHeader = sesb.prevEpochStartMeta.ShallowClone()
	clonedPrevEpochStartMeta, ok := clonedHeader.(*block.MetaBlock)
	if !ok {
		return fmt.Errorf("%w while trying to assert prevClonedHeader to *block.MetaBlock", epochStart.ErrWrongTypeAssertion)
	}

	err = sesb.applyCurrentShardIDOnMiniblocksCopy(clonedPrevEpochStartMeta)
	if err != nil {
		return err
	}

	// no need to save the peers miniblocks here as they were already fetched from the DB
	sesb.nodesConfig, sesb.baseData.shardId, _, err = sesb.nodesConfigHandler.NodesConfigFromMetaBlock(clonedEpochStartMeta, clonedPrevEpochStartMeta)
	sesb.baseData.shardId = sesb.applyShardIDAsObserverIfNeeded(sesb.baseData.shardId)

	return err
}

// applyCurrentShardIDOnMiniblocksCopy will alter the fetched metablocks making the sender shard ID for each miniblock
// header to  be exactly the shard ID used in the import-db process. This is necessary as to allow the miniblocks to be requested
// on the available resolver and should be called only from this storage-base bootstrap instance.
// This method also copies the MiniBlockHeaders slice pointer. Otherwise the node will end up stating
// "start of epoch metablock mismatch"
func (sesb *storageEpochStartBootstrap) applyCurrentShardIDOnMiniblocksCopy(metablock data.HeaderHandler) error {
	originalMiniblocksHeaders := metablock.GetMiniBlockHeaderHandlers()
	mbsHeaderHandlersToSet := make([]data.MiniBlockHeaderHandler, 0, len(originalMiniblocksHeaders))
	var err error

	for i := range originalMiniblocksHeaders {
		mb := originalMiniblocksHeaders[i].ShallowClone()
		err = mb.SetSenderShardID(sesb.importDbConfig.ImportDBTargetShardID) // it is safe to modify here as mb is a shallow clone
		if err != nil {
			return err
		}

		mbsHeaderHandlersToSet = append(mbsHeaderHandlersToSet, mb)
	}

	err = metablock.SetMiniBlockHeaderHandlers(mbsHeaderHandlersToSet)
	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (sesb *storageEpochStartBootstrap) IsInterfaceNil() bool {
	return sesb == nil
}
