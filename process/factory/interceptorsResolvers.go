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
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/resolvers"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type topicName string

const (
	// TransactionTopic is the topic used for sharing transactions
	TransactionTopic topicName = "transactions"
	// HeadersTopic is the topic used for sharing block headers
	HeadersTopic topicName = "headers"
	// TxBlockBodyTopic is the topic used for sharing transactions block bodies
	TxBlockBodyTopic topicName = "txBlockBodies"
	// PeerChBodyTopic is used for sharing peer change block bodies
	PeerChBodyTopic topicName = "peerChangeBlockBodies"
	// StateBodyTopic is used for sharing state block bodies
	StateBodyTopic topicName = "stateBlockBodies"
)

type interceptorsResolvers struct {
	interceptorContainer process.Container
	resolverContainer    process.ResolversContainer

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
	InterceptorContainer process.Container
	ResolverContainer    process.ResolversContainer

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
		interceptorContainer:     config.InterceptorContainer,
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

// CreateInterceptors creates the interceptors and initializes the interceptor container
func (ir *interceptorsResolvers) CreateInterceptors() error {
	err := ir.createTxInterceptor()
	if err != nil {
		return err
	}

	err = ir.createHdrInterceptor()
	if err != nil {
		return err
	}

	err = ir.createTxBlockBodyInterceptor()
	if err != nil {
		return err
	}

	err = ir.createPeerChBlockBodyInterceptor()
	if err != nil {
		return err
	}

	err = ir.createStateBlockBodyInterceptor()
	if err != nil {
		return err
	}

	return nil
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

	err = ir.createStateBlockBodyResolver()
	if err != nil {
		return err
	}

	return nil
}

// InterceptorContainer is a getter for interceptorContainer property
func (ir *interceptorsResolvers) InterceptorContainer() process.Container {
	return ir.interceptorContainer
}

// ResolverContainer is a getter for resolverContainer property
func (ir *interceptorsResolvers) ResolverContainer() process.ResolversContainer {
	return ir.resolverContainer
}

func (ir *interceptorsResolvers) createTxInterceptor() error {
	txStorer := ir.blockchain.GetStorer(blockchain.TransactionUnit)

	txInterceptor, err := transaction.NewTxInterceptor(
		ir.marshalizer,
		ir.dataPool.Transactions(),
		txStorer,
		ir.addrConverter,
		ir.hasher,
		ir.singleSigner,
		ir.keyGen,
		ir.shardCoordinator)

	if err != nil {
		return err
	}

	err = ir.createTopicAndAssignHandler(string(TransactionTopic), txInterceptor, true)
	if err != nil {
		return err
	}

	err = ir.interceptorContainer.Add(string(TransactionTopic), txInterceptor)
	return err
}

func (ir *interceptorsResolvers) createHdrInterceptor() error {
	headerStorer := ir.blockchain.GetStorer(blockchain.BlockHeaderUnit)

	hdrInterceptor, err := interceptors.NewHeaderInterceptor(
		ir.marshalizer,
		ir.dataPool.Headers(),
		ir.dataPool.HeadersNonces(),
		headerStorer,
		ir.multiSigner,
		ir.hasher,
		ir.shardCoordinator,
	)

	if err != nil {
		return err
	}

	err = ir.createTopicAndAssignHandler(string(HeadersTopic), hdrInterceptor, true)
	if err != nil {
		return err
	}

	err = ir.interceptorContainer.Add(string(HeadersTopic), hdrInterceptor)
	return err
}

func (ir *interceptorsResolvers) createTxBlockBodyInterceptor() error {
	txBlockBodyStorer := ir.blockchain.GetStorer(blockchain.TxBlockBodyUnit)

	txBlockBodyInterceptor, err := interceptors.NewTxBlockBodyInterceptor(
		ir.marshalizer,
		ir.dataPool.TxBlocks(),
		txBlockBodyStorer,
		ir.hasher,
		ir.shardCoordinator,
	)

	if err != nil {
		return err
	}

	err = ir.createTopicAndAssignHandler(string(TxBlockBodyTopic), txBlockBodyInterceptor, true)
	if err != nil {
		return err
	}

	err = ir.interceptorContainer.Add(string(TxBlockBodyTopic), txBlockBodyInterceptor)
	return err
}

func (ir *interceptorsResolvers) createPeerChBlockBodyInterceptor() error {
	peerBlockBodyStorer := ir.blockchain.GetStorer(blockchain.PeerBlockBodyUnit)

	peerChBodyInterceptor, err := interceptors.NewPeerBlockBodyInterceptor(
		ir.marshalizer,
		ir.dataPool.PeerChangesBlocks(),
		peerBlockBodyStorer,
		ir.hasher,
		ir.shardCoordinator,
	)

	if err != nil {
		return err
	}

	err = ir.createTopicAndAssignHandler(string(PeerChBodyTopic), peerChBodyInterceptor, true)
	if err != nil {
		return err
	}

	err = ir.interceptorContainer.Add(string(PeerChBodyTopic), peerChBodyInterceptor)
	return err
}

func (ir *interceptorsResolvers) createStateBlockBodyInterceptor() error {
	stateBlockBodyStorer := ir.blockchain.GetStorer(blockchain.StateBlockBodyUnit)

	stateBodyInterceptor, err := interceptors.NewStateBlockBodyInterceptor(
		ir.marshalizer,
		ir.dataPool.StateBlocks(),
		stateBlockBodyStorer,
		ir.hasher,
		ir.shardCoordinator,
	)

	if err != nil {
		return err
	}

	err = ir.createTopicAndAssignHandler(string(StateBodyTopic), stateBodyInterceptor, true)
	if err != nil {
		return err
	}

	err = ir.interceptorContainer.Add(string(StateBodyTopic), stateBodyInterceptor)
	return err
}

func (ir *interceptorsResolvers) createTxResolver() error {
	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		ir.messenger,
		string(TransactionTopic),
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
		string(TransactionTopic)+resolverSender.RequestTopicSuffix(),
		txResolver,
		false)
	if err != nil {
		return err
	}

	err = ir.resolverContainer.Add(string(TransactionTopic), txResolver)
	return err
}

