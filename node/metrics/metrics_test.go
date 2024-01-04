package metrics

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitBaseMetrics(t *testing.T) {
	t.Parallel()

	expectedKeys := []string{
		common.MetricSynchronizedRound,
		common.MetricNonce,
		common.MetricCountConsensus,
		common.MetricCountLeader,
		common.MetricCountAcceptedBlocks,
		common.MetricNumTxInBlock,
		common.MetricNumMiniBlocks,
		common.MetricNumProcessedTxs,
		common.MetricCurrentRoundTimestamp,
		common.MetricHeaderSize,
		common.MetricMiniBlocksSize,
		common.MetricNumShardHeadersFromPool,
		common.MetricNumShardHeadersProcessed,
		common.MetricNumTimesInForkChoice,
		common.MetricHighestFinalBlock,
		common.MetricCountConsensusAcceptedBlocks,
		common.MetricRoundsPassedInCurrentEpoch,
		common.MetricNoncesPassedInCurrentEpoch,
		common.MetricNumConnectedPeers,
		common.MetricEpochForEconomicsData,
		common.MetricConsensusState,
		common.MetricConsensusRoundState,
		common.MetricCurrentBlockHash,
		common.MetricNumConnectedPeersClassification,
		common.MetricLatestTagSoftwareVersion,
		common.MetricAreVMQueriesReady,
		common.MetricP2PNumConnectedPeersClassification,
		common.MetricP2PPeerInfo,
		common.MetricP2PIntraShardValidators,
		common.MetricP2PIntraShardObservers,
		common.MetricP2PCrossShardValidators,
		common.MetricP2PCrossShardObservers,
		common.MetricP2PUnknownPeers,
		common.MetricInflation,
		common.MetricDevRewardsInEpoch,
		common.MetricTotalFees,
		common.MetricAccountsSnapshotInProgress,
		common.MetricLastAccountsSnapshotDurationSec,
		common.MetricPeersSnapshotInProgress,
		common.MetricLastPeersSnapshotDurationSec,
		common.MetricAccountsSnapshotNumNodes,
		common.MetricTrieSyncNumProcessedNodes,
		common.MetricTrieSyncNumReceivedBytes,
	}

	keys := make(map[string]struct{})

	ash := &statusHandler.AppStatusHandlerStub{
		SetStringValueHandler: func(key string, value string) {
			okValue := value == initString || value == initZeroString || value == strconv.FormatBool(false)
			require.True(t, okValue)
			keys[key] = struct{}{}
		},
		SetUInt64ValueHandler: func(key string, value uint64) {
			require.Equal(t, value, initUint)
			keys[key] = struct{}{}
		},
		SetInt64ValueHandler: func(key string, value int64) {
			require.Equal(t, value, initInt)
			keys[key] = struct{}{}
		},
	}

	err := InitBaseMetrics(nil)
	require.Equal(t, ErrNilAppStatusHandler, err)

	err = InitBaseMetrics(ash)
	require.Nil(t, err)

	require.Equal(t, len(expectedKeys), len(keys))
	for _, key := range expectedKeys {
		_, found := keys[key]
		assert.True(t, found, fmt.Sprintf("key not found: %s", key))
	}
}

