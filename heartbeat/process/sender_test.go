package process_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/ElrondNetwork/elrond-go/heartbeat/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

// ------- NewSender

func createMockArgHeartbeatSender() process.ArgHeartbeatSender {
	return process.ArgHeartbeatSender{
		PeerMessenger: &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {},
		},
		PeerSignatureHandler: &mock.PeerSignatureHandler{},
		PrivKey:              &mock.PrivateKeyStub{},
		Marshalizer: &testscommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (i []byte, e error) {
				return nil, nil
			},
		},
		Topic:                 "",
		ShardCoordinator:      &mock.ShardCoordinatorMock{},
		PeerTypeProvider:      &mock.PeerTypeProviderStub{},
		StatusHandler:         &statusHandlerMock.AppStatusHandlerStub{},
		VersionNumber:         "v0.1",
		NodeDisplayName:       "undefined",
		HardforkTrigger:       &testscommon.HardforkTriggerStub{},
		CurrentBlockProvider:  &mock.CurrentBlockProviderStub{},
		RedundancyHandler:     &mock.RedundancyHandlerStub{},
		EpochNotifier:         &epochNotifier.EpochNotifierStub{},
		HeartbeatDisableEpoch: 1,
	}
}

func TestNewSender_NilP2PMessengerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	arg.PeerMessenger = nil
	sender, err := process.NewSender(arg)

	assert.Nil(t, sender)
	assert.Equal(t, heartbeat.ErrNilMessenger, err)
}

func TestNewSender_NilPeerSignatureHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	arg.PeerSignatureHandler = nil
	sender, err := process.NewSender(arg)

	assert.Nil(t, sender)
	assert.Equal(t, heartbeat.ErrNilPeerSignatureHandler, err)
}

func TestNewSender_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	arg.ShardCoordinator = nil
	sender, err := process.NewSender(arg)

	assert.Nil(t, sender)
	assert.Equal(t, heartbeat.ErrNilShardCoordinator, err)
}

func TestNewSender_NilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	arg.PrivKey = nil
	sender, err := process.NewSender(arg)

	assert.Nil(t, sender)
	assert.True(t, errors.Is(err, heartbeat.ErrNilPrivateKey))
}

func TestNewSender_NilMarshallerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	arg.Marshalizer = nil
	sender, err := process.NewSender(arg)

	assert.Nil(t, sender)
	assert.Equal(t, heartbeat.ErrNilMarshaller, err)
}

func TestNewSender_NilPeerTypeProviderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	arg.PeerTypeProvider = nil
	sender, err := process.NewSender(arg)

	assert.Nil(t, sender)
	assert.Equal(t, heartbeat.ErrNilPeerTypeProvider, err)
}

func TestNewSender_NilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	arg.StatusHandler = nil
	sender, err := process.NewSender(arg)

	assert.Nil(t, sender)
	assert.Equal(t, heartbeat.ErrNilAppStatusHandler, err)
}

func TestNewSender_NilHardforkTriggerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	arg.HardforkTrigger = nil
	sender, err := process.NewSender(arg)

	assert.Nil(t, sender)
	assert.Equal(t, heartbeat.ErrNilHardforkTrigger, err)
}

func TestNewSender_PropertyTooLongShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	arg.VersionNumber = strings.Repeat("a", process.MaxSizeInBytes+1)
	sender, err := process.NewSender(arg)

	assert.Nil(t, sender)
	assert.True(t, errors.Is(err, heartbeat.ErrPropertyTooLong))
}

func TestNewSender_NilCurrentBlockProviderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	arg.CurrentBlockProvider = nil
	sender, err := process.NewSender(arg)

	assert.Nil(t, sender)
	assert.True(t, errors.Is(err, heartbeat.ErrNilCurrentBlockProvider))
}

func TestNewSender_NilRedundancyHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	arg.RedundancyHandler = nil
	sender, err := process.NewSender(arg)

	assert.Nil(t, sender)
	assert.True(t, errors.Is(err, heartbeat.ErrNilRedundancyHandler))
}

