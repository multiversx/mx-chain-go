package clean

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/stretchr/testify/require"
)

func TestNewOldDataCleanerProvider_NilNodeTypeProviderShouldErr(t *testing.T) {
	t.Parallel()

	odcp, err := NewOldDataCleanerProvider(nil, config.StoragePruningConfig{})
	require.Nil(t, odcp)
	require.Equal(t, storage.ErrNilNodeTypeProvider, err)
}

func TestNewOldDataCleanerProvider_ShouldWork(t *testing.T) {
	t.Parallel()

	odcp, err := NewOldDataCleanerProvider(&nodeTypeProviderMock.NodeTypeProviderStub{}, config.StoragePruningConfig{})
	require.NoError(t, err)
	require.NotNil(t, odcp)
}

func TestOldDataCleanerProvider_ShouldCleanShouldReturnObserverIfInvalidNodeType(t *testing.T) {
	t.Parallel()

	ntp := &nodeTypeProviderMock.NodeTypeProviderStub{
		GetTypeCalled: func() core.NodeType {
			return "invalid"
		},
	}

	odcp, _ := NewOldDataCleanerProvider(ntp, config.StoragePruningConfig{
		ObserverCleanOldEpochsData:  true,
		ValidatorCleanOldEpochsData: true,
	})

	require.False(t, odcp.ShouldClean())
}

func TestOldDataCleanerProvider_ShouldClean(t *testing.T) {
	t.Parallel()

	storagePruningConfig := config.StoragePruningConfig{
		ObserverCleanOldEpochsData:  false,
		ValidatorCleanOldEpochsData: true,
	}

	ntp := &nodeTypeProviderMock.NodeTypeProviderStub{
		GetTypeCalled: func() core.NodeType {
			return core.NodeTypeValidator
		},
	}

	odcp, _ := NewOldDataCleanerProvider(ntp, storagePruningConfig)
	require.NotNil(t, odcp)

	require.True(t, odcp.ShouldClean())

	odcp.nodeTypeProvider = &nodeTypeProviderMock.NodeTypeProviderStub{
		GetTypeCalled: func() core.NodeType {
			return core.NodeTypeObserver
		},
	}
	require.False(t, odcp.ShouldClean())
}

func TestOldDataCleanerProvider_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var odcp *oldDataCleanerProvider
	require.True(t, odcp.IsInterfaceNil())

	odcp, _ = NewOldDataCleanerProvider(&nodeTypeProviderMock.NodeTypeProviderStub{}, config.StoragePruningConfig{})
	require.False(t, odcp.IsInterfaceNil())
}
