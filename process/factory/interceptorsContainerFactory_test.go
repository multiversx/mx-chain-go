package factory_test

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var errExpected = errors.New("expected error")

func createStubWireTopicHandler(matchStrToErrOnCreate string, matchStrToErrOnRegister string) process.TopicHandler {
	return &mock.TopicHandlerStub{
		CreateTopicCalled: func(name string, createChannelForTopic bool) error {
			if matchStrToErrOnCreate == "" {
				return nil
			}

			if strings.Contains(name, matchStrToErrOnCreate) {
				return errExpected
			}

			return nil
		},
		RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
			if matchStrToErrOnRegister == "" {
				return nil
			}

			if strings.Contains(topic, matchStrToErrOnRegister) {
				return errExpected
			}

			return nil
		},
	}
}

//------- NewInterceptorsContainerFactory

func TestNewInterceptorsContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := factory.NewInterceptorsContainerFactory(
		nil,
		&mock.TopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptorsContainerFactory_NilWireTopicHandlerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		nil,
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewInterceptorsContainerFactory_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewInterceptorsContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createBlockchain(),
		nil,
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptorsContainerFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		nil,
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptorsContainerFactory_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		nil,
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewInterceptorsContainerFactory_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		nil,
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewInterceptorsContainerFactory_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		nil,
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewInterceptorsContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		nil,
		&mock.AddressConverterMock{})

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewInterceptorsContainerFactory_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	icf, err := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		nil)

	assert.Nil(t, icf)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewInterceptorsContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	icf, err := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.NotNil(t, icf)
	assert.Nil(t, err)
}

//------- Create

func TestInterceptorsContainerFactory_CreateTopicCreationTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler(factory.TransactionTopic, ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateTopicCreationHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler(factory.HeadersTopic, ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateTopicCreationMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler(factory.MiniBlocksTopic, ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateTopicCreationPeerChBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler(factory.PeerChBodyTopic, ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler("", factory.TransactionTopic),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler("", factory.HeadersTopic),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler("", factory.MiniBlocksTopic),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateRegisterPeerChBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icf, _ := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler("", factory.PeerChBodyTopic),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	icf, _ := factory.NewInterceptorsContainerFactory(
		mock.NewOneShardCoordinatorMock(),
		&mock.TopicHandlerStub{
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestInterceptorsContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	icf, _ := factory.NewInterceptorsContainerFactory(
		shardCoordinator,
		&mock.TopicHandlerStub{
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, _ := icf.Create()

	assert.Equal(t, noOfShards+1+noOfShards+1, container.Len())
}
