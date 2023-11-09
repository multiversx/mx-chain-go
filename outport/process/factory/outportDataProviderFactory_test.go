package factory

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/outport/process/alteredaccounts"
	"github.com/stretchr/testify/require"
)

func TestCreateOutportDataProviderDisabled(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProviderFactory()
	arg.HasDrivers = false

	provider, err := CreateOutportDataProvider(arg)
	require.Nil(t, err)
	require.NotNil(t, provider)
	require.Equal(t, "*disabled.disabledOutportDataProvider", fmt.Sprintf("%T", provider))
}

func TestCreateOutportDataProviderError(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProviderFactory()
	arg.HasDrivers = true
	arg.AddressConverter = nil

	provider, err := CreateOutportDataProvider(arg)
	require.Nil(t, provider)
	require.Equal(t, alteredaccounts.ErrNilPubKeyConverter, err)
}

func TestCreateOutportDataProvider(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProviderFactory()
	arg.HasDrivers = true

	provider, err := CreateOutportDataProvider(arg)
	require.Nil(t, err)
	require.NotNil(t, provider)
	require.Equal(t, "*process.outportDataProvider", fmt.Sprintf("%T", provider))
}
