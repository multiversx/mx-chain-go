package bootstrap

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/partitioning"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-go/common"
	disabledCommon "github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/common/ordering"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	factoryDataPool "github.com/multiversx/mx-chain-go/dataRetriever/factory"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	factoryInterceptors "github.com/multiversx/mx-chain-go/epochStart/bootstrap/factory"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/types"
	factoryDisabled "github.com/multiversx/mx-chain-go/factory/disabled"
	"github.com/multiversx/mx-chain-go/heartbeat/sender"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/heartbeat/validator"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	disabledInterceptors "github.com/multiversx/mx-chain-go/process/interceptors/disabled"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/redundancy"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/trie/factory"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/multiversx/mx-chain-go/trie/storageMarker"
	"github.com/multiversx/mx-chain-go/update"
	updateSync "github.com/multiversx/mx-chain-go/update/sync"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("epochStart/bootstrap")

// DefaultTimeToWaitForRequestedData represents the default timespan until requested data needs to be received from the connected peers
const DefaultTimeToWaitForRequestedData = time.Minute
const timeBetweenRequests = 100 * time.Millisecond
const maxToRequest = 100
const gracePeriodInPercentage = float64(0.25)
const roundGracePeriod = 25

// thresholdForConsideringMetaBlockCorrect represents the percentage (between 0 and 100) of connected peers to send
// the same meta block in order to consider it correct
const thresholdForConsideringMetaBlockCorrect = 67

// Parameters defines the DTO for the result produced by the bootstrap component
type Parameters struct {
	Epoch       uint32
	SelfShardId uint32
	NumOfShards uint32
	NodesConfig *nodesCoordinator.NodesCoordinatorRegistry
}

// ComponentsNeededForBootstrap holds the components which need to be initialized from network
type ComponentsNeededForBootstrap struct {
	EpochStartMetaBlock data.MetaHeaderHandler
	PreviousEpochStart  data.MetaHeaderHandler
	ShardHeader         data.HeaderHandler
	NodesConfig         *nodesCoordinator.NodesCoordinatorRegistry
	Headers             map[string]data.HeaderHandler
	ShardCoordinator    sharding.Coordinator
	PendingMiniBlocks   map[string]*block.MiniBlock
	PeerMiniBlocks      []*block.MiniBlock
}

// epochStartBootstrap will handle requesting the needed data to start when joining late the network
type epochStartBootstrap struct {
	// should come via arguments
	destinationShardAsObserver uint32
	coreComponentsHolder       process.CoreComponentsHolder
	cryptoComponentsHolder     process.CryptoComponentsHolder
	mainMessenger              p2p.Messenger
	fullArchiveMessenger       p2p.Messenger
	generalConfig              config.Config
	prefsConfig                config.PreferencesConfig
	flagsConfig                config.ContextFlagsConfig
	economicsData              process.EconomicsDataHandler
	shardCoordinator           sharding.Coordinator
	genesisNodesConfig         sharding.GenesisNodesSetupHandler
	genesisShardCoordinator    sharding.Coordinator
	rater                      nodesCoordinator.ChanceComputer
	storerScheduledSCRs        storage.Storer
	trieContainer              common.TriesHolder
	trieStorageManagers        map[string]common.StorageManager
	mutTrieStorageManagers     sync.RWMutex
	nodeShuffler               nodesCoordinator.NodesShuffler
	roundHandler               epochStart.RoundHandler
	statusHandler              core.AppStatusHandler
	headerIntegrityVerifier    process.HeaderIntegrityVerifier
	numConcurrentTrieSyncers   int
	maxHardCapForMissingNodes  int
	trieSyncerVersion          int
	checkNodesOnDisk           bool
	bootstrapHeartbeatSender   update.Closer
	trieSyncStatisticsProvider common.SizeSyncStatisticsHandler
	nodeProcessingMode         common.NodeProcessingMode
	nodeOperationMode          common.NodeOperation
	// created components
	requestHandler                  process.RequestHandler
	mainInterceptorContainer        process.InterceptorsContainer
	fullArchiveInterceptorContainer process.InterceptorsContainer
	dataPool                        dataRetriever.PoolsHolder
	miniBlocksSyncer                epochStart.PendingMiniBlocksSyncHandler
	headersSyncer                   epochStart.HeadersByHashSyncer
	txSyncerForScheduled            update.TransactionsSyncHandler
	epochStartMetaBlockSyncer       epochStart.StartOfEpochMetaSyncer
	nodesConfigHandler              StartOfEpochNodesConfigHandler
	whiteListHandler                update.WhiteListHandler
	whiteListerVerifiedTxs          update.WhiteListHandler
	storageOpenerHandler            storage.UnitOpenerHandler
	latestStorageDataProvider       storage.LatestStorageDataProviderHandler
	argumentsParser                 process.ArgumentsParser
	dataSyncerFactory               types.ScheduledDataSyncerCreator
	dataSyncerWithScheduled         types.ScheduledDataSyncer
	storageService                  dataRetriever.StorageService

	// gathered data
	epochStartMeta     data.MetaHeaderHandler
	prevEpochStartMeta data.MetaHeaderHandler
	syncedHeaders      map[string]data.HeaderHandler
	nodesConfig        *nodesCoordinator.NodesCoordinatorRegistry
	baseData           baseDataInStorage
	startRound         int64
	nodeType           core.NodeType
	startEpoch         uint32
	shuffledOut        bool
}

type baseDataInStorage struct {
	shardId         uint32
	numberOfShards  uint32
	lastRound       int64
	epochStartRound uint64
	lastEpoch       uint32
	storageExists   bool
}

// ArgsEpochStartBootstrap holds the arguments needed for creating an epoch start data provider component
type ArgsEpochStartBootstrap struct {
	CoreComponentsHolder       process.CoreComponentsHolder
	CryptoComponentsHolder     process.CryptoComponentsHolder
	DestinationShardAsObserver uint32
	MainMessenger              p2p.Messenger
	FullArchiveMessenger       p2p.Messenger
	GeneralConfig              config.Config
	PrefsConfig                config.PreferencesConfig
	FlagsConfig                config.ContextFlagsConfig
	EconomicsData              process.EconomicsDataHandler
	GenesisNodesConfig         sharding.GenesisNodesSetupHandler
	GenesisShardCoordinator    sharding.Coordinator
	StorageUnitOpener          storage.UnitOpenerHandler
	LatestStorageDataProvider  storage.LatestStorageDataProviderHandler
	Rater                      nodesCoordinator.ChanceComputer
	NodeShuffler               nodesCoordinator.NodesShuffler
	RoundHandler               epochStart.RoundHandler
	ArgumentsParser            process.ArgumentsParser
	StatusHandler              core.AppStatusHandler
	HeaderIntegrityVerifier    process.HeaderIntegrityVerifier
	DataSyncerCreator          types.ScheduledDataSyncerCreator
	ScheduledSCRsStorer        storage.Storer
	TrieSyncStatisticsProvider common.SizeSyncStatisticsHandler
	NodeProcessingMode         common.NodeProcessingMode
}