func TestNewSender_RedundancyHandlerReturnsANilObserverPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	arg.RedundancyHandler = &mock.RedundancyHandlerStub{
		ObserverPrivateKeyCalled: func() crypto.PrivateKey {
			return nil
		},
	}
	sender, err := process.NewSender(arg)

	assert.Nil(t, sender)
	assert.True(t, errors.Is(err, heartbeat.ErrNilPrivateKey))
}

func TestNewSender_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	arg.EpochNotifier = nil
	sender, err := process.NewSender(arg)

	assert.Nil(t, sender)
	assert.Equal(t, heartbeat.ErrNilEpochNotifier, err)
}

func TestNewSender_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatSender()
	sender, err := process.NewSender(arg)

	assert.NotNil(t, sender)
	assert.Nil(t, err)
}

// ------- SendHeartbeat

func TestSender_SendHeartbeatGeneratePublicKeyErrShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	testSendHeartbeat(t, errExpected, nil, nil)
}

func TestSender_SendHeartbeatSignErrShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	testSendHeartbeat(t, nil, errExpected, nil)
}

func TestSender_SendHeartbeatMarshalizerErrShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("err")
	testSendHeartbeat(t, nil, nil, expectedErr)
}

func testSendHeartbeat(t *testing.T, pubKeyErr, signErr, marshalErr error) {
	var expectedErr error
	pubKey := &mock.PublicKeyMock{
		ToByteArrayHandler: func() (i []byte, e error) {
			expectedErr = pubKeyErr
			return nil, pubKeyErr
		},
	}

	arg := createMockArgHeartbeatSender()
	arg.PrivKey = &mock.PrivateKeyStub{
		GeneratePublicHandler: func() crypto.PublicKey {
			return pubKey
		},
	}
	arg.PeerMessenger = &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}

	singleSigner := &mock.SinglesignStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) (i []byte, e error) {
			expectedErr = signErr
			return nil, signErr
		},
	}
	arg.PeerSignatureHandler = &mock.PeerSignatureHandler{Signer: singleSigner}

	arg.Marshalizer = &testscommon.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			expectedErr = marshalErr
			return nil, marshalErr
		},
	}

	sender, _ := process.NewSender(arg)

	err := sender.SendHeartbeat()

	assert.Equal(t, expectedErr, err)
}

func TestSender_SendHeartbeatShouldWork(t *testing.T) {
	t.Parallel()

	testTopic := "topic"
	marshaledBuff := []byte("marshalBuff")
	pubKey := &mock.PublicKeyMock{
		ToByteArrayHandler: func() (i []byte, e error) {
			return []byte("pub key"), nil
		},
	}
	signature := []byte("signature")

	broadcastCalled := false
	signCalled := false
	genPubKeyCalled := false
	marshalCalled := false

	arg := createMockArgHeartbeatSender()
	arg.Topic = testTopic
	arg.PeerMessenger = &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			if topic == testTopic && bytes.Equal(buff, marshaledBuff) {
				broadcastCalled = true
			}
		},
	}
	singleSigner := &mock.SinglesignStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) (i []byte, e error) {
			signCalled = true
			return signature, nil
		},
	}
	arg.PeerSignatureHandler = &mock.PeerSignatureHandler{Signer: singleSigner}

	arg.PrivKey = &mock.PrivateKeyStub{
		GeneratePublicHandler: func() crypto.PublicKey {
			genPubKeyCalled = true
			return pubKey
		},
	}
	arg.Marshalizer = &testscommon.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			hb, ok := obj.(*data.Heartbeat)
			if ok {
				pubkeyBytes, _ := pubKey.ToByteArray()
				if bytes.Equal(hb.Signature, signature) &&
					bytes.Equal(hb.Pubkey, pubkeyBytes) {

					marshalCalled = true
					return marshaledBuff, nil
				}
			}

			return nil, nil
		},
	}
	sender, _ := process.NewSender(arg)

	err := sender.SendHeartbeat()

	assert.Nil(t, err)
	assert.True(t, broadcastCalled)
	assert.True(t, signCalled)
	assert.True(t, genPubKeyCalled)
	assert.True(t, marshalCalled)
}

