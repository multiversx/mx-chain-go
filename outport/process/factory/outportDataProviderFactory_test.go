package factory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/outport/process/alteredaccounts"
)

type outportFactory interface {
	CreateOutportDataProvider(arg ArgOutportDataProviderFactory) (outport.DataProviderOutport, error)
	IsInterfaceNil() bool
}

func TestCreateOutportDataProvider(t *testing.T) {
	t.Parallel()

	factory := NewOutportDataProviderFactory()
	testCreateOutportDataProviderDisabled(t, factory)
	testCreateOutportDataProviderError(t, factory)
	testCreateOutportDataProvider(t, factory, "*process.outportDataProvider")
}

func testCreateOutportDataProviderDisabled(t *testing.T, factory outportFactory) {
	arg := createArgOutportDataProviderFactory()
	arg.HasDrivers = false

	provider, err := factory.CreateOutportDataProvider(arg)
	require.Nil(t, err)
	require.NotNil(t, provider)
	require.Equal(t, "*disabled.disabledOutportDataProvider", fmt.Sprintf("%T", provider))
}

func testCreateOutportDataProviderError(t *testing.T, factory outportFactory) {
	arg := createArgOutportDataProviderFactory()
	arg.HasDrivers = true
	arg.AddressConverter = nil

	provider, err := factory.CreateOutportDataProvider(arg)
	require.Nil(t, provider)
	require.Equal(t, alteredaccounts.ErrNilPubKeyConverter, err)
}

func testCreateOutportDataProvider(t *testing.T, factory outportFactory, expectedType string) {
	arg := createArgOutportDataProviderFactory()
	arg.HasDrivers = true

	provider, err := factory.CreateOutportDataProvider(arg)
	require.Nil(t, err)
	require.NotNil(t, provider)
	require.Equal(t, expectedType, fmt.Sprintf("%T", provider))
	require.False(t, factory.IsInterfaceNil())
}
