package components

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon/components"

	"github.com/multiversx/mx-chain-go/config"
)

func createArgsCoreComponentsHolder() ArgsCoreComponentsHolder {
	return ArgsCoreComponentsHolder{
		Config: config.Config{
			Marshalizer: config.MarshalizerConfig{
				Type: "json",
			},
			TxSignMarshalizer: config.TypeConfig{
				Type: "json",
			},
			VmMarshalizer: config.TypeConfig{
				Type: "json",
			},
			Hasher: config.TypeConfig{
				Type: "blake2b",
			},
			TxSignHasher: config.TypeConfig{
				Type: "blake2b",
			},
			AddressPubkeyConverter: config.PubkeyConfig{
				Length: 32,
				Type:   "bech32",
				Hrp:    "erd",
			},
			ValidatorPubkeyConverter: config.PubkeyConfig{
				Length: 96,
				Type:   "hex",
			},
			GeneralSettings: config.GeneralSettingsConfig{
				ChainID:               "T",
				MinTransactionVersion: 1,
				ChainParametersByEpoch: []config.ChainParametersByEpochConfig{
					{
						EnableEpoch:                 0,
						RoundDuration:               4000,
						ShardConsensusGroupSize:     1,
						ShardMinNumNodes:            1,
						MetachainConsensusGroupSize: 1,
						MetachainMinNumNodes:        1,
						Hysteresis:                  0,
						Adaptivity:                  false,
					},
				},
				EpochChangeGracePeriodByEpoch: []config.EpochChangeGracePeriodByEpoch{{EnableEpoch: 0, GracePeriodInRounds: 1}},
			},
			Hardfork: config.HardforkConfig{
				PublicKeyToListenFrom: components.DummyPk,
			},
		},
		EnableEpochsConfig: config.EnableEpochs{},
		RoundsConfig: config.RoundConfig{
			RoundActivations: map[string]config.ActivationRoundByName{
				"DisableAsyncCallV1": {
					Round: "18446744073709551615",
				},
			},
		},
		EconomicsConfig: config.EconomicsConfig{
			GlobalSettings: config.GlobalSettings{
				GenesisTotalSupply: "2000000000000000000000",
				MinimumInflation:   0,
				YearSettings: []*config.YearSetting{
					{
						Year:             0,
						MaximumInflation: 0.01,
					},
				},
			},
			FeeSettings: config.FeeSettings{
				GasLimitSettings: []config.GasLimitSetting{
					{
						MaxGasLimitPerBlock:         "10000000000",
						MaxGasLimitPerMiniBlock:     "10000000000",
						MaxGasLimitPerMetaBlock:     "10000000000",
						MaxGasLimitPerMetaMiniBlock: "10000000000",
						MaxGasLimitPerTx:            "10000000000",
						MinGasLimit:                 "10",
						ExtraGasLimitGuardedTx:      "50000",
					},
				},
				GasPriceModifier:       0.01,
				MinGasPrice:            "100",
				GasPerDataByte:         "1",
				MaxGasPriceSetGuardian: "100",
			},
			RewardsSettings: config.RewardsSettings{
				RewardsConfigByEpoch: []config.EpochRewardSettings{
					{
						LeaderPercentage:                 0.1,
						DeveloperPercentage:              0.1,
						ProtocolSustainabilityPercentage: 0.1,
						ProtocolSustainabilityAddress:    testingProtocolSustainabilityAddress,
						TopUpGradientPoint:               "300000000000000000000",
						TopUpFactor:                      0.25,
						EpochEnable:                      0,
					},
				},
			},
		},
		RatingConfig: config.RatingsConfig{
			General: config.General{
				StartRating:           4000,
				MaxRating:             10000,
				MinRating:             1,
				SignedBlocksThreshold: 0.025,
				SelectionChances: []*config.SelectionChance{
					{MaxThreshold: 0, ChancePercent: 1},
					{MaxThreshold: 1, ChancePercent: 2},
					{MaxThreshold: 10000, ChancePercent: 4},
				},
			},
			ShardChain: config.ShardChain{
				RatingStepsByEpoch: []config.RatingSteps{
					{
						HoursToMaxRatingFromStartRating: 2,
						ProposerValidatorImportance:     1,
						ProposerDecreaseFactor:          -4,
						ValidatorDecreaseFactor:         -4,
						ConsecutiveMissedBlocksPenalty:  1.2,
					},
				},
			},
			MetaChain: config.MetaChain{
				RatingStepsByEpoch: []config.RatingSteps{
					{
						HoursToMaxRatingFromStartRating: 2,
						ProposerValidatorImportance:     1,
						ProposerDecreaseFactor:          -4,
						ValidatorDecreaseFactor:         -4,
						ConsecutiveMissedBlocksPenalty:  1.3,
					},
				},
			},
		},
		ChanStopNodeProcess:         make(chan endProcess.ArgEndProcess),
		InitialRound:                0,
		NodesSetupPath:              "../../../cmd/node/config/nodesSetup.json",
		GasScheduleFilename:         "../../../cmd/node/config/gasSchedules/gasScheduleV8.toml",
		NumShards:                   3,
		WorkingDir:                  ".",
		MinNodesPerShard:            1,
		MinNodesMeta:                1,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
		RoundDurationInMs:           6000,
	}
}

