package reflectcommon

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/stretchr/testify/require"
)

func TestAdaptStructureValueBasedOnPath(t *testing.T) {
	t.Parallel()

	t.Run("nil object to change, should err", func(t *testing.T) {
		t.Parallel()

		err := AdaptStructureValueBasedOnPath(nil, "path", "n/a")

		require.Equal(t, "nil structure to update", err.Error())
	})

	t.Run("wrong path, should not panic, but catch the error", func(t *testing.T) {
		t.Parallel()

		path := "StoragePruning.InvalidFieldName"
		expectedNewValue := "%5050"
		cfg := &config.Config{}

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.Equal(t, "invalid structure name: InvalidFieldName", err.Error())
	})

	t.Run("should error when trying to set int as string", func(t *testing.T) {
		t.Parallel()

		path := "StoragePruning.AccountsTrieSkipRemovalCustomPattern"
		expectedNewValue := 37
		cfg := &config.Config{}
		cfg.StoragePruning.AccountsTrieSkipRemovalCustomPattern = "%50"

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.Equal(t, "reflect.Set: value of type int is not assignable to type string", err.Error())
	})

	t.Run("should error when trying to set uint64 as uint32", func(t *testing.T) {
		t.Parallel()

		path := "StoragePruning.FullArchiveNumActivePersisters"
		expectedNewValue := uint64(37)
		cfg := &config.Config{}
		cfg.StoragePruning.FullArchiveNumActivePersisters = uint32(50)

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.Equal(t, "reflect.Set: value of type uint64 is not assignable to type uint32", err.Error())
	})

	t.Run("should error when invalid field during multiple levels depth", func(t *testing.T) {
		t.Parallel()

		path := "Hardfork.ExportKeysStorageConfig.DB2.FilePath" // DB2 instead of DB
		expectedNewValue := "new file path"
		cfg := &config.Config{}
		cfg.Hardfork.ExportKeysStorageConfig.DB.FilePath = "original file path"

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.Equal(t, "invalid structure name: DB2", err.Error())
	})

	t.Run("should error when the final value is invalid", func(t *testing.T) {
		t.Parallel()

		path := "Hardfork.ExportKeysStorageConfig.DB.FilePath2" // FilePath2 instead of FilePath
		expectedNewValue := "new file path"
		cfg := &config.Config{}
		cfg.Hardfork.ExportKeysStorageConfig.DB.FilePath = "original file path"

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.Equal(t, "invalid structure name: FilePath2", err.Error())
	})

	t.Run("should error if the field is un-settable / unexported", func(t *testing.T) {
		t.Parallel()

		type simpleStruct struct {
			a string
		}
		obj := &simpleStruct{a: "original"}
		path := "a"
		expectedNewValue := "new"

		err := AdaptStructureValueBasedOnPath(obj, path, expectedNewValue)

		require.Equal(t, "cannot set value for field. it or it's structure might be unexported. field name=a", err.Error())
	})

	t.Run("should error if the field in the middle of the path is un-settable / unexported", func(t *testing.T) {
		t.Parallel()

		type innerStruct2 struct {
			C int
		}
		type innerStruct1 struct {
			b innerStruct2
		}
		type simpleStruct struct {
			A innerStruct1
		}
		obj := &simpleStruct{
			A: innerStruct1{
				b: innerStruct2{
					C: 5,
				},
			},
		}
		path := "A.b.C"
		expectedNewValue := 37

		err := AdaptStructureValueBasedOnPath(obj, path, expectedNewValue)

		require.Equal(t, "cannot set value for field. it or it's structure might be unexported. field name=C", err.Error())
	})

	t.Run("should work for single level structures", func(t *testing.T) {
		t.Parallel()

		type simpleStruct struct {
			A string
		}
		obj := &simpleStruct{A: "original"}
		path := "A"
		expectedNewValue := "new"

		err := AdaptStructureValueBasedOnPath(obj, path, expectedNewValue)

		require.NoError(t, err)
		require.Equal(t, expectedNewValue, obj.A)
	})

	t.Run("should error if the structure is passed as value", func(t *testing.T) {
		t.Parallel()

		type simpleStruct struct {
			A string
		}
		obj := simpleStruct{A: "original"}
		path := "A"
		expectedNewValue := "new"

		err := AdaptStructureValueBasedOnPath(obj, path, expectedNewValue)

		require.Equal(t, "cannot update structures that are not passed by pointer", err.Error())
	})

	t.Run("should work and override string value", func(t *testing.T) {
		t.Parallel()

		path := "StoragePruning.AccountsTrieSkipRemovalCustomPattern"
		expectedNewValue := "%5050"
		cfg := &config.Config{}
		cfg.StoragePruning.AccountsTrieSkipRemovalCustomPattern = "%50"

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.StoragePruning.AccountsTrieSkipRemovalCustomPattern)
	})

	t.Run("should work and override bool value", func(t *testing.T) {
		t.Parallel()

		path := "StoragePruning.AccountsTrieCleanOldEpochsData"
		cfg := &config.Config{}
		cfg.StoragePruning.AccountsTrieCleanOldEpochsData = false

		err := AdaptStructureValueBasedOnPath(cfg, path, true)
		require.NoError(t, err)

		require.True(t, cfg.StoragePruning.AccountsTrieCleanOldEpochsData)
	})

	t.Run("should work and override uint32 value", func(t *testing.T) {
		t.Parallel()

		path := "StoragePruning.FullArchiveNumActivePersisters"
		cfg := &config.Config{}
		cfg.StoragePruning.FullArchiveNumActivePersisters = uint32(50)
		expectedNewValue := uint32(37)

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.StoragePruning.FullArchiveNumActivePersisters)
	})

	t.Run("should work and override string value on multiple levels depth", func(t *testing.T) {
		t.Parallel()

		path := "Hardfork.ExportKeysStorageConfig.DB.FilePath"
		expectedNewValue := "new file path"
		cfg := &config.Config{}
		cfg.Hardfork.ExportKeysStorageConfig.DB.FilePath = "original file path"

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.Hardfork.ExportKeysStorageConfig.DB.FilePath)
	})

	t.Run("should work and override int value on multiple levels depth", func(t *testing.T) {
		t.Parallel()

		path := "Hardfork.ExportKeysStorageConfig.DB.MaxBatchSize"
		cfg := &config.Config{}
		cfg.Hardfork.ExportKeysStorageConfig.DB.MaxBatchSize = 10
		expectedNewValue := 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.Hardfork.ExportKeysStorageConfig.DB.MaxBatchSize)
	})
}

func BenchmarkAdaptStructureValueBasedOnPath(b *testing.B) {
	type testStruct struct {
		InnerStruct struct {
			YetAnotherInnerStruct struct {
				Value string
			}
		}
	}

	testObject := &testStruct{}
	for i := 0; i < b.N; i++ {
		_ = AdaptStructureValueBasedOnPath(testObject, "InnerStruct.YetAnotherInnerStruct.Value", "new")
	}
}