func (ir *interceptorsResolvers) createHdrResolver() error {
	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		ir.messenger,
		string(HeadersTopic),
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
		string(HeadersTopic)+resolverSender.RequestTopicSuffix(),
		hdrResolver,
		false)
	if err != nil {
		return err
	}

	err = ir.resolverContainer.Add(string(HeadersTopic), hdrResolver)
	return err
}

func (ir *interceptorsResolvers) createTxBlockBodyResolver() error {
	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		ir.messenger,
		string(TxBlockBodyTopic),
		ir.marshalizer)
	if err != nil {
		return err
	}

	txBlkResolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		ir.dataPool.TxBlocks(),
		ir.blockchain.GetStorer(blockchain.TxBlockBodyUnit),
		ir.marshalizer)

	if err != nil {
		return err
	}

	//add on the request topic
	err = ir.createTopicAndAssignHandler(
		string(TxBlockBodyTopic)+resolverSender.RequestTopicSuffix(),
		txBlkResolver,
		false)
	if err != nil {
		return err
	}

	err = ir.resolverContainer.Add(string(TxBlockBodyTopic), txBlkResolver)
	return err
}

func (ir *interceptorsResolvers) createPeerChBlockBodyResolver() error {
	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		ir.messenger,
		string(PeerChBodyTopic),
		ir.marshalizer)
	if err != nil {
		return err
	}

	peerChBlkResolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		ir.dataPool.PeerChangesBlocks(),
		ir.blockchain.GetStorer(blockchain.PeerBlockBodyUnit),
		ir.marshalizer)

	if err != nil {
		return err
	}

	//add on the request topic
	err = ir.createTopicAndAssignHandler(
		string(PeerChBodyTopic)+resolverSender.RequestTopicSuffix(),
		peerChBlkResolver,
		false)
	if err != nil {
		return err
	}

	err = ir.resolverContainer.Add(string(PeerChBodyTopic), peerChBlkResolver)
	return err
}

func (ir *interceptorsResolvers) createStateBlockBodyResolver() error {
	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		ir.messenger,
		string(StateBodyTopic),
		ir.marshalizer)
	if err != nil {
		return err
	}

	stateBlkResolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		ir.dataPool.StateBlocks(),
		ir.blockchain.GetStorer(blockchain.StateBlockBodyUnit),
		ir.marshalizer)

	if err != nil {
		return err
	}

	//add on the request topic
	err = ir.createTopicAndAssignHandler(
		string(StateBodyTopic)+resolverSender.RequestTopicSuffix(),
		stateBlkResolver,
		false)
	if err != nil {
		return err
	}

	err = ir.resolverContainer.Add(string(StateBodyTopic), stateBlkResolver)
	return err
}

func validateRequiredProcessCreatorParams(config InterceptorsResolversConfig) error {
	if config.InterceptorContainer == nil {
		return process.ErrNilInterceptorContainer
	}
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
