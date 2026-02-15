package configs_test

import (
	"fmt"
	"strings"
	"testing"

	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/configs"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/stretchr/testify/require"
)

func getAntifloodConfigsByRound() []config.AntifloodConfigByRound {
	defaultAntifloodConfig := testscommon.GetDefaultAntifloodConfig()
	return defaultAntifloodConfig.ConfigsByRound
}

func TestNewAntifloodConfigsHandler(t *testing.T) {
	t.Parallel()

	validConfig := config.AntifloodConfig{
		Enabled:        true,
		ConfigsByRound: getAntifloodConfigsByRound(),
	}

	t.Run("should return error for empty configs by round", func(t *testing.T) {
		t.Parallel()

		afConf := validConfig
		afConf.ConfigsByRound = nil

		handler, err := configs.NewAntifloodConfigsHandler(afConf, &epochNotifier.RoundNotifierStub{})
		require.Nil(t, handler)
		require.Equal(t, configs.ErrEmptyAntifloodConfigsByRound, err)
	})

	t.Run("should return error for duplicated round config", func(t *testing.T) {
		t.Parallel()

		afConf := validConfig
		afConf.ConfigsByRound = append(afConf.ConfigsByRound, afConf.ConfigsByRound[0])

		handler, err := configs.NewAntifloodConfigsHandler(afConf, &epochNotifier.RoundNotifierStub{})
		require.Nil(t, handler)
		require.Equal(t, configs.ErrDuplicatedRoundConfig, err)
	})

	t.Run("should return error for missing round zero config", func(t *testing.T) {
		t.Parallel()

		afConf := validConfig
		afConf.ConfigsByRound = []config.AntifloodConfigByRound{
			getAntifloodConfigsByRound()[1],
		}

		handler, err := configs.NewAntifloodConfigsHandler(afConf, &epochNotifier.RoundNotifierStub{})
		require.Nil(t, handler)
		require.Equal(t, configs.ErrMissingRoundZeroConfig, err)
	})

	t.Run("should return error for invalid flood preventer config", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name        string
			mutate      func(*config.FloodPreventerConfig)
			expectedErr string
		}{
			{
				name: "zero ThresholdNumMessagesPerInterval",
				mutate: func(c *config.FloodPreventerConfig) {
					c.BlackList.ThresholdNumMessagesPerInterval = 0
				},
				expectedErr: "thresholdNumReceivedFlood == 0",
			},
			{
				name: "zero ThresholdSizePerInterval",
				mutate: func(c *config.FloodPreventerConfig) {
					c.BlackList.ThresholdSizePerInterval = 0
				},
				expectedErr: "thresholdSizeReceivedFlood == 0",
			},
			{
				name: "invalid PeerBanDurationInSeconds",
				mutate: func(c *config.FloodPreventerConfig) {
					c.BlackList.PeerBanDurationInSeconds = 0
				},
				expectedErr: "for peerBanDurationInSeconds",
			},
			{
				name: "zero NumFloodingRounds",
				mutate: func(c *config.FloodPreventerConfig) {
					c.BlackList.NumFloodingRounds = 0
				},
				expectedErr: "numFloodingRounds",
			},
			{
				name: "invalid BaseMessagesPerInterval",
				mutate: func(c *config.FloodPreventerConfig) {
					c.PeerMaxInput.BaseMessagesPerInterval = 0
				},
				expectedErr: "maxMessagesPerPeer",
			},
			{
				name: "invalid TotalSizePerInterval",
				mutate: func(c *config.FloodPreventerConfig) {
					c.PeerMaxInput.TotalSizePerInterval = 0
				},
				expectedErr: "maxTotalSizePerPeer",
			},
			{
				name: "ReservedPercent greater than max",
				mutate: func(c *config.FloodPreventerConfig) {
					c.ReservedPercent = 91.0
				},
				expectedErr: "percentReserved",
			},
			{
				name: "ReservedPercent less than min",
				mutate: func(c *config.FloodPreventerConfig) {
					c.ReservedPercent = -1.0
				},
				expectedErr: "percentReserved",
			},
			{
				name: "IncreaseFactor negative",
				mutate: func(c *config.FloodPreventerConfig) {
					c.PeerMaxInput.IncreaseFactor.Factor = -0.5
				},
				expectedErr: "increaseFactor is negative",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				afConf := config.AntifloodConfig{
					ConfigsByRound: getAntifloodConfigsByRound(),
				}

				// Mutate FastReacting config of the first round
				tc.mutate(&afConf.ConfigsByRound[0].FastReacting)

				handler, err := configs.NewAntifloodConfigsHandler(afConf, &epochNotifier.RoundNotifierStub{})
				require.Nil(t, handler)
				require.ErrorIs(t, err, process.ErrInvalidValue)
				require.True(t, strings.Contains(err.Error(), tc.expectedErr), fmt.Sprintf("expected error to contain '%s', got '%s'", tc.expectedErr, err.Error()))
			})
		}
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		afConf := validConfig

		handler, err := configs.NewAntifloodConfigsHandler(afConf, &epochNotifier.RoundNotifierStub{})
		require.NotNil(t, handler)
		require.NoError(t, err)
		require.False(t, handler.IsInterfaceNil())

		require.True(t, handler.IsEnabled())
	})
}