type dataToSync struct {
	ownShardHdr       data.ShardHeaderHandler
	rootHashToSync    []byte
	withScheduled     bool
	additionalHeaders map[string]data.HeaderHandler
	miniBlocks        map[string]*block.MiniBlock
}

// NewEpochStartBootstrap will return a new instance of epochStartBootstrap
func NewEpochStartBootstrap(args ArgsEpochStartBootstrap) (*epochStartBootstrap, error) {
	err := checkArguments(args)
	if err != nil {
		return nil, err
	}

	epochStartProvider := &epochStartBootstrap{
		coreComponentsHolder:       args.CoreComponentsHolder,
		cryptoComponentsHolder:     args.CryptoComponentsHolder,
		mainMessenger:              args.MainMessenger,
		fullArchiveMessenger:       args.FullArchiveMessenger,
		generalConfig:              args.GeneralConfig,
		prefsConfig:                args.PrefsConfig,
		flagsConfig:                args.FlagsConfig,
		economicsData:              args.EconomicsData,
		genesisNodesConfig:         args.GenesisNodesConfig,
		genesisShardCoordinator:    args.GenesisShardCoordinator,
		rater:                      args.Rater,
		destinationShardAsObserver: args.DestinationShardAsObserver,
		nodeShuffler:               args.NodeShuffler,
		roundHandler:               args.RoundHandler,
		storageOpenerHandler:       args.StorageUnitOpener,
		latestStorageDataProvider:  args.LatestStorageDataProvider,
		shuffledOut:                false,
		statusHandler:              args.StatusHandler,
		nodeType:                   core.NodeTypeObserver,
		argumentsParser:            args.ArgumentsParser,
		headerIntegrityVerifier:    args.HeaderIntegrityVerifier,
		numConcurrentTrieSyncers:   args.GeneralConfig.TrieSync.NumConcurrentTrieSyncers,
		maxHardCapForMissingNodes:  args.GeneralConfig.TrieSync.MaxHardCapForMissingNodes,
		trieSyncerVersion:          args.GeneralConfig.TrieSync.TrieSyncerVersion,
		checkNodesOnDisk:           args.GeneralConfig.TrieSync.CheckNodesOnDisk,
		dataSyncerFactory:          args.DataSyncerCreator,
		storerScheduledSCRs:        args.ScheduledSCRsStorer,
		shardCoordinator:           args.GenesisShardCoordinator,
		trieSyncStatisticsProvider: args.TrieSyncStatisticsProvider,
		nodeProcessingMode:         args.NodeProcessingMode,
		nodeOperationMode:          common.NormalOperation,
	}

	if epochStartProvider.prefsConfig.FullArchive {
		epochStartProvider.nodeOperationMode = common.FullArchiveMode
	}

	whiteListCache, err := storageunit.NewCache(storageFactory.GetCacherFromConfig(epochStartProvider.generalConfig.WhiteListPool))
	if err != nil {
		return nil, err
	}

	epochStartProvider.whiteListHandler, err = interceptors.NewWhiteListDataVerifier(whiteListCache)
	if err != nil {
		return nil, err
	}

	epochStartProvider.whiteListerVerifiedTxs, err = disabledInterceptors.NewDisabledWhiteListDataVerifier()
	if err != nil {
		return nil, err
	}

	epochStartProvider.trieContainer = state.NewDataTriesHolder()
	epochStartProvider.trieStorageManagers = make(map[string]common.StorageManager)

	if epochStartProvider.generalConfig.Hardfork.AfterHardFork {
		epochStartProvider.startEpoch = epochStartProvider.generalConfig.Hardfork.StartEpoch
		epochStartProvider.baseData.lastEpoch = epochStartProvider.startEpoch
		epochStartProvider.startRound = int64(epochStartProvider.generalConfig.Hardfork.StartRound)
		epochStartProvider.baseData.lastRound = epochStartProvider.startRound
		epochStartProvider.baseData.epochStartRound = uint64(epochStartProvider.startRound)
	}

	return epochStartProvider, nil
}

func (e *epochStartBootstrap) isStartInEpochZero() bool {
	startTime := time.Unix(e.genesisNodesConfig.GetStartTime(), 0)
	isCurrentTimeBeforeGenesis := time.Since(startTime) < 0
	if isCurrentTimeBeforeGenesis {
		return true
	}

	currentRound := e.roundHandler.Index() - e.startRound
	epochEndPlusGracePeriod := float64(e.generalConfig.EpochStartConfig.RoundsPerEpoch) * (gracePeriodInPercentage + 1.0)
	log.Debug("IsStartInEpochZero", "currentRound", currentRound, "epochEndRound", epochEndPlusGracePeriod)
	return float64(currentRound) < epochEndPlusGracePeriod
}

func (e *epochStartBootstrap) prepareEpochZero() (Parameters, error) {
	shardIDToReturn := e.genesisShardCoordinator.SelfId()
	if !e.isNodeInGenesisNodesConfig() {
		shardIDToReturn = e.applyShardIDAsObserverIfNeeded(e.genesisShardCoordinator.SelfId())
	}
	parameters := Parameters{
		Epoch:       e.startEpoch,
		SelfShardId: shardIDToReturn,
		NumOfShards: e.genesisShardCoordinator.NumberOfShards(),
	}
	return parameters, nil
}

func (e *epochStartBootstrap) isNodeInGenesisNodesConfig() bool {
	ownPubKey, err := e.cryptoComponentsHolder.PublicKey().ToByteArray()
	if err != nil {
		return false
	}

	eligibleList, waitingList := e.genesisNodesConfig.InitialNodesInfo()

	for _, nodesInShard := range eligibleList {
		for _, eligibleNode := range nodesInShard {
			if bytes.Equal(eligibleNode.PubKeyBytes(), ownPubKey) {
				return true
			}
		}
	}

	for _, nodesInShard := range waitingList {
		for _, waitingNode := range nodesInShard {
			if bytes.Equal(waitingNode.PubKeyBytes(), ownPubKey) {
				return true
			}
		}
	}

	return false
}

