package broadcast_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/broadcast"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestShardChainMessenger_NewShardChainMessengerNilMarshalizerShouldFail(t *testing.T) {
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	scm, err := broadcast.NewShardChainMessenger(
		nil,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilMessengerShouldFail(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	scm, err := broadcast.NewShardChainMessenger(
		marshalizerMock,
		nil,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilMessenger, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilPrivateKeyShouldFail(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	scm, err := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		nil,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilPrivateKey, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilShardCoordinatorShouldFail(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	scm, err := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		nil,
		singleSignerMock,
	)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilSingleSignerShouldFail(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}

	scm, err := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		nil,
	)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilSingleSigner, err)
}

func TestShardChainMessenger_NewShardChainMessengerShouldWork(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	scm, err := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.NotNil(t, scm)
	assert.Equal(t, nil, err)
}

func TestShardChainMessenger_BroadcastBlockShouldErrNilBody(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	scm, _ := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	err := scm.BroadcastBlock(nil, &block.Header{})
	assert.Equal(t, spos.ErrNilBody, err)
}

func TestShardChainMessenger_BroadcastBlockShouldErrNilHeader(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	scm, _ := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	err := scm.BroadcastBlock(&block.Body{}, nil)
	assert.Equal(t, spos.ErrNilHeader, err)
}

func TestShardChainMessenger_BroadcastBlockShouldErrMockMarshalizer(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	marshalizerMock.Fail = true

	scm, _ := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	err := scm.BroadcastBlock(&block.Body{}, &block.Header{})
	assert.Equal(t, mock.ErrMockMarshalizer, err)
}

func TestShardChainMessenger_BroadcastBlockShouldWork(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	scm, _ := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	err := scm.BroadcastBlock(&block.Body{}, &block.Header{})
	assert.Nil(t, err)
}

func TestShardChainMessenger_BroadcastShardHeaderShouldErrNilHeader(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	scm, _ := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	err := scm.BroadcastShardHeader(nil)
	assert.Equal(t, spos.ErrNilHeader, err)
}

func TestShardChainMessenger_BroadcastShardHeaderShouldErrMockMarshalizer(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	marshalizerMock.Fail = true

	scm, _ := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	err := scm.BroadcastShardHeader(&block.Header{})
	assert.Equal(t, mock.ErrMockMarshalizer, err)
}

func TestShardChainMessenger_BroadcastShardHeaderShouldWork(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	scm, _ := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	err := scm.BroadcastShardHeader(&block.Header{})
	assert.Nil(t, err)
}

func TestShardChainMessenger_BroadcastMiniBlocksShouldBeDone(t *testing.T) {
	var channelCalled chan bool
	channelCalled = make(chan bool, 100)

	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			channelCalled <- true
		},
	}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	scm, _ := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	miniBlocks := make(map[uint32][]byte)
	miniBlocks[0] = make([]byte, 0)
	miniBlocks[1] = make([]byte, 0)
	miniBlocks[2] = make([]byte, 0)
	miniBlocks[3] = make([]byte, 0)
	err := scm.BroadcastMiniBlocks(miniBlocks)

	called := 0
	for i := 0; i < 4; i++ {
		select {
		case <-channelCalled:
			called++
		case <-time.After(time.Millisecond * 100):
			break
		}
	}

	assert.Nil(t, err)
	assert.Equal(t, 4, called)
}

func TestShardChainMessenger_BroadcastTransactionsShouldNotBeCalled(t *testing.T) {
	var channelCalled chan bool
	channelCalled = make(chan bool)

	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			channelCalled <- true
		},
	}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	scm, _ := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	transactions := make(map[string][][]byte)
	err := scm.BroadcastTransactions(transactions)

	wasCalled := false
	select {
	case <-channelCalled:
		wasCalled = true
	case <-time.After(time.Millisecond * 100):
	}

	assert.Nil(t, err)
	assert.False(t, wasCalled)

	transactions[factory.TransactionTopic] = make([][]byte, 0)
	err = scm.BroadcastTransactions(transactions)

	wasCalled = false
	select {
	case <-channelCalled:
		wasCalled = true
	case <-time.After(time.Millisecond * 100):
	}

	assert.Nil(t, err)
	assert.False(t, wasCalled)
}

func TestShardChainMessenger_BroadcastTransactionsShouldBeCalled(t *testing.T) {
	var channelCalled chan bool
	channelCalled = make(chan bool)

	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			channelCalled <- true
		},
	}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	scm, _ := broadcast.NewShardChainMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	transactions := make(map[string][][]byte)
	txs := make([][]byte, 0)
	txs = append(txs, []byte(""))
	transactions[factory.TransactionTopic] = txs
	err := scm.BroadcastTransactions(transactions)

	wasCalled := false
	select {
	case <-channelCalled:
		wasCalled = true
	case <-time.After(time.Millisecond * 100):
	}

	assert.Nil(t, err)
	assert.True(t, wasCalled)
}
