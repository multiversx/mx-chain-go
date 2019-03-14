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
	dataPools                data.PoolsHolder
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

// TODO ,ddm add constructor name

// NewResolversContainerFactory ddd
func NewResolversContainerFactory(
	shardCoordinator sharding.ShardCoordinator,
	messenger process.TopicMessageHandler,
	blockchain *blockchain.BlockChain,
	marshalizer marshal.Marshalizer,
	dataPools data.PoolsHolder,
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
	if dataPools == nil {
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
		dataPools:                dataPools,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
	}, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (rcf *resolversContainerFactory) Create() (process.ResolversContainer, error) {
	container := containers.NewResolversContainer()

	keys, resolverSlice, err := rcf.generateTxResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = rcf.generateHdrResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = rcf.generateMiniBlocksResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = rcf.generatePeerChBlockBodyResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (rcf *resolversContainerFactory) createTopicAndAssignHandler(
	topicName string,
	resolver process.Resolver,
	createChannel bool,
) (process.Resolver, error) {

	err := rcf.messenger.CreateTopic(topicName, createChannel)
	if err != nil {
		return nil, err
	}

	return resolver, rcf.messenger.RegisterMessageProcessor(topicName, resolver)
}

//------- Tx resolvers

func (rcf *resolversContainerFactory) generateTxResolvers() ([]string, []process.Resolver, error) {
	shardC := rcf.shardCoordinator

	noOfShards := shardC.NoShards()

	keys := make([]string, noOfShards)
	resolverSlice := make([]process.Resolver, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := TransactionTopic + shardC.CommunicationIdentifier(idx)

		resolver, err := rcf.createOneTxResolver(identifierTx)
		if err != nil {
			return nil, nil, err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierTx
	}

	return keys, resolverSlice, nil
}

func (rcf *resolversContainerFactory) createOneTxResolver(identifier string) (process.Resolver, error) {
	txStorer := rcf.blockchain.GetStorer(blockchain.TransactionUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	resolver, err := transaction.NewTxResolver(
		resolverSender,
		rcf.dataPools.Transactions(),
		txStorer,
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		identifier+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}

//------- Hdr resolvers

func (rcf *resolversContainerFactory) generateHdrResolvers() ([]string, []process.Resolver, error) {
	shardC := rcf.shardCoordinator

	//only one intrashard header topic
	identifierHdr := HeadersTopic + shardC.CommunicationIdentifier(shardC.ShardForCurrentNode())

	resolver, err := rcf.createOneHdrResolver(identifierHdr)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHdr}, []process.Resolver{resolver}, nil
}

func (rcf *resolversContainerFactory) createOneHdrResolver(identifier string) (process.Resolver, error) {
	hdrStorer := rcf.blockchain.GetStorer(blockchain.BlockHeaderUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	resolver, err := resolvers.NewHeaderResolver(
		resolverSender,
		rcf.dataPools,
		hdrStorer,
		rcf.marshalizer,
		rcf.uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		identifier+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}

//------- MiniBlocks resolvers

func (rcf *resolversContainerFactory) generateMiniBlocksResolvers() ([]string, []process.Resolver, error) {
	shardC := rcf.shardCoordinator

	noOfShards := shardC.NoShards()

	keys := make([]string, noOfShards)
	resolverSlice := make([]process.Resolver, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := MiniBlocksTopic + shardC.CommunicationIdentifier(idx)

		resolver, err := rcf.createOneMiniBlocksResolver(identifierMiniBlocks)
		if err != nil {
			return nil, nil, err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierMiniBlocks
	}

	return keys, resolverSlice, nil
}

func (rcf *resolversContainerFactory) createOneMiniBlocksResolver(identifier string) (process.Resolver, error) {
	miniBlocksStorer := rcf.blockchain.GetStorer(blockchain.MiniBlockUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	txBlkResolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		rcf.dataPools.MiniBlocks(),
		miniBlocksStorer,
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		identifier+resolverSender.TopicRequestSuffix(),
		txBlkResolver,
		false)
}

//------- PeerChBlocks resolvers

func (rcf *resolversContainerFactory) generatePeerChBlockBodyResolvers() ([]string, []process.Resolver, error) {
	shardC := rcf.shardCoordinator

	//only one intrashard peer change blocks topic
	identifierPeerCh := PeerChBodyTopic + shardC.CommunicationIdentifier(shardC.ShardForCurrentNode())

	resolver, err := rcf.createOnePeerChBlockBodyResolver(identifierPeerCh)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierPeerCh}, []process.Resolver{resolver}, nil
}

func (rcf *resolversContainerFactory) createOnePeerChBlockBodyResolver(identifier string) (process.Resolver, error) {
	peerBlockBodyStorer := rcf.blockchain.GetStorer(blockchain.PeerChangesUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	peerChResolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		rcf.dataPools.MiniBlocks(),
		peerBlockBodyStorer,
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		identifier+resolverSender.TopicRequestSuffix(),
		peerChResolver,
		false)
}
