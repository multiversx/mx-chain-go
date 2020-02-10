package interceptorscontainer

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
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
	processInterceptors "github.com/ElrondNetwork/elrond-go/process/interceptors"
	interceptorFactory "github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.InterceptorsContainerFactory = (*metaInterceptorsContainerFactory)(nil)

// metaInterceptorsContainerFactory will handle the creation the interceptors container for metachain
type metaInterceptorsContainerFactory struct {
	*baseInterceptorsContainerFactory
}

// NewMetaInterceptorsContainerFactory is responsible for creating a new interceptors factory object
func NewMetaInterceptorsContainerFactory(
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	messenger process.TopicHandler,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	multiSigner crypto.MultiSigner,
	dataPool dataRetriever.PoolsHolder,
	accounts state.AccountsAdapter,
	addrConverter state.AddressConverter,
	singleSigner crypto.SingleSigner,
	blockSingleSigner crypto.SingleSigner,
	keyGen crypto.KeyGenerator,
	blockKeyGen crypto.KeyGenerator,
	maxTxNonceDeltaAllowed int,
	txFeeHandler process.FeeHandler,
	blackList process.BlackListHandler,
	headerSigVerifier process.InterceptedHeaderSigVerifier,
	chainID []byte,
	sizeCheckDelta uint32,
	validityAttester process.ValidityAttester,
	epochStartTrigger process.EpochStartTriggerHandler,
) (*metaInterceptorsContainerFactory, error) {
	if sizeCheckDelta > 0 {
		marshalizer = marshal.NewSizeCheckUnmarshalizer(marshalizer, sizeCheckDelta)
	}
	err := checkBaseParams(
		shardCoordinator,
		accounts,
		marshalizer,
		hasher,
		store,
		dataPool,
		messenger,
		multiSigner,
		nodesCoordinator,
		blackList,
	)
	if err != nil {
		return nil, err
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
	if check.IfNil(blockKeyGen) {
		return nil, process.ErrNilKeyGen
	}
	if check.IfNil(blockSingleSigner) {
		return nil, process.ErrNilSingleSigner
	}
	if check.IfNil(headerSigVerifier) {
		return nil, process.ErrNilHeaderSigVerifier
	}
	if check.IfNil(epochStartTrigger) {
		return nil, process.ErrNilEpochStartTrigger
	}
	if len(chainID) == 0 {
		return nil, process.ErrInvalidChainID
	}
	if check.IfNil(validityAttester) {
		return nil, process.ErrNilValidityAttester
	}

	argInterceptorFactory := &interceptorFactory.ArgInterceptedDataFactory{
		Marshalizer:       marshalizer,
		Hasher:            hasher,
		ShardCoordinator:  shardCoordinator,
		NodesCoordinator:  nodesCoordinator,
		MultiSigVerifier:  multiSigner,
		KeyGen:            keyGen,
		BlockKeyGen:       blockKeyGen,
		Signer:            singleSigner,
		BlockSigner:       blockSingleSigner,
		AddrConv:          addrConverter,
		FeeHandler:        txFeeHandler,
		HeaderSigVerifier: headerSigVerifier,
		ChainID:           chainID,
		ValidityAttester:  validityAttester,
		EpochStartTrigger: epochStartTrigger,
	}

	base := &baseInterceptorsContainerFactory{
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

	icf := &metaInterceptorsContainerFactory{
		baseInterceptorsContainerFactory: base,
	}

	icf.globalThrottler, err = throttler.NewNumGoRoutineThrottler(numGoRoutines)
	if err != nil {
		return nil, err
	}

	return icf, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (micf *metaInterceptorsContainerFactory) Create() (process.InterceptorsContainer, error) {
	container := containers.NewInterceptorsContainer()

	keys, interceptorSlice, err := micf.generateMetablockInterceptors()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = micf.generateShardHeaderInterceptors()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = micf.generateTxInterceptors()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = micf.generateUnsignedTxsInterceptors()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = micf.generateMiniBlocksInterceptors()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	keys, interceptorSlice, err = micf.generateTrieNodesInterceptors()
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	return container, nil
}

//------- Metablock interceptor

func (micf *metaInterceptorsContainerFactory) generateMetablockInterceptors() ([]string, []process.Interceptor, error) {
	identifierHdr := factory.MetachainBlocksTopic

	//TODO implement other HeaderHandlerProcessValidator that will check the header's nonce
	// against blockchain's latest nonce - k finality
	hdrValidator, err := dataValidators.NewNilHeaderValidator()
	if err != nil {
		return nil, nil, err
	}

	hdrFactory, err := interceptorFactory.NewInterceptedMetaHeaderDataFactory(micf.argInterceptorFactory)
	if err != nil {
		return nil, nil, err
	}

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:      micf.dataPool.Headers(),
		HdrValidator: hdrValidator,
		BlackList:    micf.blackList,
	}
	hdrProcessor, err := processor.NewHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, nil, err
	}

	//only one metachain header topic
	interceptor, err := processInterceptors.NewSingleDataInterceptor(
		hdrFactory,
		hdrProcessor,
		micf.globalThrottler,
	)
	if err != nil {
		return nil, nil, err
	}

	_, err = micf.createTopicAndAssignHandler(identifierHdr, interceptor, true)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHdr}, []process.Interceptor{interceptor}, nil
}

