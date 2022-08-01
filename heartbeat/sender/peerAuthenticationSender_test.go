package sender

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	ed25519SingleSig "github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
)

var providedHardforkPubKey = []byte("provided hardfork pub key")

func createMockPeerAuthenticationSenderArgs(argBase argBaseSender) argPeerAuthenticationSender {
	return argPeerAuthenticationSender{
		argBaseSender:            argBase,
		nodesCoordinator:         &shardingMocks.NodesCoordinatorStub{},
		peerSignatureHandler:     &cryptoMocks.PeerSignatureHandlerStub{},
		privKey:                  &cryptoMocks.PrivateKeyStub{},
		redundancyHandler:        &mock.RedundancyHandlerStub{},
		hardforkTrigger:          &testscommon.HardforkTriggerStub{},
		hardforkTimeBetweenSends: time.Second,
		hardforkTriggerPubKey:    providedHardforkPubKey,
	}
}

func createMockPeerAuthenticationSenderArgsSemiIntegrationTests(baseArg argBaseSender) argPeerAuthenticationSender {
	keyGen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	sk, _ := keyGen.GeneratePair()
	singleSigner := singlesig.NewBlsSigner()

	return argPeerAuthenticationSender{
		argBaseSender:    baseArg,
		nodesCoordinator: &shardingMocks.NodesCoordinatorStub{},
		peerSignatureHandler: &mock.PeerSignatureHandlerStub{
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
		privKey:                  sk,
		redundancyHandler:        &mock.RedundancyHandlerStub{},
		hardforkTrigger:          &testscommon.HardforkTriggerStub{},
		hardforkTimeBetweenSends: time.Second,
		hardforkTriggerPubKey:    providedHardforkPubKey,
	}
}

func TestNewPeerAuthenticationSender(t *testing.T) {
	t.Parallel()

	t.Run("nil peer messenger should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		argsBase.messenger = nil

		args := createMockPeerAuthenticationSenderArgs(argsBase)
		sender, err := newPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(sender))
		assert.Equal(t, heartbeat.ErrNilMessenger, err)
	})
	t.Run("nil nodes coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.nodesCoordinator = nil
		sender, err := newPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(sender))
		assert.Equal(t, heartbeat.ErrNilNodesCoordinator, err)
	})
	t.Run("nil peer signature handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.peerSignatureHandler = nil
		sender, err := newPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(sender))
		assert.Equal(t, heartbeat.ErrNilPeerSignatureHandler, err)
	})
	t.Run("nil private key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.privKey = nil
		sender, err := newPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(sender))
		assert.Equal(t, heartbeat.ErrNilPrivateKey, err)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		argsBase.marshaller = nil

		args := createMockPeerAuthenticationSenderArgs(argsBase)
		sender, err := newPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(sender))
		assert.Equal(t, heartbeat.ErrNilMarshaller, err)
	})
	t.Run("empty topic should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		argsBase.topic = ""

		args := createMockPeerAuthenticationSenderArgs(argsBase)
		sender, err := newPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(sender))
		assert.Equal(t, heartbeat.ErrEmptySendTopic, err)
	})
	t.Run("nil redundancy handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.redundancyHandler = nil
		sender, err := newPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(sender))
		assert.Equal(t, heartbeat.ErrNilRedundancyHandler, err)
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		argsBase.timeBetweenSends = time.Second - time.Nanosecond

		args := createMockPeerAuthenticationSenderArgs(argsBase)
		sender, err := newPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(sender))
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSends"))
		assert.False(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		argsBase.timeBetweenSendsWhenError = time.Second - time.Nanosecond

		args := createMockPeerAuthenticationSenderArgs(argsBase)
		sender, err := newPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(sender))
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("threshold too small should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.thresholdBetweenSends = 0.001
		sender, err := newPeerAuthenticationSender(args)

		assert.Nil(t, sender)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidThreshold))
		assert.True(t, strings.Contains(err.Error(), "thresholdBetweenSends"))
	})
	t.Run("threshold too big should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.thresholdBetweenSends = 1.001
		sender, err := newPeerAuthenticationSender(args)

		assert.Nil(t, sender)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidThreshold))
		assert.True(t, strings.Contains(err.Error(), "thresholdBetweenSends"))
	})
	t.Run("nil hardfork trigger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.hardforkTrigger = nil
		sender, err := newPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(sender))
		assert.Equal(t, heartbeat.ErrNilHardforkTrigger, err)
	})
	t.Run("invalid time between hardforks should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.hardforkTimeBetweenSends = time.Second - time.Nanosecond
		sender, err := newPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(sender))
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "hardforkTimeBetweenSends"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		sender, err := newPeerAuthenticationSender(args)

		assert.False(t, check.IfNil(sender))
		assert.Nil(t, err)
	})
}