// Bootstrap runs the fast bootstrap method from the network or local storage
func (e *epochStartBootstrap) Bootstrap() (Parameters, error) {
	defer e.closeTrieComponents()
	defer e.closeBootstrapHeartbeatSender()

	if e.flagsConfig.ForceStartFromNetwork {
		log.Warn("epochStartBootstrap.Bootstrap: forcing start from network")
	}

	shouldStartFromNetwork := e.generalConfig.GeneralSettings.StartInEpochEnabled || e.flagsConfig.ForceStartFromNetwork
	if !shouldStartFromNetwork {
		return e.bootstrapFromLocalStorage()
	}

	defer e.cleanupOnBootstrapFinish()

	var err error
	e.shardCoordinator, err = sharding.NewMultiShardCoordinator(e.genesisShardCoordinator.NumberOfShards(), core.MetachainShardId)
	if err != nil {
		return Parameters{}, err
	}

	e.dataPool, err = factoryDataPool.NewDataPoolFromConfig(
		factoryDataPool.ArgsDataPool{
			Config:           &e.generalConfig,
			EconomicsData:    e.economicsData,
			ShardCoordinator: e.shardCoordinator,
			Marshalizer:      e.coreComponentsHolder.InternalMarshalizer(),
			PathManager:      e.coreComponentsHolder.PathHandler(),
		},
	)
	if err != nil {
		return Parameters{}, err
	}

	params, shouldContinue, err := e.startFromSavedEpoch()
	shouldContinue = shouldContinue || e.flagsConfig.ForceStartFromNetwork
	if !shouldContinue {
		return params, err
	}

	err = e.prepareComponentsToSyncFromNetwork()
	if err != nil {
		return Parameters{}, err
	}

	e.epochStartMeta, err = e.epochStartMetaBlockSyncer.SyncEpochStartMeta(DefaultTimeToWaitForRequestedData)
	if err != nil {
		return Parameters{}, err
	}
	log.Debug("start in epoch bootstrap: got epoch start meta header", "epoch", e.epochStartMeta.GetEpoch(), "nonce", e.epochStartMeta.GetNonce())
	e.setEpochStartMetrics()

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

	params, err = e.requestAndProcessing()
	if err != nil {
		return Parameters{}, err
	}

	return params, nil
}

func (e *epochStartBootstrap) bootstrapFromLocalStorage() (Parameters, error) {
	log.Warn("fast bootstrap is disabled")

	e.initializeFromLocalStorage()
	if !e.baseData.storageExists {
		return Parameters{
			Epoch:       e.startEpoch,
			SelfShardId: e.genesisShardCoordinator.SelfId(),
			NumOfShards: e.genesisShardCoordinator.NumberOfShards(),
		}, nil
	}

	newShardId, shuffledOut, err := e.getShardIDForLatestEpoch()
	if err != nil {
		return Parameters{}, err
	}

	epochToStart := e.baseData.lastEpoch
	if shuffledOut {
		epochToStart = e.startEpoch
	}

	newShardId = e.applyShardIDAsObserverIfNeeded(newShardId)
	return Parameters{
		Epoch:       epochToStart,
		SelfShardId: newShardId,
		NumOfShards: e.baseData.numberOfShards,
		NodesConfig: e.nodesConfig,
	}, nil
}

func (e *epochStartBootstrap) cleanupOnBootstrapFinish() {
	log.Debug("unregistering all message processor and un-joining all topics")
	errMessenger := e.mainMessenger.UnregisterAllMessageProcessors()
	log.LogIfError(errMessenger)

	errMessenger = e.mainMessenger.UnJoinAllTopics()
	log.LogIfError(errMessenger)

	errMessenger = e.fullArchiveMessenger.UnregisterAllMessageProcessors()
	log.LogIfError(errMessenger)

	errMessenger = e.fullArchiveMessenger.UnJoinAllTopics()
	log.LogIfError(errMessenger)

	e.closeTrieNodes()
}

func (e *epochStartBootstrap) closeTrieNodes() {
	if check.IfNil(e.dataPool) {
		return
	}
	if check.IfNil(e.dataPool.TrieNodes()) {
		return
	}

	errTrieNodesClosed := e.dataPool.TrieNodes().Close()
	log.LogIfError(errTrieNodesClosed)
}

func (e *epochStartBootstrap) startFromSavedEpoch() (Parameters, bool, error) {
	isStartInEpochZero := e.isStartInEpochZero()
	isCurrentEpochSaved := e.computeIfCurrentEpochIsSaved()

	if isStartInEpochZero || isCurrentEpochSaved {
		if e.baseData.lastEpoch <= e.startEpoch {
			params, err := e.prepareEpochZero()
			return params, false, err
		}

		parameters, errPrepare := e.prepareEpochFromStorage()
		if errPrepare == nil {
			return parameters, false, nil
		}

		if e.shuffledOut {
			// sync was already tried - not need to continue from here
			return Parameters{}, false, errPrepare
		}

		log.Debug("could not start from storage - will try sync for start in epoch", "error", errPrepare)
	}

	return Parameters{}, true, nil
}

func (e *epochStartBootstrap) computeIfCurrentEpochIsSaved() bool {
	e.initializeFromLocalStorage()
	if !e.baseData.storageExists {
		return false
	}

	computedRound := e.roundHandler.Index()
	log.Debug("computed round", "round", computedRound, "lastRound", e.baseData.lastRound)
	if computedRound-e.baseData.lastRound < roundGracePeriod {
		return true
	}

	roundsSinceEpochStart := computedRound - int64(e.baseData.epochStartRound)
	log.Debug("epoch start round", "round", e.baseData.epochStartRound, "roundsSinceEpochStart", roundsSinceEpochStart)
	epochEndPlusGracePeriod := float64(e.generalConfig.EpochStartConfig.RoundsPerEpoch) * (gracePeriodInPercentage + 1.0)
	return float64(roundsSinceEpochStart) < epochEndPlusGracePeriod
}

func (e *epochStartBootstrap) prepareComponentsToSyncFromNetwork() error {
	e.closeTrieComponents()
	e.storageService = disabled.NewChainStorer()
	triesContainer, trieStorageManagers, err := factory.CreateTriesComponentsForShardId(
		e.generalConfig,
		e.coreComponentsHolder,
		e.storageService,
	)
	if err != nil {
		return err
	}

	e.trieContainer = triesContainer
	e.trieStorageManagers = trieStorageManagers

	err = e.createResolversContainer()
	if err != nil {
		return err
	}

	err = e.createRequestHandler()
	if err != nil {
		return err
	}

	epochStartConfig := e.generalConfig.EpochStartConfig
	metaBlockProcessor, err := NewEpochStartMetaBlockProcessor(
		e.mainMessenger,
		e.requestHandler,
		e.coreComponentsHolder.InternalMarshalizer(),
		e.coreComponentsHolder.Hasher(),
		thresholdForConsideringMetaBlockCorrect,
		epochStartConfig.MinNumConnectedPeersToStart,
		epochStartConfig.MinNumOfPeersToConsiderBlockValid,
	)
	if err != nil {
		return err
	}

	argsEpochStartSyncer := ArgsNewEpochStartMetaSyncer{
		CoreComponentsHolder:    e.coreComponentsHolder,
		CryptoComponentsHolder:  e.cryptoComponentsHolder,
		RequestHandler:          e.requestHandler,
		Messenger:               e.mainMessenger,
		ShardCoordinator:        e.shardCoordinator,
		EconomicsData:           e.economicsData,
		WhitelistHandler:        e.whiteListHandler,
		StartInEpochConfig:      epochStartConfig,
		HeaderIntegrityVerifier: e.headerIntegrityVerifier,
		MetaBlockProcessor:      metaBlockProcessor,
	}
	e.epochStartMetaBlockSyncer, err = NewEpochStartMetaSyncer(argsEpochStartSyncer)
	if err != nil {
		return err
	}

	return nil
}

