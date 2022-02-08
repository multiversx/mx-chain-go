package sender

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	ed25519SingleSig "github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/stretchr/testify/assert"
)

func createMockPeerAuthenticationSenderArgs() ArgPeerAuthenticationSender {
	return ArgPeerAuthenticationSender{
		Messenger:                 &mock.MessengerStub{},
		PeerSignatureHandler:      &mock.PeerSignatureHandlerStub{},
		PrivKey:                   &mock.PrivateKeyStub{},
		Marshaller:                &mock.MarshallerMock{},
		Topic:                     "topic",
		RedundancyHandler:         &mock.RedundancyHandlerStub{},
		TimeBetweenSends:          time.Second,
		TimeBetweenSendsWhenError: time.Second,
	}
}

func createMockPeerAuthenticationSenderArgsSemiIntegrationTests() ArgPeerAuthenticationSender {
	keyGen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	sk, _ := keyGen.GeneratePair()
	singleSigner := singlesig.NewBlsSigner()

	return ArgPeerAuthenticationSender{
		Messenger: &mock.MessengerStub{},
		PeerSignatureHandler: &mock.PeerSignatureHandlerStub{
			VerifyPeerSignatureCalled: func(pk []byte, pid core.PeerID, signature []byte) error {
				senderPubKey, err := keyGen.PublicKeyFromByteArray(pk)
				if err != nil {
					return err
				}
				return singleSigner.Verify(senderPubKey, pid.Bytes(), signature)
			},
			GetPeerSignatureCalled: func(privateKey crypto.PrivateKey, pid []byte) ([]byte, error) {
				return singleSigner.Sign(privateKey, pid)
			},
		},
		PrivKey:                   sk,
		Marshaller:                &marshal.GogoProtoMarshalizer{},
		Topic:                     "topic",
		RedundancyHandler:         &mock.RedundancyHandlerStub{},
		TimeBetweenSends:          time.Second,
		TimeBetweenSendsWhenError: time.Second,
	}
}

func TestNewPeerAuthenticationSender(t *testing.T) {
	t.Parallel()

	t.Run("nil peer messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		args.Messenger = nil
		sender, err := NewPeerAuthenticationSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilMessenger, err)
	})
	t.Run("nil peer signature handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		args.PeerSignatureHandler = nil
		sender, err := NewPeerAuthenticationSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilPeerSignatureHandler, err)
	})
	t.Run("nil private key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		args.PrivKey = nil
		sender, err := NewPeerAuthenticationSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilPrivateKey, err)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		args.Marshaller = nil
		sender, err := NewPeerAuthenticationSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilMarshaller, err)
	})
	t.Run("empty topic should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		args.Topic = ""
		sender, err := NewPeerAuthenticationSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrEmptySendTopic, err)
	})
	t.Run("nil redundancy handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		args.RedundancyHandler = nil
		sender, err := NewPeerAuthenticationSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilRedundancyHandler, err)
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		args.TimeBetweenSends = time.Second - time.Nanosecond
		sender, err := NewPeerAuthenticationSender(args)

		assert.Nil(t, sender)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "TimeBetweenSends"))
		assert.False(t, strings.Contains(err.Error(), "TimeBetweenSendsWhenError"))
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		args.TimeBetweenSendsWhenError = time.Second - time.Nanosecond
		sender, err := NewPeerAuthenticationSender(args)

		assert.Nil(t, sender)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "TimeBetweenSendsWhenError"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		sender, err := NewPeerAuthenticationSender(args)

		assert.NotNil(t, sender)
		assert.Nil(t, err)
	})
}

