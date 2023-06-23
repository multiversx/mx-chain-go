package requesterscontainer

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/random"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers/requesters"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type metaRequestersContainerFactory struct {
	*baseRequestersContainerFactory
}

// NewMetaRequestersContainerFactory creates a new container filled with topic requesters for metachain
func NewMetaRequestersContainerFactory(
	args FactoryArgs,
) (*metaRequestersContainerFactory, error) {
	if args.SizeCheckDelta > 0 {
		args.Marshaller = marshal.NewSizeCheckUnmarshalizer(args.Marshaller, args.SizeCheckDelta)
	}

	numIntraShardPeers := args.RequesterConfig.NumTotalPeers - args.RequesterConfig.NumCrossShardPeers
	container := containers.NewRequestersContainer()
	base := &baseRequestersContainerFactory{
		container:                       container,
		shardCoordinator:                args.ShardCoordinator,
		mainMessenger:                   args.MainMessenger,
		fullArchiveMessenger:            args.FullArchiveMessenger,
		marshaller:                      args.Marshaller,
		uint64ByteSliceConverter:        args.Uint64ByteSliceConverter,
		intRandomizer:                   &random.ConcurrentSafeIntRandomizer{},
		outputAntifloodHandler:          args.OutputAntifloodHandler,
		currentNetworkEpochProvider:     args.CurrentNetworkEpochProvider,
		mainPreferredPeersHolder:        args.MainPreferredPeersHolder,
		fullArchivePreferredPeersHolder: args.FullArchivePreferredPeersHolder,
		peersRatingHandler:              args.PeersRatingHandler,
		numCrossShardPeers:              int(args.RequesterConfig.NumCrossShardPeers),
		numIntraShardPeers:              int(numIntraShardPeers),
		numTotalPeers:                   int(args.RequesterConfig.NumTotalPeers),
		numFullHistoryPeers:             int(args.RequesterConfig.NumFullHistoryPeers),
	}

	err := base.checkParams()
	if err != nil {
		return nil, err
	}

	base.intraShardTopic = common.ConsensusTopic +
		base.shardCoordinator.CommunicationIdentifier(base.shardCoordinator.SelfId())

	return &metaRequestersContainerFactory{
		baseRequestersContainerFactory: base,
	}, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (mrcf *metaRequestersContainerFactory) Create() (dataRetriever.RequestersContainer, error) {
	err := mrcf.generateCommonRequesters()
	if err != nil {
		return nil, err
	}

	err = mrcf.generateShardHeaderRequesters()
	if err != nil {
		return nil, err
	}

	err = mrcf.generateMetaChainHeaderRequesters()
	if err != nil {
		return nil, err
	}

	err = mrcf.generateRewardsRequesters(factory.RewardsTransactionTopic)
	if err != nil {
		return nil, err
	}

	err = mrcf.generateTrieNodesRequesters()
	if err != nil {
		return nil, err
	}

	return mrcf.container, nil
}

// AddShardTrieNodeRequesters will add trie node requesters to the existing container, needed for start in epoch
func (mrcf *metaRequestersContainerFactory) AddShardTrieNodeRequesters(container dataRetriever.RequestersContainer) error {
	if check.IfNil(container) {
		return dataRetriever.ErrNilRequestersContainer
	}

	shardC := mrcf.shardCoordinator

	keys := make([]string, 0)
	requestersSlice := make([]dataRetriever.Requester, 0)

	for idx := uint32(0); idx < shardC.NumberOfShards(); idx++ {
		identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(idx)
		requester, err := mrcf.createTrieNodesRequester(
			identifierTrieNodes,
			mrcf.numCrossShardPeers,
			mrcf.numTotalPeers-mrcf.numCrossShardPeers,
			idx,
		)
		if err != nil {
			return err
		}

		requestersSlice = append(requestersSlice, requester)
		keys = append(keys, identifierTrieNodes)
	}

	return container.AddMultiple(keys, requestersSlice)
}

// ------- Shard header requesters

func (mrcf *metaRequestersContainerFactory) generateShardHeaderRequesters() error {
	shardC := mrcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	requestersSlice := make([]dataRetriever.Requester, noOfShards)

	// wire up to topics: shardBlocks_0_META, shardBlocks_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierHeader := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(idx)
		excludePeersFromTopic := EmptyExcludePeersOnTopic

		requester, err := mrcf.createShardHeaderRequester(identifierHeader, excludePeersFromTopic, idx, mrcf.numCrossShardPeers, mrcf.numIntraShardPeers)
		if err != nil {
			return err
		}

		requestersSlice[idx] = requester
		keys[idx] = identifierHeader
	}

	return mrcf.container.AddMultiple(keys, requestersSlice)
}