func (e *epochStartBootstrap) createSyncers() error {
	var err error
	args := factoryInterceptors.ArgsEpochStartInterceptorContainer{
		CoreComponents:          e.coreComponentsHolder,
		CryptoComponents:        e.cryptoComponentsHolder,
		Config:                  e.generalConfig,
		ShardCoordinator:        e.shardCoordinator,
		MainMessenger:           e.mainMessenger,
		FullArchiveMessenger:    e.fullArchiveMessenger,
		DataPool:                e.dataPool,
		WhiteListHandler:        e.whiteListHandler,
		WhiteListerVerifiedTxs:  e.whiteListerVerifiedTxs,
		ArgumentsParser:         e.argumentsParser,
		HeaderIntegrityVerifier: e.headerIntegrityVerifier,
		RequestHandler:          e.requestHandler,
		SignaturesHandler:       e.mainMessenger,
		NodeOperationMode:       e.nodeOperationMode,
	}

	e.mainInterceptorContainer, e.fullArchiveInterceptorContainer, err = factoryInterceptors.NewEpochStartInterceptorsContainer(args)
	if err != nil {
		return err
	}

	syncMiniBlocksArgs := updateSync.ArgsNewPendingMiniBlocksSyncer{
		Storage:        disabled.CreateMemUnit(),
		Cache:          e.dataPool.MiniBlocks(),
		Marshalizer:    e.coreComponentsHolder.InternalMarshalizer(),
		RequestHandler: e.requestHandler,
	}
	e.miniBlocksSyncer, err = updateSync.NewPendingMiniBlocksSyncer(syncMiniBlocksArgs)
	if err != nil {
		return err
	}

	syncMissingHeadersArgs := updateSync.ArgsNewMissingHeadersByHashSyncer{
		Storage:        disabled.CreateMemUnit(),
		Cache:          e.dataPool.Headers(),
		Marshalizer:    e.coreComponentsHolder.InternalMarshalizer(),
		RequestHandler: e.requestHandler,
	}
	e.headersSyncer, err = updateSync.NewMissingheadersByHashSyncer(syncMissingHeadersArgs)
	if err != nil {
		return err
	}

	syncTxsArgs := updateSync.ArgsNewTransactionsSyncer{
		DataPools:      e.dataPool,
		Storages:       disabled.NewChainStorer(),
		Marshaller:     e.coreComponentsHolder.InternalMarshalizer(),
		RequestHandler: e.requestHandler,
	}

	e.txSyncerForScheduled, err = updateSync.NewTransactionsSyncer(syncTxsArgs)
	if err != nil {
		return err
	}

	return nil
}

