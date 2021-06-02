package resolverscontainer

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// EmptyExcludePeersOnTopic is an empty topic
const EmptyExcludePeersOnTopic = ""

type baseResolversContainerFactory struct {
	container                   dataRetriever.ResolversContainer
	shardCoordinator            sharding.Coordinator
	messenger                   dataRetriever.TopicMessageHandler
	store                       dataRetriever.StorageService
	marshalizer                 marshal.Marshalizer
	dataPools                   dataRetriever.PoolsHolder
	uint64ByteSliceConverter    typeConverters.Uint64ByteSliceConverter
	intRandomizer               dataRetriever.IntRandomizer
	dataPacker                  dataRetriever.DataPacker
	triesContainer              state.TriesHolder
	inputAntifloodHandler       dataRetriever.P2PAntifloodHandler
	outputAntifloodHandler      dataRetriever.P2PAntifloodHandler
	throttler                   dataRetriever.ResolverThrottler
	intraShardTopic             string
	isFullHistoryNode           bool
	currentNetworkEpochProvider dataRetriever.CurrentNetworkEpochProviderHandler
	numCrossShardPeers          int
	numIntraShardPeers          int
	numFullHistoryPeers         int
}

func (brcf *baseResolversContainerFactory) checkParams() error {
	if check.IfNil(brcf.shardCoordinator) {
		return dataRetriever.ErrNilShardCoordinator
	}
	if check.IfNil(brcf.messenger) {
		return dataRetriever.ErrNilMessenger
	}
	if check.IfNil(brcf.store) {
		return dataRetriever.ErrNilStore
	}
	if check.IfNil(brcf.marshalizer) {
		return dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(brcf.dataPools) {
		return dataRetriever.ErrNilDataPoolHolder
	}
	if check.IfNil(brcf.uint64ByteSliceConverter) {
		return dataRetriever.ErrNilUint64ByteSliceConverter
	}
	if check.IfNil(brcf.dataPacker) {
		return dataRetriever.ErrNilDataPacker
	}
	if check.IfNil(brcf.triesContainer) {
		return dataRetriever.ErrNilTrieDataGetter
	}
	if check.IfNil(brcf.inputAntifloodHandler) {
		return fmt.Errorf("%w for input", dataRetriever.ErrNilAntifloodHandler)
	}
	if check.IfNil(brcf.outputAntifloodHandler) {
		return fmt.Errorf("%w for output", dataRetriever.ErrNilAntifloodHandler)
	}
	if check.IfNil(brcf.throttler) {
		return dataRetriever.ErrNilThrottler
	}
	if check.IfNil(brcf.currentNetworkEpochProvider) {
		return dataRetriever.ErrNilCurrentNetworkEpochProvider
	}
	if brcf.numCrossShardPeers <= 0 {
		return fmt.Errorf("%w for numCrossShardPeers", dataRetriever.ErrInvalidValue)
	}
	if brcf.numIntraShardPeers <= 0 {
		return fmt.Errorf("%w for numIntraShardPeers", dataRetriever.ErrInvalidValue)
	}
	if brcf.numFullHistoryPeers <= 0 {
		return fmt.Errorf("%w for numFullHistoryPeers", dataRetriever.ErrInvalidValue)
	}

	return nil
}

func (brcf *baseResolversContainerFactory) generateTxResolvers(
	topic string,
	unit dataRetriever.UnitType,
	dataPool dataRetriever.ShardedDataCacherNotifier,
) error {

	shardC := brcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards+1)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards+1)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := topic + shardC.CommunicationIdentifier(idx)
		excludePeersFromTopic := topic + shardC.CommunicationIdentifier(shardC.SelfId())

		resolver, err := brcf.createTxResolver(identifierTx, excludePeersFromTopic, unit, dataPool, idx)
		if err != nil {
			return err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierTx
	}

	identifierTx := topic + shardC.CommunicationIdentifier(core.MetachainShardId)
	excludePeersFromTopic := topic + shardC.CommunicationIdentifier(shardC.SelfId())

	resolver, err := brcf.createTxResolver(identifierTx, excludePeersFromTopic, unit, dataPool, core.MetachainShardId)
	if err != nil {
		return err
	}

	resolverSlice[noOfShards] = resolver
	keys[noOfShards] = identifierTx

	return brcf.container.AddMultiple(keys, resolverSlice)
}

func (brcf *baseResolversContainerFactory) createTxResolver(
	topic string,
	excludedTopic string,
	unit dataRetriever.UnitType,
	dataPool dataRetriever.ShardedDataCacherNotifier,
	targetShardID uint32,
) (dataRetriever.Resolver, error) {

	txStorer := brcf.store.GetStorer(unit)

	resolverSender, err := brcf.createOneResolverSender(topic, excludedTopic, targetShardID)
	if err != nil {
		return nil, err
	}

	arg := resolvers.ArgTxResolver{
		SenderResolver:    resolverSender,
		TxPool:            dataPool,
		TxStorage:         txStorer,
		Marshalizer:       brcf.marshalizer,
		DataPacker:        brcf.dataPacker,
		AntifloodHandler:  brcf.inputAntifloodHandler,
		Throttler:         brcf.throttler,
		IsFullHistoryNode: brcf.isFullHistoryNode,
	}
	resolver, err := resolvers.NewTxResolver(arg)
	if err != nil {
		return nil, err
	}

	err = brcf.messenger.RegisterMessageProcessor(resolver.RequestTopic(), core.DefaultResolversIdentifier, resolver)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func (brcf *baseResolversContainerFactory) generateMiniBlocksResolvers() error {
	shardC := brcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards+2)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards+2)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)
		excludePeersFromTopic := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(shardC.SelfId())

		resolver, err := brcf.createMiniBlocksResolver(identifierMiniBlocks, excludePeersFromTopic, idx)
		if err != nil {
			return err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierMiniBlocks
	}

	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	excludePeersFromTopic := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(shardC.SelfId())

	resolver, err := brcf.createMiniBlocksResolver(identifierMiniBlocks, excludePeersFromTopic, core.MetachainShardId)
	if err != nil {
		return err
	}

	resolverSlice[noOfShards] = resolver
	keys[noOfShards] = identifierMiniBlocks

	identifierAllShardMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.AllShardId)
	allShardMiniblocksResolver, err := brcf.createMiniBlocksResolver(identifierAllShardMiniBlocks, EmptyExcludePeersOnTopic, brcf.shardCoordinator.SelfId())
	if err != nil {
		return err
	}

	resolverSlice[noOfShards+1] = allShardMiniblocksResolver
	keys[noOfShards+1] = identifierAllShardMiniBlocks

	return brcf.container.AddMultiple(keys, resolverSlice)
}

