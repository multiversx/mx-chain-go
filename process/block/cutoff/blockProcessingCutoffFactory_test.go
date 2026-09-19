package cutoff

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

func TestCreateBlockProcessingCutoffHandler(t *testing.T) {
	t.Parallel()

	t.Run("should create disabled instance", func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled: false,
		}

		instance, err := CreateBlockProcessingCutoffHandler(cfg, nil)
		require.NoError(t, err)
		require.Equal(t, "*cutoff.disabledBlockProcessingCutoff", fmt.Sprintf("%T", instance))
	})

	t.Run("should create regular instance", func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       true,
			Mode:          "pause",
			CutoffTrigger: "nonce",
			Value:         37,
		}

		instance, err := CreateBlockProcessingCutoffHandler(cfg, nil)
		require.NoError(t, err)
		require.Equal(t, "*cutoff.blockProcessingCutoffHandler", fmt.Sprintf("%T", instance))
	})

	t.Run("should create graceful stop instance", func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       true,
			Mode:          common.BlockProcessingCutoffModeGracefulStop,
			CutoffTrigger: string(common.BlockProcessingCutoffByRound),
			Value:         37,
		}

		instance, err := CreateBlockProcessingCutoffHandler(cfg, make(chan endProcess.ArgEndProcess, 1))
		require.NoError(t, err)
		require.Equal(t, "*cutoff.blockProcessingCutoffHandler", fmt.Sprintf("%T", instance))
	})
}
