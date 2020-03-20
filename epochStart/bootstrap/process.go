package bootstrap

import (
	"encoding/hex"
	"errors"
	"math"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	trieFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	factory2 "github.com/ElrondNetwork/elrond-go/dataRetriever/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	factory3 "github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/factory"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/storagehandler"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/sync"
)

var log = logger.GetOrCreate("epochStart/bootstrap")

const delayAfterRequesting = 1 * time.Second
const thresholdForConsideringMetaBlockCorrect = 0.2
const maxNumTimesToRetry = 100

// ComponentsNeededForBootstrap holds the components which need to be initialized from network
type ComponentsNeededForBootstrap struct {
	EpochStartMetaBlock         *block.MetaBlock
	PreviousEpochStartMetaBlock *block.MetaBlock
	ShardHeader                 *block.Header //only for shards, nil for meta
	NodesConfig                 *sharding.NodesSetup
	ShardHeaders                map[uint32]*block.Header
	ShardCoordinator            sharding.Coordinator
	Tries                       state.TriesHolder
	PendingMiniBlocks           map[string]*block.MiniBlock
}

// epochStartBootstrap will handle requesting the needed data to start when joining late the network
type epochStartBootstrap struct {
	publicKey                      crypto.PublicKey
	marshalizer                    marshal.Marshalizer
	hasher                         hashing.Hasher
	messenger                      p2p.Messenger
	generalConfig                  config.Config
	economicsConfig                config.EconomicsConfig
	pathManager                    PathManagerHandler
	nodesConfigProvider            NodesConfigProviderHandler
	epochStartMetaBlockInterceptor EpochStartMetaBlockInterceptorHandler
	metaBlockInterceptor           MetaBlockInterceptorHandler
	shardHeaderInterceptor         ShardHeaderInterceptorHandler
	miniBlockInterceptor           MiniBlockInterceptorHandler
	singleSigner                   crypto.SingleSigner
	blockSingleSigner              crypto.SingleSigner
	keyGen                         crypto.KeyGenerator
	blockKeyGen                    crypto.KeyGenerator
	requestHandler                 process.RequestHandler
	whiteListHandler               dataRetriever.WhiteListHandler
	shardCoordinator               sharding.Coordinator
	genesisNodesConfig             *sharding.NodesSetup
	workingDir                     string
	defaultDBPath                  string
	defaultEpochString             string

	dataPool      dataRetriever.PoolsHolder
	computedEpoch uint32
}

// ArgsEpochStartBootstrap holds the arguments needed for creating an epoch start data provider component
type ArgsEpochStartBootstrap struct {
	PublicKey                      crypto.PublicKey
	Messenger                      p2p.Messenger
	Marshalizer                    marshal.Marshalizer
	Hasher                         hashing.Hasher
	GeneralConfig                  config.Config
	EconomicsConfig                config.EconomicsConfig
	GenesisShardCoordinator        sharding.Coordinator
	PathManager                    PathManagerHandler
	NodesConfigProvider            NodesConfigProviderHandler
	EpochStartMetaBlockInterceptor EpochStartMetaBlockInterceptorHandler
	SingleSigner                   crypto.SingleSigner
	BlockSingleSigner              crypto.SingleSigner
	KeyGen                         crypto.KeyGenerator
	BlockKeyGen                    crypto.KeyGenerator
	WhiteListHandler               dataRetriever.WhiteListHandler
	GenesisNodesConfig             *sharding.NodesSetup
	WorkingDir                     string
	DefaultDBPath                  string
	DefaultEpochString             string
}

