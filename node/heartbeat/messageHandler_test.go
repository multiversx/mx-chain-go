package heartbeat

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/stretchr/testify/assert"
)

func TestNewMessageHandler(t *testing.T) {
	t.Parallel()

	mh := NewMessageHandler(
		&mock.SingleSignerMock{},
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
	)

	assert.NotNil(t, mh)
}

func TestHeartbeatMessageHandler_CreateHeartbeatFromP2pMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	mh := NewMessageHandler(
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

	mh := NewMessageHandler(
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

	mh := NewMessageHandler(
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

	mh := NewMessageHandler(
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

	mh := NewMessageHandler(
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

func TestHeartbeatMessageHandler_convertToExportedStruct(t *testing.T) {
	t.Parallel()

	mh := NewMessageHandler(
		&mock.SingleSignerMock{},
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
	)

	timeHandler := func() time.Time {
		return time.Now()
	}
	hbmi, err := newHeartbeatMessageInfo(10, false, time.Now(), timeHandler)
	assert.Nil(t, err)
	res := mh.convertToExportedStruct(hbmi)

	assert.Equal(t, hbmi.nodeDisplayName, res.NodeDisplayName)
	assert.Equal(t, hbmi.isActive, res.IsActive)
}

func TestHeartbeatMessageHandler_convertFromExportedStruct(t *testing.T) {
	t.Parallel()

	mh := NewMessageHandler(
		&mock.SingleSignerMock{},
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
	)

	hbmiDto := HeartbeatDTO{
		NodeDisplayName: "node0",
		IsValidator:     false,
	}

	hbmi := mh.convertFromExportedStruct(hbmiDto, 10)

	assert.Equal(t, hbmiDto.NodeDisplayName, hbmi.nodeDisplayName)
	assert.Equal(t, hbmiDto.IsValidator, hbmi.isActive)
}
