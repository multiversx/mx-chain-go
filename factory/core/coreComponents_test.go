package core_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	coreComp "github.com/multiversx/mx-chain-go/factory/core"
	"github.com/multiversx/mx-chain-go/state"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
)

func TestNewCoreComponentsFactory_OkValuesShouldWork(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	require.NotNil(t, ccf)
}

func TestCoreComponentsFactory_CreateCoreComponentsNoHasherConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           componentsMock.TestMarshalizer,
			SizeCheckDelta: 0,
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidHasherConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           componentsMock.TestMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: "invalid_type",
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsNoInternalMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Hasher: config.TypeConfig{
			Type: componentsMock.TestHasher,
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidInternalMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           "invalid_marshalizer_type",
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: componentsMock.TestHasher,
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsNoVmMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Hasher: config.TypeConfig{
			Type: componentsMock.TestHasher,
		},
		Marshalizer: config.MarshalizerConfig{
			Type:           componentsMock.TestMarshalizer,
			SizeCheckDelta: 0,
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidVmMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           componentsMock.TestMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: componentsMock.TestHasher,
		},
		VmMarshalizer: config.TypeConfig{
			Type: "invalid",
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsNoTxSignMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Hasher: config.TypeConfig{
			Type: componentsMock.TestHasher,
		},
		Marshalizer: config.MarshalizerConfig{
			Type:           componentsMock.TestMarshalizer,
			SizeCheckDelta: 0,
		},
		VmMarshalizer: config.TypeConfig{
			Type: componentsMock.TestMarshalizer,
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidTxSignMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           componentsMock.TestMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: componentsMock.TestHasher,
		},
		VmMarshalizer: config.TypeConfig{
			Type: componentsMock.TestMarshalizer,
		},
		TxSignMarshalizer: config.TypeConfig{
			Type: "invalid",
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidTxSignHasherConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config.TxSignHasher = config.TypeConfig{
		Type: "invalid",
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidValPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config.ValidatorPubkeyConverter.Type = "invalid"
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, state.ErrInvalidPubkeyConverterType))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidAddrPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config.AddressPubkeyConverter.Type = "invalid"
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, state.ErrInvalidPubkeyConverterType))
}