func TestInitConfigMetrics(t *testing.T) {
	t.Parallel()

	cfg := config.EpochConfig{
		EnableEpochs: config.EnableEpochs{
			SCDeployEnableEpoch:                         1,
			BuiltInFunctionsEnableEpoch:                 2,
			RelayedTransactionsEnableEpoch:              3,
			PenalizedTooMuchGasEnableEpoch:              4,
			SwitchJailWaitingEnableEpoch:                5,
			SwitchHysteresisForMinNodesEnableEpoch:      6,
			BelowSignedThresholdEnableEpoch:             7,
			TransactionSignedWithTxHashEnableEpoch:      8,
			MetaProtectionEnableEpoch:                   9,
			AheadOfTimeGasUsageEnableEpoch:              10,
			GasPriceModifierEnableEpoch:                 11,
			RepairCallbackEnableEpoch:                   12,
			BlockGasAndFeesReCheckEnableEpoch:           13,
			StakingV2EnableEpoch:                        14,
			StakeEnableEpoch:                            15,
			DoubleKeyProtectionEnableEpoch:              16,
			ESDTEnableEpoch:                             17,
			GovernanceEnableEpoch:                       18,
			DelegationManagerEnableEpoch:                19,
			DelegationSmartContractEnableEpoch:          20,
			CorrectLastUnjailedEnableEpoch:              21,
			BalanceWaitingListsEnableEpoch:              22,
			ReturnDataToLastTransferEnableEpoch:         23,
			SenderInOutTransferEnableEpoch:              24,
			RelayedTransactionsV2EnableEpoch:            25,
			UnbondTokensV2EnableEpoch:                   26,
			SaveJailedAlwaysEnableEpoch:                 27,
			ValidatorToDelegationEnableEpoch:            28,
			ReDelegateBelowMinCheckEnableEpoch:          29,
			IncrementSCRNonceInMultiTransferEnableEpoch: 30,
			ESDTMultiTransferEnableEpoch:                31,
			GlobalMintBurnDisableEpoch:                  32,
			ESDTTransferRoleEnableEpoch:                 33,
			SetGuardianEnableEpoch:                      34,
			ScToScLogEventEnableEpoch:                   35,
			MaxNodesChangeEnableEpoch: []config.MaxNodesChangeConfig{
				{
					EpochEnable:            0,
					MaxNumNodes:            1,
					NodesToShufflePerShard: 2,
				},
			},
		},
	}

	lastSnapshotTrieNodesConfig := config.GatewayMetricsConfig{
		URL: "http://localhost:8080",
	}

	expectedValues := map[string]interface{}{
		"erd_smart_contract_deploy_enable_epoch":                        uint32(1),
		"erd_built_in_functions_enable_epoch":                           uint32(2),
		"erd_relayed_transactions_enable_epoch":                         uint32(3),
		"erd_penalized_too_much_gas_enable_epoch":                       uint32(4),
		"erd_switch_jail_waiting_enable_epoch":                          uint32(5),
		"erd_switch_hysteresis_for_min_nodes_enable_epoch":              uint32(6),
		"erd_below_signed_threshold_enable_epoch":                       uint32(7),
		"erd_transaction_signed_with_txhash_enable_epoch":               uint32(8),
		"erd_meta_protection_enable_epoch":                              uint32(9),
		"erd_ahead_of_time_gas_usage_enable_epoch":                      uint32(10),
		"erd_gas_price_modifier_enable_epoch":                           uint32(11),
		"erd_repair_callback_enable_epoch":                              uint32(12),
		"erd_block_gas_and_fee_recheck_enable_epoch":                    uint32(13),
		"erd_staking_v2_enable_epoch":                                   uint32(14),
		"erd_stake_enable_epoch":                                        uint32(15),
		"erd_double_key_protection_enable_epoch":                        uint32(16),
		"erd_esdt_enable_epoch":                                         uint32(17),
		"erd_governance_enable_epoch":                                   uint32(18),
		"erd_delegation_manager_enable_epoch":                           uint32(19),
		"erd_delegation_smart_contract_enable_epoch":                    uint32(20),
		"erd_correct_last_unjailed_enable_epoch":                        uint32(21),
		"erd_balance_waiting_lists_enable_epoch":                        uint32(22),
		"erd_return_data_to_last_transfer_enable_epoch":                 uint32(23),
		"erd_sender_in_out_transfer_enable_epoch":                       uint32(24),
		"erd_relayed_transactions_v2_enable_epoch":                      uint32(25),
		"erd_unbond_tokens_v2_enable_epoch":                             uint32(26),
		"erd_save_jailed_always_enable_epoch":                           uint32(27),
		"erd_validator_to_delegation_enable_epoch":                      uint32(28),
		"erd_redelegate_below_min_check_enable_epoch":                   uint32(29),
		"erd_increment_scr_nonce_in_multi_transfer_enable_epoch":        uint32(30),
		"erd_esdt_multi_transfer_enable_epoch":                          uint32(31),
		"erd_global_mint_burn_disable_epoch":                            uint32(32),
		"erd_esdt_transfer_role_enable_epoch":                           uint32(33),
		"erd_max_nodes_change_enable_epoch":                             nil,
		"erd_total_supply":                                              "12345",
		"erd_hysteresis":                                                "0.100000",
		"erd_adaptivity":                                                "true",
		"erd_max_nodes_change_enable_epoch0_epoch_enable":               uint32(0),
		"erd_max_nodes_change_enable_epoch0_max_num_nodes":              uint32(1),
		"erd_max_nodes_change_enable_epoch0_nodes_to_shuffle_per_shard": uint32(2),
		"erd_set_guardian_feature_enable_epoch":                         uint32(34),
		"erd_set_sc_to_sc_log_event_enable_epoch":                       uint32(35),
		common.MetricGatewayMetricsEndpoint:                             "http://localhost:8080",
	}

	economicsConfig := config.EconomicsConfig{
		GlobalSettings: config.GlobalSettings{
			GenesisTotalSupply: "12345",
		},
	}

	genesisNodesConfig := &genesisMocks.NodesSetupStub{
		GetAdaptivityCalled: func() bool {
			return true
		},
		GetHysteresisCalled: func() float32 {
			return 0.1
		},
	}

	keys := make(map[string]interface{})

	ash := &statusHandler.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			keys[key] = uint32(value)
		},
		SetStringValueHandler: func(key string, value string) {
			keys[key] = value
		},
	}

	err := InitConfigMetrics(nil, cfg, economicsConfig, genesisNodesConfig, lastSnapshotTrieNodesConfig)
	require.Equal(t, ErrNilAppStatusHandler, err)

	err = InitConfigMetrics(ash, cfg, economicsConfig, genesisNodesConfig, lastSnapshotTrieNodesConfig)
	require.Nil(t, err)

	assert.Equal(t, len(expectedValues), len(keys))
	for k, v := range expectedValues {
		assert.Equal(t, v, keys[k])
	}

	genesisNodesConfig = &genesisMocks.NodesSetupStub{
		GetAdaptivityCalled: func() bool {
			return false
		},
		GetHysteresisCalled: func() float32 {
			return 0
		},
	}
	expectedValues["erd_adaptivity"] = "false"
	expectedValues["erd_hysteresis"] = "0.000000"

	err = InitConfigMetrics(ash, cfg, economicsConfig, genesisNodesConfig, lastSnapshotTrieNodesConfig)
	require.Nil(t, err)

	assert.Equal(t, expectedValues["erd_adaptivity"], keys["erd_adaptivity"])
	assert.Equal(t, expectedValues["erd_hysteresis"], keys["erd_hysteresis"])
}

