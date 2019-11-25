package metachain

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	processInterceptors "github.com/ElrondNetwork/elrond-go/process/interceptors"
	interceptorFactory "github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const numGoRoutines = 2000

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
	messenger              process.TopicHandler
	multiSigner            crypto.MultiSigner
	nodesCoordinator       sharding.NodesCoordinator
	blackList              process.BlackListHandler
	tpsBenchmark           *statistics.TpsBenchmark
	argInterceptorFactory  *interceptorFactory.ArgInterceptedDataFactory
	globalThrottler        process.InterceptorThrottler
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
	blockSingleSigner crypto.SingleSigner,
	keyGen crypto.KeyGenerator,
	blockKeyGen crypto.KeyGenerator,
	maxTxNonceDeltaAllowed int,
	txFeeHandler process.FeeHandler,
	blackList process.BlackListHandler,
) (*interceptorsContainerFactory, error) {

	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(messenger) {
		return nil, process.ErrNilMessenger
	}
	if check.IfNil(store) {
		return nil, process.ErrNilStore
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(multiSigner) {
		return nil, process.ErrNilMultiSigVerifier
	}
	if check.IfNil(dataPool) {
		return nil, process.ErrNilDataPoolHolder
	}
	if check.IfNil(nodesCoordinator) {
		return nil, process.ErrNilNodesCoordinator
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(addrConverter) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(singleSigner) {
		return nil, process.ErrNilSingleSigner
	}
	if check.IfNil(keyGen) {
		return nil, process.ErrNilKeyGen
	}
	if check.IfNil(txFeeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(blackList) {
		return nil, process.ErrNilBlackListHandler
	}
	if check.IfNil(blockKeyGen) {
		return nil, process.ErrNilKeyGen
	}
	if check.IfNil(blockSingleSigner) {
		return nil, process.ErrNilSingleSigner
	}

	argInterceptorFactory := &interceptorFactory.ArgInterceptedDataFactory{
		Marshalizer:      marshalizer,
		Hasher:           hasher,
		ShardCoordinator: shardCoordinator,
		NodesCoordinator: nodesCoordinator,
		MultiSigVerifier: multiSigner,
		KeyGen:           keyGen,
		BlockKeyGen:      blockKeyGen,
		Signer:           singleSigner,
		BlockSigner:      blockSingleSigner,
		AddrConv:         addrConverter,
		FeeHandler:       txFeeHandler,
	}

	icf := &interceptorsContainerFactory{
		shardCoordinator:       shardCoordinator,
		messenger:              messenger,
		store:                  store,
		marshalizer:            marshalizer,
		hasher:                 hasher,
		multiSigner:            multiSigner,
		dataPool:               dataPool,
		nodesCoordinator:       nodesCoordinator,
		blackList:              blackList,
		argInterceptorFactory:  argInterceptorFactory,
		maxTxNonceDeltaAllowed: maxTxNonceDeltaAllowed,
		accounts:               accounts,
	}

	var err error
	icf.globalThrottler, err = throttler.NewNumGoRoutineThrottler(numGoRoutines)
	if err != nil {
		return nil, err
	}

	return icf, nil
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

	keys, interceptorSlice, err = icf.generateTrieNodesInterceptors()
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

	hdrFactory, err := interceptorFactory.NewInterceptedMetaHeaderDataFactory(icf.argInterceptorFactory)
	if err != nil {
		return nil, nil, err
	}

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:       icf.dataPool.MetaBlocks(),
		HeadersNonces: icf.dataPool.HeadersNonces(),
		HdrValidator:  hdrValidator,
		BlackList:     icf.blackList,
	}
	hdrProcessor, err := processor.NewHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, nil, err
	}

	//only one metachain header topic
	interceptor, err := processInterceptors.NewSingleDataInterceptor(
		hdrFactory,
		hdrProcessor,
		icf.globalThrottler,
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

func (icf *interceptorsContainerFactory) createOneShardHeaderInterceptor(topic string) (process.Interceptor, error) {
	//TODO implement other HeaderHandlerProcessValidator that will check the header's nonce
	// against blockchain's latest nonce - k finality
	hdrValidator, err := dataValidators.NewNilHeaderValidator()
	if err != nil {
		return nil, err
	}

	hdrFactory, err := interceptorFactory.NewInterceptedShardHeaderDataFactory(icf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:       icf.dataPool.ShardHeaders(),
		HeadersNonces: icf.dataPool.HeadersNonces(),
		HdrValidator:  hdrValidator,
		BlackList:     icf.blackList,
	}
	hdrProcessor, err := processor.NewHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	interceptor, err := processInterceptors.NewSingleDataInterceptor(
		hdrFactory,
		hdrProcessor,
		icf.globalThrottler,
	)
	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(topic, interceptor, true)
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

func (icf *interceptorsContainerFactory) createOneTxInterceptor(topic string) (process.Interceptor, error) {
	txValidator, err := dataValidators.NewTxValidator(icf.accounts, icf.shardCoordinator, icf.maxTxNonceDeltaAllowed)
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

	txFactory, err := interceptorFactory.NewInterceptedTxDataFactory(icf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewMultiDataInterceptor(
		icf.marshalizer,
		txFactory,
		txProcessor,
		icf.globalThrottler,
	)
	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(topic, interceptor, true)
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

func (icf *interceptorsContainerFactory) createOneMiniBlocksInterceptor(topic string) (process.Interceptor, error) {
	argProcessor := &processor.ArgTxBodyInterceptorProcessor{
		MiniblockCache:   icf.dataPool.MiniBlocks(),
		Marshalizer:      icf.marshalizer,
		Hasher:           icf.hasher,
		ShardCoordinator: icf.shardCoordinator,
	}
	txBlockBodyProcessor, err := processor.NewTxBodyInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	txFactory, err := interceptorFactory.NewInterceptedTxBlockBodyDataFactory(icf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewSingleDataInterceptor(
		txFactory,
		txBlockBodyProcessor,
		icf.globalThrottler,
	)
	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (icf *interceptorsContainerFactory) generateTrieNodesInterceptors() ([]string, []process.Interceptor, error) {
	shardC := icf.shardCoordinator

	identifierTrieNodes := factory.TrieNodesTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)

	interceptor, err := icf.createOneTrieNodesInterceptor(identifierTrieNodes)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierTrieNodes}, []process.Interceptor{interceptor}, nil
}

func (icf *interceptorsContainerFactory) createOneTrieNodesInterceptor(topic string) (process.Interceptor, error) {
	trieNodesProcessor, err := processor.NewTrieNodesInterceptorProcessor(icf.dataPool.TrieNodes())
	if err != nil {
		return nil, err
	}

	trieNodesFactory, err := interceptorFactory.NewInterceptedTrieNodeDataFactory(icf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewMultiDataInterceptor(
		icf.marshalizer,
		trieNodesFactory,
		trieNodesProcessor,
		icf.globalThrottler,
	)
	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(topic, interceptor, true)
}

// IsInterfaceNil returns true if there is no value under the interface
func (icf *interceptorsContainerFactory) IsInterfaceNil() bool {
	if icf == nil {
		return true
	}
	return false
}
