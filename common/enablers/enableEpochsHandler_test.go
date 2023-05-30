package enablers

import (
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
		SCDeployEnableEpoch:                               1,
		BuiltInFunctionsEnableEpoch:                       2,
		RelayedTransactionsEnableEpoch:                    3,
		PenalizedTooMuchGasEnableEpoch:                    4,
		SwitchJailWaitingEnableEpoch:                      5,
		SwitchHysteresisForMinNodesEnableEpoch:            6,
		BelowSignedThresholdEnableEpoch:                   7,
		TransactionSignedWithTxHashEnableEpoch:            8,
		MetaProtectionEnableEpoch:                         9,
		AheadOfTimeGasUsageEnableEpoch:                    10,
		GasPriceModifierEnableEpoch:                       11,
		RepairCallbackEnableEpoch:                         12,
		BlockGasAndFeesReCheckEnableEpoch:                 13,
		StakingV2EnableEpoch:                              14,
		StakeEnableEpoch:                                  15,
		DoubleKeyProtectionEnableEpoch:                    16,
		ESDTEnableEpoch:                                   17,
		GovernanceEnableEpoch:                             18,
		DelegationManagerEnableEpoch:                      19,
		DelegationSmartContractEnableEpoch:                20,
		CorrectLastUnjailedEnableEpoch:                    21,
		BalanceWaitingListsEnableEpoch:                    22,
		ReturnDataToLastTransferEnableEpoch:               23,
		SenderInOutTransferEnableEpoch:                    24,
		RelayedTransactionsV2EnableEpoch:                  25,
		UnbondTokensV2EnableEpoch:                         26,
		SaveJailedAlwaysEnableEpoch:                       27,
		ValidatorToDelegationEnableEpoch:                  28,
		ReDelegateBelowMinCheckEnableEpoch:                29,
		WaitingListFixEnableEpoch:                         30,
		IncrementSCRNonceInMultiTransferEnableEpoch:       31,
		ScheduledMiniBlocksEnableEpoch:                    32,
		ESDTMultiTransferEnableEpoch:                      33,
		GlobalMintBurnDisableEpoch:                        34,
		ESDTTransferRoleEnableEpoch:                       35,
		BuiltInFunctionOnMetaEnableEpoch:                  36,
		ComputeRewardCheckpointEnableEpoch:                37,
		SCRSizeInvariantCheckEnableEpoch:                  38,
		BackwardCompSaveKeyValueEnableEpoch:               39,
		ESDTNFTCreateOnMultiShardEnableEpoch:              40,
		MetaESDTSetEnableEpoch:                            41,
		AddTokensToDelegationEnableEpoch:                  42,
		MultiESDTTransferFixOnCallBackOnEnableEpoch:       43,
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch:       44,
		CorrectFirstQueuedEpoch:                           45,
		CorrectJailedNotUnstakedEmptyQueueEpoch:           46,
		FixOOGReturnCodeEnableEpoch:                       47,
		RemoveNonUpdatedStorageEnableEpoch:                48,
		DeleteDelegatorAfterClaimRewardsEnableEpoch:       49,
		OptimizeNFTStoreEnableEpoch:                       50,
		CreateNFTThroughExecByCallerEnableEpoch:           51,
		StopDecreasingValidatorRatingWhenStuckEnableEpoch: 52,
		FrontRunningProtectionEnableEpoch:                 53,
		IsPayableBySCEnableEpoch:                          54,
		CleanUpInformativeSCRsEnableEpoch:                 55,
		StorageAPICostOptimizationEnableEpoch:             56,
		TransformToMultiShardCreateEnableEpoch:            57,
		ESDTRegisterAndSetAllRolesEnableEpoch:             58,
		DoNotReturnOldBlockInBlockchainHookEnableEpoch:    59,
		AddFailedRelayedTxToInvalidMBsDisableEpoch:        60,
		SCRSizeInvariantOnBuiltInResultEnableEpoch:        61,
		CheckCorrectTokenIDForTransferRoleEnableEpoch:     62,
		FailExecutionOnEveryAPIErrorEnableEpoch:           63,
		MiniBlockPartialExecutionEnableEpoch:              64,
		ManagedCryptoAPIsEnableEpoch:                      65,
		ESDTMetadataContinuousCleanupEnableEpoch:          66,
		DisableExecByCallerEnableEpoch:                    67,
		RefactorContextEnableEpoch:                        68,
		CheckFunctionArgumentEnableEpoch:                  69,
		CheckExecuteOnReadOnlyEnableEpoch:                 70,
		FixAsyncCallBackArgsListEnableEpoch:               71,
		FixOldTokenLiquidityEnableEpoch:                   72,
		RuntimeMemStoreLimitEnableEpoch:                   73,
		MaxBlockchainHookCountersEnableEpoch:              74,
		WipeSingleNFTLiquidityDecreaseEnableEpoch:         75,
		AlwaysSaveTokenMetaDataEnableEpoch:                76,
		RuntimeCodeSizeFixEnableEpoch:                     77,
		MultiClaimOnDelegationEnableEpoch:                 78,
		KeepExecOrderOnCreatedSCRsEnableEpoch:             79,
		ChangeUsernameEnableEpoch:                         80,
		AutoBalanceDataTriesEnableEpoch:                   81,
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

		handler.EpochConfirmed(77, 0)

		assert.Equal(t, cfg.BlockGasAndFeesReCheckEnableEpoch, handler.BlockGasAndFeesReCheckEnableEpoch())
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
		assert.True(t, handler.IsFailExecutionOnEveryAPIErrorFlagEnabled())
		assert.True(t, handler.IsMiniBlockPartialExecutionFlagEnabled())
		assert.True(t, handler.IsManagedCryptoAPIsFlagEnabled())
		assert.True(t, handler.IsESDTMetadataContinuousCleanupFlagEnabled())
		assert.True(t, handler.IsDisableExecByCallerFlagEnabled())
		assert.True(t, handler.IsRefactorContextFlagEnabled())
		assert.True(t, handler.IsCheckFunctionArgumentFlagEnabled())
		assert.True(t, handler.IsCheckExecuteOnReadOnlyFlagEnabled())
		assert.True(t, handler.IsChangeDelegationOwnerFlagEnabled())
		assert.True(t, handler.IsFixOldTokenLiquidityEnabled())
		assert.True(t, handler.IsRuntimeMemStoreLimitEnabled())
		assert.True(t, handler.IsMaxBlockchainHookCountersFlagEnabled())
		assert.True(t, handler.IsAlwaysSaveTokenMetaDataEnabled())
		assert.True(t, handler.IsRuntimeCodeSizeFixEnabled())
		assert.True(t, handler.IsConsistentTokensValuesLengthCheckEnabled())
		assert.False(t, handler.IsKeepExecOrderOnCreatedSCRsEnabled())
		assert.False(t, handler.IsMultiClaimOnDelegationEnabled())
		assert.False(t, handler.IsChangeUsernameEnabled())
		assert.False(t, handler.IsAutoBalanceDataTriesEnabled())
	})
	t.Run("flags with == condition should be set, along with all >=", func(t *testing.T) {
		t.Parallel()

		epoch := uint32(81)
		cfg := createEnableEpochsConfig()
		cfg.StakingV2EnableEpoch = epoch
		cfg.ESDTEnableEpoch = epoch
		cfg.GovernanceEnableEpoch = epoch
		cfg.CorrectLastUnjailedEnableEpoch = epoch

		handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
		require.False(t, check.IfNil(handler))

		handler.EpochConfirmed(epoch, 0)

		assert.Equal(t, cfg.BlockGasAndFeesReCheckEnableEpoch, handler.BlockGasAndFeesReCheckEnableEpoch())
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
		assert.True(t, handler.IsFailExecutionOnEveryAPIErrorFlagEnabled())
		assert.True(t, handler.IsMiniBlockPartialExecutionFlagEnabled())
		assert.True(t, handler.IsManagedCryptoAPIsFlagEnabled())
		assert.True(t, handler.IsESDTMetadataContinuousCleanupFlagEnabled())
		assert.True(t, handler.IsDisableExecByCallerFlagEnabled())
		assert.True(t, handler.IsRefactorContextFlagEnabled())
		assert.True(t, handler.IsCheckFunctionArgumentFlagEnabled())
		assert.True(t, handler.IsCheckExecuteOnReadOnlyFlagEnabled())
		assert.True(t, handler.IsChangeDelegationOwnerFlagEnabled())
		assert.True(t, handler.IsFixAsyncCallBackArgsListFlagEnabled())
		assert.True(t, handler.IsFixOldTokenLiquidityEnabled())
		assert.True(t, handler.IsRuntimeMemStoreLimitEnabled())
		assert.True(t, handler.IsMaxBlockchainHookCountersFlagEnabled())
		assert.True(t, handler.IsWipeSingleNFTLiquidityDecreaseEnabled())
		assert.True(t, handler.IsAlwaysSaveTokenMetaDataEnabled())
		assert.True(t, handler.IsRuntimeCodeSizeFixEnabled())
		assert.True(t, handler.IsKeepExecOrderOnCreatedSCRsEnabled())
		assert.True(t, handler.IsChangeUsernameEnabled())
		assert.True(t, handler.IsAutoBalanceDataTriesEnabled())
	})
	t.Run("flags with < should be set", func(t *testing.T) {
		t.Parallel()

		epoch := uint32(0)
		cfg := createEnableEpochsConfig()
		handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
		require.False(t, check.IfNil(handler))

		handler.EpochConfirmed(epoch, 0)

		assert.Equal(t, cfg.BlockGasAndFeesReCheckEnableEpoch, handler.BlockGasAndFeesReCheckEnableEpoch())
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
		assert.False(t, handler.IsFailExecutionOnEveryAPIErrorFlagEnabled())
		assert.False(t, handler.IsMiniBlockPartialExecutionFlagEnabled())
		assert.False(t, handler.IsManagedCryptoAPIsFlagEnabled())
		assert.False(t, handler.IsESDTMetadataContinuousCleanupFlagEnabled())
		assert.False(t, handler.IsDisableExecByCallerFlagEnabled())
		assert.False(t, handler.IsRefactorContextFlagEnabled())
		assert.False(t, handler.IsCheckFunctionArgumentFlagEnabled())
		assert.False(t, handler.IsCheckExecuteOnReadOnlyFlagEnabled())
		assert.False(t, handler.IsChangeDelegationOwnerFlagEnabled())
		assert.False(t, handler.IsFixAsyncCallBackArgsListFlagEnabled())
		assert.False(t, handler.IsFixOldTokenLiquidityEnabled())
		assert.False(t, handler.IsRuntimeMemStoreLimitEnabled())
		assert.False(t, handler.IsMaxBlockchainHookCountersFlagEnabled())
		assert.False(t, handler.IsWipeSingleNFTLiquidityDecreaseEnabled())
		assert.False(t, handler.IsAlwaysSaveTokenMetaDataEnabled())
		assert.False(t, handler.IsRuntimeCodeSizeFixEnabled())
		assert.False(t, handler.IsKeepExecOrderOnCreatedSCRsEnabled())
		assert.False(t, handler.IsChangeUsernameEnabled())
		assert.False(t, handler.IsAutoBalanceDataTriesEnabled())
	})
}
