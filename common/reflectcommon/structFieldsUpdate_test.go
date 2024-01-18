package reflectcommon

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon/toml"
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

		require.ErrorContains(t, err, "unsupported type <string> when trying to set the value of type <struct>")
	})

	t.Run("should error when setting invalid uint32", func(t *testing.T) {
		t.Parallel()

		path := "TrieSyncStorage.Capacity"
		expectedNewValue := "invalid uint32"
		cfg := &config.Config{}
		cfg.TrieSyncStorage.Capacity = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast value 'invalid uint32' of type <string> to kind <uint32>")
	})

	t.Run("should error when setting invalid uint64", func(t *testing.T) {
		t.Parallel()

		path := "TrieSyncStorage.SizeInBytes"
		expectedNewValue := "invalid uint64"
		cfg := &config.Config{}
		cfg.TrieSyncStorage.SizeInBytes = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast value 'invalid uint64' of type <string> to kind <uint64>")
	})

	t.Run("should error when setting invalid float32", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.MinPeersThreshold"
		expectedNewValue := "invalid float32"
		cfg := &config.Config{}
		cfg.HeartbeatV2.MinPeersThreshold = 37.0

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast value 'invalid float32' of type <string> to kind <float32>")
	})

	t.Run("should error when setting invalid float64", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.PeerShardTimeThresholdBetweenSends"
		expectedNewValue := "invalid float64"
		cfg := &config.Config{}
		cfg.HeartbeatV2.PeerShardTimeThresholdBetweenSends = 37.0

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast value 'invalid float64' of type <string> to kind <float64>")
	})

	t.Run("should error when setting invalid int64", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.HeartbeatExpiryTimespanInSec"
		expectedNewValue := "invalid int64"
		cfg := &config.Config{}
		cfg.HeartbeatV2.HeartbeatExpiryTimespanInSec = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast value 'invalid int64' of type <string> to kind <int64>")
	})

	t.Run("should error when setting invalid int64", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.HeartbeatExpiryTimespanInSec"
		expectedNewValue := "invalid int64"
		cfg := &config.Config{}
		cfg.HeartbeatV2.HeartbeatExpiryTimespanInSec = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast value 'invalid int64' of type <string> to kind <int64>")
	})

	t.Run("should error when setting invalid int", func(t *testing.T) {
		t.Parallel()

		path := "Debug.InterceptorResolver.DebugLineExpiration"
		expectedNewValue := "invalid int"
		cfg := &config.Config{}
		cfg.Debug.InterceptorResolver.DebugLineExpiration = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast value 'invalid int' of type <string> to kind <int>")
	})

	t.Run("should error when setting invalid bool", func(t *testing.T) {
		t.Parallel()

		path := "Debug.InterceptorResolver.EnablePrint"
		expectedNewValue := "invalid bool"
		cfg := &config.Config{}
		cfg.Debug.InterceptorResolver.EnablePrint = false

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.ErrorContains(t, err, "cannot cast value 'invalid bool' of type <string> to kind <bool>")
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

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.StoragePruning.FullArchiveNumActivePersisters)
	})

	t.Run("should work and override float32 value", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.MinPeersThreshold"
		expectedNewValue := float32(38.0)
		cfg := &config.Config{}
		cfg.HeartbeatV2.MinPeersThreshold = 37.0

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.HeartbeatV2.MinPeersThreshold)
	})

	t.Run("should work and override float64 value", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.PeerAuthenticationTimeThresholdBetweenSends"
		expectedNewValue := 38.0
		cfg := &config.Config{}
		cfg.HeartbeatV2.PeerAuthenticationTimeThresholdBetweenSends = 37.0

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.HeartbeatV2.PeerAuthenticationTimeThresholdBetweenSends)
	})

	t.Run("should work and override int value", func(t *testing.T) {
		t.Parallel()

		path := "Debug.InterceptorResolver.DebugLineExpiration"
		expectedNewValue := 38
		cfg := &config.Config{}
		cfg.Debug.InterceptorResolver.DebugLineExpiration = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.Debug.InterceptorResolver.DebugLineExpiration)
	})

	t.Run("should work and override int64 value", func(t *testing.T) {
		t.Parallel()

		path := "Hardfork.GenesisTime"
		expectedNewValue := int64(38)
		cfg := &config.Config{}
		cfg.Hardfork.GenesisTime = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.Hardfork.GenesisTime)
	})

	t.Run("should work and override uint64 value", func(t *testing.T) {
		t.Parallel()

		path := "TrieSyncStorage.SizeInBytes"
		expectedNewValue := uint64(38)
		cfg := &config.Config{}
		cfg.TrieSyncStorage.SizeInBytes = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.TrieSyncStorage.SizeInBytes)
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

	t.Run("should work and override int32 value", func(t *testing.T) {
		t.Parallel()

		path := "Antiflood.NumConcurrentResolverJobs"
		cfg := &config.Config{}
		cfg.Antiflood.NumConcurrentResolverJobs = int32(50)
		expectedNewValue := int32(37)

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)
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

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)
		require.NoError(t, err)

		require.Equal(t, expectedNewValue, cfg.Hardfork.ExportKeysStorageConfig.DB.MaxBatchSize)
	})

	t.Run("should work and override int8 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI8.Int8.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[0].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[0].Value, int64(testConfig.Int8.Value))
	})

	t.Run("should error int8 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI8.Int8.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[1].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(int64=128)' does not fit within the range of <int8>")
	})

	t.Run("should work and override int8 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI8.Int8.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[2].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[2].Value, int64(testConfig.Int8.Value))
	})

	t.Run("should error int8 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI8.Int8.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[3].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(int64=-129)' does not fit within the range of <int8>")
	})

	t.Run("should work and override int16 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI16.Int16.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[4].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[4].Value, int64(testConfig.Int16.Value))
	})

	t.Run("should error int16 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI16.Int16.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[5].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(int64=32768)' does not fit within the range of <int16>")
	})

	t.Run("should work and override int16 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI16.Int16.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[6].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[6].Value, int64(testConfig.Int16.Value))
	})

	t.Run("should error int16 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI16.Int16.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[7].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(int64=-32769)' does not fit within the range of <int16>")
	})

	t.Run("should work and override int32 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI32.Int32.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[8].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[8].Value, int64(testConfig.Int32.Value))
	})

	t.Run("should error int32 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI32.Int32.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[9].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(int64=2147483648)' does not fit within the range of <int32>")
	})

	t.Run("should work and override int32 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI32.Int32.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[10].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[10].Value, int64(testConfig.Int32.Value))
	})

	t.Run("should error int32 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI32.Int32.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[11].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(int64=-2147483649)' does not fit within the range of <int32>")
	})

	t.Run("should work and override int64 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI64.Int64.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[12].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[12].Value, int64(testConfig.Int64.Value))
	})

	t.Run("should work and override int64 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigI64.Int64.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[13].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[13].Value, int64(testConfig.Int64.Value))
	})

	t.Run("should work and override uint8 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigU8.Uint8.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[14].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[14].Value, int64(testConfig.Uint8.Value))
	})

	t.Run("should error uint8 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigU8.Uint8.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[15].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(int64=256)' does not fit within the range of <uint8>")
	})

	t.Run("should error uint8 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigU8.Uint8.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[16].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(int64=-256)' does not fit within the range of <uint8>")
	})

	t.Run("should work and override uint16 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigU16.Uint16.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[17].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[17].Value, int64(testConfig.Uint16.Value))
	})

	t.Run("should error uint16 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigU16.Uint16.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[18].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(int64=65536)' does not fit within the range of <uint16>")
	})

	t.Run("should error uint16 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigU16.Uint16.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[19].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(int64=-65536)' does not fit within the range of <uint16>")
	})

	t.Run("should work and override uint32 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigU32.Uint32.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[20].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[20].Value, int64(testConfig.Uint32.Value))
	})

	t.Run("should error uint32 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigU32.Uint32.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[21].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(int64=4294967296)' does not fit within the range of <uint32>")
	})

	t.Run("should error uint32 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigU32.Uint32.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[22].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(int64=-4294967296)' does not fit within the range of <uint32>")
	})

	t.Run("should work and override uint64 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigU64.Uint64.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[23].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[23].Value, int64(testConfig.Uint64.Value))
	})

	t.Run("should error uint64 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigU64.Uint64.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[24].Value)
		require.ErrorContains(t, err, "value '%!s(int64=-9223372036854775808)' does not fit within the range of <uint64>")
	})

	t.Run("should work and override float32 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigF32.Float32.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[25].Value)
		require.NoError(t, err)
		require.Equal(t, testConfig.Float32.Value, float32(3.4))
	})

	t.Run("should error float32 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigF32.Float32.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[26].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(float64=3.4e+39)' does not fit within the range of <float32>")
	})

	t.Run("should work and override float32 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigF32.Float32.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[27].Value)
		require.NoError(t, err)
		require.Equal(t, testConfig.Float32.Value, float32(-3.4))
	})

	t.Run("should error float32 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigF32.Float32.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[28].Value)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "value '%!s(float64=-3.4e+40)' does not fit within the range of <float32>")
	})

	t.Run("should work and override float64 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigF64.Float64.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[29].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[29].Value, testConfig.Float64.Value)
	})

	t.Run("should work and override float64 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigF64.Float64.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[30].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[30].Value, testConfig.Float64.Value)
	})

	t.Run("should work and override struct", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigStruct.ConfigStruct.Description"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[31].Value)
		require.NoError(t, err)
		require.Equal(t, testConfig.TestConfigStruct.ConfigStruct.Description.Number, uint32(11))
	})

	t.Run("should work and override nested struct", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		path := "TestConfigNestedStruct.ConfigNestedStruct"

		err = AdaptStructureValueBasedOnPath(testConfig, path, overrideConfig.OverridableConfigTomlValues[32].Value)
		require.NoError(t, err)
		require.Equal(t, testConfig.TestConfigNestedStruct.ConfigNestedStruct.Text, "Overwritten text")
		require.Equal(t, testConfig.TestConfigNestedStruct.ConfigNestedStruct.Message.Public, false)
		require.Equal(t, testConfig.TestConfigNestedStruct.ConfigNestedStruct.Message.MessageDescription[0].Text, "Overwritten Text1")
	})

}

func loadTestConfig(filepath string) (*toml.Config, error) {
	cfg := &toml.Config{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
func loadOverrideConfig(filepath string) (*toml.OverrideConfig, error) {
	cfg := &toml.OverrideConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
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
