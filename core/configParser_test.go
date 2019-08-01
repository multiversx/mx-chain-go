package core_test

import (
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestLoadP2PConfig_InvalidFileShouldErr(t *testing.T) {
	t.Parallel()

	conf, err := core.LoadP2PConfig("testFile")

	assert.Nil(t, conf)
	assert.Error(t, err)
}

func TestLoadP2PConfig_ShouldPass(t *testing.T) {
	t.Parallel()

	_, err := os.Create("testFile")

	conf, err := core.LoadP2PConfig("testFile")

	if _, errF := os.Stat("testFile"); errF == nil {
		_ = os.Remove("testFile")
	}

	assert.NotNil(t, conf)
	assert.Nil(t, err)
}

func TestLoadServersPConfig_InvalidFileShouldErr(t *testing.T) {
	t.Parallel()

	conf, err := core.LoadServersPConfig("testFile")

	assert.Nil(t, conf)
	assert.Error(t, err)
}

func TestLoadServersPConfig_ShouldPass(t *testing.T) {
	t.Parallel()

	_, err := os.Create("testFile")

	conf, err := core.LoadServersPConfig("testFile")

	if _, errF := os.Stat("testFile"); errF == nil {
		_ = os.Remove("testFile")
	}

	assert.NotNil(t, conf)
	assert.Nil(t, err)
}