func TestCreateCoreComponents(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		comp, err := CreateCoreComponents(createArgsCoreComponentsHolder())
		require.NoError(t, err)
		require.NotNil(t, comp)

		require.Nil(t, comp.Create())
		require.Nil(t, comp.Close())
	})
	t.Run("internal NewMarshalizer failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsCoreComponentsHolder()
		args.Config.Marshalizer.Type = "invalid"
		comp, err := CreateCoreComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("tx NewMarshalizer failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsCoreComponentsHolder()
		args.Config.TxSignMarshalizer.Type = "invalid"
		comp, err := CreateCoreComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("vm NewMarshalizer failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsCoreComponentsHolder()
		args.Config.VmMarshalizer.Type = "invalid"
		comp, err := CreateCoreComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("main NewHasher failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsCoreComponentsHolder()
		args.Config.Hasher.Type = "invalid"
		comp, err := CreateCoreComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("tx NewHasher failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsCoreComponentsHolder()
		args.Config.TxSignHasher.Type = "invalid"
		comp, err := CreateCoreComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("address NewPubkeyConverter failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsCoreComponentsHolder()
		args.Config.AddressPubkeyConverter.Type = "invalid"
		comp, err := CreateCoreComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("validator NewPubkeyConverter failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsCoreComponentsHolder()
		args.Config.ValidatorPubkeyConverter.Type = "invalid"
		comp, err := CreateCoreComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("NewNodesSetup failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsCoreComponentsHolder()
		args.NumShards = 0
		comp, err := CreateCoreComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("NewEconomicsData failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsCoreComponentsHolder()
		args.EconomicsConfig.GlobalSettings.MinimumInflation = -1.0
		comp, err := CreateCoreComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("validatorPubKeyConverter.Decode failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsCoreComponentsHolder()
		args.Config.Hardfork.PublicKeyToListenFrom = "invalid"
		comp, err := CreateCoreComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
}

func TestCoreComponentsHolder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var comp *coreComponentsHolder
	require.True(t, comp.IsInterfaceNil())

	comp, _ = CreateCoreComponents(createArgsCoreComponentsHolder())
	require.False(t, comp.IsInterfaceNil())
	require.Nil(t, comp.Close())
}

func TestCoreComponents_GettersSetters(t *testing.T) {
	t.Parallel()

	comp, err := CreateCoreComponents(createArgsCoreComponentsHolder())
	require.NoError(t, err)

	require.NotNil(t, comp.InternalMarshalizer())
	require.Nil(t, comp.SetInternalMarshalizer(nil))
	require.Nil(t, comp.InternalMarshalizer())

	require.NotNil(t, comp.TxMarshalizer())
	require.NotNil(t, comp.VmMarshalizer())
	require.NotNil(t, comp.Hasher())
	require.NotNil(t, comp.TxSignHasher())
	require.NotNil(t, comp.Uint64ByteSliceConverter())
	require.NotNil(t, comp.AddressPubKeyConverter())
	require.NotNil(t, comp.ValidatorPubKeyConverter())
	require.NotNil(t, comp.PathHandler())
	require.NotNil(t, comp.Watchdog())
	require.NotNil(t, comp.AlarmScheduler())
	require.NotNil(t, comp.SyncTimer())
	require.NotNil(t, comp.RoundHandler())
	require.NotNil(t, comp.EconomicsData())
	require.NotNil(t, comp.APIEconomicsData())
	require.NotNil(t, comp.RatingsData())
	require.NotNil(t, comp.Rater())
	require.NotNil(t, comp.GenesisNodesSetup())
	require.NotNil(t, comp.NodesShuffler())
	require.NotNil(t, comp.EpochNotifier())
	require.NotNil(t, comp.EnableRoundsHandler())
	require.NotNil(t, comp.RoundNotifier())
	require.NotNil(t, comp.EpochStartNotifierWithConfirm())
	require.NotNil(t, comp.ChanStopNodeProcess())
	require.NotNil(t, comp.GenesisTime())
	require.Equal(t, "T", comp.ChainID())
	require.Equal(t, uint32(1), comp.MinTransactionVersion())
	require.NotNil(t, comp.TxVersionChecker())
	require.Equal(t, uint32(62), comp.EncodedAddressLen())
	hfPk, _ := hex.DecodeString(components.DummyPk)
	require.Equal(t, hfPk, comp.HardforkTriggerPubKey())
	require.NotNil(t, comp.NodeTypeProvider())
	require.NotNil(t, comp.WasmVMChangeLocker())
	require.NotNil(t, comp.ProcessStatusHandler())
	require.NotNil(t, comp.ProcessStatusHandler())
	require.NotNil(t, comp.EnableEpochsHandler())
	require.Nil(t, comp.CheckSubcomponents())
	require.Empty(t, comp.String())
	require.Nil(t, comp.Close())
}
