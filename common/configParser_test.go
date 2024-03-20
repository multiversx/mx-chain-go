package common_test

import (
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadP2PConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid file should error", func(t *testing.T) {
		t.Parallel()

		conf, err := common.LoadP2PConfig("testFile01")
		assert.Nil(t, conf)
		assert.Error(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testString := `
[Node]
    Port = "1234"
[KadDhtPeerDiscovery]
    Enabled = true
`

		filePath := path.Join(t.TempDir(), "testFile02")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString(testString)
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		conf, err := common.LoadP2PConfig(filePath)
		assert.NotNil(t, conf)
		assert.Nil(t, err)
		assert.True(t, conf.KadDhtPeerDiscovery.Enabled)
		assert.Equal(t, "1234", conf.Node.Port)
	})
}

func TestLoadMainConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid file should error", func(t *testing.T) {
		t.Parallel()

		conf, err := common.LoadMainConfig("testFile01")
		assert.Nil(t, conf)
		assert.Error(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testString := `
[GeneralSettings]
    ChainID = "1"
[Versions]
    DefaultVersion = "default"
[MiniBlocksStorage]
    [MiniBlocksStorage.Cache]
        Name = "MiniBlocksStorage"
`

		filePath := path.Join(t.TempDir(), "testFile02")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString(testString)
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		conf, err := common.LoadMainConfig(filePath)
		assert.NotNil(t, conf)
		assert.Nil(t, err)
		assert.Equal(t, "1", conf.GeneralSettings.ChainID)
		assert.Equal(t, "default", conf.Versions.DefaultVersion)
		assert.Equal(t, "MiniBlocksStorage", conf.MiniBlocksStorage.Cache.Name)
	})
}

func TestLoadApiConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid file should error", func(t *testing.T) {
		t.Parallel()

		conf, err := common.LoadApiConfig("testFile01")
		assert.Nil(t, conf)
		assert.Error(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testString := `
[Logging]
    LoggingEnabled = true
[APIPackages]
[APIPackages.node]
    Routes = [
        { Name = "/status", Open = true },
	]
`

		filePath := path.Join(t.TempDir(), "testFile02")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString(testString)
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		conf, err := common.LoadApiConfig(filePath)
		assert.NotNil(t, conf)
		assert.Nil(t, err)
		assert.True(t, conf.Logging.LoggingEnabled)
		nodePackage, found := conf.APIPackages["node"]
		assert.True(t, found)
		assert.Equal(t, "/status", nodePackage.Routes[0].Name)
	})
}

func TestLoadEconomicsConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid file should error", func(t *testing.T) {
		t.Parallel()

		conf, err := common.LoadEconomicsConfig("testFile01")
		assert.Nil(t, conf)
		assert.Error(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testString := `
[GlobalSettings]
    GenesisTotalSupply = "20000000000000000000000000" #20MIL eGLD
`

		filePath := path.Join(t.TempDir(), "testFile02")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString(testString)
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		conf, err := common.LoadEconomicsConfig(filePath)
		assert.NotNil(t, conf)
		assert.Nil(t, err)
		assert.Equal(t, "20000000000000000000000000", conf.GlobalSettings.GenesisTotalSupply)
	})
}

func TestLoadSystemSmartContractsConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid file should error", func(t *testing.T) {
		t.Parallel()

		conf, err := common.LoadSystemSmartContractsConfig("testFile01")

		assert.Nil(t, conf)
		assert.Error(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testString := `
[StakingSystemSCConfig]
    GenesisNodePrice = "2500000000000000000000" #2.5K eGLD
`

		filePath := path.Join(t.TempDir(), "testFile02")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString(testString)
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		conf, err := common.LoadSystemSmartContractsConfig(filePath)
		assert.NotNil(t, conf)
		assert.Nil(t, err)
		assert.Equal(t, "2500000000000000000000", conf.StakingSystemSCConfig.GenesisNodePrice)
	})
}

func TestLoadRatingsConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid file should error", func(t *testing.T) {
		t.Parallel()

		conf, err := common.LoadRatingsConfig("testFile01")
		assert.Equal(t, &config.RatingsConfig{}, conf)
		assert.Error(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testString := `
[General]
    MinRating = 1
    SelectionChances = [
        { MaxThreshold = 1, ChancePercent = 5},
	]
`

		filePath := path.Join(t.TempDir(), "testFile02")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString(testString)
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		conf, err := common.LoadRatingsConfig(filePath)
		assert.NotNil(t, conf)
		assert.Nil(t, err)
		assert.Equal(t, uint32(1), conf.General.MinRating)
		assert.Equal(t, uint32(1), conf.General.SelectionChances[0].MaxThreshold)
		assert.Equal(t, uint32(5), conf.General.SelectionChances[0].ChancePercent)
	})
}

func TestLoadPreferencesConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid file should error", func(t *testing.T) {
		t.Parallel()

		conf, err := common.LoadPreferencesConfig("testFile01")
		assert.Nil(t, conf)
		assert.Error(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testString := `
[Preferences]
   DestinationShardAsObserver = "metachain"
`

		filePath := path.Join(t.TempDir(), "testFile02")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString(testString)
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		conf, err := common.LoadPreferencesConfig(filePath)
		assert.NotNil(t, conf)
		assert.Nil(t, err)
		assert.Equal(t, "metachain", conf.Preferences.DestinationShardAsObserver)
	})
}

func TestLoadExternalConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid file should error", func(t *testing.T) {
		t.Parallel()

		conf, err := common.LoadExternalConfig("testFile01")
		assert.Nil(t, conf)
		assert.Error(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testString := `
[ElasticSearchConnector]
    Enabled = true
`

		filePath := path.Join(t.TempDir(), "testFile02")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString(testString)
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		conf, err := common.LoadExternalConfig(filePath)
		assert.NotNil(t, conf)
		assert.Nil(t, err)
		assert.True(t, conf.ElasticSearchConnector.Enabled)
	})
}

func TestLoadGasScheduleConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid file should error", func(t *testing.T) {
		t.Parallel()

		conf, err := common.LoadGasScheduleConfig("testFile01")
		assert.Nil(t, conf)
		assert.Error(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testString := `
[GasSchedule]
    GetFunction = 84
    StorageStore = 11
    BigIntNew = 9
    BigIntAdd = 44
    Drop = 1
    I64Load = 12
    MemoryCopy = 14623
`

		filePath := path.Join(t.TempDir(), "testGasSchedule.toml")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString(testString)
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		gasSchedule, err := common.LoadGasScheduleConfig(filePath)
		assert.Nil(t, err)

		expectedGasSchedule := make(map[string]uint64)
		expectedGasSchedule["GetFunction"] = 84
		expectedGasSchedule["StorageStore"] = 11
		expectedGasSchedule["BigIntNew"] = 9
		expectedGasSchedule["BigIntAdd"] = 44
		expectedGasSchedule["Drop"] = 1
		expectedGasSchedule["I64Load"] = 12
		expectedGasSchedule["MemoryCopy"] = 14623

		assert.Equal(t, expectedGasSchedule, gasSchedule["GasSchedule"])
	})
}

func TestLoadRoundConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid file should error", func(t *testing.T) {
		t.Parallel()

		conf, err := common.LoadRoundConfig("testFile01")
		assert.Nil(t, conf)
		assert.Error(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testString := ""

		filePath := path.Join(t.TempDir(), "testEnableRounds.toml")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString(testString)
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		roundActivationConfig, err := common.LoadRoundConfig(filePath)
		assert.Nil(t, err)

		require.Equal(t, &config.RoundConfig{}, roundActivationConfig)
	})
}

func TestLoadEpochConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid file should error", func(t *testing.T) {
		t.Parallel()

		conf, err := common.LoadEpochConfig("testFile01")
		assert.Nil(t, conf)
		assert.Error(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testString := `
[EnableEpochs]
    SCDeployEnableEpoch = 1
`

		filePath := path.Join(t.TempDir(), "testFile02")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString(testString)
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		conf, err := common.LoadEpochConfig(filePath)
		assert.NotNil(t, conf)
		assert.Nil(t, err)
		require.Equal(t, uint32(1), conf.EnableEpochs.SCDeployEnableEpoch)
	})
}

