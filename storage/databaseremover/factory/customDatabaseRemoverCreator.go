package factory

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/databaseremover"
	"github.com/ElrondNetwork/elrond-go/storage/databaseremover/disabled"
)

// CreateCustomDatabaseRemover will handle the creation of a custom database remover based on the configuration
func CreateCustomDatabaseRemover(storagePruningConfig config.StoragePruningConfig) (storage.CustomDatabaseRemoverHandler, error) {
	if storagePruningConfig.AccountsTrieCleanOldEpochsData {
		return databaseremover.NewCustomDatabaseRemover(storagePruningConfig)
	}

	return disabled.NewDisabledCustomDatabaseRemover(), nil
}
