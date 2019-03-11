package factory

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/resolvers"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

const (
	// TransactionTopic is the topic used for sharing transactions
	TransactionTopic = "transactions"
	// HeadersTopic is the topic used for sharing block headers
	HeadersTopic = "headers"
	// MiniBlocksTopic is the topic used for sharing mini blocks
	MiniBlocksTopic = "txBlockBodies"
	// PeerChBodyTopic is used for sharing peer change block bodies
	PeerChBodyTopic = "peerChangeBlockBodies"
)

type interceptorsResolvers struct {
	resolverContainer process.ResolversContainer

	messenger                p2p.Messenger
	blockchain               *blockchain.BlockChain
	dataPool                 data.TransientDataHolder
	shardCoordinator         sharding.ShardCoordinator
	addrConverter            state.AddressConverter
	hasher                   hashing.Hasher
	marshalizer              marshal.Marshalizer
	multiSigner              crypto.MultiSigner
	singleSigner             crypto.SingleSigner
	keyGen                   crypto.KeyGenerator
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

// InterceptorsResolversConfig is the struct containing the needed params to be
//  provided when initialising a new interceptorsResolvers factory
type InterceptorsResolversConfig struct {
	ResolverContainer process.ResolversContainer

	Messenger                p2p.Messenger
	Blockchain               *blockchain.BlockChain
	DataPool                 data.TransientDataHolder
	ShardCoordinator         sharding.ShardCoordinator
	AddrConverter            state.AddressConverter
	Hasher                   hashing.Hasher
	Marshalizer              marshal.Marshalizer
	MultiSigner              crypto.MultiSigner
	SingleSigner             crypto.SingleSigner
	KeyGen                   crypto.KeyGenerator
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

// NewInterceptorsResolversCreator is responsible for creating a new interceptorsResolvers factory object
func NewInterceptorsResolversCreator(config InterceptorsResolversConfig) (*interceptorsResolvers, error) {
	err := validateRequiredProcessCreatorParams(config)
	if err != nil {
		return nil, err
	}
	return &interceptorsResolvers{
		resolverContainer:        config.ResolverContainer,
		messenger:                config.Messenger,
		blockchain:               config.Blockchain,
		dataPool:                 config.DataPool,
		shardCoordinator:         config.ShardCoordinator,
		addrConverter:            config.AddrConverter,
		hasher:                   config.Hasher,
		marshalizer:              config.Marshalizer,
		multiSigner:              config.MultiSigner,
		singleSigner:             config.SingleSigner,
		keyGen:                   config.KeyGen,
		uint64ByteSliceConverter: config.Uint64ByteSliceConverter,
	}, nil
}

// CreateResolvers creates the resolvers and initializes the resolver container
func (ir *interceptorsResolvers) CreateResolvers() error {
	err := ir.createTxResolver()
	if err != nil {
		return err
	}

	err = ir.createHdrResolver()
	if err != nil {
		return err
	}

	err = ir.createTxBlockBodyResolver()
	if err != nil {
		return err
	}

	err = ir.createPeerChBlockBodyResolver()
	if err != nil {
		return err
	}

	return nil
}

// ResolverContainer is a getter for resolverContainer property
func (ir *interceptorsResolvers) ResolverContainer() process.ResolversContainer {
	return ir.resolverContainer
}

func (ir *interceptorsResolvers) createTxResolver() error {
	//TODO temporary, will be refactored in EN-1104
	identifier := TransactionTopic + ir.shardCoordinator.CommunicationIdentifier(ir.shardCoordinator.ShardForCurrentNode())

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		ir.messenger,
		identifier,
		ir.marshalizer)
	if err != nil {
		return err
	}

	txResolver, err := transaction.NewTxResolver(
		resolverSender,
		ir.dataPool.Transactions(),
		ir.blockchain.GetStorer(blockchain.TransactionUnit),
		ir.marshalizer)

	if err != nil {
		return err
	}

	//add on the request topic
	err = ir.createTopicAndAssignHandler(
		identifier+resolverSender.RequestTopicSuffix(),
		txResolver,
		false)
	if err != nil {
		return err
	}

	err = ir.resolverContainer.Add(TransactionTopic, txResolver)
	return err
}

func (ir *interceptorsResolvers) createHdrResolver() error {
	//TODO temporary, will be refactored in EN-1104
	identifier := HeadersTopic + ir.shardCoordinator.CommunicationIdentifier(ir.shardCoordinator.ShardForCurrentNode())

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		ir.messenger,
		identifier,
		ir.marshalizer)
	if err != nil {
		return err
	}

	hdrResolver, err := resolvers.NewHeaderResolver(
		resolverSender,
		ir.dataPool,
		ir.blockchain.GetStorer(blockchain.BlockHeaderUnit),
		ir.marshalizer,
		ir.uint64ByteSliceConverter)

	if err != nil {
		return err
	}

	//add on the request topic
	err = ir.createTopicAndAssignHandler(
		identifier+resolverSender.RequestTopicSuffix(),
		hdrResolver,
		false)
	if err != nil {
		return err
	}

	err = ir.resolverContainer.Add(HeadersTopic, hdrResolver)
	return err
}