func TestSender_SendHeartbeatNotABackupNodeShouldWork(t *testing.T) {
	t.Parallel()

	testTopic := "topic"
	pkBytes := []byte("pub key")
	pubKey := &mock.PublicKeyMock{
		ToByteArrayHandler: func() (i []byte, e error) {
			return pkBytes, nil
		},
	}
	signature := []byte("signature")

	broadcastCalled := false
	signCalled := false
	genPubKeyCalled := false

	arg := createMockArgHeartbeatSender()
	arg.Marshalizer = &testscommon.MarshalizerMock{}
	arg.Topic = testTopic
	arg.PeerMessenger = &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			fmt.Println(string(buff))
			pkEncoded := base64.StdEncoding.EncodeToString(pkBytes)
			if topic == testTopic && bytes.Contains(buff, []byte(pkEncoded)) {
				broadcastCalled = true
			}
		},
	}
	singleSigner := &mock.SinglesignStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) (i []byte, e error) {
			signCalled = true
			return signature, nil
		},
	}
	arg.PeerSignatureHandler = &mock.PeerSignatureHandler{Signer: singleSigner}

	arg.PrivKey = &mock.PrivateKeyStub{
		GeneratePublicHandler: func() crypto.PublicKey {
			genPubKeyCalled = true
			return pubKey
		},
	}
	sender, _ := process.NewSender(arg)

	err := sender.SendHeartbeat()

	assert.Nil(t, err)
	assert.True(t, broadcastCalled)
	assert.True(t, signCalled)
	assert.True(t, genPubKeyCalled)
}

func TestSender_SendHeartbeatBackupNodeShouldWork(t *testing.T) {
	t.Parallel()

	testTopic := "topic"
	originalPkBytes := []byte("aaaa")
	pkBytes := []byte("bbbb")
	pubKey := &mock.PublicKeyMock{
		ToByteArrayHandler: func() (i []byte, e error) {
			return originalPkBytes, nil
		},
	}
	signature := []byte("signature")

	broadcastCalled := false
	signCalled := false
	genPubKeyCalled := false

	arg := createMockArgHeartbeatSender()
	arg.RedundancyHandler = &mock.RedundancyHandlerStub{
		IsRedundancyNodeCalled: func() bool {
			return true
		},
		IsMainMachineActiveCalled: func() bool {
			return true
		},
		ObserverPrivateKeyCalled: func() crypto.PrivateKey {
			return &mock.PrivateKeyStub{
				GeneratePublicHandler: func() crypto.PublicKey {
					return &mock.PublicKeyMock{
						ToByteArrayHandler: func() (i []byte, e error) {
							return pkBytes, nil
						},
					}
				},
			}
		},
	}
	arg.Marshalizer = &testscommon.MarshalizerMock{}
	arg.Topic = testTopic
	arg.PeerMessenger = &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			fmt.Println(string(buff))
			pkEncoded := base64.StdEncoding.EncodeToString(pkBytes)
			if topic == testTopic && bytes.Contains(buff, []byte(pkEncoded)) {
				broadcastCalled = true
			}
		},
	}
	singleSigner := &mock.SinglesignStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) (i []byte, e error) {
			signCalled = true
			return signature, nil
		},
	}
	arg.PeerSignatureHandler = &mock.PeerSignatureHandler{Signer: singleSigner}

	arg.PrivKey = &mock.PrivateKeyStub{
		GeneratePublicHandler: func() crypto.PublicKey {
			genPubKeyCalled = true
			return pubKey
		},
	}
	sender, _ := process.NewSender(arg)

	err := sender.SendHeartbeat()

	assert.Nil(t, err)
	assert.True(t, broadcastCalled)
	assert.True(t, signCalled)
	assert.True(t, genPubKeyCalled)
}

