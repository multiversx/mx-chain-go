package sender

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
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
		keysHolder:               &testscommon.KeysHolderStub{},
		timeBetweenChecks:        time.Second,
		shardCoordinator:         createShardCoordinatorInShard(0),
	}
}

func TestPeerAuthenticationSenderFactory_Create(t *testing.T) {
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
		assert.Equal(t, expectedErr, err)
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
		args.keysHolder = &testscommon.KeysHolderStub{
			GetManagedKeysByCurrentNodeCalled: func() map[string]crypto.PrivateKey {
				keygen := signing.NewKeyGenerator(&mcl.SuiteBLS12{})
				sk, pk := keygen.GeneratePair()
				pkBytes, err := pk.ToByteArray()
				assert.Nil(t, err)
				keysMap := make(map[string]crypto.PrivateKey)
				keysMap[string(pkBytes)] = sk
				return keysMap
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
		args.keysHolder = &testscommon.KeysHolderStub{
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
		args.keysHolder = &testscommon.KeysHolderStub{
			GetManagedKeysByCurrentNodeCalled: func() map[string]crypto.PrivateKey {
				keygen := signing.NewKeyGenerator(&mcl.SuiteBLS12{})
				sk1, pk1 := keygen.GeneratePair()
				pk1Bytes, err := pk1.ToByteArray()
				assert.Nil(t, err)
				sk2, pk2 := keygen.GeneratePair()
				pk2Bytes, err := pk2.ToByteArray()
				assert.Nil(t, err)
				keysMap := make(map[string]crypto.PrivateKey)
				keysMap[string(pk1Bytes)] = sk1
				keysMap[string(pk2Bytes)] = sk2
				return keysMap
			},
		}
		peerAuthSender, err := createPeerAuthenticationSender(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(peerAuthSender))
		assert.Equal(t, "*sender.multikeyPeerAuthenticationSender", fmt.Sprintf("%T", peerAuthSender))
	})

}
