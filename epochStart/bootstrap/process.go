package bootstrap

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/syncer"
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
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
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
	NodesConfig                 *sharding.NodesCoordinatorRegistry
	Headers                     map[string]data.HeaderHandler
	ShardCoordinator            sharding.Coordinator
	UserAccountTries            map[string]data.Trie
	PeerAccountTries            map[string]data.Trie
	PendingMiniBlocks           map[string]*block.MiniBlock
}

// epochStartBootstrap will handle requesting the needed data to start when joining late the network
type epochStartBootstrap struct {
	// should come via arguments
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
	shardCoordinator   sharding.Coordinator
	genesisNodesConfig *sharding.NodesSetup
	pathManager        storage.PathManagerHandler
	workingDir         string
	defaultDBPath      string
	defaultEpochString string
	defaultShardString string
	rater              sharding.ChanceComputer

	// created components
	requestHandler            process.RequestHandler
	interceptorContainer      process.InterceptorsContainer
	dataPool                  dataRetriever.PoolsHolder
	miniBlocksSyncer          epochStart.PendingMiniBlocksSyncHandler
	headersSyncer             epochStart.HeadersByHashSyncer
	epochStartMetaBlockSyncer epochStart.StartOfEpochMetaSyncer
	nodesConfigHandler        StartOfEpochNodesConfigHandler
	userTrieStorageManager    data.StorageManager
	peerTrieStorageManager    data.StorageManager
	whiteListHandler          update.WhiteListHandler

	// gathered data
	epochStartMeta     *block.MetaBlock
	prevEpochStartMeta *block.MetaBlock
	syncedHeaders      map[string]data.HeaderHandler
	nodesConfig        *sharding.NodesCoordinatorRegistry
	userAccountTries   map[string]data.Trie
	peerAccountTries   map[string]data.Trie
	baseData           baseDataInStorage
	computedEpoch      uint32
}

type baseDataInStorage struct {
	shardId        uint32
	numberOfShards uint32
	lastRound      int64
	lastEpoch      uint32
}

// ArgsEpochStartBootstrap holds the arguments needed for creating an epoch start data provider component
type ArgsEpochStartBootstrap struct {
	PublicKey          crypto.PublicKey
	Marshalizer        marshal.Marshalizer
	Hasher             hashing.Hasher
	Messenger          p2p.Messenger
	GeneralConfig      config.Config
	EconomicsData      *economics.EconomicsData
	SingleSigner       crypto.SingleSigner
	BlockSingleSigner  crypto.SingleSigner
	KeyGen             crypto.KeyGenerator
	BlockKeyGen        crypto.KeyGenerator
	GenesisNodesConfig *sharding.NodesSetup
	PathManager        storage.PathManagerHandler
	WorkingDir         string
	DefaultDBPath      string
	DefaultEpochString string
	DefaultShardString string
	Rater              sharding.ChanceComputer
}

// NewEpochStartBootstrap will return a new instance of epochStartBootstrap
func NewEpochStartBootstrapHandler(args ArgsEpochStartBootstrap) (*epochStartBootstrap, error) {
	epochStartProvider := &epochStartBootstrap{
		publicKey:          args.PublicKey,
		marshalizer:        args.Marshalizer,
		hasher:             args.Hasher,
		messenger:          args.Messenger,
		generalConfig:      args.GeneralConfig,
		genesisNodesConfig: args.GenesisNodesConfig,
		workingDir:         args.WorkingDir,
		defaultEpochString: args.DefaultEpochString,
		defaultDBPath:      args.DefaultEpochString,
		defaultShardString: args.DefaultShardString,
		keyGen:             args.KeyGen,
		blockKeyGen:        args.BlockKeyGen,
		singleSigner:       args.SingleSigner,
		blockSingleSigner:  args.BlockSingleSigner,
		rater:              args.Rater,
	}

	return epochStartProvider, nil
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

	// TODO: make a better decision according to lastRound, lastEpoch
	isCurrentEpochSaved := e.baseData.lastEpoch+1 >= e.computedEpoch
	if isCurrentEpochSaved {
		epoch, shardId, numOfShards, err := e.prepareEpochFromStorage()
		if err == nil {
			return epoch, shardId, numOfShards, nil
		}
	}

	var err error
	e.shardCoordinator, err = sharding.NewMultiShardCoordinator(e.genesisNodesConfig.NumberOfShards(), core.MetachainShardId)
	if err != nil {
		return 0, 0, 0, err
	}

	err = e.prepareComponentsToSyncFromNetwork()
	if err != nil {
		return 0, 0, 0, err
	}

	return e.requestAndProcessing()
}

