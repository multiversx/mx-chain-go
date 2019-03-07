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

func createStubWireTopicHandler(matchStrToErrOnCreate string, matchStrToErrOnRegister string) process.WireTopicHandler {
	return &mock.WireTopicHandlerStub{
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

//------- NewInterceptorsContainerCreator

func TestNewInterceptorsContainerCreator_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	icc, err := factory.NewInterceptorsContainerCreator(
		nil,
		&mock.WireTopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icc)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptorsContainerCreator_NilWireTopicHandlerShouldErr(t *testing.T) {
	t.Parallel()

	icc, err := factory.NewInterceptorsContainerCreator(
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

	assert.Nil(t, icc)
	assert.Equal(t, process.ErrNilMessenger, err)
}

func TestNewInterceptorsContainerCreator_NilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	icc, err := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		&mock.WireTopicHandlerStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icc)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewInterceptorsContainerCreator_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	icc, err := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		&mock.WireTopicHandlerStub{},
		createBlockchain(),
		nil,
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icc)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptorsContainerCreator_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	icc, err := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		&mock.WireTopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		nil,
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icc)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptorsContainerCreator_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	icc, err := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		&mock.WireTopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		nil,
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icc)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewInterceptorsContainerCreator_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	icc, err := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		&mock.WireTopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		nil,
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icc)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewInterceptorsContainerCreator_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	icc, err := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		&mock.WireTopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		nil,
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.Nil(t, icc)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewInterceptorsContainerCreator_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	icc, err := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		&mock.WireTopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		nil,
		&mock.AddressConverterMock{})

	assert.Nil(t, icc)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewInterceptorsContainerCreator_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	icc, err := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		&mock.WireTopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		nil)

	assert.Nil(t, icc)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewInterceptorsContainerCreator_ShouldWork(t *testing.T) {
	t.Parallel()

	icc, err := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		&mock.WireTopicHandlerStub{},
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	assert.NotNil(t, icc)
	assert.Nil(t, err)
}

//------- Create

func TestInterceptorsContainerCreator_CreateTopicCreationTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	icc, _ := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler(string(factory.TransactionTopic), ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icc.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerCreator_CreateTopicCreationHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	icc, _ := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler(string(factory.HeadersTopic), ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icc.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerCreator_CreateTopicCreationMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icc, _ := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler(string(factory.MiniBlocksTopic), ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icc.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerCreator_CreateTopicCreationPeerChBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icc, _ := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler(string(factory.PeerChBodyTopic), ""),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icc.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerCreator_CreateRegisterTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	icc, _ := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler("", string(factory.TransactionTopic)),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icc.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerCreator_CreateRegisterHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	icc, _ := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler("", string(factory.HeadersTopic)),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icc.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerCreator_CreateRegisterMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icc, _ := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler("", string(factory.MiniBlocksTopic)),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icc.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerCreator_CreateRegisterPeerChBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	icc, _ := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		createStubWireTopicHandler("", string(factory.PeerChBodyTopic)),
		createBlockchain(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		mock.NewMultiSigner(),
		createDataPool(),
		&mock.AddressConverterMock{})

	container, err := icc.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestInterceptorsContainerCreator_CreateShouldWork(t *testing.T) {
	t.Parallel()

	icc, _ := factory.NewInterceptorsContainerCreator(
		mock.NewOneShardCoordinatorMock(),
		&mock.WireTopicHandlerStub{
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

	container, err := icc.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestInterceptorsContainerCreator_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	icc, _ := factory.NewInterceptorsContainerCreator(
		shardCoordinator,
		&mock.WireTopicHandlerStub{
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

	container, _ := icc.Create()

	assert.Equal(t, noOfShards+1+noOfShards+1, container.Len())
}