func TestInitRatingsMetrics(t *testing.T) {
	t.Parallel()

	cfg := config.RatingsConfig{
		General: config.General{
			StartRating:           1,
			MaxRating:             10,
			MinRating:             0,
			SignedBlocksThreshold: 0.1,
			SelectionChances: []*config.SelectionChance{
				{
					MaxThreshold:  10,
					ChancePercent: 5,
				},
			},
		},
		ShardChain: config.ShardChain{
			RatingSteps: config.RatingSteps{
				HoursToMaxRatingFromStartRating: 10,
				ProposerValidatorImportance:     0.1,
				ProposerDecreaseFactor:          0.1,
				ValidatorDecreaseFactor:         0.1,
				ConsecutiveMissedBlocksPenalty:  0.1,
			},
		},
		MetaChain: config.MetaChain{
			RatingSteps: config.RatingSteps{
				HoursToMaxRatingFromStartRating: 10,
				ProposerValidatorImportance:     0.1,
				ProposerDecreaseFactor:          0.1,
				ValidatorDecreaseFactor:         0.1,
				ConsecutiveMissedBlocksPenalty:  0.1,
			},
		},
		PeerHonesty: config.PeerHonestyConfig{
			DecayCoefficient:             0.1,
			DecayUpdateIntervalInSeconds: 10,
			MaxScore:                     0.1,
			MinScore:                     0.1,
			BadPeerThreshold:             0.1,
			UnitValue:                    0.1,
		},
	}

	maxThresholdStr := fmt.Sprintf("%s%d%s", common.MetricRatingsGeneralSelectionChances, 0, common.SelectionChancesMaxThresholdSuffix)
	chancePercentStr := fmt.Sprintf("%s%d%s", common.MetricRatingsGeneralSelectionChances, 0, common.SelectionChancesChancePercentSuffix)

	expectedValues := map[string]interface{}{
		common.MetricRatingsGeneralStartRating:                 uint64(1),
		common.MetricRatingsGeneralMaxRating:                   uint64(10),
		common.MetricRatingsGeneralMinRating:                   uint64(0),
		common.MetricRatingsGeneralSignedBlocksThreshold:       "0.100000",
		common.MetricRatingsGeneralSelectionChances + "_count": uint64(1),
		maxThresholdStr:  uint64(10),
		chancePercentStr: uint64(5),
		common.MetricRatingsShardChainHoursToMaxRatingFromStartRating: uint64(10),
		common.MetricRatingsShardChainProposerValidatorImportance:     "0.100000",
		common.MetricRatingsShardChainProposerDecreaseFactor:          "0.100000",
		common.MetricRatingsShardChainValidatorDecreaseFactor:         "0.100000",
		common.MetricRatingsShardChainConsecutiveMissedBlocksPenalty:  "0.100000",
		common.MetricRatingsMetaChainHoursToMaxRatingFromStartRating:  uint64(10),
		common.MetricRatingsMetaChainProposerValidatorImportance:      "0.100000",
		common.MetricRatingsMetaChainProposerDecreaseFactor:           "0.100000",
		common.MetricRatingsMetaChainValidatorDecreaseFactor:          "0.100000",
		common.MetricRatingsMetaChainConsecutiveMissedBlocksPenalty:   "0.100000",
		common.MetricRatingsPeerHonestyDecayCoefficient:               "0.100000",
		common.MetricRatingsPeerHonestyDecayUpdateIntervalInSeconds:   uint64(10),
		common.MetricRatingsPeerHonestyMaxScore:                       "0.100000",
		common.MetricRatingsPeerHonestyMinScore:                       "0.100000",
		common.MetricRatingsPeerHonestyBadPeerThreshold:               "0.100000",
		common.MetricRatingsPeerHonestyUnitValue:                      "0.100000",
	}

	keys := make(map[string]interface{})

	ash := &statusHandler.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			keys[key] = value
		},
		SetStringValueHandler: func(key string, value string) {
			keys[key] = value
		},
	}

	err := InitRatingsMetrics(nil, cfg)
	require.Equal(t, ErrNilAppStatusHandler, err)

	err = InitRatingsMetrics(ash, cfg)
	require.Nil(t, err)

	assert.Equal(t, len(expectedValues), len(keys))
	for k, v := range expectedValues {
		assert.Equal(t, v, keys[k])
	}
}

