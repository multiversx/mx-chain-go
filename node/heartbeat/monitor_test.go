package heartbeat_test

import (
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/stretchr/testify/assert"
)

//------- NewMonitor

func TestNewMonitor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMonitor(
		nil,
		&mock.SinglesignMock{},
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
		0,
		[]string{""},
	)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilMessenger, err)
}

func TestNewMonitor_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMonitor(
		&mock.MessengerStub{},
		nil,
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
		0,
		[]string{""},
	)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilSingleSigner, err)
}

func TestNewMonitor_NilKeygenShouldErr(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMonitor(
		&mock.MessengerStub{},
		&mock.SinglesignMock{},
		nil,
		&mock.MarshalizerMock{},
		0,
		[]string{""},
	)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilKeyGenerator, err)
}

func TestNewMonitor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMonitor(
		&mock.MessengerStub{},
		&mock.SinglesignMock{},
		&mock.KeyGenMock{},
		nil,
		0,
		[]string{""},
	)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilMarshalizer, err)
}

func TestNewMonitor_EmptyPublicKeyListShouldErr(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMonitor(
		&mock.MessengerStub{},
		&mock.SinglesignMock{},
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
		0,
		make([]string, 0),
	)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrEmptyPublicKeyList, err)
}

func TestNewMonitor_OkValsShouldCreatePubkeyMap(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMonitor(
		&mock.MessengerStub{},
		&mock.SinglesignMock{},
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
		1,
		[]string{"pk1", "pk2"},
	)

	assert.NotNil(t, mon)
	assert.Nil(t, err)
	hbStatus := mon.GetHeartbeats()
	assert.Equal(t, 2, len(hbStatus))
}

//------- ProcessReceivedMessage

func TestMonitor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	mon, _ := heartbeat.NewMonitor(
		&mock.MessengerStub{},
		&mock.SinglesignMock{},
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
		1,
		[]string{"pk1"},
	)

	err := mon.ProcessReceivedMessage(nil)

	assert.Equal(t, heartbeat.ErrNilMessage, err)
}

func TestMonitor_ProcessReceivedMessageNilDataShouldErr(t *testing.T) {
	t.Parallel()

	mon, _ := heartbeat.NewMonitor(
		&mock.MessengerStub{},
		&mock.SinglesignMock{},
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{},
		0,
		[]string{"pk1"},
	)

	err := mon.ProcessReceivedMessage(&mock.P2PMessageStub{})

	assert.Equal(t, heartbeat.ErrNilDataToProcess, err)
}

func TestMonitor_ProcessReceivedMessageMarshalFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")

	mon, _ := heartbeat.NewMonitor(
		&mock.MessengerStub{},
		&mock.SinglesignMock{},
		&mock.KeyGenMock{},
		&mock.MarshalizerMock{
			UnmarshalHandler: func(obj interface{}, buff []byte) error {
				return errExpected
			},
		},
		1,
		[]string{"pk1"},
	)

	err := mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: []byte("")})

	assert.Equal(t, errExpected, err)
}

func TestMonitor_ProcessReceivedMessageWrongPubkeyShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")

	mon, _ := heartbeat.NewMonitor(
		&mock.MessengerStub{},
		&mock.SinglesignMock{},
		&mock.KeyGenMock{
			PublicKeyFromByteArrayMock: func(b []byte) (key crypto.PublicKey, e error) {
				return nil, errExpected
			},
		},
		&mock.MarshalizerMock{
			UnmarshalHandler: func(obj interface{}, buff []byte) error {
				return nil
			},
		},
		1,
		[]string{"pk1"},
	)

	err := mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: []byte("")})

	assert.Equal(t, errExpected, err)
}

func TestMonitor_ProcessReceivedMessageVerifyFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")

	mon, _ := heartbeat.NewMonitor(
		&mock.MessengerStub{},
		&mock.SinglesignStub{
			VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
				return errExpected
			},
		},
		&mock.KeyGenMock{
			PublicKeyFromByteArrayMock: func(b []byte) (key crypto.PublicKey, e error) {
				return nil, nil
			},
		},
		&mock.MarshalizerMock{
			UnmarshalHandler: func(obj interface{}, buff []byte) error {
				return nil
			},
		},
		1,
		[]string{"pk1"},
	)

	err := mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: []byte("")})

	assert.Equal(t, errExpected, err)
}

func TestMonitor_ProcessReceivedMessageShouldWork(t *testing.T) {
	t.Parallel()

	peerAddress := "peer address"
	pubKey := "pk1"

	mon, _ := heartbeat.NewMonitor(
		&mock.MessengerStub{
			PeerAddressCalled: func(pid p2p.PeerID) string {
				return peerAddress
			},
		},
		&mock.SinglesignStub{
			VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
				return nil
			},
		},
		&mock.KeyGenMock{
			PublicKeyFromByteArrayMock: func(b []byte) (key crypto.PublicKey, e error) {
				return nil, nil
			},
		},
		&mock.MarshalizerMock{
			UnmarshalHandler: func(obj interface{}, buff []byte) error {
				(obj.(*heartbeat.Heartbeat)).Pubkey = []byte(pubKey)
				return nil
			},
		},
		time.Second*1000,
		[]string{pubKey},
	)

	err := mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: []byte("")})
	assert.Nil(t, err)

	//a delay is mandatory for the go routine to finish its job
	time.Sleep(time.Second)

	hbStatus := mon.GetHeartbeats()
	assert.Equal(t, 1, len(hbStatus))
	assert.Equal(t, hex.EncodeToString([]byte(pubKey)), hbStatus[0].HexPublicKey)
	assert.Equal(t, 1, len(hbStatus[0].PeerHeartBeats))
	assert.True(t, hbStatus[0].PeerHeartBeats[0].IsActive)
	assert.Equal(t, peerAddress, hbStatus[0].PeerHeartBeats[0].P2PAddress)
}
