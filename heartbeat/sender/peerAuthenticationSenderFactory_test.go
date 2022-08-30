package sender

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
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
		privKey:                  &cryptoMocks.PrivateKeyStub{},
		redundancyHandler:        &mock.RedundancyHandlerStub{},
	}
}

func Test_peerAuthenticationSenderFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("both configs should error", func(t *testing.T) {
		t.Parallel()

		peerAuthFactory, err := newPeerAuthenticationSenderFactory(createMockPeerAuthenticationSenderFactoryArgs())
		assert.Nil(t, err)
		assert.NotNil(t, peerAuthFactory)

		peerAuth, err := peerAuthFactory.create()
		assert.True(t, check.IfNil(peerAuth))
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidConfiguration))
	})
	t.Run("none of the configs should error", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderFactoryArgs()
		args.keysHolder = nil
		args.privKey = nil
		peerAuthFactory, err := newPeerAuthenticationSenderFactory(args)
		assert.Nil(t, err)
		assert.NotNil(t, peerAuthFactory)

		peerAuth, err := peerAuthFactory.create()
		assert.True(t, check.IfNil(peerAuth))
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidConfiguration))
	})
	t.Run("should create validator peer authentication", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderFactoryArgs()
		args.keysHolder = nil
		peerAuthFactory, err := newPeerAuthenticationSenderFactory(args)
		assert.Nil(t, err)
		assert.NotNil(t, peerAuthFactory)

		peerAuth, err := peerAuthFactory.create()
		assert.Nil(t, err)
		assert.False(t, check.IfNil(peerAuth))

		_, ok := peerAuth.(*peerAuthenticationSender)
		assert.True(t, ok)
	})
	t.Run("should create multikey peer authentication", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerAuthenticationSenderFactoryArgs()
		args.privKey = nil
		peerAuthFactory, err := newPeerAuthenticationSenderFactory(args)
		assert.Nil(t, err)
		assert.NotNil(t, peerAuthFactory)

		peerAuth, err := peerAuthFactory.create()
		assert.Nil(t, err)
		assert.False(t, check.IfNil(peerAuth))

		_, ok := peerAuth.(*multikeyPeerAuthenticationSender)
		assert.True(t, ok)
	})
}
