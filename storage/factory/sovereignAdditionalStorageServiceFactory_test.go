package factory_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignAdditionalStorageServiceFactory(t *testing.T) {
	t.Parallel()

	f, err := factory.NewSovereignAdditionalStorageServiceFactory()
	require.NotNil(t, f)
	require.NoError(t, err)
}

func TestSovereignAdditionalStorageServiceFactory_CreateAdditionalStorageUnits(t *testing.T) {
	t.Parallel()

	t.Run("nil function should err", func(t *testing.T) {
		t.Parallel()

		f, _ := factory.NewSovereignAdditionalStorageServiceFactory()
		err := f.CreateAdditionalStorageUnits(nil, nil, "")
		require.ErrorIs(t, errors.ErrNilFunction, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		f, _ := factory.NewSovereignAdditionalStorageServiceFactory()

		wasCalled := false
		fParam := func(store dataRetriever.StorageService, shardID string) error {
			wasCalled = true
			return nil
		}

		err := f.CreateAdditionalStorageUnits(fParam, nil, "")
		require.NoError(t, err)
		require.True(t, wasCalled)
	})
}

func TestSovereignAdditionalStorageServiceFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	f, _ := factory.NewSovereignAdditionalStorageServiceFactory()
	require.False(t, f.IsInterfaceNil())
}
