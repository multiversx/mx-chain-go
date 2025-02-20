package storageBootstrap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSovereignShardStorageBootstrapperFactory(t *testing.T) {
	t.Parallel()

	ssbf := NewSovereignShardStorageBootstrapperFactory()
	require.NotNil(t, ssbf)
}

func TestSovereignShardStorageBootstrapperFactory_CreateShardStorageBootstrapper(t *testing.T) {
	t.Parallel()

	ssbf := NewSovereignShardStorageBootstrapperFactory()
	sb, err := ssbf.CreateBootstrapperFromStorage(ArgsShardStorageBootstrapper{})

	require.Nil(t, sb)
	require.NotNil(t, err)

	sb, err = ssbf.CreateBootstrapperFromStorage(getDefaultArgShardBootstrapper())

	require.NotNil(t, sb)
	require.Nil(t, err)
}

func TestSovereignShardStorageBootstrapperFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	ssbf := NewSovereignShardStorageBootstrapperFactory()

	require.False(t, ssbf.IsInterfaceNil())

	ssbf = nil
	require.True(t, ssbf.IsInterfaceNil())
}