func (e *epochStartBootstrap) prepareComponentsToSyncFromNetwork() error {
	var err error

	whiteListCache, err := storageUnit.NewCache(
		storageUnit.CacheType(e.generalConfig.WhiteListPool.Type),
		e.generalConfig.WhiteListPool.Size,
		e.generalConfig.WhiteListPool.Shards,
	)
	if err != nil {
		return err
	}

	e.whiteListHandler, err = interceptors.NewWhiteListDataVerifier(whiteListCache)
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

	argsEpochStartSyncer := ArgsNewEpochStartMetaSyncer{
		RequestHandler: e.requestHandler,
		Messenger:      e.messenger,
		Marshalizer:    e.marshalizer,
		Hasher:         e.hasher,
	}
	e.epochStartMetaBlockSyncer, err = NewEpochStartMetaSyncer(argsEpochStartSyncer)
	if err != nil {
		return err
	}

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
	var err error
	e.epochStartMeta, err = e.epochStartMetaBlockSyncer.SyncEpochStartMeta(timeToWait)
	if err != nil {
		return 0, 0, 0, err
	}

	e.baseData.numberOfShards = uint32(len(e.epochStartMeta.EpochStart.LastFinalizedHeaders))
	e.baseData.lastEpoch = e.epochStartMeta.Epoch

	e.syncedHeaders, err = e.syncHeadersFrom(e.epochStartMeta)
	if err != nil {
		return 0, 0, 0, err
	}

	prevEpochStartMetaHash := e.epochStartMeta.EpochStart.Economics.PrevEpochStartHash
	prevEpochStartMeta, ok := e.syncedHeaders[string(prevEpochStartMetaHash)].(*block.MetaBlock)
	if !ok {
		return 0, 0, 0, epochStart.ErrWrongTypeAssertion
	}
	e.prevEpochStartMeta = prevEpochStartMeta

	pubKeyBytes, err := e.publicKey.ToByteArray()
	if err != nil {
		return 0, 0, 0, err
	}

	err = e.createTrieStorageManagers()
	if err != nil {
		return 0, 0, 0, err
	}

	err = e.syncPeerAccountsState(e.epochStartMeta.ValidatorStatsRootHash)
	if err != nil {
		return 0, 0, 0, err
	}

	err = e.processNodesConfig(pubKeyBytes, e.epochStartMeta.ValidatorStatsRootHash)
	if err != nil {
		return 0, 0, 0, err
	}

	e.shardCoordinator, err = sharding.NewMultiShardCoordinator(e.baseData.numberOfShards, e.baseData.shardId)
	if err != nil {
		return 0, 0, 0, err
	}

	if e.shardCoordinator.SelfId() == core.MetachainShardId {
		err = e.requestAndProcessForShard()
		if err != nil {
			return 0, 0, 0, err
		}
	}

	err = e.requestAndProcessForMeta()
	if err != nil {
		return 0, 0, 0, err
	}

	return e.baseData.shardId, e.baseData.numberOfShards, e.baseData.lastEpoch, nil
}

func (e *epochStartBootstrap) processNodesConfig(pubKey []byte, rootHash []byte) error {
	accountFactory, err := factory.NewAccountFactoryCreator(state.ValidatorAccount)
	if err != nil {
		return err
	}
	peerAccountsDB, err := state.NewPeerAccountsDB(e.peerAccountTries[string(rootHash)], e.hasher, e.marshalizer, accountFactory)
	if err != nil {
		return err
	}

	blsAddressConverter, err := addressConverters.NewPlainAddressConverter(
		e.generalConfig.BLSPublicKey.Length,
		e.generalConfig.BLSPublicKey.Prefix,
	)
	if err != nil {
		return err
	}
	argsNewValidatorStatusSyncers := ArgsNewSyncValidatorStatus{
		DataPool:            e.dataPool,
		Marshalizer:         e.marshalizer,
		RequestHandler:      e.requestHandler,
		Rater:               e.rater,
		GenesisNodesConfig:  e.genesisNodesConfig,
		ValidatorAccountsDB: peerAccountsDB,
		AdrConv:             blsAddressConverter,
	}
	e.nodesConfigHandler, err = NewSyncValidatorStatus(argsNewValidatorStatusSyncers)
	if err != nil {
		return err
	}

	e.nodesConfig, e.baseData.shardId, err = e.nodesConfigHandler.NodesConfigFromMetaBlock(e.epochStartMeta, e.prevEpochStartMeta, pubKey)
	return err
}

