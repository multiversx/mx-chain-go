package resolverscontainer

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// EmptyExcludePeersOnTopic is an empty topic
const EmptyExcludePeersOnTopic = ""

var log = logger.GetOrCreate("dataRetriever/factory/resolverscontainer")

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
	triesContainer              common.TriesHolder
	inputAntifloodHandler       dataRetriever.P2PAntifloodHandler
	outputAntifloodHandler      dataRetriever.P2PAntifloodHandler
	throttler                   dataRetriever.ResolverThrottler
	intraShardTopic             string
	isFullHistoryNode           bool
	currentNetworkEpochProvider dataRetriever.CurrentNetworkEpochProviderHandler
	preferredPeersHolder        dataRetriever.PreferredPeersHolderHandler
	peersRatingHandler          dataRetriever.PeersRatingHandler
	numCrossShardPeers          int
	numIntraShardPeers          int
	numTotalPeers               int
	numFullHistoryPeers         int
	payloadValidator            dataRetriever.PeerAuthenticationPayloadValidator
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
	if check.IfNil(brcf.preferredPeersHolder) {
		return dataRetriever.ErrNilPreferredPeersHolder
	}
	if check.IfNil(brcf.peersRatingHandler) {
		return dataRetriever.ErrNilPeersRatingHandler
	}
	if brcf.numCrossShardPeers <= 0 {
		return fmt.Errorf("%w for numCrossShardPeers", dataRetriever.ErrInvalidValue)
	}
	if brcf.numTotalPeers <= brcf.numCrossShardPeers {
		return fmt.Errorf("%w for numTotalPeers", dataRetriever.ErrInvalidValue)
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

		resolver, err := brcf.createTxResolver(identifierTx, excludePeersFromTopic, unit, dataPool, idx, brcf.numCrossShardPeers, brcf.numIntraShardPeers)
		if err != nil {
			return err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierTx
	}

	identifierTx := topic + shardC.CommunicationIdentifier(core.MetachainShardId)
	excludePeersFromTopic := topic + shardC.CommunicationIdentifier(shardC.SelfId())

	resolver, err := brcf.createTxResolver(identifierTx, excludePeersFromTopic, unit, dataPool, core.MetachainShardId, brcf.numCrossShardPeers, brcf.numIntraShardPeers)
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
	numCrossShardPeers int,
	numIntraShardPeers int,
) (dataRetriever.Resolver, error) {

	txStorer, err := brcf.store.GetStorer(unit)
	if err != nil {
		return nil, err
	}

	resolverSender, err := brcf.createOneResolverSenderWithSpecifiedNumRequests(topic, excludedTopic, targetShardID, numCrossShardPeers, numIntraShardPeers)
	if err != nil {
		return nil, err
	}

	arg := resolvers.ArgTxResolver{
		ArgBaseResolver: resolvers.ArgBaseResolver{
			SenderResolver:   resolverSender,
			Marshaller:       brcf.marshalizer,
			AntifloodHandler: brcf.inputAntifloodHandler,
			Throttler:        brcf.throttler,
		},
		TxPool:            dataPool,
		TxStorage:         txStorer,
		DataPacker:        brcf.dataPacker,
		IsFullHistoryNode: brcf.isFullHistoryNode,
	}
	resolver, err := resolvers.NewTxResolver(arg)
	if err != nil {
		return nil, err
	}

	err = brcf.messenger.RegisterMessageProcessor(resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
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

		resolver, err := brcf.createMiniBlocksResolver(identifierMiniBlocks, excludePeersFromTopic, idx, brcf.numCrossShardPeers, brcf.numIntraShardPeers)
		if err != nil {
			return err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierMiniBlocks
	}

	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	excludePeersFromTopic := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(shardC.SelfId())

	resolver, err := brcf.createMiniBlocksResolver(identifierMiniBlocks, excludePeersFromTopic, core.MetachainShardId, brcf.numCrossShardPeers, brcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	resolverSlice[noOfShards] = resolver
	keys[noOfShards] = identifierMiniBlocks

	identifierAllShardMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.AllShardId)
	allShardMiniblocksResolver, err := brcf.createMiniBlocksResolver(identifierAllShardMiniBlocks, EmptyExcludePeersOnTopic, brcf.shardCoordinator.SelfId(), brcf.numCrossShardPeers, brcf.numIntraShardPeers)
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
	numCrossShardPeers int,
	numIntraShardPeers int,
) (dataRetriever.Resolver, error) {
	miniBlocksStorer, err := brcf.store.GetStorer(dataRetriever.MiniBlockUnit)
	if err != nil {
		return nil, err
	}

	resolverSender, err := brcf.createOneResolverSenderWithSpecifiedNumRequests(topic, excludedTopic, targetShardID, numCrossShardPeers, numIntraShardPeers)
	if err != nil {
		return nil, err
	}

	arg := resolvers.ArgMiniblockResolver{
		ArgBaseResolver: resolvers.ArgBaseResolver{
			SenderResolver:   resolverSender,
			Marshaller:       brcf.marshalizer,
			AntifloodHandler: brcf.inputAntifloodHandler,
			Throttler:        brcf.throttler,
		},
		MiniBlockPool:     brcf.dataPools.MiniBlocks(),
		MiniBlockStorage:  miniBlocksStorer,
		DataPacker:        brcf.dataPacker,
		IsFullHistoryNode: brcf.isFullHistoryNode,
	}
	txBlkResolver, err := resolvers.NewMiniblockResolver(arg)
	if err != nil {
		return nil, err
	}

	err = brcf.messenger.RegisterMessageProcessor(txBlkResolver.RequestTopic(), common.DefaultResolversIdentifier, txBlkResolver)
	if err != nil {
		return nil, err
	}

	return txBlkResolver, nil
}

func (brcf *baseResolversContainerFactory) generatePeerAuthenticationResolver() error {
	identifierPeerAuth := common.PeerAuthenticationTopic
	shardC := brcf.shardCoordinator
	resolverSender, err := brcf.createOneResolverSenderWithSpecifiedNumRequests(identifierPeerAuth, EmptyExcludePeersOnTopic, shardC.SelfId(), brcf.numCrossShardPeers, brcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	arg := resolvers.ArgPeerAuthenticationResolver{
		ArgBaseResolver: resolvers.ArgBaseResolver{
			SenderResolver:   resolverSender,
			Marshaller:       brcf.marshalizer,
			AntifloodHandler: brcf.inputAntifloodHandler,
			Throttler:        brcf.throttler,
		},
		PeerAuthenticationPool: brcf.dataPools.PeerAuthentications(),
		DataPacker:             brcf.dataPacker,
		PayloadValidator:       brcf.payloadValidator,
	}
	peerAuthResolver, err := resolvers.NewPeerAuthenticationResolver(arg)
	if err != nil {
		return err
	}

	err = brcf.messenger.RegisterMessageProcessor(peerAuthResolver.RequestTopic(), common.DefaultResolversIdentifier, peerAuthResolver)
	if err != nil {
		return err
	}

	return brcf.container.Add(identifierPeerAuth, peerAuthResolver)
}

func (brcf *baseResolversContainerFactory) createOneResolverSenderWithSpecifiedNumRequests(
	topic string,
	excludedTopic string,
	targetShardId uint32,
	numCrossShardPeers int,
	numIntraShardPeers int,
) (dataRetriever.TopicResolverSender, error) {

	log.Trace("baseResolversContainerFactory.createOneResolverSenderWithSpecifiedNumRequests",
		"topic", topic, "intraShardTopic", brcf.intraShardTopic, "excludedTopic", excludedTopic,
		"numCrossShardPeers", numCrossShardPeers, "numIntraShardPeers", numIntraShardPeers)

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
		NumCrossShardPeers:          numCrossShardPeers,
		NumIntraShardPeers:          numIntraShardPeers,
		NumFullHistoryPeers:         brcf.numFullHistoryPeers,
		CurrentNetworkEpochProvider: brcf.currentNetworkEpochProvider,
		PreferredPeersHolder:        brcf.preferredPeersHolder,
		SelfShardIdProvider:         brcf.shardCoordinator,
		PeersRatingHandler:          brcf.peersRatingHandler,
	}
	// TODO instantiate topic sender resolver with the shard IDs for which this resolver is supposed to serve the data
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
	numCrossShardPeers int,
	numIntraShardPeers int,
	targetShardID uint32,
) (dataRetriever.Resolver, error) {
	resolverSender, err := brcf.createOneResolverSenderWithSpecifiedNumRequests(
		topic,
		EmptyExcludePeersOnTopic,
		targetShardID,
		numCrossShardPeers,
		numIntraShardPeers,
	)
	if err != nil {
		return nil, err
	}

	trie := brcf.triesContainer.Get([]byte(trieId))
	argTrie := resolvers.ArgTrieNodeResolver{
		ArgBaseResolver: resolvers.ArgBaseResolver{
			SenderResolver:   resolverSender,
			Marshaller:       brcf.marshalizer,
			AntifloodHandler: brcf.inputAntifloodHandler,
			Throttler:        brcf.throttler,
		},
		TrieDataGetter: trie,
	}
	resolver, err := resolvers.NewTrieNodeResolver(argTrie)
	if err != nil {
		return nil, err
	}

	err = brcf.messenger.RegisterMessageProcessor(resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func (brcf *baseResolversContainerFactory) generateValidatorInfoResolver() error {
	identifierValidatorInfo := common.ValidatorInfoTopic
	shardC := brcf.shardCoordinator
	resolverSender, err := brcf.createOneResolverSenderWithSpecifiedNumRequests(identifierValidatorInfo, EmptyExcludePeersOnTopic, shardC.SelfId(), brcf.numCrossShardPeers, brcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	validatorInfoStorage, err := brcf.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	if err != nil {
		return err
	}

	arg := resolvers.ArgValidatorInfoResolver{
		SenderResolver:       resolverSender,
		Marshaller:           brcf.marshalizer,
		AntifloodHandler:     brcf.inputAntifloodHandler,
		Throttler:            brcf.throttler,
		ValidatorInfoPool:    brcf.dataPools.ValidatorsInfo(),
		ValidatorInfoStorage: validatorInfoStorage,
		DataPacker:           brcf.dataPacker,
		IsFullHistoryNode:    brcf.isFullHistoryNode,
	}
	validatorInfoResolver, err := resolvers.NewValidatorInfoResolver(arg)
	if err != nil {
		return err
	}

	err = brcf.messenger.RegisterMessageProcessor(validatorInfoResolver.RequestTopic(), common.DefaultResolversIdentifier, validatorInfoResolver)
	if err != nil {
		return err
	}

	return brcf.container.Add(identifierValidatorInfo, validatorInfoResolver)
}
