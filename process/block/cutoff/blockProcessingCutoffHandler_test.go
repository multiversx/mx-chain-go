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
		require.Equal(t, "invalid block processing cutoff mode. provided value=invalid", err.Error())
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
		require.Equal(t, "invalid block processing cutoff trigger. provided value=invalid", err.Error())
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
			Enabled: false,
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.NoError(t, err)

		b.HandlePauseCutoff(nil)
		b.HandlePauseCutoff(&block.MetaBlock{})
		b.config.CutoffTrigger = common.BlockProcessingCutoffModeProcessError
		b.HandlePauseCutoff(&block.MetaBlock{})
	})

	t.Run("pause via round - should work", func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       true,
			Mode:          common.BlockProcessingCutoffModePause,
			CutoffTrigger: string(common.BlockProcessingCutoffByRound),
			Value:         20,
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.NoError(t, err)

		err = b.HandleProcessErrorCutoff(&block.MetaBlock{Round: 19}) // not the desired round
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			b.HandlePauseCutoff(&block.MetaBlock{Round: 20})
			done <- struct{}{}
		}()

		select {
		case <-done:
			require.Fail(t, "should have not advanced")
		case <-time.After(100 * time.Millisecond):
		}
	})

	t.Run("pause via nonce - should work", func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       true,
			Mode:          common.BlockProcessingCutoffModePause,
			CutoffTrigger: string(common.BlockProcessingCutoffByNonce),
			Value:         20,
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.NoError(t, err)

		err = b.HandleProcessErrorCutoff(&block.MetaBlock{Nonce: 19}) // not the desired round
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			b.HandlePauseCutoff(&block.MetaBlock{Nonce: 20})
			done <- struct{}{}
		}()

		select {
		case <-done:
			require.Fail(t, "should have not advanced")
		case <-time.After(100 * time.Millisecond):
		}
	})

	t.Run("pause via epoch - should work", func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       true,
			Mode:          common.BlockProcessingCutoffModePause,
			CutoffTrigger: string(common.BlockProcessingCutoffByEpoch),
			Value:         20,
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.NoError(t, err)

		err = b.HandleProcessErrorCutoff(&block.MetaBlock{Epoch: 19}) // not the desired round
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			b.HandlePauseCutoff(&block.MetaBlock{Epoch: 20})
			done <- struct{}{}
		}()

		select {
		case <-done:
			require.Fail(t, "should have not advanced")
		case <-time.After(time.Millisecond):
		}
	})
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
			Enabled: false,
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

	t.Run("process error via round", func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       true,
			Mode:          common.BlockProcessingCutoffModeProcessError,
			CutoffTrigger: string(common.BlockProcessingCutoffByRound),
			Value:         20,
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.NoError(t, err)

		err = b.HandleProcessErrorCutoff(&block.MetaBlock{Round: 19}) // not the desired round
		require.NoError(t, err)

		err = b.HandleProcessErrorCutoff(&block.MetaBlock{Round: 20})
		require.Equal(t, errProcess, err)
	})

	t.Run("process error via nonce", func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       true,
			Mode:          common.BlockProcessingCutoffModeProcessError,
			CutoffTrigger: string(common.BlockProcessingCutoffByNonce),
			Value:         20,
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.NoError(t, err)

		err = b.HandleProcessErrorCutoff(&block.MetaBlock{Nonce: 19}) // not the desired nonce
		require.NoError(t, err)

		err = b.HandleProcessErrorCutoff(&block.MetaBlock{Nonce: 20})
		require.Equal(t, errProcess, err)
	})

	t.Run("process error via epoch", func(t *testing.T) {
		t.Parallel()

		cfg := config.BlockProcessingCutoffConfig{
			Enabled:       true,
			Mode:          common.BlockProcessingCutoffModeProcessError,
			CutoffTrigger: string(common.BlockProcessingCutoffByEpoch),
			Value:         20,
		}
		b, err := NewBlockProcessingCutoffHandler(cfg)
		require.NoError(t, err)

		dummyEpochStartData := block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					ShardID: 0,
				},
			},
		}
		err = b.HandleProcessErrorCutoff(&block.MetaBlock{Epoch: 19, EpochStart: dummyEpochStartData}) // not the desired nonce
		require.NoError(t, err)

		err = b.HandleProcessErrorCutoff(&block.MetaBlock{Epoch: 20, EpochStart: dummyEpochStartData})
		require.Equal(t, errProcess, err)
	})
}