// NewEpochStartBootstrap will return a new instance of epochStartBootstrap
func NewEpochStartBootstrapHandler(args ArgsEpochStartBootstrap) (*epochStartBootstrap, error) {
	if check.IfNil(args.PublicKey) {
		return nil, ErrNilPublicKey
	}
	if check.IfNil(args.Messenger) {
		return nil, ErrNilMessenger
	}
	if check.IfNil(args.Marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(args.PathManager) {
		return nil, ErrNilPathManager
	}
	if check.IfNil(args.NodesConfigProvider) {
		return nil, ErrNilNodesConfigProvider
	}
	if check.IfNil(args.EpochStartMetaBlockInterceptor) {
		return nil, ErrNilEpochStartMetaBlockInterceptor
	}
	if check.IfNil(args.WhiteListHandler) {
		return nil, ErrNilWhiteListHandler
	}
	if check.IfNil(args.DefaultShardCoordinator) {
		return nil, ErrNilDefaultShardCoordinator
	}
	if check.IfNil(args.BlockKeyGen) {
		return nil, ErrNilBlockKeyGen
	}
	if check.IfNil(args.KeyGen) {
		return nil, ErrNilKeyGen
	}
	if check.IfNil(args.SingleSigner) {
		return nil, ErrNilSingleSigner
	}
	if check.IfNil(args.BlockSingleSigner) {
		return nil, ErrNilBlockSingleSigner
	}

	epochStartProvider := &epochStartBootstrap{
		publicKey:                      args.PublicKey,
		marshalizer:                    args.Marshalizer,
		hasher:                         args.Hasher,
		messenger:                      args.Messenger,
		generalConfig:                  args.GeneralConfig,
		economicsConfig:                args.EconomicsConfig,
		pathManager:                    args.PathManager,
		nodesConfigProvider:            args.NodesConfigProvider,
		epochStartMetaBlockInterceptor: args.EpochStartMetaBlockInterceptor,
		metaBlockInterceptor:           args.MetaBlockInterceptor,
		shardHeaderInterceptor:         args.ShardHeaderInterceptor,
		miniBlockInterceptor:           args.MiniBlockInterceptor,
		whiteListHandler:               args.WhiteListHandler,
		genesisNodesConfig:             args.GenesisNodesConfig,
		workingDir:                     args.WorkingDir,
		defaultEpochString:             args.DefaultEpochString,
		defaultDBPath:                  args.DefaultEpochString,
		keyGen:                         args.KeyGen,
		blockKeyGen:                    args.BlockKeyGen,
		singleSigner:                   args.SingleSigner,
		blockSingleSigner:              args.BlockSingleSigner,
	}

	return epochStartProvider, nil
}

func (e *epochStartBootstrap) initInternalComponents() error {
	var err error
	e.shardCoordinator, err = sharding.NewMultiShardCoordinator(e.genesisNodesConfig.NumberOfShards(), core.MetachainShardId)
	if err != nil {
		return err
	}

	err = e.initTopicsAndInterceptors()
	if err != nil {
		return err
	}
	defer func() {
		e.resetTopicsAndInterceptors()
	}()

	err = e.createRequestHandler()
	if err != nil {
		return err
	}

	e.dataPool, err = factory2.NewDataPoolFromConfig(
		factory2.ArgsDataPool{
			Config:           &e.generalConfig,
			EconomicsData:    e.economicsData,
			ShardCoordinator: e.shardCoordinator,
		},
	)

	return nil
}

func (e *epochStartBootstrap) searchDataInLocalStorage() {
	currentEpoch, errNotCritical := storageFactory.FindLastEpochFromStorage(
		e.workingDir,
		e.genesisNodesConfig.ChainID,
		e.defaultDBPath,
		e.defaultEpochString,
	)
	if errNotCritical != nil {
		log.Debug("no epoch db found in storage", "error", errNotCritical.Error())
	}

	log.Debug("current epoch from the storage : ", "epoch", currentEpoch)
}

func (e *epochStartBootstrap) isStartInEpochZero() bool {
	startTime := time.Unix(e.genesisNodesConfig.StartTime, 0)
	isCurrentTimeBeforeGenesis := time.Now().Sub(startTime) < 0
	if isCurrentTimeBeforeGenesis {
		return true
	}

	timeInFirstEpochAtRoundsPerEpoch := startTime.Add(time.Duration(e.genesisNodesConfig.RoundDuration *
		uint64(e.generalConfig.EpochStartConfig.RoundsPerEpoch)))
	isEpochZero := time.Now().Sub(timeInFirstEpochAtRoundsPerEpoch) < 0

	return isEpochZero
}

func (e *epochStartBootstrap) prepareEpochZero() (uint32, uint32, uint32, error) {
	currentEpoch := uint32(0)
	return currentEpoch, e.shardCoordinator.SelfId(), e.shardCoordinator.NumberOfShards(), nil
}

func (e *epochStartBootstrap) isCurrentEpochSavedInStorage() bool {
	// TODO: implement
	return true
}

func (e *epochStartBootstrap) requestDataFromNetwork() {

}

func (e *epochStartBootstrap) saveGatheredDataToStorage() {

}

func (e *epochStartBootstrap) computeMostProbableEpoch() {
	startTime := time.Unix(e.genesisNodesConfig.StartTime, 0)
	elapsedTime := time.Since(startTime)

	timeForOneEpoch := time.Duration(e.genesisNodesConfig.RoundDuration *
		uint64(e.generalConfig.EpochStartConfig.RoundsPerEpoch))

	elaspedTimeInSeconds := uint64(elapsedTime.Seconds())
	timeForOneEpochInSeconds := uint64(timeForOneEpoch.Seconds())

	e.computedEpoch = uint32(elaspedTimeInSeconds / timeForOneEpochInSeconds)
}

func (e *epochStartBootstrap) Bootstrap() (uint32, uint32, uint32, error) {
	if e.isStartInEpochZero() {
		return e.prepareEpochZero()
	}

}

// Bootstrap will handle requesting and receiving the needed information the node will bootstrap from
func (e *epochStartBootstrap) requestAndProcessing() (uint32, uint32, uint32, error) {
	// TODO: add searching for epoch start metablock and other data inside this component

	err := e.initInternalComponents()
	if err != nil {
		return nil, err
	}

	interceptorsContainer, err := e.createInterceptors(commonDataPool)
	if err != nil || interceptorsContainer == nil {
		return nil, err
	}

	miniBlocksSyncer, err := e.getMiniBlockSyncer(commonDataPool.MiniBlocks())
	if err != nil {
		return nil, err
	}

	missingHeadersSyncer, err := e.getHeaderHandlerSyncer(commonDataPool.Headers())
	if err != nil {
		return nil, err
	}

	epochNumForRequestingTheLatestAvailable := uint32(math.MaxUint32)
	metaBlock, err := e.getEpochStartMetaBlock(epochNumForRequestingTheLatestAvailable)
	if err != nil {
		return nil, err
	}

	prevMetaBlock, err := e.getMetaBlock(missingHeadersSyncer, metaBlock.EpochStart.Economics.PrevEpochStartHash)
	if err != nil {
		return nil, err
	}

	e.changeMessageProcessorsForMetaBlocks()

	log.Info("previous meta block", "epoch", prevMetaBlock.Epoch)
	nodesConfig, err := e.nodesConfigProvider.GetNodesConfigForMetaBlock(metaBlock)
	if err != nil {
		return nil, err
	}

	e.shardCoordinator, err = e.getShardCoordinator(metaBlock, nodesConfig)
	if err != nil {
		return nil, err
	}

	shardHeaders, err := e.getShardHeaders(missingHeadersSyncer, metaBlock, nodesConfig, shardCoordinator)
	if err != nil {
		log.Debug("shard headers not found", "error", err)
	}

	var shardHeaderForShard *block.Header
	if e.shardCoordinator.SelfId() < e.shardCoordinator.NumberOfShards() {
		shardHeaderForShard = shardHeaders[e.shardCoordinator.SelfId()]
	}

	epochStartData, err := e.getCurrentEpochStartData(e.shardCoordinator, metaBlock)
	if err != nil {
		return nil, err
	}

	pendingMiniBlocks, err := e.getMiniBlocks(miniBlocksSyncer, epochStartData.PendingMiniBlockHeaders, shardCoordinator.SelfId())
	if err != nil {
		return nil, err
	}

	lastFinalizedMetaBlock, err := e.getMetaBlock(missingHeadersSyncer, epochStartData.LastFinishedMetaBlock)
	if err != nil {
		return nil, err
	}
	log.Info("received last finalized meta block", "nonce", lastFinalizedMetaBlock.Nonce)

	firstPendingMetaBlock, err := e.getMetaBlock(missingHeadersSyncer, epochStartData.FirstPendingMetaBlock)
	if err != nil {
		return nil, err
	}
	log.Info("received first pending meta block", "nonce", firstPendingMetaBlock.Nonce)

	trieToReturn, err := e.getTrieFromRootHash(epochStartData.RootHash)
	if err != nil {
		return nil, err
	}

	components := &structs.ComponentsNeededForBootstrap{
		EpochStartMetaBlock:         metaBlock,
		PreviousEpochStartMetaBlock: prevMetaBlock,
		ShardHeader:                 shardHeaderForShard,
		NodesConfig:                 nodesConfig,
		ShardHeaders:                shardHeaders,
		ShardCoordinator:            e.shardCoordinator,
		Tries:                       trieToReturn,
		PendingMiniBlocks:           pendingMiniBlocks,
	}

	var storageHandlerComponent StorageHandler
	if e.shardCoordinator.SelfId() > e.shardCoordinator.NumberOfShards() {
		storageHandlerComponent, err = storagehandler.NewMetaStorageHandler(
			e.generalConfig,
			e.shardCoordinator,
			e.pathManager,
			e.marshalizer,
			e.hasher,
			metaBlock.Epoch,
		)
		if err != nil {
			return nil, err
		}
	} else {
		storageHandlerComponent, err = storagehandler.NewShardStorageHandler(
			e.generalConfig,
			e.shardCoordinator,
			e.pathManager,
			e.marshalizer,
			e.hasher,
			metaBlock.Epoch,
		)
		if err != nil {
			return nil, err
		}
	}

	errSavingToStorage := storageHandlerComponent.SaveDataToStorage(*components)
	if errSavingToStorage != nil {
		return nil, errSavingToStorage
	}

	return components, nil
}

func (e *epochStartBootstrap) getMiniBlockSyncer(dataPool storage.Cacher) (update.EpochStartPendingMiniBlocksSyncHandler, error) {
	syncMiniBlocksArgs := sync.ArgsNewPendingMiniBlocksSyncer{
		Storage:        &disabled.Storer{},
		Cache:          dataPool,
		Marshalizer:    e.marshalizer,
		RequestHandler: e.requestHandler,
	}
	return sync.NewPendingMiniBlocksSyncer(syncMiniBlocksArgs)
}

func (e *epochStartBootstrap) getHeaderHandlerSyncer(pool dataRetriever.HeadersPool) (update.MissingHeadersByHashSyncer, error) {
	syncMissingHeadersArgs := sync.ArgsNewMissingHeadersByHashSyncer{
		Storage:        &disabled.Storer{},
		Cache:          pool,
		Marshalizer:    e.marshalizer,
		RequestHandler: e.requestHandler,
	}
	return sync.NewMissingheadersByHashSyncer(syncMissingHeadersArgs)
}

func (e *epochStartBootstrap) getMiniBlocks(
	handler update.EpochStartPendingMiniBlocksSyncHandler,
	pendingMiniBlocks []block.ShardMiniBlockHeader,
	shardID uint32,
) (map[string]*block.MiniBlock, error) {

	waitTime := 1 * time.Minute
	err := handler.SyncPendingMiniBlocksForEpochStart(pendingMiniBlocks, waitTime)
	if err != nil {
		return nil, err
	}

	return handler.GetMiniBlocks()
}

func (e *epochStartBootstrap) createInterceptors(dataPool dataRetriever.PoolsHolder) (process.InterceptorsContainer, error) {
	args := factory3.ArgsEpochStartInterceptorContainer{
		Config:            e.generalConfig,
		ShardCoordinator:  e.defaultShardCoordinator,
		Marshalizer:       e.marshalizer,
		Hasher:            e.hasher,
		Messenger:         e.messenger,
		DataPool:          dataPool,
		SingleSigner:      e.singleSigner,
		BlockSingleSigner: e.blockSingleSigner,
		KeyGen:            e.keyGen,
		BlockKeyGen:       e.blockKeyGen,
		WhiteListHandler:  e.whiteListHandler,
	}

	return factory3.NewEpochStartInterceptorsContainer(args)
}

func (e *epochStartBootstrap) changeMessageProcessorsForMetaBlocks() {
	err := e.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
	if err != nil {
		log.Info("error unregistering message processor", "error", err)
	}

	err = e.messenger.RegisterMessageProcessor(factory.MetachainBlocksTopic, e.metaBlockInterceptor)
	if err != nil {
		log.Info("error unregistering message processor", "error", err)
	}
}

func (e *epochStartBootstrap) createRequestHandler() error {
	dataPacker, err := partitioning.NewSimpleDataPacker(e.marshalizer)
	if err != nil {
		return err
	}

	storageService := &disabled.ChainStorer{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return disabled.NewDisabledStorer()
		},
	}

	cacher := disabled.NewDisabledPoolsHolder()
	triesHolder := state.NewDataTriesHolder()

	stateTrieStorageManager, err := trie.NewTrieStorageManagerWithoutPruning(disabled.NewDisabledStorer())
	if err != nil {
		return err
	}
	stateTrie, err := trie.NewTrie(stateTrieStorageManager, e.marshalizer, e.hasher)
	if err != nil {
		return err
	}
	triesHolder.Put([]byte(trieFactory.UserAccountTrie), stateTrie)

	peerTrieStorageManager, err := trie.NewTrieStorageManagerWithoutPruning(disabled.NewDisabledStorer())
	if err != nil {
		return err
	}

	peerTrie, err := trie.NewTrie(peerTrieStorageManager, e.marshalizer, e.hasher)
	if err != nil {
		return err
	}
	triesHolder.Put([]byte(trieFactory.PeerAccountTrie), peerTrie)

	resolversContainerArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:         e.shardCoordinator,
		Messenger:                e.messenger,
		Store:                    storageService,
		Marshalizer:              e.marshalizer,
		DataPools:                cacher,
		Uint64ByteSliceConverter: uint64ByteSlice.NewBigEndianConverter(),
		DataPacker:               dataPacker,
		TriesContainer:           triesHolder,
		SizeCheckDelta:           0,
	}

	resolverFactory, err := resolverscontainer.NewMetaResolversContainerFactory(resolversContainerArgs)
	if err != nil {
		return err
	}

	container, err := resolverFactory.Create()
	if err != nil {
		return err
	}

	finder, err := containers.NewResolversFinder(container, e.shardCoordinator)
	if err != nil {
		return err
	}

	requestedItemsHandler := timecache.NewTimeCache(100)

	maxToRequest := 100

	e.requestHandler, err = requestHandlers.NewResolverRequestHandler(finder, requestedItemsHandler, e.whiteListHandler, maxToRequest, core.MetachainShardId)
	return err
}

