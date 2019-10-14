package metachain

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const maxGoRoutineTxInterceptor = 100

type interceptorsContainerFactory struct {
	accounts               state.AccountsAdapter
	addrConverter          state.AddressConverter
	singleSigner           crypto.SingleSigner
	keyGen                 crypto.KeyGenerator
	maxTxNonceDeltaAllowed int
	txFeeHandler           process.FeeHandler
	txInterceptorThrottler process.InterceptorThrottler
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	store                  dataRetriever.StorageService
	dataPool               dataRetriever.MetaPoolsHolder
	shardCoordinator       sharding.Coordinator
	nodesCoordinator       sharding.NodesCoordinator
	messenger              process.TopicHandler
	multiSigner            crypto.MultiSigner
	tpsBenchmark           *statistics.TpsBenchmark
}

// NewInterceptorsContainerFactory is responsible for creating a new interceptors factory object
func NewInterceptorsContainerFactory(
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	messenger process.TopicHandler,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	multiSigner crypto.MultiSigner,
	dataPool dataRetriever.MetaPoolsHolder,
	accounts state.AccountsAdapter,
	addrConverter state.AddressConverter,
	singleSigner crypto.SingleSigner,
	keyGen crypto.KeyGenerator,
	maxTxNonceDeltaAllowed int,
	txFeeHandler process.FeeHandler,
) (*interceptorsContainerFactory, error) {

	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if nodesCoordinator == nil || nodesCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilNodesCoordinator
	}
	if messenger == nil {
		return nil, process.ErrNilMessenger
	}
	if store == nil || store.IsInterfaceNil() {
		return nil, process.ErrNilStore
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if multiSigner == nil || multiSigner.IsInterfaceNil() {
		return nil, process.ErrNilMultiSigVerifier
	}
	if dataPool == nil || dataPool.IsInterfaceNil() {
		return nil, process.ErrNilDataPoolHolder
	}
	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if addrConverter == nil || addrConverter.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if singleSigner == nil || singleSigner.IsInterfaceNil() {
		return nil, process.ErrNilSingleSigner
	}
	if keyGen == nil || keyGen.IsInterfaceNil() {
		return nil, process.ErrNilKeyGen
	}
	if txFeeHandler == nil || txFeeHandler.IsInterfaceNil() {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	txInterceptorThrottler, err := throttler.NewNumGoRoutineThrottler(maxGoRoutineTxInterceptor)
	if err != nil {
		return nil, err
	}

	return &interceptorsContainerFactory{
		accounts:               accounts,
		addrConverter:          addrConverter,
		singleSigner:           singleSigner,
		keyGen:                 keyGen,
		maxTxNonceDeltaAllowed: maxTxNonceDeltaAllowed,
		txFeeHandler:           txFeeHandler,
		txInterceptorThrottler: txInterceptorThrottler,
		shardCoordinator:       shardCoordinator,
		nodesCoordinator:       nodesCoordinator,
		messenger:              messenger,
		store:                  store,
		marshalizer:            marshalizer,
		hasher:                 hasher,
		multiSigner:            multiSigner,
		dataPool:               dataPool,
	}, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (icf *interceptorsContainerFactory) Create() (process.InterceptorsContainer, error) {
	container := containers.NewInterceptorsContainer()

	keys, interceptorSlice, err := icf.generateMetablockInterceptor()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateShardHeaderInterceptors()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = icf.generateTxInterceptors()
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

//------- Metablock interceptor

func (icf *interceptorsContainerFactory) generateMetablockInterceptor() ([]string, []process.Interceptor, error) {
	identifierHdr := factory.MetachainBlocksTopic

	//TODO implement other HeaderHandlerProcessValidator that will check the header's nonce
	// against blockchain's latest nonce - k finality
	hdrValidator, err := dataValidators.NewNilHeaderValidator()
	if err != nil {
		return nil, nil, err
	}

	interceptor, err := interceptors.NewMetachainHeaderInterceptor(
		icf.marshalizer,
		icf.dataPool.MetaChainBlocks(),
		icf.dataPool.HeadersNonces(),
		hdrValidator,
		icf.multiSigner,
		icf.hasher,
		icf.shardCoordinator,
		icf.nodesCoordinator,
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

//------- Shard header interceptors

func (icf *interceptorsContainerFactory) generateShardHeaderInterceptors() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	interceptorSlice := make([]process.Interceptor, noOfShards)

	//wire up to topics: shardHeadersForMetachain_0_META, shardHeadersForMetachain_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierHeader := factory.ShardHeadersForMetachainTopic + shardC.CommunicationIdentifier(idx)
		interceptor, err := icf.createOneShardHeaderInterceptor(identifierHeader)
		if err != nil {
			return nil, nil, err
		}

		keys[int(idx)] = identifierHeader
		interceptorSlice[int(idx)] = interceptor
	}

	return keys, interceptorSlice, nil
}

func (icf *interceptorsContainerFactory) createOneShardHeaderInterceptor(identifier string) (process.Interceptor, error) {
	//TODO implement other HeaderHandlerProcessValidator that will check the header's nonce
	// against blockchain's latest nonce - k finality
	hdrValidator, err := dataValidators.NewNilHeaderValidator()
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewHeaderInterceptor(
		icf.marshalizer,
		icf.dataPool.ShardHeaders(),
		icf.dataPool.HeadersNonces(),
		hdrValidator,
		icf.multiSigner,
		icf.hasher,
		icf.shardCoordinator,
		icf.nodesCoordinator,
	)
	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
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
	txValidator, err := dataValidators.NewTxValidator(icf.accounts, icf.shardCoordinator, icf.maxTxNonceDeltaAllowed)
	if err != nil {
		return nil, err
	}

	interceptor, err := transaction.NewTxInterceptor(
		icf.marshalizer,
		icf.dataPool.Transactions(),
		txValidator,
		icf.addrConverter,
		icf.hasher,
		icf.singleSigner,
		icf.keyGen,
		icf.shardCoordinator,
		icf.txInterceptorThrottler,
		icf.txFeeHandler,
	)

	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}

//------- MiniBlocks interceptors

func (icf *interceptorsContainerFactory) generateMiniBlocksInterceptors() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards+1)
	interceptorSlice := make([]process.Interceptor, noOfShards+1)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := icf.createOneMiniBlocksInterceptor(identifierMiniBlocks)
		if err != nil {
			return nil, nil, err
		}

		keys[int(idx)] = identifierMiniBlocks
		interceptorSlice[int(idx)] = interceptor
	}

	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)

	interceptor, err := icf.createOneMiniBlocksInterceptor(identifierMiniBlocks)
	if err != nil {
		return nil, nil, err
	}

	keys[noOfShards] = identifierMiniBlocks
	interceptorSlice[noOfShards] = interceptor

	return keys, interceptorSlice, nil
}

func (icf *interceptorsContainerFactory) createOneMiniBlocksInterceptor(identifier string) (process.Interceptor, error) {
	txBlockBodyStorer := icf.store.GetStorer(dataRetriever.MiniBlockUnit)

	interceptor, err := interceptors.NewTxBlockBodyInterceptor(
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

// IsInterfaceNil returns true if there is no value under the interface
func (icf *interceptorsContainerFactory) IsInterfaceNil() bool {
	if icf == nil {
		return true
	}
	return false
}
