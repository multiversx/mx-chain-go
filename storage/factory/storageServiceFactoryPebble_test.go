package factory

import (
	"reflect"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-storage-go/pebbledb"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage/mock"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

// switchAllUnitsToPebble sets every top-level storage unit of the config to the pebble engine
func switchAllUnitsToPebble(cfg *config.Config) {
	storageConfigType := reflect.TypeOf(config.StorageConfig{})
	value := reflect.ValueOf(cfg).Elem()
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Type() != storageConfigType {
			continue
		}
		dbConfig := value.Field(i).FieldByName("DB")
		dbConfig.FieldByName("Type").SetString(string(storageunit.PebbleDB))
		dbConfig.FieldByName("PebbleProfile").SetString(pebbledb.HeavyWriteProfile)
	}
}

func TestStorageServiceFactory_CreateWithPebbleUnits(t *testing.T) {
	t.Parallel()

	t.Run("for shard", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		switchAllUnitsToPebble(&args.Config)
		resources, err := pebbledb.NewSharedResources(8<<20, 100)
		require.Nil(t, err)
		args.PebbleResources = resources

		storageServiceFactory, err := NewStorageServiceFactory(args)
		require.Nil(t, err)
		storageService, err := storageServiceFactory.CreateForShard()
		require.Nil(t, err)
		require.Equal(t, 25, len(storageService.GetAllStorers()))

		key, val := []byte("key"), []byte("value")
		for _, unit := range []dataRetriever.UnitType{dataRetriever.TransactionUnit, dataRetriever.UserAccountsUnit, dataRetriever.BlockHeaderUnit} {
			storer, errGet := storageService.GetStorer(unit)
			require.Nil(t, errGet)
			require.Nil(t, storer.Put(key, val))
			res, errGet := storer.Get(key)
			require.Nil(t, errGet)
			require.Equal(t, val, res)
		}

		require.Nil(t, storageService.CloseAll())
		resources.Close()
	})
	t.Run("for meta", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.ShardCoordinator = mock.NewShardCoordinatorMock(core.MetachainShardId, 3)
		switchAllUnitsToPebble(&args.Config)

		storageServiceFactory, err := NewStorageServiceFactory(args)
		require.Nil(t, err)
		storageService, err := storageServiceFactory.CreateForMeta()
		require.Nil(t, err)

		storer, err := storageService.GetStorer(dataRetriever.PeerAccountsUnit)
		require.Nil(t, err)
		require.Nil(t, storer.Put([]byte("key"), []byte("value")))
		require.Nil(t, storer.Has([]byte("key")))

		require.Nil(t, storageService.CloseAll())
	})
}