func (e *epochStartBootstrap) syncHeadersFrom(meta data.MetaHeaderHandler) (map[string]data.HeaderHandler, error) {
	hashesToRequest := make([][]byte, 0, len(meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers())+1)
	shardIds := make([]uint32, 0, len(meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers())+1)

	for _, epochStartData := range meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers() {
		hashesToRequest = append(hashesToRequest, epochStartData.GetHeaderHash())
		shardIds = append(shardIds, epochStartData.GetShardID())
	}

	if meta.GetEpoch() > e.startEpoch+1 { // no need to request genesis block
		hashesToRequest = append(hashesToRequest, meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())
		shardIds = append(shardIds, core.MetachainShardId)
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err := e.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	syncedHeaders, err := e.headersSyncer.GetHeaders()
	if err != nil {
		return nil, err
	}

	if meta.GetEpoch() == e.startEpoch+1 {
		syncedHeaders[string(meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())] = &block.MetaBlock{}
	}

	return syncedHeaders, nil
}

// Bootstrap will handle requesting and receiving the needed information the node will bootstrap from
func (e *epochStartBootstrap) requestAndProcessing() (Parameters, error) {
	var err error
	e.baseData.numberOfShards = uint32(len(e.epochStartMeta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers()))
	e.baseData.lastEpoch = e.epochStartMeta.GetEpoch()

	e.syncedHeaders, err = e.syncHeadersFrom(e.epochStartMeta)
	if err != nil {
		return Parameters{}, err
	}
	log.Debug("start in epoch bootstrap: got shard headers and previous epoch start meta block")

	prevEpochStartMetaHash := e.epochStartMeta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash()
	prevEpochStartMeta, ok := e.syncedHeaders[string(prevEpochStartMetaHash)].(*block.MetaBlock)
	if !ok {
		return Parameters{}, epochStart.ErrWrongTypeAssertion
	}
	e.prevEpochStartMeta = prevEpochStartMeta

	pubKeyBytes, err := e.cryptoComponentsHolder.PublicKey().ToByteArray()
	if err != nil {
		return Parameters{}, err
	}

	miniBlocks, err := e.processNodesConfig(pubKeyBytes)
	if err != nil {
		return Parameters{}, err
	}
	log.Debug("start in epoch bootstrap: processNodesConfig")

	e.saveSelfShardId()
	e.shardCoordinator, err = sharding.NewMultiShardCoordinator(e.baseData.numberOfShards, e.baseData.shardId)
	if err != nil {
		return Parameters{}, fmt.Errorf("%w numberOfShards=%v shardId=%v", err, e.baseData.numberOfShards, e.baseData.shardId)
	}
	log.Debug("start in epoch bootstrap: shardCoordinator", "numOfShards", e.baseData.numberOfShards, "shardId", e.baseData.shardId)

	consensusTopic := common.ConsensusTopic + e.shardCoordinator.CommunicationIdentifier(e.shardCoordinator.SelfId())
	err = e.mainMessenger.CreateTopic(consensusTopic, true)
	if err != nil {
		return Parameters{}, err
	}

	err = e.createHeartbeatSender()
	if err != nil {
		return Parameters{}, err
	}

	if e.shardCoordinator.SelfId() == core.MetachainShardId {
		err = e.requestAndProcessForMeta(miniBlocks)
		if err != nil {
			return Parameters{}, err
		}
	} else {
		err = e.requestAndProcessForShard(miniBlocks)
		if err != nil {
			return Parameters{}, err
		}
	}

	log.Debug("removing cached received trie nodes")
	e.dataPool.TrieNodes().Clear()

	parameters := Parameters{
		Epoch:       e.baseData.lastEpoch,
		SelfShardId: e.baseData.shardId,
		NumOfShards: e.baseData.numberOfShards,
		NodesConfig: e.nodesConfig,
	}

	return parameters, nil
}

func (e *epochStartBootstrap) saveSelfShardId() {
	if e.baseData.shardId != core.AllShardId {
		return
	}

	e.baseData.shardId = e.destinationShardAsObserver

	if e.baseData.shardId > e.baseData.numberOfShards &&
		e.baseData.shardId != core.MetachainShardId {
		e.baseData.shardId = e.genesisShardCoordinator.SelfId()
	}
}

func (e *epochStartBootstrap) processNodesConfig(pubKey []byte) ([]*block.MiniBlock, error) {
	var err error
	shardId := e.destinationShardAsObserver
	if shardId > e.baseData.numberOfShards && shardId != core.MetachainShardId {
		shardId = e.genesisShardCoordinator.SelfId()
	}
	argsNewValidatorStatusSyncers := ArgsNewSyncValidatorStatus{
		DataPool:            e.dataPool,
		Marshalizer:         e.coreComponentsHolder.InternalMarshalizer(),
		RequestHandler:      e.requestHandler,
		ChanceComputer:      e.rater,
		GenesisNodesConfig:  e.genesisNodesConfig,
		NodeShuffler:        e.nodeShuffler,
		Hasher:              e.coreComponentsHolder.Hasher(),
		PubKey:              pubKey,
		ShardIdAsObserver:   shardId,
		ChanNodeStop:        e.coreComponentsHolder.ChanStopNodeProcess(),
		NodeTypeProvider:    e.coreComponentsHolder.NodeTypeProvider(),
		IsFullArchive:       e.prefsConfig.FullArchive,
		EnableEpochsHandler: e.coreComponentsHolder.EnableEpochsHandler(),
	}

	e.nodesConfigHandler, err = NewSyncValidatorStatus(argsNewValidatorStatusSyncers)
	if err != nil {
		return nil, err
	}

	var miniBlocks []*block.MiniBlock
	e.nodesConfig, e.baseData.shardId, miniBlocks, err = e.nodesConfigHandler.NodesConfigFromMetaBlock(e.epochStartMeta, e.prevEpochStartMeta)
	e.baseData.shardId = e.applyShardIDAsObserverIfNeeded(e.baseData.shardId)

	return miniBlocks, err
}

func (e *epochStartBootstrap) requestAndProcessForMeta(peerMiniBlocks []*block.MiniBlock) error {
	var err error

	storageHandlerComponent, err := NewMetaStorageHandler(
		e.generalConfig,
		e.prefsConfig,
		e.shardCoordinator,
		e.coreComponentsHolder.PathHandler(),
		e.coreComponentsHolder.InternalMarshalizer(),
		e.coreComponentsHolder.Hasher(),
		e.epochStartMeta.GetEpoch(),
		e.coreComponentsHolder.Uint64ByteSliceConverter(),
		e.coreComponentsHolder.NodeTypeProvider(),
		e.nodeProcessingMode,
		e.cryptoComponentsHolder.ManagedPeersHolder(),
	)
	if err != nil {
		return err
	}

	defer storageHandlerComponent.CloseStorageService()

	e.closeTrieComponents()
	triesContainer, trieStorageManagers, err := factory.CreateTriesComponentsForShardId(
		e.generalConfig,
		e.coreComponentsHolder,
		storageHandlerComponent.storageService,
	)
	if err != nil {
		return err
	}

	e.trieContainer = triesContainer
	e.trieStorageManagers = trieStorageManagers

	log.Debug("start in epoch bootstrap: started syncValidatorAccountsState")
	err = e.syncValidatorAccountsState(e.epochStartMeta.GetValidatorStatsRootHash())
	if err != nil {
		return err
	}
	log.Debug("start in epoch bootstrap: syncUserAccountsState")

	err = e.syncUserAccountsState(e.epochStartMeta.GetRootHash())
	if err != nil {
		return err
	}

	pendingMiniBlocks, err := e.getPendingMiniblocks()
	if err != nil {
		return err
	}

	log.Debug("start in epoch bootstrap: GetMiniBlocks", "num synced", len(pendingMiniBlocks))

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: e.epochStartMeta,
		PreviousEpochStart:  e.prevEpochStartMeta,
		NodesConfig:         e.nodesConfig,
		Headers:             e.syncedHeaders,
		ShardCoordinator:    e.shardCoordinator,
		PendingMiniBlocks:   pendingMiniBlocks,
		PeerMiniBlocks:      peerMiniBlocks,
	}

	errSavingToStorage := storageHandlerComponent.SaveDataToStorage(components)
	if errSavingToStorage != nil {
		return errSavingToStorage
	}

	return nil
}

