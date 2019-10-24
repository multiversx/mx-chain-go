package heartbeat_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/stretchr/testify/assert"
)

func CreateHeartbeat() *heartbeat.Heartbeat {
	hb := heartbeat.Heartbeat{
		Payload:         []byte("Payload"),
		Pubkey:          []byte("PubKey"),
		Signature:       []byte("Signature"),
		ShardID:         0,
		VersionNumber:   "VersionNumber",
		NodeDisplayName: "NodeDisplayName",
	}
	return &hb
}

func TestNewMessageProcessor_SingleSignerNilShouldErr(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMessageProcessor(nil, &mock.KeyGenMock{}, &mock.MarshalizerMock{})

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilSingleSigner, err)
}

func TestNewMessageProcessor_KeyGeneratorNilShouldErr(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMessageProcessor(&mock.SinglesignMock{}, nil, &mock.MarshalizerMock{})

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilKeyGenerator, err)
}

func TestNewMessageProcessor_MarshalizerNilShouldErr(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMessageProcessor(&mock.SinglesignMock{}, &mock.KeyGenMock{}, nil)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilMarshalizer, err)
}

func TestNewMessageProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMessageProcessor(&mock.SinglesignMock{}, &mock.KeyGenMock{}, &mock.MarshalizerMock{})

	assert.Nil(t, err)
	assert.NotNil(t, mon)
}

func TestNewMessageProcessor_VerifyMessageAllSmallerShouldWork(t *testing.T) {
	t.Parallel()

	hbmi := CreateHeartbeat()
	err := heartbeat.VerifyLengths(hbmi)

	assert.Nil(t, err)
}

func TestNewMessageProcessor_VerifyMessageAllNilShouldWork(t *testing.T) {
	t.Parallel()

	hbmi := CreateHeartbeat()
	hbmi.Signature = nil
	hbmi.Payload = nil
	hbmi.Pubkey = nil
	hbmi.VersionNumber = ""
	hbmi.NodeDisplayName = ""

	err := heartbeat.VerifyLengths(hbmi)

	assert.Nil(t, err)
}

func TestNewMessageProcessor_VerifyMessageBiggerPublicKeyShouldErr(t *testing.T) {
	t.Parallel()

	hbmi := CreateHeartbeat()
	hbmi.Pubkey = make([]byte, heartbeat.GetMaxSizeInBytes()+1)
	err := heartbeat.VerifyLengths(hbmi)

	assert.NotNil(t, err)
}

func TestNewMessageProcessor_VerifyMessageAllSmallerPublicKeyShouldWork(t *testing.T) {
	t.Parallel()

	hbmi := CreateHeartbeat()
	hbmi.Pubkey = make([]byte, heartbeat.GetMaxSizeInBytes())
	err := heartbeat.VerifyLengths(hbmi)

	assert.Nil(t, err)
}

func TestNewMessageProcessor_VerifyMessageBiggerPayloadShouldErr(t *testing.T) {
	t.Parallel()

	hbmi := CreateHeartbeat()
	hbmi.Payload = make([]byte, heartbeat.GetMaxSizeInBytes()+1)
	err := heartbeat.VerifyLengths(hbmi)

	assert.NotNil(t, err)
}

func TestNewMessageProcessor_VerifyMessageSmallerPayloadShouldWork(t *testing.T) {
	t.Parallel()

	hbmi := CreateHeartbeat()
	hbmi.Payload = make([]byte, heartbeat.GetMaxSizeInBytes())
	err := heartbeat.VerifyLengths(hbmi)

	assert.Nil(t, err)
}

func TestNewMessageProcessor_VerifyMessageBiggerSignatureShouldErr(t *testing.T) {
	t.Parallel()

	hbmi := CreateHeartbeat()
	hbmi.Signature = make([]byte, heartbeat.GetMaxSizeInBytes()+1)
	err := heartbeat.VerifyLengths(hbmi)

	assert.NotNil(t, err)
}