func TestAntifloodConfigs_GetCurrentConfig(t *testing.T) {
	t.Parallel()

	t.Run("first config", func(t *testing.T) {
		t.Parallel()

		afConf := config.AntifloodConfig{
			Enabled:        true,
			ConfigsByRound: getAntifloodConfigsByRound(),
		}

		roundNotifier := &epochNotifier.RoundNotifierStub{
			CurrentRoundCalled: func() uint64 {
				return 60
			},
		}
		handler, _ := configs.NewAntifloodConfigsHandler(afConf, roundNotifier)

		currentConfig := handler.GetCurrentConfig()

		// check some configs vars
		require.Equal(t, uint64(0), currentConfig.Round)
		require.Equal(t, float32(50), currentConfig.FastReacting.ReservedPercent)
	})

	t.Run("latest config", func(t *testing.T) {
		t.Parallel()

		afConf := config.AntifloodConfig{
			Enabled:        true,
			ConfigsByRound: getAntifloodConfigsByRound(),
		}

		roundNotifier := &epochNotifier.RoundNotifierStub{
			CurrentRoundCalled: func() uint64 {
				return 120
			},
		}
		handler, _ := configs.NewAntifloodConfigsHandler(afConf, roundNotifier)

		currentConfig := handler.GetCurrentConfig()

		// check some configs vars
		require.Equal(t, uint64(100), currentConfig.Round)
		require.Equal(t, float32(60), currentConfig.FastReacting.ReservedPercent)
	})
}

func TestAntifloodConfigs_GetFloodPreventerConfigByType(t *testing.T) {
	t.Parallel()

	conf := config.AntifloodConfig{
		Enabled:        true,
		ConfigsByRound: getAntifloodConfigsByRound(),
	}

	roundNotifier := &epochNotifier.RoundNotifierStub{
		CurrentRoundCalled: func() uint64 {
			return 120
		},
	}
	handler, _ := configs.NewAntifloodConfigsHandler(conf, roundNotifier)

	fastReacting := handler.GetFloodPreventerConfigByType(common.FastReacting)
	require.Equal(t, conf.ConfigsByRound[1].FastReacting, fastReacting)

	slowReacting := handler.GetFloodPreventerConfigByType(common.SlowReacting)
	require.Equal(t, conf.ConfigsByRound[1].SlowReacting, slowReacting)

	outOfSpecs := handler.GetFloodPreventerConfigByType(common.OutOfSpecs)
	require.Equal(t, conf.ConfigsByRound[1].OutOfSpecs, outOfSpecs)

	peerOutput := handler.GetFloodPreventerConfigByType(common.Output)
	require.Equal(t, conf.ConfigsByRound[1].PeerMaxOutput, peerOutput)

	other := handler.GetFloodPreventerConfigByType(common.FloodPreventerType("other"))
	require.Equal(t, config.FloodPreventerConfig{}, other) // empty config
}

func TestAntifloodConfigs_SetActivationRound(t *testing.T) {
	t.Parallel()

	afConf := config.AntifloodConfig{
		Enabled:        true,
		ConfigsByRound: getAntifloodConfigsByRound(),
	}

	currentRound := uint64(50)
	roundNotifier := &epochNotifier.RoundNotifierStub{
		CurrentRoundCalled: func() uint64 {
			return currentRound
		},
	}
	handler, err := configs.NewAntifloodConfigsHandler(afConf, roundNotifier)
	require.NoError(t, err)
	require.NotNil(t, handler)

	// Before SetActivationRound, the last config has round 100
	currentRound = 120
	currentConfig := handler.GetCurrentConfig()
	require.Equal(t, uint64(100), currentConfig.Round)

	// Set activation round to 500
	testLog := logger.GetOrCreate("test")
	handler.SetActivationRound(500, testLog)

	// Now at round 120, the last config (round 500) should not be active
	currentConfig = handler.GetCurrentConfig()
	require.Equal(t, uint64(0), currentConfig.Round)

	// At round 500, the last config should be active
	currentRound = 500
	currentConfig = handler.GetCurrentConfig()
	require.Equal(t, uint64(500), currentConfig.Round)
}