func (e *epochStartBootstrap) getPendingMiniblocks() (map[string]*block.MiniBlock, error) {
	allPendingMiniblocksHeaders := e.computeAllPendingMiniblocksHeaders()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err := e.miniBlocksSyncer.SyncPendingMiniBlocks(allPendingMiniblocksHeaders, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	return e.miniBlocksSyncer.GetMiniBlocks()
}

func (e *epochStartBootstrap) computeAllPendingMiniblocksHeaders() []data.MiniBlockHeaderHandler {
	allPendingMiniblocksHeaders := make([]data.MiniBlockHeaderHandler, 0)
	lastFinalizedHeaderHandlers := e.epochStartMeta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers()

	for _, hdr := range lastFinalizedHeaderHandlers {
		allPendingMiniblocksHeaders = append(allPendingMiniblocksHeaders, hdr.GetPendingMiniBlockHeaderHandlers()...)
	}

	return allPendingMiniblocksHeaders
}

func (e *epochStartBootstrap) findSelfShardEpochStartData() (data.EpochStartShardDataHandler, error) {
	var epochStartData data.EpochStartShardDataHandler
	lastFinalizedHeaderHandlers := e.epochStartMeta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers()
	for i, shardData := range lastFinalizedHeaderHandlers {
		if shardData.GetShardID() == e.shardCoordinator.SelfId() {
			return lastFinalizedHeaderHandlers[i], nil
		}
	}
	return epochStartData, epochStart.ErrEpochStartDataForShardNotFound
}

func (e *epochStartBootstrap) requestAndProcessForShard(peerMiniBlocks []*block.MiniBlock) error {
	epochStartData, err := e.findSelfShardEpochStartData()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err = e.miniBlocksSyncer.SyncPendingMiniBlocks(epochStartData.GetPendingMiniBlockHeaderHandlers(), ctx)
	cancel()
	if err != nil {
		return err
	}

	pendingMiniBlocks, err := e.miniBlocksSyncer.GetMiniBlocks()
	if err != nil {
		return err
	}
	log.Debug("start in epoch bootstrap: GetMiniBlocks", "num synced", len(pendingMiniBlocks))

	shardIds := []uint32{
		core.MetachainShardId,
		core.MetachainShardId,
	}
	lastFinishedMeta := epochStartData.GetLastFinishedMetaBlock()
	firstPendingMetaBlock := epochStartData.GetFirstPendingMetaBlock()
	hashesToRequest := [][]byte{
		lastFinishedMeta,
		firstPendingMetaBlock,
	}

	e.headersSyncer.ClearFields()
	ctx, cancel = context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err = e.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return err
	}

	neededHeaders, err := e.headersSyncer.GetHeaders()
	if err != nil {
		return err
	}
	log.Debug("start in epoch bootstrap: SyncMissingHeadersByHash")

	for hash, hdr := range neededHeaders {
		e.syncedHeaders[hash] = hdr
	}

	shardNotarizedHeader, ok := e.syncedHeaders[string(epochStartData.GetHeaderHash())].(data.ShardHeaderHandler)
	if !ok {
		return epochStart.ErrWrongTypeAssertion
	}

	dts, err := e.getDataToSync(
		epochStartData,
		shardNotarizedHeader,
	)
	if err != nil {
		return err
	}

	for hash, hdr := range dts.additionalHeaders {
		e.syncedHeaders[hash] = hdr
	}

	storageHandlerComponent, err := NewShardStorageHandler(
		e.generalConfig,
		e.prefsConfig,
		e.shardCoordinator,
		e.coreComponentsHolder.PathHandler(),
		e.coreComponentsHolder.InternalMarshalizer(),
		e.coreComponentsHolder.Hasher(),
		e.baseData.lastEpoch,
		e.coreComponentsHolder.Uint64ByteSliceConverter(),
		e.coreComponentsHolder.NodeTypeProvider(),
		e.nodeProcessingMode,
		e.cryptoComponentsHolder.ManagedPeersHolder(),
	)
	if err != nil {
		return err
	}

	defer storageHandlerComponent.CloseStorageService()

	e.closeTrieComponents()
	triesContainer, trieStorageManagers, err := factory.CreateTriesComponentsForShardId(
		e.generalConfig,
		e.coreComponentsHolder,
		storageHandlerComponent.storageService,
	)
	if err != nil {
		return err
	}

	e.trieContainer = triesContainer
	e.trieStorageManagers = trieStorageManagers

	log.Debug("start in epoch bootstrap: started syncUserAccountsState", "rootHash", dts.rootHashToSync)
	err = e.syncUserAccountsState(dts.rootHashToSync)
	if err != nil {
		return err
	}
	log.Debug("start in epoch bootstrap: syncUserAccountsState")

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: e.epochStartMeta,
		PreviousEpochStart:  e.prevEpochStartMeta,
		ShardHeader:         dts.ownShardHdr,
		NodesConfig:         e.nodesConfig,
		Headers:             e.syncedHeaders,
		ShardCoordinator:    e.shardCoordinator,
		PendingMiniBlocks:   pendingMiniBlocks,
		PeerMiniBlocks:      peerMiniBlocks,
	}

	errSavingToStorage := storageHandlerComponent.SaveDataToStorage(components, shardNotarizedHeader, dts.withScheduled, dts.miniBlocks)
	if errSavingToStorage != nil {
		return errSavingToStorage
	}

	return nil
}

func (e *epochStartBootstrap) getDataToSync(
	epochStartData data.EpochStartShardDataHandler,
	shardNotarizedHeader data.ShardHeaderHandler,
) (*dataToSync, error) {
	var err error
	e.storerScheduledSCRs, err = e.storageOpenerHandler.OpenDB(
		e.generalConfig.ScheduledSCRsStorage.DB,
		epochStartData.GetShardID(),
		epochStartData.GetEpoch(),
	)
	if err != nil {
		return nil, err
	}

	res, err := e.updateDataForScheduled(shardNotarizedHeader)
	if err != nil {
		return nil, err
	}

	errClose := e.storerScheduledSCRs.Close()
	log.LogIfError(errClose)
	res.withScheduled = res.ownShardHdr != shardNotarizedHeader

	return res, nil
}

func (e *epochStartBootstrap) updateDataForScheduled(
	shardNotarizedHeader data.ShardHeaderHandler,
) (*dataToSync, error) {

	orderedCollection := ordering.NewOrderedCollection()
	scheduledTxsHandler, err := preprocess.NewScheduledTxsExecution(
		&factoryDisabled.TxProcessor{},
		&factoryDisabled.TxCoordinator{},
		e.storerScheduledSCRs,
		e.coreComponentsHolder.InternalMarshalizer(),
		e.coreComponentsHolder.Hasher(),
		e.shardCoordinator,
		orderedCollection,
	)
	if err != nil {
		return nil, err
	}

	argsScheduledDataSyncer := &types.ScheduledDataSyncerCreateArgs{
		ScheduledTxsHandler:  scheduledTxsHandler,
		HeadersSyncer:        e.headersSyncer,
		MiniBlocksSyncer:     e.miniBlocksSyncer,
		TxSyncer:             e.txSyncerForScheduled,
		ScheduledEnableEpoch: e.coreComponentsHolder.EnableEpochsHandler().GetActivationEpoch(common.ScheduledMiniBlocksFlag),
	}

	e.dataSyncerWithScheduled, err = e.dataSyncerFactory.Create(argsScheduledDataSyncer)
	if err != nil {
		return nil, err
	}

	res := &dataToSync{
		ownShardHdr:       nil,
		rootHashToSync:    nil,
		withScheduled:     false,
		additionalHeaders: nil,
		miniBlocks:        nil,
	}

	res.ownShardHdr, res.additionalHeaders, res.miniBlocks, err = e.dataSyncerWithScheduled.UpdateSyncDataIfNeeded(shardNotarizedHeader)
	if err != nil {
		return nil, err
	}

	res.rootHashToSync = e.dataSyncerWithScheduled.GetRootHashToSync(shardNotarizedHeader)

	return res, nil
}

