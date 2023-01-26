package sender

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	ed25519SingleSig "github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl/singlesig"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/heartbeat/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createShardCoordinatorInShard(shardID uint32) *testscommon.ShardsCoordinatorMock {
	shardCoordinator := testscommon.NewMultiShardsCoordinatorMock(3)
	shardCoordinator.CurrentShard = shardID

	return shardCoordinator
}

func createMockMultikeyPeerAuthenticationSenderArgs(argBase argBaseSender) argMultikeyPeerAuthenticationSender {
	return argMultikeyPeerAuthenticationSender{
		argBaseSender:            argBase,
		nodesCoordinator:         &shardingMocks.NodesCoordinatorStub{},
		peerSignatureHandler:     &cryptoMocks.PeerSignatureHandlerStub{},
		hardforkTrigger:          &testscommon.HardforkTriggerStub{},
		hardforkTimeBetweenSends: time.Second,
		hardforkTriggerPubKey:    providedHardforkPubKey,
		timeBetweenChecks:        time.Second,
		managedPeersHolder: &testscommon.ManagedPeersHolderStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				return true
			},
		},
		shardCoordinator: createShardCoordinatorInShard(0),
	}
}

func createMockMultikeyPeerAuthenticationSenderArgsSemiIntegrationTests(
	numKeys int,
) (argMultikeyPeerAuthenticationSender, *p2pmocks.MessengerStub) {
	keyGenForBLS := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	keyGenForP2P := signing.NewKeyGenerator(ed25519.NewEd25519())
	signerMessenger := ed25519SingleSig.Ed25519Signer{}

	keyMap := make(map[string]crypto.PrivateKey)
	for i := 0; i < numKeys; i++ {
		sk, pk := keyGenForBLS.GeneratePair()
		pkBytes, _ := pk.ToByteArray()

		keyMap[string(pkBytes)] = sk
	}

	p2pSkPkMap := make(map[string][]byte)
	peerIdPkMap := make(map[string]core.PeerID)
	for pk := range keyMap {
		p2pSk, p2pPk := keyGenForP2P.GeneratePair()
		p2pSkBytes, _ := p2pSk.ToByteArray()
		p2pPkBytes, _ := p2pPk.ToByteArray()

		p2pSkPkMap[pk] = p2pSkBytes
		peerIdPkMap[pk] = core.PeerID(p2pPkBytes)
	}

	mutTimeMap := sync.RWMutex{}
	peerAuthTimeMap := make(map[string]time.Time)
	managedPeersHolder := &testscommon.ManagedPeersHolderStub{
		GetP2PIdentityCalled: func(pkBytes []byte) ([]byte, core.PeerID, error) {
			return p2pSkPkMap[string(pkBytes)], peerIdPkMap[string(pkBytes)], nil
		},
		GetManagedKeysByCurrentNodeCalled: func() map[string]crypto.PrivateKey {
			return keyMap
		},
		IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
			return true
		},
		IsKeyValidatorCalled: func(pkBytes []byte) bool {
			return true
		},
		GetNextPeerAuthenticationTimeCalled: func(pkBytes []byte) (time.Time, error) {
			mutTimeMap.RLock()
			defer mutTimeMap.RUnlock()
			return peerAuthTimeMap[string(pkBytes)], nil
		},
		SetNextPeerAuthenticationTimeCalled: func(pkBytes []byte, nextTime time.Time) {
			mutTimeMap.Lock()
			defer mutTimeMap.Unlock()
			peerAuthTimeMap[string(pkBytes)] = nextTime
		},
	}

	singleSigner := singlesig.NewBlsSigner()

	baseArgs := createMockBaseArgs()
	args := argMultikeyPeerAuthenticationSender{
		argBaseSender:    baseArgs,
		nodesCoordinator: &shardingMocks.NodesCoordinatorStub{},
		peerSignatureHandler: &mock.PeerSignatureHandlerStub{
			VerifyPeerSignatureCalled: func(pk []byte, pid core.PeerID, signature []byte) error {
				senderPubKey, err := keyGenForBLS.PublicKeyFromByteArray(pk)
				if err != nil {
					return err
				}
				return singleSigner.Verify(senderPubKey, pid.Bytes(), signature)
			},
			GetPeerSignatureCalled: func(privateKey crypto.PrivateKey, pid []byte) ([]byte, error) {
				return singleSigner.Sign(privateKey, pid)
			},
		},
		hardforkTrigger:          &testscommon.HardforkTriggerStub{},
		hardforkTimeBetweenSends: time.Second,
		hardforkTriggerPubKey:    providedHardforkPubKey,
		timeBetweenChecks:        time.Second,
		managedPeersHolder:       managedPeersHolder,
		shardCoordinator:         createShardCoordinatorInShard(0),
	}

	messenger := &p2pmocks.MessengerStub{
		SignUsingPrivateKeyCalled: func(skBytes []byte, payload []byte) ([]byte, error) {
			p2pSk, _ := keyGenForP2P.PrivateKeyFromByteArray(skBytes)

			return signerMessenger.Sign(p2pSk, payload)
		},
		VerifyCalled: func(payload []byte, pid core.PeerID, signature []byte) error {
			pk, _ := keyGenForP2P.PublicKeyFromByteArray(pid.Bytes())

			return signerMessenger.Verify(pk, payload, signature)
		},
	}

	args.messenger = messenger

	return args, messenger
}