func (ir *interceptorsResolvers) createTxBlockBodyResolver() error {
	//TODO temporary, will be refactored in EN-1104
	identifier := MiniBlocksTopic + ir.shardCoordinator.CommunicationIdentifier(ir.shardCoordinator.ShardForCurrentNode())

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		ir.messenger,
		identifier,
		ir.marshalizer)
	if err != nil {
		return err
	}

	txBlkResolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		ir.dataPool.MiniBlocks(),
		ir.blockchain.GetStorer(blockchain.MiniBlockUnit),
		ir.marshalizer)

	if err != nil {
		return err
	}

	//add on the request topic
	err = ir.createTopicAndAssignHandler(
		identifier+resolverSender.RequestTopicSuffix(),
		txBlkResolver,
		false)
	if err != nil {
		return err
	}

	err = ir.resolverContainer.Add(MiniBlocksTopic, txBlkResolver)
	return err
}

func (ir *interceptorsResolvers) createPeerChBlockBodyResolver() error {
	//TODO temporary, will be refactored in EN-1104
	identifier := PeerChBodyTopic + ir.shardCoordinator.CommunicationIdentifier(ir.shardCoordinator.ShardForCurrentNode())

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		ir.messenger,
		identifier,
		ir.marshalizer)
	if err != nil {
		return err
	}

	peerChBlkResolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		ir.dataPool.PeerChangesBlocks(),
		ir.blockchain.GetStorer(blockchain.PeerChangesUnit),
		ir.marshalizer)

	if err != nil {
		return err
	}

	//add on the request topic
	err = ir.createTopicAndAssignHandler(
		identifier+resolverSender.RequestTopicSuffix(),
		peerChBlkResolver,
		false)
	if err != nil {
		return err
	}

	err = ir.resolverContainer.Add(PeerChBodyTopic, peerChBlkResolver)
	return err
}

func validateRequiredProcessCreatorParams(config InterceptorsResolversConfig) error {
	if config.ResolverContainer == nil {
		return process.ErrNilResolverContainer
	}
	if config.Messenger == nil {
		return process.ErrNilMessenger
	}
	if config.Blockchain == nil {
		return process.ErrNilBlockChain
	}
	if config.DataPool == nil {
		return process.ErrNilDataPoolHolder
	}
	if config.ShardCoordinator == nil {
		return process.ErrNilShardCoordinator
	}
	if config.AddrConverter == nil {
		return process.ErrNilAddressConverter
	}
	if config.Hasher == nil {
		return process.ErrNilHasher
	}
	if config.Marshalizer == nil {
		return process.ErrNilMarshalizer
	}
	if config.SingleSigner == nil {
		return process.ErrNilSingleSigner
	}
	if config.MultiSigner == nil {
		return process.ErrNilMultiSigVerifier
	}
	if config.KeyGen == nil {
		return process.ErrNilKeyGen
	}
	if config.Uint64ByteSliceConverter == nil {
		return process.ErrNilUint64ByteSliceConverter
	}

	return nil
}

func (ir *interceptorsResolvers) createTopicAndAssignHandler(topic string, handler p2p.MessageProcessor, createPipe bool) error {
	err := ir.messenger.CreateTopic(topic, createPipe)
	if err != nil {
		return err
	}

	return ir.messenger.RegisterMessageProcessor(topic, handler)
}
