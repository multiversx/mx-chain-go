package chainSimulator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../cmd/node/config/"
)

func TestNewChainSimulator(t *testing.T) {
	chainSimulator, err := NewChainSimulator(3, defaultPathToInitialConfig)
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)
}
