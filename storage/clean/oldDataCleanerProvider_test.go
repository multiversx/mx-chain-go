package clean

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/stretchr/testify/require"
)

func createMockArgOldDataCleanerProvider() ArgOldDataCleanerProvider {
	return ArgOldDataCleanerProvider{
		NodeTypeProvider:    &nodeTypeProviderMock.NodeTypeProviderStub{},
		PruningStorerConfig: config.StoragePruningConfig{},
		ManagedPeersHolder:  &testscommon.ManagedPeersHolderStub{},
	}
}

func TestNewOldDataCleanerProvider(t *testing.T) {
	t.Parallel()

	t.Run("nil NodeTypeProvider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgOldDataCleanerProvider()
		args.NodeTypeProvider = nil
		odcp, err := NewOldDataCleanerProvider(args)
		require.Nil(t, odcp)
		require.Equal(t, storage.ErrNilNodeTypeProvider, err)
	})
	t.Run("nil ManagedPeersHolder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgOldDataCleanerProvider()
		args.ManagedPeersHolder = nil
		odcp, err := NewOldDataCleanerProvider(args)
		require.Nil(t, odcp)
		require.Equal(t, storage.ErrNilManagedPeersHolder, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		odcp, err := NewOldDataCleanerProvider(createMockArgOldDataCleanerProvider())
		require.NoError(t, err)
		require.NotNil(t, odcp)
	})
}

func TestOldDataCleanerProvider_ShouldClean(t *testing.T) {
	t.Parallel()

	t.Run("invalid type should not clean", func(t *testing.T) {
		t.Parallel()

		args := createMockArgOldDataCleanerProvider()
		args.NodeTypeProvider = &nodeTypeProviderMock.NodeTypeProviderStub{
			GetTypeCalled: func() core.NodeType {
				return "invalid"
			},
		}
		args.PruningStorerConfig = config.StoragePruningConfig{
			ObserverCleanOldEpochsData:  true,
			ValidatorCleanOldEpochsData: true,
		}
		odcp, _ := NewOldDataCleanerProvider(args)

		require.False(t, odcp.ShouldClean())
	})
	t.Run("observer should clean", func(t *testing.T) {
		t.Parallel()

		args := createMockArgOldDataCleanerProvider()
		args.PruningStorerConfig = config.StoragePruningConfig{
			ObserverCleanOldEpochsData:  true,
			ValidatorCleanOldEpochsData: false,
		}

		args.NodeTypeProvider = &nodeTypeProviderMock.NodeTypeProviderStub{
			GetTypeCalled: func() core.NodeType {
				return core.NodeTypeObserver
			},
		}

		odcp, _ := NewOldDataCleanerProvider(args)
		require.NotNil(t, odcp)

		require.True(t, odcp.ShouldClean())
	})
	t.Run("validator should clean", func(t *testing.T) {
		t.Parallel()

		args := createMockArgOldDataCleanerProvider()
		args.PruningStorerConfig = config.StoragePruningConfig{
			ObserverCleanOldEpochsData:  false,
			ValidatorCleanOldEpochsData: true,
		}

		args.NodeTypeProvider = &nodeTypeProviderMock.NodeTypeProviderStub{
			GetTypeCalled: func() core.NodeType {
				return core.NodeTypeValidator
			},
		}

		odcp, _ := NewOldDataCleanerProvider(args)
		require.NotNil(t, odcp)

		require.True(t, odcp.ShouldClean())
	})
	t.Run("multi key observer should clean", func(t *testing.T) {
		t.Parallel()

		args := createMockArgOldDataCleanerProvider()
		args.ManagedPeersHolder = &testscommon.ManagedPeersHolderStub{
			IsMultiKeyModeCalled: func() bool {
				return true
			},
		}
		args.PruningStorerConfig = config.StoragePruningConfig{
			ObserverCleanOldEpochsData:  false,
			ValidatorCleanOldEpochsData: true,
		}

		args.NodeTypeProvider = &nodeTypeProviderMock.NodeTypeProviderStub{
			GetTypeCalled: func() core.NodeType {
				return core.NodeTypeObserver
			},
		}

		odcp, _ := NewOldDataCleanerProvider(args)
		require.NotNil(t, odcp)

		require.True(t, odcp.ShouldClean())
	})
}

func TestOldDataCleanerProvider_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var odcp *oldDataCleanerProvider
	require.True(t, odcp.IsInterfaceNil())

	args := createMockArgOldDataCleanerProvider()
	odcp, _ = NewOldDataCleanerProvider(args)
	require.False(t, odcp.IsInterfaceNil())
}
