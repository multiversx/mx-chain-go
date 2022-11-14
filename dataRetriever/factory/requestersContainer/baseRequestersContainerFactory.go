package requesterscontainer

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers/requesters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/topicSender"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// EmptyExcludePeersOnTopic is an empty topic
const EmptyExcludePeersOnTopic = ""

var log = logger.GetOrCreate("dataRetriever/factory/requesterscontainer")

type baseRequestersContainerFactory struct {
	container                   dataRetriever.RequestersContainer
	shardCoordinator            sharding.Coordinator
	messenger                   dataRetriever.TopicMessageHandler
	marshaller                  marshal.Marshalizer
	uint64ByteSliceConverter    typeConverters.Uint64ByteSliceConverter
	intRandomizer               dataRetriever.IntRandomizer
	outputAntifloodHandler      dataRetriever.P2PAntifloodHandler
	intraShardTopic             string
	currentNetworkEpochProvider dataRetriever.CurrentNetworkEpochProviderHandler
	preferredPeersHolder        dataRetriever.PreferredPeersHolderHandler
	peersRatingHandler          dataRetriever.PeersRatingHandler
	numCrossShardPeers          int
	numIntraShardPeers          int
	numTotalPeers               int
	numFullHistoryPeers         int
}

func (brcf *baseRequestersContainerFactory) checkParams() error {
	if check.IfNil(brcf.shardCoordinator) {
		return dataRetriever.ErrNilShardCoordinator
	}
	if check.IfNil(brcf.messenger) {
		return dataRetriever.ErrNilMessenger
	}
	if check.IfNil(brcf.marshaller) {
		return dataRetriever.ErrNilMarshaller
	}
	if check.IfNil(brcf.uint64ByteSliceConverter) {
		return dataRetriever.ErrNilUint64ByteSliceConverter
	}
	if check.IfNil(brcf.outputAntifloodHandler) {
		return fmt.Errorf("%w for output", dataRetriever.ErrNilAntifloodHandler)
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

func (brcf *baseRequestersContainerFactory) generateTxRequesters(topic string) error {

	shardC := brcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards+1)
	requestersSlice := make([]dataRetriever.Requester, noOfShards+1)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := topic + shardC.CommunicationIdentifier(idx)
		excludePeersFromTopic := topic + shardC.CommunicationIdentifier(shardC.SelfId())

		requester, err := brcf.createTxRequester(identifierTx, excludePeersFromTopic, idx, brcf.numCrossShardPeers, brcf.numIntraShardPeers)
		if err != nil {
			return err
		}

		requestersSlice[idx] = requester
		keys[idx] = identifierTx
	}

	identifierTx := topic + shardC.CommunicationIdentifier(core.MetachainShardId)
	excludePeersFromTopic := topic + shardC.CommunicationIdentifier(shardC.SelfId())

	requester, err := brcf.createTxRequester(identifierTx, excludePeersFromTopic, core.MetachainShardId, brcf.numCrossShardPeers, brcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	requestersSlice[noOfShards] = requester
	keys[noOfShards] = identifierTx

	return brcf.container.AddMultiple(keys, requestersSlice)
}

func (brcf *baseRequestersContainerFactory) createTxRequester(
	topic string,
	excludedTopic string,
	targetShardID uint32,
	numCrossShardPeers int,
	numIntraShardPeers int,
) (dataRetriever.Requester, error) {

	requestSender, err := brcf.createOneRequestSenderWithSpecifiedNumRequests(topic, excludedTopic, targetShardID, numCrossShardPeers, numIntraShardPeers)
	if err != nil {
		return nil, err
	}

	arg := requesters.ArgTransactionRequester{
		ArgBaseRequester: requesters.ArgBaseRequester{
			RequestSender: requestSender,
			Marshaller:    brcf.marshaller,
		},
	}
	return requesters.NewTransactionRequester(arg)
}

func (brcf *baseRequestersContainerFactory) generateMiniBlocksRequesters() error {
	shardC := brcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards+2)
	requestersSlice := make([]dataRetriever.Requester, noOfShards+2)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)
		excludePeersFromTopic := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(shardC.SelfId())

		requester, err := brcf.createMiniBlocksRequester(identifierMiniBlocks, excludePeersFromTopic, idx, brcf.numCrossShardPeers, brcf.numIntraShardPeers)
		if err != nil {
			return err
		}

		requestersSlice[idx] = requester
		keys[idx] = identifierMiniBlocks
	}

	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	excludePeersFromTopic := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(shardC.SelfId())

	requester, err := brcf.createMiniBlocksRequester(identifierMiniBlocks, excludePeersFromTopic, core.MetachainShardId, brcf.numCrossShardPeers, brcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	requestersSlice[noOfShards] = requester
	keys[noOfShards] = identifierMiniBlocks

	identifierAllShardMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.AllShardId)
	allShardMiniblocksResolver, err := brcf.createMiniBlocksRequester(identifierAllShardMiniBlocks, EmptyExcludePeersOnTopic, brcf.shardCoordinator.SelfId(), brcf.numCrossShardPeers, brcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	requestersSlice[noOfShards+1] = allShardMiniblocksResolver
	keys[noOfShards+1] = identifierAllShardMiniBlocks

	return brcf.container.AddMultiple(keys, requestersSlice)
}