func TestPeerAuthenticationSender_execute(t *testing.T) {
	t.Parallel()

	t.Run("messenger Sign method fails, should return error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		argsBase.messenger = &mock.MessengerStub{
			SignCalled: func(payload []byte) ([]byte, error) {
				return nil, expectedErr
			},
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Fail(t, "should have not called Messenger.BroadcastCalled")
			},
		}

		args := createMockPeerAuthenticationSenderArgs(argsBase)
		sender, _ := newPeerAuthenticationSender(args)

		err, isHardforkTriggered := sender.execute()
		assert.Equal(t, expectedErr, err)
		assert.False(t, isHardforkTriggered)
	})
	t.Run("marshaller fails in first time, should return error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		argsBase.messenger = &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Fail(t, "should have not called Messenger.BroadcastCalled")
			},
		}
		argsBase.marshaller = &testscommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}

		args := createMockPeerAuthenticationSenderArgs(argsBase)
		sender, _ := newPeerAuthenticationSender(args)

		err, isHardforkTriggered := sender.execute()
		assert.Equal(t, expectedErr, err)
		assert.False(t, isHardforkTriggered)
	})
	t.Run("get peer signature method fails, should return error", func(t *testing.T) {
		t.Parallel()

		baseArgs := createMockBaseArgs()
		baseArgs.messenger = &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Fail(t, "should have not called Messenger.BroadcastCalled")
			},
		}
		args := createMockPeerAuthenticationSenderArgs(baseArgs)
		args.peerSignatureHandler = &mock.PeerSignatureHandlerStub{
			GetPeerSignatureCalled: func(key crypto.PrivateKey, pid []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		sender, _ := newPeerAuthenticationSender(args)

		err, isHardforkTriggered := sender.execute()
		assert.Equal(t, expectedErr, err)
		assert.False(t, isHardforkTriggered)
	})
	t.Run("marshaller fails for the second time, should return error", func(t *testing.T) {
		t.Parallel()

		numCalls := 0
		argsBase := createMockBaseArgs()
		argsBase.messenger = &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Fail(t, "should have not called Messenger.BroadcastCalled")
			},
		}
		argsBase.marshaller = &testscommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				numCalls++
				if numCalls < 2 {
					return make([]byte, 0), nil
				}
				return nil, expectedErr
			},
		}

		args := createMockPeerAuthenticationSenderArgs(argsBase)
		sender, _ := newPeerAuthenticationSender(args)

		err, isHardforkTriggered := sender.execute()
		assert.Equal(t, expectedErr, err)
		assert.False(t, isHardforkTriggered)
	})
	t.Run("should work with stubs", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		broadcastCalled := false
		argsBase.messenger = &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Equal(t, argsBase.topic, topic)
				broadcastCalled = true
			},
		}

		args := createMockPeerAuthenticationSenderArgs(argsBase)
		sender, _ := newPeerAuthenticationSender(args)

		err, isHardforkTriggered := sender.execute()
		assert.Nil(t, err)
		assert.True(t, broadcastCalled)
		assert.False(t, isHardforkTriggered)
	})
	t.Run("should work with some real components", func(t *testing.T) {
		t.Parallel()

		startTime := time.Now()
		// use the Elrond defined ed25519 operations instead of the secp256k1 implemented in the "real" network messenger,
		// should work with both
		keyGen := signing.NewKeyGenerator(ed25519.NewEd25519())
		skMessenger, pkMessenger := keyGen.GeneratePair()
		signerMessenger := ed25519SingleSig.Ed25519Signer{}

		argsBase := createMockBaseArgs()
		var buffResulted []byte
		messenger := &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Equal(t, argsBase.topic, topic)
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
		argsBase.messenger = messenger
		args := createMockPeerAuthenticationSenderArgsSemiIntegrationTests(argsBase)
		sender, _ := newPeerAuthenticationSender(args)

		err, isHardforkTriggered := sender.execute()
		assert.Nil(t, err)
		assert.False(t, isHardforkTriggered)

		skBytes, _ := sender.privKey.ToByteArray()
		pkBytes, _ := sender.publicKey.ToByteArray()
		log.Info("args", "pid", argsBase.messenger.ID().Pretty(), "bls sk", skBytes, "bls pk", pkBytes)

		// verify the received bytes if they can be converted in a valid peer authentication message
		recoveredBatch := batch.Batch{}
		err = argsBase.marshaller.Unmarshal(&recoveredBatch, buffResulted)
		assert.Nil(t, err)
		recoveredMessage := &heartbeat.PeerAuthentication{}
		err = argsBase.marshaller.Unmarshal(recoveredMessage, recoveredBatch.Data[0])
		assert.Nil(t, err)
		assert.Equal(t, pkBytes, recoveredMessage.Pubkey)
		assert.Equal(t, argsBase.messenger.ID().Pretty(), core.PeerID(recoveredMessage.Pid).Pretty())
		t.Run("verify BLS sig on having the payload == message's pid", func(t *testing.T) {
			errVerify := args.peerSignatureHandler.VerifyPeerSignature(recoveredMessage.Pubkey, core.PeerID(recoveredMessage.Pid), recoveredMessage.Signature)
			assert.Nil(t, errVerify)
		})
		t.Run("verify ed25519 sig having the payload == message's payload", func(t *testing.T) {
			errVerify := messenger.Verify(recoveredMessage.Payload, core.PeerID(recoveredMessage.Pid), recoveredMessage.PayloadSignature)
			assert.Nil(t, errVerify)
		})
		t.Run("verify payload", func(t *testing.T) {
			recoveredPayload := &heartbeat.Payload{}
			err = argsBase.marshaller.Unmarshal(recoveredPayload, recoveredMessage.Payload)
			assert.Nil(t, err)

			endTime := time.Now()

			messageTime := time.Unix(recoveredPayload.Timestamp, 0)
			assert.True(t, startTime.Unix() <= messageTime.Unix())
			assert.True(t, messageTime.Unix() <= endTime.Unix())
		})
	})
}

