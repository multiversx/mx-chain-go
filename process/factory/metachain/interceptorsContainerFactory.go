package metachain

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/metablock"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type interceptorsContainerFactory struct {
	marshalizer         marshal.Marshalizer
	hasher              hashing.Hasher
	store               dataRetriever.StorageService
	dataPool            dataRetriever.MetaPoolsHolder
	shardCoordinator    sharding.Coordinator
	messenger           process.TopicHandler
	multiSigner         crypto.MultiSigner
	chronologyValidator process.ChronologyValidator
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

	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if messenger == nil {
		return nil, process.ErrNilMessenger
	}
	if store == nil {
		return nil, process.ErrNilStore
	}
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if multiSigner == nil {
		return nil, process.ErrNilMultiSigVerifier
	}
	if dataPool == nil {
		return nil, process.ErrNilDataPoolHolder
	}
	if chronologyValidator == nil {
		return nil, process.ErrNilChronologyValidator
	}

	return &interceptorsContainerFactory{
		shardCoordinator:    shardCoordinator,
		messenger:           messenger,
		store:               store,
		marshalizer:         marshalizer,
		hasher:              hasher,
		multiSigner:         multiSigner,
		dataPool:            dataPool,
		chronologyValidator: chronologyValidator,
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
	metachainHeaderStorer := icf.store.GetStorer(dataRetriever.MetaBlockUnit)

	interceptor, err := interceptors.NewMetachainHeaderInterceptor(
		icf.marshalizer,
		icf.dataPool.MetaChainBlocks(),
		icf.dataPool.MetaBlockNonces(),
		nil,
		metachainHeaderStorer,
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
	hdrStorer := icf.store.GetStorer(dataRetriever.BlockHeaderUnit)
	interceptor, err := metablock.NewShardHeaderInterceptor(
		icf.marshalizer,
		icf.dataPool.ShardHeaders(),
		hdrStorer,
		icf.multiSigner,
		icf.hasher,
		icf.shardCoordinator,
		icf.chronologyValidator,
	)
	if err != nil {
		return nil, err
	}

	return icf.createTopicAndAssignHandler(identifier, interceptor, true)
}