func (brcf *baseResolversContainerFactory) createMiniBlocksResolver(
	topic string,
	excludedTopic string,
	targetShardID uint32,
) (dataRetriever.Resolver, error) {
	miniBlocksStorer := brcf.store.GetStorer(dataRetriever.MiniBlockUnit)

	resolverSender, err := brcf.createOneResolverSender(topic, excludedTopic, targetShardID)
	if err != nil {
		return nil, err
	}

	arg := resolvers.ArgMiniblockResolver{
		SenderResolver:    resolverSender,
		MiniBlockPool:     brcf.dataPools.MiniBlocks(),
		MiniBlockStorage:  miniBlocksStorer,
		Marshalizer:       brcf.marshalizer,
		AntifloodHandler:  brcf.inputAntifloodHandler,
		Throttler:         brcf.throttler,
		DataPacker:        brcf.dataPacker,
		IsFullHistoryNode: brcf.isFullHistoryNode,
	}
	txBlkResolver, err := resolvers.NewMiniblockResolver(arg)
	if err != nil {
		return nil, err
	}

	err = brcf.messenger.RegisterMessageProcessor(txBlkResolver.RequestTopic(), core.DefaultResolversIdentifier, txBlkResolver)
	if err != nil {
		return nil, err
	}

	return txBlkResolver, nil
}

func (brcf *baseResolversContainerFactory) createOneResolverSender(
	topic string,
	excludedTopic string,
	targetShardId uint32,
) (dataRetriever.TopicResolverSender, error) {
	return brcf.createOneResolverSenderWithSpecifiedNumRequests(
		topic,
		excludedTopic,
		targetShardId,
		brcf.numCrossShardPeers,
		brcf.numIntraShardPeers,
		brcf.numFullHistoryPeers,
		brcf.currentNetworkEpochProvider,
	)
}

func (brcf *baseResolversContainerFactory) createOneResolverSenderWithSpecifiedNumRequests(
	topic string,
	excludedTopic string,
	targetShardId uint32,
	numCrossShard int,
	numIntraShard int,
	numFullHistory int,
	currentNetworkEpochProvider dataRetriever.CurrentNetworkEpochProviderHandler,
) (dataRetriever.TopicResolverSender, error) {

	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(brcf.messenger, topic, brcf.intraShardTopic, excludedTopic)
	if err != nil {
		return nil, err
	}

	arg := topicResolverSender.ArgTopicResolverSender{
		Messenger:                   brcf.messenger,
		TopicName:                   topic,
		PeerListCreator:             peerListCreator,
		Marshalizer:                 brcf.marshalizer,
		Randomizer:                  brcf.intRandomizer,
		TargetShardId:               targetShardId,
		OutputAntiflooder:           brcf.outputAntifloodHandler,
		NumCrossShardPeers:          numCrossShard,
		NumIntraShardPeers:          numIntraShard,
		NumFullHistoryPeers:         numFullHistory,
		CurrentNetworkEpochProvider: currentNetworkEpochProvider,
	}
	//TODO instantiate topic sender resolver with the shard IDs for which this resolver is supposed to serve the data
	// this will improve the serving of transactions as the searching will be done only on 2 sharded data units
	resolverSender, err := topicResolverSender.NewTopicResolverSender(arg)
	if err != nil {
		return nil, err
	}

	return resolverSender, nil
}

func (brcf *baseResolversContainerFactory) createTrieNodesResolver(
	topic string,
	trieId string,
	numCrossShard int,
	numIntraShard int,
	numFullHistory int,
	targetShardID uint32,
	currentNetworkEpochProviderHandler dataRetriever.CurrentNetworkEpochProviderHandler,
) (dataRetriever.Resolver, error) {
	resolverSender, err := brcf.createOneResolverSenderWithSpecifiedNumRequests(
		topic,
		EmptyExcludePeersOnTopic,
		targetShardID,
		numCrossShard,
		numIntraShard,
		numFullHistory,
		currentNetworkEpochProviderHandler,
	)
	if err != nil {
		return nil, err
	}

	trie := brcf.triesContainer.Get([]byte(trieId))
	argTrie := resolvers.ArgTrieNodeResolver{
		SenderResolver:   resolverSender,
		TrieDataGetter:   trie,
		Marshalizer:      brcf.marshalizer,
		AntifloodHandler: brcf.inputAntifloodHandler,
		Throttler:        brcf.throttler,
	}
	resolver, err := resolvers.NewTrieNodeResolver(argTrie)
	if err != nil {
		return nil, err
	}

	err = brcf.messenger.RegisterMessageProcessor(resolver.RequestTopic(), core.DefaultResolversIdentifier, resolver)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}