func TestPeerAuthenticationSender_Execute(t *testing.T) {
	t.Parallel()

	t.Run("observer should not execute", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		wasBroadcastCalled := false
		argsBase.messenger = &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				wasBroadcastCalled = true
			},
		}
		args := createMockPeerAuthenticationSenderArgs(argsBase)
		args.nodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
				return nil, 0, errors.New("observer")
			},
		}
		sender, _ := newPeerAuthenticationSender(args)

		sender.Execute()
		assert.False(t, wasBroadcastCalled)
	})
	t.Run("execute errors, should set the error time duration value", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		argsBase := createMockBaseArgs()
		argsBase.timeBetweenSendsWhenError = time.Second * 3
		argsBase.timeBetweenSends = time.Second * 2

		args := createMockPeerAuthenticationSenderArgs(argsBase)
		args.peerSignatureHandler = &mock.PeerSignatureHandlerStub{
			GetPeerSignatureCalled: func(key crypto.PrivateKey, pid []byte) ([]byte, error) {
				return nil, errors.New("error")
			},
		}

		sender, _ := newPeerAuthenticationSender(args)
		sender.timerHandler = &mock.TimerHandlerStub{
			CreateNewTimerCalled: func(duration time.Duration) {
				assert.Equal(t, argsBase.timeBetweenSendsWhenError, duration)
				wasCalled = true
			},
		}

		sender.Execute()
		assert.True(t, wasCalled)
	})
	t.Run("execute worked, should set the normal time duration value", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		argsBase := createMockBaseArgs()
		argsBase.timeBetweenSendsWhenError = time.Second * 3
		argsBase.timeBetweenSends = time.Second * 2
		args := createMockPeerAuthenticationSenderArgs(argsBase)

		sender, _ := newPeerAuthenticationSender(args)
		sender.timerHandler = &mock.TimerHandlerStub{
			CreateNewTimerCalled: func(duration time.Duration) {
				floatTBS := float64(argsBase.timeBetweenSends.Nanoseconds())
				maxDuration := floatTBS + floatTBS*argsBase.thresholdBetweenSends
				assert.True(t, time.Duration(maxDuration) > duration)
				assert.True(t, argsBase.timeBetweenSends <= duration)
				wasCalled = true
			},
		}

		sender.Execute()
		assert.True(t, wasCalled)
	})
	t.Run("observer->validator->observer should work", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		counterBroadcast := 0
		argsBase.messenger = &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				counterBroadcast++
			},
		}
		args := createMockPeerAuthenticationSenderArgs(argsBase)
		counter := 0
		args.nodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
				counter++
				if counter == 2 {
					return nil, 0, nil // validator
				}

				return nil, 0, errors.New("observer") // observer
			},
		}

		sender, _ := newPeerAuthenticationSender(args)

		sender.Execute() // observer
		sender.Execute() // validator
		sender.Execute() // observer
		assert.Equal(t, 1, counterBroadcast)
	})
	t.Run("execute worked, should set the hardfork time duration value", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		argsBase := createMockBaseArgs()
		args := createMockPeerAuthenticationSenderArgs(argsBase)
		args.hardforkTimeBetweenSends = time.Second * 3
		args.hardforkTrigger = &testscommon.HardforkTriggerStub{
			RecordedTriggerMessageCalled: func() ([]byte, bool) {
				return make([]byte, 0), true
			},
		}
		sender, _ := newPeerAuthenticationSender(args)
		sender.timerHandler = &mock.TimerHandlerStub{
			CreateNewTimerCalled: func(duration time.Duration) {
				floatTBH := float64(args.hardforkTimeBetweenSends.Nanoseconds())
				maxDuration := floatTBH + floatTBH*argsBase.thresholdBetweenSends
				assert.True(t, time.Duration(maxDuration) > duration)
				assert.True(t, args.hardforkTimeBetweenSends <= duration)
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

		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.redundancyHandler = &mock.RedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool {
				return false
			},
		}
		sender, _ := newPeerAuthenticationSender(args)
		sk, pk := sender.getCurrentPrivateAndPublicKeys()
		assert.True(t, sk == args.privKey)     // pointer testing
		assert.True(t, pk == sender.publicKey) // pointer testing
	})
	t.Run("is redundancy node but the main machine is not active should return regular keys", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.redundancyHandler = &mock.RedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool {
				return true
			},
			IsMainMachineActiveCalled: func() bool {
				return false
			},
		}
		sender, _ := newPeerAuthenticationSender(args)
		sk, pk := sender.getCurrentPrivateAndPublicKeys()
		assert.True(t, sk == args.privKey)     // pointer testing
		assert.True(t, pk == sender.publicKey) // pointer testing
	})
	t.Run("is redundancy node but the main machine is active should return the observer keys", func(t *testing.T) {
		t.Parallel()

		observerSk := &mock.PrivateKeyStub{}
		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.redundancyHandler = &mock.RedundancyHandlerStub{
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
		sender, _ := newPeerAuthenticationSender(args)
		sk, pk := sender.getCurrentPrivateAndPublicKeys()
		assert.True(t, sk == args.redundancyHandler.ObserverPrivateKey()) // pointer testing
		assert.True(t, pk == sender.observerPublicKey)                    // pointer testing
	})
	t.Run("call from multiple threads", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.redundancyHandler = &mock.RedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool {
				return false
			},
		}
		sender, _ := newPeerAuthenticationSender(args)

		numOfThreads := 10
		var wg sync.WaitGroup
		wg.Add(numOfThreads)
		for i := 0; i < numOfThreads; i++ {
			go func() {
				defer wg.Done()
				sk, pk := sender.getCurrentPrivateAndPublicKeys()
				assert.True(t, sk == args.privKey)     // pointer testing
				assert.True(t, pk == sender.publicKey) // pointer testing
			}()
		}

		wg.Wait()
	})
}

