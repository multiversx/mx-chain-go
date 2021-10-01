package metrics

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
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
		common.MetricRoundAtEpochStart,
		common.MetricNonceAtEpochStart,
		common.MetricRoundsPassedInCurrentEpoch,
		common.MetricNoncesPassedInCurrentEpoch,
		common.MetricNumConnectedPeers,
		common.MetricEpochForEconomicsData,
		common.MetricConsensusState,
		common.MetricConsensusRoundState,
		common.MetricCurrentBlockHash,
		common.MetricNumConnectedPeersClassification,
		common.MetricLatestTagSoftwareVersion,
		common.MetricP2PNumConnectedPeersClassification,
		common.MetricP2PPeerInfo,
		common.MetricP2PIntraShardValidators,
		common.MetricP2PIntraShardObservers,
		common.MetricP2PCrossShardValidators,
		common.MetricP2PCrossShardObservers,
		common.MetricP2PFullHistoryObservers,
		common.MetricP2PUnknownPeers,
		common.MetricInflation,
		common.MetricDevRewardsInEpoch,
		common.MetricTotalFees,
	}

	keys := make(map[string]struct{})

	ash := &statusHandler.AppStatusHandlerStub{
		SetStringValueHandler: func(key string, value string) {
			okValue := value == initString || value == initZeroString
			require.True(t, okValue)
			keys[key] = struct{}{}
		},
		SetUInt64ValueHandler: func(key string, value uint64) {
			require.Equal(t, value, initUint)
			keys[key] = struct{}{}
		},
	}

	sm := &statusHandler.StatusHandlersUtilsMock{
		AppStatusHandler: ash,
	}

	err := InitBaseMetrics(nil)
	require.Equal(t, ErrNilStatusHandlerUtils, err)

	err = InitBaseMetrics(sm)
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
			BuiltInFunctionOnMetaEnableEpoch:            34,
		},
	}

	expectedValues := map[string]interface{}{
		"erd_smart_contract_deploy_enable_epoch":                 uint32(1),
		"erd_built_in_functions_enable_epoch":                    uint32(2),
		"erd_relayed_transactions_enable_epoch":                  uint32(3),
		"erd_penalized_too_much_gas_enable_epoch":                uint32(4),
		"erd_switch_jail_waiting_enable_epoch":                   uint32(5),
		"erd_switch_hysteresis_for_min_nodes_enable_epoch":       uint32(6),
		"erd_below_signed_threshold_enable_epoch":                uint32(7),
		"erd_transaction_signed_with_txhash_enable_epoch":        uint32(8),
		"erd_meta_protection_enable_epoch":                       uint32(9),
		"erd_ahead_of_time_gas_usage_enable_epoch":               uint32(10),
		"erd_gas_price_modifier_enable_epoch":                    uint32(11),
		"erd_repair_callback_enable_epoch":                       uint32(12),
		"erd_block_gas_and_fee_recheck_enable_epoch":             uint32(13),
		"erd_staking_v2_enable_epoch":                            uint32(14),
		"erd_stake_enable_epoch":                                 uint32(15),
		"erd_double_key_protection_enable_epoch":                 uint32(16),
		"erd_esdt_enable_epoch":                                  uint32(17),
		"erd_governance_enable_epoch":                            uint32(18),
		"erd_delegation_manager_enable_epoch":                    uint32(19),
		"erd_delegation_smart_contract_enable_epoch":             uint32(20),
		"erd_correct_last_unjailed_enable_epoch":                 uint32(21),
		"erd_balance_waiting_lists_enable_epoch":                 uint32(22),
		"erd_return_data_to_last_transfer_enable_epoch":          uint32(23),
		"erd_sender_in_out_transfer_enable_epoch":                uint32(24),
		"erd_relayed_transactions_v2_enable_epoch":               uint32(25),
		"erd_unbond_tokens_v2_enable_epoch":                      uint32(26),
		"erd_save_jailed_always_enable_epoch":                    uint32(27),
		"erd_validator_to_delegation_enable_epoch":               uint32(28),
		"erd_redelegate_below_min_check_enable_epoch":            uint32(29),
		"erd_increment_scr_nonce_in_multi_transfer_enable_epoch": uint32(30),
		"erd_esdt_multi_transfer_enable_epoch":                   uint32(31),
		"erd_global_mint_burn_disable_epoch":                     uint32(32),
		"erd_esdt_transfer_role_enable_epoch":                    uint32(33),
		"erd_builtin_function_on_meta_enable_epoch":              uint32(34),
		"erd_total_supply":                                       "12345",
	}

	economicsConfig := config.EconomicsConfig{
		GlobalSettings: config.GlobalSettings{
			GenesisTotalSupply: "12345",
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

	sm := &statusHandler.StatusHandlersUtilsMock{
		AppStatusHandler: ash,
	}

	err := InitConfigMetrics(nil, cfg, economicsConfig)
	require.Equal(t, ErrNilStatusHandlerUtils, err)

	err = InitConfigMetrics(sm, cfg, economicsConfig)
	require.Nil(t, err)

	assert.Equal(t, len(expectedValues), len(keys))
	for k, v := range expectedValues {
		assert.Equal(t, v, keys[k])
	}
}
