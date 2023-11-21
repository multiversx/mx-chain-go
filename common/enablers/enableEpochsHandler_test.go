package enablers

import (
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createEnableEpochsConfig() config.EnableEpochs {
	return config.EnableEpochs{
		SCDeployEnableEpoch:                                      1,
		BuiltInFunctionsEnableEpoch:                              2,
		RelayedTransactionsEnableEpoch:                           3,
		PenalizedTooMuchGasEnableEpoch:                           4,
		SwitchJailWaitingEnableEpoch:                             5,
		BelowSignedThresholdEnableEpoch:                          6,
		SwitchHysteresisForMinNodesEnableEpoch:                   7,
		TransactionSignedWithTxHashEnableEpoch:                   8,
		MetaProtectionEnableEpoch:                                9,
		AheadOfTimeGasUsageEnableEpoch:                           10,
		GasPriceModifierEnableEpoch:                              11,
		RepairCallbackEnableEpoch:                                12,
		BlockGasAndFeesReCheckEnableEpoch:                        13,
		BalanceWaitingListsEnableEpoch:                           14,
		ReturnDataToLastTransferEnableEpoch:                      15,
		SenderInOutTransferEnableEpoch:                           16,
		StakeEnableEpoch:                                         17,
		StakingV2EnableEpoch:                                     18,
		DoubleKeyProtectionEnableEpoch:                           19,
		ESDTEnableEpoch:                                          20,
		GovernanceEnableEpoch:                                    21,
		DelegationManagerEnableEpoch:                             22,
		DelegationSmartContractEnableEpoch:                       23,
		CorrectLastUnjailedEnableEpoch:                           24,
		RelayedTransactionsV2EnableEpoch:                         25,
		UnbondTokensV2EnableEpoch:                                26,
		SaveJailedAlwaysEnableEpoch:                              27,
		ReDelegateBelowMinCheckEnableEpoch:                       28,
		ValidatorToDelegationEnableEpoch:                         29,
		WaitingListFixEnableEpoch:                                30,
		IncrementSCRNonceInMultiTransferEnableEpoch:              31,
		ESDTMultiTransferEnableEpoch:                             32,
		GlobalMintBurnDisableEpoch:                               33,
		ESDTTransferRoleEnableEpoch:                              34,
		BuiltInFunctionOnMetaEnableEpoch:                         35,
		ComputeRewardCheckpointEnableEpoch:                       36,
		SCRSizeInvariantCheckEnableEpoch:                         37,
		BackwardCompSaveKeyValueEnableEpoch:                      38,
		ESDTNFTCreateOnMultiShardEnableEpoch:                     39,
		MetaESDTSetEnableEpoch:                                   40,
		AddTokensToDelegationEnableEpoch:                         41,
		MultiESDTTransferFixOnCallBackOnEnableEpoch:              42,
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch:              43,
		CorrectFirstQueuedEpoch:                                  44,
		DeleteDelegatorAfterClaimRewardsEnableEpoch:              45,
		FixOOGReturnCodeEnableEpoch:                              46,
		RemoveNonUpdatedStorageEnableEpoch:                       47,
		OptimizeNFTStoreEnableEpoch:                              48,
		CreateNFTThroughExecByCallerEnableEpoch:                  49,
		StopDecreasingValidatorRatingWhenStuckEnableEpoch:        50,
		FrontRunningProtectionEnableEpoch:                        51,
		IsPayableBySCEnableEpoch:                                 52,
		CleanUpInformativeSCRsEnableEpoch:                        53,
		StorageAPICostOptimizationEnableEpoch:                    54,
		TransformToMultiShardCreateEnableEpoch:                   55,
		ESDTRegisterAndSetAllRolesEnableEpoch:                    56,
		ScheduledMiniBlocksEnableEpoch:                           57,
		CorrectJailedNotUnstakedEmptyQueueEpoch:                  58,
		DoNotReturnOldBlockInBlockchainHookEnableEpoch:           59,
		AddFailedRelayedTxToInvalidMBsDisableEpoch:               60,
		SCRSizeInvariantOnBuiltInResultEnableEpoch:               61,
		CheckCorrectTokenIDForTransferRoleEnableEpoch:            62,
		DisableExecByCallerEnableEpoch:                           63,
		RefactorContextEnableEpoch:                               64,
		FailExecutionOnEveryAPIErrorEnableEpoch:                  65,
		ManagedCryptoAPIsEnableEpoch:                             66,
		CheckFunctionArgumentEnableEpoch:                         67,
		CheckExecuteOnReadOnlyEnableEpoch:                        68,
		ESDTMetadataContinuousCleanupEnableEpoch:                 69,
		MiniBlockPartialExecutionEnableEpoch:                     70,
		FixAsyncCallBackArgsListEnableEpoch:                      71,
		FixOldTokenLiquidityEnableEpoch:                          72,
		RuntimeMemStoreLimitEnableEpoch:                          73,
		SetSenderInEeiOutputTransferEnableEpoch:                  74,
		RefactorPeersMiniBlocksEnableEpoch:                       75,
		MaxBlockchainHookCountersEnableEpoch:                     76,
		WipeSingleNFTLiquidityDecreaseEnableEpoch:                77,
		AlwaysSaveTokenMetaDataEnableEpoch:                       78,
		RuntimeCodeSizeFixEnableEpoch:                            79,
		RelayedNonceFixEnableEpoch:                               80,
		SetGuardianEnableEpoch:                                   81,
		AutoBalanceDataTriesEnableEpoch:                          82,
		KeepExecOrderOnCreatedSCRsEnableEpoch:                    83,
		MultiClaimOnDelegationEnableEpoch:                        84,
		ChangeUsernameEnableEpoch:                                85,
		ConsistentTokensValuesLengthCheckEnableEpoch:             86,
		FixDelegationChangeOwnerOnAccountEnableEpoch:             87,
		DeterministicSortOnValidatorsInfoEnableEpoch:             79,
		ScToScLogEventEnableEpoch:                                88,
		NFTStopCreateEnableEpoch:                                 89,
		FixGasRemainingForSaveKeyValueBuiltinFunctionEnableEpoch: 90,
	}
}

func TestNewEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil epoch notifier should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewEnableEpochsHandler(createEnableEpochsConfig(), nil)
		assert.Equal(t, process.ErrNilEpochNotifier, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		handler, err := NewEnableEpochsHandler(createEnableEpochsConfig(), &epochNotifier.EpochNotifierStub{
			RegisterNotifyHandlerCalled: func(handler vmcommon.EpochSubscriberHandler) {
				wasCalled = true
			},
		})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
		assert.True(t, wasCalled)
	})
}