func TestPeerAuthenticationSender_execute(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("messenger Sign method fails, should return error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		args.Messenger = &mock.MessengerStub{
			SignCalled: func(payload []byte) ([]byte, error) {
				return nil, expectedErr
			},
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Fail(t, "should have not called Messenger.BroadcastCalled")
			},
		}
		sender, _ := NewPeerAuthenticationSender(args)

		err := sender.execute()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("marshaller fails in first time, should return error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		args.Messenger = &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Fail(t, "should have not called Messenger.BroadcastCalled")
			},
		}
		args.Marshaller = &mock.MarshallerStub{
			MarshalHandler: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		sender, _ := NewPeerAuthenticationSender(args)

		err := sender.execute()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("get peer signature method fails, should return error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		args.Messenger = &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Fail(t, "should have not called Messenger.BroadcastCalled")
			},
		}
		args.PeerSignatureHandler = &mock.PeerSignatureHandlerStub{
			GetPeerSignatureCalled: func(key crypto.PrivateKey, pid []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		sender, _ := NewPeerAuthenticationSender(args)

		err := sender.execute()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("marshaller fails fot the second time, should return error", func(t *testing.T) {
		t.Parallel()

		numCalls := 0
		args := createMockPeerAuthenticationSenderArgs()
		args.Messenger = &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Fail(t, "should have not called Messenger.BroadcastCalled")
			},
		}
		args.Marshaller = &mock.MarshallerStub{
			MarshalHandler: func(obj interface{}) ([]byte, error) {
				numCalls++
				if numCalls < 2 {
					return make([]byte, 0), nil
				}
				return nil, expectedErr
			},
		}
		sender, _ := NewPeerAuthenticationSender(args)

		err := sender.execute()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work with stubs", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		broadcastCalled := false
		args.Messenger = &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Equal(t, args.Topic, topic)
				broadcastCalled = true
			},
		}
		sender, _ := NewPeerAuthenticationSender(args)

		err := sender.execute()
		assert.Nil(t, err)
		assert.True(t, broadcastCalled)
	})
	t.Run("should work with some real components", func(t *testing.T) {
		t.Parallel()

		startTime := time.Now()
		// use the Elrond defined ed25519 operations instead of the secp256k1 implemented in the "real" network messenger,
		// should work with both
		keyGen := signing.NewKeyGenerator(ed25519.NewEd25519())
		skMessenger, pkMessenger := keyGen.GeneratePair()
		signerMessenger := ed25519SingleSig.Ed25519Signer{}

		args := createMockPeerAuthenticationSenderArgsSemiIntegrationTests()
		var buffResulted []byte
		messenger := &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Equal(t, args.Topic, topic)
				buffResulted = buff
			},
			SignCalled: func(payload []byte) ([]byte, error) {
				return signerMessenger.Sign(skMessenger, payload)
			},
			VerifyCalled: func(payload []byte, pid core.PeerID, signature []byte) error {
				pk, _ := keyGen.PublicKeyFromByteArray(pid.Bytes())

				return signerMessenger.Verify(pk, payload, signature)
			},
			IDCalled: func() core.PeerID {
				pkBytes, _ := pkMessenger.ToByteArray()
				return core.PeerID(pkBytes)
			},
		}
		args.Messenger = messenger
		sender, _ := NewPeerAuthenticationSender(args)

		err := sender.execute()
		assert.Nil(t, err)

		skBytes, _ := sender.privKey.ToByteArray()
		pkBytes, _ := sender.publicKey.ToByteArray()
		log.Info("args", "pid", args.Messenger.ID().Pretty(), "bls sk", skBytes, "bls pk", pkBytes)

		// verify the received bytes if they can be converted in a valid peer authentication message
		recoveredMessage := &heartbeat.PeerAuthentication{}
		err = args.Marshaller.Unmarshal(recoveredMessage, buffResulted)
		assert.Nil(t, err)
		assert.Equal(t, pkBytes, recoveredMessage.Pubkey)
		assert.Equal(t, args.Messenger.ID().Pretty(), core.PeerID(recoveredMessage.Pid).Pretty())
		// verify BLS sig on having the payload == message's pid
		err = args.PeerSignatureHandler.VerifyPeerSignature(recoveredMessage.Pubkey, core.PeerID(recoveredMessage.Pid), recoveredMessage.Signature)
		assert.Nil(t, err)
		// verify ed25519 sig having the payload == message's payload
		err = messenger.Verify(recoveredMessage.Payload, core.PeerID(recoveredMessage.Pid), recoveredMessage.PayloadSignature)
		assert.Nil(t, err)

		recoveredPayload := &heartbeat.Payload{}
		err = args.Marshaller.Unmarshal(recoveredPayload, recoveredMessage.Payload)
		assert.Nil(t, err)

		endTime := time.Now()

		messageTime := time.Unix(recoveredPayload.Timestamp, 0)
		assert.True(t, startTime.Unix() <= messageTime.Unix())
		assert.True(t, messageTime.Unix() <= endTime.Unix())
	})
}

