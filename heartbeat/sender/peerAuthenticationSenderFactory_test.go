package sender

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
)

func createMockPeerAuthenticationSenderFactoryArgs() argPeerAuthenticationSenderFactory {
	return argPeerAuthenticationSenderFactory{
		argBaseSender:            createMockBaseArgs(),
		nodesCoordinator:         &shardingMocks.NodesCoordinatorStub{},
		peerSignatureHandler:     &cryptoMocks.PeerSignatureHandlerStub{},
		hardforkTrigger:          &testscommon.HardforkTriggerStub{},
		hardforkTimeBetweenSends: time.Second,
		hardforkTriggerPubKey:    providedHardforkPubKey,
		managedPeersHolder:       &testscommon.ManagedPeersHolderStub{},
		timeBetweenChecks:        time.Second,
		shardCoordinator:         createShardCoordinatorInShard(0),
	}
}

func TestPeerAuthenticationSenderFactory_createPeerAuthenticationSender(t *testing.T) {
	t.Parallel()

	t.Run("ToByteArray fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderFactoryArgs()
		args.privKey = &cryptoMocks.PrivateKeyStub{
			GeneratePublicStub: func() crypto.PublicKey {
				return &cryptoMocks.PublicKeyStub{
					ToByteArrayStub: func() ([]byte, error) {
						return nil, expectedErr
					},
				}
			},
		}
		peerAuthSender, err := createPeerAuthenticationSender(args)
		assert.True(t, errors.Is(err, expectedErr))
		assert.True(t, check.IfNil(peerAuthSender))
	})
	t.Run("validator with keys managed should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderFactoryArgs()
		args.nodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
				return nil, 0, nil
			},
		}
		args.managedPeersHolder = &testscommon.ManagedPeersHolderStub{
			IsMultiKeyModeCalled: func() bool {
				return true
			},
		}
		peerAuthSender, err := createPeerAuthenticationSender(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidConfiguration))
		assert.True(t, check.IfNil(peerAuthSender))
	})
	t.Run("validator should create regular sender", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderFactoryArgs()
		args.nodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
				return nil, 0, nil
			},
		}
		peerAuthSender, err := createPeerAuthenticationSender(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(peerAuthSender))
		assert.Equal(t, "*sender.peerAuthenticationSender", fmt.Sprintf("%T", peerAuthSender))
	})
	t.Run("regular observer should create regular sender", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderFactoryArgs()
		args.nodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
				return nil, 0, errors.New("not validator")
			},
		}
		args.managedPeersHolder = &testscommon.ManagedPeersHolderStub{
			GetManagedKeysByCurrentNodeCalled: func() map[string]crypto.PrivateKey {
				return make(map[string]crypto.PrivateKey)
			},
		}
		peerAuthSender, err := createPeerAuthenticationSender(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(peerAuthSender))
		assert.Equal(t, "*sender.peerAuthenticationSender", fmt.Sprintf("%T", peerAuthSender))
	})
	t.Run("not validator with keys managed should create multikey sender", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderFactoryArgs()
		args.nodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
				return nil, 0, errors.New("not validator")
			},
		}
		args.managedPeersHolder = &testscommon.ManagedPeersHolderStub{
			IsMultiKeyModeCalled: func() bool {
				return true
			},
		}
		peerAuthSender, err := createPeerAuthenticationSender(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(peerAuthSender))
		assert.Equal(t, "*sender.multikeyPeerAuthenticationSender", fmt.Sprintf("%T", peerAuthSender))
	})
}
