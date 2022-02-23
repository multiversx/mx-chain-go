package epochActivation_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process/epochActivation"
	"github.com/stretchr/testify/require"
)

func TestEpochActivation_IsEnabledInRound_IsEnabled(t *testing.T) {
	t.Parallel()
	ea, err := epochActivation.NewEpochActivation(*createMockConfig())
	require.Nil(t, err)
	require.False(t, ea.IsEnabledInEpoch("Fix1", 100))
	require.False(t, ea.IsEnabledInEpoch("Fix2", 1000))
	require.False(t, ea.IsEnabled("Fix1"))
	require.False(t, ea.IsEnabled("Fix2"))
}

func TestRoundActivation_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	ra, _ := epochActivation.NewEpochActivation(config.EnableEpochs{})
	require.True(t, ra.IsInterfaceNil())
}

func createMockConfig() *config.EnableEpochs {
	cfg := &config.EnableEpochs{
		SCDeployEnableEpoch:                               180,
		BuiltInFunctionsEnableEpoch:                       180,
		RelayedTransactionsEnableEpoch:                    180,
		PenalizedTooMuchGasEnableEpoch:                    124,
		SwitchJailWaitingEnableEpoch:                      124,
		SwitchHysteresisForMinNodesEnableEpoch:            124,
		BelowSignedThresholdEnableEpoch:                   124,
		TransactionSignedWithTxHashEnableEpoch:            239,
		MetaProtectionEnableEpoch:                         180,
		AheadOfTimeGasUsageEnableEpoch:                    180,
		GasPriceModifierEnableEpoch:                       180,
		RepairCallbackEnableEpoch:                         239,
		MaxNodesChangeEnableEpoch:                         make([]config.MaxNodesChangeConfig, 0, 0),
		BlockGasAndFeesReCheckEnableEpoch:                 239,
		StakingV2EnableEpoch:                              239,
		StakeEnableEpoch:                                  124,
		DoubleKeyProtectionEnableEpoch:                    180,
		ESDTEnableEpoch:                                   272,
		GovernanceEnableEpoch:                             1000000,
		DelegationManagerEnableEpoch:                      239,
		DelegationSmartContractEnableEpoch:                239,
		CorrectLastUnjailedEnableEpoch:                    272,
		BalanceWaitingListsEnableEpoch:                    272,
		ReturnDataToLastTransferEnableEpoch:               272,
		SenderInOutTransferEnableEpoch:                    272,
		RelayedTransactionsV2EnableEpoch:                  432,
		UnbondTokensV2EnableEpoch:                         319,
		SaveJailedAlwaysEnableEpoch:                       319,
		ValidatorToDelegationEnableEpoch:                  319,
		ReDelegateBelowMinCheckEnableEpoch:                319,
		WaitingListFixEnableEpoch:                         1000000,
		IncrementSCRNonceInMultiTransferEnableEpoch:       432,
		ScheduledMiniBlocksEnableEpoch:                    510,
		ESDTMultiTransferEnableEpoch:                      432,
		GlobalMintBurnDisableEpoch:                        432,
		ESDTTransferRoleEnableEpoch:                       432,
		BuiltInFunctionOnMetaEnableEpoch:                  1000000,
		ComputeRewardCheckpointEnableEpoch:                427,
		SCRSizeInvariantCheckEnableEpoch:                  427,
		BackwardCompSaveKeyValueEnableEpoch:               427,
		ESDTNFTCreateOnMultiShardEnableEpoch:              460,
		MetaESDTSetEnableEpoch:                            460,
		AddTokensToDelegationEnableEpoch:                  460,
		MultiESDTTransferFixOnCallBackOnEnableEpoch:       460,
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch:       460,
		CorrectFirstQueuedEpoch:                           460,
		CorrectJailedNotUnstakedEmptyQueueEpoch:           460,
		FixOOGReturnCodeEnableEpoch:                       460,
		RemoveNonUpdatedStorageEnableEpoch:                460,
		DeleteDelegatorAfterClaimRewardsEnableEpoch:       460,
		OptimizeNFTStoreEnableEpoch:                       510,
		CreateNFTThroughExecByCallerEnableEpoch:           510,
		StopDecreasingValidatorRatingWhenStuckEnableEpoch: 510,
		FrontRunningProtectionEnableEpoch:                 510,
		DisableOldTrieStorageEpoch:                        510,
		IsPayableBySCEnableEpoch:                          510,
		CleanUpInformativeSCRsEnableEpoch:                 510,
		StorageAPICostOptimizationEnableEpoch:             510,
		TransformToMultiShardCreateEnableEpoch:            510,
		ESDTRegisterAndSetAllRolesEnableEpoch:             510,
		DoNotReturnOldBlockInBlockchainHookEnableEpoch:    510,
		AddFailedRelayedTxToInvalidMBsDisableEpoch:        510,
	}
	cfg.MaxNodesChangeEnableEpoch = append(cfg.MaxNodesChangeEnableEpoch, config.MaxNodesChangeConfig{
		EpochEnable:            0,
		MaxNumNodes:            2169,
		NodesToShufflePerShard: 143,
	})
	cfg.MaxNodesChangeEnableEpoch = append(cfg.MaxNodesChangeEnableEpoch, config.MaxNodesChangeConfig{
		EpochEnable:            249,
		MaxNumNodes:            3200,
		NodesToShufflePerShard: 80,
	})
	return cfg
}
