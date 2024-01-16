package keysManagement_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/keysManagement"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/stretchr/testify/assert"
)

var testPrivateKeyBytes = []byte("private key bytes")
var testPublicKeyBytes = []byte("public key bytes")
var randomPublicKeyBytes = []byte("random key bytes")

func createMockArgsKeysHandler() keysManagement.ArgsKeysHandler {
	return keysManagement.ArgsKeysHandler{
		ManagedPeersHolder: &testscommon.ManagedPeersHolderStub{},
		PrivateKey: &cryptoMocks.PrivateKeyStub{
			ToByteArrayStub: func() ([]byte, error) {
				return testPrivateKeyBytes, nil
			},
			GeneratePublicStub: func() crypto.PublicKey {
				return &cryptoMocks.PublicKeyStub{
					ToByteArrayStub: func() ([]byte, error) {
						return testPublicKeyBytes, nil
					},
				}
			},
		},
		Pid: pid,
	}
}

func TestNewKeysHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil managed peers holder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKeysHandler()
		args.ManagedPeersHolder = nil
		handler, err := keysManagement.NewKeysHandler(args)

		assert.True(t, check.IfNil(handler))
		assert.Equal(t, keysManagement.ErrNilManagedPeersHolder, err)
	})
	t.Run("nil private key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKeysHandler()
		args.PrivateKey = nil
		handler, err := keysManagement.NewKeysHandler(args)

		assert.True(t, check.IfNil(handler))
		assert.Equal(t, keysManagement.ErrNilPrivateKey, err)
	})
	t.Run("empty pid should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKeysHandler()
		args.Pid = ""
		handler, err := keysManagement.NewKeysHandler(args)

		assert.True(t, check.IfNil(handler))
		assert.Equal(t, keysManagement.ErrEmptyPeerID, err)
	})
	t.Run("public key bytes generation errors should error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockArgsKeysHandler()
		args.PrivateKey = &cryptoMocks.PrivateKeyStub{
			GeneratePublicStub: func() crypto.PublicKey {
				return &cryptoMocks.PublicKeyStub{
					ToByteArrayStub: func() ([]byte, error) {
						return nil, expectedErr
					},
				}
			},
		}
		handler, err := keysManagement.NewKeysHandler(args)

		assert.True(t, check.IfNil(handler))
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKeysHandler()
		handler, err := keysManagement.NewKeysHandler(args)

		assert.False(t, check.IfNil(handler))
		assert.Nil(t, err)
		assert.Equal(t, testPublicKeyBytes, handler.PublicKeyBytes())
		assert.Equal(t, pid, handler.Pid())
		assert.False(t, check.IfNil(handler.PrivateKey()))
		assert.False(t, check.IfNil(handler.ManagedPeersHolder()))
		assert.False(t, check.IfNil(handler.PrivateKey()))
	})
}

func TestKeysHandler_GetHandledPrivateKey(t *testing.T) {
	t.Parallel()

	t.Run("is original public key of the node", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKeysHandler()
		handler, _ := keysManagement.NewKeysHandler(args)

		sk := handler.GetHandledPrivateKey(testPublicKeyBytes)
		assert.True(t, sk == handler.PrivateKey()) // pointer testing
	})
	t.Run("managedPeersHolder.GetPrivateKey errors", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKeysHandler()
		args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
			GetPrivateKeyCalled: func(pkBytes []byte) (crypto.PrivateKey, error) {
				return nil, errors.New("private key not found")
			},
		}
		handler, _ := keysManagement.NewKeysHandler(args)

		sk := handler.GetHandledPrivateKey(randomPublicKeyBytes)
		assert.True(t, sk == handler.PrivateKey()) // pointer testing
	})
	t.Run("managedPeersHolder.GetPrivateKey returns the private key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKeysHandler()
		handledPrivateKey := &cryptoMocks.PrivateKeyStub{}
		args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
			GetPrivateKeyCalled: func(pkBytes []byte) (crypto.PrivateKey, error) {
				assert.Equal(t, randomPublicKeyBytes, pkBytes)
				return handledPrivateKey, nil
			},
		}
		handler, _ := keysManagement.NewKeysHandler(args)

		sk := handler.GetHandledPrivateKey(randomPublicKeyBytes)
		assert.True(t, sk == handledPrivateKey) // pointer testing
	})
}

