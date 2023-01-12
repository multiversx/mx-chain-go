package factory

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

func TestCreateCustomDatabaseRemover(t *testing.T) {
	t.Parallel()

	t.Run("should create real custom database remover", func(t *testing.T) {
		t.Parallel()

		storagePruningArgs := config.StoragePruningConfig{
			AccountsTrieCleanOldEpochsData:       true,
			AccountsTrieSkipRemovalCustomPattern: "%1",
		}

		removerInstance, err := CreateCustomDatabaseRemover(storagePruningArgs)
		require.NoError(t, err)

		require.Equal(t, "*databaseremover.customDatabaseRemover", fmt.Sprintf("%T", removerInstance))
	})

	t.Run("should create disabled custom database remover", func(t *testing.T) {
		t.Parallel()

		storagePruningArgs := config.StoragePruningConfig{
			AccountsTrieCleanOldEpochsData:       false,
			AccountsTrieSkipRemovalCustomPattern: "%1",
		}

		removerInstance, err := CreateCustomDatabaseRemover(storagePruningArgs)
		require.NoError(t, err)

		require.Equal(t, "*disabled.disabledCustomDatabaseRemover", fmt.Sprintf("%T", removerInstance))
	})
}