func TestSender_SendHeartbeatIsBackupNodeButMainIsNotActiveShouldWork(t *testing.T) {
	t.Parallel()

	testTopic := "topic"
	originalPkBytes := []byte("aaaa")
	pkBytes := []byte("bbbb")
	pubKey := &mock.PublicKeyMock{
		ToByteArrayHandler: func() (i []byte, e error) {
			return originalPkBytes, nil
		},
	}
	signature := []byte("signature")

	broadcastCalled := false
	signCalled := false
	genPubKeyCalled := false

	arg := createMockArgHeartbeatSender()
	arg.RedundancyHandler = &mock.RedundancyHandlerStub{
		IsRedundancyNodeCalled: func() bool {
			return true
		},
		IsMainMachineActiveCalled: func() bool {
			return false
		},
		ObserverPrivateKeyCalled: func() crypto.PrivateKey {
			return &mock.PrivateKeyStub{
				GeneratePublicHandler: func() crypto.PublicKey {
					return &mock.PublicKeyMock{
						ToByteArrayHandler: func() (i []byte, e error) {
							return pkBytes, nil
						},
					}
				},
			}
		},
	}
	arg.Marshalizer = &testscommon.MarshalizerMock{}
	arg.Topic = testTopic
	arg.PeerMessenger = &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			fmt.Println(string(buff))
			originalPkEncoded := base64.StdEncoding.EncodeToString(originalPkBytes)
			if topic == testTopic && bytes.Contains(buff, []byte(originalPkEncoded)) {
				broadcastCalled = true
			}
		},
	}
	singleSigner := &mock.SinglesignStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) (i []byte, e error) {
			signCalled = true
			return signature, nil
		},
	}
	arg.PeerSignatureHandler = &mock.PeerSignatureHandler{Signer: singleSigner}

	arg.PrivKey = &mock.PrivateKeyStub{
		GeneratePublicHandler: func() crypto.PublicKey {
			genPubKeyCalled = true
			return pubKey
		},
	}
	sender, _ := process.NewSender(arg)

	err := sender.SendHeartbeat()

	assert.Nil(t, err)
	assert.True(t, broadcastCalled)
	assert.True(t, signCalled)
	assert.True(t, genPubKeyCalled)
}

func TestSender_SendHeartbeatAfterTriggerShouldWork(t *testing.T) {
	t.Parallel()

	testTopic := "topic"
	marshaledBuff := []byte("marshalBuff")
	pubKey := &mock.PublicKeyMock{
		ToByteArrayHandler: func() (i []byte, e error) {
			return []byte("pub key"), nil
		},
	}
	signature := []byte("signature")

	broadcastCalled := false
	signCalled := false
	genPubKeyCalled := false
	marshalCalled := false

	dataPayload := []byte("payload")
	arg := createMockArgHeartbeatSender()
	arg.Topic = testTopic
	arg.PeerMessenger = &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			if topic != testTopic {
				return
			}
			if bytes.Equal(buff, marshaledBuff) {
				broadcastCalled = true
			}
		},
	}
	singleSigner := &mock.SinglesignStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) (i []byte, e error) {
			signCalled = true
			return signature, nil
		},
	}
	arg.PeerSignatureHandler = &mock.PeerSignatureHandler{Signer: singleSigner}

	arg.PrivKey = &mock.PrivateKeyStub{
		GeneratePublicHandler: func() crypto.PublicKey {
			genPubKeyCalled = true
			return pubKey
		},
	}
	arg.Marshalizer = &testscommon.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			hb, ok := obj.(*data.Heartbeat)
			if ok {
				pubkeyBytes, _ := pubKey.ToByteArray()
				if bytes.Equal(hb.Signature, signature) &&
					bytes.Equal(hb.Pubkey, pubkeyBytes) &&
					bytes.Equal(hb.Payload, dataPayload) {

					marshalCalled = true
					return marshaledBuff, nil
				}
			}

			return nil, nil
		},
	}
	arg.HardforkTrigger = &testscommon.HardforkTriggerStub{
		RecordedTriggerMessageCalled: func() (i []byte, b bool) {
			return nil, true
		},
		CreateDataCalled: func() []byte {
			return dataPayload
		},
	}
	sender, _ := process.NewSender(arg)

	err := sender.SendHeartbeat()

	assert.Nil(t, err)
	assert.True(t, broadcastCalled)
	assert.True(t, signCalled)
	assert.True(t, genPubKeyCalled)
	assert.True(t, marshalCalled)
}

