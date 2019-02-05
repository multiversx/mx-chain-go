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
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/interceptor"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/resolver"
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

type processorsCreator struct {
	interceptorContainer process.InterceptorContainer
	resolverContainer    process.ResolverContainer

	messenger                p2p.Messenger
	blockchain               *blockchain.BlockChain
	dataPool                 data.TransientDataHolder
	shardCoordinator         sharding.ShardCoordinator
	addrConverter            state.AddressConverter
	hasher                   hashing.Hasher
	marshalizer              marshal.Marshalizer
	singleSignKeyGen         crypto.KeyGenerator
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

// ProcessorsCreatorConfig is the struct containing the needed params to be
//  provided when initialising a new processorsCreator
type ProcessorsCreatorConfig struct {
	InterceptorContainer process.InterceptorContainer
	ResolverContainer    process.ResolverContainer

	Messenger                p2p.Messenger
	Blockchain               *blockchain.BlockChain
	DataPool                 data.TransientDataHolder
	ShardCoordinator         sharding.ShardCoordinator
	AddrConverter            state.AddressConverter
	Hasher                   hashing.Hasher
	Marshalizer              marshal.Marshalizer
	SingleSignKeyGen         crypto.KeyGenerator
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

// NewProcessorsCreator is responsible for creating a new processorsCreator object
func NewProcessorsCreator(config ProcessorsCreatorConfig) (*processorsCreator, error) {
	err := validateRequiredProcessCreatorParams(config)
	if err != nil {
		return nil, err
	}
	return &processorsCreator{
		interceptorContainer:     config.InterceptorContainer,
		resolverContainer:        config.ResolverContainer,
		messenger:                config.Messenger,
		blockchain:               config.Blockchain,
		dataPool:                 config.DataPool,
		shardCoordinator:         config.ShardCoordinator,
		addrConverter:            config.AddrConverter,
		hasher:                   config.Hasher,
		marshalizer:              config.Marshalizer,
		singleSignKeyGen:         config.SingleSignKeyGen,
		uint64ByteSliceConverter: config.Uint64ByteSliceConverter,
	}, nil
}

// CreateInterceptors creates the interceptors and initializes the interceptor container
func (p *processorsCreator) CreateInterceptors() error {
	err := p.createTxInterceptor()
	if err != nil {
		return err
	}

	err = p.createHdrInterceptor()
	if err != nil {
		return err
	}

	err = p.createTxBlockBodyInterceptor()
	if err != nil {
		return err
	}

	err = p.createPeerChBlockBodyInterceptor()
	if err != nil {
		return err
	}

	err = p.createStateBlockBodyInterceptor()
	if err != nil {
		return err
	}

	return nil
}

// CreateResolvers creates the resolvers and initializes the resolver container
func (p *processorsCreator) CreateResolvers() error {
	err := p.createTxResolver()
	if err != nil {
		return err
	}

	err = p.createHdrResolver()
	if err != nil {
		return err
	}

	err = p.createTxBlockBodyResolver()
	if err != nil {
		return err
	}

	err = p.createPeerChBlockBodyResolver()
	if err != nil {
		return err
	}

	err = p.createStateBlockBodyResolver()
	if err != nil {
		return err
	}

	return nil
}

// InterceptorContainer is a getter for interceptorContainer property
func (p *processorsCreator) InterceptorContainer() process.InterceptorContainer {
	return p.interceptorContainer
}

// ResolverContainer is a getter for resolverContainer property
func (p *processorsCreator) ResolverContainer() process.ResolverContainer {
	return p.resolverContainer
}

func (p *processorsCreator) createTxInterceptor() error {
	intercept, err := interceptor.NewTopicInterceptor(string(TransactionTopic), p.messenger, transaction.NewInterceptedTransaction())
	if err != nil {
		return err
	}

	txStorer := p.blockchain.GetStorer(blockchain.TransactionUnit)

	txInterceptor, err := transaction.NewTxInterceptor(
		intercept,
		p.dataPool.Transactions(),
		txStorer,
		p.addrConverter,
		p.hasher,
		p.singleSignKeyGen,
		p.shardCoordinator)

	if err != nil {
		return err
	}

	err = p.interceptorContainer.Add(string(TransactionTopic), txInterceptor)
	return err
}

func (p *processorsCreator) createHdrInterceptor() error {
	intercept, err := interceptor.NewTopicInterceptor(string(HeadersTopic), p.messenger, block.NewInterceptedHeader())
	if err != nil {
		return err
	}

	headerStorer := p.blockchain.GetStorer(blockchain.BlockHeaderUnit)

	hdrInterceptor, err := block.NewHeaderInterceptor(
		intercept,
		p.dataPool.Headers(),
		p.dataPool.HeadersNonces(),
		headerStorer,
		p.hasher,
		p.shardCoordinator,
	)

	if err != nil {
		return err
	}

	err = p.interceptorContainer.Add(string(HeadersTopic), hdrInterceptor)
	return err
}

func (p *processorsCreator) createTxBlockBodyInterceptor() error {
	intercept, err := interceptor.NewTopicInterceptor(string(TxBlockBodyTopic), p.messenger, block.NewInterceptedTxBlockBody())
	if err != nil {
		return err
	}

	txBlockBodyStorer := p.blockchain.GetStorer(blockchain.TxBlockBodyUnit)

	txBlockBodyInterceptor, err := block.NewGenericBlockBodyInterceptor(
		intercept,
		p.dataPool.TxBlocks(),
		txBlockBodyStorer,
		p.hasher,
		p.shardCoordinator,
	)

	if err != nil {
		return err
	}

	err = p.interceptorContainer.Add(string(TxBlockBodyTopic), txBlockBodyInterceptor)
	return err
}

func (p *processorsCreator) createPeerChBlockBodyInterceptor() error {
	intercept, err := interceptor.NewTopicInterceptor(string(PeerChBodyTopic), p.messenger, block.NewInterceptedPeerBlockBody())
	if err != nil {
		return err
	}

	peerBlockBodyStorer := p.blockchain.GetStorer(blockchain.PeerBlockBodyUnit)

	peerChBodyInterceptor, err := block.NewGenericBlockBodyInterceptor(
		intercept,
		p.dataPool.PeerChangesBlocks(),
		peerBlockBodyStorer,
		p.hasher,
		p.shardCoordinator,
	)

	if err != nil {
		return err
	}

	err = p.interceptorContainer.Add(string(PeerChBodyTopic), peerChBodyInterceptor)
	return err
}

func (p *processorsCreator) createStateBlockBodyInterceptor() error {
	intercept, err := interceptor.NewTopicInterceptor(string(StateBodyTopic), p.messenger, block.NewInterceptedStateBlockBody())
	if err != nil {
		return err
	}

	stateBlockBodyStorer := p.blockchain.GetStorer(blockchain.StateBlockBodyUnit)

	stateBodyInterceptor, err := block.NewGenericBlockBodyInterceptor(
		intercept,
		p.dataPool.StateBlocks(),
		stateBlockBodyStorer,
		p.hasher,
		p.shardCoordinator,
	)

	if err != nil {
		return err
	}

	err = p.interceptorContainer.Add(string(StateBodyTopic), stateBodyInterceptor)
	return err
}

func (p *processorsCreator) createTxResolver() error {
	resolve, err := resolver.NewTopicResolver(string(TransactionTopic), p.messenger, p.marshalizer)
	if err != nil {
		return err
	}

	txResolver, err := transaction.NewTxResolver(
		resolve,
		p.dataPool.Transactions(),
		p.blockchain.GetStorer(blockchain.TransactionUnit),
		p.marshalizer)

	if err != nil {
		return err
	}

	err = p.resolverContainer.Add(string(TransactionTopic), txResolver)
	return err
}

func (p *processorsCreator) createHdrResolver() error {
	resolve, err := resolver.NewTopicResolver(string(HeadersTopic), p.messenger, p.marshalizer)
	if err != nil {
		return err
	}

	hdrResolver, err := block.NewHeaderResolver(
		resolve,
		p.dataPool,
		p.blockchain.GetStorer(blockchain.BlockHeaderUnit),
		p.marshalizer,
		p.uint64ByteSliceConverter)

	if err != nil {
		return err
	}

	err = p.resolverContainer.Add(string(HeadersTopic), hdrResolver)
	return err
}

func (p *processorsCreator) createTxBlockBodyResolver() error {
	resolve, err := resolver.NewTopicResolver(string(TxBlockBodyTopic), p.messenger, p.marshalizer)
	if err != nil {
		return err
	}

	txBlkResolver, err := block.NewGenericBlockBodyResolver(
		resolve,
		p.dataPool.TxBlocks(),
		p.blockchain.GetStorer(blockchain.TxBlockBodyUnit),
		p.marshalizer)

	if err != nil {
		return err
	}

	err = p.resolverContainer.Add(string(TxBlockBodyTopic), txBlkResolver)
	return err
}

func (p *processorsCreator) createPeerChBlockBodyResolver() error {
	resolve, err := resolver.NewTopicResolver(string(PeerChBodyTopic), p.messenger, p.marshalizer)
	if err != nil {
		return err
	}

	peerChBlkResolver, err := block.NewGenericBlockBodyResolver(
		resolve,
		p.dataPool.PeerChangesBlocks(),
		p.blockchain.GetStorer(blockchain.PeerBlockBodyUnit),
		p.marshalizer)

	if err != nil {
		return err
	}

	err = p.resolverContainer.Add(string(PeerChBodyTopic), peerChBlkResolver)
	return err
}

func (p *processorsCreator) createStateBlockBodyResolver() error {
	resolve, err := resolver.NewTopicResolver(string(StateBodyTopic), p.messenger, p.marshalizer)
	if err != nil {
		return err
	}

	stateBlkResolver, err := block.NewGenericBlockBodyResolver(
		resolve,
		p.dataPool.StateBlocks(),
		p.blockchain.GetStorer(blockchain.StateBlockBodyUnit),
		p.marshalizer)

	if err != nil {
		return err
	}

	err = p.resolverContainer.Add(string(StateBodyTopic), stateBlkResolver)
	return err
}

func validateRequiredProcessCreatorParams(config ProcessorsCreatorConfig) error {
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
	if config.SingleSignKeyGen == nil {
		return process.ErrNilSingleSignKeyGen
	}
	if config.Uint64ByteSliceConverter == nil {
		return process.ErrNilUint64ByteSliceConverter
	}

	return nil
}
