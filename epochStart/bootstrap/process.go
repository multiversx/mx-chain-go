package bootstrap

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-core/core/throttler"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters/uint64ByteSlice"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	factoryDataPool "github.com/ElrondNetwork/elrond-go/dataRetriever/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	factoryInterceptors "github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/factory"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	disabledInterceptors "github.com/ElrondNetwork/elrond-go/process/interceptors/disabled"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/syncer"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/trie/factory"
	"github.com/ElrondNetwork/elrond-go/update"
	updateSync "github.com/ElrondNetwork/elrond-go/update/sync"
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
	NodesConfig *sharding.NodesCoordinatorRegistry
}

// ComponentsNeededForBootstrap holds the components which need to be initialized from network
type ComponentsNeededForBootstrap struct {
	EpochStartMetaBlock *block.MetaBlock
	PreviousEpochStart  *block.MetaBlock
	ShardHeader         *block.Header
	NodesConfig         *sharding.NodesCoordinatorRegistry
	Headers             map[string]data.HeaderHandler
	ShardCoordinator    sharding.Coordinator
	PendingMiniBlocks   map[string]*block.MiniBlock
}

// epochStartBootstrap will handle requesting the needed data to start when joining late the network
type epochStartBootstrap struct {
	// should come via arguments
	destinationShardAsObserver uint32
	coreComponentsHolder       process.CoreComponentsHolder
	cryptoComponentsHolder     process.CryptoComponentsHolder
	messenger                  Messenger
	generalConfig              config.Config
	prefsConfig                config.PreferencesConfig
	economicsData              process.EconomicsDataHandler
	shardCoordinator           sharding.Coordinator
	genesisNodesConfig         sharding.GenesisNodesSetupHandler
	genesisShardCoordinator    sharding.Coordinator
	rater                      sharding.ChanceComputer
	trieContainer              state.TriesHolder
	trieStorageManagers        map[string]temporary.StorageManager
	mutTrieStorageManagers     sync.RWMutex
	nodeShuffler               sharding.NodesShuffler
	roundHandler               epochStart.RoundHandler
	statusHandler              core.AppStatusHandler
	headerIntegrityVerifier    process.HeaderIntegrityVerifier
	enableSignTxWithHashEpoch  uint32
	epochNotifier              process.EpochNotifier
	numConcurrentTrieSyncers   int
	maxHardCapForMissingNodes  int
	trieSyncerVersion          int

	// created components
	requestHandler            process.RequestHandler
	interceptorContainer      process.InterceptorsContainer
	dataPool                  dataRetriever.PoolsHolder
	miniBlocksSyncer          epochStart.PendingMiniBlocksSyncHandler
	headersSyncer             epochStart.HeadersByHashSyncer
	epochStartMetaBlockSyncer epochStart.StartOfEpochMetaSyncer
	nodesConfigHandler        StartOfEpochNodesConfigHandler
	whiteListHandler          update.WhiteListHandler
	whiteListerVerifiedTxs    update.WhiteListHandler
	storageOpenerHandler      storage.UnitOpenerHandler
	latestStorageDataProvider storage.LatestStorageDataProviderHandler
	argumentsParser           process.ArgumentsParser
	waitingListFixEnableEpoch uint32

	// gathered data
	epochStartMeta     *block.MetaBlock
	prevEpochStartMeta *block.MetaBlock
	syncedHeaders      map[string]data.HeaderHandler
	nodesConfig        *sharding.NodesCoordinatorRegistry
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
	Messenger                  Messenger
	GeneralConfig              config.Config
	PrefsConfig                config.PreferencesConfig
	EpochConfig                config.EpochConfig
	EconomicsData              process.EconomicsDataHandler
	GenesisNodesConfig         sharding.GenesisNodesSetupHandler
	GenesisShardCoordinator    sharding.Coordinator
	StorageUnitOpener          storage.UnitOpenerHandler
	LatestStorageDataProvider  storage.LatestStorageDataProviderHandler
	Rater                      sharding.ChanceComputer
	NodeShuffler               sharding.NodesShuffler
	RoundHandler               epochStart.RoundHandler
	ArgumentsParser            process.ArgumentsParser
	StatusHandler              core.AppStatusHandler
	HeaderIntegrityVerifier    process.HeaderIntegrityVerifier
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
		messenger:                  args.Messenger,
		generalConfig:              args.GeneralConfig,
		prefsConfig:                args.PrefsConfig,
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
		enableSignTxWithHashEpoch:  args.EpochConfig.EnableEpochs.TransactionSignedWithTxHashEnableEpoch,
		epochNotifier:              args.CoreComponentsHolder.EpochNotifier(),
		numConcurrentTrieSyncers:   args.GeneralConfig.TrieSync.NumConcurrentTrieSyncers,
		maxHardCapForMissingNodes:  args.GeneralConfig.TrieSync.MaxHardCapForMissingNodes,
		trieSyncerVersion:          args.GeneralConfig.TrieSync.TrieSyncerVersion,
		waitingListFixEnableEpoch:  args.EpochConfig.EnableEpochs.WaitingListFixEnableEpoch,
	}

	log.Debug("process: enable epoch for transaction signed with tx hash", "epoch", epochStartProvider.enableSignTxWithHashEpoch)

	whiteListCache, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(epochStartProvider.generalConfig.WhiteListPool))
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
	epochStartProvider.trieStorageManagers = make(map[string]temporary.StorageManager)

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
	err := e.createTriesComponentsForShardId(e.genesisShardCoordinator.SelfId())
	if err != nil {
		return Parameters{}, err
	}

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