func TestGetSkBytesFromP2pKey(t *testing.T) {
	t.Parallel()

	t.Run("empty file name should early return", func(t *testing.T) {
		t.Parallel()

		key, err := common.GetSkBytesFromP2pKey("")
		assert.Equal(t, []byte{}, key)
		assert.NoError(t, err)
	})
	t.Run("invalid file should return nil", func(t *testing.T) {
		t.Parallel()

		key, err := common.GetSkBytesFromP2pKey("testFile01")
		assert.Equal(t, []byte{}, key)
		assert.NoError(t, err)
	})
	t.Run("invalid file content should error", func(t *testing.T) {
		t.Parallel()

		filePath := path.Join(t.TempDir(), "testFile01")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString("invalid pem file")
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		key, err := common.GetSkBytesFromP2pKey(filePath)
		assert.Nil(t, key)
		assert.Error(t, err)
	})
	t.Run("invalid secret key should error", func(t *testing.T) {
		t.Parallel()

		invalidPem := `
-----BEGIN PRIVATE KEY for erd1ecdwux5tvanwryhr7cn5l9kc07ayquec7h2jc608kz0ycychzexsj6qw4j-----
INVALIDiZTI4ODY1ZGQ1MzZmOWYyOGEwOTZhNmY5NmQ2MzZjMjhmNzMzNDAwNDBj
MTU0ZDA4Nzg3MTBhOTE5ZWNlMWFlZTFhOGI2NzY2ZTE5MmUzZjYyNzRmOTZkODdm
YmE0MDczMzhmNWQ1MmM2OWU3YjA5ZTRjMTMxNzE2NGQ=
-----END PRIVATE KEY for erd1ecdwux5tvanwryhr7cn5l9kc07ayquec7h2jc608kz0ycychzexsj6qw4j-----
`
		filePath := path.Join(t.TempDir(), "testFile01")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString(invalidPem)
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		key, err := common.GetSkBytesFromP2pKey(filePath)
		assert.Nil(t, key)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "for encoded secret key"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		validPem := `
-----BEGIN PRIVATE KEY for erd1ecdwux5tvanwryhr7cn5l9kc07ayquec7h2jc608kz0ycychzexsj6qw4j-----
NDliYzliZTI4ODY1ZGQ1MzZmOWYyOGEwOTZhNmY5NmQ2MzZjMjhmNzMzNDAwNDBj
MTU0ZDA4Nzg3MTBhOTE5ZWNlMWFlZTFhOGI2NzY2ZTE5MmUzZjYyNzRmOTZkODdm
YmE0MDczMzhmNWQ1MmM2OWU3YjA5ZTRjMTMxNzE2NGQ=
-----END PRIVATE KEY for erd1ecdwux5tvanwryhr7cn5l9kc07ayquec7h2jc608kz0ycychzexsj6qw4j-----
`
		filePath := path.Join(t.TempDir(), "testFile01")
		file, err := os.Create(filePath)
		assert.Nil(t, err)

		_, err = file.WriteString(validPem)
		assert.Nil(t, err)

		assert.Nil(t, file.Close())

		key, err := common.GetSkBytesFromP2pKey(filePath)
		assert.NotNil(t, key)
		assert.NoError(t, err)
		assert.Equal(t, 64, len(key))
	})
}

func TestGetNodeProcessingMode(t *testing.T) {
	t.Parallel()

	mode := common.GetNodeProcessingMode(&config.ImportDbConfig{
		IsImportDBMode: true,
	})
	assert.Equal(t, common.ImportDb, mode)

	mode = common.GetNodeProcessingMode(&config.ImportDbConfig{})
	assert.Equal(t, common.Normal, mode)
}

func TestLatestVersionOfGasScheduleIsUsed(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir("../cmd/node/config/gasSchedules")
	if err != nil {
		log.Fatal(err)
	}

	assert.True(t, len(entries) > 0, "No gas schedule files found")

	sort.Slice(entries, func(i, j int) bool {
		return common.GasScheduleSortName(entries[i].Name()) > common.GasScheduleSortName(entries[j].Name())
	})

	assert.Equal(t, common.LatestGasScheduleFileName, entries[0].Name(), "Latest gas schedule file is not used. Please update common.LatestGasScheduleFileName and enableEpochs.toml.")
}
