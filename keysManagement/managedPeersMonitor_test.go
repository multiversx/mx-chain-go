package keysManagement

import (
	"errors"
	"testing"

	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")

func createMockArgManagedPeersMonitor() ArgManagedPeersMonitor {
	return ArgManagedPeersMonitor{
		ManagedPeersHolder: &testscommon.ManagedPeersHolderStub{},
		NodesCoordinator:   &shardingMocks.NodesCoordinatorStub{},
		ShardProvider:      &testscommon.ShardsCoordinatorMock{},
		EpochProvider:      &epochNotifier.EpochNotifierStub{},
	}
}

func TestNewManagedPeersMonitor(t *testing.T) {
	t.Parallel()

	t.Run("nil ManagedPeersHolder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgManagedPeersMonitor()
		args.ManagedPeersHolder = nil
		monitor, err := NewManagedPeersMonitor(args)
		require.Equal(t, ErrNilManagedPeersHolder, err)
		require.Nil(t, monitor)
	})
	t.Run("nil NodesCoordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgManagedPeersMonitor()
		args.NodesCoordinator = nil
		monitor, err := NewManagedPeersMonitor(args)
		require.Equal(t, ErrNilNodesCoordinator, err)
		require.Nil(t, monitor)
	})
	t.Run("nil ShardProvider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgManagedPeersMonitor()
		args.ShardProvider = nil
		monitor, err := NewManagedPeersMonitor(args)
		require.Equal(t, ErrNilShardProvider, err)
		require.Nil(t, monitor)
	})
	t.Run("nil EpochProvider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgManagedPeersMonitor()
		args.EpochProvider = nil
		monitor, err := NewManagedPeersMonitor(args)
		require.Equal(t, ErrNilEpochProvider, err)
		require.Nil(t, monitor)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		monitor, err := NewManagedPeersMonitor(createMockArgManagedPeersMonitor())
		require.NoError(t, err)
		require.NotNil(t, monitor)
	})
}

func TestManagedPeersMonitor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var monitor *managedPeersMonitor
	require.True(t, monitor.IsInterfaceNil())

	monitor, _ = NewManagedPeersMonitor(createMockArgManagedPeersMonitor())
	require.False(t, monitor.IsInterfaceNil())
}

func TestManagedPeersMonitor_GetManagedKeysCount(t *testing.T) {
	t.Parallel()

	managedKeys := map[string]crypto.PrivateKey{
		"pk1": &cryptoMocks.PrivateKeyStub{},
		"pk2": &cryptoMocks.PrivateKeyStub{},
		"pk3": &cryptoMocks.PrivateKeyStub{},
		"pk4": &cryptoMocks.PrivateKeyStub{},
	}
	args := createMockArgManagedPeersMonitor()
	args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
		GetManagedKeysByCurrentNodeCalled: func() map[string]crypto.PrivateKey {
			return managedKeys
		},
	}
	monitor, err := NewManagedPeersMonitor(args)
	require.NoError(t, err)

	require.Equal(t, len(managedKeys), monitor.GetManagedKeysCount())
}

func TestManagedPeersMonitor_GetEligibleManagedKeys(t *testing.T) {
	t.Parallel()

	t.Run("nodes coordinator returns error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgManagedPeersMonitor()
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
				return nil, expectedErr
			},
		}
		monitor, err := NewManagedPeersMonitor(args)
		require.NoError(t, err)

		keys, err := monitor.GetEligibleManagedKeys()
		require.Equal(t, expectedErr, err)
		require.Nil(t, keys)
	})
	t.Run("shard not found should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgManagedPeersMonitor()
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
				return map[uint32][][]byte{}, nil
			},
		}
		monitor, err := NewManagedPeersMonitor(args)
		require.NoError(t, err)

		keys, err := monitor.GetEligibleManagedKeys()
		require.True(t, errors.Is(err, ErrInvalidValue))
		require.Nil(t, keys)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		selfShard := uint32(1)
		eligibleValidators := map[uint32][][]byte{
			selfShard: {
				[]byte("managed 1"),
				[]byte("managed 2"),
				[]byte("not managed 1"),
				[]byte("not managed 2"),
			},
			100: {
				[]byte("managed 3"),
				[]byte("not managed 3"),
			},
		}
		args := createMockArgManagedPeersMonitor()
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
				return eligibleValidators, nil
			},
		}
		args.ShardProvider = &testscommon.ShardsCoordinatorMock{
			SelfIDCalled: func() uint32 {
				return selfShard
			},
		}
		args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				return string(pkBytes) == "managed 1" ||
					string(pkBytes) == "managed 2" ||
					string(pkBytes) == "managed 3"
			},
		}
		monitor, err := NewManagedPeersMonitor(args)
		require.NoError(t, err)

		keys, err := monitor.GetEligibleManagedKeys()
		require.NoError(t, err)

		require.Equal(t, [][]byte{[]byte("managed 1"), []byte("managed 2")}, keys)
	})
}

