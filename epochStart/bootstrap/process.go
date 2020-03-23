package bootstrap

import (
	"encoding/hex"
	"errors"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	trieFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	factoryDataPool "github.com/ElrondNetwork/elrond-go/dataRetriever/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	factoryInterceptors "github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/factory"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/storagehandler"
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

var log = logger.GetOrCreate("epochStart/bootstrap")

const timeToWait = 5 * time.Second

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
	publicKey          crypto.PublicKey
	marshalizer        marshal.Marshalizer
	hasher             hashing.Hasher
	messenger          p2p.Messenger
	generalConfig      config.Config
	economicsData      *economics.EconomicsData
	singleSigner       crypto.SingleSigner
	blockSingleSigner  crypto.SingleSigner
	keyGen             crypto.KeyGenerator
	blockKeyGen        crypto.KeyGenerator
	requestHandler     process.RequestHandler
	whiteListHandler   update.WhiteListHandler
	shardCoordinator   sharding.Coordinator
	genesisNodesConfig *sharding.NodesSetup
	workingDir         string
	defaultDBPath      string
	defaultEpochString string

	interceptorContainer process.InterceptorsContainer
	dataPool             dataRetriever.PoolsHolder
	computedEpoch        uint32

	miniBlocksSyncer          epochStart.PendingMiniBlocksSyncHandler
	headersSyncer             epochStart.HeadersByHashSyncer
	epochStartMetaBlockSyncer epochStart.StartOfEpochMetaSyncer

	baseData baseDataInStorage
}

type baseDataInStorage struct {
	shardId   uint32
	lastRound uint64
	lastEpoch uint32
}

// ArgsEpochStartBootstrap holds the arguments needed for creating an epoch start data provider component
type ArgsEpochStartBootstrap struct {
	PublicKey               crypto.PublicKey
	Messenger               p2p.Messenger
	Marshalizer             marshal.Marshalizer
	Hasher                  hashing.Hasher
	GeneralConfig           config.Config
	EconomicsConfig         config.EconomicsConfig
	GenesisShardCoordinator sharding.Coordinator
	SingleSigner            crypto.SingleSigner
	BlockSingleSigner       crypto.SingleSigner
	KeyGen                  crypto.KeyGenerator
	BlockKeyGen             crypto.KeyGenerator
	WhiteListHandler        update.WhiteListHandler
	GenesisNodesConfig      *sharding.NodesSetup
	WorkingDir              string
	DefaultDBPath           string
	DefaultEpochString      string
}

// NewEpochStartBootstrap will return a new instance of epochStartBootstrap
func NewEpochStartBootstrapHandler(args ArgsEpochStartBootstrap) (*epochStartBootstrap, error) {
	epochStartProvider := &epochStartBootstrap{
		publicKey:          args.PublicKey,
		marshalizer:        args.Marshalizer,
		hasher:             args.Hasher,
		messenger:          args.Messenger,
		generalConfig:      args.GeneralConfig,
		whiteListHandler:   args.WhiteListHandler,
		genesisNodesConfig: args.GenesisNodesConfig,
		workingDir:         args.WorkingDir,
		defaultEpochString: args.DefaultEpochString,
		defaultDBPath:      args.DefaultEpochString,
		keyGen:             args.KeyGen,
		blockKeyGen:        args.BlockKeyGen,
		singleSigner:       args.SingleSigner,
		blockSingleSigner:  args.BlockSingleSigner,
	}

	return epochStartProvider, nil
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
	// TODO: write gathered data in baseDataInStorage
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

	e.computeMostProbableEpoch()
	e.searchDataInLocalStorage()

	isCurrentEpochSaved := e.baseData.lastEpoch+1 >= e.computedEpoch
	if isCurrentEpochSaved {
		return e.prepareEpochFromStorage()
	}

	err := e.prepareComponentsToSyncFromNetwork()
	if err != nil {
		return 0, 0, 0, err
	}

	return e.requestAndProcessing()
}

