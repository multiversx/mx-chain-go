package factory

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/resolvers"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type resolversContainerFactory struct {
	shardCoordinator         sharding.ShardCoordinator
	messenger                process.TopicMessageHandler
	blockchain               *blockchain.BlockChain
	marshalizer              marshal.Marshalizer
	dataPool                 data.TransientDataHolder
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

func NewResolversContainerFactory(
	shardCoordinator sharding.ShardCoordinator,
	messenger process.TopicMessageHandler,
	blockchain *blockchain.BlockChain,
	marshalizer marshal.Marshalizer,
	dataPool data.TransientDataHolder,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
) (*resolversContainerFactory, error) {

	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if messenger == nil {
		return nil, process.ErrNilMessenger
	}
	if blockchain == nil {
		return nil, process.ErrNilBlockChain
	}
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if dataPool == nil {
		return nil, process.ErrNilDataPoolHolder
	}
	if uint64ByteSliceConverter == nil {
		return nil, process.ErrNilUint64ByteSliceConverter
	}

	return &resolversContainerFactory{
		shardCoordinator:         shardCoordinator,
		messenger:                messenger,
		blockchain:               blockchain,
		marshalizer:              marshalizer,
		dataPool:                 dataPool,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
	}, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (rcf *resolversContainerFactory) Create() (process.ResolversContainer, error) {
	container := containers.NewResolversContainer()

	err := rcf.addTxResolvers(container)
	if err != nil {
		return nil, err
	}

	err = rcf.addHdrResolvers(container)
	if err != nil {
		return nil, err
	}

	err = rcf.addMiniBlocksResolvers(container)
	if err != nil {
		return nil, err
	}

	err = rcf.addPeerChBlockBodyResolvers(container)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (rcf *resolversContainerFactory) createTopicAndAssignHandler(
	topic string,
	resolver process.Resolver,
	createChannel bool) (process.Resolver, error) {

	err := rcf.messenger.CreateTopic(topic, createChannel)
	if err != nil {
		return nil, err
	}

	return resolver, rcf.messenger.RegisterMessageProcessor(topic, resolver)
}

//------- Tx resolvers

func (rcf *resolversContainerFactory) addTxResolvers(container process.ResolversContainer) error {
	shardC := rcf.shardCoordinator

	noOfShards := shardC.NoShards()

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := TransactionTopic + shardC.CommunicationIdentifier(idx)

		resolver, err := rcf.createOneTxResolver(identifierTx)
		if err != nil {
			return err
		}

		err = container.Add(identifierTx, resolver)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rcf *resolversContainerFactory) createOneTxResolver(identifier string) (process.Resolver, error) {
	txStorer := rcf.blockchain.GetStorer(blockchain.TransactionUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		rcf.marshalizer)
	if err != nil {
		return nil, err
	}

	resolver, err := transaction.NewTxResolver(
		resolverSender,
		rcf.dataPool.Transactions(),
		txStorer,
		rcf.marshalizer)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		identifier+resolverSender.RequestTopicSuffix(),
		resolver,
		false)
}

//------- Hdr resolvers

func (rcf *resolversContainerFactory) addHdrResolvers(container process.ResolversContainer) error {
	shardC := rcf.shardCoordinator

	//only one intrashard header topic
	identifierHdr := HeadersTopic + shardC.CommunicationIdentifier(shardC.ShardForCurrentNode())

	resolver, err := rcf.createOneHdrResolver(identifierHdr)
	if err != nil {
		return err
	}

	return container.Add(identifierHdr, resolver)
}

func (rcf *resolversContainerFactory) createOneHdrResolver(identifier string) (process.Resolver, error) {
	hdrStorer := rcf.blockchain.GetStorer(blockchain.BlockHeaderUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		rcf.marshalizer)
	if err != nil {
		return nil, err
	}

	resolver, err := resolvers.NewHeaderResolver(
		resolverSender,
		rcf.dataPool,
		hdrStorer,
		rcf.marshalizer,
		rcf.uint64ByteSliceConverter)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		identifier+resolverSender.RequestTopicSuffix(),
		resolver,
		false)
}

//------- MiniBlocks resolvers

func (rcf *resolversContainerFactory) addMiniBlocksResolvers(container process.ResolversContainer) error {
	shardC := rcf.shardCoordinator

	noOfShards := shardC.NoShards()

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := MiniBlocksTopic + shardC.CommunicationIdentifier(idx)

		resolver, err := rcf.createOneMiniBlocksResolver(identifierMiniBlocks)
		if err != nil {
			return err
		}

		err = container.Add(identifierMiniBlocks, resolver)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rcf *resolversContainerFactory) createOneMiniBlocksResolver(identifier string) (process.Resolver, error) {
	miniBlocksStorer := rcf.blockchain.GetStorer(blockchain.MiniBlockUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		rcf.marshalizer)
	if err != nil {
		return nil, err
	}

	txBlkResolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		rcf.dataPool.MiniBlocks(),
		miniBlocksStorer,
		rcf.marshalizer)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		identifier+resolverSender.RequestTopicSuffix(),
		txBlkResolver,
		false)
}

//------- PeerChBlocks resolvers

func (rcf *resolversContainerFactory) addPeerChBlockBodyResolvers(container process.ResolversContainer) error {
	shardC := rcf.shardCoordinator

	//only one intrashard peer change blocks topic
	identifierPeerCh := PeerChBodyTopic + shardC.CommunicationIdentifier(shardC.ShardForCurrentNode())

	resolver, err := rcf.createOnePeerChBlockBodyResolver(identifierPeerCh)
	if err != nil {
		return err
	}

	return container.Add(identifierPeerCh, resolver)
}

func (rcf *resolversContainerFactory) createOnePeerChBlockBodyResolver(identifier string) (process.Resolver, error) {
	peerBlockBodyStorer := rcf.blockchain.GetStorer(blockchain.PeerChangesUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		rcf.marshalizer)
	if err != nil {
		return nil, err
	}

	peerChResolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		rcf.dataPool.MiniBlocks(),
		peerBlockBodyStorer,
		rcf.marshalizer)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		identifier+resolverSender.RequestTopicSuffix(),
		peerChResolver,
		false)
}
