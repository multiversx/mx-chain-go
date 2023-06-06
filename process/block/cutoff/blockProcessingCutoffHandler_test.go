package cutoff

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

func TestNewBlockProcessingCutoffHandler(t *testing.T) {
	t.Parallel()

	t.Run("invalid mode - should error", func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled: true,
			Mode:    "invalid",
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.Equal(t, "invalid block processing cutoff mode, provided value=invalid", err.Error())
		require.Nil(t, b)
	})

	t.Run("invalid cutoff trigger - should error", func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       true,
			Mode:          "pause",
			CutoffTrigger: "invalid",
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.Equal(t, "invalid block processing cutoff trigger, provided value=invalid", err.Error())
		require.Nil(t, b)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       true,
			Mode:          "pause",
			CutoffTrigger: "epoch",
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.NoError(t, err)
		require.False(t, check.IfNil(b))
	})
}

func TestBlockProcessingCutoffHandler_HandlePauseBackoff(t *testing.T) {
	t.Parallel()

	t.Run("bad config - should skip", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			require.Nil(t, r)
		}()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       false,
			Mode:          "pause",
			CutoffTrigger: "nonce",
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.NoError(t, err)

		b.HandlePauseCutoff(nil)
		b.HandlePauseCutoff(&block.MetaBlock{})
		b.config.CutoffTrigger = common.BlockProcessingCutoffModeProcessError
		b.HandlePauseCutoff(&block.MetaBlock{})
	})

	t.Run("pause via round - should work", testHandlePauseCutoff(string(common.BlockProcessingCutoffByRound)))
	t.Run("pause via nonce - should work", testHandlePauseCutoff(string(common.BlockProcessingCutoffByNonce)))
	t.Run("pause via epoch - should work", testHandlePauseCutoff(string(common.BlockProcessingCutoffByEpoch)))
}

func testHandlePauseCutoff(trigger string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       true,
			Mode:          common.BlockProcessingCutoffModePause,
			CutoffTrigger: trigger,
			Value:         20,
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.NoError(t, err)

		b.HandlePauseCutoff(&block.MetaBlock{
			Epoch: 19, // not the desired epoch
			Nonce: 19, // not the desired nonce
			Round: 19, // not the desired round
		})

		done := make(chan struct{})
		go func() {
			b.HandlePauseCutoff(&block.MetaBlock{
				Epoch: 20,
				Nonce: 20,
				Round: 20,
			})
			done <- struct{}{}
		}()

		select {
		case <-done:
			require.Fail(t, "should have not advanced")
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func TestBlockProcessingCutoffHandler_HandleProcessErrorBackoff(t *testing.T) {
	t.Parallel()

	t.Run("bad config - should skip", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			require.Nil(t, r)
		}()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       false,
			Mode:          "pause",
			CutoffTrigger: "nonce",
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.NoError(t, err)

		err = b.HandleProcessErrorCutoff(nil)
		require.NoError(t, err)
		err = b.HandleProcessErrorCutoff(&block.MetaBlock{})
		require.NoError(t, err)
		b.config.CutoffTrigger = "pause"
		err = b.HandleProcessErrorCutoff(&block.MetaBlock{})
		require.NoError(t, err)
	})

	t.Run("process error via round - should work", testHandleProcessErrorCutoff(string(common.BlockProcessingCutoffByRound)))
	t.Run("process error via nonce - should work", testHandleProcessErrorCutoff(string(common.BlockProcessingCutoffByNonce)))
	t.Run("process error via epoch - should work", testHandleProcessErrorCutoff(string(common.BlockProcessingCutoffByEpoch)))
}

func testHandleProcessErrorCutoff(trigger string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       true,
			Mode:          common.BlockProcessingCutoffModeProcessError,
			CutoffTrigger: trigger,
			Value:         20,
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.NoError(t, err)

		err = b.HandleProcessErrorCutoff(&block.MetaBlock{
			Epoch: 19, // not the desired epoch
			Nonce: 19, // not the desired nonce
			Round: 19, // not the desired round
		})
		require.NoError(t, err)

		err = b.HandleProcessErrorCutoff(&block.MetaBlock{
			Epoch: 20,
			Nonce: 20,
			Round: 20,
		})
		require.Equal(t, errProcess, err)
	}
}
