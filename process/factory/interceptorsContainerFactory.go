package factory

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type interceptorsContainerFactory struct {
	shardCoordinator sharding.ShardCoordinator
	messenger        process.TopicHandler
	blockchain       *blockchain.BlockChain
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	keyGen           crypto.KeyGenerator
	singleSigner     crypto.SingleSigner
	multiSigner      crypto.MultiSigner
	dataPool         data.TransientDataHolder
	addrConverter    state.AddressConverter
}

// NewInterceptorsContainerFactory is responsible for creating a new interceptors factory object
func NewInterceptorsContainerFactory(
	shardCoordinator sharding.ShardCoordinator,
	messenger process.TopicHandler,
	blockchain *blockchain.BlockChain,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	keyGen crypto.KeyGenerator,
	singleSigner crypto.SingleSigner,
	multiSigner crypto.MultiSigner,
	dataPool data.TransientDataHolder,
	addrConverter state.AddressConverter,
) (*interceptorsContainerFactory, error) {

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
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if keyGen == nil {
		return nil, process.ErrNilKeyGen
	}
	if singleSigner == nil {
		return nil, process.ErrNilSingleSigner
	}
	if multiSigner == nil {
		return nil, process.ErrNilMultiSigVerifier
	}
	if dataPool == nil {
		return nil, process.ErrNilDataPoolHolder
	}
	if addrConverter == nil {
		return nil, process.ErrNilAddressConverter
	}

	return &interceptorsContainerFactory{
		shardCoordinator: shardCoordinator,
		messenger:        messenger,
		blockchain:       blockchain,
		marshalizer:      marshalizer,
		hasher:           hasher,
		keyGen:           keyGen,
		singleSigner:     singleSigner,
		multiSigner:      multiSigner,
		dataPool:         dataPool,
		addrConverter:    addrConverter,
	}, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (icf *interceptorsContainerFactory) Create() (process.InterceptorsContainer, error) {
	container := containers.NewInterceptorsContainer()

	err := icf.addTxInterceptors(container)
	if err != nil {
		return nil, err
	}

	err = icf.addHdrInterceptors(container)
	if err != nil {
		return nil, err
	}

	err = icf.addMiniBlocksInterceptors(container)
	if err != nil {
		return nil, err
	}

	err = icf.addPeerChBlockBodyInterceptors(container)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (icf *interceptorsContainerFactory) createTopicAndAssignHandler(
	topic string,
	interceptor process.Interceptor,
	createChannel bool) (process.Interceptor, error) {

	err := icf.messenger.CreateTopic(topic, createChannel)
	if err != nil {
		return nil, err
	}

	return interceptor, icf.messenger.RegisterMessageProcessor(topic, interceptor)
}

//------- Tx interceptors

func (icf *interceptorsContainerFactory) addTxInterceptors(container process.InterceptorsContainer) error {
	shardC := icf.shardCoordinator

	noOfShards := shardC.NoShards()

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := TransactionTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := icf.createOneTxInterceptor(identifierTx)
		if err != nil {
			return err
		}

		err = container.Add(identifierTx, interceptor)
		if err != nil {
			return err
		}
	}

	return nil
}

func (icf *interceptorsContainerFactory) createOneTxInterceptor(identifier string) (process.Interceptor, error) {
	txStorer := icf.blockchain.GetStorer(blockchain.TransactionUnit)

	interceptor, err := transaction.NewTxInterceptor(
		icf.marshalizer,
		icf.dataPool.Transactions(),
		txStorer,
		icf.addrConverter,
		icf.hasher,
		icf.singleSigner,
		icf.keyGen,
		icf.shardCoordinator)

	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- Hdr interceptors

func (icf *interceptorsContainerFactory) addHdrInterceptors(container process.InterceptorsContainer) error {
	shardC := icf.shardCoordinator

	//only one intrashard header topic
	identifierHdr := HeadersTopic + shardC.CommunicationIdentifier(shardC.ShardForCurrentNode())

	interceptor, err := icf.createOneHdrInterceptor(identifierHdr)
	if err != nil {
		return err
	}

	return container.Add(identifierHdr, interceptor)
}

func (icf *interceptorsContainerFactory) createOneHdrInterceptor(identifier string) (process.Interceptor, error) {
	headerStorer := icf.blockchain.GetStorer(blockchain.BlockHeaderUnit)

	interceptor, err := interceptors.NewHeaderInterceptor(
		icf.marshalizer,
		icf.dataPool.Headers(),
		icf.dataPool.HeadersNonces(),
		headerStorer,
		icf.multiSigner,
		icf.hasher,
		icf.shardCoordinator,
	)

	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- MiniBlocks interceptors

func (icf *interceptorsContainerFactory) addMiniBlocksInterceptors(container process.InterceptorsContainer) error {
	shardC := icf.shardCoordinator

	noOfShards := shardC.NoShards()

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := MiniBlocksTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := icf.createOneMiniBlocksInterceptor(identifierMiniBlocks)
		if err != nil {
			return err
		}

		err = container.Add(identifierMiniBlocks, interceptor)
		if err != nil {
			return err
		}
	}

	return nil
}

func (icf *interceptorsContainerFactory) createOneMiniBlocksInterceptor(identifier string) (process.Interceptor, error) {
	txBlockBodyStorer := icf.blockchain.GetStorer(blockchain.MiniBlockUnit)

	interceptor, err := interceptors.NewMiniBlocksInterceptor(
		icf.marshalizer,
		icf.dataPool.MiniBlocks(),
		txBlockBodyStorer,
		icf.hasher,
		icf.shardCoordinator,
	)

	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- PeerChBlocks interceptors

func (icf *interceptorsContainerFactory) addPeerChBlockBodyInterceptors(container process.InterceptorsContainer) error {
	shardC := icf.shardCoordinator

	//only one intrashard peer change blocks topic
	identifierPeerCh := PeerChBodyTopic + shardC.CommunicationIdentifier(shardC.ShardForCurrentNode())

	interceptor, err := icf.createOnePeerChBlockBodyInterceptor(identifierPeerCh)
	if err != nil {
		return err
	}

	return container.Add(identifierPeerCh, interceptor)
}

func (icf *interceptorsContainerFactory) createOnePeerChBlockBodyInterceptor(identifier string) (process.Interceptor, error) {
	peerBlockBodyStorer := icf.blockchain.GetStorer(blockchain.PeerChangesUnit)

	interceptor, err := interceptors.NewPeerBlockBodyInterceptor(
		icf.marshalizer,
		icf.dataPool.PeerChangesBlocks(),
		peerBlockBodyStorer,
		icf.hasher,
		icf.shardCoordinator,
	)

	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}
