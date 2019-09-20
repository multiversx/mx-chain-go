package shard

import (
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	blockInterceptors "github.com/ElrondNetwork/elrond-go/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	interceptorFactory "github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/process/unsigned"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const numGoRoutinesInTxInterceptors = 100

type interceptorsContainerFactory struct {
	accounts              state.AccountsAdapter
	shardCoordinator      sharding.Coordinator
	messenger             process.TopicHandler
	store                 dataRetriever.StorageService
	marshalizer           marshal.Marshalizer
	hasher                hashing.Hasher
	keyGen                crypto.KeyGenerator
	singleSigner          crypto.SingleSigner
	multiSigner           crypto.MultiSigner
	dataPool              dataRetriever.PoolsHolder
	addrConverter         state.AddressConverter
	chronologyValidator   process.ChronologyValidator
	argInterceptorFactory *interceptorFactory.ArgInterceptedDataFactory
	globalTxThrottler     process.InterceptorThrottler
}

// NewInterceptorsContainerFactory is responsible for creating a new interceptors factory object
func NewInterceptorsContainerFactory(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	messenger process.TopicHandler,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	keyGen crypto.KeyGenerator,
	singleSigner crypto.SingleSigner,
	multiSigner crypto.MultiSigner,
	dataPool dataRetriever.PoolsHolder,
	addrConverter state.AddressConverter,
	chronologyValidator process.ChronologyValidator,
) (*interceptorsContainerFactory, error) {
	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if messenger == nil || messenger.IsInterfaceNil() {
		return nil, process.ErrNilMessenger
	}
	if store == nil || store.IsInterfaceNil() {
		return nil, process.ErrNilBlockChain
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if keyGen == nil || keyGen.IsInterfaceNil() {
		return nil, process.ErrNilKeyGen
	}
	if singleSigner == nil || singleSigner.IsInterfaceNil() {
		return nil, process.ErrNilSingleSigner
	}
	if multiSigner == nil || multiSigner.IsInterfaceNil() {
		return nil, process.ErrNilMultiSigVerifier
	}
	if dataPool == nil || dataPool.IsInterfaceNil() {
		return nil, process.ErrNilDataPoolHolder
	}
	if addrConverter == nil || addrConverter.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if chronologyValidator == nil || chronologyValidator.IsInterfaceNil() {
		return nil, process.ErrNilChronologyValidator
	}

	argInterceptorFactory := &interceptorFactory.ArgInterceptedDataFactory{
		Marshalizer:      marshalizer,
		Hasher:           hasher,
		KeyGen:           keyGen,
		Signer:           singleSigner,
		AddrConv:         addrConverter,
		ShardCoordinator: shardCoordinator,
	}

	icf := &interceptorsContainerFactory{
		accounts:              accounts,
		shardCoordinator:      shardCoordinator,
		messenger:             messenger,
		store:                 store,
		marshalizer:           marshalizer,
		hasher:                hasher,
		keyGen:                keyGen,
		singleSigner:          singleSigner,
		multiSigner:           multiSigner,
		dataPool:              dataPool,
		addrConverter:         addrConverter,
		chronologyValidator:   chronologyValidator,
		argInterceptorFactory: argInterceptorFactory,
	}

	var err error
	icf.globalTxThrottler, err = throttler.NewNumGoRoutineThrottler(numGoRoutinesInTxInterceptors)
	if err != nil {
		return nil, err
	}

	return icf, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (icf *interceptorsContainerFactory) Create() (process.InterceptorsContainer, error) {
	container := containers.NewInterceptorsContainer()

	keys, interceptorSlice, err := icf.generateTxInterceptors()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateUnsignedTxsInterceptors()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateHdrInterceptor()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateMiniBlocksInterceptors()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generatePeerChBlockBodyInterceptor()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateMetachainHeaderInterceptor()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (icf *interceptorsContainerFactory) createTopicAndAssignHandler(
	topic string,
	interceptor process.Interceptor,
	createChannel bool,
) (process.Interceptor, error) {

	err := icf.messenger.CreateTopic(topic, createChannel)
	if err != nil {
		return nil, err
	}

	return interceptor, icf.messenger.RegisterMessageProcessor(topic, interceptor)
}

//------- Tx interceptors

func (icf *interceptorsContainerFactory) generateTxInterceptors() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	interceptorSlice := make([]process.Interceptor, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := factory.TransactionTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := icf.createOneTxInterceptor(identifierTx)
		if err != nil {
			return nil, nil, err
		}

		keys[int(idx)] = identifierTx
		interceptorSlice[int(idx)] = interceptor
	}

	//tx interceptor for metachain topic
	identifierTx := factory.TransactionTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)

	interceptor, err := icf.createOneTxInterceptor(identifierTx)
	if err != nil {
		return nil, nil, err
	}

	keys = append(keys, identifierTx)
	interceptorSlice = append(interceptorSlice, interceptor)
	return keys, interceptorSlice, nil
}

func (icf *interceptorsContainerFactory) createOneTxInterceptor(identifier string) (process.Interceptor, error) {
	txValidator, err := dataValidators.NewTxValidator(icf.accounts, icf.shardCoordinator)
	if err != nil {
		return nil, err
	}

	argProcessor := &processor.ArgTxInterceptorProcessor{
		ShardedDataCache: icf.dataPool.Transactions(),
		TxValidator:      txValidator,
	}
	txProcessor, err := processor.NewTxInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	txFactory, err := interceptorFactory.NewInterceptedDataFactory(
		icf.argInterceptorFactory,
		interceptorFactory.InterceptedTx,
	)

	interceptor, err := interceptors.NewMultiDataInterceptor(
		icf.marshalizer,
		txFactory,
		txProcessor,
		icf.globalTxThrottler,
	)

	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- Unsigned transactions interceptors

func (icf *interceptorsContainerFactory) generateUnsignedTxsInterceptors() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	interceptorSlice := make([]process.Interceptor, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierScr := factory.UnsignedTransactionTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := icf.createOneUnsignedTxInterceptor(identifierScr)
		if err != nil {
			return nil, nil, err
		}

		keys[int(idx)] = identifierScr
		interceptorSlice[int(idx)] = interceptor
	}

	identifierTx := factory.UnsignedTransactionTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)

	interceptor, err := icf.createOneUnsignedTxInterceptor(identifierTx)
	if err != nil {
		return nil, nil, err
	}

	keys = append(keys, identifierTx)
	interceptorSlice = append(interceptorSlice, interceptor)
	return keys, interceptorSlice, nil
}

func (icf *interceptorsContainerFactory) createOneUnsignedTxInterceptor(identifier string) (process.Interceptor, error) {
	uTxStorer := icf.store.GetStorer(dataRetriever.UnsignedTransactionUnit)

	interceptor, err := unsigned.NewUnsignedTxInterceptor(
		icf.marshalizer,
		icf.dataPool.UnsignedTransactions(),
		uTxStorer,
		icf.addrConverter,
		icf.hasher,
		icf.shardCoordinator)

	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- Hdr interceptor

func (icf *interceptorsContainerFactory) generateHdrInterceptor() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator
	//TODO implement other HeaderHandlerProcessValidator that will check the header's nonce
	// against blockchain's latest nonce - k finality
	hdrValidator, err := dataValidators.NewNilHeaderValidator()
	if err != nil {
		return nil, nil, err
	}

	//only one intrashard header topic
	identifierHdr := factory.HeadersTopic + shardC.CommunicationIdentifier(shardC.SelfId())
	interceptor, err := blockInterceptors.NewHeaderInterceptor(
		icf.marshalizer,
		icf.dataPool.Headers(),
		icf.dataPool.HeadersNonces(),
		hdrValidator,
		icf.multiSigner,
		icf.hasher,
		icf.shardCoordinator,
		icf.chronologyValidator,
	)
	if err != nil {
		return nil, nil, err
	}
	_, err = icf.createTopicAndAssignHandler(identifierHdr, interceptor, true)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHdr}, []process.Interceptor{interceptor}, nil
}