func TestPeerAuthenticationSender_Execute(t *testing.T) {
	t.Parallel()

	t.Run("execute errors, should set the error time duration value", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockPeerAuthenticationSenderArgs()
		args.TimeBetweenSendsWhenError = time.Second * 3
		args.TimeBetweenSends = time.Second * 2
		args.PeerSignatureHandler = &mock.PeerSignatureHandlerStub{
			GetPeerSignatureCalled: func(key crypto.PrivateKey, pid []byte) ([]byte, error) {
				return nil, errors.New("error")
			},
		}
		sender, _ := NewPeerAuthenticationSender(args)
		sender.timerHandler = &mock.TimerHandlerStub{
			CreateNewTimerCalled: func(duration time.Duration) {
				assert.Equal(t, args.TimeBetweenSendsWhenError, duration)
				wasCalled = true
			},
		}

		sender.Execute()
		assert.True(t, wasCalled)
	})
	t.Run("execute worked, should set the normal time duration value", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockPeerAuthenticationSenderArgs()
		args.TimeBetweenSendsWhenError = time.Second * 3
		args.TimeBetweenSends = time.Second * 2
		sender, _ := NewPeerAuthenticationSender(args)
		sender.timerHandler = &mock.TimerHandlerStub{
			CreateNewTimerCalled: func(duration time.Duration) {
				assert.Equal(t, args.TimeBetweenSends, duration)
				wasCalled = true
			},
		}

		sender.Execute()
		assert.True(t, wasCalled)
	})
}

func TestPeerAuthenticationSender_getCurrentPrivateAndPublicKeys(t *testing.T) {
	t.Parallel()

	t.Run("is not redundancy node should return regular keys", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		args.RedundancyHandler = &mock.RedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool {
				return false
			},
		}
		sender, _ := NewPeerAuthenticationSender(args)
		sk, pk := sender.getCurrentPrivateAndPublicKeys()
		assert.True(t, sk == args.PrivKey)     // pointer testing
		assert.True(t, pk == sender.publicKey) // pointer testing
	})
	t.Run("is redundancy node but the main machine is not active should return regular keys", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs()
		args.RedundancyHandler = &mock.RedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool {
				return true
			},
			IsMainMachineActiveCalled: func() bool {
				return false
			},
		}
		sender, _ := NewPeerAuthenticationSender(args)
		sk, pk := sender.getCurrentPrivateAndPublicKeys()
		assert.True(t, sk == args.PrivKey)     // pointer testing
		assert.True(t, pk == sender.publicKey) // pointer testing
	})
	t.Run("is redundancy node but the main machine is active should return the observer keys", func(t *testing.T) {
		t.Parallel()

		observerSk := &mock.PrivateKeyStub{}
		args := createMockPeerAuthenticationSenderArgs()
		args.RedundancyHandler = &mock.RedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool {
				return true
			},
			IsMainMachineActiveCalled: func() bool {
				return true
			},
			ObserverPrivateKeyCalled: func() crypto.PrivateKey {
				return observerSk
			},
		}
		sender, _ := NewPeerAuthenticationSender(args)
		sk, pk := sender.getCurrentPrivateAndPublicKeys()
		assert.True(t, sk == args.RedundancyHandler.ObserverPrivateKey()) // pointer testing
		assert.True(t, pk == sender.observerPublicKey)                    // pointer testing
	})

}
