package broadcast_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/broadcast"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/assert"
)

func TestMetaChainMessenger_NewMetaChainMessengerNilMarshalizerShouldFail(t *testing.T) {
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	mcm, err := broadcast.NewMetaChainMessenger(
		nil,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerNilMessengerShouldFail(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	mcm, err := broadcast.NewMetaChainMessenger(
		marshalizerMock,
		nil,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilMessenger, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerNilPrivateKeyShouldFail(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	mcm, err := broadcast.NewMetaChainMessenger(
		marshalizerMock,
		messengerMock,
		nil,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilPrivateKey, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerNilShardCoordinatorShouldFail(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	mcm, err := broadcast.NewMetaChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		nil,
		singleSignerMock,
	)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerNilSingleSignerShouldFail(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}

	mcm, err := broadcast.NewMetaChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		nil,
	)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilSingleSigner, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerShouldWork(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	mcm, err := broadcast.NewMetaChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.NotNil(t, mcm)
	assert.Equal(t, nil, err)
}

func TestMetaChainMessenger_BroadcastBlockShouldErrNilMetaHeader(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	mcm, _ := broadcast.NewMetaChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	err := mcm.BroadcastBlock(&block.Body{}, nil)
	assert.Equal(t, spos.ErrNilMetaHeader, err)
}

func TestMetaChainMessenger_BroadcastBlockShouldErrMockMarshalizer(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	marshalizerMock.Fail = true

	mcm, _ := broadcast.NewMetaChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	err := mcm.BroadcastBlock(&block.Body{}, &block.MetaBlock{})
	assert.Equal(t, mock.ErrMockMarshalizer, err)
}

func TestMetaChainMessenger_BroadcastBlockShouldWork(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	mcm, _ := broadcast.NewMetaChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	err := mcm.BroadcastBlock(&block.Body{}, &block.MetaBlock{})
	assert.Nil(t, err)
}

func TestMetaChainMessenger_BroadcastShardHeaderShouldWork(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	mcm, _ := broadcast.NewMetaChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	err := mcm.BroadcastShardHeader(nil)
	assert.Nil(t, err)
}

func TestMetaChainMessenger_BroadcastMiniBlocksShouldWork(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	mcm, _ := broadcast.NewMetaChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	err := mcm.BroadcastMiniBlocks(nil)
	assert.Nil(t, err)
}

func TestMetaChainMessenger_BroadcastTransactionsShouldWork(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	mcm, _ := broadcast.NewMetaChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	err := mcm.BroadcastTransactions(nil)
	assert.Nil(t, err)
}
