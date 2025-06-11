package metrics

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
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
		common.MetricBlockTimestamp,
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
		common.MetricRoundAtEpochStart,
		common.MetricNonceAtEpochStart,
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
			SCDeployEnableEpoch:                                      1,
			BuiltInFunctionsEnableEpoch:                              2,
			RelayedTransactionsEnableEpoch:                           3,
			PenalizedTooMuchGasEnableEpoch:                           4,
			SwitchJailWaitingEnableEpoch:                             5,
			SwitchHysteresisForMinNodesEnableEpoch:                   6,
			BelowSignedThresholdEnableEpoch:                          7,
			TransactionSignedWithTxHashEnableEpoch:                   8,
			MetaProtectionEnableEpoch:                                9,
			AheadOfTimeGasUsageEnableEpoch:                           10,
			GasPriceModifierEnableEpoch:                              11,
			RepairCallbackEnableEpoch:                                12,
			BlockGasAndFeesReCheckEnableEpoch:                        13,
			StakingV2EnableEpoch:                                     14,
			StakeEnableEpoch:                                         15,
			DoubleKeyProtectionEnableEpoch:                           16,
			ESDTEnableEpoch:                                          17,
			GovernanceEnableEpoch:                                    18,
			DelegationManagerEnableEpoch:                             19,
			DelegationSmartContractEnableEpoch:                       20,
			CorrectLastUnjailedEnableEpoch:                           21,
			BalanceWaitingListsEnableEpoch:                           22,
			ReturnDataToLastTransferEnableEpoch:                      23,
			SenderInOutTransferEnableEpoch:                           24,
			RelayedTransactionsV2EnableEpoch:                         25,
			UnbondTokensV2EnableEpoch:                                26,
			SaveJailedAlwaysEnableEpoch:                              27,
			ValidatorToDelegationEnableEpoch:                         28,
			ReDelegateBelowMinCheckEnableEpoch:                       29,
			IncrementSCRNonceInMultiTransferEnableEpoch:              30,
			ScheduledMiniBlocksEnableEpoch:                           31,
			ESDTMultiTransferEnableEpoch:                             32,
			GlobalMintBurnDisableEpoch:                               33,
			ESDTTransferRoleEnableEpoch:                              34,
			ComputeRewardCheckpointEnableEpoch:                       35,
			SCRSizeInvariantCheckEnableEpoch:                         36,
			BackwardCompSaveKeyValueEnableEpoch:                      37,
			ESDTNFTCreateOnMultiShardEnableEpoch:                     38,
			MetaESDTSetEnableEpoch:                                   39,
			AddTokensToDelegationEnableEpoch:                         40,
			MultiESDTTransferFixOnCallBackOnEnableEpoch:              41,
			OptimizeGasUsedInCrossMiniBlocksEnableEpoch:              42,
			CorrectFirstQueuedEpoch:                                  43,
			CorrectJailedNotUnstakedEmptyQueueEpoch:                  44,
			FixOOGReturnCodeEnableEpoch:                              45,
			RemoveNonUpdatedStorageEnableEpoch:                       46,
			DeleteDelegatorAfterClaimRewardsEnableEpoch:              47,
			OptimizeNFTStoreEnableEpoch:                              48,
			CreateNFTThroughExecByCallerEnableEpoch:                  49,
			StopDecreasingValidatorRatingWhenStuckEnableEpoch:        50,
			FrontRunningProtectionEnableEpoch:                        51,
			IsPayableBySCEnableEpoch:                                 52,
			CleanUpInformativeSCRsEnableEpoch:                        53,
			StorageAPICostOptimizationEnableEpoch:                    54,
			TransformToMultiShardCreateEnableEpoch:                   55,
			ESDTRegisterAndSetAllRolesEnableEpoch:                    56,
			DoNotReturnOldBlockInBlockchainHookEnableEpoch:           57,
			AddFailedRelayedTxToInvalidMBsDisableEpoch:               58,
			SCRSizeInvariantOnBuiltInResultEnableEpoch:               59,
			CheckCorrectTokenIDForTransferRoleEnableEpoch:            60,
			DisableExecByCallerEnableEpoch:                           61,
			FailExecutionOnEveryAPIErrorEnableEpoch:                  62,
			ManagedCryptoAPIsEnableEpoch:                             63,
			RefactorContextEnableEpoch:                               64,
			CheckFunctionArgumentEnableEpoch:                         65,
			CheckExecuteOnReadOnlyEnableEpoch:                        66,
			MiniBlockPartialExecutionEnableEpoch:                     67,
			ESDTMetadataContinuousCleanupEnableEpoch:                 68,
			FixAsyncCallBackArgsListEnableEpoch:                      69,
			FixOldTokenLiquidityEnableEpoch:                          70,
			RuntimeMemStoreLimitEnableEpoch:                          71,
			RuntimeCodeSizeFixEnableEpoch:                            72,
			SetSenderInEeiOutputTransferEnableEpoch:                  73,
			RefactorPeersMiniBlocksEnableEpoch:                       74,
			SCProcessorV2EnableEpoch:                                 75,
			MaxBlockchainHookCountersEnableEpoch:                     76,
			WipeSingleNFTLiquidityDecreaseEnableEpoch:                77,
			AlwaysSaveTokenMetaDataEnableEpoch:                       78,
			SetGuardianEnableEpoch:                                   79,
			RelayedNonceFixEnableEpoch:                               80,
			DeterministicSortOnValidatorsInfoEnableEpoch:             81,
			KeepExecOrderOnCreatedSCRsEnableEpoch:                    82,
			MultiClaimOnDelegationEnableEpoch:                        83,
			ChangeUsernameEnableEpoch:                                84,
			AutoBalanceDataTriesEnableEpoch:                          85,
			MigrateDataTrieEnableEpoch:                               86,
			ConsistentTokensValuesLengthCheckEnableEpoch:             87,
			FixDelegationChangeOwnerOnAccountEnableEpoch:             88,
			DynamicGasCostForDataTrieStorageLoadEnableEpoch:          89,
			NFTStopCreateEnableEpoch:                                 90,
			ChangeOwnerAddressCrossShardThroughSCEnableEpoch:         91,
			FixGasRemainingForSaveKeyValueBuiltinFunctionEnableEpoch: 92,
			CurrentRandomnessOnSortingEnableEpoch:                    93,
			StakeLimitsEnableEpoch:                                   94,
			StakingV4Step1EnableEpoch:                                95,
			StakingV4Step2EnableEpoch:                                96,
			StakingV4Step3EnableEpoch:                                97,
			CleanupAuctionOnLowWaitingListEnableEpoch:                98,
			AlwaysMergeContextsInEEIEnableEpoch:                      99,
			DynamicESDTEnableEpoch:                                   100,
			EGLDInMultiTransferEnableEpoch:                           101,
			CryptoOpcodesV2EnableEpoch:                               102,
			ScToScLogEventEnableEpoch:                                103,
			FixRelayedBaseCostEnableEpoch:                            104,
			MultiESDTNFTTransferAndExecuteByUserEnableEpoch:          105,
			FixRelayedMoveBalanceToNonPayableSCEnableEpoch:           106,
			RelayedTransactionsV3EnableEpoch:                         107,
			RelayedTransactionsV3FixESDTTransferEnableEpoch:          108,
			CheckBuiltInCallOnTransferValueAndFailEnableRound:        109,
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
		"erd_smart_contract_deploy_enable_epoch":                               uint32(1),
		"erd_built_in_functions_enable_epoch":                                  uint32(2),
		"erd_relayed_transactions_enable_epoch":                                uint32(3),
		"erd_penalized_too_much_gas_enable_epoch":                              uint32(4),
		"erd_switch_jail_waiting_enable_epoch":                                 uint32(5),
		"erd_switch_hysteresis_for_min_nodes_enable_epoch":                     uint32(6),
		"erd_below_signed_threshold_enable_epoch":                              uint32(7),
		"erd_transaction_signed_with_txhash_enable_epoch":                      uint32(8),
		"erd_meta_protection_enable_epoch":                                     uint32(9),
		"erd_ahead_of_time_gas_usage_enable_epoch":                             uint32(10),
		"erd_gas_price_modifier_enable_epoch":                                  uint32(11),
		"erd_repair_callback_enable_epoch":                                     uint32(12),
		"erd_block_gas_and_fee_recheck_enable_epoch":                           uint32(13),
		"erd_staking_v2_enable_epoch":                                          uint32(14),
		"erd_stake_enable_epoch":                                               uint32(15),
		"erd_double_key_protection_enable_epoch":                               uint32(16),
		"erd_esdt_enable_epoch":                                                uint32(17),
		"erd_governance_enable_epoch":                                          uint32(18),
		"erd_delegation_manager_enable_epoch":                                  uint32(19),
		"erd_delegation_smart_contract_enable_epoch":                           uint32(20),
		"erd_correct_last_unjailed_enable_epoch":                               uint32(21),
		"erd_balance_waiting_lists_enable_epoch":                               uint32(22),
		"erd_return_data_to_last_transfer_enable_epoch":                        uint32(23),
		"erd_sender_in_out_transfer_enable_epoch":                              uint32(24),
		"erd_relayed_transactions_v2_enable_epoch":                             uint32(25),
		"erd_unbond_tokens_v2_enable_epoch":                                    uint32(26),
		"erd_save_jailed_always_enable_epoch":                                  uint32(27),
		"erd_validator_to_delegation_enable_epoch":                             uint32(28),
		"erd_redelegate_below_min_check_enable_epoch":                          uint32(29),
		"erd_increment_scr_nonce_in_multi_transfer_enable_epoch":               uint32(30),
		"erd_scheduled_miniblocks_enable_epoch":                                uint32(31),
		"erd_esdt_multi_transfer_enable_epoch":                                 uint32(32),
		"erd_global_mint_burn_disable_epoch":                                   uint32(33),
		"erd_compute_reward_checkpoint_enable_epoch":                           uint32(35),
		"erd_esdt_transfer_role_enable_epoch":                                  uint32(34),
		"erd_scr_size_invariant_check_enable_epoch":                            uint32(36),
		"erd_backward_comp_save_keyvalue_enable_epoch":                         uint32(37),
		"erd_esdt_nft_create_on_multi_shard_enable_epoch":                      uint32(38),
		"erd_meta_esdt_set_enable_epoch":                                       uint32(39),
		"erd_add_tokens_to_delegation_enable_epoch":                            uint32(40),
		"erd_multi_esdt_transfer_fix_on_callback_enable_epoch":                 uint32(41),
		"erd_optimize_gas_used_in_cross_miniblocks_enable_epoch":               uint32(42),
		"erd_correct_first_queued_enable_epoch":                                uint32(43),
		"erd_correct_jailed_not_unstaked_empty_queue_enable_epoch":             uint32(44),
		"erd_fix_oog_return_code_enable_epoch":                                 uint32(45),
		"erd_remove_non_updated_storage_enable_epoch":                          uint32(46),
		"erd_delete_delegator_after_claim_rewards_enable_epoch":                uint32(47),
		"erd_optimize_nft_store_enable_epoch":                                  uint32(48),
		"erd_create_nft_through_exec_by_caller_enable_epoch":                   uint32(49),
		"erd_stop_decreasing_validator_rating_when_stuck_enable_epoch":         uint32(50),
		"erd_front_running_protection_enable_epoch":                            uint32(51),
		"erd_is_payable_by_sc_enable_epoch":                                    uint32(52),
		"erd_cleanup_informative_scrs_enable_epoch":                            uint32(53),
		"erd_storage_api_cost_optimization_enable_epoch":                       uint32(54),
		"erd_transform_to_multi_shard_create_enable_epoch":                     uint32(55),
		"erd_esdt_register_and_set_all_roles_enable_epoch":                     uint32(56),
		"erd_do_not_returns_old_block_in_blockchain_hook_enable_epoch":         uint32(57),
		"erd_add_failed_relayed_tx_to_invalid_mbs_enable_epoch":                uint32(58),
		"erd_scr_size_invariant_on_builtin_result_enable_epoch":                uint32(59),
		"erd_check_correct_tokenid_for_transfer_role_enable_epoch":             uint32(60),
		"erd_disable_exec_by_caller_enable_epoch":                              uint32(61),
		"erd_fail_execution_on_every_api_error_enable_epoch":                   uint32(62),
		"erd_managed_crypto_apis_enable_epoch":                                 uint32(63),
		"erd_refactor_context_enable_epoch":                                    uint32(64),
		"erd_check_function_argument_enable_epoch":                             uint32(65),
		"erd_check_execute_on_readonly_enable_epoch":                           uint32(66),
		"erd_miniblock_partial_execution_enable_epoch":                         uint32(67),
		"erd_esdt_metadata_continuous_cleanup_enable_epoch":                    uint32(68),
		"erd_fix_async_callback_args_list_enable_epoch":                        uint32(69),
		"erd_fix_old_token_liquidity_enable_epoch":                             uint32(70),
		"erd_runtime_mem_store_limit_enable_epoch":                             uint32(71),
		"erd_runtime_code_size_fix_enable_epoch":                               uint32(72),
		"erd_set_sender_in_eei_output_transfer_enable_epoch":                   uint32(73),
		"erd_refactor_peers_miniblocks_enable_epoch":                           uint32(74),
		"erd_sc_processorv2_enable_epoch":                                      uint32(75),
		"erd_max_blockchain_hook_counters_enable_epoch":                        uint32(76),
		"erd_wipe_single_nft_liquidity_decrease_enable_epoch":                  uint32(77),
		"erd_always_save_token_metadata_enable_epoch":                          uint32(78),
		"erd_set_guardian_feature_enable_epoch":                                uint32(79),
		"erd_relayed_nonce_fix_enable_epoch":                                   uint32(80),
		"erd_deterministic_sort_on_validators_info_enable_epoch":               uint32(81),
		"erd_keep_exec_order_on_created_scrs_enable_epoch":                     uint32(82),
		"erd_multi_claim_on_delegation_enable_epoch":                           uint32(83),
		"erd_change_username_enable_epoch":                                     uint32(84),
		"erd_auto_balance_data_tries_enable_epoch":                             uint32(85),
		"erd_migrate_datatrie_enable_epoch":                                    uint32(86),
		"erd_consistent_tokens_values_length_check_enable_epoch":               uint32(87),
		"erd_fix_delegation_change_owner_on_account_enable_epoch":              uint32(88),
		"erd_dynamic_gas_cost_for_datatrie_storage_load_enable_epoch":          uint32(89),
		"erd_nft_stop_create_enable_epoch":                                     uint32(90),
		"erd_change_owner_address_cross_shard_through_sc_enable_epoch":         uint32(91),
		"erd_fix_gas_remainig_for_save_keyvalue_builtin_function_enable_epoch": uint32(92),
		"erd_current_randomness_on_sorting_enable_epoch":                       uint32(93),
		"erd_stake_limits_enable_epoch":                                        uint32(94),
		"erd_staking_v4_step1_enable_epoch":                                    uint32(95),
		"erd_staking_v4_step2_enable_epoch":                                    uint32(96),
		"erd_staking_v4_step3_enable_epoch":                                    uint32(97),
		"erd_cleanup_auction_on_low_waiting_list_enable_epoch":                 uint32(98),
		"erd_always_merge_contexts_in_eei_enable_epoch":                        uint32(99),
		"erd_dynamic_esdt_enable_epoch":                                        uint32(100),
		"erd_egld_in_multi_transfer_enable_epoch":                              uint32(101),
		"erd_crypto_opcodes_v2_enable_epoch":                                   uint32(102),
		"erd_set_sc_to_sc_log_event_enable_epoch":                              uint32(103),
		"erd_fix_relayed_base_cost_enable_epoch":                               uint32(104),
		"erd_multi_esdt_transfer_execute_by_user_enable_epoch":                 uint32(105),
		"erd_fix_relayed_move_balance_to_non_payable_sc_enable_epoch":          uint32(106),
		"erd_relayed_transactions_v3_enable_epoch":                             uint32(107),
		"erd_relayed_transactions_v3_fix_esdt_transfer_enable_epoch":           uint32(108),
		"erd_checkbuiltincall_ontransfervalueandfail_enable_round":             uint32(109),
		"erd_max_nodes_change_enable_epoch":                                    nil,
		"erd_total_supply":                                                     "12345",
		"erd_hysteresis":                                                       "0.100000",
		"erd_adaptivity":                                                       "true",
		"erd_max_nodes_change_enable_epoch0_epoch_enable":                      uint32(0),
		"erd_max_nodes_change_enable_epoch0_max_num_nodes":                     uint32(1),
		"erd_max_nodes_change_enable_epoch0_nodes_to_shuffle_per_shard":        uint32(2),
		common.MetricGatewayMetricsEndpoint:                                    "http://localhost:8080",
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
			RatingStepsByEpoch: []config.RatingSteps{
				{
					HoursToMaxRatingFromStartRating: 10,
					ProposerValidatorImportance:     0.1,
					ProposerDecreaseFactor:          0.1,
					ValidatorDecreaseFactor:         0.1,
					ConsecutiveMissedBlocksPenalty:  0.1,
				},
			},
		},
		MetaChain: config.MetaChain{
			RatingStepsByEpoch: []config.RatingSteps{
				{
					HoursToMaxRatingFromStartRating: 10,
					ProposerValidatorImportance:     0.1,
					ProposerDecreaseFactor:          0.1,
					ValidatorDecreaseFactor:         0.1,
					ConsecutiveMissedBlocksPenalty:  0.1,
				},
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
	nodesSetup := &genesisMocks.NodesSetupStub{
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
