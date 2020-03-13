package bootstrap

import (
	"encoding/hex"
	"errors"
	"math"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	factory2 "github.com/ElrondNetwork/elrond-go/data/state/factory"
	factory3 "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
)

var log = logger.GetOrCreate("registration")
var _ process.Interceptor = (*simpleMetaBlockInterceptor)(nil)

const delayBetweenRequests = 1 * time.Second
const delayAfterRequesting = 1 * time.Second
const thresholdForConsideringMetaBlockCorrect = 0.2
const numRequestsToSendOnce = 4

// ComponentsNeededForBootstrap holds the components which need to be initialized from network
type ComponentsNeededForBootstrap struct {
	EpochStartMetaBlock *block.MetaBlock
	NodesConfig         *sharding.NodesSetup
	ShardHeaders        map[uint32]*block.Header
	ShardCoordinator    sharding.Coordinator
	Tries               state.TriesHolder
}

// epochStartDataProvider will handle requesting the needed data to start when joining late the network
type epochStartDataProvider struct {
	publicKey                      crypto.PublicKey
	marshalizer                    marshal.Marshalizer
	hasher                         hashing.Hasher
	messenger                      p2p.Messenger
	nodesConfigProvider            NodesConfigProviderHandler
	epochStartMetaBlockInterceptor EpochStartMetaBlockInterceptorHandler
	metaBlockInterceptor           MetaBlockInterceptorHandler
	shardHeaderInterceptor         ShardHeaderInterceptorHandler
	miniBlockInterceptor           MiniBlockInterceptorHandler
	requestHandlerMeta             process.RequestHandler
}

// ArgsEpochStartDataProvider holds the arguments needed for creating an epoch start data provider component
type ArgsEpochStartDataProvider struct {
	PublicKey                      crypto.PublicKey
	Messenger                      p2p.Messenger
	Marshalizer                    marshal.Marshalizer
	Hasher                         hashing.Hasher
	NodesConfigProvider            NodesConfigProviderHandler
	EpochStartMetaBlockInterceptor EpochStartMetaBlockInterceptorHandler
	MetaBlockInterceptor           MetaBlockInterceptorHandler
	ShardHeaderInterceptor         ShardHeaderInterceptorHandler
	MiniBlockInterceptor           MiniBlockInterceptorHandler
}

// NewEpochStartDataProvider will return a new instance of epochStartDataProvider
func NewEpochStartDataProvider(args ArgsEpochStartDataProvider) (*epochStartDataProvider, error) {
	// TODO: maybe remove these nil checks as all of them have been done in the factory
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
	return &epochStartDataProvider{
		publicKey:                      args.PublicKey,
		marshalizer:                    args.Marshalizer,
		hasher:                         args.Hasher,
		messenger:                      args.Messenger,
		nodesConfigProvider:            args.NodesConfigProvider,
		epochStartMetaBlockInterceptor: args.EpochStartMetaBlockInterceptor,
		metaBlockInterceptor:           args.MetaBlockInterceptor,
		shardHeaderInterceptor:         args.ShardHeaderInterceptor,
		miniBlockInterceptor:           args.MiniBlockInterceptor,
	}, nil
}

