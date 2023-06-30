package bootstrap

import (
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignEpochStartBootstrapperFactory(t *testing.T) {
	t.Parallel()

	sebf, err := NewSovereignEpochStartBootstrapperFactory(nil)

	require.Nil(t, sebf)
	require.Equal(t, errors.ErrNilEpochStartBootstrapperFactory, err)

	esbf, _ := NewEpochStartBootstrapperFactory()
	sebf, err = NewSovereignEpochStartBootstrapperFactory(esbf)

	require.Nil(t, err)
	require.NotNil(t, sebf)
}

func TestSovereignEpochStartBootstrapperFactory_CreateEpochStartBootstrapper(t *testing.T) {
	t.Parallel()

	esbf, _ := NewEpochStartBootstrapperFactory()
	sebf, _ := NewSovereignEpochStartBootstrapperFactory(esbf)

	seb, err := sebf.CreateEpochStartBootstrapper(getDefaultArgs())

	require.Nil(t, err)
	require.NotNil(t, seb)
}

func TestSovereignEpochStartBootstrapperFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	esbf, _ := NewEpochStartBootstrapperFactory()
	sebf, _ := NewSovereignEpochStartBootstrapperFactory(esbf)

	require.False(t, sebf.IsInterfaceNil())

	sebf = (*sovereignEpochStartBootstrapperFactory)(nil)
	require.True(t, sebf.IsInterfaceNil())
}