func (e *epochStartBootstrap) getMiniBlock(miniBlockHeader *block.ShardMiniBlockHeader) (*block.MiniBlock, error) {
	e.requestMiniBlock(miniBlockHeader)

	time.Sleep(delayAfterRequesting)

	for {
		numConnectedPeers := len(e.messenger.Peers())
		threshold := int(thresholdForConsideringMetaBlockCorrect * float64(numConnectedPeers))
		mb, errConsensusNotReached := e.miniBlockInterceptor.GetMiniBlock(miniBlockHeader.Hash, threshold)
		if errConsensusNotReached == nil {
			return mb, nil
		}
		log.Info("consensus not reached for epoch start meta block. re-requesting and trying again...")
		e.requestMiniBlock(miniBlockHeader)
	}
}

func (e *epochStartBootstrap) requestMiniBlock(miniBlockHeader *block.ShardMiniBlockHeader) {
	e.requestHandler.RequestMiniBlock(miniBlockHeader.ReceiverShardID, miniBlockHeader.Hash)
}

func (e *epochStartBootstrap) getCurrentEpochStartData(
	shardCoordinator sharding.epochStartBootstrap,
	metaBlock *block.MetaBlock,
) (*block.EpochStartShardData, error) {
	shardID := shardCoordinator.SelfId()
	for _, epochStartData := range metaBlock.EpochStart.LastFinalizedHeaders {
		if epochStartData.ShardID == shardID {
			return &epochStartData, nil
		}
	}

	return nil, errors.New("not found")
}