func (e *epochStartBootstrap) prepareEpochFromStorage() (uint32, uint32, uint32, error) {
	return 0, 0, 0, nil
}

func (e *epochStartBootstrap) prepareComponentsToSyncFromNetwork() error {
	var err error
	e.shardCoordinator, err = sharding.NewMultiShardCoordinator(e.genesisNodesConfig.NumberOfShards(), core.MetachainShardId)
	if err != nil {
		return err
	}

	err = e.createRequestHandler()
	if err != nil {
		return err
	}

	e.dataPool, err = factoryDataPool.NewDataPoolFromConfig(
		factoryDataPool.ArgsDataPool{
			Config:           &e.generalConfig,
			EconomicsData:    e.economicsData,
			ShardCoordinator: e.shardCoordinator,
		},
	)

	args := factoryInterceptors.ArgsEpochStartInterceptorContainer{
		Config:            e.generalConfig,
		ShardCoordinator:  e.shardCoordinator,
		Marshalizer:       e.marshalizer,
		Hasher:            e.hasher,
		Messenger:         e.messenger,
		DataPool:          e.dataPool,
		SingleSigner:      e.singleSigner,
		BlockSingleSigner: e.blockSingleSigner,
		KeyGen:            e.keyGen,
		BlockKeyGen:       e.blockKeyGen,
		WhiteListHandler:  e.whiteListHandler,
	}

	e.interceptorContainer, err = factoryInterceptors.NewEpochStartInterceptorsContainer(args)
	if err != nil {
		return err
	}

	// TODO epochStart meta syncer
	e.epochStartMetaBlockSyncer = NewEpochStartmetaBlockSyncer()

	syncMiniBlocksArgs := sync.ArgsNewPendingMiniBlocksSyncer{
		Storage:        &disabled.Storer{},
		Cache:          e.dataPool.MiniBlocks(),
		Marshalizer:    e.marshalizer,
		RequestHandler: e.requestHandler,
	}
	e.miniBlocksSyncer, err = sync.NewPendingMiniBlocksSyncer(syncMiniBlocksArgs)

	syncMissingHeadersArgs := sync.ArgsNewMissingHeadersByHashSyncer{
		Storage:        &disabled.Storer{},
		Cache:          e.dataPool.Headers(),
		Marshalizer:    e.marshalizer,
		RequestHandler: e.requestHandler,
	}
	e.headersSyncer, err = sync.NewMissingheadersByHashSyncer(syncMissingHeadersArgs)

	return nil
}

func (e *epochStartBootstrap) syncHeadersFrom(meta *block.MetaBlock) (map[string]data.HeaderHandler, error) {
	hashesToRequest := make([][]byte, 0, len(meta.EpochStart.LastFinalizedHeaders)+1)
	shardIds := make([]uint32, 0, len(meta.EpochStart.LastFinalizedHeaders)+1)

	for _, epochStartData := range meta.EpochStart.LastFinalizedHeaders {
		hashesToRequest = append(hashesToRequest, epochStartData.HeaderHash)
		shardIds = append(shardIds, epochStartData.ShardID)
	}

	hashesToRequest = append(hashesToRequest, meta.EpochStart.Economics.PrevEpochStartHash)
	shardIds = append(shardIds, core.MetachainShardId)

	err := e.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, timeToWait)
	if err != nil {
		return nil, err
	}

	return e.headersSyncer.GetHeaders()
}

// Bootstrap will handle requesting and receiving the needed information the node will bootstrap from
func (e *epochStartBootstrap) requestAndProcessing() (uint32, uint32, uint32, error) {
	metaBlock, err := e.epochStartMetaBlockSyncer.SyncEpochStartMeta(timeToWait)
	if err != nil {
		return 0, 0, 0, err
	}

	headers, err := e.syncHeadersFrom(metaBlock)
	if err != nil {
		return 0, 0, 0, err
	}

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
		DataPools:                e.dataPool,
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

// IsInterfaceNil returns true if there is no value under the interface
func (e *epochStartBootstrap) IsInterfaceNil() bool {
	return e == nil
}
