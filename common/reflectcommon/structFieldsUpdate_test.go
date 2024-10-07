package reflectcommon

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon/toml"

	"github.com/multiversx/mx-chain-core-go/core"
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

	t.Run("should error when setting unsupported type on struct", func(t *testing.T) {
		t.Parallel()

		path := "TrieSyncStorage.DB"
		expectedNewValue := "provided value"
		cfg := &config.Config{}

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.Equal(t, "unsupported type <string> when trying to set the value of type <struct>", err.Error())
	})

	t.Run("should error when setting invalid type on struct", func(t *testing.T) {
		t.Parallel()

		path := "TrieSyncStorage.DB"
		cfg := &config.Config{}

		err := AdaptStructureValueBasedOnPath(cfg, path, nil)

		require.Equal(t, "invalid new value kind", err.Error())
	})

	t.Run("should error when setting invalid uint32", func(t *testing.T) {
		t.Parallel()

		path := "TrieSyncStorage.Capacity"
		expectedNewValue := "invalid uint32"
		cfg := &config.Config{}
		cfg.TrieSyncStorage.Capacity = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.Equal(t, "unable to cast value 'invalid uint32' of type <string> to type <uint32>", err.Error())
	})

	t.Run("should error when setting invalid uint64", func(t *testing.T) {
		t.Parallel()

		path := "TrieSyncStorage.SizeInBytes"
		expectedNewValue := "invalid uint64"
		cfg := &config.Config{}
		cfg.TrieSyncStorage.SizeInBytes = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.Equal(t, "unable to cast value 'invalid uint64' of type <string> to type <uint64>", err.Error())
	})

	t.Run("should error when setting invalid float32", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.MinPeersThreshold"
		expectedNewValue := "invalid float32"
		cfg := &config.Config{}
		cfg.HeartbeatV2.MinPeersThreshold = 37.0

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.Equal(t, "unable to cast value 'invalid float32' of type <string> to type <float32>", err.Error())
	})

	t.Run("should error when setting invalid float64", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.PeerShardTimeThresholdBetweenSends"
		expectedNewValue := "invalid float64"
		cfg := &config.Config{}
		cfg.HeartbeatV2.PeerShardTimeThresholdBetweenSends = 37.0

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.Equal(t, "unable to cast value 'invalid float64' of type <string> to type <float64>", err.Error())
	})

	t.Run("should error when setting invalid int64", func(t *testing.T) {
		t.Parallel()

		path := "HeartbeatV2.HeartbeatExpiryTimespanInSec"
		expectedNewValue := "invalid int64"
		cfg := &config.Config{}
		cfg.HeartbeatV2.HeartbeatExpiryTimespanInSec = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.Equal(t, "unable to cast value 'invalid int64' of type <string> to type <int64>", err.Error())
	})

	t.Run("should error when setting invalid int", func(t *testing.T) {
		t.Parallel()

		path := "Debug.InterceptorResolver.DebugLineExpiration"
		expectedNewValue := "invalid int"
		cfg := &config.Config{}
		cfg.Debug.InterceptorResolver.DebugLineExpiration = 37

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.Equal(t, "unable to cast value 'invalid int' of type <string> to type <int>", err.Error())
	})

	t.Run("should error when setting invalid bool", func(t *testing.T) {
		t.Parallel()

		path := "Debug.InterceptorResolver.EnablePrint"
		expectedNewValue := "invalid bool"
		cfg := &config.Config{}
		cfg.Debug.InterceptorResolver.EnablePrint = false

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)

		require.Equal(t, "unable to cast value 'invalid bool' of type <string> to type <bool>", err.Error())
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

	t.Run("should error if setting int into string", func(t *testing.T) {
		t.Parallel()

		path := "GeneralSettings.ChainID"
		cfg := &config.Config{}
		cfg.GeneralSettings.ChainID = "D"
		expectedNewValue := 1

		err := AdaptStructureValueBasedOnPath(cfg, path, expectedNewValue)
		require.Equal(t, "unable to cast value '1' of type <int> to type <string>", err.Error())
	})

	t.Run("should error for unsupported type", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		expectedNewValue := make(map[string]int)
		expectedNewValue["first"] = 1
		expectedNewValue["second"] = 2

		path := "TestInterface.Value"

		err = AdaptStructureValueBasedOnPath(testConfig, path, expectedNewValue)
		require.Equal(t, "unsupported type <interface> when trying to set the value 'map[first:1 second:2]' of type <map[string]int>", err.Error())
	})

	t.Run("should error fit signed for target type not int", func(t *testing.T) {
		t.Parallel()

		expectedNewValue := 10
		reflectNewValue := reflect.ValueOf(expectedNewValue)
		targetType := reflect.TypeOf("string")

		res := FitsWithinSignedIntegerRange(reflectNewValue, targetType)
		require.False(t, res)
	})

	t.Run("should error fit signed for value not int and target type int", func(t *testing.T) {
		t.Parallel()

		expectedNewValue := "value"
		reflectNewValue := reflect.ValueOf(expectedNewValue)
		targetType := reflect.TypeOf(10)

		res := FitsWithinSignedIntegerRange(reflectNewValue, targetType)
		require.False(t, res)
	})

	t.Run("should error fit unsigned for target type not uint", func(t *testing.T) {
		t.Parallel()

		expectedNewValue := uint(10)
		reflectNewValue := reflect.ValueOf(expectedNewValue)
		targetType := reflect.TypeOf("string")

		res := FitsWithinUnsignedIntegerRange(reflectNewValue, targetType)
		require.False(t, res)
	})

	t.Run("should error fit unsigned for value not uint and target type uint", func(t *testing.T) {
		t.Parallel()

		expectedNewValue := "value"
		reflectNewValue := reflect.ValueOf(expectedNewValue)
		targetType := reflect.TypeOf(uint(10))

		res := FitsWithinUnsignedIntegerRange(reflectNewValue, targetType)
		require.False(t, res)
	})

	t.Run("should error fit float for target type not float", func(t *testing.T) {
		t.Parallel()

		expectedNewValue := float32(10)
		reflectNewValue := reflect.ValueOf(expectedNewValue)
		targetType := reflect.TypeOf("string")

		res := FitsWithinFloatRange(reflectNewValue, targetType)
		require.False(t, res)
	})

	t.Run("should work and override int8 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[0].Path, overrideConfig.OverridableConfigTomlValues[0].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[0].Value, int64(testConfig.Int8.Number))
	})

	t.Run("should error int8 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[1].Path, overrideConfig.OverridableConfigTomlValues[1].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '128' of type <int64> to type <int8>", err.Error())
	})

	t.Run("should work and override int8 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[2].Path, overrideConfig.OverridableConfigTomlValues[2].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[2].Value, int64(testConfig.Int8.Number))
	})

	t.Run("should error int8 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[3].Path, overrideConfig.OverridableConfigTomlValues[3].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '-129' of type <int64> to type <int8>", err.Error())
	})

	t.Run("should work and override int16 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[4].Path, overrideConfig.OverridableConfigTomlValues[4].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[4].Value, int64(testConfig.Int16.Number))
	})

	t.Run("should error int16 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[5].Path, overrideConfig.OverridableConfigTomlValues[5].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '32768' of type <int64> to type <int16>", err.Error())
	})

	t.Run("should work and override int16 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[6].Path, overrideConfig.OverridableConfigTomlValues[6].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[6].Value, int64(testConfig.Int16.Number))
	})

	t.Run("should error int16 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[7].Path, overrideConfig.OverridableConfigTomlValues[7].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '-32769' of type <int64> to type <int16>", err.Error())
	})

	t.Run("should work and override int32 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[8].Path, overrideConfig.OverridableConfigTomlValues[8].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[8].Value, int64(testConfig.Int32.Number))
	})

	t.Run("should work and override int32 value with uint16", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		expectedNewValue := uint16(10)

		path := "TestConfigI32.Int32.Number"

		err = AdaptStructureValueBasedOnPath(testConfig, path, expectedNewValue)
		require.NoError(t, err)
		require.Equal(t, int32(expectedNewValue), testConfig.Int32.Number)
	})

	t.Run("should error int32 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[9].Path, overrideConfig.OverridableConfigTomlValues[9].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '2147483648' of type <int64> to type <int32>", err.Error())
	})

	t.Run("should work and override int32 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[10].Path, overrideConfig.OverridableConfigTomlValues[10].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[10].Value, int64(testConfig.Int32.Number))
	})

	t.Run("should error int32 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[11].Path, overrideConfig.OverridableConfigTomlValues[11].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '-2147483649' of type <int64> to type <int32>", err.Error())
	})

	t.Run("should work and override int64 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[12].Path, overrideConfig.OverridableConfigTomlValues[12].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[12].Value, int64(testConfig.Int64.Number))
	})

	t.Run("should work and override int64 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[13].Path, overrideConfig.OverridableConfigTomlValues[13].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[13].Value, int64(testConfig.Int64.Number))
	})

	t.Run("should work and override uint8 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[14].Path, overrideConfig.OverridableConfigTomlValues[14].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[14].Value, int64(testConfig.Uint8.Number))
	})

	t.Run("should error uint8 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[15].Path, overrideConfig.OverridableConfigTomlValues[15].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '256' of type <int64> to type <uint8>", err.Error())
	})

	t.Run("should error uint8 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[16].Path, overrideConfig.OverridableConfigTomlValues[16].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '-256' of type <int64> to type <uint8>", err.Error())
	})

	t.Run("should work and override uint16 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[17].Path, overrideConfig.OverridableConfigTomlValues[17].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[17].Value, int64(testConfig.Uint16.Number))
	})

	t.Run("should error uint16 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[18].Path, overrideConfig.OverridableConfigTomlValues[18].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '65536' of type <int64> to type <uint16>", err.Error())
	})

	t.Run("should error uint16 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[19].Path, overrideConfig.OverridableConfigTomlValues[19].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '-65536' of type <int64> to type <uint16>", err.Error())
	})

	t.Run("should work and override uint32 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[20].Path, overrideConfig.OverridableConfigTomlValues[20].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[20].Value, int64(testConfig.Uint32.Number))
	})

	t.Run("should error uint32 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[21].Path, overrideConfig.OverridableConfigTomlValues[21].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '4294967296' of type <int64> to type <uint32>", err.Error())
	})

	t.Run("should error uint32 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[22].Path, overrideConfig.OverridableConfigTomlValues[22].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '-4294967296' of type <int64> to type <uint32>", err.Error())
	})

	t.Run("should work and override uint64 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[23].Path, overrideConfig.OverridableConfigTomlValues[23].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[23].Value, int64(testConfig.Uint64.Number))
	})

	t.Run("should error uint64 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[24].Path, overrideConfig.OverridableConfigTomlValues[24].Value)
		require.Equal(t, "unable to cast value '-9223372036854775808' of type <int64> to type <uint64>", err.Error())
	})

	t.Run("should work and override float32 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[25].Path, overrideConfig.OverridableConfigTomlValues[25].Value)
		require.NoError(t, err)
		require.Equal(t, float32(3.4), testConfig.Float32.Number)
	})

	t.Run("should error float32 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[26].Path, overrideConfig.OverridableConfigTomlValues[26].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '3.4e+39' of type <float64> to type <float32>", err.Error())
	})

	t.Run("should work and override float32 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[27].Path, overrideConfig.OverridableConfigTomlValues[27].Value)
		require.NoError(t, err)
		require.Equal(t, float32(-3.4), testConfig.Float32.Number)
	})

	t.Run("should error float32 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[28].Path, overrideConfig.OverridableConfigTomlValues[28].Value)
		require.NotNil(t, err)
		require.Equal(t, "unable to cast value '-3.4e+40' of type <float64> to type <float32>", err.Error())
	})

	t.Run("should work and override float64 value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[29].Path, overrideConfig.OverridableConfigTomlValues[29].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[29].Value, testConfig.Float64.Number)
	})

	t.Run("should work and override float64 negative value", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[30].Path, overrideConfig.OverridableConfigTomlValues[30].Value)
		require.NoError(t, err)
		require.Equal(t, overrideConfig.OverridableConfigTomlValues[30].Value, testConfig.Float64.Number)
	})

	t.Run("should work and override struct", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		expectedNewValue := toml.Description{
			Number: 11,
		}

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[31].Path, overrideConfig.OverridableConfigTomlValues[31].Value)
		require.NoError(t, err)
		require.Equal(t, expectedNewValue, testConfig.TestConfigStruct.ConfigStruct.Description)
	})

	t.Run("should error with field not found", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[32].Path, overrideConfig.OverridableConfigTomlValues[32].Value)
		require.Equal(t, "field <Nr> not found or cannot be set", err.Error())
	})

	t.Run("should error with different types", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[33].Path, overrideConfig.OverridableConfigTomlValues[33].Value)
		require.Equal(t, "unable to cast value '11' of type <string> to type <uint32>", err.Error())
	})

	t.Run("should work and override nested struct", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		expectedNewValue := toml.ConfigNestedStruct{
			Text: "Overwritten text",
			Message: toml.Message{
				Public: false,
				MessageDescription: []toml.MessageDescription{
					{Text: "Overwritten Text1"},
				},
			},
		}

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[34].Path, overrideConfig.OverridableConfigTomlValues[34].Value)
		require.NoError(t, err)
		require.Equal(t, expectedNewValue, testConfig.TestConfigNestedStruct.ConfigNestedStruct)
	})

	t.Run("should work on slice and override map", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		expectedNewValue := []toml.MessageDescription{
			{Text: "Overwritten Text1"},
			{Text: "Overwritten Text2"},
		}

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[35].Path, overrideConfig.OverridableConfigTomlValues[35].Value)
		require.NoError(t, err)
		require.Equal(t, expectedNewValue, testConfig.TestConfigNestedStruct.ConfigNestedStruct.Message.MessageDescription)
	})

	t.Run("should error on slice when override int", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		path := "TestConfigNestedStruct.ConfigNestedStruct.Message.MessageDescription"

		err = AdaptStructureValueBasedOnPath(testConfig, path, 10)
		require.Equal(t, "reflect: call of reflect.Value.Len on int Value", err.Error())
	})

	t.Run("should error on slice when override different type", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		path := "TestConfigNestedStruct.ConfigNestedStruct.Message.MessageDescription"

		expectedNewValue := []int{10, 20}

		err = AdaptStructureValueBasedOnPath(testConfig, path, expectedNewValue)
		require.Equal(t, "unsupported type <int> when trying to set the value of type <struct>", err.Error())
	})

	t.Run("should error on slice when override different struct", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		path := "TestConfigNestedStruct.ConfigNestedStruct.Message.MessageDescription"

		var expectedNewValue = []toml.MessageDescriptionOtherName{
			{Value: "10"},
			{Value: "20"},
		}

		err = AdaptStructureValueBasedOnPath(testConfig, path, expectedNewValue)
		require.Equal(t, "field <Value> not found or cannot be set", err.Error())
	})

	t.Run("should error on slice when override different struct types", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		path := "TestConfigNestedStruct.ConfigNestedStruct.Message.MessageDescription"

		var expectedNewValue = []toml.MessageDescriptionOtherType{
			{Text: 10},
			{Text: 20},
		}

		err = AdaptStructureValueBasedOnPath(testConfig, path, expectedNewValue)
		require.Equal(t, "unable to cast value '10' of type <int> to type <string>", err.Error())
	})

	t.Run("should work on slice and override struct", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		path := "TestConfigNestedStruct.ConfigNestedStruct.Message.MessageDescription"

		var expectedNewValue = []toml.MessageDescription{
			{Text: "Text 1"},
			{Text: "Text 2"},
		}

		err = AdaptStructureValueBasedOnPath(testConfig, path, expectedNewValue)
		require.NoError(t, err)
		require.Equal(t, expectedNewValue, testConfig.TestConfigNestedStruct.ConfigNestedStruct.Message.MessageDescription)
	})

	t.Run("should work on map, override and insert from config", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[36].Path, overrideConfig.OverridableConfigTomlValues[36].Value)
		require.NoError(t, err)
		require.Equal(t, 2, len(testConfig.TestMap.Map))
		require.Equal(t, 10, testConfig.TestMap.Map["Key1"].Number)
		require.Equal(t, 11, testConfig.TestMap.Map["Key2"].Number)
	})

	t.Run("should work on map and insert from config", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		overrideConfig, err := loadOverrideConfig("../../testscommon/toml/overwrite.toml")
		require.NoError(t, err)

		err = AdaptStructureValueBasedOnPath(testConfig, overrideConfig.OverridableConfigTomlValues[37].Path, overrideConfig.OverridableConfigTomlValues[37].Value)
		require.NoError(t, err)
		require.Equal(t, 3, len(testConfig.TestMap.Map))
		require.Equal(t, 999, testConfig.TestMap.Map["Key1"].Number)
		require.Equal(t, 2, testConfig.TestMap.Map["Key2"].Number)
		require.Equal(t, 3, testConfig.TestMap.Map["Key3"].Number)
	})

	t.Run("should work on map, override and insert values in map", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		expectedNewValue := make(map[string]toml.MapValues)
		expectedNewValue["Key1"] = toml.MapValues{Number: 100}
		expectedNewValue["Key2"] = toml.MapValues{Number: 200}

		path := "TestMap.Map"

		err = AdaptStructureValueBasedOnPath(testConfig, path, expectedNewValue)
		require.NoError(t, err)
		require.Equal(t, 2, len(testConfig.TestMap.Map))
		require.Equal(t, 100, testConfig.TestMap.Map["Key1"].Number)
		require.Equal(t, 200, testConfig.TestMap.Map["Key2"].Number)
	})

	t.Run("should error on map when override different map", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		expectedNewValue := make(map[string]toml.MessageDescription)
		expectedNewValue["Key1"] = toml.MessageDescription{Text: "A"}
		expectedNewValue["Key2"] = toml.MessageDescription{Text: "B"}

		path := "TestMap.Map"

		err = AdaptStructureValueBasedOnPath(testConfig, path, expectedNewValue)
		require.Equal(t, "field <Text> not found or cannot be set", err.Error())
	})

	t.Run("should error on map when override anything else other than map", func(t *testing.T) {
		t.Parallel()

		testConfig, err := loadTestConfig("../../testscommon/toml/config.toml")
		require.NoError(t, err)

		expectedNewValue := 1

		path := "TestMap.Map"

		err = AdaptStructureValueBasedOnPath(testConfig, path, expectedNewValue)
		require.Equal(t, "unsupported type <int> when trying to add value in type <map>", err.Error())
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
