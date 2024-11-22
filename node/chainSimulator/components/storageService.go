package components

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

// CreateStorageService creates a storage service for shard nodes
func CreateStorageService(numOfShards uint32, trieStoragePath string, config *config.Config) (dataRetriever.StorageService, error) {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.RewardTransactionUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.BootstrapUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.StatusMetricsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.ReceiptsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.ScheduledSCRsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.TxLogsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.UserAccountsUnit, CreateMemUnitForTries())
	store.AddStorer(dataRetriever.PeerAccountsUnit, CreateMemUnitForTries())
	store.AddStorer(dataRetriever.ESDTSuppliesUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.RoundHdrHashDataUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MiniblocksMetadataUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MiniblockHashByTxHashUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.EpochByHashUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.ResultsHashesByTxHashUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.TrieEpochRootHashUnit, CreateMemUnit())

	for i := uint32(0); i < numOfShards; i++ {
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i)
		store.AddStorer(hdrNonceHashDataUnit, CreateMemUnit())
	}

	if trieStoragePath == "" {
		return store, nil
	}

	config.AccountsTrieStorage.DB.FilePath = trieStoragePath
	storer, err := createStaticStorageUnit(config.AccountsTrieStorage)
	if err != nil {
		return nil, err
	}

	store.AddStorer(dataRetriever.UserAccountsUnit, storer)

	return store, nil
}

func createStaticStorageUnit(
	storageConf config.StorageConfig,
) (*storageunit.Unit, error) {
	storageUnitDBConf := factory.GetDBFromConfig(storageConf.DB)
	dbPath := storageConf.DB.FilePath
	storageUnitDBConf.FilePath = dbPath

	persisterCreator, err := factory.NewPersisterFactory(storageConf.DB)
	if err != nil {
		return nil, err
	}

	return storageunit.NewStorageUnitFromConf(
		factory.GetCacherFromConfig(storageConf.Cache),
		storageUnitDBConf,
		persisterCreator,
	)
}
