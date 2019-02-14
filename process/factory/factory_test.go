package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
)

func TestProcessorsCreator_NilInterceptorContainerShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.InterceptorContainer = nil
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil interceptor container")
}

func TestProcessorsCreator_NilResolverContainerShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.ResolverContainer = nil
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil resolver container")
}

func TestProcessorsCreator_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.Messenger = nil
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil Messenger")
}

func TestProcessorsCreator_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.Blockchain = nil
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil block chain")
}

func TestProcessorsCreator_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.DataPool = nil
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil data pool")
}

func TestProcessorsCreator_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.ShardCoordinator = nil
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil shard coordinator")
}

func TestProcessorsCreator_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.AddrConverter = nil
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil AddressConverter")
}

func TestProcessorsCreator_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.Hasher = nil
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil Hasher")
}

func TestProcessorsCreator_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.Marshalizer = nil
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil Marshalizer")
}

func TestProcessorsCreator_NilSingleSignKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.KeyGen = nil
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.NotNil(t, err)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestProcessorsCreator_NilUint64ByteSliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.Uint64ByteSliceConverter = nil
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil byte slice converter")
}

func TestProcessorsCreator_ShouldWork(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	count := 10
	pFactoryConfig.InterceptorContainer = &mock.InterceptorContainer{
		LenCalled: func() int {
			return count
		},
	}
	pFactoryConfig.ResolverContainer = &mock.ResolverContainer{
		LenCalled: func() int {
			return count
		},
	}
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)

	assert.NotNil(t, pFactory)
	assert.NotNil(t, pFactory.ResolverContainer())
	assert.NotNil(t, pFactory.InterceptorContainer())
	assert.Nil(t, err)
	assert.Equal(t, pFactory.ResolverContainer().Len(), count)
	assert.Equal(t, pFactory.InterceptorContainer().Len(), count)
}

func TestCreateInterceptors_ReturnsSuccessfully(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)
	assert.Nil(t, err)

	err = pFactory.CreateInterceptors()
	assert.Nil(t, err)
}

func TestCreateInterceptors_NewTopicInterceptorErrorsWillMakeCreateInterceptorsError(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactory, _ := factory.NewProcessorsCreator(pFactoryConfig)

	pFactory.SetMessenger(nil)
	err := pFactory.CreateInterceptors()
	assert.NotNil(t, err)
}

func TestCreateResolvers_ReturnsSuccessfully(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactory, err := factory.NewProcessorsCreator(pFactoryConfig)
	assert.Nil(t, err)

	err = pFactory.CreateResolvers()
	assert.Nil(t, err)
}

func TestCreateResolvers_NewTopicInterceptorErrorsWillMakeCreateInterceptorsError(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactory, _ := factory.NewProcessorsCreator(pFactoryConfig)

	pFactory.SetMessenger(nil)
	err := pFactory.CreateResolvers()
	assert.NotNil(t, err)
}

func createConfig() factory.ProcessorsCreatorConfig {

	mockMessenger := createMessenger()
	mockTransientDataPool := createDataPool()
	mockInterceptorContainer := createInterceptorContainer()
	mockResolverContainer := createResolverContainer()

	return factory.ProcessorsCreatorConfig{
		InterceptorContainer:     mockInterceptorContainer,
		ResolverContainer:        mockResolverContainer,
		Messenger:                mockMessenger,
		DataPool:                 mockTransientDataPool,
		Blockchain:               createBlockchain(),
		ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
		AddrConverter:            &mock.AddressConverterMock{},
		Hasher:                   mock.HasherMock{},
		Marshalizer:              &mock.MarshalizerMock{},
		SingleSigner:             &mock.SignerMock{},
		KeyGen:                   &mock.SingleSignKeyGenMock{},
		Uint64ByteSliceConverter: &mock.Uint64ByteSliceConverterMock{},
	}
}

func createBlockchain() *blockchain.BlockChain {
	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{})

	return blkc
}

func createMessenger() p2p.Messenger {
	mockMessenger := mock.NewMessengerStub()
	mockMessenger.GetTopicCalled = func(name string) *p2p.Topic {
		topic := &p2p.Topic{}
		topic.RegisterTopicValidator = func(v pubsub.Validator) error {
			return nil
		}
		return topic
	}
	return mockMessenger
}

func createDataPool() data.TransientDataHolder {
	mockTransientDataPool := &mock.TransientDataPoolMock{}
	mockTransientDataPool.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	mockTransientDataPool.HeadersCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{}
	}
	mockTransientDataPool.HeadersNoncesCalled = func() data.Uint64Cacher {
		return &mock.Uint64CacherStub{}
	}
	mockTransientDataPool.TxBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	mockTransientDataPool.PeerChangesBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	mockTransientDataPool.StateBlocksCalled = func() storage.Cacher {
		return &mock.CacherStub{}
	}
	return mockTransientDataPool
}

func createInterceptorContainer() process.InterceptorContainer {
	mockInterceptorContainer := &mock.InterceptorContainer{}
	mockInterceptorContainer.AddCalled = func(key string, interceptor process.Interceptor) error {
		return nil
	}
	return mockInterceptorContainer
}
func createResolverContainer() process.ResolverContainer {
	mockResolverContainer := &mock.ResolverContainer{}
	mockResolverContainer.AddCalled = func(key string, resolver process.Resolver) error {
		return nil
	}
	return mockResolverContainer
}
