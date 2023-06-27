package storageBootstrap

import (
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignShardStorageBootstrapperFactory(t *testing.T) {
	t.Parallel()

	ssbf, err := NewSovereignShardStorageBootstrapperFactory(nil)

	require.Nil(t, ssbf)
	require.Equal(t, errors.ErrNilShardStorageBootstrapperFactory, err)

	sbf, _ := NewShardStorageBootstrapperFactory()
	ssbf, err = NewSovereignShardStorageBootstrapperFactory(sbf)

	require.NotNil(t, ssbf)
	require.Nil(t, err)
}

func TestSovereignShardStorageBootstrapperFactory_CreateShardStorageBootstrapper(t *testing.T) {
	t.Parallel()

	sbf, _ := NewShardStorageBootstrapperFactory()
	ssbf, _ := NewSovereignShardStorageBootstrapperFactory(sbf)
	sb, err := ssbf.CreateShardStorageBootstrapper(ArgsShardStorageBootstrapper{})

	require.Nil(t, sb)
	require.NotNil(t, err)

	sb, err = ssbf.CreateShardStorageBootstrapper(getDefaultArgShardBootstrapper())

	require.NotNil(t, sb)
	require.Nil(t, err)
}

func TestSovereignShardStorageBootstrapperFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	sbf, _ := NewShardStorageBootstrapperFactory()
	ssbf, _ := NewSovereignShardStorageBootstrapperFactory(sbf)

	require.False(t, ssbf.IsInterfaceNil())

	ssbf = nil
	require.True(t, ssbf.IsInterfaceNil())
}