func TestNewEnableEpochsHandler_EpochConfirmed(t *testing.T) {
	t.Parallel()

	t.Run("higher epoch should set only >= and > flags", func(t *testing.T) {
		t.Parallel()

		cfg := createEnableEpochsConfig()
		handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
		require.False(t, check.IfNil(handler))

		handler.EpochConfirmed(math.MaxUint32, 0)

		assert.True(t, handler.IsSCDeployFlagEnabled())
		assert.True(t, handler.IsBuiltInFunctionsFlagEnabled())
		assert.True(t, handler.IsRelayedTransactionsFlagEnabled())
		assert.True(t, handler.IsPenalizedTooMuchGasFlagEnabled())
		assert.True(t, handler.IsSwitchJailWaitingFlagEnabled())
		assert.True(t, handler.IsBelowSignedThresholdFlagEnabled())
		assert.True(t, handler.IsSwitchHysteresisForMinNodesFlagEnabled())
		assert.False(t, handler.IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.True(t, handler.IsTransactionSignedWithTxHashFlagEnabled())
		assert.True(t, handler.IsMetaProtectionFlagEnabled())
		assert.True(t, handler.IsAheadOfTimeGasUsageFlagEnabled())
		assert.True(t, handler.IsGasPriceModifierFlagEnabled())
		assert.True(t, handler.IsRepairCallbackFlagEnabled())
		assert.True(t, handler.IsBalanceWaitingListsFlagEnabled())
		assert.True(t, handler.IsReturnDataToLastTransferFlagEnabled())
		assert.True(t, handler.IsSenderInOutTransferFlagEnabled())
		assert.True(t, handler.IsStakeFlagEnabled())
		assert.True(t, handler.IsStakingV2FlagEnabled())
		assert.False(t, handler.IsStakingV2OwnerFlagEnabled()) // epoch == limit
		assert.True(t, handler.IsStakingV2FlagEnabledForActivationEpochCompleted())
		assert.True(t, handler.IsDoubleKeyProtectionFlagEnabled())
		assert.True(t, handler.IsESDTFlagEnabled())
		assert.False(t, handler.IsESDTFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.True(t, handler.IsGovernanceFlagEnabled())
		assert.False(t, handler.IsGovernanceFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.True(t, handler.IsDelegationManagerFlagEnabled())
		assert.True(t, handler.IsDelegationSmartContractFlagEnabled())
		assert.False(t, handler.IsDelegationSmartContractFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.True(t, handler.IsCorrectLastUnJailedFlagEnabled())
		assert.False(t, handler.IsCorrectLastUnJailedFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.True(t, handler.IsRelayedTransactionsV2FlagEnabled())
		assert.True(t, handler.IsUnBondTokensV2FlagEnabled())
		assert.True(t, handler.IsSaveJailedAlwaysFlagEnabled())
		assert.True(t, handler.IsReDelegateBelowMinCheckFlagEnabled())
		assert.True(t, handler.IsValidatorToDelegationFlagEnabled())
		assert.True(t, handler.IsWaitingListFixFlagEnabled())
		assert.True(t, handler.IsIncrementSCRNonceInMultiTransferFlagEnabled())
		assert.True(t, handler.IsESDTMultiTransferFlagEnabled())
		assert.False(t, handler.IsGlobalMintBurnFlagEnabled())
		assert.True(t, handler.IsESDTTransferRoleFlagEnabled())
		assert.True(t, handler.IsBuiltInFunctionOnMetaFlagEnabled())
		assert.True(t, handler.IsComputeRewardCheckpointFlagEnabled())
		assert.True(t, handler.IsSCRSizeInvariantCheckFlagEnabled())
		assert.False(t, handler.IsBackwardCompSaveKeyValueFlagEnabled())
		assert.True(t, handler.IsESDTNFTCreateOnMultiShardFlagEnabled())
		assert.True(t, handler.IsMetaESDTSetFlagEnabled())
		assert.True(t, handler.IsAddTokensToDelegationFlagEnabled())
		assert.True(t, handler.IsMultiESDTTransferFixOnCallBackFlagEnabled())
		assert.True(t, handler.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled())
		assert.True(t, handler.IsCorrectFirstQueuedFlagEnabled())
		assert.True(t, handler.IsDeleteDelegatorAfterClaimRewardsFlagEnabled())
		assert.True(t, handler.IsFixOOGReturnCodeFlagEnabled())
		assert.True(t, handler.IsRemoveNonUpdatedStorageFlagEnabled())
		assert.True(t, handler.IsOptimizeNFTStoreFlagEnabled())
		assert.True(t, handler.IsCreateNFTThroughExecByCallerFlagEnabled())
		assert.True(t, handler.IsStopDecreasingValidatorRatingWhenStuckFlagEnabled())
		assert.True(t, handler.IsFrontRunningProtectionFlagEnabled())
		assert.True(t, handler.IsPayableBySCFlagEnabled())
		assert.True(t, handler.IsCleanUpInformativeSCRsFlagEnabled())
		assert.True(t, handler.IsStorageAPICostOptimizationFlagEnabled())
		assert.True(t, handler.IsESDTRegisterAndSetAllRolesFlagEnabled())
		assert.True(t, handler.IsScheduledMiniBlocksFlagEnabled())
		assert.True(t, handler.IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled())
		assert.True(t, handler.IsDoNotReturnOldBlockInBlockchainHookFlagEnabled())
		assert.False(t, handler.IsAddFailedRelayedTxToInvalidMBsFlag())
		assert.True(t, handler.IsSCRSizeInvariantOnBuiltInResultFlagEnabled())
		assert.True(t, handler.IsCheckCorrectTokenIDForTransferRoleFlagEnabled())
		assert.True(t, handler.IsDisableExecByCallerFlagEnabled())
		assert.True(t, handler.IsRefactorContextFlagEnabled())
		assert.True(t, handler.IsFailExecutionOnEveryAPIErrorFlagEnabled())
		assert.True(t, handler.IsManagedCryptoAPIsFlagEnabled())
		assert.True(t, handler.IsCheckFunctionArgumentFlagEnabled())
		assert.True(t, handler.IsCheckExecuteOnReadOnlyFlagEnabled())
		assert.True(t, handler.IsESDTMetadataContinuousCleanupFlagEnabled())
		assert.True(t, handler.IsChangeDelegationOwnerFlagEnabled())
		assert.True(t, handler.IsMiniBlockPartialExecutionFlagEnabled())
		assert.True(t, handler.IsFixAsyncCallBackArgsListFlagEnabled())
		assert.True(t, handler.IsFixOldTokenLiquidityEnabled())
		assert.True(t, handler.IsRuntimeMemStoreLimitEnabled())
		assert.True(t, handler.IsSetSenderInEeiOutputTransferFlagEnabled())
		assert.True(t, handler.IsRefactorPeersMiniBlocksFlagEnabled())
		assert.True(t, handler.IsMaxBlockchainHookCountersFlagEnabled())
		assert.True(t, handler.IsWipeSingleNFTLiquidityDecreaseEnabled())
		assert.True(t, handler.IsAlwaysSaveTokenMetaDataEnabled())
		assert.True(t, handler.IsRuntimeCodeSizeFixEnabled())
		assert.True(t, handler.IsRelayedNonceFixEnabled())
		assert.True(t, handler.IsSetGuardianEnabled())
		assert.True(t, handler.IsDeterministicSortOnValidatorsInfoFixEnabled())
		assert.True(t, handler.IsScToScEventLogEnabled())
		assert.True(t, handler.IsAutoBalanceDataTriesEnabled())
		assert.True(t, handler.IsKeepExecOrderOnCreatedSCRsEnabled())
		assert.True(t, handler.IsMultiClaimOnDelegationEnabled())
		assert.True(t, handler.IsChangeUsernameEnabled())
		assert.True(t, handler.IsConsistentTokensValuesLengthCheckEnabled())
		assert.True(t, handler.IsFixAsyncCallbackCheckFlagEnabled())
		assert.True(t, handler.IsSaveToSystemAccountFlagEnabled())
		assert.True(t, handler.IsCheckFrozenCollectionFlagEnabled())
		assert.True(t, handler.IsSendAlwaysFlagEnabled())
		assert.True(t, handler.IsValueLengthCheckFlagEnabled())
		assert.True(t, handler.IsCheckTransferFlagEnabled())
		assert.True(t, handler.IsTransferToMetaFlagEnabled())
		assert.True(t, handler.IsESDTNFTImprovementV1FlagEnabled())
		assert.True(t, handler.FixDelegationChangeOwnerOnAccountEnabled())
		assert.True(t, handler.NFTStopCreateEnabled())
		assert.True(t, handler.FixGasRemainingForSaveKeyValueBuiltinFunctionEnabled())
	})
	t.Run("flags with == condition should not be set, the ones with >= should be set", func(t *testing.T) {
		t.Parallel()

		epoch := uint32(math.MaxUint32)
		cfg := createEnableEpochsConfig()
		cfg.StakingV2EnableEpoch = epoch
		cfg.ESDTEnableEpoch = epoch
		cfg.GovernanceEnableEpoch = epoch
		cfg.CorrectLastUnjailedEnableEpoch = epoch

		handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
		require.False(t, check.IfNil(handler))

		handler.EpochConfirmed(epoch, 0)

		assert.True(t, handler.IsSCDeployFlagEnabled())
		assert.True(t, handler.IsBuiltInFunctionsFlagEnabled())
		assert.True(t, handler.IsRelayedTransactionsFlagEnabled())
		assert.True(t, handler.IsPenalizedTooMuchGasFlagEnabled())
		assert.True(t, handler.IsSwitchJailWaitingFlagEnabled())
		assert.True(t, handler.IsBelowSignedThresholdFlagEnabled())
		assert.True(t, handler.IsSwitchHysteresisForMinNodesFlagEnabled())
		assert.False(t, handler.IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.True(t, handler.IsTransactionSignedWithTxHashFlagEnabled())
		assert.True(t, handler.IsMetaProtectionFlagEnabled())
		assert.True(t, handler.IsAheadOfTimeGasUsageFlagEnabled())
		assert.True(t, handler.IsGasPriceModifierFlagEnabled())
		assert.True(t, handler.IsRepairCallbackFlagEnabled())
		assert.True(t, handler.IsBalanceWaitingListsFlagEnabled())
		assert.True(t, handler.IsReturnDataToLastTransferFlagEnabled())
		assert.True(t, handler.IsSenderInOutTransferFlagEnabled())
		assert.True(t, handler.IsStakeFlagEnabled())
		assert.True(t, handler.IsStakingV2FlagEnabled())
		assert.True(t, handler.IsStakingV2OwnerFlagEnabled()) // epoch == limit
		assert.False(t, handler.IsStakingV2FlagEnabledForActivationEpochCompleted())
		assert.True(t, handler.IsDoubleKeyProtectionFlagEnabled())
		assert.True(t, handler.IsESDTFlagEnabled())
		assert.True(t, handler.IsESDTFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.True(t, handler.IsGovernanceFlagEnabled())
		assert.True(t, handler.IsGovernanceFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.True(t, handler.IsDelegationManagerFlagEnabled())
		assert.True(t, handler.IsDelegationSmartContractFlagEnabled())
		assert.False(t, handler.IsDelegationSmartContractFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.True(t, handler.IsCorrectLastUnJailedFlagEnabled())
		assert.True(t, handler.IsCorrectLastUnJailedFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.True(t, handler.IsRelayedTransactionsV2FlagEnabled())
		assert.True(t, handler.IsUnBondTokensV2FlagEnabled())
		assert.True(t, handler.IsSaveJailedAlwaysFlagEnabled())
		assert.True(t, handler.IsReDelegateBelowMinCheckFlagEnabled())
		assert.True(t, handler.IsValidatorToDelegationFlagEnabled())
		assert.True(t, handler.IsWaitingListFixFlagEnabled())
		assert.True(t, handler.IsIncrementSCRNonceInMultiTransferFlagEnabled())
		assert.True(t, handler.IsESDTMultiTransferFlagEnabled())
		assert.False(t, handler.IsGlobalMintBurnFlagEnabled())
		assert.True(t, handler.IsESDTTransferRoleFlagEnabled())
		assert.True(t, handler.IsBuiltInFunctionOnMetaFlagEnabled())
		assert.True(t, handler.IsComputeRewardCheckpointFlagEnabled())
		assert.True(t, handler.IsSCRSizeInvariantCheckFlagEnabled())
		assert.False(t, handler.IsBackwardCompSaveKeyValueFlagEnabled())
		assert.True(t, handler.IsESDTNFTCreateOnMultiShardFlagEnabled())
		assert.True(t, handler.IsMetaESDTSetFlagEnabled())
		assert.True(t, handler.IsAddTokensToDelegationFlagEnabled())
		assert.True(t, handler.IsMultiESDTTransferFixOnCallBackFlagEnabled())
		assert.True(t, handler.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled())
		assert.True(t, handler.IsCorrectFirstQueuedFlagEnabled())
		assert.True(t, handler.IsDeleteDelegatorAfterClaimRewardsFlagEnabled())
		assert.True(t, handler.IsFixOOGReturnCodeFlagEnabled())
		assert.True(t, handler.IsRemoveNonUpdatedStorageFlagEnabled())
		assert.True(t, handler.IsOptimizeNFTStoreFlagEnabled())
		assert.True(t, handler.IsCreateNFTThroughExecByCallerFlagEnabled())
		assert.True(t, handler.IsStopDecreasingValidatorRatingWhenStuckFlagEnabled())
		assert.True(t, handler.IsFrontRunningProtectionFlagEnabled())
		assert.True(t, handler.IsPayableBySCFlagEnabled())
		assert.True(t, handler.IsCleanUpInformativeSCRsFlagEnabled())
		assert.True(t, handler.IsStorageAPICostOptimizationFlagEnabled())
		assert.True(t, handler.IsESDTRegisterAndSetAllRolesFlagEnabled())
		assert.True(t, handler.IsScheduledMiniBlocksFlagEnabled())
		assert.True(t, handler.IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled())
		assert.True(t, handler.IsDoNotReturnOldBlockInBlockchainHookFlagEnabled())
		assert.False(t, handler.IsAddFailedRelayedTxToInvalidMBsFlag())
		assert.True(t, handler.IsSCRSizeInvariantOnBuiltInResultFlagEnabled())
		assert.True(t, handler.IsCheckCorrectTokenIDForTransferRoleFlagEnabled())
		assert.True(t, handler.IsDisableExecByCallerFlagEnabled())
		assert.True(t, handler.IsRefactorContextFlagEnabled())
		assert.True(t, handler.IsFailExecutionOnEveryAPIErrorFlagEnabled())
		assert.True(t, handler.IsManagedCryptoAPIsFlagEnabled())
		assert.True(t, handler.IsCheckFunctionArgumentFlagEnabled())
		assert.True(t, handler.IsCheckExecuteOnReadOnlyFlagEnabled())
		assert.True(t, handler.IsESDTMetadataContinuousCleanupFlagEnabled())
		assert.True(t, handler.IsChangeDelegationOwnerFlagEnabled())
		assert.True(t, handler.IsMiniBlockPartialExecutionFlagEnabled())
		assert.True(t, handler.IsFixAsyncCallBackArgsListFlagEnabled())
		assert.True(t, handler.IsFixOldTokenLiquidityEnabled())
		assert.True(t, handler.IsRuntimeMemStoreLimitEnabled())
		assert.True(t, handler.IsSetSenderInEeiOutputTransferFlagEnabled())
		assert.True(t, handler.IsRefactorPeersMiniBlocksFlagEnabled())
		assert.True(t, handler.IsMaxBlockchainHookCountersFlagEnabled())
		assert.True(t, handler.IsWipeSingleNFTLiquidityDecreaseEnabled())
		assert.True(t, handler.IsAlwaysSaveTokenMetaDataEnabled())
		assert.True(t, handler.IsRuntimeCodeSizeFixEnabled())
		assert.True(t, handler.IsRelayedNonceFixEnabled())
		assert.True(t, handler.IsSetGuardianEnabled())
		assert.True(t, handler.IsDeterministicSortOnValidatorsInfoFixEnabled())
		assert.True(t, handler.IsScToScEventLogEnabled())
		assert.True(t, handler.IsAutoBalanceDataTriesEnabled())
		assert.True(t, handler.IsKeepExecOrderOnCreatedSCRsEnabled())
		assert.True(t, handler.IsMultiClaimOnDelegationEnabled())
		assert.True(t, handler.IsChangeUsernameEnabled())
		assert.True(t, handler.IsConsistentTokensValuesLengthCheckEnabled())
		assert.True(t, handler.IsFixAsyncCallbackCheckFlagEnabled())
		assert.True(t, handler.IsSaveToSystemAccountFlagEnabled())
		assert.True(t, handler.IsCheckFrozenCollectionFlagEnabled())
		assert.True(t, handler.IsSendAlwaysFlagEnabled())
		assert.True(t, handler.IsValueLengthCheckFlagEnabled())
		assert.True(t, handler.IsCheckTransferFlagEnabled())
		assert.True(t, handler.IsTransferToMetaFlagEnabled())
		assert.True(t, handler.IsESDTNFTImprovementV1FlagEnabled())
		assert.True(t, handler.FixDelegationChangeOwnerOnAccountEnabled())
		assert.True(t, handler.NFTStopCreateEnabled())
		assert.True(t, handler.FixGasRemainingForSaveKeyValueBuiltinFunctionEnabled())
	})
	t.Run("flags with < should be set", func(t *testing.T) {
		t.Parallel()

		epoch := uint32(0)
		cfg := createEnableEpochsConfig()
		handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
		require.False(t, check.IfNil(handler))

		handler.EpochConfirmed(epoch, 0)

		assert.False(t, handler.IsSCDeployFlagEnabled())
		assert.False(t, handler.IsBuiltInFunctionsFlagEnabled())
		assert.False(t, handler.IsRelayedTransactionsFlagEnabled())
		assert.False(t, handler.IsPenalizedTooMuchGasFlagEnabled())
		assert.False(t, handler.IsSwitchJailWaitingFlagEnabled())
		assert.False(t, handler.IsBelowSignedThresholdFlagEnabled())
		assert.False(t, handler.IsSwitchHysteresisForMinNodesFlagEnabled())
		assert.False(t, handler.IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.False(t, handler.IsTransactionSignedWithTxHashFlagEnabled())
		assert.False(t, handler.IsMetaProtectionFlagEnabled())
		assert.False(t, handler.IsAheadOfTimeGasUsageFlagEnabled())
		assert.False(t, handler.IsGasPriceModifierFlagEnabled())
		assert.False(t, handler.IsRepairCallbackFlagEnabled())
		assert.False(t, handler.IsBalanceWaitingListsFlagEnabled())
		assert.False(t, handler.IsReturnDataToLastTransferFlagEnabled())
		assert.False(t, handler.IsSenderInOutTransferFlagEnabled())
		assert.False(t, handler.IsStakeFlagEnabled())
		assert.False(t, handler.IsStakingV2FlagEnabled())
		assert.False(t, handler.IsStakingV2OwnerFlagEnabled()) // epoch == limit
		assert.False(t, handler.IsStakingV2FlagEnabledForActivationEpochCompleted())
		assert.False(t, handler.IsDoubleKeyProtectionFlagEnabled())
		assert.False(t, handler.IsESDTFlagEnabled())
		assert.False(t, handler.IsESDTFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.False(t, handler.IsGovernanceFlagEnabled())
		assert.False(t, handler.IsGovernanceFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.False(t, handler.IsDelegationManagerFlagEnabled())
		assert.False(t, handler.IsDelegationSmartContractFlagEnabled())
		assert.False(t, handler.IsDelegationSmartContractFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.False(t, handler.IsCorrectLastUnJailedFlagEnabled())
		assert.False(t, handler.IsCorrectLastUnJailedFlagEnabledForCurrentEpoch()) // epoch == limit
		assert.False(t, handler.IsRelayedTransactionsV2FlagEnabled())
		assert.False(t, handler.IsUnBondTokensV2FlagEnabled())
		assert.False(t, handler.IsSaveJailedAlwaysFlagEnabled())
		assert.False(t, handler.IsReDelegateBelowMinCheckFlagEnabled())
		assert.False(t, handler.IsValidatorToDelegationFlagEnabled())
		assert.False(t, handler.IsWaitingListFixFlagEnabled())
		assert.False(t, handler.IsIncrementSCRNonceInMultiTransferFlagEnabled())
		assert.False(t, handler.IsESDTMultiTransferFlagEnabled())
		assert.True(t, handler.IsGlobalMintBurnFlagEnabled())
		assert.False(t, handler.IsESDTTransferRoleFlagEnabled())
		assert.False(t, handler.IsBuiltInFunctionOnMetaFlagEnabled())
		assert.False(t, handler.IsComputeRewardCheckpointFlagEnabled())
		assert.False(t, handler.IsSCRSizeInvariantCheckFlagEnabled())
		assert.True(t, handler.IsBackwardCompSaveKeyValueFlagEnabled())
		assert.False(t, handler.IsESDTNFTCreateOnMultiShardFlagEnabled())
		assert.False(t, handler.IsMetaESDTSetFlagEnabled())
		assert.False(t, handler.IsAddTokensToDelegationFlagEnabled())
		assert.False(t, handler.IsMultiESDTTransferFixOnCallBackFlagEnabled())
		assert.False(t, handler.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled())
		assert.False(t, handler.IsCorrectFirstQueuedFlagEnabled())
		assert.False(t, handler.IsDeleteDelegatorAfterClaimRewardsFlagEnabled())
		assert.False(t, handler.IsFixOOGReturnCodeFlagEnabled())
		assert.False(t, handler.IsRemoveNonUpdatedStorageFlagEnabled())
		assert.False(t, handler.IsOptimizeNFTStoreFlagEnabled())
		assert.False(t, handler.IsCreateNFTThroughExecByCallerFlagEnabled())
		assert.False(t, handler.IsStopDecreasingValidatorRatingWhenStuckFlagEnabled())
		assert.False(t, handler.IsFrontRunningProtectionFlagEnabled())
		assert.False(t, handler.IsPayableBySCFlagEnabled())
		assert.False(t, handler.IsCleanUpInformativeSCRsFlagEnabled())
		assert.False(t, handler.IsStorageAPICostOptimizationFlagEnabled())
		assert.False(t, handler.IsESDTRegisterAndSetAllRolesFlagEnabled())
		assert.False(t, handler.IsScheduledMiniBlocksFlagEnabled())
		assert.False(t, handler.IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled())
		assert.False(t, handler.IsDoNotReturnOldBlockInBlockchainHookFlagEnabled())
		assert.True(t, handler.IsAddFailedRelayedTxToInvalidMBsFlag())
		assert.False(t, handler.IsSCRSizeInvariantOnBuiltInResultFlagEnabled())
		assert.False(t, handler.IsCheckCorrectTokenIDForTransferRoleFlagEnabled())
		assert.False(t, handler.IsDisableExecByCallerFlagEnabled())
		assert.False(t, handler.IsRefactorContextFlagEnabled())
		assert.False(t, handler.IsFailExecutionOnEveryAPIErrorFlagEnabled())
		assert.False(t, handler.IsManagedCryptoAPIsFlagEnabled())
		assert.False(t, handler.IsCheckFunctionArgumentFlagEnabled())
		assert.False(t, handler.IsCheckExecuteOnReadOnlyFlagEnabled())
		assert.False(t, handler.IsESDTMetadataContinuousCleanupFlagEnabled())
		assert.False(t, handler.IsChangeDelegationOwnerFlagEnabled())
		assert.False(t, handler.IsMiniBlockPartialExecutionFlagEnabled())
		assert.False(t, handler.IsFixAsyncCallBackArgsListFlagEnabled())
		assert.False(t, handler.IsFixOldTokenLiquidityEnabled())
		assert.False(t, handler.IsRuntimeMemStoreLimitEnabled())
		assert.False(t, handler.IsSetSenderInEeiOutputTransferFlagEnabled())
		assert.False(t, handler.IsRefactorPeersMiniBlocksFlagEnabled())
		assert.False(t, handler.IsMaxBlockchainHookCountersFlagEnabled())
		assert.False(t, handler.IsWipeSingleNFTLiquidityDecreaseEnabled())
		assert.False(t, handler.IsAlwaysSaveTokenMetaDataEnabled())
		assert.False(t, handler.IsRuntimeCodeSizeFixEnabled())
		assert.False(t, handler.IsRelayedNonceFixEnabled())
		assert.False(t, handler.IsSetGuardianEnabled())
		assert.False(t, handler.IsDeterministicSortOnValidatorsInfoFixEnabled())
		assert.False(t, handler.IsScToScEventLogEnabled())
		assert.False(t, handler.IsAutoBalanceDataTriesEnabled())
		assert.False(t, handler.IsKeepExecOrderOnCreatedSCRsEnabled())
		assert.False(t, handler.IsMultiClaimOnDelegationEnabled())
		assert.False(t, handler.IsChangeUsernameEnabled())
		assert.False(t, handler.IsConsistentTokensValuesLengthCheckEnabled())
		assert.False(t, handler.IsFixAsyncCallbackCheckFlagEnabled())
		assert.False(t, handler.IsSaveToSystemAccountFlagEnabled())
		assert.False(t, handler.IsCheckFrozenCollectionFlagEnabled())
		assert.False(t, handler.IsSendAlwaysFlagEnabled())
		assert.False(t, handler.IsValueLengthCheckFlagEnabled())
		assert.False(t, handler.IsCheckTransferFlagEnabled())
		assert.False(t, handler.IsTransferToMetaFlagEnabled())
		assert.False(t, handler.IsESDTNFTImprovementV1FlagEnabled())
		assert.False(t, handler.FixDelegationChangeOwnerOnAccountEnabled())
		assert.False(t, handler.NFTStopCreateEnabled())
		assert.False(t, handler.FixGasRemainingForSaveKeyValueBuiltinFunctionEnabled())
	})
}

func TestNewEnableEpochsHandler_Getters(t *testing.T) {
	t.Parallel()

	cfg := createEnableEpochsConfig()
	handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
	require.NotNil(t, handler)

	require.Equal(t, cfg.ScheduledMiniBlocksEnableEpoch, handler.ScheduledMiniBlocksEnableEpoch())
	assert.Equal(t, cfg.BlockGasAndFeesReCheckEnableEpoch, handler.BlockGasAndFeesReCheckEnableEpoch())
	require.Equal(t, cfg.StakingV2EnableEpoch, handler.StakingV2EnableEpoch())
	require.Equal(t, cfg.SwitchJailWaitingEnableEpoch, handler.SwitchJailWaitingEnableEpoch())
	require.Equal(t, cfg.BalanceWaitingListsEnableEpoch, handler.BalanceWaitingListsEnableEpoch())
	require.Equal(t, cfg.WaitingListFixEnableEpoch, handler.WaitingListFixEnableEpoch())
	require.Equal(t, cfg.MultiESDTTransferFixOnCallBackOnEnableEpoch, handler.MultiESDTTransferAsyncCallBackEnableEpoch())
	require.Equal(t, cfg.FixOOGReturnCodeEnableEpoch, handler.FixOOGReturnCodeEnableEpoch())
	require.Equal(t, cfg.RemoveNonUpdatedStorageEnableEpoch, handler.RemoveNonUpdatedStorageEnableEpoch())
	require.Equal(t, cfg.CreateNFTThroughExecByCallerEnableEpoch, handler.CreateNFTThroughExecByCallerEnableEpoch())
	require.Equal(t, cfg.FailExecutionOnEveryAPIErrorEnableEpoch, handler.FixFailExecutionOnErrorEnableEpoch())
	require.Equal(t, cfg.ManagedCryptoAPIsEnableEpoch, handler.ManagedCryptoAPIEnableEpoch())
	require.Equal(t, cfg.DisableExecByCallerEnableEpoch, handler.DisableExecByCallerEnableEpoch())
	require.Equal(t, cfg.RefactorContextEnableEpoch, handler.RefactorContextEnableEpoch())
	require.Equal(t, cfg.CheckExecuteOnReadOnlyEnableEpoch, handler.CheckExecuteReadOnlyEnableEpoch())
	require.Equal(t, cfg.StorageAPICostOptimizationEnableEpoch, handler.StorageAPICostOptimizationEnableEpoch())
	require.Equal(t, cfg.MiniBlockPartialExecutionEnableEpoch, handler.MiniBlockPartialExecutionEnableEpoch())
	require.Equal(t, cfg.RefactorPeersMiniBlocksEnableEpoch, handler.RefactorPeersMiniBlocksEnableEpoch())
	require.Equal(t, cfg.RelayedNonceFixEnableEpoch, handler.RelayedNonceFixEnableEpoch())
}

func TestEnableEpochsHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var handler *enableEpochsHandler
	require.True(t, handler.IsInterfaceNil())

	handler, _ = NewEnableEpochsHandler(createEnableEpochsConfig(), &epochNotifier.EpochNotifierStub{})
	require.False(t, handler.IsInterfaceNil())
}