func TestKeysHandler_GetP2PIdentity(t *testing.T) {
	t.Parallel()

	p2pPrivateKeyBytes := []byte("p2p private key bytes")
	wasCalled := false
	args := createMockArgsKeysHandler()
	args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
		GetP2PIdentityCalled: func(pkBytes []byte) ([]byte, core.PeerID, error) {
			assert.Equal(t, randomPublicKeyBytes, pkBytes)
			wasCalled = true

			return p2pPrivateKeyBytes, pid, nil
		},
	}
	handler, _ := keysManagement.NewKeysHandler(args)

	recoveredPrivateKey, recoveredPid, err := handler.GetP2PIdentity(randomPublicKeyBytes)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, p2pPrivateKeyBytes, recoveredPrivateKey)
	assert.Equal(t, pid, recoveredPid)
}

func TestKeysHandler_IsKeyManagedByCurrentNode(t *testing.T) {
	t.Parallel()

	wasCalled := false
	args := createMockArgsKeysHandler()
	args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
		IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
			assert.Equal(t, randomPublicKeyBytes, pkBytes)
			wasCalled = true
			return true
		},
	}
	handler, _ := keysManagement.NewKeysHandler(args)

	isManaged := handler.IsKeyManagedByCurrentNode(randomPublicKeyBytes)
	assert.True(t, wasCalled)
	assert.True(t, isManaged)
}

func TestKeysHandler_IncrementRoundsWithoutReceivedMessages(t *testing.T) {
	t.Parallel()

	wasCalled := false
	args := createMockArgsKeysHandler()
	args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
		IncrementRoundsWithoutReceivedMessagesCalled: func(pkBytes []byte) {
			assert.Equal(t, randomPublicKeyBytes, pkBytes)
			wasCalled = true
		},
	}
	handler, _ := keysManagement.NewKeysHandler(args)

	handler.IncrementRoundsWithoutReceivedMessages(randomPublicKeyBytes)
	assert.True(t, wasCalled)
}

func TestKeysHandler_GetAssociatedPid(t *testing.T) {
	t.Parallel()

	t.Run("is original public key of the node", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKeysHandler()
		handler, _ := keysManagement.NewKeysHandler(args)

		recoveredPid := handler.GetAssociatedPid(testPublicKeyBytes)
		assert.True(t, recoveredPid == args.Pid)
	})
	t.Run("managedPeersHolder.GetP2PIdentity errors", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKeysHandler()
		args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
			GetP2PIdentityCalled: func(pkBytes []byte) ([]byte, core.PeerID, error) {
				return nil, "", errors.New("identity not found")
			},
		}
		handler, _ := keysManagement.NewKeysHandler(args)

		recoveredPid := handler.GetAssociatedPid(randomPublicKeyBytes)
		assert.True(t, recoveredPid == args.Pid)
	})
	t.Run("managedPeersHolder.GetP2PIdentity returns the identity", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKeysHandler()
		args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
			GetP2PIdentityCalled: func(pkBytes []byte) ([]byte, core.PeerID, error) {
				assert.Equal(t, randomPublicKeyBytes, pkBytes)

				return make([]byte, 0), pid, nil
			},
		}
		handler, _ := keysManagement.NewKeysHandler(args)

		recoveredPid := handler.GetAssociatedPid(randomPublicKeyBytes)
		assert.Equal(t, pid, recoveredPid)
	})
}

func TestKeysHandler_ResetRoundsWithoutReceivedMessages(t *testing.T) {
	t.Parallel()

	mapResetCalled := make(map[string]int)
	args := createMockArgsKeysHandler()
	args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
		ResetRoundsWithoutReceivedMessagesCalled: func(pkBytes []byte, pid core.PeerID) {
			mapResetCalled[string(pkBytes)]++
		},
	}
	handler, _ := keysManagement.NewKeysHandler(args)

	randomPid := core.PeerID("random pid")
	handler.ResetRoundsWithoutReceivedMessages(randomPublicKeyBytes, randomPid)
	assert.Equal(t, 1, len(mapResetCalled))
	assert.Equal(t, 1, mapResetCalled[string(randomPublicKeyBytes)])
}

func TestKeysHandler_GetRedundancyStepInReason(t *testing.T) {
	t.Parallel()

	expectedString := "expected string"
	args := createMockArgsKeysHandler()
	args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
		GetRedundancyStepInReasonCalled: func() string {
			return expectedString
		},
	}

	handler, _ := keysManagement.NewKeysHandler(args)
	assert.Equal(t, expectedString, handler.GetRedundancyStepInReason())
}
