package broadcast_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/broadcast"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCommonMessenger_BroadcastConsensusMessageShouldErrWhenSignMessageFail(t *testing.T) {
	err := errors.New("sign message error")
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return nil, err
		},
	}

	cm, _ := broadcast.NewCommonMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	msg := &consensus.Message{}
	err2 := cm.BroadcastConsensusMessage(msg)
	assert.Equal(t, err, err2)
}

func TestCommonMessenger_BroadcastConsensusMessageShouldWork(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return []byte(""), nil
		},
	}

	cm, _ := broadcast.NewCommonMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	msg := &consensus.Message{}
	err := cm.BroadcastConsensusMessage(msg)
	assert.Nil(t, err)
}

func TestCommonMessenger_SignMessageShouldErrWhenMarshalFail(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return []byte(""), nil
		},
	}
	marshalizerMock.Fail = true

	cm, _ := broadcast.NewCommonMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	msg := &consensus.Message{}
	_, err := cm.SignMessage(msg)
	assert.Equal(t, err, mock.ErrMockMarshalizer)
}

func TestCommonMessenger_SignMessageShouldErrWhenSignFail(t *testing.T) {
	err := errors.New("sign message error")
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return nil, err
		},
	}

	cm, _ := broadcast.NewCommonMessenger(
		marshalizerMock,
		messengerMock,
		privateKeyMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	msg := &consensus.Message{}
	_, err2 := cm.SignMessage(msg)
	assert.Equal(t, err, err2)
}