//------- MiniBlocks interceptors

func (icf *interceptorsContainerFactory) generateMiniBlocksInterceptors() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	interceptorSlice := make([]process.Interceptor, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := icf.createOneMiniBlocksInterceptor(identifierMiniBlocks)
		if err != nil {
			return nil, nil, err
		}

		keys[int(idx)] = identifierMiniBlocks
		interceptorSlice[int(idx)] = interceptor
	}

	return keys, interceptorSlice, nil
}

func (icf *interceptorsContainerFactory) createOneMiniBlocksInterceptor(identifier string) (process.Interceptor, error) {
	txBlockBodyStorer := icf.store.GetStorer(dataRetriever.MiniBlockUnit)

	interceptor, err := blockInterceptors.NewTxBlockBodyInterceptor(
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

//------- PeerChBlocks interceptor

func (icf *interceptorsContainerFactory) generatePeerChBlockBodyInterceptor() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator

	//only one intrashard peer change blocks topic
	identifierPeerCh := factory.PeerChBodyTopic + shardC.CommunicationIdentifier(shardC.SelfId())
	peerBlockBodyStorer := icf.store.GetStorer(dataRetriever.PeerChangesUnit)

	interceptor, err := blockInterceptors.NewPeerBlockBodyInterceptor(
		icf.marshalizer,
		icf.dataPool.PeerChangesBlocks(),
		peerBlockBodyStorer,
		icf.hasher,
		shardC,
	)
	if err != nil {
		return nil, nil, err
	}
	_, err = icf.createTopicAndAssignHandler(identifierPeerCh, interceptor, true)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierPeerCh}, []process.Interceptor{interceptor}, nil
}

//------- MetachainHeader interceptors

func (icf *interceptorsContainerFactory) generateMetachainHeaderInterceptor() ([]string, []process.Interceptor, error) {
	identifierHdr := factory.MetachainBlocksTopic
	//TODO implement other HeaderHandlerProcessValidator that will check the header's nonce
	// against blockchain's latest nonce - k finality
	hdrValidator, err := dataValidators.NewNilHeaderValidator()
	if err != nil {
		return nil, nil, err
	}

	interceptor, err := blockInterceptors.NewMetachainHeaderInterceptor(
		icf.marshalizer,
		icf.dataPool.MetaBlocks(),
		icf.dataPool.HeadersNonces(),
		hdrValidator,
		icf.multiSigner,
		icf.hasher,
		icf.shardCoordinator,
		icf.chronologyValidator,
	)
	if err != nil {
		return nil, nil, err
	}
	_, err = icf.createTopicAndAssignHandler(identifierHdr, interceptor, true)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHdr}, []process.Interceptor{interceptor}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (icf *interceptorsContainerFactory) IsInterfaceNil() bool {
	if icf == nil {
		return true
	}
	return false
}
