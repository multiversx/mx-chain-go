package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/throttler"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/topicSender"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/genesis"
)

const defaultTargetShardID = uint32(0)

type resolversContainerFactory struct {
	shardCoordinator       sharding.Coordinator
	messenger              dataRetriever.TopicMessageHandler
	marshalizer            marshal.Marshalizer
	dataTrieContainer      common.TriesHolder
	container              dataRetriever.ResolversContainer
	inputAntifloodHandler  dataRetriever.P2PAntifloodHandler
	outputAntifloodHandler dataRetriever.P2PAntifloodHandler
	throttler              dataRetriever.ResolverThrottler
}

// ArgsNewResolversContainerFactory defines the arguments for the resolversContainerFactory constructor
type ArgsNewResolversContainerFactory struct {
	ShardCoordinator           sharding.Coordinator
	Messenger                  dataRetriever.TopicMessageHandler
	Marshalizer                marshal.Marshalizer
	DataTrieContainer          common.TriesHolder
	ExistingResolvers          dataRetriever.ResolversContainer
	InputAntifloodHandler      dataRetriever.P2PAntifloodHandler
	OutputAntifloodHandler     dataRetriever.P2PAntifloodHandler
	NumConcurrentResolvingJobs int32
}

// NewResolversContainerFactory creates a new container filled with topic resolvers
func NewResolversContainerFactory(args ArgsNewResolversContainerFactory) (*resolversContainerFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.Messenger) {
		return nil, update.ErrNilMessenger
	}
	if check.IfNil(args.Marshalizer) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.DataTrieContainer) {
		return nil, update.ErrNilTrieDataGetter
	}
	if check.IfNil(args.ExistingResolvers) {
		return nil, fmt.Errorf("%w for resolvers", update.ErrNilContainer)
	}

	thr, err := throttler.NewNumGoRoutinesThrottler(args.NumConcurrentResolvingJobs)
	if err != nil {
		return nil, err
	}
	return &resolversContainerFactory{
		shardCoordinator:       args.ShardCoordinator,
		messenger:              args.Messenger,
		marshalizer:            args.Marshalizer,
		dataTrieContainer:      args.DataTrieContainer,
		container:              args.ExistingResolvers,
		inputAntifloodHandler:  args.InputAntifloodHandler,
		outputAntifloodHandler: args.OutputAntifloodHandler,
		throttler:              thr,
	}, nil
}

// Create returns a resolver container that will hold all resolvers in the system
func (rcf *resolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	err := rcf.generateTrieNodesResolvers()
	if err != nil {
		return nil, err
	}

	return rcf.container, nil
}

func (rcf *resolversContainerFactory) generateTrieNodesResolvers() error {
	shardC := rcf.shardCoordinator

	keys := make([]string, 0)
	resolversSlice := make([]dataRetriever.Resolver, 0)

	for i := uint32(0); i < shardC.NumberOfShards(); i++ {
		identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(i, core.MetachainShardId)
		if rcf.checkIfResolverExists(identifierTrieNodes) {
			continue
		}

		trieId := genesis.CreateTrieIdentifier(i, genesis.UserAccount)
		resolver, err := rcf.createTrieNodesResolver(identifierTrieNodes, trieId)
		if err != nil {
			return err
		}

		resolversSlice = append(resolversSlice, resolver)
		keys = append(keys, identifierTrieNodes)
	}

	identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	if !rcf.checkIfResolverExists(identifierTrieNodes) {
		trieId := genesis.CreateTrieIdentifier(core.MetachainShardId, genesis.UserAccount)
		resolver, err := rcf.createTrieNodesResolver(identifierTrieNodes, trieId)
		if err != nil {
			return err
		}

		resolversSlice = append(resolversSlice, resolver)
		keys = append(keys, identifierTrieNodes)
	}

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	if !rcf.checkIfResolverExists(identifierTrieNodes) {
		trieID := genesis.CreateTrieIdentifier(core.MetachainShardId, genesis.ValidatorAccount)
		resolver, err := rcf.createTrieNodesResolver(identifierTrieNodes, trieID)
		if err != nil {
			return err
		}

		resolversSlice = append(resolversSlice, resolver)
		keys = append(keys, identifierTrieNodes)
	}

	return rcf.container.AddMultiple(keys, resolversSlice)
}

func (rcf *resolversContainerFactory) checkIfResolverExists(topic string) bool {
	_, err := rcf.container.Get(topic)
	return err == nil
}

func (rcf *resolversContainerFactory) createTrieNodesResolver(baseTopic string, trieId string) (dataRetriever.Resolver, error) {
	arg := topicSender.ArgTopicResolverSender{
		ArgBaseTopicSender: topicSender.ArgBaseTopicSender{
			Messenger:            rcf.messenger,
			TopicName:            baseTopic,
			OutputAntiflooder:    rcf.outputAntifloodHandler,
			PreferredPeersHolder: disabled.NewPreferredPeersHolder(),
			TargetShardId:        defaultTargetShardID,
		},
	}
	resolverSender, err := topicSender.NewTopicResolverSender(arg)
	if err != nil {
		return nil, err
	}

	trie := rcf.dataTrieContainer.Get([]byte(trieId))
	argTrieResolver := resolvers.ArgTrieNodeResolver{
		ArgBaseResolver: resolvers.ArgBaseResolver{
			SenderResolver:   resolverSender,
			Marshaller:       rcf.marshalizer,
			AntifloodHandler: rcf.inputAntifloodHandler,
			Throttler:        rcf.throttler,
		},
		TrieDataGetter: trie,
	}
	resolver, err := resolvers.NewTrieNodeResolver(argTrieResolver)
	if err != nil {
		return nil, err
	}

	err = rcf.messenger.RegisterMessageProcessor(resolver.RequestTopic(), common.HardforkResolversIdentifier, resolver)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcf *resolversContainerFactory) IsInterfaceNil() bool {
	return rcf == nil
}
