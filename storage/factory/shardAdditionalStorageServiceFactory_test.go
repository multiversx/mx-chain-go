package factory_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/stretchr/testify/require"
)

func TestNewShardAdditionalStorageServiceFactory(t *testing.T) {
	t.Parallel()

	f, err := factory.NewShardAdditionalStorageServiceFactory()
	require.NotNil(t, f)
	require.NoError(t, err)
}

func TestShardAdditionalStorageServiceFactory_CreateAdditionalStorageUnits(t *testing.T) {
	t.Parallel()

	f, _ := factory.NewShardAdditionalStorageServiceFactory()
	err := f.CreateAdditionalStorageUnits(nil, nil, "")
	require.NoError(t, err)
}

func TestShardAdditionalStorageServiceFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	f, _ := factory.NewShardAdditionalStorageServiceFactory()
	require.False(t, f.IsInterfaceNil())
}
