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
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

//------- NewInterceptorsResolversCreator

func TestNewInterceptorsResolversCreator_NilInterceptorContainerShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.InterceptorContainer = nil
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.Equal(t, process.ErrNilInterceptorContainer, err)
}

func TestNewInterceptorsResolversCreator_NilResolverContainerShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.ResolverContainer = nil
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.Equal(t, process.ErrNilResolverContainer, err)
}

func TestNewInterceptorsResolversCreator_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.Messenger = nil
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewInterceptorsResolversCreator_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.Blockchain = nil
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewInterceptorsResolversCreator_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.DataPool = nil
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewInterceptorsResolversCreator_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.ShardCoordinator = nil
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptorsResolversCreator_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.AddrConverter = nil
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewInterceptorsResolversCreator_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.Hasher = nil
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptorsResolversCreator_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.Marshalizer = nil
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptorsResolversCreator_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.KeyGen = nil
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewInterceptorsResolversCreator_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.SingleSigner = nil
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewInterceptorsResolversCreator_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.MultiSigner = nil
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewInterceptorsResolversCreator_NilUint64ByteSliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactoryConfig.Uint64ByteSliceConverter = nil
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.Nil(t, pFactory)
	assert.Equal(t, process.ErrNilUint64ByteSliceConverter, err)
}

func TestNewInterceptorsResolversCreator_ShouldWork(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()

	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)
	assert.NotNil(t, pFactory)
	assert.Nil(t, err)
}

func TestNewInterceptorsResolversCreator_ShouldNotModifyContainerPointers(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()

	cm1 := &mock.ObjectsContainerStub{}
	cm2 := &mock.ResolversContainerStub{}

	pFactoryConfig.InterceptorContainer = cm1
	pFactoryConfig.ResolverContainer = cm2

	pFactory, _ := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	assert.True(t, cm1 == pFactory.InterceptorContainer())
	assert.True(t, cm2 == pFactory.ResolverContainer())
}

//------- CreateInterceptors

func TestInterceptorsResolversCreator_CreateInterceptorsReturnsSuccessfully(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)
	assert.Nil(t, err)

	err = pFactory.CreateInterceptors()
	assert.Nil(t, err)
}

func TestInterceptorsResolversCreator_CreateInterceptorsNewTopicInterceptorErrorsWillMakeCreateInterceptorsError(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")

	pFactoryConfig := createConfig()
	pFactoryConfig.Messenger = &mock.MessengerStub{
		HasTopicCalled: func(name string) bool {
			return true
		},
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
			return errExpected
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
	}
	pFactory, _ := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	err := pFactory.CreateInterceptors()
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsResolversCreator_CreateResolversReturnsSuccessfully(t *testing.T) {
	t.Parallel()

	pFactoryConfig := createConfig()
	pFactory, err := factory.NewInterceptorsResolversCreator(pFactoryConfig)
	assert.Nil(t, err)

	err = pFactory.CreateResolvers()
	assert.Nil(t, err)
}

//------- CreateResolvers

func TestInterceptorsResolversCreator_CreateResolversNewTopicInterceptorErrorsWillMakeCreateInterceptorsError(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")

	pFactoryConfig := createConfig()
	pFactoryConfig.Messenger = &mock.MessengerStub{
		HasTopicCalled: func(name string) bool {
			return true
		},
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
			return errExpected
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
	}

	pFactory, _ := factory.NewInterceptorsResolversCreator(pFactoryConfig)

	err := pFactory.CreateResolvers()
	assert.Equal(t, errExpected, err)
}

func createConfig() factory.InterceptorsResolversConfig {

	mockMessenger := createMessenger()
	mockTransientDataPool := createDataPool()
	mockInterceptorContainer := &mock.ObjectsContainerStub{
		AddCalled: func(key string, val interface{}) error {
			return nil
		},
	}
	mockResolverContainer := &mock.ResolversContainerStub{
		AddCalled: func(key string, val process.Resolver) error {
			return nil
		},
	}

	return factory.InterceptorsResolversConfig{
		InterceptorContainer:     mockInterceptorContainer,
		ResolverContainer:        mockResolverContainer,
		Messenger:                mockMessenger,
		DataPool:                 mockTransientDataPool,
		Blockchain:               createBlockchain(),
		ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
		AddrConverter:            &mock.AddressConverterMock{},
		Hasher:                   mock.HasherMock{},
		Marshalizer:              &mock.MarshalizerMock{},
		MultiSigner:              mock.NewMultiSigner(),
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
	mockMessenger := &mock.MessengerStub{
		HasTopicCalled: func(name string) bool {
			return true
		},
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
			return nil
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
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