func (mrcf *metaRequestersContainerFactory) createShardHeaderRequester(
	topic string,
	excludedTopic string,
	shardID uint32,
	numCrossShardPeers int,
	numIntraShardPeers int,
) (dataRetriever.Requester, error) {
	requestSender, err := mrcf.createOneRequestSenderWithSpecifiedNumRequests(topic, excludedTopic, shardID, numCrossShardPeers, numIntraShardPeers)
	if err != nil {
		return nil, err
	}

	arg := requesters.ArgHeaderRequester{
		ArgBaseRequester: requesters.ArgBaseRequester{
			RequestSender: requestSender,
			Marshaller:    mrcf.marshaller,
		},
		NonceConverter: mrcf.uint64ByteSliceConverter,
	}
	return requesters.NewHeaderRequester(arg)
}

// ------- Meta header requester

func (mrcf *metaRequestersContainerFactory) generateMetaChainHeaderRequesters() error {
	identifierHeader := factory.MetachainBlocksTopic
	requester, err := mrcf.createMetaChainHeaderRequester(identifierHeader, core.MetachainShardId, mrcf.numCrossShardPeers, mrcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	return mrcf.container.Add(identifierHeader, requester)
}

func (mrcf *metaRequestersContainerFactory) createMetaChainHeaderRequester(
	identifier string,
	shardId uint32,
	numCrossShardPeers int,
	numIntraShardPeers int,
) (dataRetriever.Requester, error) {
	requestSender, err := mrcf.createOneRequestSenderWithSpecifiedNumRequests(identifier, EmptyExcludePeersOnTopic, shardId, numCrossShardPeers, numIntraShardPeers)
	if err != nil {
		return nil, err
	}

	arg := requesters.ArgHeaderRequester{
		ArgBaseRequester: requesters.ArgBaseRequester{
			RequestSender: requestSender,
			Marshaller:    mrcf.marshaller,
		},
		NonceConverter: mrcf.uint64ByteSliceConverter,
	}
	return requesters.NewHeaderRequester(arg)
}

func (mrcf *metaRequestersContainerFactory) generateTrieNodesRequesters() error {
	keys := make([]string, 0)
	requestersSlice := make([]dataRetriever.Requester, 0)

	identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	requester, err := mrcf.createTrieNodesRequester(
		identifierTrieNodes,
		0,
		mrcf.numTotalPeers,
		core.MetachainShardId,
	)
	if err != nil {
		return err
	}

	requestersSlice = append(requestersSlice, requester)
	keys = append(keys, identifierTrieNodes)

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	requester, err = mrcf.createTrieNodesRequester(
		identifierTrieNodes,
		0,
		mrcf.numTotalPeers,
		core.MetachainShardId,
	)
	if err != nil {
		return err
	}

	requestersSlice = append(requestersSlice, requester)
	keys = append(keys, identifierTrieNodes)

	return mrcf.container.AddMultiple(keys, requestersSlice)
}

func (mrcf *metaRequestersContainerFactory) generateRewardsRequesters(topic string) error {

	shardC := mrcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	requestersSlice := make([]dataRetriever.Requester, noOfShards)

	// wire up to topics: shardBlocks_0_META, shardBlocks_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := topic + shardC.CommunicationIdentifier(idx)
		excludePeersFromTopic := EmptyExcludePeersOnTopic

		requester, err := mrcf.createTxRequester(identifierTx, excludePeersFromTopic, idx, mrcf.numCrossShardPeers, mrcf.numIntraShardPeers)
		if err != nil {
			return err
		}

		requestersSlice[idx] = requester
		keys[idx] = identifierTx
	}

	return mrcf.container.AddMultiple(keys, requestersSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mrcf *metaRequestersContainerFactory) IsInterfaceNil() bool {
	return mrcf == nil
}
