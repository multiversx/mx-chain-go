package factory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/outport/process/alteredaccounts"
)

func TestCreateOutportDataProviderDisabled(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProviderFactory()
	arg.HasDrivers = false

	factory := NewOutportDataProviderFactory()
	provider, err := factory.CreateOutportDataProvider(arg)
	require.Nil(t, err)
	require.NotNil(t, provider)
	require.Equal(t, "*disabled.disabledOutportDataProvider", fmt.Sprintf("%T", provider))
}

func TestCreateOutportDataProviderError(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProviderFactory()
	arg.HasDrivers = true
	arg.AddressConverter = nil

	factory := NewOutportDataProviderFactory()
	provider, err := factory.CreateOutportDataProvider(arg)
	require.Nil(t, provider)
	require.Equal(t, alteredaccounts.ErrNilPubKeyConverter, err)
}

func TestCreateOutportDataProvider(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProviderFactory()
	arg.HasDrivers = true

	factory := NewOutportDataProviderFactory()
	provider, err := factory.CreateOutportDataProvider(arg)
	require.Nil(t, err)
	require.NotNil(t, provider)
	require.Equal(t, "*process.outportDataProvider", fmt.Sprintf("%T", provider))
	require.False(t, factory.IsInterfaceNil())
}