// GetTriesComponents returns the created tries components according to the shardID for the current epoch
func (e *epochStartBootstrap) GetTriesComponents() (state.TriesHolder, map[string]temporary.StorageManager) {
	e.mutTrieStorageManagers.RLock()
	defer e.mutTrieStorageManagers.RUnlock()

	storageManagers := make(map[string]temporary.StorageManager)
	for k, v := range e.trieStorageManagers {
		storageManagers[k] = v
	}

	return e.trieContainer, storageManagers
}

// Bootstrap runs the fast bootstrap method from the network or local storage
func (e *epochStartBootstrap) Bootstrap() (Parameters, error) {
	if !e.generalConfig.GeneralSettings.StartInEpochEnabled {
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
	log.Debug("start in epoch bootstrap: got epoch start meta header", "epoch", e.epochStartMeta.Epoch, "nonce", e.epochStartMeta.Nonce)
	e.setEpochStartMetrics()

	err = e.createSyncers()
	if err != nil {
		return Parameters{}, err
	}

	defer func() {
		errClose := e.interceptorContainer.Close()
		if errClose != nil {
			log.Warn("prepareEpochFromStorage interceptorContainer.Close()", "error", errClose)
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
		err := e.createTriesComponentsForShardId(e.genesisShardCoordinator.SelfId())
		if err != nil {
			return Parameters{}, err
		}

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

	err = e.createTriesComponentsForShardId(newShardId)
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
	errMessenger := e.messenger.UnregisterAllMessageProcessors()
	log.LogIfError(errMessenger)

	errMessenger = e.messenger.UnjoinAllTopics()
	log.LogIfError(errMessenger)
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
	err := e.createTriesComponentsForShardId(core.MetachainShardId)
	if err != nil {
		return err
	}

	err = e.createRequestHandler()
	if err != nil {
		return err
	}

	epochStartConfig := e.generalConfig.EpochStartConfig
	metablockProcessor, err := NewEpochStartMetaBlockProcessor(
		e.messenger,
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
		Messenger:               e.messenger,
		ShardCoordinator:        e.shardCoordinator,
		EconomicsData:           e.economicsData,
		WhitelistHandler:        e.whiteListHandler,
		StartInEpochConfig:      epochStartConfig,
		HeaderIntegrityVerifier: e.headerIntegrityVerifier,
		MetaBlockProcessor:      metablockProcessor,
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
		CoreComponents:            e.coreComponentsHolder,
		CryptoComponents:          e.cryptoComponentsHolder,
		Config:                    e.generalConfig,
		ShardCoordinator:          e.shardCoordinator,
		Messenger:                 e.messenger,
		DataPool:                  e.dataPool,
		WhiteListHandler:          e.whiteListHandler,
		WhiteListerVerifiedTxs:    e.whiteListerVerifiedTxs,
		ArgumentsParser:           e.argumentsParser,
		HeaderIntegrityVerifier:   e.headerIntegrityVerifier,
		EnableSignTxWithHashEpoch: e.enableSignTxWithHashEpoch,
		EpochNotifier:             e.epochNotifier,
		RequestHandler:            e.requestHandler,
	}

	e.interceptorContainer, err = factoryInterceptors.NewEpochStartInterceptorsContainer(args)
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

	return nil
}

func (e *epochStartBootstrap) syncHeadersFrom(meta *block.MetaBlock) (map[string]data.HeaderHandler, error) {
	hashesToRequest := make([][]byte, 0, len(meta.EpochStart.LastFinalizedHeaders)+1)
	shardIds := make([]uint32, 0, len(meta.EpochStart.LastFinalizedHeaders)+1)

	for _, epochStartData := range meta.EpochStart.LastFinalizedHeaders {
		hashesToRequest = append(hashesToRequest, epochStartData.HeaderHash)
		shardIds = append(shardIds, epochStartData.ShardID)
	}

	if meta.Epoch > e.startEpoch+1 { // no need to request genesis block
		hashesToRequest = append(hashesToRequest, meta.EpochStart.Economics.PrevEpochStartHash)
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

	if meta.Epoch == e.startEpoch+1 {
		syncedHeaders[string(meta.EpochStart.Economics.PrevEpochStartHash)] = &block.MetaBlock{}
	}

	return syncedHeaders, nil
}

// Bootstrap will handle requesting and receiving the needed information the node will bootstrap from
func (e *epochStartBootstrap) requestAndProcessing() (Parameters, error) {
	var err error
	e.baseData.numberOfShards = uint32(len(e.epochStartMeta.EpochStart.LastFinalizedHeaders))
	e.baseData.lastEpoch = e.epochStartMeta.Epoch

	e.syncedHeaders, err = e.syncHeadersFrom(e.epochStartMeta)
	if err != nil {
		return Parameters{}, err
	}
	log.Debug("start in epoch bootstrap: got shard headers and previous epoch start meta block")

	prevEpochStartMetaHash := e.epochStartMeta.EpochStart.Economics.PrevEpochStartHash
	prevEpochStartMeta, ok := e.syncedHeaders[string(prevEpochStartMetaHash)].(*block.MetaBlock)
	if !ok {
		return Parameters{}, epochStart.ErrWrongTypeAssertion
	}
	e.prevEpochStartMeta = prevEpochStartMeta

	pubKeyBytes, err := e.cryptoComponentsHolder.PublicKey().ToByteArray()
	if err != nil {
		return Parameters{}, err
	}

	err = e.processNodesConfig(pubKeyBytes)
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

	err = e.messenger.CreateTopic(common.ConsensusTopic+e.shardCoordinator.CommunicationIdentifier(e.shardCoordinator.SelfId()), true)
	if err != nil {
		return Parameters{}, err
	}

	if e.shardCoordinator.SelfId() == core.MetachainShardId {
		err = e.requestAndProcessForMeta()
		if err != nil {
			return Parameters{}, err
		}
	} else {
		err = e.createTriesComponentsForShardId(e.shardCoordinator.SelfId())
		if err != nil {
			return Parameters{}, err
		}

		err = e.requestAndProcessForShard()
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

func (e *epochStartBootstrap) processNodesConfig(pubKey []byte) error {
	var err error
	shardId := e.destinationShardAsObserver
	if shardId > e.baseData.numberOfShards && shardId != core.MetachainShardId {
		shardId = e.genesisShardCoordinator.SelfId()
	}
	argsNewValidatorStatusSyncers := ArgsNewSyncValidatorStatus{
		DataPool:                  e.dataPool,
		Marshalizer:               e.coreComponentsHolder.InternalMarshalizer(),
		RequestHandler:            e.requestHandler,
		ChanceComputer:            e.rater,
		GenesisNodesConfig:        e.genesisNodesConfig,
		NodeShuffler:              e.nodeShuffler,
		Hasher:                    e.coreComponentsHolder.Hasher(),
		PubKey:                    pubKey,
		ShardIdAsObserver:         shardId,
		WaitingListFixEnableEpoch: e.waitingListFixEnableEpoch,
		ChanNodeStop:              e.coreComponentsHolder.ChanStopNodeProcess(),
		NodeTypeProvider:          e.coreComponentsHolder.NodeTypeProvider(),
		IsFullArchive:             e.prefsConfig.FullArchive,
	}

	e.nodesConfigHandler, err = NewSyncValidatorStatus(argsNewValidatorStatusSyncers)
	if err != nil {
		return err
	}

	e.nodesConfig, e.baseData.shardId, err = e.nodesConfigHandler.NodesConfigFromMetaBlock(e.epochStartMeta, e.prevEpochStartMeta)
	e.baseData.shardId = e.applyShardIDAsObserverIfNeeded(e.baseData.shardId)

	return err
}

func (e *epochStartBootstrap) requestAndProcessForMeta() error {
	var err error

	log.Debug("start in epoch bootstrap: started syncValidatorAccountsState")
	err = e.syncValidatorAccountsState(e.epochStartMeta.ValidatorStatsRootHash)
	if err != nil {
		return err
	}
	log.Debug("start in epoch bootstrap: syncUserAccountsState")

	err = e.syncUserAccountsState(e.epochStartMeta.RootHash)
	if err != nil {
		return err
	}

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: e.epochStartMeta,
		PreviousEpochStart:  e.prevEpochStartMeta,
		NodesConfig:         e.nodesConfig,
		Headers:             e.syncedHeaders,
		ShardCoordinator:    e.shardCoordinator,
	}

	storageHandlerComponent, err := NewMetaStorageHandler(
		e.generalConfig,
		e.prefsConfig,
		e.shardCoordinator,
		e.coreComponentsHolder.PathHandler(),
		e.coreComponentsHolder.InternalMarshalizer(),
		e.coreComponentsHolder.Hasher(),
		e.epochStartMeta.Epoch,
		e.coreComponentsHolder.Uint64ByteSliceConverter(),
		e.coreComponentsHolder.NodeTypeProvider(),
	)
	if err != nil {
		return err
	}

	errSavingToStorage := storageHandlerComponent.SaveDataToStorage(components)
	if errSavingToStorage != nil {
		return errSavingToStorage
	}

	return nil
}

func (e *epochStartBootstrap) findSelfShardEpochStartData() (block.EpochStartShardData, error) {
	var epochStartData block.EpochStartShardData
	for _, shardData := range e.epochStartMeta.EpochStart.LastFinalizedHeaders {
		if shardData.ShardID == e.shardCoordinator.SelfId() {
			return shardData, nil
		}
	}
	return epochStartData, epochStart.ErrEpochStartDataForShardNotFound
}

func (e *epochStartBootstrap) requestAndProcessForShard() error {
	epochStartData, err := e.findSelfShardEpochStartData()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err = e.miniBlocksSyncer.SyncPendingMiniBlocks(epochStartData.PendingMiniBlockHeaders, ctx)
	cancel()
	if err != nil {
		return err
	}

	pendingMiniBlocks, err := e.miniBlocksSyncer.GetMiniBlocks()
	if err != nil {
		return err
	}
	log.Debug("start in epoch bootstrap: GetMiniBlocks")

	shardIds := []uint32{
		core.MetachainShardId,
		core.MetachainShardId,
	}
	hashesToRequest := [][]byte{
		epochStartData.LastFinishedMetaBlock,
		epochStartData.FirstPendingMetaBlock,
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

	ownShardHdr, ok := e.syncedHeaders[string(epochStartData.HeaderHash)].(*block.Header)
	if !ok {
		return epochStart.ErrWrongTypeAssertion
	}

	log.Debug("start in epoch bootstrap: started syncUserAccountsState")
	err = e.syncUserAccountsState(ownShardHdr.RootHash)
	if err != nil {
		return err
	}
	log.Debug("start in epoch bootstrap: syncUserAccountsState")

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: e.epochStartMeta,
		PreviousEpochStart:  e.prevEpochStartMeta,
		ShardHeader:         ownShardHdr,
		NodesConfig:         e.nodesConfig,
		Headers:             e.syncedHeaders,
		ShardCoordinator:    e.shardCoordinator,
		PendingMiniBlocks:   pendingMiniBlocks,
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
	)
	if err != nil {
		return err
	}

	errSavingToStorage := storageHandlerComponent.SaveDataToStorage(components)
	if errSavingToStorage != nil {
		return errSavingToStorage
	}

	return nil
}

func (e *epochStartBootstrap) syncUserAccountsState(rootHash []byte) error {
	thr, err := throttler.NewNumGoRoutinesThrottler(int32(e.numConcurrentTrieSyncers))
	if err != nil {
		return err
	}

	e.mutTrieStorageManagers.RLock()
	trieStorageManager := e.trieStorageManagers[factory.UserAccountTrie]
	e.mutTrieStorageManagers.RUnlock()

	argsUserAccountsSyncer := syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:                    e.coreComponentsHolder.Hasher(),
			Marshalizer:               e.coreComponentsHolder.InternalMarshalizer(),
			TrieStorageManager:        trieStorageManager,
			RequestHandler:            e.requestHandler,
			Timeout:                   common.TimeoutGettingTrieNodes,
			Cacher:                    e.dataPool.TrieNodes(),
			MaxTrieLevelInMemory:      e.generalConfig.StateTriesConfig.MaxStateTrieLevelInMemory,
			MaxHardCapForMissingNodes: e.maxHardCapForMissingNodes,
			TrieSyncerVersion:         e.trieSyncerVersion,
		},
		ShardId:   e.shardCoordinator.SelfId(),
		Throttler: thr,
	}
	accountsDBSyncer, err := syncer.NewUserAccountsSyncer(argsUserAccountsSyncer)
	if err != nil {
		return err
	}

	err = accountsDBSyncer.SyncAccounts(rootHash)
	if err != nil {
		return err
	}

	return nil
}

func (e *epochStartBootstrap) createTriesComponentsForShardId(shardId uint32) error {
	e.tryCloseExisting(factory.UserAccountTrie)
	e.tryCloseExisting(factory.PeerAccountTrie)

	trieFactoryArgs := factory.TrieFactoryArgs{
		SnapshotDbCfg:            e.generalConfig.TrieSnapshotDB,
		Marshalizer:              e.coreComponentsHolder.InternalMarshalizer(),
		Hasher:                   e.coreComponentsHolder.Hasher(),
		PathManager:              e.coreComponentsHolder.PathHandler(),
		TrieStorageManagerConfig: e.generalConfig.TrieStorageManagerConfig,
	}
	trieFactory, err := factory.NewTrieFactory(trieFactoryArgs)
	if err != nil {
		return err
	}

	args := temporary.TrieCreateArgs{
		TrieStorageConfig:  e.generalConfig.AccountsTrieStorage,
		ShardID:            core.GetShardIDString(shardId),
		PruningEnabled:     e.generalConfig.StateTriesConfig.AccountsStatePruningEnabled,
		CheckpointsEnabled: e.generalConfig.StateTriesConfig.CheckpointsEnabled,
		MaxTrieLevelInMem:  e.generalConfig.StateTriesConfig.MaxStateTrieLevelInMemory,
	}
	userStorageManager, userAccountTrie, err := trieFactory.Create(args)
	if err != nil {
		return err
	}

	e.trieContainer.Put([]byte(factory.UserAccountTrie), userAccountTrie)
	e.mutTrieStorageManagers.Lock()
	e.trieStorageManagers[factory.UserAccountTrie] = userStorageManager
	e.mutTrieStorageManagers.Unlock()

	args = temporary.TrieCreateArgs{
		TrieStorageConfig:  e.generalConfig.PeerAccountsTrieStorage,
		ShardID:            core.GetShardIDString(shardId),
		PruningEnabled:     e.generalConfig.StateTriesConfig.PeerStatePruningEnabled,
		CheckpointsEnabled: e.generalConfig.StateTriesConfig.CheckpointsEnabled,
		MaxTrieLevelInMem:  e.generalConfig.StateTriesConfig.MaxPeerTrieLevelInMemory,
	}
	peerStorageManager, peerAccountsTrie, err := trieFactory.Create(args)
	if err != nil {
		return err
	}

	e.mutTrieStorageManagers.Lock()
	e.trieContainer.Put([]byte(factory.PeerAccountTrie), peerAccountsTrie)
	e.trieStorageManagers[factory.PeerAccountTrie] = peerStorageManager
	e.mutTrieStorageManagers.Unlock()

	return nil
}

func (e *epochStartBootstrap) tryCloseExisting(trieType string) {
	e.mutTrieStorageManagers.RLock()
	existingStorageManager := e.trieStorageManagers[trieType]
	e.mutTrieStorageManagers.RUnlock()
	if !check.IfNil(existingStorageManager) {
		err := existingStorageManager.Close()
		if err != nil {
			log.Warn("failed to close existing storage manager", "error", err)
		}
	}
}

func (e *epochStartBootstrap) syncValidatorAccountsState(rootHash []byte) error {
	e.mutTrieStorageManagers.RLock()
	peerTrieStorageManager := e.trieStorageManagers[factory.PeerAccountTrie]
	e.mutTrieStorageManagers.RUnlock()

	argsValidatorAccountsSyncer := syncer.ArgsNewValidatorAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:                    e.coreComponentsHolder.Hasher(),
			Marshalizer:               e.coreComponentsHolder.InternalMarshalizer(),
			TrieStorageManager:        peerTrieStorageManager,
			RequestHandler:            e.requestHandler,
			Timeout:                   common.TimeoutGettingTrieNodes,
			Cacher:                    e.dataPool.TrieNodes(),
			MaxTrieLevelInMemory:      e.generalConfig.StateTriesConfig.MaxPeerTrieLevelInMemory,
			MaxHardCapForMissingNodes: e.maxHardCapForMissingNodes,
			TrieSyncerVersion:         e.trieSyncerVersion,
		},
	}
	accountsDBSyncer, err := syncer.NewValidatorAccountsSyncer(argsValidatorAccountsSyncer)
	if err != nil {
		return err
	}

	err = accountsDBSyncer.SyncAccounts(rootHash)
	if err != nil {
		return err
	}

	return nil
}

func (e *epochStartBootstrap) createRequestHandler() error {
	dataPacker, err := partitioning.NewSimpleDataPacker(e.coreComponentsHolder.InternalMarshalizer())
	if err != nil {
		return err
	}

	storageService := disabled.NewChainStorer()

	resolversContainerArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:            e.shardCoordinator,
		Messenger:                   e.messenger,
		Store:                       storageService,
		Marshalizer:                 e.coreComponentsHolder.InternalMarshalizer(),
		DataPools:                   e.dataPool,
		Uint64ByteSliceConverter:    uint64ByteSlice.NewBigEndianConverter(),
		NumConcurrentResolvingJobs:  10,
		DataPacker:                  dataPacker,
		TriesContainer:              e.trieContainer,
		SizeCheckDelta:              0,
		InputAntifloodHandler:       disabled.NewAntiFloodHandler(),
		OutputAntifloodHandler:      disabled.NewAntiFloodHandler(),
		CurrentNetworkEpochProvider: disabled.NewCurrentNetworkEpochProviderHandler(),
		PreferredPeersHolder:        disabled.NewPreferredPeersHolder(),
		ResolverConfig:              e.generalConfig.Resolvers,
	}
	resolverFactory, err := resolverscontainer.NewMetaResolversContainerFactory(resolversContainerArgs)
	if err != nil {
		return err
	}

	container, err := resolverFactory.Create()
	if err != nil {
		return err
	}

	err = resolverFactory.AddShardTrieNodeResolvers(container)
	if err != nil {
		return err
	}

	finder, err := containers.NewResolversFinder(container, e.shardCoordinator)
	if err != nil {
		return err
	}

	requestedItemsHandler := timecache.NewTimeCache(timeBetweenRequests)
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
	if e.epochStartMeta != nil {
		e.statusHandler.SetUInt64Value(common.MetricNonceAtEpochStart, e.epochStartMeta.Nonce)
		e.statusHandler.SetUInt64Value(common.MetricRoundAtEpochStart, e.epochStartMeta.Round)
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

// Close closes the component's opened storage services/started go-routines
func (e *epochStartBootstrap) Close() error {
	e.mutTrieStorageManagers.RLock()
	defer e.mutTrieStorageManagers.RUnlock()

	if e.trieStorageManagers != nil {
		log.Debug("closing all trieStorageManagers....")
		for _, tsm := range e.trieStorageManagers {
			err := tsm.Close()
			log.LogIfError(err)
		}
	}

	if !check.IfNil(e.dataPool) && !check.IfNil(e.dataPool.TrieNodes()) {
		log.Debug("closing trie nodes data pool....")
		err := e.dataPool.TrieNodes().Close()
		log.LogIfError(err)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (e *epochStartBootstrap) IsInterfaceNil() bool {
	return e == nil
}
