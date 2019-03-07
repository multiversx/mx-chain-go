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

type interceptorsContainerCreator struct {
	shardCoordinator sharding.ShardCoordinator
	messenger        process.WireTopicHandler
	blockchain       *blockchain.BlockChain
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	keyGen           crypto.KeyGenerator
	singleSigner     crypto.SingleSigner
	multiSigner      crypto.MultiSigner
	dataPool         data.TransientDataHolder
	addrConverter    state.AddressConverter
}

// NewInterceptorsContainerCreator is responsible for creating a new interceptors factory object
func NewInterceptorsContainerCreator(
	shardCoordinator sharding.ShardCoordinator,
	messenger process.WireTopicHandler,
	blockchain *blockchain.BlockChain,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	keyGen crypto.KeyGenerator,
	singleSigner crypto.SingleSigner,
	multiSigner crypto.MultiSigner,
	dataPool data.TransientDataHolder,
	addrConverter state.AddressConverter,
) (*interceptorsContainerCreator, error) {

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

	return &interceptorsContainerCreator{
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
func (icc *interceptorsContainerCreator) Create() (process.InterceptorsContainer, error) {
	container := containers.NewInterceptorsContainer()

	err := icc.addTxInterceptors(container)
	if err != nil {
		return nil, err
	}

	err = icc.addHdrInterceptors(container)
	if err != nil {
		return nil, err
	}

	err = icc.addMiniBlocksInterceptors(container)
	if err != nil {
		return nil, err
	}

	err = icc.addPeerChBlockBodyInterceptors(container)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (icc *interceptorsContainerCreator) createTopicAndAssignHandler(
	topic string,
	interceptor process.Interceptor,
	createPipe bool) (process.Interceptor, error) {

	err := icc.messenger.CreateTopic(topic, createPipe)
	if err != nil {
		return nil, err
	}

	return interceptor, icc.messenger.RegisterMessageProcessor(topic, interceptor)
}

//------- Tx interceptors

func (icc *interceptorsContainerCreator) addTxInterceptors(container process.InterceptorsContainer) error {
	shardC := icc.shardCoordinator

	for idx := uint32(0); idx < shardC.NoShards(); idx++ {
		identifierTx := string(TransactionTopic) + shardC.CrossShardIdentifier(idx)

		interceptor, err := icc.createOneTxInterceptor(identifierTx)
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

func (icc *interceptorsContainerCreator) createOneTxInterceptor(identifier string) (process.Interceptor, error) {
	txStorer := icc.blockchain.GetStorer(blockchain.TransactionUnit)

	interceptor, err := transaction.NewTxInterceptor(
		icc.marshalizer,
		icc.dataPool.Transactions(),
		txStorer,
		icc.addrConverter,
		icc.hasher,
		icc.singleSigner,
		icc.keyGen,
		icc.shardCoordinator)

	if err != nil {
		return nil, err
	}

	return icc.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- Hdr interceptors

func (icc *interceptorsContainerCreator) addHdrInterceptors(container process.InterceptorsContainer) error {
	shardC := icc.shardCoordinator

	//only one intrashard header topic
	identifierHdr := string(HeadersTopic) + shardC.CrossShardIdentifier(shardC.ShardForCurrentNode())

	interceptor, err := icc.createOneHdrInterceptor(identifierHdr)
	if err != nil {
		return err
	}

	return container.Add(identifierHdr, interceptor)
}

func (icc *interceptorsContainerCreator) createOneHdrInterceptor(identifier string) (process.Interceptor, error) {
	headerStorer := icc.blockchain.GetStorer(blockchain.BlockHeaderUnit)

	interceptor, err := interceptors.NewHeaderInterceptor(
		icc.marshalizer,
		icc.dataPool.Headers(),
		icc.dataPool.HeadersNonces(),
		headerStorer,
		icc.multiSigner,
		icc.hasher,
		icc.shardCoordinator,
	)

	if err != nil {
		return nil, err
	}

	return icc.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- MiniBlocks interceptors

func (icc *interceptorsContainerCreator) addMiniBlocksInterceptors(container process.InterceptorsContainer) error {
	shardC := icc.shardCoordinator

	for idx := uint32(0); idx < shardC.NoShards(); idx++ {
		identifierMiniBlocks := string(MiniBlocksTopic) + shardC.CrossShardIdentifier(idx)

		interceptor, err := icc.createOneMiniBlocksInterceptor(identifierMiniBlocks)
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

func (icc *interceptorsContainerCreator) createOneMiniBlocksInterceptor(identifier string) (process.Interceptor, error) {
	txBlockBodyStorer := icc.blockchain.GetStorer(blockchain.MiniBlockUnit)

	interceptor, err := interceptors.NewMiniBlocksInterceptor(
		icc.marshalizer,
		icc.dataPool.MiniBlocks(),
		txBlockBodyStorer,
		icc.hasher,
		icc.shardCoordinator,
	)

	if err != nil {
		return nil, err
	}

	return icc.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- PeerChBlocks interceptors

func (icc *interceptorsContainerCreator) addPeerChBlockBodyInterceptors(container process.InterceptorsContainer) error {
	shardC := icc.shardCoordinator

	//only one intrashard peer change blocks topic
	identifierPeerCh := string(PeerChBodyTopic) + shardC.CrossShardIdentifier(shardC.ShardForCurrentNode())

	interceptor, err := icc.createOnePeerChBlockBodyInterceptor(identifierPeerCh)
	if err != nil {
		return err
	}

	return container.Add(identifierPeerCh, interceptor)
}

func (icc *interceptorsContainerCreator) createOnePeerChBlockBodyInterceptor(identifier string) (process.Interceptor, error) {
	peerBlockBodyStorer := icc.blockchain.GetStorer(blockchain.PeerChangesUnit)

	interceptor, err := interceptors.NewPeerBlockBodyInterceptor(
		icc.marshalizer,
		icc.dataPool.PeerChangesBlocks(),
		peerBlockBodyStorer,
		icc.hasher,
		icc.shardCoordinator,
	)

	if err != nil {
		return nil, err
	}

	return icc.createTopicAndAssignHandler(identifier, interceptor, true)
}
