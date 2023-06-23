package bootstrap

import (
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignEpochStartBootstrapperFactory(t *testing.T) {
	sebf, err := NewSovereignEpochStartBootstrapperFactory(nil)

	require.Nil(t, sebf)
	require.Equal(t, errors.ErrNilEpochStartBootstrapperFactory, err)

	esbf, _ := NewEpochStartBootstrapperFactory()
	sebf, err = NewSovereignEpochStartBootstrapperFactory(esbf)

	require.Nil(t, err)
	require.NotNil(t, sebf)
}

func TestSovereignEpochStartBootstrapperFactory_CreateEpochStartBootstrapper(t *testing.T) {
	esbf, _ := NewEpochStartBootstrapperFactory()
	sebf, _ := NewSovereignEpochStartBootstrapperFactory(esbf)

	seb, err := sebf.CreateEpochStartBootstrapper(GetDefaultArgs())

	require.Nil(t, err)
	require.NotNil(t, seb)
}

func TestSovereignEpochStartBootstrapperFactory_IsInterfaceNil(t *testing.T) {
	esbf, _ := NewEpochStartBootstrapperFactory()
	sebf, _ := NewSovereignEpochStartBootstrapperFactory(esbf)

	require.False(t, sebf.IsInterfaceNil())
}
