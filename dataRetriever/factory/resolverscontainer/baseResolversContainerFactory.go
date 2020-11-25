package resolverscontainer

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer/disabled"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// EmptyExcludePeersOnTopic is an empty topic
const EmptyExcludePeersOnTopic = ""
const defaultTargetShardID = uint32(0)

//TODO extract these in config
const numCrossShardPeers = 2
const numIntraShardPeers = 1
const numFullHistoryPeers = 3

type baseResolversContainerFactory struct {
	container                dataRetriever.ResolversContainer
	shardCoordinator         sharding.Coordinator
	messenger                dataRetriever.TopicMessageHandler
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	dataPools                dataRetriever.PoolsHolder
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	intRandomizer            dataRetriever.IntRandomizer
	dataPacker               dataRetriever.DataPacker
	triesContainer           state.TriesHolder
	inputAntifloodHandler    dataRetriever.P2PAntifloodHandler
	outputAntifloodHandler   dataRetriever.P2PAntifloodHandler
	throttler                dataRetriever.ResolverThrottler
	intraShardTopic          string
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

		resolver, err := brcf.createTxResolver(identifierTx, excludePeersFromTopic, unit, dataPool)
		if err != nil {
			return err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierTx
	}

	identifierTx := topic + shardC.CommunicationIdentifier(core.MetachainShardId)
	excludePeersFromTopic := topic + shardC.CommunicationIdentifier(shardC.SelfId())

	resolver, err := brcf.createTxResolver(identifierTx, excludePeersFromTopic, unit, dataPool)
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
) (dataRetriever.Resolver, error) {

	txStorer := brcf.store.GetStorer(unit)

	resolverSender, err := brcf.createOneResolverSender(topic, excludedTopic, defaultTargetShardID)
	if err != nil {
		return nil, err
	}

	arg := resolvers.ArgTxResolver{
		SenderResolver:   resolverSender,
		TxPool:           dataPool,
		TxStorage:        txStorer,
		Marshalizer:      brcf.marshalizer,
		DataPacker:       brcf.dataPacker,
		AntifloodHandler: brcf.inputAntifloodHandler,
		Throttler:        brcf.throttler,
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

		resolver, err := brcf.createMiniBlocksResolver(identifierMiniBlocks, excludePeersFromTopic)
		if err != nil {
			return err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierMiniBlocks
	}

	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	excludePeersFromTopic := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(shardC.SelfId())

	resolver, err := brcf.createMiniBlocksResolver(identifierMiniBlocks, excludePeersFromTopic)
	if err != nil {
		return err
	}

	resolverSlice[noOfShards] = resolver
	keys[noOfShards] = identifierMiniBlocks

	identifierAllShardMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.AllShardId)
	allShardMiniblocksResolver, err := brcf.createMiniBlocksResolver(identifierAllShardMiniBlocks, EmptyExcludePeersOnTopic)
	if err != nil {
		return err
	}

	resolverSlice[noOfShards+1] = allShardMiniblocksResolver
	keys[noOfShards+1] = identifierAllShardMiniBlocks

	return brcf.container.AddMultiple(keys, resolverSlice)
}

func (brcf *baseResolversContainerFactory) createMiniBlocksResolver(topic string, excludedTopic string) (dataRetriever.Resolver, error) {
	miniBlocksStorer := brcf.store.GetStorer(dataRetriever.MiniBlockUnit)

	resolverSender, err := brcf.createOneResolverSender(topic, excludedTopic, defaultTargetShardID)
	if err != nil {
		return nil, err
	}

	arg := resolvers.ArgMiniblockResolver{
		SenderResolver:   resolverSender,
		MiniBlockPool:    brcf.dataPools.MiniBlocks(),
		MiniBlockStorage: miniBlocksStorer,
		Marshalizer:      brcf.marshalizer,
		AntifloodHandler: brcf.inputAntifloodHandler,
		Throttler:        brcf.throttler,
		DataPacker:       brcf.dataPacker,
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
	return brcf.createOneResolverSenderWithSpecifiedNumRequests(topic, excludedTopic, targetShardId,
		numCrossShardPeers, numIntraShardPeers, numFullHistoryPeers, &disabled.NilCurrentNetworkEpochProviderHandler{})
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
	currentNetworkEpochProviderHandler dataRetriever.CurrentNetworkEpochProviderHandler,
) (dataRetriever.Resolver, error) {
	resolverSender, err := brcf.createOneResolverSenderWithSpecifiedNumRequests(
		topic,
		EmptyExcludePeersOnTopic,
		defaultTargetShardID,
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
