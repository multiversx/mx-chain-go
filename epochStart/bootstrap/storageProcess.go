package bootstrap

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	factoryDataPool "github.com/multiversx/mx-chain-go/dataRetriever/factory"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers"
	storagerequesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/storageRequestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage/cache"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	trieFactory "github.com/multiversx/mx-chain-go/trie/factory"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/partitioning"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
)

// ArgsStorageEpochStartBootstrap holds the arguments needed for creating an epoch start data provider component
// from storage
type ArgsStorageEpochStartBootstrap struct {
	ArgsEpochStartBootstrap
	ImportDbConfig             config.ImportDbConfig
	ChanGracefullyClose        chan endProcess.ArgEndProcess
	TimeToWaitForRequestedData time.Duration
	EpochStartBootStrap        *epochStartBootstrap
}

type storageEpochStartBootstrap struct {
	*epochStartBootstrap
	container                  dataRetriever.RequestersContainer
	store                      dataRetriever.StorageService
	importDbConfig             config.ImportDbConfig
	chanGracefullyClose        chan endProcess.ArgEndProcess
	chainID                    string
	timeToWaitForRequestedData time.Duration
}

// NewStorageEpochStartBootstrap will return a new instance of storageEpochStartBootstrap that can bootstrap
// the node with the help of storage requesters through the import-db process
func NewStorageEpochStartBootstrap(args ArgsStorageEpochStartBootstrap) (*storageEpochStartBootstrap, error) {
	err := checkArguments(args.ArgsEpochStartBootstrap)
	if err != nil {
		return nil, err
	}
	if check.IfNil(args.EpochStartBootStrap) {
		return nil, errors.ErrNilEpochStartBootstrapper
	}
	if args.ChanGracefullyClose == nil {
		return nil, dataRetriever.ErrNilGracefullyCloseChannel
	}

	sesb := &storageEpochStartBootstrap{
		epochStartBootstrap:        args.EpochStartBootStrap,
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
	defer sesb.closeBootstrapHeartbeatSender()

	if !sesb.generalConfig.GeneralSettings.StartInEpochEnabled {
		return sesb.bootstrapFromLocalStorage()
	}

	defer func() {
		sesb.cleanupOnBootstrapFinish()

		if !check.IfNil(sesb.container) {
			err := sesb.container.Close()
			if err != nil {
				log.Debug("non critical error closing requesters", "error", err)
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
	sesb.shardCoordinator, err = sesb.runTypeComponents.ShardCoordinatorCreator().CreateShardCoordinator(sesb.genesisShardCoordinator.NumberOfShards(), core.MetachainShardId)
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
	triesContainer, trieStorageManagers, err := trieFactory.CreateTriesComponentsForShardId(
		sesb.generalConfig,
		sesb.coreComponentsHolder,
		sesb.storageService,
		sesb.stateStatsHandler,
	)
	if err != nil {
		return err
	}

	sesb.trieContainer = triesContainer
	sesb.trieStorageManagers = trieStorageManagers

	requestHandler, err := sesb.createStorageRequestHandler()
	if err != nil {
		return err
	}

	sesb.requestHandler = requestHandler

	metablockProcessor, err := NewStorageEpochStartMetaBlockProcessor(
		sesb.mainMessenger,
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
		Messenger:               sesb.mainMessenger,
		ShardCoordinator:        sesb.shardCoordinator,
		EconomicsData:           sesb.economicsData,
		WhitelistHandler:        sesb.whiteListHandler,
		StartInEpochConfig:      sesb.generalConfig.EpochStartConfig,
		HeaderIntegrityVerifier: sesb.headerIntegrityVerifier,
		MetaBlockProcessor:      metablockProcessor,
	}

	sesb.epochStartMetaBlockSyncer, err = sesb.bootStrapShardProcessor.createStorageEpochStartMetaSyncer(argsEpochStartSyncer)
	if err != nil {
		return err
	}

	return nil
}

func (sesb *storageEpochStartBootstrap) createStorageRequestHandler() (process.RequestHandler, error) {
	err := sesb.createStorageRequesters()
	if err != nil {
		return nil, err
	}

	finder, err := containers.NewRequestersFinder(sesb.container, sesb.shardCoordinator)
	if err != nil {
		return nil, err
	}

	requestedItemsHandler := cache.NewTimeCache(timeBetweenRequests)
	args := requestHandlers.RequestHandlerArgs{
		RequestersFinder:      finder,
		RequestedItemsHandler: requestedItemsHandler,
		WhiteListHandler:      sesb.whiteListHandler,
		MaxTxsToRequest:       maxToRequest,
		ShardID:               core.MetachainShardId,
		RequestInterval:       timeBetweenRequests,
	}

	return sesb.runTypeComponents.RequestHandlerCreator().CreateRequestHandler(args)
}

func (sesb *storageEpochStartBootstrap) createStorageRequesters() error {
	dataPacker, err := partitioning.NewSimpleDataPacker(sesb.coreComponentsHolder.InternalMarshalizer())
	if err != nil {
		return err
	}

	shardCoordinator, err := sesb.runTypeComponents.ShardCoordinatorCreator().CreateShardCoordinator(sesb.genesisShardCoordinator.NumberOfShards(), sesb.genesisShardCoordinator.SelfId())
	if err != nil {
		return err
	}

	initialEpoch := uint32(1)
	mesn := notifier.NewManualEpochStartNotifier()
	mesn.NewEpoch(initialEpoch)
	sesb.store, err = sesb.createStoreForStorageResolvers(shardCoordinator, mesn)
	if err != nil {
		return err
	}

	requestersContainerFactoryArgs := storagerequesterscontainer.FactoryArgs{
		GeneralConfig:            sesb.generalConfig,
		ShardIDForTries:          sesb.importDbConfig.ImportDBTargetShardID,
		ChainID:                  sesb.chainID,
		WorkingDirectory:         sesb.importDbConfig.ImportDBWorkingDir,
		Hasher:                   sesb.coreComponentsHolder.Hasher(),
		ShardCoordinator:         shardCoordinator,
		Messenger:                sesb.mainMessenger,
		Store:                    sesb.store,
		Marshalizer:              sesb.coreComponentsHolder.InternalMarshalizer(),
		Uint64ByteSliceConverter: sesb.coreComponentsHolder.Uint64ByteSliceConverter(),
		DataPacker:               dataPacker,
		ManualEpochStartNotifier: mesn,
		ChanGracefullyClose:      sesb.chanGracefullyClose,
		EnableEpochsHandler:      sesb.coreComponentsHolder.EnableEpochsHandler(),
		StateStatsHandler:        sesb.stateStatsHandler,
	}

	var requestersContainerFactory dataRetriever.RequestersContainerFactory
	if sesb.importDbConfig.ImportDBTargetShardID == core.MetachainShardId {
		requestersContainerFactory, err = storagerequesterscontainer.NewMetaRequestersContainerFactory(requestersContainerFactoryArgs)
	} else {
		requestersContainerFactory, err = sesb.runTypeComponents.ShardRequestersContainerCreatorHandler().CreateShardRequestersContainerFactory(requestersContainerFactoryArgs)
	}

	if err != nil {
		return err
	}

	sesb.container, err = requestersContainerFactory.Create()

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

	return sesb.createStorageServiceForImportDB(
		shardCoordinator,
		pathManager,
		mesn,
		sesb.importDbConfig.ImportDbSaveTrieEpochRootHash,
		sesb.importDbConfig.ImportDBTargetShardID,
	)
}

func (sesb *storageEpochStartBootstrap) requestAndProcessFromStorage() (Parameters, error) {
	var err error
	sesb.baseData.numberOfShards = sesb.bootStrapShardProcessor.computeNumShards(sesb.epochStartMeta)
	sesb.baseData.lastEpoch = sesb.epochStartMeta.GetEpoch()

	sesb.syncedHeaders, err = sesb.bootStrapShardProcessor.syncHeadersFromStorage(
		sesb.epochStartMeta,
		sesb.destinationShardAsObserver,
		sesb.importDbConfig.ImportDBTargetShardID,
		sesb.timeToWaitForRequestedData,
	)
	if err != nil {
		return Parameters{}, err
	}
	log.Debug("start in epoch bootstrap: got shard header and previous epoch start meta block")

	prevEpochStartMetaHash := sesb.epochStartMeta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash()
	prevEpochStartMeta, ok := sesb.syncedHeaders[string(prevEpochStartMetaHash)].(data.MetaHeaderHandler)
	if !ok {
		return Parameters{}, epochStart.ErrWrongTypeAssertion
	}
	sesb.prevEpochStartMeta = prevEpochStartMeta

	pubKeyBytes, err := sesb.cryptoComponentsHolder.PublicKey().ToByteArray()
	if err != nil {
		return Parameters{}, err
	}

	sesb.nodesConfig, sesb.baseData.shardId, err = sesb.bootStrapShardProcessor.processNodesConfigFromStorage(pubKeyBytes, sesb.importDbConfig.ImportDBTargetShardID)
	if err != nil {
		return Parameters{}, err
	}
	log.Debug("start in epoch bootstrap: processNodesConfig")

	sesb.saveSelfShardId()
	sesb.shardCoordinator, err = sesb.runTypeComponents.ShardCoordinatorCreator().CreateShardCoordinator(sesb.baseData.numberOfShards, sesb.baseData.shardId)
	if err != nil {
		return Parameters{}, fmt.Errorf("%w numberOfShards=%v shardId=%v", err, sesb.baseData.numberOfShards, sesb.baseData.shardId)
	}
	log.Debug("start in epoch bootstrap: shardCoordinator", "numOfShards", sesb.baseData.numberOfShards, "shardId", sesb.baseData.shardId)

	consensusTopic := common.ConsensusTopic + sesb.shardCoordinator.CommunicationIdentifier(sesb.shardCoordinator.SelfId())
	err = sesb.mainMessenger.CreateTopic(consensusTopic, true)
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
		err = sesb.bootStrapShardProcessor.requestAndProcessForShard(emptyPeerMiniBlocksSlice)
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

// IsInterfaceNil returns true if there is no value under the interface
func (sesb *storageEpochStartBootstrap) IsInterfaceNil() bool {
	return sesb == nil
}