// Bootstrap will handle requesting and receiving the needed information the node will bootstrap from
func (esdp *epochStartDataProvider) Bootstrap() (*ComponentsNeededForBootstrap, error) {
	err := esdp.initTopicsAndInterceptors()
	if err != nil {
		return nil, err
	}
	defer func() {
		esdp.resetTopicsAndInterceptors()
	}()

	requestHandlerMeta, err := esdp.createRequestHandler()
	if err != nil {
		return nil, err
	}

	esdp.requestHandlerMeta = requestHandlerMeta

	epochNumForRequestingTheLatestAvailable := uint32(math.MaxUint32)
	metaBlock, err := esdp.getEpochStartMetaBlock(epochNumForRequestingTheLatestAvailable)
	if err != nil {
		return nil, err
	}

	prevMetaBlock, err := esdp.getEpochStartMetaBlock(metaBlock.Epoch - 1)
	if err != nil {
		return nil, err
	}

	esdp.changeMessageProcessorsForMetaBlocks()

	log.Info("previous meta block", "epoch", prevMetaBlock.Epoch)
	nodesConfig, err := esdp.nodesConfigProvider.GetNodesConfigForMetaBlock(metaBlock)
	if err != nil {
		return nil, err
	}

	shardCoordinator, err := esdp.getShardCoordinator(metaBlock, nodesConfig)
	if err != nil {
		return nil, err
	}

	shardHeaders, err := esdp.getShardHeaders(metaBlock, nodesConfig, shardCoordinator)
	if err != nil {
		log.Debug("shard headers not found", "error", err)
	}

	epochStartData, err := esdp.getCurrentEpochStartData(shardCoordinator, metaBlock)
	if err != nil {
		return nil, err
	}

	for _, mb := range epochStartData.PendingMiniBlockHeaders {
		receivedMb, err := esdp.getMiniBlock(&mb)
		if err != nil {
			return nil, err
		}
		log.Info("received miniblock", "type", receivedMb.Type)
	}

	lastFinalizedMetaBlock, err := esdp.getMetaBlock(epochStartData.LastFinishedMetaBlock)
	if err != nil {
		return nil, err
	}
	log.Info("received last finalized meta block", "nonce", lastFinalizedMetaBlock.Nonce)

	firstPendingMetaBlock, err := esdp.getMetaBlock(epochStartData.FirstPendingMetaBlock)
	if err != nil {
		return nil, err
	}
	log.Info("received first pending meta block", "nonce", firstPendingMetaBlock.Nonce)

	trie, err := esdp.getTrieFromRootHash(epochStartData.RootHash)
	if err != nil {
		return nil, err
	}

	return &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: metaBlock,
		NodesConfig:         nodesConfig,
		ShardHeaders:        shardHeaders,
		ShardCoordinator:    shardCoordinator,
		Tries:               trie,
	}, nil
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

func (esdp *epochStartDataProvider) createRequestHandler() (process.RequestHandler, error) {
	dataPacker, err := partitioning.NewSimpleDataPacker(esdp.marshalizer)
	if err != nil {
		return nil, err
	}

	shardC, err := sharding.NewMultiShardCoordinator(2, core.MetachainShardId)
	if err != nil {
		return nil, err
	}

	storageService := &disabled.ChainStorer{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return disabled.NewDisabledStorer()
		},
	}

	cacher := disabled.NewDisabledPoolsHolder()
	triesHolder := state.NewDataTriesHolder()
	var stateTrie data.Trie
	// TODO: change from integrationsTests.CreateAccountsDB
	_, stateTrie, _ = integrationTests.CreateAccountsDB(factory2.UserAccount)
	triesHolder.Put([]byte(factory3.UserAccountTrie), stateTrie)

	var peerTrie data.Trie
	_, peerTrie, _ = integrationTests.CreateAccountsDB(factory2.ValidatorAccount)
	triesHolder.Put([]byte(factory3.PeerAccountTrie), peerTrie)

	resolversContainerArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:         shardC,
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
		return nil, err
	}

	container, err := resolverFactory.Create()
	if err != nil {
		return nil, err
	}

	finder, err := containers.NewResolversFinder(container, shardC)
	if err != nil {
		return nil, err
	}

	requestedItemsHandler := timecache.NewTimeCache(100)

	maxToRequest := 100

	return requestHandlers.NewMetaResolverRequestHandler(finder, requestedItemsHandler, maxToRequest)
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
	esdp.requestHandlerMeta.RequestMiniBlock(miniBlockHeader.ReceiverShardID, miniBlockHeader.Hash)
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

