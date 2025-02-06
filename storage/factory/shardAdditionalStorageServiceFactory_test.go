package factory_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/storage/factory"
)

func TestNewShardAdditionalStorageServiceFactory(t *testing.T) {
	t.Parallel()

	f := factory.NewShardAdditionalStorageServiceFactory()
	require.False(t, f.IsInterfaceNil())
}

func TestShardAdditionalStorageServiceFactory_CreateAdditionalStorageUnits(t *testing.T) {
	t.Parallel()

	f := factory.NewShardAdditionalStorageServiceFactory()
	err := f.CreateAdditionalStorageUnits(nil, nil, "")
	require.NoError(t, err)
}