func (e *epochStartBootstrap) initTopicForEpochStartMetaBlockInterceptor() error {
	err := e.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
	if err != nil {
		log.Info("error unregistering message processor", "error", err)
		return err
	}

	err = e.messenger.CreateTopic(factory.MetachainBlocksTopic, true)
	if err != nil {
		log.Info("error registering message processor", "error", err)
		return err
	}

	err = e.messenger.RegisterMessageProcessor(factory.MetachainBlocksTopic, e.epochStartMetaBlockInterceptor)
	if err != nil {
		return err
	}

	return nil
}

func (e *epochStartBootstrap) getShardID(nodesConfig *sharding.NodesSetup) (uint32, error) {
	pubKeyBytes, err := e.publicKey.ToByteArray()
	if err != nil {
		return 0, err
	}
	pubKeyStr := hex.EncodeToString(pubKeyBytes)
	for shardID, nodesPerShard := range nodesConfig.InitialNodesPubKeys() {
		for _, nodePubKey := range nodesPerShard {
			if nodePubKey == pubKeyStr {
				return shardID, nil
			}
		}
	}

	return 0, nil
}

func (e *epochStartBootstrap) getTrieFromRootHash(_ []byte) (state.TriesHolder, error) {
	// TODO: get trie from trie syncer
	return state.NewDataTriesHolder(), nil
}