func (e *epochStartBootstrap) syncUserAccountsState(rootHash []byte) error {
	thr, err := throttler.NewNumGoRoutinesThrottler(int32(e.numConcurrentTrieSyncers))
	if err != nil {
		return err
	}

	e.mutTrieStorageManagers.RLock()
	trieStorageManager := e.trieStorageManagers[dataRetriever.UserAccountsUnit.String()]
	e.mutTrieStorageManagers.RUnlock()

	argsUserAccountsSyncer := syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:                            e.coreComponentsHolder.Hasher(),
			Marshalizer:                       e.coreComponentsHolder.InternalMarshalizer(),
			TrieStorageManager:                trieStorageManager,
			RequestHandler:                    e.requestHandler,
			Timeout:                           common.TimeoutGettingTrieNodes,
			Cacher:                            e.dataPool.TrieNodes(),
			MaxTrieLevelInMemory:              e.generalConfig.StateTriesConfig.MaxStateTrieLevelInMemory,
			MaxHardCapForMissingNodes:         e.maxHardCapForMissingNodes,
			TrieSyncerVersion:                 e.trieSyncerVersion,
			CheckNodesOnDisk:                  e.checkNodesOnDisk,
			UserAccountsSyncStatisticsHandler: e.trieSyncStatisticsProvider,
			AppStatusHandler:                  e.statusHandler,
			EnableEpochsHandler:               e.coreComponentsHolder.EnableEpochsHandler(),
		},
		ShardId:                e.shardCoordinator.SelfId(),
		Throttler:              thr,
		AddressPubKeyConverter: e.coreComponentsHolder.AddressPubKeyConverter(),
	}
	accountsDBSyncer, err := syncer.NewUserAccountsSyncer(argsUserAccountsSyncer)
	if err != nil {
		return err
	}

	err = accountsDBSyncer.SyncAccounts(rootHash, storageMarker.NewTrieStorageMarker())
	if err != nil {
		return err
	}

	return nil
}

func (e *epochStartBootstrap) createStorageService(
	shardCoordinator sharding.Coordinator,
	pathManager storage.PathManagerHandler,
	epochStartNotifier epochStart.EpochStartNotifier,
	startEpoch uint32,
	createTrieEpochRootHashStorer bool,
	targetShardId uint32,
) (dataRetriever.StorageService, error) {
	storageServiceCreator, err := storageFactory.NewStorageServiceFactory(
		storageFactory.StorageServiceFactoryArgs{
			Config:                        e.generalConfig,
			PrefsConfig:                   e.prefsConfig,
			ShardCoordinator:              shardCoordinator,
			PathManager:                   pathManager,
			EpochStartNotifier:            epochStartNotifier,
			NodeTypeProvider:              e.coreComponentsHolder.NodeTypeProvider(),
			CurrentEpoch:                  startEpoch,
			StorageType:                   storageFactory.BootstrapStorageService,
			CreateTrieEpochRootHashStorer: createTrieEpochRootHashStorer,
			NodeProcessingMode:            e.nodeProcessingMode,
			RepopulateTokensSupplies:      e.flagsConfig.RepopulateTokensSupplies,
			ManagedPeersHolder:            e.cryptoComponentsHolder.ManagedPeersHolder(),
		})
	if err != nil {
		return nil, err
	}

	if targetShardId == core.MetachainShardId {
		return storageServiceCreator.CreateForMeta()
	}

	return storageServiceCreator.CreateForShard()
}

func (e *epochStartBootstrap) syncValidatorAccountsState(rootHash []byte) error {
	e.mutTrieStorageManagers.RLock()
	peerTrieStorageManager := e.trieStorageManagers[dataRetriever.PeerAccountsUnit.String()]
	e.mutTrieStorageManagers.RUnlock()

	argsValidatorAccountsSyncer := syncer.ArgsNewValidatorAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:                            e.coreComponentsHolder.Hasher(),
			Marshalizer:                       e.coreComponentsHolder.InternalMarshalizer(),
			TrieStorageManager:                peerTrieStorageManager,
			RequestHandler:                    e.requestHandler,
			Timeout:                           common.TimeoutGettingTrieNodes,
			Cacher:                            e.dataPool.TrieNodes(),
			MaxTrieLevelInMemory:              e.generalConfig.StateTriesConfig.MaxPeerTrieLevelInMemory,
			MaxHardCapForMissingNodes:         e.maxHardCapForMissingNodes,
			TrieSyncerVersion:                 e.trieSyncerVersion,
			CheckNodesOnDisk:                  e.checkNodesOnDisk,
			UserAccountsSyncStatisticsHandler: statistics.NewTrieSyncStatistics(),
			AppStatusHandler:                  disabledCommon.NewAppStatusHandler(),
			EnableEpochsHandler:               e.coreComponentsHolder.EnableEpochsHandler(),
		},
	}
	accountsDBSyncer, err := syncer.NewValidatorAccountsSyncer(argsValidatorAccountsSyncer)
	if err != nil {
		return err
	}

	err = accountsDBSyncer.SyncAccounts(rootHash, storageMarker.NewTrieStorageMarker())
	if err != nil {
		return err
	}

	return nil
}

func (e *epochStartBootstrap) createResolversContainer() error {
	dataPacker, err := partitioning.NewSimpleDataPacker(e.coreComponentsHolder.InternalMarshalizer())
	if err != nil {
		return err
	}

	storageService := disabled.NewChainStorer()

	payloadValidator, err := validator.NewPeerAuthenticationPayloadValidator(e.generalConfig.HeartbeatV2.HeartbeatExpiryTimespanInSec)
	if err != nil {
		return err
	}

	// TODO - create a dedicated request handler to be used when fetching required data with the correct shard coordinator
	//  this one should only be used before determining the correct shard where the node should reside
	log.Debug("epochStartBootstrap.createRequestHandler", "shard", e.shardCoordinator.SelfId())
	resolversContainerArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:                e.shardCoordinator,
		MainMessenger:                   e.mainMessenger,
		FullArchiveMessenger:            e.fullArchiveMessenger,
		Store:                           storageService,
		Marshalizer:                     e.coreComponentsHolder.InternalMarshalizer(),
		DataPools:                       e.dataPool,
		Uint64ByteSliceConverter:        uint64ByteSlice.NewBigEndianConverter(),
		NumConcurrentResolvingJobs:      10,
		DataPacker:                      dataPacker,
		TriesContainer:                  e.trieContainer,
		SizeCheckDelta:                  0,
		InputAntifloodHandler:           disabled.NewAntiFloodHandler(),
		OutputAntifloodHandler:          disabled.NewAntiFloodHandler(),
		MainPreferredPeersHolder:        disabled.NewPreferredPeersHolder(),
		FullArchivePreferredPeersHolder: disabled.NewPreferredPeersHolder(),
		PayloadValidator:                payloadValidator,
	}
	resolverFactory, err := resolverscontainer.NewMetaResolversContainerFactory(resolversContainerArgs)
	if err != nil {
		return err
	}

	container, err := resolverFactory.Create()
	if err != nil {
		return err
	}

	return resolverFactory.AddShardTrieNodeResolvers(container)
}