func TestNewMultikeyPeerAuthenticationSender(t *testing.T) {
	t.Parallel()

	t.Run("nil peer messenger should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		argsBase.messenger = nil

		args := createMockMultikeyPeerAuthenticationSenderArgs(argsBase)
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.Equal(t, heartbeat.ErrNilMessenger, err)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		argsBase.marshaller = nil

		args := createMockMultikeyPeerAuthenticationSenderArgs(argsBase)
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.Equal(t, heartbeat.ErrNilMarshaller, err)
	})
	t.Run("empty topic should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		argsBase.topic = ""

		args := createMockMultikeyPeerAuthenticationSenderArgs(argsBase)
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.Equal(t, heartbeat.ErrEmptySendTopic, err)
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		argsBase.timeBetweenSends = time.Second - time.Nanosecond

		args := createMockMultikeyPeerAuthenticationSenderArgs(argsBase)
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSends"))
		assert.False(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		argsBase.timeBetweenSendsWhenError = time.Second - time.Nanosecond

		args := createMockMultikeyPeerAuthenticationSenderArgs(argsBase)
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("threshold too small should error", func(t *testing.T) {
		t.Parallel()

		args := createMockMultikeyPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.thresholdBetweenSends = 0.001
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidThreshold))
		assert.True(t, strings.Contains(err.Error(), "thresholdBetweenSends"))
	})
	t.Run("threshold too big should error", func(t *testing.T) {
		t.Parallel()

		args := createMockMultikeyPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.thresholdBetweenSends = 1.001
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidThreshold))
		assert.True(t, strings.Contains(err.Error(), "thresholdBetweenSends"))
	})
	t.Run("nil nodes coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockMultikeyPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.nodesCoordinator = nil
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.Equal(t, heartbeat.ErrNilNodesCoordinator, err)
	})
	t.Run("nil peer signature handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockMultikeyPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.peerSignatureHandler = nil
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.Equal(t, heartbeat.ErrNilPeerSignatureHandler, err)
	})
	t.Run("nil hardfork trigger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockMultikeyPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.hardforkTrigger = nil
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.Equal(t, heartbeat.ErrNilHardforkTrigger, err)
	})
	t.Run("invalid time between hardforks should error", func(t *testing.T) {
		t.Parallel()

		args := createMockMultikeyPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.hardforkTimeBetweenSends = time.Second - time.Nanosecond
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "hardforkTimeBetweenSends"))
	})
	t.Run("nil managed peers holder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockMultikeyPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.managedPeersHolder = nil
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.Equal(t, heartbeat.ErrNilManagedPeersHolder, err)
	})
	t.Run("invalid time between checks should error", func(t *testing.T) {
		t.Parallel()

		args := createMockMultikeyPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.timeBetweenChecks = time.Second - time.Nanosecond
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenChecks"))
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockMultikeyPeerAuthenticationSenderArgs(createMockBaseArgs())
		args.shardCoordinator = nil
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.Equal(t, heartbeat.ErrNilShardCoordinator, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockMultikeyPeerAuthenticationSenderArgs(createMockBaseArgs())
		senderInstance, err := newMultikeyPeerAuthenticationSender(args)

		assert.False(t, check.IfNil(senderInstance))
		assert.Nil(t, err)
	})
}