func TestManagedPeersMonitor_GetGetWaitingManagedKeys(t *testing.T) {
	t.Parallel()

	t.Run("nodes coordinator returns error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgManagedPeersMonitor()
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetAllWaitingValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
				return nil, expectedErr
			},
		}
		monitor, err := NewManagedPeersMonitor(args)
		require.NoError(t, err)

		keys, err := monitor.GetWaitingManagedKeys()
		require.Equal(t, expectedErr, err)
		require.Nil(t, keys)
	})
	t.Run("shard not found should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgManagedPeersMonitor()
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetAllWaitingValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
				return map[uint32][][]byte{}, nil
			},
		}
		monitor, err := NewManagedPeersMonitor(args)
		require.NoError(t, err)

		keys, err := monitor.GetWaitingManagedKeys()
		require.True(t, errors.Is(err, ErrInvalidValue))
		require.Nil(t, keys)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		selfShard := uint32(1)
		eligibleValidators := map[uint32][][]byte{
			selfShard: {
				[]byte("managed 1"),
				[]byte("managed 2"),
				[]byte("not managed 1"),
				[]byte("not managed 2"),
			},
			100: {
				[]byte("managed 3"),
				[]byte("not managed 3"),
			},
		}
		args := createMockArgManagedPeersMonitor()
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetAllWaitingValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
				return eligibleValidators, nil
			},
		}
		args.ShardProvider = &testscommon.ShardsCoordinatorMock{
			SelfIDCalled: func() uint32 {
				return selfShard
			},
		}
		args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				return string(pkBytes) == "managed 1" ||
					string(pkBytes) == "managed 2" ||
					string(pkBytes) == "managed 3"
			},
		}
		monitor, err := NewManagedPeersMonitor(args)
		require.NoError(t, err)

		keys, err := monitor.GetWaitingManagedKeys()
		require.NoError(t, err)

		require.Equal(t, [][]byte{[]byte("managed 1"), []byte("managed 2")}, keys)
	})
}

func TestManagedPeersMonitor_GetManagedKeys(t *testing.T) {
	t.Parallel()

	managedKeys := map[string]crypto.PrivateKey{
		"pk2": &cryptoMocks.PrivateKeyStub{},
		"pk1": &cryptoMocks.PrivateKeyStub{},
		"pk3": &cryptoMocks.PrivateKeyStub{},
	}
	expectedManagedKeys := [][]byte{[]byte("pk1"), []byte("pk2"), []byte("pk3")}
	args := createMockArgManagedPeersMonitor()
	args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
		GetManagedKeysByCurrentNodeCalled: func() map[string]crypto.PrivateKey {
			return managedKeys
		},
	}
	monitor, err := NewManagedPeersMonitor(args)
	require.NoError(t, err)

	keys := monitor.GetManagedKeys()
	require.Equal(t, expectedManagedKeys, keys)
}

func TestManagedPeersMonitor_GetLoadedKeys(t *testing.T) {
	t.Parallel()

	loadedKeys := [][]byte{[]byte("pk1"), []byte("pk2"), []byte("pk3")}
	args := createMockArgManagedPeersMonitor()
	args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
		GetLoadedKeysByCurrentNodeCalled: func() [][]byte {
			return loadedKeys
		},
	}
	monitor, err := NewManagedPeersMonitor(args)
	require.NoError(t, err)

	keys := monitor.GetLoadedKeys()
	require.Equal(t, loadedKeys, keys)
}