//------- Shard header interceptors

func (micf *metaInterceptorsContainerFactory) generateShardHeaderInterceptors() ([]string, []process.Interceptor, error) {
	shardC := micf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	interceptorSlice := make([]process.Interceptor, noOfShards)

	//wire up to topics: shardBlocks_0_META, shardBlocks_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierHeader := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(idx)
		interceptor, err := micf.createOneShardHeaderInterceptor(identifierHeader)
		if err != nil {
			return nil, nil, err
		}

		keys[int(idx)] = identifierHeader
		interceptorSlice[int(idx)] = interceptor
	}

	return keys, interceptorSlice, nil
}

func (micf *metaInterceptorsContainerFactory) createOneShardHeaderInterceptor(topic string) (process.Interceptor, error) {
	//TODO implement other HeaderHandlerProcessValidator that will check the header's nonce
	// against blockchain's latest nonce - k finality
	hdrValidator, err := dataValidators.NewNilHeaderValidator()
	if err != nil {
		return nil, err
	}

	hdrFactory, err := interceptorFactory.NewInterceptedShardHeaderDataFactory(micf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:      micf.dataPool.Headers(),
		HdrValidator: hdrValidator,
		BlackList:    micf.blackList,
	}
	hdrProcessor, err := processor.NewHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	interceptor, err := processInterceptors.NewSingleDataInterceptor(
		hdrFactory,
		hdrProcessor,
		micf.globalThrottler,
	)
	if err != nil {
		return nil, err
	}

	return micf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (micf *metaInterceptorsContainerFactory) generateTrieNodesInterceptors() ([]string, []process.Interceptor, error) {
	shardC := micf.shardCoordinator

	keys := make([]string, 0)
	trieInterceptors := make([]process.Interceptor, 0)

	for i := uint32(0); i < shardC.NumberOfShards(); i++ {
		identifierTrieNodes := factory.ValidatorTrieNodesTopic + shardC.CommunicationIdentifier(i)
		interceptor, err := micf.createOneTrieNodesInterceptor(identifierTrieNodes)
		if err != nil {
			return nil, nil, err
		}

		keys = append(keys, identifierTrieNodes)
		trieInterceptors = append(trieInterceptors, interceptor)

		identifierTrieNodes = factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(i)
		interceptor, err = micf.createOneTrieNodesInterceptor(identifierTrieNodes)
		if err != nil {
			return nil, nil, err
		}

		keys = append(keys, identifierTrieNodes)
		trieInterceptors = append(trieInterceptors, interceptor)
	}

	identifierTrieNodes := factory.ValidatorTrieNodesTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	interceptor, err := micf.createOneTrieNodesInterceptor(identifierTrieNodes)
	if err != nil {
		return nil, nil, err
	}

	keys = append(keys, identifierTrieNodes)
	trieInterceptors = append(trieInterceptors, interceptor)

	identifierTrieNodes = factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	interceptor, err = micf.createOneTrieNodesInterceptor(identifierTrieNodes)
	if err != nil {
		return nil, nil, err
	}

	keys = append(keys, identifierTrieNodes)
	trieInterceptors = append(trieInterceptors, interceptor)

	return keys, trieInterceptors, nil
}

func (micf *metaInterceptorsContainerFactory) generateUnsignedTxsInterceptors() ([]string, []process.Interceptor, error) {
	shardC := micf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	interceptorSlice := make([]process.Interceptor, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierScr := factory.UnsignedTransactionTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := micf.createOneUnsignedTxInterceptor(identifierScr)
		if err != nil {
			return nil, nil, err
		}

		keys[int(idx)] = identifierScr
		interceptorSlice[int(idx)] = interceptor
	}

	return keys, interceptorSlice, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (micf *metaInterceptorsContainerFactory) IsInterfaceNil() bool {
	return micf == nil
}