func (e *epochStartBootstrap) resetTopicsAndInterceptors() {
	err := e.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
	if err != nil {
		log.Info("error unregistering message processors", "error", err)
	}
}

func (e *epochStartBootstrap) getMetaBlock(syncer update.MissingHeadersByHashSyncer, hash []byte) (*block.MetaBlock, error) {
	//e.requestMetaBlock(hash)
	//
	//time.Sleep(delayAfterRequesting)
	//
	//for {
	//	numConnectedPeers := len(e.messenger.Peers())
	//	threshold := int(thresholdForConsideringMetaBlockCorrect * float64(numConnectedPeers))
	//	mb, errConsensusNotReached := e.metaBlockInterceptor.GetMetaBlock(hash, threshold)
	//	if errConsensusNotReached == nil {
	//		return mb, nil
	//	}
	//	log.Info("consensus not reached for meta block. re-requesting and trying again...")
	//	e.requestMetaBlock(hash)
	//}
	waitTime := 1 * time.Minute
	err := syncer.SyncMissingHeadersByHash(e.defaultShardCoordinator.SelfId(), [][]byte{hash}, waitTime)
	if err != nil {
		return nil, err
	}

	hdrs, err := syncer.GetHeaders()
	if err != nil {
		return nil, err
	}

	syncer.ClearFields()

	return hdrs[string(hash)].(*block.MetaBlock), nil
}