func (esdp *epochStartDataProvider) initTopicsAndInterceptors() error {
	err := esdp.messenger.CreateTopic(factory.MetachainBlocksTopic, true)
	if err != nil {
		log.Info("error unregistering message processor", "error", err)
		return err
	}

	err = esdp.messenger.RegisterMessageProcessor(factory.MetachainBlocksTopic, esdp.epochStartMetaBlockInterceptor)
	if err != nil {
		return err
	}

	err = esdp.messenger.CreateTopic(factory.ShardBlocksTopic+"_1_META", true)
	if err != nil {
		log.Info("error unregistering message processor", "error", err)
		return err
	}

	err = esdp.messenger.RegisterMessageProcessor(factory.ShardBlocksTopic+"_1_META", esdp.shardHeaderInterceptor)
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
	err := esdp.messenger.UnregisterAllMessageProcessors()
	if err != nil {
		log.Info("error unregistering message processors", "error", err)
	}
}

func (esdp *epochStartDataProvider) getMetaBlock(hash []byte) (*block.MetaBlock, error) {
	esdp.requestMetaBlock(hash)

	time.Sleep(delayAfterRequesting)

	for {
		numConnectedPeers := len(esdp.messenger.Peers())
		threshold := int(thresholdForConsideringMetaBlockCorrect * float64(numConnectedPeers))
		mb, errConsensusNotReached := esdp.metaBlockInterceptor.GetMetaBlock(hash, threshold)
		if errConsensusNotReached == nil {
			return mb, nil
		}
		log.Info("consensus not reached for meta block. re-requesting and trying again...")
		esdp.requestMetaBlock(hash)
	}
}

func (esdp *epochStartDataProvider) getEpochStartMetaBlock(epoch uint32) (*block.MetaBlock, error) {
	esdp.requestEpochStartMetaBlock(epoch)

	time.Sleep(delayAfterRequesting)

	for {
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

	numOfShards := nodesConfig.NumberOfShards()
	if numOfShards == 1 {
		return &sharding.OneShardCoordinator{}, nil
	}

	return sharding.NewMultiShardCoordinator(numOfShards, shardID)
}

func (esdp *epochStartDataProvider) getShardHeaders(
	metaBlock *block.MetaBlock,
	nodesConfig *sharding.NodesSetup,
	shardCoordinator sharding.Coordinator,
) (map[uint32]*block.Header, error) {
	headersMap := make(map[uint32]*block.Header)

	shardID := shardCoordinator.SelfId()
	if shardID == core.MetachainShardId {
		for _, entry := range metaBlock.EpochStart.LastFinalizedHeaders {
			var hdr *block.Header
			hdr, err := esdp.getShardHeader(entry.HeaderHash, entry.ShardID)
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
	hash []byte,
	shardID uint32,
) (*block.Header, error) {
	esdp.requestShardHeader(shardID, hash)
	time.Sleep(delayBetweenRequests)

	for {
		numConnectedPeers := len(esdp.messenger.Peers())
		threshold := int(thresholdForConsideringMetaBlockCorrect * float64(numConnectedPeers))
		mb, errConsensusNotReached := esdp.shardHeaderInterceptor.GetShardHeader(threshold)
		if errConsensusNotReached == nil {
			return mb, nil
		}
		log.Info("consensus not reached for shard header. re-requesting and trying again...")
		esdp.requestShardHeader(shardID, hash)
	}
}

func (esdp *epochStartDataProvider) requestMetaBlock(hash []byte) {
	// send more requests
	log.Debug("requested meta block", "hash", hash)
	for i := 0; i < numRequestsToSendOnce; i++ {
		time.Sleep(delayBetweenRequests)
		esdp.requestHandlerMeta.RequestMetaHeader(hash)
	}
}

func (esdp *epochStartDataProvider) requestShardHeader(shardID uint32, hash []byte) {
	// send more requests
	log.Debug("requested shard block", "shard ID", shardID, "hash", hash)
	for i := 0; i < numRequestsToSendOnce; i++ {
		time.Sleep(delayBetweenRequests)
		esdp.requestHandlerMeta.RequestShardHeader(shardID, hash)
	}
}

func (esdp *epochStartDataProvider) requestEpochStartMetaBlock(epoch uint32) {
	// send more requests
	for i := 0; i < numRequestsToSendOnce; i++ {
		time.Sleep(delayBetweenRequests)
		esdp.requestHandlerMeta.RequestStartOfEpochMetaBlock(epoch)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (esdp *epochStartDataProvider) IsInterfaceNil() bool {
	return esdp == nil
}
