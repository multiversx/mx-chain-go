package common_test

import (
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadP2PConfig_InvalidFileShouldErr(t *testing.T) {
	t.Parallel()

	conf, err := common.LoadP2PConfig("testFile01")

	assert.Nil(t, conf)
	assert.Error(t, err)
}

func TestLoadP2PConfig_ShouldPass(t *testing.T) {
	t.Parallel()

	_, err := os.Create("testFile02")
	assert.Nil(t, err)

	conf, err := common.LoadP2PConfig("testFile02")
	if _, errF := os.Stat("testFile02"); errF == nil {
		_ = os.Remove("testFile02")
	}

	assert.NotNil(t, conf)
	assert.Nil(t, err)
}

func TestLoadGasScheduleConfig(t *testing.T) {
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

	file, err := os.Create("testGasSchedule.toml")
	assert.Nil(t, err)

	_, _ = file.WriteString(testString)
	_ = file.Close()

	gasSchedule, err := common.LoadGasScheduleConfig("testGasSchedule.toml")
	assert.Nil(t, err)

	if _, err = os.Stat("testGasSchedule.toml"); err == nil {
		_ = os.Remove("testGasSchedule.toml")
	}

	expectedGasSchedule := make(map[string]uint64)
	expectedGasSchedule["GetFunction"] = 84
	expectedGasSchedule["StorageStore"] = 11
	expectedGasSchedule["BigIntNew"] = 9
	expectedGasSchedule["BigIntAdd"] = 44
	expectedGasSchedule["Drop"] = 1
	expectedGasSchedule["I64Load"] = 12
	expectedGasSchedule["MemoryCopy"] = 14623

	assert.Equal(t, expectedGasSchedule, gasSchedule["GasSchedule"])
}

func TestLoadRoundConfig(t *testing.T) {
	roundConfig, err := common.LoadRoundConfig("invalid path")
	require.Nil(t, roundConfig)
	require.NotNil(t, err)

	roundConfig, err = common.LoadRoundConfig("../cmd/node/config/enableRounds.toml")
	require.True(t, len(roundConfig.EnableRoundsByName) != 0)
	require.Nil(t, err)
}