func (e *epochStartBootstrap) requestAndProcessForMeta() error {
	err := e.syncUserAccountsState(e.epochStartMeta.RootHash)
	if err != nil {
		return err
	}

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock:         e.epochStartMeta,
		PreviousEpochStartMetaBlock: e.prevEpochStartMeta,
		NodesConfig:                 e.nodesConfig,
		Headers:                     e.syncedHeaders,
		ShardCoordinator:            e.shardCoordinator,
		UserAccountTries:            e.userAccountTries,
		PeerAccountTries:            e.peerAccountTries,
	}

	storageHandlerComponent, err := NewMetaStorageHandler(
		e.generalConfig,
		e.shardCoordinator,
		e.pathManager,
		e.marshalizer,
		e.hasher,
		e.epochStartMeta.Epoch,
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

func (e *epochStartBootstrap) requestAndProcessForShard() error {
	var epochStartData block.EpochStartShardData
	found := false
	for _, shardData := range e.epochStartMeta.EpochStart.LastFinalizedHeaders {
		if shardData.ShardID == e.shardCoordinator.SelfId() {
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
	shardIds = append(shardIds, e.shardCoordinator.SelfId())
	shardIds = append(shardIds, e.shardCoordinator.SelfId())

	e.headersSyncer.ClearFields()
	err = e.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, timeToWait)
	if err != nil {
		return err
	}

	neededHeaders, err := e.headersSyncer.GetHeaders()
	if err != nil {
		return err
	}

	for hash, hdr := range neededHeaders {
		e.syncedHeaders[hash] = hdr
	}

	ownShardHdr, ok := e.syncedHeaders[string(epochStartData.HeaderHash)].(*block.Header)
	if !ok {
		return epochStart.ErrWrongTypeAssertion
	}

	err = e.syncUserAccountsState(ownShardHdr.RootHash)
	if err != nil {
		return err
	}

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock:         e.epochStartMeta,
		PreviousEpochStartMetaBlock: e.prevEpochStartMeta,
		ShardHeader:                 ownShardHdr,
		NodesConfig:                 e.nodesConfig,
		Headers:                     e.syncedHeaders,
		ShardCoordinator:            e.shardCoordinator,
		UserAccountTries:            e.userAccountTries,
		PeerAccountTries:            e.peerAccountTries,
		PendingMiniBlocks:           pendingMiniBlocks,
	}

	storageHandlerComponent, err := NewShardStorageHandler(
		e.generalConfig,
		e.shardCoordinator,
		e.pathManager,
		e.marshalizer,
		e.hasher,
		e.baseData.lastEpoch,
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
	argsUserAccountsSyncer := syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:             e.hasher,
			Marshalizer:        e.marshalizer,
			TrieStorageManager: e.userTrieStorageManager,
			RequestHandler:     e.requestHandler,
			WaitTime:           timeToWait,
			Cacher:             e.dataPool.TrieNodes(),
		},
		ShardId: e.shardCoordinator.SelfId(),
	}
	accountsDBSyncer, err := syncer.NewUserAccountsSyncer(argsUserAccountsSyncer)
	if err != nil {
		return err
	}

	err = accountsDBSyncer.SyncAccounts(rootHash)
	if err != nil {
		return err
	}

	e.userAccountTries = accountsDBSyncer.GetSyncedTries()
	return nil
}

func (e *epochStartBootstrap) createTrieStorageManagers() error {
	dbConfig := storageFactory.GetDBFromConfig(e.generalConfig.AccountsTrieStorage.DB)
	trieStorage, err := storageUnit.NewStorageUnitFromConf(
		storageFactory.GetCacherFromConfig(e.generalConfig.AccountsTrieStorage.Cache),
		dbConfig,
		storageFactory.GetBloomFromConfig(e.generalConfig.AccountsTrieStorage.Bloom),
	)
	if err != nil {
		return err
	}

	e.userTrieStorageManager, err = trie.NewTrieStorageManagerWithoutPruning(trieStorage)
	if err != nil {
		return err
	}

	dbConfig = storageFactory.GetDBFromConfig(e.generalConfig.PeerAccountsTrieStorage.DB)
	trieStorage, err = storageUnit.NewStorageUnitFromConf(
		storageFactory.GetCacherFromConfig(e.generalConfig.PeerAccountsTrieStorage.Cache),
		dbConfig,
		storageFactory.GetBloomFromConfig(e.generalConfig.PeerAccountsTrieStorage.Bloom),
	)
	if err != nil {
		return err
	}

	e.peerTrieStorageManager, err = trie.NewTrieStorageManagerWithoutPruning(trieStorage)
	if err != nil {
		return err
	}

	return nil
}

func (e *epochStartBootstrap) syncPeerAccountsState(rootHash []byte) error {
	argsValidatorAccountsSyncer := syncer.ArgsNewValidatorAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:             e.hasher,
			Marshalizer:        e.marshalizer,
			TrieStorageManager: e.peerTrieStorageManager,
			RequestHandler:     e.requestHandler,
			WaitTime:           timeToWait,
			Cacher:             e.dataPool.TrieNodes(),
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

	e.peerAccountTries = accountsDBSyncer.GetSyncedTries()
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
	err = e.createTrieStorageManagers()
	if err != nil {
		return err
	}

	stateTrie, err := trie.NewTrie(e.userTrieStorageManager, e.marshalizer, e.hasher)
	if err != nil {
		return err
	}
	triesHolder.Put([]byte(trieFactory.UserAccountTrie), stateTrie)

	peerTrie, err := trie.NewTrie(e.peerTrieStorageManager, e.marshalizer, e.hasher)
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

// IsInterfaceNil returns true if there is no value under the interface
func (e *epochStartBootstrap) IsInterfaceNil() bool {
	return e == nil
}
