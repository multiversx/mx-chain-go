package heartbeat

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/stretchr/testify/assert"
)

func TestNewMessageHandler_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	mh, err := NewMessageProcessor(
		nil,
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, mh)
	assert.Equal(t, ErrNilSingleSigner, err)
}

func TestNewMessageHandler_NilKeygenShouldErr(t *testing.T) {
	t.Parallel()

	mh, err := NewMessageProcessor(
		&mock.SingleSignerMock{},
		nil,
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, mh)
	assert.Equal(t, ErrNilKeyGenerator, err)
}

func TestNewMessageHandler_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	mh, err := NewMessageProcessor(
		&mock.SingleSignerMock{},
		&mock.KeyGenMock{},
		nil,
	)

	assert.Nil(t, mh)
	assert.Equal(t, ErrNilMarshalizer, err)
}

func TestNewMessageHandler_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	mh, err := NewMessageProcessor(
		&mock.SingleSignerMock{},
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
	)

	assert.NotNil(t, mh)
	assert.Nil(t, err)
}

func TestHeartbeatMessageHandler_CreateHeartbeatFromP2pMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	mh, _ := NewMessageProcessor(
		&mock.SingleSignerMock{},
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
	)

	hbm, err := mh.CreateHeartbeatFromP2pMessage(nil)
	assert.Nil(t, hbm)
	assert.Equal(t, ErrNilMessage, err)
}

func TestHeartbeatMessageHandler_CreateHeartbeatFromP2pMessageNilMessageDataShouldErr(t *testing.T) {
	t.Parallel()

	mh, _ := NewMessageProcessor(
		&mock.SingleSignerMock{},
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
	)

	message := mock.P2PMessageMock{
		DataField: nil,
	}
	hbm, err := mh.CreateHeartbeatFromP2pMessage(&message)
	assert.Nil(t, hbm)
	assert.Equal(t, ErrNilDataToProcess, err)
}

func TestHeartbeatMessageHandler_CreateHeartbeatFromP2pMessageUnmarshalError(t *testing.T) {
	t.Parallel()

	mh, _ := NewMessageProcessor(
		&mock.SingleSignerMock{},
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
	)

	message := mock.P2PMessageMock{
		DataField: []byte("invalid heartbeat"),
	}
	hbm, err := mh.CreateHeartbeatFromP2pMessage(&message)
	assert.Nil(t, hbm)
	assert.NotNil(t, err)
}

func TestHeartbeatMessageHandler_CreateHeartbeatFromP2pMessageSignatureShouldError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("invalid signature")

	mh, _ := NewMessageProcessor(
		&mock.SingleSignerMock{
			VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
				return expectedErr
			},
		},
		&mock.KeyGenMock{
			PublicKeyFromByteArrayMock: func(b []byte) (crypto.PublicKey, error) {
				return &mock.PublicKeyMock{}, nil
			},
		},
		&mock.MarshalizerMock{},
	)

	heartbeat := HeartbeatDTO{
		NodeDisplayName: "node0",
	}
	heartbeatBytes, _ := json.Marshal(heartbeat)
	message := mock.P2PMessageMock{
		DataField: heartbeatBytes,
	}
	hbm, err := mh.CreateHeartbeatFromP2pMessage(&message)
	assert.Nil(t, hbm)
	assert.Equal(t, expectedErr, err)
}

func TestHeartbeatMessageHandler_CreateHeartbeatFromP2pOkValsShouldWork(t *testing.T) {
	t.Parallel()

	mh, _ := NewMessageProcessor(
		&mock.SingleSignerMock{
			VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
				return nil
			},
		},
		&mock.KeyGenMock{
			PublicKeyFromByteArrayMock: func(b []byte) (crypto.PublicKey, error) {
				return &mock.PublicKeyMock{}, nil
			},
		},
		&mock.MarshalizerMock{},
	)

	heartbeat := HeartbeatDTO{
		NodeDisplayName: "node0",
	}
	heartbeatBytes, _ := json.Marshal(heartbeat)
	message := mock.P2PMessageMock{
		DataField: heartbeatBytes,
	}
	hbm, err := mh.CreateHeartbeatFromP2pMessage(&message)
	assert.NotNil(t, hbm)
	assert.Nil(t, err)
}
