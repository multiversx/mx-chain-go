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
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/factory/interceptors"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/storagehandler"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/structs"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/sync"
)

var log = logger.GetOrCreate("registration")
var _ process.Interceptor = (*simpleMetaBlockInterceptor)(nil)

const delayBetweenRequests = 1 * time.Second
const delayAfterRequesting = 1 * time.Second
const thresholdForConsideringMetaBlockCorrect = 0.2
const numRequestsToSendOnce = 4
const maxNumTimesToRetry = 100

// epochStartDataProvider will handle requesting the needed data to start when joining late the network
type epochStartDataProvider struct {
	publicKey                      crypto.PublicKey
	marshalizer                    marshal.Marshalizer
	hasher                         hashing.Hasher
	messenger                      p2p.Messenger
	generalConfig                  config.Config
	economicsConfig                config.EconomicsConfig
	defaultShardCoordinator        sharding.Coordinator
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
}

// ArgsEpochStartDataProvider holds the arguments needed for creating an epoch start data provider component
type ArgsEpochStartDataProvider struct {
	PublicKey                      crypto.PublicKey
	Messenger                      p2p.Messenger
	Marshalizer                    marshal.Marshalizer
	Hasher                         hashing.Hasher
	GeneralConfig                  config.Config
	EconomicsConfig                config.EconomicsConfig
	DefaultShardCoordinator        sharding.Coordinator
	PathManager                    PathManagerHandler
	NodesConfigProvider            NodesConfigProviderHandler
	EpochStartMetaBlockInterceptor EpochStartMetaBlockInterceptorHandler
	MetaBlockInterceptor           MetaBlockInterceptorHandler
	ShardHeaderInterceptor         ShardHeaderInterceptorHandler
	MiniBlockInterceptor           MiniBlockInterceptorHandler
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

// NewEpochStartDataProvider will return a new instance of epochStartDataProvider
func NewEpochStartDataProvider(args ArgsEpochStartDataProvider) (*epochStartDataProvider, error) {
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
	if check.IfNil(args.MetaBlockInterceptor) {
		return nil, ErrNilMetaBlockInterceptor
	}
	if check.IfNil(args.ShardHeaderInterceptor) {
		return nil, ErrNilShardHeaderInterceptor
	}
	if check.IfNil(args.MiniBlockInterceptor) {
		return nil, ErrNilMiniBlockInterceptor
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

	epochStartProvider := &epochStartDataProvider{
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
		defaultShardCoordinator:        args.DefaultShardCoordinator,
		keyGen:                         args.KeyGen,
		blockKeyGen:                    args.BlockKeyGen,
		singleSigner:                   args.SingleSigner,
		blockSingleSigner:              args.BlockSingleSigner,
	}

	err := epochStartProvider.initInternalComponents()
	if err != nil {
		return nil, err
	}

	return epochStartProvider, nil
}

func (esdp *epochStartDataProvider) initInternalComponents() error {
	var err error
	esdp.shardCoordinator, err = sharding.NewMultiShardCoordinator(esdp.genesisNodesConfig.NumberOfShards(), core.MetachainShardId)
	if err != nil {
		return err
	}

	err = esdp.initTopicsAndInterceptors()
	if err != nil {
		return err
	}
	defer func() {
		esdp.resetTopicsAndInterceptors()
	}()

	err = esdp.createRequestHandler()
	if err != nil {
		return err
	}

	economicsData, err := economics.NewEconomicsData(&esdp.economicsConfig)
	if err != nil {
		return nil, err
	}

	commonDataPool, err := factory2.NewDataPoolFromConfig(
		factory2.ArgsDataPool{
			Config:           &esdp.generalConfig,
			EconomicsData:    economicsData,
			ShardCoordinator: esdp.defaultShardCoordinator,
		},
	)

	return nil
}

func (esdp *epochStartDataProvider) searchDataInLocalStorage() {
	var errNotCritical error
	// TODO: add a component which opens headers storer and gets the last epoch start metablock
	// in order to provide the last known epoch in storage. Right now, it won't work as expected
	// if storage pruning is disabled
	isEpochFoundInStorage := true
	currentEpoch, errNotCritical := storageFactory.FindLastEpochFromStorage(
		esdp.workingDir,
		esdp.genesisNodesConfig.ChainID,
		esdp.defaultDBPath,
		esdp.defaultEpochString,
	)
	if errNotCritical != nil {
		log.Debug("no epoch db found in storage", "error", errNotCritical.Error())
		isEpochFoundInStorage = false
	}

	log.Debug("current epoch from the storage : ", "epoch", currentEpoch)

	shouldSync := ShouldSyncWithTheNetwork(
		args.StartTime,
		isEpochFoundInStorage,
		args.OriginalNodesConfig,
		args.GeneralConfig,
	)

	log.Debug("shouldSync epochStartData", "shouldSync", shouldSync)
}

// Bootstrap will handle requesting and receiving the needed information the node will bootstrap from
func (esdp *epochStartDataProvider) Bootstrap() (*structs.ComponentsNeededForBootstrap, error) {
	// TODO: add searching for epoch start metablock and other data inside this component

	interceptorsContainer, err := esdp.createInterceptors(commonDataPool)
	if err != nil || interceptorsContainer == nil {
		return nil, err
	}

	miniBlocksSyncer, err := esdp.getMiniBlockSyncer(commonDataPool.MiniBlocks())
	if err != nil {
		return nil, err
	}

	missingHeadersSyncer, err := esdp.getHeaderHandlerSyncer(commonDataPool.Headers())
	if err != nil {
		return nil, err
	}

	epochNumForRequestingTheLatestAvailable := uint32(math.MaxUint32)
	metaBlock, err := esdp.getEpochStartMetaBlock(epochNumForRequestingTheLatestAvailable)
	if err != nil {
		return nil, err
	}

	prevMetaBlock, err := esdp.getMetaBlock(missingHeadersSyncer, metaBlock.EpochStart.Economics.PrevEpochStartHash)
	if err != nil {
		return nil, err
	}

	esdp.changeMessageProcessorsForMetaBlocks()

	log.Info("previous meta block", "epoch", prevMetaBlock.Epoch)
	nodesConfig, err := esdp.nodesConfigProvider.GetNodesConfigForMetaBlock(metaBlock)
	if err != nil {
		return nil, err
	}

	esdp.shardCoordinator, err = esdp.getShardCoordinator(metaBlock, nodesConfig)
	if err != nil {
		return nil, err
	}

	shardHeaders, err := esdp.getShardHeaders(missingHeadersSyncer, metaBlock, nodesConfig, shardCoordinator)
	if err != nil {
		log.Debug("shard headers not found", "error", err)
	}

	var shardHeaderForShard *block.Header
	if esdp.shardCoordinator.SelfId() < esdp.shardCoordinator.NumberOfShards() {
		shardHeaderForShard = shardHeaders[esdp.shardCoordinator.SelfId()]
	}

	epochStartData, err := esdp.getCurrentEpochStartData(esdp.shardCoordinator, metaBlock)
	if err != nil {
		return nil, err
	}

	pendingMiniBlocks, err := esdp.getMiniBlocks(miniBlocksSyncer, epochStartData.PendingMiniBlockHeaders, shardCoordinator.SelfId())
	if err != nil {
		return nil, err
	}
	//pendingMiniBlocks := make([]*block.MiniBlock, 0)
	//for _, mb := range epochStartData.PendingMiniBlockHeaders {
	//	receivedMb, errGetMb := esdp.getMiniBlock(&mb)
	//	if errGetMb != nil {
	//		return nil, errGetMb
	//	}
	//	pendingMiniBlocks = append(pendingMiniBlocks, receivedMb)
	//	log.Info("received miniblock", "type", receivedMb.Type)
	//}

	lastFinalizedMetaBlock, err := esdp.getMetaBlock(missingHeadersSyncer, epochStartData.LastFinishedMetaBlock)
	if err != nil {
		return nil, err
	}
	log.Info("received last finalized meta block", "nonce", lastFinalizedMetaBlock.Nonce)

	firstPendingMetaBlock, err := esdp.getMetaBlock(missingHeadersSyncer, epochStartData.FirstPendingMetaBlock)
	if err != nil {
		return nil, err
	}
	log.Info("received first pending meta block", "nonce", firstPendingMetaBlock.Nonce)

	trieToReturn, err := esdp.getTrieFromRootHash(epochStartData.RootHash)
	if err != nil {
		return nil, err
	}

	components := &structs.ComponentsNeededForBootstrap{
		EpochStartMetaBlock:         metaBlock,
		PreviousEpochStartMetaBlock: prevMetaBlock,
		ShardHeader:                 shardHeaderForShard,
		NodesConfig:                 nodesConfig,
		ShardHeaders:                shardHeaders,
		ShardCoordinator:            esdp.shardCoordinator,
		Tries:                       trieToReturn,
		PendingMiniBlocks:           pendingMiniBlocks,
	}

	var storageHandlerComponent StorageHandler
	if esdp.shardCoordinator.SelfId() > esdp.shardCoordinator.NumberOfShards() {
		storageHandlerComponent, err = storagehandler.NewMetaStorageHandler(
			esdp.generalConfig,
			esdp.shardCoordinator,
			esdp.pathManager,
			esdp.marshalizer,
			esdp.hasher,
			metaBlock.Epoch,
		)
		if err != nil {
			return nil, err
		}
	} else {
		storageHandlerComponent, err = storagehandler.NewShardStorageHandler(
			esdp.generalConfig,
			esdp.shardCoordinator,
			esdp.pathManager,
			esdp.marshalizer,
			esdp.hasher,
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

func (esdp *epochStartDataProvider) getMiniBlockSyncer(dataPool storage.Cacher) (update.EpochStartPendingMiniBlocksSyncHandler, error) {
	syncMiniBlocksArgs := sync.ArgsNewPendingMiniBlocksSyncer{
		Storage:        &disabled.Storer{},
		Cache:          dataPool,
		Marshalizer:    esdp.marshalizer,
		RequestHandler: esdp.requestHandler,
	}
	return sync.NewPendingMiniBlocksSyncer(syncMiniBlocksArgs)
}

func (esdp *epochStartDataProvider) getHeaderHandlerSyncer(pool dataRetriever.HeadersPool) (update.MissingHeadersByHashSyncer, error) {
	syncMissingHeadersArgs := sync.ArgsNewMissingHeadersByHashSyncer{
		Storage:        &disabled.Storer{},
		Cache:          pool,
		Marshalizer:    esdp.marshalizer,
		RequestHandler: esdp.requestHandler,
	}
	return sync.NewMissingheadersByHashSyncer(syncMissingHeadersArgs)
}

func (esdp *epochStartDataProvider) getMiniBlocks(
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

func (esdp *epochStartDataProvider) createInterceptors(dataPool dataRetriever.PoolsHolder) (process.InterceptorsContainer, error) {
	args := interceptors.ArgsEpochStartInterceptorContainer{
		Config:            esdp.generalConfig,
		ShardCoordinator:  esdp.defaultShardCoordinator,
		Marshalizer:       esdp.marshalizer,
		Hasher:            esdp.hasher,
		Messenger:         esdp.messenger,
		DataPool:          dataPool,
		SingleSigner:      esdp.singleSigner,
		BlockSingleSigner: esdp.blockSingleSigner,
		KeyGen:            esdp.keyGen,
		BlockKeyGen:       esdp.blockKeyGen,
		WhiteListHandler:  esdp.whiteListHandler,
	}

	return interceptors.NewEpochStartInterceptorsContainer(args)
}

func (esdp *epochStartDataProvider) changeMessageProcessorsForMetaBlocks() {
	err := esdp.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
	if err != nil {
		log.Info("error unregistering message processor", "error", err)
	}

	err = esdp.messenger.RegisterMessageProcessor(factory.MetachainBlocksTopic, esdp.metaBlockInterceptor)
	if err != nil {
		log.Info("error unregistering message processor", "error", err)
	}
}

func (esdp *epochStartDataProvider) createRequestHandler() error {
	dataPacker, err := partitioning.NewSimpleDataPacker(esdp.marshalizer)
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
	stateTrie, err := trie.NewTrie(stateTrieStorageManager, esdp.marshalizer, esdp.hasher)
	if err != nil {
		return err
	}
	triesHolder.Put([]byte(trieFactory.UserAccountTrie), stateTrie)

	peerTrieStorageManager, err := trie.NewTrieStorageManagerWithoutPruning(disabled.NewDisabledStorer())
	if err != nil {
		return err
	}

	peerTrie, err := trie.NewTrie(peerTrieStorageManager, esdp.marshalizer, esdp.hasher)
	if err != nil {
		return err
	}
	triesHolder.Put([]byte(trieFactory.PeerAccountTrie), peerTrie)

	resolversContainerArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:         esdp.shardCoordinator,
		Messenger:                esdp.messenger,
		Store:                    storageService,
		Marshalizer:              esdp.marshalizer,
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

	finder, err := containers.NewResolversFinder(container, esdp.shardCoordinator)
	if err != nil {
		return err
	}

	requestedItemsHandler := timecache.NewTimeCache(100)

	maxToRequest := 100

	esdp.requestHandler, err = requestHandlers.NewResolverRequestHandler(finder, requestedItemsHandler, esdp.whiteListHandler, maxToRequest, core.MetachainShardId)
	return err
}

func (esdp *epochStartDataProvider) getMiniBlock(miniBlockHeader *block.ShardMiniBlockHeader) (*block.MiniBlock, error) {
	esdp.requestMiniBlock(miniBlockHeader)

	time.Sleep(delayAfterRequesting)

	for {
		numConnectedPeers := len(esdp.messenger.Peers())
		threshold := int(thresholdForConsideringMetaBlockCorrect * float64(numConnectedPeers))
		mb, errConsensusNotReached := esdp.miniBlockInterceptor.GetMiniBlock(miniBlockHeader.Hash, threshold)
		if errConsensusNotReached == nil {
			return mb, nil
		}
		log.Info("consensus not reached for epoch start meta block. re-requesting and trying again...")
		esdp.requestMiniBlock(miniBlockHeader)
	}
}

func (esdp *epochStartDataProvider) requestMiniBlock(miniBlockHeader *block.ShardMiniBlockHeader) {
	esdp.requestHandler.RequestMiniBlock(miniBlockHeader.ReceiverShardID, miniBlockHeader.Hash)
}

func (esdp *epochStartDataProvider) getCurrentEpochStartData(
	shardCoordinator sharding.Coordinator,
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

func (esdp *epochStartDataProvider) initTopicForEpochStartMetaBlockInterceptor() error {
	err := esdp.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
	if err != nil {
		log.Info("error unregistering message processor", "error", err)
		return err
	}

	err = esdp.messenger.CreateTopic(factory.MetachainBlocksTopic, true)
	if err != nil {
		log.Info("error registering message processor", "error", err)
		return err
	}

	err = esdp.messenger.RegisterMessageProcessor(factory.MetachainBlocksTopic, esdp.epochStartMetaBlockInterceptor)
	if err != nil {
		return err
	}

	return nil
}

func (esdp *epochStartDataProvider) getShardID(nodesConfig *sharding.NodesSetup) (uint32, error) {
	pubKeyBytes, err := esdp.publicKey.ToByteArray()
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

func (esdp *epochStartDataProvider) getTrieFromRootHash(_ []byte) (state.TriesHolder, error) {
	// TODO: get trie from trie syncer
	return state.NewDataTriesHolder(), nil
}

func (esdp *epochStartDataProvider) resetTopicsAndInterceptors() {
	err := esdp.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
	if err != nil {
		log.Info("error unregistering message processors", "error", err)
	}
}

func (esdp *epochStartDataProvider) getMetaBlock(syncer update.MissingHeadersByHashSyncer, hash []byte) (*block.MetaBlock, error) {
	//esdp.requestMetaBlock(hash)
	//
	//time.Sleep(delayAfterRequesting)
	//
	//for {
	//	numConnectedPeers := len(esdp.messenger.Peers())
	//	threshold := int(thresholdForConsideringMetaBlockCorrect * float64(numConnectedPeers))
	//	mb, errConsensusNotReached := esdp.metaBlockInterceptor.GetMetaBlock(hash, threshold)
	//	if errConsensusNotReached == nil {
	//		return mb, nil
	//	}
	//	log.Info("consensus not reached for meta block. re-requesting and trying again...")
	//	esdp.requestMetaBlock(hash)
	//}
	waitTime := 1 * time.Minute
	err := syncer.SyncMissingHeadersByHash(esdp.defaultShardCoordinator.SelfId(), [][]byte{hash}, waitTime)
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

func (esdp *epochStartDataProvider) getEpochStartMetaBlock(epoch uint32) (*block.MetaBlock, error) {
	err := esdp.initTopicForEpochStartMetaBlockInterceptor()
	if err != nil {
		return nil, err
	}
	defer func() {
		esdp.resetTopicsAndInterceptors()
	}()

	esdp.requestEpochStartMetaBlock(epoch)

	time.Sleep(delayAfterRequesting)
	count := 0

	for {
		if count > maxNumTimesToRetry {
			panic("can't sync with other peers")
		}
		count++
		numConnectedPeers := len(esdp.messenger.Peers())
		threshold := int(thresholdForConsideringMetaBlockCorrect * float64(numConnectedPeers))
		mb, errConsensusNotReached := esdp.epochStartMetaBlockInterceptor.GetEpochStartMetaBlock(threshold, epoch)
		if errConsensusNotReached == nil {
			return mb, nil
		}
		log.Info("consensus not reached for meta block. re-requesting and trying again...")
		esdp.requestEpochStartMetaBlock(epoch)
	}
}

func (esdp *epochStartDataProvider) getShardCoordinator(metaBlock *block.MetaBlock, nodesConfig *sharding.NodesSetup) (sharding.Coordinator, error) {
	shardID, err := esdp.getShardID(nodesConfig)
	if err != nil {
		return nil, err
	}

	numOfShards := len(metaBlock.EpochStart.LastFinalizedHeaders)
	return sharding.NewMultiShardCoordinator(uint32(numOfShards), shardID)
}

func (esdp *epochStartDataProvider) getShardHeaders(
	syncer update.MissingHeadersByHashSyncer,
	metaBlock *block.MetaBlock,
	shardCoordinator sharding.Coordinator,
) (map[uint32]*block.Header, error) {
	headersMap := make(map[uint32]*block.Header)

	shardID := shardCoordinator.SelfId()
	if shardID == core.MetachainShardId {
		for _, entry := range metaBlock.EpochStart.LastFinalizedHeaders {
			var hdr *block.Header
			hdr, err := esdp.getShardHeader(syncer, entry.HeaderHash, entry.ShardID)
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

	hdr, err := esdp.getShardHeader(
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

func (esdp *epochStartDataProvider) getShardHeader(
	syncer update.MissingHeadersByHashSyncer,
	hash []byte,
	shardID uint32,
) (*block.Header, error) {
	waitTime := 1 * time.Minute
	err := syncer.SyncMissingHeadersByHash(esdp.defaultShardCoordinator.SelfId(), [][]byte{hash}, waitTime)
	if err != nil {
		return nil, err
	}

	hdrs, err := syncer.GetHeaders()
	if err != nil {
		return nil, err
	}

	syncer.ClearFields()

	return hdrs[string(hash)].(*block.Header), nil

	//esdp.requestShardHeader(shardID, hash)
	//time.Sleep(delayBetweenRequests)
	//
	//count := 0
	//for {
	//	if count > maxNumTimesToRetry {
	//		panic("can't sync with the other peers")
	//	}
	//	count++
	//	numConnectedPeers := len(esdp.messenger.Peers())
	//	threshold := int(thresholdForConsideringMetaBlockCorrect * float64(numConnectedPeers))
	//	mb, errConsensusNotReached := esdp.shardHeaderInterceptor.GetShardHeader(hash, threshold)
	//	if errConsensusNotReached == nil {
	//		return mb, nil
	//	}
	//	log.Info("consensus not reached for shard header. re-requesting and trying again...")
	//	esdp.requestShardHeader(shardID, hash)
	//}
}

func (esdp *epochStartDataProvider) requestMetaBlock(hash []byte) {
	// send more requests
	log.Debug("requested meta block", "hash", hash)
	for i := 0; i < numRequestsToSendOnce; i++ {
		time.Sleep(delayBetweenRequests)
		esdp.requestHandler.RequestMetaHeader(hash)
	}
}

func (esdp *epochStartDataProvider) requestShardHeader(shardID uint32, hash []byte) {
	// send more requests
	log.Debug("requested shard block", "shard ID", shardID, "hash", hash)
	for i := 0; i < numRequestsToSendOnce; i++ {
		time.Sleep(delayBetweenRequests)
		esdp.requestHandler.RequestShardHeader(shardID, hash)
	}
}

func (esdp *epochStartDataProvider) requestEpochStartMetaBlock(epoch uint32) {
	// send more requests
	for i := 0; i < numRequestsToSendOnce; i++ {
		time.Sleep(delayBetweenRequests)
		esdp.requestHandler.RequestStartOfEpochMetaBlock(epoch)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (esdp *epochStartDataProvider) IsInterfaceNil() bool {
	return esdp == nil
}