func TestCoreComponentsFactory_CreateCoreComponentsNilChanStopNodeProcessShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.ChanStopNodeProcess = nil
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.NotNil(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidRoundConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.RoundConfig = config.RoundConfig{}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.NotNil(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidGenesisMaxNumberOfShardsShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config.GeneralSettings.GenesisMaxNumberOfShards = 0
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.NotNil(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidEconomicsConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.EconomicsConfig = config.EconomicsConfig{}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.NotNil(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidRatingsConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.RatingsConfig = config.RatingsConfig{}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.NotNil(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidHardforkPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config.Hardfork.PublicKeyToListenFrom = "invalid"
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.NotNil(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponentsShouldWork(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func TestCoreComponentsFactory_CreateCoreComponentsShouldWorkAfterHardfork(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config.Hardfork.AfterHardFork = true
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func TestCoreComponentsFactory_CreateSupernovaActivationTupleMismatchShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 0
	args.RoundConfig.RoundActivations["SupernovaEnableRound"] = config.ActivationRoundByName{Round: "100"}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
}

func TestCoreComponentsFactory_CreateSupernovaDisabledEpochNearRoundShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 1000000
	args.RoundConfig.RoundActivations["SupernovaEnableRound"] = config.ActivationRoundByName{Round: "100"}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
}

func TestCoreComponentsFactory_CreateSupernovaDisabledMainnetStyleShouldWork(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 999999
	args.RoundConfig.RoundActivations["SupernovaEnableRound"] = config.ActivationRoundByName{Round: "99999999999"}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
	// must stay far in the future, otherwise round.isSupernovaActivated force-activates
	require.True(t, cc.SupernovaGenesisTime().After(time.Now().AddDate(100, 0, 0)))
}

func TestValidateSupernovaActivationTuple(t *testing.T) {
	t.Parallel()

	supernovaEpoch := uint32(2)
	supernovaRound := uint64(440)

	coherentConfig := func() config.Config {
		return config.Config{
			Versions: config.VersionsConfig{
				VersionsByEpochs: []config.VersionByEpochs{
					{StartEpoch: 0, StartRound: 0, Version: "*"},
					{StartEpoch: 2, StartRound: 440, Version: "3"},
				},
			},
			GeneralSettings: config.GeneralSettingsConfig{
				ChainParametersByEpoch:        []config.ChainParametersByEpochConfig{{EnableEpoch: 0}, {EnableEpoch: 2}},
				EpochChangeGracePeriodByEpoch: []config.EpochChangeGracePeriodByEpoch{{EnableEpoch: 0}, {EnableEpoch: 2}},
				ProcessConfigsByEpoch:         []config.ProcessConfigByEpoch{{EnableEpoch: 0}, {EnableEpoch: 2}},
				EpochStartConfigsByEpoch:      []config.EpochStartConfigByEpoch{{EnableEpoch: 0}, {EnableEpoch: 2}},
				ConsensusConfigsByEpoch:       []config.ConsensusConfigByEpoch{{EnableEpoch: 0}, {EnableEpoch: 2}},
				ProcessConfigsByRound:         []config.ProcessConfigByRound{{EnableRound: 0}, {EnableRound: 440}},
				EpochStartConfigsByRound:      []config.EpochStartConfigByRound{{EnableRound: 0}, {EnableRound: 440}},
				ConsensusConfigsByRound:       []config.ConsensusConfigByRound{{EnableRound: 0}, {EnableRound: 440}},
			},
			Antiflood: config.AntifloodConfig{
				ConfigsByRound: []config.AntifloodConfigByRound{{Round: 0}, {Round: 440}},
			},
		}
	}

	coherentEconomics := func() config.EconomicsConfig {
		return config.EconomicsConfig{
			FeeSettings: config.FeeSettings{
				GasLimitSettings: []config.GasLimitSetting{{EnableEpoch: 0}, {EnableEpoch: 2}},
			},
		}
	}
	coherentRatings := func() config.RatingsConfig {
		return config.RatingsConfig{
			ShardChain: config.ShardChain{RatingStepsByEpoch: []config.RatingSteps{{EnableEpoch: 0}, {EnableEpoch: 2}}},
			MetaChain:  config.MetaChain{RatingStepsByEpoch: []config.RatingSteps{{EnableEpoch: 0}, {EnableEpoch: 2}}},
		}
	}

	t.Run("coherent tuple should work", func(t *testing.T) {
		t.Parallel()

		err := coreComp.ValidateSupernovaActivationTuple(coherentConfig(), coherentEconomics(), coherentRatings(), supernovaEpoch, supernovaRound)
		require.NoError(t, err)
	})

	t.Run("far away epoch skips the boundary alignment check", func(t *testing.T) {
		t.Parallel()

		err := coreComp.ValidateSupernovaActivationTuple(config.Config{}, config.EconomicsConfig{}, config.RatingsConfig{}, 999999, 99_999_999_999)
		require.NoError(t, err)
	})

	t.Run("disabled supernova with near activation round should error", func(t *testing.T) {
		t.Parallel()

		err := coreComp.ValidateSupernovaActivationTuple(config.Config{}, config.EconomicsConfig{}, config.RatingsConfig{}, 999999, supernovaRound)
		require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
		require.ErrorContains(t, err, "below")
	})

	t.Run("enabled supernova with missing antiflood round entry should error", func(t *testing.T) {
		t.Parallel()

		cfg := coherentConfig()
		cfg.Antiflood.ConfigsByRound = []config.AntifloodConfigByRound{{Round: 0}}
		err := coreComp.ValidateSupernovaActivationTuple(cfg, coherentEconomics(), coherentRatings(), supernovaEpoch, supernovaRound)
		require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
		require.ErrorContains(t, err, "Antiflood.ConfigsByRound")
	})

	t.Run("disabled supernova with near antiflood round entry should error", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			Antiflood: config.AntifloodConfig{
				ConfigsByRound: []config.AntifloodConfigByRound{{Round: 0}, {Round: 440}},
			},
		}
		err := coreComp.ValidateSupernovaActivationTuple(cfg, config.EconomicsConfig{}, config.RatingsConfig{}, 999999, 99_999_999_999)
		require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
		require.ErrorContains(t, err, "Antiflood.ConfigsByRound")
	})

	t.Run("enabled supernova with missing paired gas limit entry should error", func(t *testing.T) {
		t.Parallel()

		economics := config.EconomicsConfig{
			FeeSettings: config.FeeSettings{GasLimitSettings: []config.GasLimitSetting{{EnableEpoch: 0}}},
		}
		err := coreComp.ValidateSupernovaActivationTuple(coherentConfig(), economics, coherentRatings(), supernovaEpoch, supernovaRound)
		require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
		require.ErrorContains(t, err, "FeeSettings.GasLimitSettings")
	})

	t.Run("enabled supernova with missing paired shard rating steps should error", func(t *testing.T) {
		t.Parallel()

		ratings := coherentRatings()
		ratings.ShardChain.RatingStepsByEpoch = []config.RatingSteps{{EnableEpoch: 0}}
		err := coreComp.ValidateSupernovaActivationTuple(coherentConfig(), coherentEconomics(), ratings, supernovaEpoch, supernovaRound)
		require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
		require.ErrorContains(t, err, "ShardChain.RatingStepsByEpoch")
	})

	t.Run("enabled supernova with missing paired meta rating steps should error", func(t *testing.T) {
		t.Parallel()

		ratings := coherentRatings()
		ratings.MetaChain.RatingStepsByEpoch = []config.RatingSteps{{EnableEpoch: 0}}
		err := coreComp.ValidateSupernovaActivationTuple(coherentConfig(), coherentEconomics(), ratings, supernovaEpoch, supernovaRound)
		require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
		require.ErrorContains(t, err, "MetaChain.RatingStepsByEpoch")
	})

	t.Run("disabled supernova with uniform round duration should work", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			GeneralSettings: config.GeneralSettingsConfig{
				ChainParametersByEpoch: []config.ChainParametersByEpochConfig{
					{EnableEpoch: 0, RoundDuration: 6000},
					{EnableEpoch: 2, RoundDuration: 6000},
				},
			},
		}
		err := coreComp.ValidateSupernovaActivationTuple(cfg, config.EconomicsConfig{}, config.RatingsConfig{}, 999999, 99_999_999_999)
		require.NoError(t, err)
	})

	t.Run("disabled supernova with near round duration change should error regardless of order", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			GeneralSettings: config.GeneralSettingsConfig{
				ChainParametersByEpoch: []config.ChainParametersByEpochConfig{
					{EnableEpoch: 999999, RoundDuration: 600},
					{EnableEpoch: 1, RoundDuration: 600},
					{EnableEpoch: 0, RoundDuration: 6000},
				},
			},
		}
		err := coreComp.ValidateSupernovaActivationTuple(cfg, config.EconomicsConfig{}, config.RatingsConfig{}, 999999, 99_999_999_999)
		require.ErrorIs(t, err, errorsMx.ErrSupernovaActivationConfigMismatch)
		require.ErrorContains(t, err, "entry at epoch 1")
	})

	t.Run("disabled supernova accepts descending parameters with change at activation", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			GeneralSettings: config.GeneralSettingsConfig{
				ChainParametersByEpoch: []config.ChainParametersByEpochConfig{
					{EnableEpoch: 999999, RoundDuration: 600},
					{EnableEpoch: 1, RoundDuration: 6000},
					{EnableEpoch: 0, RoundDuration: 6000},
				},
			},
		}
		err := coreComp.ValidateSupernovaActivationTuple(cfg, config.EconomicsConfig{}, config.RatingsConfig{}, 999999, 99_999_999_999)
		require.NoError(t, err)
	})

	t.Run("disabled supernova with duration change parked at far away epoch should work", func(t *testing.T) {
		t.Parallel()

		// mainnet-style disabled convention: the supernova chain-params entry moved to the sentinel
		cfg := config.Config{
			GeneralSettings: config.GeneralSettingsConfig{
				ChainParametersByEpoch: []config.ChainParametersByEpochConfig{
					{EnableEpoch: 0, RoundDuration: 6000},
					{EnableEpoch: 1763, RoundDuration: 6000},
					{EnableEpoch: 999999, RoundDuration: 600},
				},
			},
		}
		err := coreComp.ValidateSupernovaActivationTuple(cfg, config.EconomicsConfig{}, config.RatingsConfig{}, 999999, 99_999_999_999)
		require.NoError(t, err)
	})

	t.Run("disabled supernova with coherent far away values should work", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			Versions: config.VersionsConfig{
				VersionsByEpochs: []config.VersionByEpochs{
					{StartEpoch: 0, StartRound: 0, Version: "*"},
					{StartEpoch: 9999999, StartRound: 99_999_999_999, Version: "3"},
				},
			},
			GeneralSettings: config.GeneralSettingsConfig{
				ProcessConfigsByRound:    []config.ProcessConfigByRound{{EnableRound: 0}, {EnableRound: 99_999_999_999}},
				EpochStartConfigsByRound: []config.EpochStartConfigByRound{{EnableRound: 0}},
				ConsensusConfigsByRound:  []config.ConsensusConfigByRound{{EnableRound: 0}},
			},
		}
		err := coreComp.ValidateSupernovaActivationTuple(cfg, config.EconomicsConfig{}, config.RatingsConfig{}, 9999999, 99_999_999_999)
		require.NoError(t, err)
	})

	t.Run("disabled supernova with near version 3 entry should error", func(t *testing.T) {
		t.Parallel()

		cfg := coherentConfig()
		err := coreComp.ValidateSupernovaActivationTuple(cfg, config.EconomicsConfig{}, config.RatingsConfig{}, 9999999, 99_999_999_999)
		require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
		require.ErrorContains(t, err, "StartEpoch 2")
	})

	t.Run("disabled supernova with near round-keyed entries should error", func(t *testing.T) {
		t.Parallel()

		cfg := coherentConfig()
		cfg.Versions.VersionsByEpochs[1].StartEpoch = 9999999
		cfg.Versions.VersionsByEpochs[1].StartRound = 9999999
		err := coreComp.ValidateSupernovaActivationTuple(cfg, config.EconomicsConfig{}, config.RatingsConfig{}, 9999999, 99_999_999_999)
		require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
		require.ErrorContains(t, err, "GeneralSettings.ProcessConfigsByRound")
		require.ErrorContains(t, err, "GeneralSettings.EpochStartConfigsByRound")
		require.ErrorContains(t, err, "GeneralSettings.ConsensusConfigsByRound")
	})

	t.Run("disabled supernova with mainnet-scale round leftover should error", func(t *testing.T) {
		t.Parallel()

		cfg := coherentConfig()
		cfg.Versions.VersionsByEpochs[1].StartEpoch = 9999999
		cfg.GeneralSettings.EpochStartConfigsByRound = []config.EpochStartConfigByRound{{EnableRound: 0}}
		cfg.GeneralSettings.ConsensusConfigsByRound = []config.ConsensusConfigByRound{{EnableRound: 0}}
		cfg.GeneralSettings.ProcessConfigsByRound = []config.ProcessConfigByRound{{EnableRound: 0}, {EnableRound: 31608234}}
		err := coreComp.ValidateSupernovaActivationTuple(cfg, config.EconomicsConfig{}, config.RatingsConfig{}, 9999999, 99_999_999_999)
		require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
		require.ErrorContains(t, err, "GeneralSettings.ProcessConfigsByRound")
	})

	t.Run("missing version 3 entry should error", func(t *testing.T) {
		t.Parallel()

		cfg := coherentConfig()
		cfg.Versions.VersionsByEpochs = cfg.Versions.VersionsByEpochs[:1]
		err := coreComp.ValidateSupernovaActivationTuple(cfg, coherentEconomics(), coherentRatings(), supernovaEpoch, supernovaRound)
		require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
	})

	t.Run("wrong version start epoch should error", func(t *testing.T) {
		t.Parallel()

		cfg := coherentConfig()
		cfg.Versions.VersionsByEpochs[1].StartEpoch = 3
		err := coreComp.ValidateSupernovaActivationTuple(cfg, coherentEconomics(), coherentRatings(), supernovaEpoch, supernovaRound)
		require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
	})

	t.Run("wrong version start round should error", func(t *testing.T) {
		t.Parallel()

		cfg := coherentConfig()
		cfg.Versions.VersionsByEpochs[1].StartRound = 441
		err := coreComp.ValidateSupernovaActivationTuple(cfg, coherentEconomics(), coherentRatings(), supernovaEpoch, supernovaRound)
		require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
	})

	listMutations := map[string]func(cfg *config.Config){
		"GeneralSettings.ChainParametersByEpoch": func(cfg *config.Config) {
			cfg.GeneralSettings.ChainParametersByEpoch = cfg.GeneralSettings.ChainParametersByEpoch[:1]
		},
		"GeneralSettings.EpochChangeGracePeriodByEpoch": func(cfg *config.Config) {
			cfg.GeneralSettings.EpochChangeGracePeriodByEpoch = cfg.GeneralSettings.EpochChangeGracePeriodByEpoch[:1]
		},
		"GeneralSettings.ProcessConfigsByEpoch": func(cfg *config.Config) {
			cfg.GeneralSettings.ProcessConfigsByEpoch = cfg.GeneralSettings.ProcessConfigsByEpoch[:1]
		},
		"GeneralSettings.EpochStartConfigsByEpoch": func(cfg *config.Config) {
			cfg.GeneralSettings.EpochStartConfigsByEpoch = cfg.GeneralSettings.EpochStartConfigsByEpoch[:1]
		},
		"GeneralSettings.ConsensusConfigsByEpoch": func(cfg *config.Config) {
			cfg.GeneralSettings.ConsensusConfigsByEpoch = cfg.GeneralSettings.ConsensusConfigsByEpoch[:1]
		},
		"GeneralSettings.ProcessConfigsByRound": func(cfg *config.Config) {
			cfg.GeneralSettings.ProcessConfigsByRound = cfg.GeneralSettings.ProcessConfigsByRound[:1]
		},
		"GeneralSettings.EpochStartConfigsByRound": func(cfg *config.Config) {
			cfg.GeneralSettings.EpochStartConfigsByRound = cfg.GeneralSettings.EpochStartConfigsByRound[:1]
		},
		"GeneralSettings.ConsensusConfigsByRound": func(cfg *config.Config) {
			cfg.GeneralSettings.ConsensusConfigsByRound = cfg.GeneralSettings.ConsensusConfigsByRound[:1]
		},
	}
	for listName, mutate := range listMutations {
		t.Run("missing boundary entry in "+listName, func(t *testing.T) {
			t.Parallel()

			cfg := coherentConfig()
			mutate(&cfg)
			err := coreComp.ValidateSupernovaActivationTuple(cfg, coherentEconomics(), coherentRatings(), supernovaEpoch, supernovaRound)
			require.True(t, errors.Is(err, errorsMx.ErrSupernovaActivationConfigMismatch))
			require.ErrorContains(t, err, listName)
		})
	}
}

// ------------ Test CoreComponents --------------------
func TestCoreComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	ccf, _ := coreComp.NewCoreComponentsFactory(args)
	cc, _ := ccf.Create()
	err := cc.Close()

	require.NoError(t, err)
}