func TestNewMessageProcessor_VerifyMessageSignatureShouldWork(t *testing.T) {
	t.Parallel()

	hbmi := CreateHeartbeat()
	hbmi.Signature = make([]byte, heartbeat.GetMaxSizeInBytes()-1)
	err := heartbeat.VerifyLengths(hbmi)

	assert.Nil(t, err)
}

func TestNewMessageProcessor_VerifyMessageBiggerNodeDisplayNameShouldErr(t *testing.T) {
	t.Parallel()

	hbmi := CreateHeartbeat()
	hbmi.NodeDisplayName = string(make([]byte, heartbeat.GetMaxSizeInBytes()+1))
	err := heartbeat.VerifyLengths(hbmi)

	assert.NotNil(t, err)
}

func TestNewMessageProcessor_VerifyMessageNodeDisplayNameShouldWork(t *testing.T) {
	t.Parallel()

	hbmi := CreateHeartbeat()
	hbmi.NodeDisplayName = string(make([]byte, heartbeat.GetMaxSizeInBytes()))
	err := heartbeat.VerifyLengths(hbmi)

	assert.Nil(t, err)
}

func TestNewMessageProcessor_VerifyMessageBiggerVersionNumberShouldErr(t *testing.T) {
	t.Parallel()

	hbmi := CreateHeartbeat()
	hbmi.VersionNumber = string(make([]byte, heartbeat.GetMaxSizeInBytes()+1))
	err := heartbeat.VerifyLengths(hbmi)

	assert.NotNil(t, err)
}

func TestNewMessageProcessor_VerifyMessageVersionNumberShouldWork(t *testing.T) {
	t.Parallel()

	hbmi := CreateHeartbeat()
	hbmi.VersionNumber = string(make([]byte, heartbeat.GetMaxSizeInBytes()))
	err := heartbeat.VerifyLengths(hbmi)

	assert.Nil(t, err)
}

func TestNewMessageProcessor_CreateHeartbeatFromP2pMessage(t *testing.T) {
	t.Parallel()

	hb := heartbeat.Heartbeat{
		Payload:         []byte("Payload"),
		Pubkey:          []byte("PubKey"),
		Signature:       []byte("signed"),
		ShardID:         0,
		VersionNumber:   "VersionNumber",
		NodeDisplayName: "NodeDisplayName",
	}

	marshalizer := &mock.MarshalizerMock{}

	marshalizer.UnmarshalHandler = func(obj interface{}, buff []byte) error {
		(obj.(*heartbeat.Heartbeat)).Pubkey = hb.Pubkey
		(obj.(*heartbeat.Heartbeat)).Payload = hb.Payload
		(obj.(*heartbeat.Heartbeat)).Signature = hb.Signature
		(obj.(*heartbeat.Heartbeat)).ShardID = hb.ShardID
		(obj.(*heartbeat.Heartbeat)).VersionNumber = hb.VersionNumber
		(obj.(*heartbeat.Heartbeat)).NodeDisplayName = hb.NodeDisplayName

		return nil
	}

	singleSigner := &mock.SinglesignMock{}
	keyGen := &mock.KeyGenMock{
		PublicKeyFromByteArrayMock: func(b []byte) (key crypto.PublicKey, e error) {
			return &mock.PublicKeyMock{}, nil
		},
	}

	mon, err := heartbeat.NewMessageProcessor(singleSigner, keyGen, marshalizer)

	message := &mock.P2PMessageStub{
		FromField:      nil,
		DataField:      make([]byte, 5),
		SeqNoField:     nil,
		TopicIDsField:  nil,
		SignatureField: nil,
		KeyField:       nil,
		PeerField:      "",
	}

	ret, err := mon.CreateHeartbeatFromP2pMessage(message)

	assert.Nil(t, err)
	assert.NotNil(t, ret)
}

