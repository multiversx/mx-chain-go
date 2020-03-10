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

const requestSuffix = "_REQUEST"
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

type shardData struct {
	headerResolver ShardHeaderResolverHandler
	epochStartData *block.EpochStartShardData
}

// epochStartDataProvider will handle requesting the needed data to start when joining late the network
type epochStartDataProvider struct {
	publicKey              crypto.PublicKey
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	messenger              p2p.Messenger
	nodesConfigProvider    NodesConfigProviderHandler
	metaBlockInterceptor   MetaBlockInterceptorHandler
	shardHeaderInterceptor ShardHeaderInterceptorHandler
	metaBlockResolver      MetaBlockResolverHandler
	requestHandlerMeta     process.RequestHandler
}

// ArgsEpochStartDataProvider holds the arguments needed for creating an epoch start data provider component
type ArgsEpochStartDataProvider struct {
	PublicKey              crypto.PublicKey
	Messenger              p2p.Messenger
	Marshalizer            marshal.Marshalizer
	Hasher                 hashing.Hasher
	NodesConfigProvider    NodesConfigProviderHandler
	MetaBlockInterceptor   MetaBlockInterceptorHandler
	ShardHeaderInterceptor ShardHeaderInterceptorHandler
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
	if check.IfNil(args.NodesConfigProvider) {
		return nil, ErrNilNodesConfigProvider
	}
	if check.IfNil(args.MetaBlockInterceptor) {
		return nil, ErrNilMetaBlockInterceptor
	}
	if check.IfNil(args.ShardHeaderInterceptor) {
		return nil, ErrNilShardHeaderInterceptor
	}
	return &epochStartDataProvider{
		publicKey:              args.PublicKey,
		marshalizer:            args.Marshalizer,
		hasher:                 args.Hasher,
		messenger:              args.Messenger,
		nodesConfigProvider:    args.NodesConfigProvider,
		metaBlockInterceptor:   args.MetaBlockInterceptor,
		shardHeaderInterceptor: args.ShardHeaderInterceptor,
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
		return nil, err
	}

	epochStartData, err := esdp.getCurrentEpochStartData(shardCoordinator, metaBlock)
	if err != nil {
		return nil, err
	}

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
		if epochStartData.ShardId == shardID {
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

	err = esdp.messenger.RegisterMessageProcessor(factory.MetachainBlocksTopic, esdp.metaBlockInterceptor)
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
	err := esdp.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
	if err != nil {
		log.Info("error unregistering message processor", "error", err)
	}
}

func (esdp *epochStartDataProvider) getEpochStartMetaBlock(epoch uint32) (*block.MetaBlock, error) {
	esdp.requestMetaBlock(epoch)

	time.Sleep(delayAfterRequesting)

	for {
		numConnectedPeers := len(esdp.messenger.Peers())
		threshold := int(thresholdForConsideringMetaBlockCorrect * float64(numConnectedPeers))
		mb, errConsensusNotReached := esdp.metaBlockInterceptor.GetMetaBlock(threshold, epoch)
		if errConsensusNotReached == nil {
			return mb, nil
		}
		log.Info("consensus not reached for epoch start meta block. re-requesting and trying again...")
		esdp.requestMetaBlock(epoch)
	}
}

func (esdp *epochStartDataProvider) getShardCoordinator(metaBlock *block.MetaBlock, nodesConfig *sharding.NodesSetup) (sharding.Coordinator, error) {
	shardID, err := esdp.getShardID(nodesConfig)
	if err != nil {
		return nil, err
	}

	numOfShards := len(metaBlock.EpochStart.LastFinalizedHeaders)
	if numOfShards == 1 {
		return &sharding.OneShardCoordinator{}, nil
	}

	return sharding.NewMultiShardCoordinator(uint32(numOfShards), shardID)
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
			hdr, err := esdp.getShardHeader(entry.HeaderHash, entry.ShardId)
			if err != nil {
				return nil, err
			}
			headersMap[entry.ShardId] = hdr
		}

		return headersMap, nil
	}

	var entryForShard *block.EpochStartShardData
	for _, entry := range metaBlock.EpochStart.LastFinalizedHeaders {
		if entry.ShardId == shardID {
			entryForShard = &entry
		}
	}

	if entryForShard == nil {
		return nil, errors.New("shard data not found")
	}

	hdr, err := esdp.getShardHeader(
		entryForShard.HeaderHash,
		entryForShard.ShardId,
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

func (esdp *epochStartDataProvider) requestShardHeader(shardID uint32, hash []byte) {
	// send more requests
	log.Debug("requsted shard block", "shard ID", shardID, "hash", hash)
	for i := 0; i < numRequestsToSendOnce; i++ {
		time.Sleep(delayBetweenRequests)
		esdp.requestHandlerMeta.RequestShardHeader(shardID, hash)
	}
}

func (esdp *epochStartDataProvider) requestMetaBlock(epoch uint32) {
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
