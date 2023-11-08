package chainSimulator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../cmd/node/config/"
)

func TestNewChainSimulator(t *testing.T) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(3, defaultPathToInitialConfig, startTime, roundDurationInMillis)
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)
	defer chainSimulator.Stop()

	time.Sleep(time.Second)
}

func TestChainSimulator_GenerateBlocksShouldWork(t *testing.T) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(3, defaultPathToInitialConfig, startTime, roundDurationInMillis)
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)
	defer chainSimulator.Stop()

	time.Sleep(time.Second)

	err = chainSimulator.GenerateBlocks(10)
	require.Nil(t, err)
}
