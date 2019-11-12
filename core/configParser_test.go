package core_test

import (
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestLoadP2PConfig_InvalidFileShouldErr(t *testing.T) {
	t.Parallel()

	conf, err := core.LoadP2PConfig("testFile01")

	assert.Nil(t, conf)
	assert.Error(t, err)
}

func TestLoadP2PConfig_ShouldPass(t *testing.T) {
	t.Parallel()

	_, err := os.Create("testFile02")
	assert.Nil(t, err)

	conf, err := core.LoadP2PConfig("testFile02")
	if _, errF := os.Stat("testFile02"); errF == nil {
		_ = os.Remove("testFile02")
	}

	assert.NotNil(t, conf)
	assert.Nil(t, err)
}

func TestLoadServersPConfig_InvalidFileShouldErr(t *testing.T) {
	t.Parallel()

	conf, err := core.LoadServersPConfig("testFile03")

	assert.Nil(t, conf)
	assert.Error(t, err)
}

func TestLoadServersPConfig_ShouldPass(t *testing.T) {
	t.Parallel()

	_, err := os.Create("testFile04")
	assert.Nil(t, err)

	conf, err := core.LoadServersPConfig("testFile04")
	if _, errF := os.Stat("testFile04"); errF == nil {
		_ = os.Remove("testFile04")
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

	file.WriteString(testString)
	file.Close()

	gasSchedule, err := core.LoadGasScheduleConfig("testGasSchedule.toml")
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

	assert.Equal(t, expectedGasSchedule, gasSchedule)
}