func TestInitMetrics(t *testing.T) {
	t.Parallel()

	appStatusHandler := &statusHandler.AppStatusHandlerStub{}
	pubkeyString := "pub key"
	nodeType := core.NodeTypeValidator
	shardCoordinator := &testscommon.ShardsCoordinatorMock{
		NoShards: 3,
		SelfIDCalled: func() uint32 {
			return 0
		},
	}
	nodesSetup := &testscommon.NodesSetupStub{
		GetShardConsensusGroupSizeCalled: func() uint32 {
			return 63
		},
		GetMetaConsensusGroupSizeCalled: func() uint32 {
			return 400
		},
		GetRoundDurationCalled: func() uint64 {
			return 6000
		},
		MinNumberOfMetaNodesCalled: func() uint32 {
			return 401
		},
		MinNumberOfShardNodesCalled: func() uint32 {
			return 402
		},
		InitialNodesInfoCalled: func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			validators := map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
				0: {
					&shardingMocks.NodeInfoMock{},
					&shardingMocks.NodeInfoMock{},
				},
				1: {
					&shardingMocks.NodeInfoMock{},
				},
			}

			return validators, make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
		},
		GetStartTimeCalled: func() int64 {
			return 111111
		},
	}
	version := "version"
	economicsConfigs := &config.EconomicsConfig{
		RewardsSettings: config.RewardsSettings{
			RewardsConfigByEpoch: []config.EpochRewardSettings{
				{
					LeaderPercentage: 2,
				},
				{
					LeaderPercentage: 2,
				},
			},
		},
		GlobalSettings: config.GlobalSettings{
			Denomination: 4,
		},
	}
	roundsPerEpoch := int64(200)
	minTransactionVersion := uint32(1)

	t.Run("nil app status handler should error", func(t *testing.T) {
		t.Parallel()

		err := InitMetrics(nil, pubkeyString, nodeType, shardCoordinator, nodesSetup, version, economicsConfigs, roundsPerEpoch, minTransactionVersion)
		assert.Equal(t, ErrNilAppStatusHandler, err)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		expectedErrorString := "nil shard coordinator when initializing metrics"
		err := InitMetrics(appStatusHandler, pubkeyString, nodeType, nil, nodesSetup, version, economicsConfigs, roundsPerEpoch, minTransactionVersion)
		assert.Equal(t, expectedErrorString, err.Error())
	})
	t.Run("nil nodes configs should error", func(t *testing.T) {
		t.Parallel()

		expectedErrorString := "nil nodes config when initializing metrics"
		err := InitMetrics(appStatusHandler, pubkeyString, nodeType, shardCoordinator, nil, version, economicsConfigs, roundsPerEpoch, minTransactionVersion)
		assert.Equal(t, expectedErrorString, err.Error())
	})
	t.Run("nil economics configs should error", func(t *testing.T) {
		t.Parallel()

		expectedErrorString := "nil economics config when initializing metrics"
		err := InitMetrics(appStatusHandler, pubkeyString, nodeType, shardCoordinator, nodesSetup, version, nil, roundsPerEpoch, minTransactionVersion)
		assert.Equal(t, expectedErrorString, err.Error())
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		keys := make(map[string]interface{})
		localStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetUInt64ValueHandler: func(key string, value uint64) {
				keys[key] = value
			},
			SetStringValueHandler: func(key string, value string) {
				keys[key] = value
			},
		}

		err := InitMetrics(localStatusHandler, pubkeyString, nodeType, shardCoordinator, nodesSetup, version, economicsConfigs, roundsPerEpoch, minTransactionVersion)
		assert.Nil(t, err)

		expectedValues := map[string]interface{}{
			common.MetricPublicKeyBlockSign:           pubkeyString,
			common.MetricShardId:                      uint64(shardCoordinator.SelfId()),
			common.MetricNumShardsWithoutMetachain:    uint64(shardCoordinator.NoShards),
			common.MetricNodeType:                     string(nodeType),
			common.MetricRoundTime:                    uint64(6),
			common.MetricAppVersion:                   version,
			common.MetricRoundsPerEpoch:               uint64(roundsPerEpoch),
			common.MetricCrossCheckBlockHeight:        "0",
			common.MetricCrossCheckBlockHeight + "_0": uint64(0),
			common.MetricCrossCheckBlockHeight + "_1": uint64(0),
			common.MetricCrossCheckBlockHeight + "_2": uint64(0),
			common.MetricCrossCheckBlockHeightMeta:    uint64(0),
			common.MetricIsSyncing:                    uint64(1),
			common.MetricLeaderPercentage:             fmt.Sprintf("%f", 2.0),
			common.MetricDenomination:                 uint64(4),
			common.MetricShardConsensusGroupSize:      uint64(63),
			common.MetricMetaConsensusGroupSize:       uint64(400),
			common.MetricNumNodesPerShard:             uint64(402),
			common.MetricNumMetachainNodes:            uint64(401),
			common.MetricStartTime:                    uint64(111111),
			common.MetricRoundDuration:                uint64(6000),
			common.MetricMinTransactionVersion:        uint64(1),
			common.MetricNumValidators:                uint64(2),
			common.MetricConsensusGroupSize:           uint64(63),
		}

		assert.Equal(t, len(expectedValues), len(keys))
		for k, v := range expectedValues {
			assert.Equal(t, v, keys[k], fmt.Sprintf("for key %s", k))
		}
	})
	t.Run("should work - metachain", func(t *testing.T) {
		t.Parallel()

		keys := make(map[string]interface{})
		localStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetUInt64ValueHandler: func(key string, value uint64) {
				keys[key] = value
			},
			SetStringValueHandler: func(key string, value string) {
				keys[key] = value
			},
		}
		localShardCoordinator := &testscommon.ShardsCoordinatorMock{
			NoShards: 3,
			SelfIDCalled: func() uint32 {
				return common.MetachainShardId
			},
		}

		err := InitMetrics(localStatusHandler, pubkeyString, nodeType, localShardCoordinator, nodesSetup, version, economicsConfigs, roundsPerEpoch, minTransactionVersion)
		assert.Nil(t, err)

		expectedValues := map[string]interface{}{
			common.MetricPublicKeyBlockSign:           pubkeyString,
			common.MetricShardId:                      uint64(localShardCoordinator.SelfId()),
			common.MetricNumShardsWithoutMetachain:    uint64(localShardCoordinator.NoShards),
			common.MetricNodeType:                     string(nodeType),
			common.MetricRoundTime:                    uint64(6),
			common.MetricAppVersion:                   version,
			common.MetricRoundsPerEpoch:               uint64(roundsPerEpoch),
			common.MetricCrossCheckBlockHeight:        "0",
			common.MetricCrossCheckBlockHeight + "_0": uint64(0),
			common.MetricCrossCheckBlockHeight + "_1": uint64(0),
			common.MetricCrossCheckBlockHeight + "_2": uint64(0),
			common.MetricCrossCheckBlockHeightMeta:    uint64(0),
			common.MetricIsSyncing:                    uint64(1),
			common.MetricLeaderPercentage:             fmt.Sprintf("%f", 2.0),
			common.MetricDenomination:                 uint64(4),
			common.MetricShardConsensusGroupSize:      uint64(63),
			common.MetricMetaConsensusGroupSize:       uint64(400),
			common.MetricNumNodesPerShard:             uint64(402),
			common.MetricNumMetachainNodes:            uint64(401),
			common.MetricStartTime:                    uint64(111111),
			common.MetricRoundDuration:                uint64(6000),
			common.MetricMinTransactionVersion:        uint64(1),
			common.MetricNumValidators:                uint64(0),
			common.MetricConsensusGroupSize:           uint64(400),
		}

		assert.Equal(t, len(expectedValues), len(keys))
		for k, v := range expectedValues {
			assert.Equal(t, v, keys[k], fmt.Sprintf("for key %s", k))
		}
	})
	t.Run("should work - invalid shard id", func(t *testing.T) {
		t.Parallel()

		keys := make(map[string]interface{})
		localStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetUInt64ValueHandler: func(key string, value uint64) {
				keys[key] = value
			},
			SetStringValueHandler: func(key string, value string) {
				keys[key] = value
			},
		}
		localShardCoordinator := &testscommon.ShardsCoordinatorMock{
			NoShards: 3,
			SelfIDCalled: func() uint32 {
				return 10
			},
		}

		err := InitMetrics(localStatusHandler, pubkeyString, nodeType, localShardCoordinator, nodesSetup, version, economicsConfigs, roundsPerEpoch, minTransactionVersion)
		assert.Nil(t, err)

		assert.Equal(t, uint64(0), keys[common.MetricConsensusGroupSize])
	})
}

func TestSaveStringMetric(t *testing.T) {
	t.Parallel()

	t.Run("should not panic if appStatusHandler is nil", func(t *testing.T) {
		assert.NotPanics(t, func() {
			SaveStringMetric(nil, "key", "value")
		})
	})
	t.Run("should work", func(t *testing.T) {
		wasCalled := false
		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetStringValueHandler: func(key string, value string) {
				wasCalled = true
				assert.Equal(t, "key", key)
				assert.Equal(t, "value", value)
			},
		}
		SaveStringMetric(appStatusHandler, "key", "value")
		assert.True(t, wasCalled)
	})
}

func TestSaveUint64Metric(t *testing.T) {
	t.Parallel()

	t.Run("should not panic if appStatusHandler is nil", func(t *testing.T) {
		assert.NotPanics(t, func() {
			SaveUint64Metric(nil, "key", 1)
		})
	})
	t.Run("should work", func(t *testing.T) {
		wasCalled := false
		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetUInt64ValueHandler: func(key string, value uint64) {
				wasCalled = true
				assert.Equal(t, "key", key)
				assert.Equal(t, uint64(1), value)
			},
		}
		SaveUint64Metric(appStatusHandler, "key", 1)
		assert.True(t, wasCalled)
	})
}
