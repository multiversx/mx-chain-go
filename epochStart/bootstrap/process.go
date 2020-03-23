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
	ShardHeader                 *block.Header
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
	// TODO: compute self shard ID for current epoch
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

	numShards := uint32(len(metaBlock.EpochStart.LastFinalizedHeaders))

	headers, err := e.syncHeadersFrom(metaBlock)
	if err != nil {
		return 0, 0, 0, err
	}

	prevEpochStartMetaHash := metaBlock.EpochStart.Economics.PrevEpochStartHash
	prevEpochStartMeta, ok := headers[string(prevEpochStartMetaHash)].(*block.MetaBlock)
	if !ok {
		return 0, 0, 0, epochStart.ErrWrongTypeAssertion
	}

	nodesConfigForCurrEpoch, err := e.nodesConfigProvider.GetNodesConfigForMetaBlock(metaBlock)
	if err != nil {
		return 0, 0, 0, err
	}

	nodesConfigForPrevEpoch, err := e.nodesConfigProvider.GetNodesConfigForMetaBlock(prevEpochStartMeta)
	if err != nil {
		return 0, 0, 0, err
	}

	selfShardId, err := e.getShardID(nodesConfigForCurrEpoch)
	if err != nil {
		return 0, 0, 0, err
	}

	e.shardCoordinator, err = sharding.NewMultiShardCoordinator(numShards, selfShardId)
	if err != nil {
		return 0, 0, 0, err
	}

	if e.shardCoordinator.SelfId() == core.MetachainShardId {
		return e.requestAndProcessForShard(e.shardCoordinator.SelfId())
	}

	return e.requestAndProcessForMeta()
}

func (e *epochStartBootstrap) requestAndProcessForMeta() error {
	// accounts and peer accounts syncer

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock:         metaBlock,
		PreviousEpochStartMetaBlock: prevEpochStartMeta,
		ShardHeader:                 shardHeaderForShard,
		NodesConfig:                 nodesConfig,
		ShardHeaders:                shardHeaders,
		ShardCoordinator:            e.shardCoordinator,
		Tries:                       trieToReturn,
		PendingMiniBlocks:           pendingMiniBlocks,
	}

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

	errSavingToStorage := storageHandlerComponent.SaveDataToStorage(*components)
	if errSavingToStorage != nil {
		return errSavingToStorage
	}

	return nil
}

func (e *epochStartBootstrap) requestAndProcessForShard(shardId uint32, metaBlock *block.MetaBlock) error {
	var epochStartData block.EpochStartShardData
	found := false
	for _, shardData := range metaBlock.EpochStart.LastFinalizedHeaders {
		if shardData.ShardID == shardId {
			epochStartData = shardData
			found = true
			break
		}
	}
	if !found {
		return epochStart.ErrEpochStartDataForShardNotFound
	}

	err := e.miniBlocksSyncer.SyncPendingMiniBlocks(epochStartData.PendingMiniBlockHeaders, timeToWait)
	if err != nil {
		return err
	}

	pendingMiniBlocks, err := e.miniBlocksSyncer.GetMiniBlocks()
	if err != nil {
		return err
	}

	shardIds := make([]uint32, 0, 2)
	hashesToRequest := make([][]byte, 0, 2)
	hashesToRequest = append(hashesToRequest, epochStartData.LastFinishedMetaBlock)
	hashesToRequest = append(hashesToRequest, epochStartData.FirstPendingMetaBlock)
	shardIds = append(shardIds, shardId)
	shardIds = append(shardIds, shardId)

	e.headersSyncer.ClearFields()
	err = e.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, timeToWait)
	if err != nil {
		return err
	}

	neededHeaders, err := e.headersSyncer.GetHeaders()
	if err != nil {
		return err
	}

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock:         metaBlock,
		PreviousEpochStartMetaBlock: prevEpochStartMeta,
		ShardHeader:                 shardHeaderForShard,
		NodesConfig:                 nodesConfig,
		ShardHeaders:                shardHeaders,
		ShardCoordinator:            e.shardCoordinator,
		Tries:                       trieToReturn,
		PendingMiniBlocks:           pendingMiniBlocks,
	}

	storageHandlerComponent, err := storagehandler.NewShardStorageHandler(
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

	errSavingToStorage := storageHandlerComponent.SaveDataToStorage(*components)
	if errSavingToStorage != nil {
		return errSavingToStorage
	}

	return nil
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

// IsInterfaceNil returns true if there is no value under the interface
func (e *epochStartBootstrap) IsInterfaceNil() bool {
	return e == nil
}