func TestSender_SendHeartbeatAfterTriggerWithRecorededPayloadShouldWork(t *testing.T) {
	t.Parallel()

	testTopic := "topic"
	marshaledBuff := []byte("marshalBuff")
	pubKey := &mock.PublicKeyMock{
		ToByteArrayHandler: func() (i []byte, e error) {
			return []byte("pub key"), nil
		},
	}
	signature := []byte("signature")
	originalTriggerPayload := []byte("original trigger payload")

	broadcastCalled := false
	broadcastTriggerCalled := false
	signCalled := false
	genPubKeyCalled := false
	marshalCalled := false

	arg := createMockArgHeartbeatSender()
	arg.Topic = testTopic
	arg.PeerMessenger = &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			if topic != testTopic {
				return
			}
			if bytes.Equal(buff, marshaledBuff) {
				broadcastCalled = true
			}
			if bytes.Equal(buff, originalTriggerPayload) {
				broadcastTriggerCalled = true
			}
		},
	}
	singleSigner := &mock.SinglesignStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) (i []byte, e error) {
			signCalled = true
			return signature, nil
		},
	}
	arg.PeerSignatureHandler = &mock.PeerSignatureHandler{Signer: singleSigner}

	arg.PrivKey = &mock.PrivateKeyStub{
		GeneratePublicHandler: func() crypto.PublicKey {
			genPubKeyCalled = true
			return pubKey
		},
	}
	arg.Marshalizer = &testscommon.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			hb, ok := obj.(*data.Heartbeat)
			if ok {
				pubkeyBytes, _ := pubKey.ToByteArray()
				if bytes.Equal(hb.Signature, signature) &&
					bytes.Equal(hb.Pubkey, pubkeyBytes) {

					marshalCalled = true
					return marshaledBuff, nil
				}
			}

			return nil, nil
		},
	}
	arg.HardforkTrigger = &testscommon.HardforkTriggerStub{
		RecordedTriggerMessageCalled: func() (i []byte, b bool) {
			return originalTriggerPayload, true
		},
	}
	sender, _ := process.NewSender(arg)

	err := sender.SendHeartbeat()

	assert.Nil(t, err)
	assert.True(t, broadcastCalled)
	assert.True(t, broadcastTriggerCalled)
	assert.True(t, signCalled)
	assert.True(t, genPubKeyCalled)
	assert.True(t, marshalCalled)
}

func TestSender_SendHeartbeatShouldNotSendAfterEpoch(t *testing.T) {
	t.Parallel()

	providedEpoch := uint32(210)
	arg := createMockArgHeartbeatSender()
	arg.HeartbeatDisableEpoch = providedEpoch

	wasBroadcastCalled := false
	arg.PeerMessenger = &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			wasBroadcastCalled = true
		},
	}

	sender, _ := process.NewSender(arg)

	sender.EpochConfirmed(providedEpoch-1, 0)
	err := sender.SendHeartbeat()
	assert.Nil(t, err)
	assert.True(t, wasBroadcastCalled)

	wasBroadcastCalled = false
	sender.EpochConfirmed(providedEpoch, 0)
	err = sender.SendHeartbeat()
	assert.Nil(t, err)
	assert.False(t, wasBroadcastCalled)
}
