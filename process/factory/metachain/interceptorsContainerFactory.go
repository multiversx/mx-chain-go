package metachain

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	processInterceptors "github.com/ElrondNetwork/elrond-go/process/interceptors"
	interceptorFactory "github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const numGoRoutines = 100

type interceptorsContainerFactory struct {
	marshalizer           marshal.Marshalizer
	hasher                hashing.Hasher
	store                 dataRetriever.StorageService
	dataPool              dataRetriever.MetaPoolsHolder
	shardCoordinator      sharding.Coordinator
	messenger             process.TopicHandler
	multiSigner           crypto.MultiSigner
	chronologyValidator   process.ChronologyValidator
	tpsBenchmark          *statistics.TpsBenchmark
	argInterceptorFactory *interceptorFactory.ArgMetaInterceptedDataFactory
	globalThrottler       process.InterceptorThrottler
}

// NewInterceptorsContainerFactory is responsible for creating a new interceptors factory object
func NewInterceptorsContainerFactory(
	shardCoordinator sharding.Coordinator,
	messenger process.TopicHandler,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	multiSigner crypto.MultiSigner,
	dataPool dataRetriever.MetaPoolsHolder,
	chronologyValidator process.ChronologyValidator,
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
	if check.IfNil(chronologyValidator) {
		return nil, process.ErrNilChronologyValidator
	}

	argInterceptorFactory := &interceptorFactory.ArgMetaInterceptedDataFactory{
		Marshalizer:         marshalizer,
		Hasher:              hasher,
		ShardCoordinator:    shardCoordinator,
		ChronologyValidator: chronologyValidator,
		MultiSigVerifier:    multiSigner,
	}

	icf := &interceptorsContainerFactory{
		shardCoordinator:      shardCoordinator,
		messenger:             messenger,
		store:                 store,
		marshalizer:           marshalizer,
		hasher:                hasher,
		multiSigner:           multiSigner,
		dataPool:              dataPool,
		chronologyValidator:   chronologyValidator,
		argInterceptorFactory: argInterceptorFactory,
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

	hdrFactory, err := interceptorFactory.NewMetaInterceptedDataFactory(
		icf.argInterceptorFactory,
		interceptorFactory.InterceptedShardHeader,
	)

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:       icf.dataPool.ShardHeaders(),
		HeadersNonces: icf.dataPool.HeadersNonces(),
		HdrValidator:  hdrValidator,
	}
	hdrProcessor, err := processor.NewHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	//TODO rename interceptors2 when process/block/interceptors are no longer required
	interceptor, err := processInterceptors.NewSingleDataInterceptor(
		hdrFactory,
		hdrProcessor,
		icf.globalThrottler,
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