func (e *epochStartBootstrap) createRequestHandler() error {
	requestersContainerArgs := requesterscontainer.FactoryArgs{
		RequesterConfig:                 e.generalConfig.Requesters,
		ShardCoordinator:                e.shardCoordinator,
		MainMessenger:                   e.mainMessenger,
		FullArchiveMessenger:            e.fullArchiveMessenger,
		Marshaller:                      e.coreComponentsHolder.InternalMarshalizer(),
		Uint64ByteSliceConverter:        uint64ByteSlice.NewBigEndianConverter(),
		OutputAntifloodHandler:          disabled.NewAntiFloodHandler(),
		CurrentNetworkEpochProvider:     disabled.NewCurrentNetworkEpochProviderHandler(),
		MainPreferredPeersHolder:        disabled.NewPreferredPeersHolder(),
		FullArchivePreferredPeersHolder: disabled.NewPreferredPeersHolder(),
		PeersRatingHandler:              disabled.NewDisabledPeersRatingHandler(),
		SizeCheckDelta:                  0,
	}
	requestersFactory, err := requesterscontainer.NewMetaRequestersContainerFactory(requestersContainerArgs)
	if err != nil {
		return err
	}

	container, err := requestersFactory.Create()
	if err != nil {
		return err
	}

	err = requestersFactory.AddShardTrieNodeRequesters(container)
	if err != nil {
		return err
	}

	finder, err := containers.NewRequestersFinder(container, e.shardCoordinator)
	if err != nil {
		return err
	}

	requestedItemsHandler := cache.NewTimeCache(timeBetweenRequests)
	e.requestHandler, err = requestHandlers.NewResolverRequestHandler(
		finder,
		requestedItemsHandler,
		e.whiteListHandler,
		maxToRequest,
		core.MetachainShardId,
		timeBetweenRequests,
	)
	return err
}

func (e *epochStartBootstrap) setEpochStartMetrics() {
	if !check.IfNil(e.epochStartMeta) {
		metablockEconomics := e.epochStartMeta.GetEpochStartHandler().GetEconomicsHandler()
		e.statusHandler.SetStringValue(common.MetricTotalSupply, metablockEconomics.GetTotalSupply().String())
		e.statusHandler.SetStringValue(common.MetricInflation, metablockEconomics.GetTotalNewlyMinted().String())
		e.statusHandler.SetStringValue(common.MetricTotalFees, e.epochStartMeta.GetAccumulatedFees().String())
		e.statusHandler.SetStringValue(common.MetricDevRewardsInEpoch, e.epochStartMeta.GetDevFeesInEpoch().String())
		e.statusHandler.SetUInt64Value(common.MetricEpochForEconomicsData, uint64(e.epochStartMeta.GetEpoch()))
	}
}

func (e *epochStartBootstrap) applyShardIDAsObserverIfNeeded(receivedShardID uint32) uint32 {
	if e.nodeType == core.NodeTypeObserver &&
		e.destinationShardAsObserver != common.DisabledShardIDAsObserver &&
		e.destinationShardAsObserver != receivedShardID {
		log.Debug("shard id as observer applied", "destination shard ID", e.destinationShardAsObserver, "computed", receivedShardID)
		receivedShardID = e.destinationShardAsObserver
	}

	return receivedShardID
}

func (e *epochStartBootstrap) createHeartbeatSender() error {
	privateKey := e.cryptoComponentsHolder.PrivateKey()
	bootstrapRedundancy, err := redundancy.NewBootstrapNodeRedundancy(privateKey)
	if err != nil {
		return err
	}

	heartbeatTopic := common.HeartbeatV2Topic + e.shardCoordinator.CommunicationIdentifier(e.shardCoordinator.SelfId())
	if !e.mainMessenger.HasTopic(heartbeatTopic) {
		err = e.mainMessenger.CreateTopic(heartbeatTopic, true)
		if err != nil {
			return err
		}
	}

	if !e.fullArchiveMessenger.HasTopic(heartbeatTopic) {
		err = e.fullArchiveMessenger.CreateTopic(heartbeatTopic, true)
		if err != nil {
			return err
		}
	}

	peerSubType := core.RegularPeer
	if e.prefsConfig.FullArchive {
		peerSubType = core.FullHistoryObserver
	}
	heartbeatCfg := e.generalConfig.HeartbeatV2
	argsHeartbeatSender := sender.ArgBootstrapSender{
		MainMessenger:                      e.mainMessenger,
		FullArchiveMessenger:               e.fullArchiveMessenger,
		Marshaller:                         e.coreComponentsHolder.InternalMarshalizer(),
		HeartbeatTopic:                     heartbeatTopic,
		HeartbeatTimeBetweenSends:          time.Second * time.Duration(heartbeatCfg.HeartbeatTimeBetweenSendsDuringBootstrapInSec),
		HeartbeatTimeBetweenSendsWhenError: time.Second * time.Duration(heartbeatCfg.HeartbeatTimeBetweenSendsWhenErrorInSec),
		HeartbeatTimeThresholdBetweenSends: heartbeatCfg.HeartbeatTimeThresholdBetweenSends,
		VersionNumber:                      e.flagsConfig.Version,
		NodeDisplayName:                    e.prefsConfig.NodeDisplayName,
		Identity:                           e.prefsConfig.Identity,
		PeerSubType:                        peerSubType,
		CurrentBlockProvider:               blockchain.NewBootstrapBlockchain(),
		PrivateKey:                         privateKey,
		RedundancyHandler:                  bootstrapRedundancy,
		PeerTypeProvider:                   peer.NewBootstrapPeerTypeProvider(),
		TrieSyncStatisticsProvider:         e.trieSyncStatisticsProvider,
	}

	e.bootstrapHeartbeatSender, err = sender.NewBootstrapSender(argsHeartbeatSender)
	return err
}

func (e *epochStartBootstrap) closeTrieComponents() {
	if e.trieStorageManagers != nil {
		log.Debug("closing all trieStorageManagers", "num", len(e.trieStorageManagers))
		for _, tsm := range e.trieStorageManagers {
			err := tsm.Close()
			log.LogIfError(err)
		}
	}

	if !check.IfNil(e.trieContainer) {
		tries := e.trieContainer.GetAll()
		log.Debug("closing all tries", "num", len(tries))
		for _, trie := range tries {
			err := trie.Close()
			log.LogIfError(err)
		}
	}

	if !check.IfNil(e.storageService) {
		err := e.storageService.Destroy()
		log.LogIfError(err)
	}
}

func (e *epochStartBootstrap) closeBootstrapHeartbeatSender() {
	if !check.IfNil(e.bootstrapHeartbeatSender) {
		log.LogIfError(e.bootstrapHeartbeatSender.Close())
	}
}

// Close closes the component's opened storage services/started go-routines
func (e *epochStartBootstrap) Close() error {
	e.mutTrieStorageManagers.RLock()
	defer e.mutTrieStorageManagers.RUnlock()

	e.closeTrieComponents()

	var err error
	if !check.IfNil(e.dataPool) {
		err = e.dataPool.Close()
	}

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (e *epochStartBootstrap) IsInterfaceNil() bool {
	return e == nil
}
