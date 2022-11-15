package requesterscontainer

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/random"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers/requesters"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

type shardRequestersContainerFactory struct {
	*baseRequestersContainerFactory
}

// NewShardRequestersContainerFactory creates a new container filled with topic requesters for shards
func NewShardRequestersContainerFactory(
	args FactoryArgs,
) (*shardRequestersContainerFactory, error) {
	if args.SizeCheckDelta > 0 {
		args.Marshaller = marshal.NewSizeCheckUnmarshalizer(args.Marshaller, args.SizeCheckDelta)
	}

	numIntraShardPeers := args.RequesterConfig.NumTotalPeers - args.RequesterConfig.NumCrossShardPeers
	base := &baseRequestersContainerFactory{
		container:                   nil, // TODO[Sorin]: add real component
		shardCoordinator:            args.ShardCoordinator,
		messenger:                   args.Messenger,
		marshaller:                  args.Marshaller,
		uint64ByteSliceConverter:    args.Uint64ByteSliceConverter,
		intRandomizer:               &random.ConcurrentSafeIntRandomizer{},
		outputAntifloodHandler:      args.OutputAntifloodHandler,
		currentNetworkEpochProvider: args.CurrentNetworkEpochProvider,
		preferredPeersHolder:        args.PreferredPeersHolder,
		peersRatingHandler:          args.PeersRatingHandler,
		numCrossShardPeers:          int(args.RequesterConfig.NumCrossShardPeers),
		numIntraShardPeers:          int(numIntraShardPeers),
		numTotalPeers:               int(args.RequesterConfig.NumTotalPeers),
		numFullHistoryPeers:         int(args.RequesterConfig.NumFullHistoryPeers),
	}

	err := base.checkParams()
	if err != nil {
		return nil, err
	}

	base.intraShardTopic = common.ConsensusTopic +
		base.shardCoordinator.CommunicationIdentifier(base.shardCoordinator.SelfId())

	return &shardRequestersContainerFactory{
		baseRequestersContainerFactory: base,
	}, nil
}

// Create returns a requesters container that will hold all requesters in the system
func (srcf *shardRequestersContainerFactory) Create() (dataRetriever.RequestersContainer, error) {
	err := srcf.generateTxRequesters(factory.TransactionTopic)
	if err != nil {
		return nil, err
	}

	err = srcf.generateTxRequesters(factory.UnsignedTransactionTopic)
	if err != nil {
		return nil, err
	}

	err = srcf.generateRewardRequester(factory.RewardsTransactionTopic)
	if err != nil {
		return nil, err
	}

	err = srcf.generateHeaderRequesters()
	if err != nil {
		return nil, err
	}

	err = srcf.generateMiniBlocksRequesters()
	if err != nil {
		return nil, err
	}

	err = srcf.generateMetablockHeaderRequesters()
	if err != nil {
		return nil, err
	}

	err = srcf.generateTrieNodesRequesters()
	if err != nil {
		return nil, err
	}

	err = srcf.generatePeerAuthenticationRequester()
	if err != nil {
		return nil, err
	}

	err = srcf.generateValidatorInfoRequester()
	if err != nil {
		return nil, err
	}

	return srcf.container, nil
}

// ------- Hdr requester

func (srcf *shardRequestersContainerFactory) generateHeaderRequesters() error {
	shardC := srcf.shardCoordinator

	// only one shard header topic, for example: shardBlocks_0_META
	identifierHdr := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	requestSender, err := srcf.createOneRequestSenderWithSpecifiedNumRequests(identifierHdr, EmptyExcludePeersOnTopic, core.MetachainShardId, srcf.numCrossShardPeers, srcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	arg := requesters.ArgHeaderRequester{
		ArgBaseRequester: requesters.ArgBaseRequester{
			RequestSender: requestSender,
			Marshaller:    srcf.marshaller,
		},
		NonceConverter: srcf.uint64ByteSliceConverter,
	}
	requester, err := requesters.NewHeaderRequester(arg)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierHdr, requester)
}

// ------- MetaBlockHeaderRequesters

func (srcf *shardRequestersContainerFactory) generateMetablockHeaderRequesters() error {
	// only one metachain header block topic
	// this is: metachainBlocks
	identifierHdr := factory.MetachainBlocksTopic
	requestSender, err := srcf.createOneRequestSenderWithSpecifiedNumRequests(identifierHdr, EmptyExcludePeersOnTopic, core.MetachainShardId, srcf.numCrossShardPeers, srcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	arg := requesters.ArgHeaderRequester{
		ArgBaseRequester: requesters.ArgBaseRequester{
			RequestSender: requestSender,
			Marshaller:    srcf.marshaller,
		},
		NonceConverter: srcf.uint64ByteSliceConverter,
	}
	requester, err := requesters.NewHeaderRequester(arg)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierHdr, requester)
}

func (srcf *shardRequestersContainerFactory) generateTrieNodesRequesters() error {
	shardC := srcf.shardCoordinator

	keys := make([]string, 0)
	requestersSlice := make([]dataRetriever.Requester, 0)

	identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	requester, err := srcf.createTrieNodesRequester(
		identifierTrieNodes,
		0,
		srcf.numTotalPeers,
		core.MetachainShardId,
	)
	if err != nil {
		return err
	}

	requestersSlice = append(requestersSlice, requester)
	keys = append(keys, identifierTrieNodes)

	return srcf.container.AddMultiple(keys, requestersSlice)
}

func (srcf *shardRequestersContainerFactory) generateRewardRequester(topic string) error {
	shardC := srcf.shardCoordinator

	keys := make([]string, 0)
	requestersSlice := make([]dataRetriever.Requester, 0)

	identifierTx := topic + shardC.CommunicationIdentifier(core.MetachainShardId)
	excludedPeersOnTopic := factory.TransactionTopic + shardC.CommunicationIdentifier(shardC.SelfId())

	requester, err := srcf.createTxRequester(identifierTx, excludedPeersOnTopic, core.MetachainShardId, srcf.numCrossShardPeers, srcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	requestersSlice = append(requestersSlice, requester)
	keys = append(keys, identifierTx)

	return srcf.container.AddMultiple(keys, requestersSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (srcf *shardRequestersContainerFactory) IsInterfaceNil() bool {
	return srcf == nil
}