func (e *epochStartBootstrap) getEpochStartMetaBlock(epoch uint32) (*block.MetaBlock, error) {
	err := e.initTopicForEpochStartMetaBlockInterceptor()
	if err != nil {
		return nil, err
	}
	defer func() {
		e.resetTopicsAndInterceptors()
	}()

	e.requestEpochStartMetaBlock(epoch)

	time.Sleep(delayAfterRequesting)
	count := 0

	for {
		if count > maxNumTimesToRetry {
			panic("can't sync with other peers")
		}
		count++
		numConnectedPeers := len(e.messenger.Peers())
		threshold := int(thresholdForConsideringMetaBlockCorrect * float64(numConnectedPeers))
		mb, errConsensusNotReached := e.epochStartMetaBlockInterceptor.GetEpochStartMetaBlock(threshold, epoch)
		if errConsensusNotReached == nil {
			return mb, nil
		}
		log.Info("consensus not reached for meta block. re-requesting and trying again...")
		e.requestEpochStartMetaBlock(epoch)
	}
}

func (e *epochStartBootstrap) getShardCoordinator(metaBlock *block.MetaBlock, nodesConfig *sharding.NodesSetup) (sharding.epochStartBootstrap, error) {
	shardID, err := e.getShardID(nodesConfig)
	if err != nil {
		return nil, err
	}

	numOfShards := len(metaBlock.EpochStart.LastFinalizedHeaders)
	return sharding.NewMultiShardCoordinator(uint32(numOfShards), shardID)
}

func (e *epochStartBootstrap) getShardHeaders(
	syncer update.MissingHeadersByHashSyncer,
	metaBlock *block.MetaBlock,
	shardCoordinator sharding.epochStartBootstrap,
) (map[uint32]*block.Header, error) {
	headersMap := make(map[uint32]*block.Header)

	shardID := shardCoordinator.SelfId()
	if shardID == core.MetachainShardId {
		for _, entry := range metaBlock.EpochStart.LastFinalizedHeaders {
			var hdr *block.Header
			hdr, err := e.getShardHeader(syncer, entry.HeaderHash, entry.ShardID)
			if err != nil {
				return nil, err
			}
			headersMap[entry.ShardID] = hdr
		}

		return headersMap, nil
	}

	var entryForShard *block.EpochStartShardData
	for _, entry := range metaBlock.EpochStart.LastFinalizedHeaders {
		if entry.ShardID == shardID {
			entryForShard = &entry
		}
	}

	if entryForShard == nil {
		return nil, ErrShardDataNotFound
	}

	hdr, err := e.getShardHeader(
		syncer,
		entryForShard.HeaderHash,
		entryForShard.ShardID,
	)
	if err != nil {
		return nil, err
	}

	headersMap[shardID] = hdr
	return headersMap, nil
}

func (e *epochStartBootstrap) getShardHeader(
	syncer update.MissingHeadersByHashSyncer,
	hash []byte,
	shardID uint32,
) (*block.Header, error) {
	waitTime := 1 * time.Minute
	err := syncer.SyncMissingHeadersByHash(e.defaultShardCoordinator.SelfId(), [][]byte{hash}, waitTime)
	if err != nil {
		return nil, err
	}

	hdrs, err := syncer.GetHeaders()
	if err != nil {
		return nil, err
	}

	syncer.ClearFields()

	return hdrs[string(hash)].(*block.Header), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (e *epochStartBootstrap) IsInterfaceNil() bool {
	return e == nil
}
