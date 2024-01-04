package reflectcommon

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
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

	t.Run("empty path, should not panic, but catch the error", func(t *testing.T) {
		t.Parallel()

		expectedNewValue := "%5050"
		cfg := &config.Config{}

		err := AdaptStructureValueBasedOnPath(cfg, "", expectedNewValue)

		require.Equal(t, "empty path to update", err.Error())
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

	t.Run("should error when setting on unsupported type", func(t *testing.T) {
		t.Parallel()

		path := "TrieSyncStorage.DB"
		expectedNewValue := "provided value"
		cfg := &config.Config{}

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "unsupported type <struct> when trying to set the value <provided value>")
	})

	t.Run("should error when setting invalid uint32", func(t *testing.T) {
		t.Parallel()

		path := "TrieSyncStorage.Capacity"
		expectedNewValue := "invalid uint32"
		cfg := &config.Config{}
		cfg.TrieSyncStorage.Capacity = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast field <invalid uint32> to kind <uint32>")
	})

	t.Run("should error when setting invalid uint64", func(t *testing.T) {
		t.Parallel()

		path := "TrieSyncStorage.SizeInBytes"
		expectedNewValue := "invalid uint64"
		cfg := &config.Config{}
		cfg.TrieSyncStorage.SizeInBytes = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast field <invalid uint64> to kind <uint64>")
	})

	t.Run("should error when setting invalid float32", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.MinPeersThreshold"
		expectedNewValue := "invalid float32"
		cfg := &config.Config{}
		cfg.HeartbeatV2.MinPeersThreshold = 37.0

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast field <invalid float32> to kind <float32>")
	})

	t.Run("should error when setting invalid float64", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.PeerShardTimeThresholdBetweenSends"
		expectedNewValue := "invalid float64"
		cfg := &config.Config{}
		cfg.HeartbeatV2.PeerShardTimeThresholdBetweenSends = 37.0

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast field <invalid float64> to kind <float64>")
	})

	t.Run("should error when setting invalid int64", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.HeartbeatExpiryTimespanInSec"
		expectedNewValue := "invalid int64"
		cfg := &config.Config{}
		cfg.HeartbeatV2.HeartbeatExpiryTimespanInSec = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast field <invalid int64> to kind <int64>")
	})

	t.Run("should error when setting invalid int64", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.HeartbeatExpiryTimespanInSec"
		expectedNewValue := "invalid int64"
		cfg := &config.Config{}
		cfg.HeartbeatV2.HeartbeatExpiryTimespanInSec = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast field <invalid int64> to kind <int64>")
	})

	t.Run("should error when setting invalid int", func(t *testing.T) {
		t.Parallel()

		path := "Debug.InterceptorResolver.DebugLineExpiration"
		expectedNewValue := "invalid int"
		cfg := &config.Config{}
		cfg.Debug.InterceptorResolver.DebugLineExpiration = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast field <invalid int> to kind <int>")
	})

	t.Run("should error when setting invalid bool", func(t *testing.T) {
		t.Parallel()

		path := "Debug.InterceptorResolver.EnablePrint"
		expectedNewValue := "invalid bool"
		cfg := &config.Config{}
		cfg.Debug.InterceptorResolver.EnablePrint = false

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast field <invalid bool> to kind <bool>")
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

		err := AdaptStructureValueBasedOnPath(obj, path, fmt.Sprintf("%d", expectedNewValue))

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

	t.Run("should work and override uint32 value", func(t *testing.T) {
		t.Parallel()

		path := "StoragePruning.FullArchiveNumActivePersisters"
		expectedNewValue := uint32(38)
		cfg := &config.Config{}
		cfg.StoragePruning.FullArchiveNumActivePersisters = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, fmt.Sprintf("%d", expectedNewValue))
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.StoragePruning.FullArchiveNumActivePersisters)
	})

	t.Run("should work and override float32 value", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.MinPeersThreshold"
		expectedNewValue := float32(38.0)
		cfg := &config.Config{}
		cfg.HeartbeatV2.MinPeersThreshold = 37.0

		err := AdaptStructureValueBasedOnPath(cfg, path, fmt.Sprintf("%f", expectedNewValue))
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.HeartbeatV2.MinPeersThreshold)
	})

	t.Run("should work and override float64 value", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.PeerAuthenticationTimeThresholdBetweenSends"
		expectedNewValue := 38.0
		cfg := &config.Config{}
		cfg.HeartbeatV2.PeerAuthenticationTimeThresholdBetweenSends = 37.0

		err := AdaptStructureValueBasedOnPath(cfg, path, fmt.Sprintf("%f", expectedNewValue))
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.HeartbeatV2.PeerAuthenticationTimeThresholdBetweenSends)
	})

	t.Run("should work and override int value", func(t *testing.T) {
		t.Parallel()

		path := "Debug.InterceptorResolver.DebugLineExpiration"
		expectedNewValue := 38
		cfg := &config.Config{}
		cfg.Debug.InterceptorResolver.DebugLineExpiration = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, fmt.Sprintf("%d", expectedNewValue))
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.Debug.InterceptorResolver.DebugLineExpiration)
	})

	t.Run("should work and override int64 value", func(t *testing.T) {
		t.Parallel()

		path := "Hardfork.GenesisTime"
		expectedNewValue := int64(38)
		cfg := &config.Config{}
		cfg.Hardfork.GenesisTime = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, fmt.Sprintf("%d", expectedNewValue))
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.Hardfork.GenesisTime)
	})

	t.Run("should work and override uint64 value", func(t *testing.T) {
		t.Parallel()

		path := "TrieSyncStorage.SizeInBytes"
		expectedNewValue := uint64(38)
		cfg := &config.Config{}
		cfg.TrieSyncStorage.SizeInBytes = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, fmt.Sprintf("%d", expectedNewValue))
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.TrieSyncStorage.SizeInBytes)
	})

	t.Run("should work and override bool value", func(t *testing.T) {
		t.Parallel()

		path := "StoragePruning.AccountsTrieCleanOldEpochsData"
		cfg := &config.Config{}
		cfg.StoragePruning.AccountsTrieCleanOldEpochsData = false

		err := AdaptStructureValueBasedOnPath(cfg, path, fmt.Sprintf("%v", true))
		require.NoError(t, err)

		require.True(t, cfg.StoragePruning.AccountsTrieCleanOldEpochsData)
	})

	t.Run("should work and override uint32 value", func(t *testing.T) {
		t.Parallel()

		path := "StoragePruning.FullArchiveNumActivePersisters"
		cfg := &config.Config{}
		cfg.StoragePruning.FullArchiveNumActivePersisters = uint32(50)
		expectedNewValue := uint32(37)

		err := AdaptStructureValueBasedOnPath(cfg, path, fmt.Sprintf("%d", expectedNewValue))
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.StoragePruning.FullArchiveNumActivePersisters)
	})

	t.Run("should work and override int32 value", func(t *testing.T) {
		t.Parallel()

		path := "Antiflood.NumConcurrentResolverJobs"
		cfg := &config.Config{}
		cfg.Antiflood.NumConcurrentResolverJobs = int32(50)
		expectedNewValue := int32(37)

		err := AdaptStructureValueBasedOnPath(cfg, path, fmt.Sprintf("%d", expectedNewValue))
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.Antiflood.NumConcurrentResolverJobs)
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

		err := AdaptStructureValueBasedOnPath(cfg, path, fmt.Sprintf("%d", expectedNewValue))
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
