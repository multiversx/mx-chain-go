package resolverscontainer

import (
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

// numPeersToQuery number of peers to send the message
const numPeersToQuery = 2

const emptyExcludePeersOnTopic = ""

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

	return nil
}

func (brcf *baseResolversContainerFactory) createTopicAndAssignHandler(
	topicName string,
	resolver dataRetriever.Resolver,
	createChannel bool,
) (dataRetriever.Resolver, error) {

	err := brcf.messenger.CreateTopic(topicName, createChannel)
	if err != nil {
		return nil, err
	}

	return resolver, brcf.messenger.RegisterMessageProcessor(topicName, resolver)
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

	resolverSender, err := brcf.createOneResolverSender(topic, excludedTopic)
	if err != nil {
		return nil, err
	}

	resolver, err := resolvers.NewTxResolver(
		resolverSender,
		dataPool,
		txStorer,
		brcf.marshalizer,
		brcf.dataPacker,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return brcf.createTopicAndAssignHandler(
		topic+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}

func (brcf *baseResolversContainerFactory) generateMiniBlocksResolvers() error {
	shardC := brcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards+1)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards+1)

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

	return brcf.container.AddMultiple(keys, resolverSlice)
}

func (brcf *baseResolversContainerFactory) createMiniBlocksResolver(topic string, excludedTopic string) (dataRetriever.Resolver, error) {
	miniBlocksStorer := brcf.store.GetStorer(dataRetriever.MiniBlockUnit)

	resolverSender, err := brcf.createOneResolverSender(topic, excludedTopic)
	if err != nil {
		return nil, err
	}

	txBlkResolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		brcf.dataPools.MiniBlocks(),
		miniBlocksStorer,
		brcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return brcf.createTopicAndAssignHandler(
		topic+resolverSender.TopicRequestSuffix(),
		txBlkResolver,
		false)
}

func (brcf *baseResolversContainerFactory) createOneResolverSender(
	topic string,
	excludedTopic string,
) (dataRetriever.TopicResolverSender, error) {

	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(brcf.messenger, topic, excludedTopic)
	if err != nil {
		return nil, err
	}

	//TODO instantiate topic sender resolver with the shard IDs for which this resolver is supposed to serve the data
	// this will improve the serving of transactions as the searching will be done only on 2 sharded data units
	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		brcf.messenger,
		topic,
		peerListCreator,
		brcf.marshalizer,
		brcf.intRandomizer,
		numPeersToQuery,
		uint32(0),
	)
	if err != nil {
		return nil, err
	}

	return resolverSender, nil
}

func (brcf *baseResolversContainerFactory) createTrieNodesResolver(topic string, trieId string) (dataRetriever.Resolver, error) {
	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(brcf.messenger, topic, emptyExcludePeersOnTopic)
	if err != nil {
		return nil, err
	}

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		brcf.messenger,
		topic,
		peerListCreator,
		brcf.marshalizer,
		brcf.intRandomizer,
		numPeersToQuery,
		brcf.shardCoordinator.SelfId(),
	)
	if err != nil {
		return nil, err
	}

	trie := brcf.triesContainer.Get([]byte(trieId))
	resolver, err := resolvers.NewTrieNodeResolver(
		resolverSender,
		trie,
		brcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return brcf.createTopicAndAssignHandler(
		topic+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}