func TestNewMessageProcessor_CreateHeartbeatFromP2pMessageWithNilDataShouldErr(t *testing.T) {
	t.Parallel()

	message := &mock.P2PMessageStub{
		FromField:      nil,
		DataField:      nil,
		SeqNoField:     nil,
		TopicIDsField:  nil,
		SignatureField: nil,
		KeyField:       nil,
		PeerField:      "",
	}

	mon, _ := heartbeat.NewMessageProcessor(&mock.SinglesignMock{}, &mock.KeyGenMock{}, &mock.MarshalizerMock{})

	ret, err := mon.CreateHeartbeatFromP2pMessage(message)

	assert.Nil(t, ret)
	assert.Equal(t, heartbeat.ErrNilDataToProcess, err)
}

func TestNewMessageProcessor_CreateHeartbeatFromP2pMessageWithUnmarshaliableDataShouldErr(t *testing.T) {
	t.Parallel()

	message := &mock.P2PMessageStub{
		FromField:      nil,
		DataField:      ([]byte("hello")),
		SeqNoField:     nil,
		TopicIDsField:  nil,
		SignatureField: nil,
		KeyField:       nil,
		PeerField:      "",
	}

	expectedErr := errors.New("Marshal didn't work")

	mon, _ := heartbeat.NewMessageProcessor(&mock.SinglesignMock{}, &mock.KeyGenMock{}, &mock.MarshalizerMock{
		UnmarshalHandler: func(obj interface{}, buff []byte) error {
			return expectedErr
		},
	})

	ret, err := mon.CreateHeartbeatFromP2pMessage(message)

	assert.Nil(t, ret)
	assert.Equal(t, expectedErr, err)
}

func TestNewMessageProcessor_CreateHeartbeatFromP2pMessageWithTooLongLengthsShouldErr(t *testing.T) {
	t.Parallel()

	length := 129
	buff := make([]byte, length)

	for i := 0; i < length; i++ {
		buff[i] = byte(97)
	}
	bigNodeName := string(buff)

	hb := heartbeat.Heartbeat{
		Payload:         []byte("Payload"),
		Pubkey:          []byte("PubKey"),
		Signature:       []byte("signed"),
		ShardID:         0,
		VersionNumber:   "VersionNumber",
		NodeDisplayName: bigNodeName,
	}

	marshalizer := &mock.MarshalizerMock{}

	marshalizer.UnmarshalHandler = func(obj interface{}, buff []byte) error {
		(obj.(*heartbeat.Heartbeat)).Pubkey = hb.Pubkey
		(obj.(*heartbeat.Heartbeat)).Payload = hb.Payload
		(obj.(*heartbeat.Heartbeat)).Signature = hb.Signature
		(obj.(*heartbeat.Heartbeat)).ShardID = hb.ShardID
		(obj.(*heartbeat.Heartbeat)).VersionNumber = hb.VersionNumber
		(obj.(*heartbeat.Heartbeat)).NodeDisplayName = hb.NodeDisplayName

		return nil
	}

	singleSigner := &mock.SinglesignMock{}

	keyGen := &mock.KeyGenMock{
		PublicKeyFromByteArrayMock: func(b []byte) (key crypto.PublicKey, e error) {
			return &mock.PublicKeyMock{}, nil
		},
	}

	mon, err := heartbeat.NewMessageProcessor(singleSigner, keyGen, marshalizer)

	message := &mock.P2PMessageStub{
		FromField:      nil,
		DataField:      make([]byte, 5),
		SeqNoField:     nil,
		TopicIDsField:  nil,
		SignatureField: nil,
		KeyField:       nil,
		PeerField:      "",
	}

	ret, err := mon.CreateHeartbeatFromP2pMessage(message)

	assert.Nil(t, ret)
	assert.Equal(t, heartbeat.ErrPropertyTooLong, err)
}

func TestNewMessageProcessor_CreateHeartbeatFromP2pNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	mon, _ := heartbeat.NewMessageProcessor(&mock.SinglesignMock{}, &mock.KeyGenMock{}, &mock.MarshalizerMock{})

	ret, err := mon.CreateHeartbeatFromP2pMessage(nil)

	assert.Nil(t, ret)
	assert.Equal(t, heartbeat.ErrNilMessage, err)
}