func (brcf *baseRequestersContainerFactory) createMiniBlocksRequester(
	topic string,
	excludedTopic string,
	targetShardID uint32,
	numCrossShardPeers int,
	numIntraShardPeers int,
) (dataRetriever.Requester, error) {
	requestSender, err := brcf.createOneRequestSenderWithSpecifiedNumRequests(topic, excludedTopic, targetShardID, numCrossShardPeers, numIntraShardPeers)
	if err != nil {
		return nil, err
	}

	arg := requesters.ArgMiniblockRequester{
		ArgBaseRequester: requesters.ArgBaseRequester{
			RequestSender: requestSender,
			Marshaller:    brcf.marshaller,
		},
	}
	return requesters.NewMiniblockRequester(arg)
}

func (brcf *baseRequestersContainerFactory) generatePeerAuthenticationRequester() error {
	identifierPeerAuth := common.PeerAuthenticationTopic
	shardC := brcf.shardCoordinator
	requestSender, err := brcf.createOneRequestSenderWithSpecifiedNumRequests(identifierPeerAuth, EmptyExcludePeersOnTopic, shardC.SelfId(), brcf.numCrossShardPeers, brcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	arg := requesters.ArgPeerAuthenticationRequester{
		ArgBaseRequester: requesters.ArgBaseRequester{
			RequestSender: requestSender,
			Marshaller:    brcf.marshaller,
		},
	}
	requester, err := requesters.NewPeerAuthenticationRequester(arg)
	if err != nil {
		return err
	}

	return brcf.container.Add(identifierPeerAuth, requester)
}

func (brcf *baseRequestersContainerFactory) createOneRequestSenderWithSpecifiedNumRequests(
	topic string,
	excludedTopic string,
	targetShardId uint32,
	numCrossShardPeers int,
	numIntraShardPeers int,
) (dataRetriever.TopicRequestSender, error) {

	log.Trace("baseRequestersContainerFactory.createOneRequestSenderWithSpecifiedNumRequests",
		"topic", topic, "intraShardTopic", brcf.intraShardTopic, "excludedTopic", excludedTopic,
		"numCrossShardPeers", numCrossShardPeers, "numIntraShardPeers", numIntraShardPeers)

	peerListCreator, err := topicSender.NewDiffPeerListCreator(brcf.messenger, topic, brcf.intraShardTopic, excludedTopic)
	if err != nil {
		return nil, err
	}

	arg := topicSender.ArgTopicRequestSender{
		ArgBaseTopicSender: topicSender.ArgBaseTopicSender{
			Messenger:            brcf.messenger,
			TopicName:            topic,
			OutputAntiflooder:    brcf.outputAntifloodHandler,
			PreferredPeersHolder: brcf.preferredPeersHolder,
			TargetShardId:        targetShardId,
		},
		Marshaller:                  brcf.marshaller,
		Randomizer:                  brcf.intRandomizer,
		PeerListCreator:             peerListCreator,
		NumIntraShardPeers:          numIntraShardPeers,
		NumCrossShardPeers:          numCrossShardPeers,
		NumFullHistoryPeers:         brcf.numFullHistoryPeers,
		CurrentNetworkEpochProvider: brcf.currentNetworkEpochProvider,
		SelfShardIdProvider:         brcf.shardCoordinator,
		PeersRatingHandler:          brcf.peersRatingHandler,
	}
	return topicSender.NewTopicRequestSender(arg)
}

func (brcf *baseRequestersContainerFactory) createTrieNodesRequester(
	topic string,
	numCrossShardPeers int,
	numIntraShardPeers int,
	targetShardID uint32,
) (dataRetriever.Requester, error) {
	requestSender, err := brcf.createOneRequestSenderWithSpecifiedNumRequests(
		topic,
		EmptyExcludePeersOnTopic,
		targetShardID,
		numCrossShardPeers,
		numIntraShardPeers,
	)
	if err != nil {
		return nil, err
	}

	arg := requesters.ArgTrieNodeRequester{
		ArgBaseRequester: requesters.ArgBaseRequester{
			RequestSender: requestSender,
			Marshaller:    brcf.marshaller,
		},
	}
	return requesters.NewTrieNodeRequester(arg)
}

func (brcf *baseRequestersContainerFactory) generateValidatorInfoRequester() error {
	identifierValidatorInfo := common.ValidatorInfoTopic
	shardC := brcf.shardCoordinator
	requestSender, err := brcf.createOneRequestSenderWithSpecifiedNumRequests(identifierValidatorInfo, EmptyExcludePeersOnTopic, shardC.SelfId(), brcf.numCrossShardPeers, brcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	arg := requesters.ArgValidatorInfoRequester{
		ArgBaseRequester: requesters.ArgBaseRequester{
			RequestSender: requestSender,
			Marshaller:    brcf.marshaller,
		},
	}
	requester, err := requesters.NewValidatorInfoRequester(arg)
	if err != nil {
		return err
	}

	return brcf.container.Add(identifierValidatorInfo, requester)
}