func TestNewMultikeyPeerAuthenticationSender_Execute(t *testing.T) {
	t.Parallel()

	t.Run("should work for the first time with some real components", func(t *testing.T) {
		t.Parallel()

		numKeys := 3
		mutData := sync.Mutex{}
		var buffResulted [][]byte
		var pids []core.PeerID
		var skBytesBroadcast [][]byte

		args, messenger := createMockMultikeyPeerAuthenticationSenderArgsSemiIntegrationTests(numKeys)
		messenger.BroadcastUsingPrivateKeyCalled = func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			assert.Equal(t, args.topic, topic)

			mutData.Lock()
			buffResulted = append(buffResulted, buff)
			pids = append(pids, pid)
			skBytesBroadcast = append(skBytesBroadcast, skBytes)
			mutData.Unlock()
		}

		senderInstance, _ := newMultikeyPeerAuthenticationSender(args)
		senderInstance.Execute()

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, numKeys, "")
		mutData.Unlock()
	})
	t.Run("should work with some real components", func(t *testing.T) {
		t.Parallel()

		numKeys := 3
		mutData := sync.Mutex{}
		var buffResulted [][]byte
		var pids []core.PeerID
		var skBytesBroadcast [][]byte

		args, messenger := createMockMultikeyPeerAuthenticationSenderArgsSemiIntegrationTests(numKeys)
		messenger.BroadcastUsingPrivateKeyCalled = func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			assert.Equal(t, args.topic, topic)

			mutData.Lock()
			buffResulted = append(buffResulted, buff)
			pids = append(pids, pid)
			skBytesBroadcast = append(skBytesBroadcast, skBytes)
			mutData.Unlock()
		}
		args.timeBetweenSends = time.Second * 3
		args.thresholdBetweenSends = 0.20

		senderInstance, _ := newMultikeyPeerAuthenticationSender(args)
		senderInstance.Execute()

		// reset data from initialization
		mutData.Lock()
		buffResulted = make([][]byte, 0)
		pids = make([]core.PeerID, 0)
		skBytesBroadcast = make([][]byte, 0)
		mutData.Unlock()

		senderInstance.Execute() // this will not add messages

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, 0, "")
		mutData.Unlock()

		time.Sleep(time.Second * 5) // allow the resending of the messages
		senderInstance.Execute()

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, numKeys, "")
		mutData.Unlock()
	})
	t.Run("should work with some real components and one key not handled by the current node", func(t *testing.T) {
		t.Parallel()

		numKeys := 3
		mutData := sync.Mutex{}
		var buffResulted [][]byte
		var pids []core.PeerID
		var skBytesBroadcast [][]byte

		args, messenger := createMockMultikeyPeerAuthenticationSenderArgsSemiIntegrationTests(numKeys)
		messenger.BroadcastUsingPrivateKeyCalled = func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			assert.Equal(t, args.topic, topic)

			mutData.Lock()
			buffResulted = append(buffResulted, buff)
			pids = append(pids, pid)
			skBytesBroadcast = append(skBytesBroadcast, skBytes)
			mutData.Unlock()
		}
		args.timeBetweenSends = time.Second * 3
		args.thresholdBetweenSends = 0.20
		firstKeyFound := ""
		pkSkMap := args.managedPeersHolder.GetManagedKeysByCurrentNode()
		for key := range pkSkMap {
			firstKeyFound = key
			break
		}
		stub := args.managedPeersHolder.(*testscommon.ManagedPeersHolderStub)
		stub.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return firstKeyFound != string(pkBytes)
		}

		senderInstance, _ := newMultikeyPeerAuthenticationSender(args)
		senderInstance.Execute()

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, numKeys-1, "")
		mutData.Unlock()

		// reset data from initialization
		mutData.Lock()
		buffResulted = make([][]byte, 0)
		pids = make([]core.PeerID, 0)
		skBytesBroadcast = make([][]byte, 0)
		mutData.Unlock()

		senderInstance.Execute() // this will not add messages

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, 0, "")
		mutData.Unlock()

		time.Sleep(time.Second * 5) // allow the resending of the messages
		senderInstance.Execute()

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, numKeys-1, "")
		mutData.Unlock()
	})
	t.Run("should work with some real components one key not a validator", func(t *testing.T) {
		t.Parallel()

		numKeys := 3
		mutData := sync.Mutex{}
		var buffResulted [][]byte
		var pids []core.PeerID
		var skBytesBroadcast [][]byte

		args, messenger := createMockMultikeyPeerAuthenticationSenderArgsSemiIntegrationTests(numKeys)
		messenger.BroadcastUsingPrivateKeyCalled = func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			assert.Equal(t, args.topic, topic)

			mutData.Lock()
			buffResulted = append(buffResulted, buff)
			pids = append(pids, pid)
			skBytesBroadcast = append(skBytesBroadcast, skBytes)
			mutData.Unlock()
		}
		args.timeBetweenSends = time.Second * 3
		args.thresholdBetweenSends = 0.20
		firstKeyFound := ""
		pkSkMap := args.managedPeersHolder.GetManagedKeysByCurrentNode()
		for key := range pkSkMap {
			firstKeyFound = key
			break
		}
		args.nodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
				if firstKeyFound == string(publicKey) {
					return nil, 0, errors.New("not a validator")
				}

				return nil, 0, nil
			},
		}
		senderInstance, _ := newMultikeyPeerAuthenticationSender(args)
		senderInstance.Execute()

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, numKeys-1, "")
		mutData.Unlock()

		// reset data from initialization
		mutData.Lock()
		buffResulted = make([][]byte, 0)
		pids = make([]core.PeerID, 0)
		skBytesBroadcast = make([][]byte, 0)
		mutData.Unlock()

		senderInstance.Execute() // this will not add messages

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, 0, "")
		mutData.Unlock()

		time.Sleep(time.Second * 5) // allow the resending of the messages
		senderInstance.Execute()

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, numKeys-1, "")
		mutData.Unlock()
	})
	t.Run("should work with some real components one key is on a different shard", func(t *testing.T) {
		t.Parallel()

		numKeys := 3
		mutData := sync.Mutex{}
		var buffResulted [][]byte
		var pids []core.PeerID
		var skBytesBroadcast [][]byte

		args, messenger := createMockMultikeyPeerAuthenticationSenderArgsSemiIntegrationTests(numKeys)
		messenger.BroadcastUsingPrivateKeyCalled = func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			assert.Equal(t, args.topic, topic)

			mutData.Lock()
			buffResulted = append(buffResulted, buff)
			pids = append(pids, pid)
			skBytesBroadcast = append(skBytesBroadcast, skBytes)
			mutData.Unlock()
		}
		args.timeBetweenSends = time.Second * 3
		args.thresholdBetweenSends = 0.20
		firstKeyFound := ""
		pkSkMap := args.managedPeersHolder.GetManagedKeysByCurrentNode()
		for key := range pkSkMap {
			firstKeyFound = key
			break
		}
		args.nodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
				if firstKeyFound == string(publicKey) {
					return nil, 1, nil
				}

				return nil, 0, nil
			},
		}
		senderInstance, _ := newMultikeyPeerAuthenticationSender(args)
		senderInstance.Execute()

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, numKeys-1, "")
		mutData.Unlock()

		// reset data from initialization
		mutData.Lock()
		buffResulted = make([][]byte, 0)
		pids = make([]core.PeerID, 0)
		skBytesBroadcast = make([][]byte, 0)
		mutData.Unlock()

		senderInstance.Execute() // this will not add messages

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, 0, "")
		mutData.Unlock()

		time.Sleep(time.Second * 5) // allow the resending of the messages
		senderInstance.Execute()

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, numKeys-1, "")
		mutData.Unlock()
	})
	t.Run("should work with some real components and hardfork trigger", func(t *testing.T) {
		t.Parallel()

		numKeys := 3
		mutData := sync.Mutex{}
		var buffResulted [][]byte
		var pids []core.PeerID
		var skBytesBroadcast [][]byte

		args, messenger := createMockMultikeyPeerAuthenticationSenderArgsSemiIntegrationTests(numKeys)
		messenger.BroadcastUsingPrivateKeyCalled = func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			assert.Equal(t, args.topic, topic)

			mutData.Lock()
			buffResulted = append(buffResulted, buff)
			pids = append(pids, pid)
			skBytesBroadcast = append(skBytesBroadcast, skBytes)
			mutData.Unlock()
		}
		args.timeBetweenSends = time.Second * 3
		args.thresholdBetweenSends = 0.20
		hardforkTriggerPayload := []byte("hardfork payload")
		args.hardforkTrigger = &testscommon.HardforkTriggerStub{
			RecordedTriggerMessageCalled: func() ([]byte, bool) {
				return make([]byte, 0), true
			},
			CreateDataCalled: func() []byte {
				return hardforkTriggerPayload
			},
		}

		senderInstance, _ := newMultikeyPeerAuthenticationSender(args)
		senderInstance.Execute()

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, numKeys, string(hardforkTriggerPayload))
		mutData.Unlock()

		// reset data from initialization
		mutData.Lock()
		buffResulted = make([][]byte, 0)
		pids = make([]core.PeerID, 0)
		skBytesBroadcast = make([][]byte, 0)
		mutData.Unlock()

		time.Sleep(time.Second * 2)
		senderInstance.Execute() // this will add messages because we are in hardfork mode

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, numKeys, string(hardforkTriggerPayload))
		mutData.Unlock()

		// reset data
		mutData.Lock()
		buffResulted = make([][]byte, 0)
		pids = make([]core.PeerID, 0)
		skBytesBroadcast = make([][]byte, 0)
		mutData.Unlock()

		time.Sleep(time.Second * 5) // allow the resending of the messages
		senderInstance.Execute()

		mutData.Lock()
		testRecoveredMessages(t, args, buffResulted, pids, skBytesBroadcast, numKeys, string(hardforkTriggerPayload))
		mutData.Unlock()
	})
}