func TestPeerAuthenticationSender_getHardforkPayload(t *testing.T) {
	t.Parallel()

	t.Run("hardfork not triggered should work", func(t *testing.T) {
		t.Parallel()

		providedPayload := make([]byte, 0)
		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.hardforkTrigger = &testscommon.HardforkTriggerStub{
			RecordedTriggerMessageCalled: func() ([]byte, bool) {
				return nil, false
			},
		}

		sender, _ := newPeerAuthenticationSender(args)

		payload, isTriggered := sender.getHardforkPayload()
		assert.False(t, isTriggered)
		assert.Equal(t, providedPayload, payload)
	})
	t.Run("hardfork triggered should work", func(t *testing.T) {
		t.Parallel()

		providedPayload := []byte("provided payload")
		args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.hardforkTrigger = &testscommon.HardforkTriggerStub{
			RecordedTriggerMessageCalled: func() ([]byte, bool) {
				return nil, true
			},
			CreateDataCalled: func() []byte {
				return providedPayload
			},
		}

		sender, _ := newPeerAuthenticationSender(args)

		payload, isTriggered := sender.getHardforkPayload()
		assert.True(t, isTriggered)
		assert.Equal(t, providedPayload, payload)
	})
}

func TestPeerAuthenticationSender_ShouldTriggerHardfork(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	ch := make(chan struct{})
	args := createMockPeerAuthenticationSenderArgs(createMockBaseArgs())
	args.hardforkTrigger = &testscommon.HardforkTriggerStub{
		NotifyTriggerReceivedV2Called: func() <-chan struct{} {
			return ch
		},
	}

	go func() {
		time.Sleep(time.Second)
		ch <- struct{}{}
	}()

	sender, _ := newPeerAuthenticationSender(args)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	select {
	case <-sender.ShouldTriggerHardfork():
		return
	case <-ctx.Done():
		assert.Fail(t, "should not reach timeout")
	}
}