func testRecoveredMessages(
	tb testing.TB,
	args argMultikeyPeerAuthenticationSender,
	payloads [][]byte,
	pids []core.PeerID,
	skBytesBroadcast [][]byte,
	numExpected int,

	hardforkPayload string,
) {
	require.Equal(tb, numExpected, len(payloads))
	require.Equal(tb, numExpected, len(pids))
	require.Equal(tb, numExpected, len(skBytesBroadcast))

	for i := 0; i < len(payloads); i++ {
		payload := payloads[i]
		pid := pids[i]

		testSingleMessage(tb, args, payload, pid, hardforkPayload)
	}
}

func testSingleMessage(
	tb testing.TB,
	args argMultikeyPeerAuthenticationSender,
	payload []byte,
	pid core.PeerID,
	hardforkPayload string,
) {
	recoveredBatch := batch.Batch{}
	err := args.marshaller.Unmarshal(&recoveredBatch, payload)
	assert.Nil(tb, err)

	recoveredMessage := &heartbeat.PeerAuthentication{}
	err = args.marshaller.Unmarshal(recoveredMessage, recoveredBatch.Data[0])
	assert.Nil(tb, err)

	_, correspondingPid, err := args.managedPeersHolder.GetP2PIdentity(recoveredMessage.Pubkey)
	assert.Nil(tb, err)
	assert.Equal(tb, correspondingPid.Pretty(), core.PeerID(recoveredMessage.Pid).Pretty())
	assert.Equal(tb, correspondingPid, pid)

	errVerify := args.peerSignatureHandler.VerifyPeerSignature(recoveredMessage.Pubkey, core.PeerID(recoveredMessage.Pid), recoveredMessage.Signature)
	assert.Nil(tb, errVerify)

	messenger := args.messenger.(*p2pmocks.MessengerStub)
	errVerify = messenger.Verify(recoveredMessage.Payload, core.PeerID(recoveredMessage.Pid), recoveredMessage.PayloadSignature)
	assert.Nil(tb, errVerify)

	recoveredPayload := &heartbeat.Payload{}
	err = args.marshaller.Unmarshal(recoveredPayload, recoveredMessage.Payload)
	assert.Nil(tb, err)

	endTime := time.Now()

	messageTime := time.Unix(recoveredPayload.Timestamp, 0)
	assert.True(tb, messageTime.Unix() <= endTime.Unix())
	assert.Equal(tb, hardforkPayload, recoveredPayload.HardforkMessage)
}
