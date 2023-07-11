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
		SCDeployEnableEpoch:                               1,
		BuiltInFunctionsEnableEpoch:                       2,
		RelayedTransactionsEnableEpoch:                    3,
		PenalizedTooMuchGasEnableEpoch:                    4,
		SwitchJailWaitingEnableEpoch:                      5,
		BelowSignedThresholdEnableEpoch:                   6,
		SwitchHysteresisForMinNodesEnableEpoch:            7,
		TransactionSignedWithTxHashEnableEpoch:            8,
		MetaProtectionEnableEpoch:                         9,
		AheadOfTimeGasUsageEnableEpoch:                    10,
		GasPriceModifierEnableEpoch:                       11,
		RepairCallbackEnableEpoch:                         12,
		BlockGasAndFeesReCheckEnableEpoch:                 13,
		BalanceWaitingListsEnableEpoch:                    14,
		ReturnDataToLastTransferEnableEpoch:               15,
		SenderInOutTransferEnableEpoch:                    16,
		StakeEnableEpoch:                                  17,
		StakingV2EnableEpoch:                              18,
		DoubleKeyProtectionEnableEpoch:                    19,
		ESDTEnableEpoch:                                   20,
		GovernanceEnableEpoch:                             21,
		DelegationManagerEnableEpoch:                      22,
		DelegationSmartContractEnableEpoch:                23,
		CorrectLastUnjailedEnableEpoch:                    24,
		RelayedTransactionsV2EnableEpoch:                  25,
		UnbondTokensV2EnableEpoch:                         26,
		SaveJailedAlwaysEnableEpoch:                       27,
		ReDelegateBelowMinCheckEnableEpoch:                28,
		ValidatorToDelegationEnableEpoch:                  29,
		WaitingListFixEnableEpoch:                         30,
		IncrementSCRNonceInMultiTransferEnableEpoch:       31,
		ESDTMultiTransferEnableEpoch:                      32,
		GlobalMintBurnDisableEpoch:                        33,
		ESDTTransferRoleEnableEpoch:                       34,
		BuiltInFunctionOnMetaEnableEpoch:                  35,
		ComputeRewardCheckpointEnableEpoch:                36,
		SCRSizeInvariantCheckEnableEpoch:                  37,
		BackwardCompSaveKeyValueEnableEpoch:               38,
		ESDTNFTCreateOnMultiShardEnableEpoch:              39,
		MetaESDTSetEnableEpoch:                            40,
		AddTokensToDelegationEnableEpoch:                  41,
		MultiESDTTransferFixOnCallBackOnEnableEpoch:       42,
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch:       43,
		CorrectFirstQueuedEpoch:                           44,
		DeleteDelegatorAfterClaimRewardsEnableEpoch:       45,
		FixOOGReturnCodeEnableEpoch:                       46,
		RemoveNonUpdatedStorageEnableEpoch:                47,
		OptimizeNFTStoreEnableEpoch:                       48,
		CreateNFTThroughExecByCallerEnableEpoch:           49,
		StopDecreasingValidatorRatingWhenStuckEnableEpoch: 50,
		FrontRunningProtectionEnableEpoch:                 51,
		IsPayableBySCEnableEpoch:                          52,
		CleanUpInformativeSCRsEnableEpoch:                 53,
		StorageAPICostOptimizationEnableEpoch:             54,
		TransformToMultiShardCreateEnableEpoch:            55,
		ESDTRegisterAndSetAllRolesEnableEpoch:             56,
		ScheduledMiniBlocksEnableEpoch:                    57,
		CorrectJailedNotUnstakedEmptyQueueEpoch:           58,
		DoNotReturnOldBlockInBlockchainHookEnableEpoch:    59,
		AddFailedRelayedTxToInvalidMBsDisableEpoch:        60,
		SCRSizeInvariantOnBuiltInResultEnableEpoch:        61,
		CheckCorrectTokenIDForTransferRoleEnableEpoch:     62,
		DisableExecByCallerEnableEpoch:                    63,
		RefactorContextEnableEpoch:                        64,
		FailExecutionOnEveryAPIErrorEnableEpoch:           65,
		ManagedCryptoAPIsEnableEpoch:                      66,
		CheckFunctionArgumentEnableEpoch:                  67,
		CheckExecuteOnReadOnlyEnableEpoch:                 68,
		ESDTMetadataContinuousCleanupEnableEpoch:          69,
		MiniBlockPartialExecutionEnableEpoch:              70,
		FixAsyncCallBackArgsListEnableEpoch:               71,
		FixOldTokenLiquidityEnableEpoch:                   72,
		RuntimeMemStoreLimitEnableEpoch:                   73,
		SetSenderInEeiOutputTransferEnableEpoch:           74,
		RefactorPeersMiniBlocksEnableEpoch:                75,
		MaxBlockchainHookCountersEnableEpoch:              76,
		WipeSingleNFTLiquidityDecreaseEnableEpoch:         77,
		AlwaysSaveTokenMetaDataEnableEpoch:                78,
		RuntimeCodeSizeFixEnableEpoch:                     79,
		RelayedNonceFixEnableEpoch:                        80,
		SetGuardianEnableEpoch:                            81,
		AutoBalanceDataTriesEnableEpoch:                   82,
		KeepExecOrderOnCreatedSCRsEnableEpoch:             83,
		MultiClaimOnDelegationEnableEpoch:                 84,
		ChangeUsernameEnableEpoch:                         85,
		ConsistentTokensValuesLengthCheckEnableEpoch:      86,
		FixDelegationChangeOwnerOnAccountEnableEpoch:      87,
		SCProcessorV2EnableEpoch:                          88,
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

		currentEpoch := uint32(math.MaxUint32 - 1)

		assert.True(t, handler.IsSCDeployFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsBuiltInFunctionsFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsRelayedTransactionsFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsPenalizedTooMuchGasFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsSwitchJailWaitingFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsBelowSignedThresholdFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsSwitchHysteresisForMinNodesFlagEnabledInEpoch(currentEpoch))
		assert.False(t, handler.IsSwitchHysteresisForMinNodesFlagEnabledInSpecificEpochOnly(currentEpoch)) // epoch == limit
		assert.True(t, handler.IsTransactionSignedWithTxHashFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsMetaProtectionFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsAheadOfTimeGasUsageFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsGasPriceModifierFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsRepairCallbackFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsBalanceWaitingListsFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsReturnDataToLastTransferFlagEnabledAfterEpoch(currentEpoch))
		assert.True(t, handler.IsSenderInOutTransferFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsStakeFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsStakingV2FlagEnabledInEpoch(currentEpoch))
		assert.False(t, handler.IsStakingV2OwnerFlagEnabledInSpecificEpochOnly(currentEpoch)) // epoch == limit
		assert.True(t, handler.IsStakingV2FlagEnabledAfterEpoch(currentEpoch))
		assert.True(t, handler.IsDoubleKeyProtectionFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsESDTFlagEnabledInEpoch(currentEpoch))
		assert.False(t, handler.IsESDTFlagEnabledInSpecificEpochOnly(currentEpoch)) // epoch == limit
		assert.True(t, handler.IsGovernanceFlagEnabledInEpoch(currentEpoch))
		assert.False(t, handler.IsGovernanceFlagEnabledInSpecificEpochOnly(currentEpoch)) // epoch == limit
		assert.True(t, handler.IsDelegationManagerFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsDelegationSmartContractFlagEnabledInEpoch(currentEpoch))
		assert.False(t, handler.IsDelegationSmartContractFlagEnabledInSpecificEpochOnly(currentEpoch)) // epoch == limit
		assert.True(t, handler.IsCorrectLastUnJailedFlagEnabledInEpoch(currentEpoch))
		assert.False(t, handler.IsCorrectLastUnJailedFlagEnabledInSpecificEpochOnly(currentEpoch)) // epoch == limit
		assert.True(t, handler.IsRelayedTransactionsV2FlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsUnBondTokensV2FlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsSaveJailedAlwaysFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsReDelegateBelowMinCheckFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsValidatorToDelegationFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsWaitingListFixFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsIncrementSCRNonceInMultiTransferFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsESDTMultiTransferFlagEnabledInEpoch(currentEpoch))
		assert.False(t, handler.IsGlobalMintBurnFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsESDTTransferRoleFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsBuiltInFunctionOnMetaFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsComputeRewardCheckpointFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsSCRSizeInvariantCheckFlagEnabledInEpoch(currentEpoch))
		assert.False(t, handler.IsBackwardCompSaveKeyValueFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsESDTNFTCreateOnMultiShardFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsMetaESDTSetFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsAddTokensToDelegationFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsMultiESDTTransferFixOnCallBackFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsCorrectFirstQueuedFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsDeleteDelegatorAfterClaimRewardsFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsFixOOGReturnCodeFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsRemoveNonUpdatedStorageFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsOptimizeNFTStoreFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsCreateNFTThroughExecByCallerFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsStopDecreasingValidatorRatingWhenStuckFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsFrontRunningProtectionFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsPayableBySCFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsCleanUpInformativeSCRsFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsStorageAPICostOptimizationFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsESDTRegisterAndSetAllRolesFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsScheduledMiniBlocksFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsDoNotReturnOldBlockInBlockchainHookFlagEnabledInEpoch(currentEpoch))
		assert.False(t, handler.IsAddFailedRelayedTxToInvalidMBsFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsSCRSizeInvariantOnBuiltInResultFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsCheckCorrectTokenIDForTransferRoleFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsDisableExecByCallerFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsRefactorContextFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsFailExecutionOnEveryAPIErrorFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsManagedCryptoAPIsFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsCheckFunctionArgumentFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsCheckExecuteOnReadOnlyFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsESDTMetadataContinuousCleanupFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsChangeDelegationOwnerFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsMiniBlockPartialExecutionFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsFixAsyncCallBackArgsListFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsFixOldTokenLiquidityEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsRuntimeMemStoreLimitEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsSetSenderInEeiOutputTransferFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsRefactorPeersMiniBlocksFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsMaxBlockchainHookCountersFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsWipeSingleNFTLiquidityDecreaseEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsAlwaysSaveTokenMetaDataEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsRuntimeCodeSizeFixEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsRelayedNonceFixEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsSetGuardianEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsAutoBalanceDataTriesEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsKeepExecOrderOnCreatedSCRsEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsMultiClaimOnDelegationEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsChangeUsernameEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsConsistentTokensValuesLengthCheckEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsFixAsyncCallbackCheckFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsSaveToSystemAccountFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsCheckFrozenCollectionFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsSendAlwaysFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsValueLengthCheckFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsCheckTransferFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsTransferToMetaFlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsESDTNFTImprovementV1FlagEnabledInEpoch(currentEpoch))
		assert.True(t, handler.FixDelegationChangeOwnerOnAccountEnabledInEpoch(currentEpoch))
		assert.True(t, handler.IsSCProcessorV2FlagEnabledInEpoch(currentEpoch))
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

		assert.True(t, handler.IsSCDeployFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsBuiltInFunctionsFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsRelayedTransactionsFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsPenalizedTooMuchGasFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsSwitchJailWaitingFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsBelowSignedThresholdFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsSwitchHysteresisForMinNodesFlagEnabledInEpoch(epoch))
		assert.False(t, handler.IsSwitchHysteresisForMinNodesFlagEnabledInSpecificEpochOnly(epoch)) // epoch == limit
		assert.True(t, handler.IsTransactionSignedWithTxHashFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsMetaProtectionFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsAheadOfTimeGasUsageFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsGasPriceModifierFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsRepairCallbackFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsBalanceWaitingListsFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsReturnDataToLastTransferFlagEnabledAfterEpoch(epoch-1))
		assert.True(t, handler.IsSenderInOutTransferFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsStakeFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsStakingV2FlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsStakingV2OwnerFlagEnabledInSpecificEpochOnly(epoch)) // epoch == limit
		assert.False(t, handler.IsStakingV2FlagEnabledAfterEpoch(epoch-1))
		assert.True(t, handler.IsDoubleKeyProtectionFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsESDTFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsESDTFlagEnabledInSpecificEpochOnly(epoch)) // epoch == limit
		assert.True(t, handler.IsGovernanceFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsGovernanceFlagEnabledInSpecificEpochOnly(epoch)) // epoch == limit
		assert.True(t, handler.IsDelegationManagerFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsDelegationSmartContractFlagEnabledInEpoch(epoch))
		assert.False(t, handler.IsDelegationSmartContractFlagEnabledInSpecificEpochOnly(epoch)) // epoch == limit
		assert.True(t, handler.IsCorrectLastUnJailedFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsCorrectLastUnJailedFlagEnabledInSpecificEpochOnly(epoch)) // epoch == limit
		assert.True(t, handler.IsRelayedTransactionsV2FlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsUnBondTokensV2FlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsSaveJailedAlwaysFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsReDelegateBelowMinCheckFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsValidatorToDelegationFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsWaitingListFixFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsIncrementSCRNonceInMultiTransferFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsESDTMultiTransferFlagEnabledInEpoch(epoch))
		assert.False(t, handler.IsGlobalMintBurnFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsESDTTransferRoleFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsBuiltInFunctionOnMetaFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsComputeRewardCheckpointFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsSCRSizeInvariantCheckFlagEnabledInEpoch(epoch))
		assert.False(t, handler.IsBackwardCompSaveKeyValueFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsESDTNFTCreateOnMultiShardFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsMetaESDTSetFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsAddTokensToDelegationFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsMultiESDTTransferFixOnCallBackFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsCorrectFirstQueuedFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsDeleteDelegatorAfterClaimRewardsFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsFixOOGReturnCodeFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsRemoveNonUpdatedStorageFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsOptimizeNFTStoreFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsCreateNFTThroughExecByCallerFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsStopDecreasingValidatorRatingWhenStuckFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsFrontRunningProtectionFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsPayableBySCFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsCleanUpInformativeSCRsFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsStorageAPICostOptimizationFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsESDTRegisterAndSetAllRolesFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsScheduledMiniBlocksFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsDoNotReturnOldBlockInBlockchainHookFlagEnabledInEpoch(epoch))
		assert.False(t, handler.IsAddFailedRelayedTxToInvalidMBsFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsSCRSizeInvariantOnBuiltInResultFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsCheckCorrectTokenIDForTransferRoleFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsDisableExecByCallerFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsRefactorContextFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsFailExecutionOnEveryAPIErrorFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsManagedCryptoAPIsFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsCheckFunctionArgumentFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsCheckExecuteOnReadOnlyFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsESDTMetadataContinuousCleanupFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsChangeDelegationOwnerFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsMiniBlockPartialExecutionFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsFixAsyncCallBackArgsListFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsFixOldTokenLiquidityEnabledInEpoch(epoch))
		assert.True(t, handler.IsRuntimeMemStoreLimitEnabledInEpoch(epoch))
		assert.True(t, handler.IsSetSenderInEeiOutputTransferFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsRefactorPeersMiniBlocksFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsMaxBlockchainHookCountersFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsWipeSingleNFTLiquidityDecreaseEnabledInEpoch(epoch))
		assert.True(t, handler.IsAlwaysSaveTokenMetaDataEnabledInEpoch(epoch))
		assert.True(t, handler.IsRuntimeCodeSizeFixEnabledInEpoch(epoch))
		assert.True(t, handler.IsRelayedNonceFixEnabledInEpoch(epoch))
		assert.True(t, handler.IsSetGuardianEnabledInEpoch(epoch))
		assert.True(t, handler.IsAutoBalanceDataTriesEnabledInEpoch(epoch))
		assert.True(t, handler.IsKeepExecOrderOnCreatedSCRsEnabledInEpoch(epoch))
		assert.True(t, handler.IsMultiClaimOnDelegationEnabledInEpoch(epoch))
		assert.True(t, handler.IsChangeUsernameEnabledInEpoch(epoch))
		assert.True(t, handler.IsConsistentTokensValuesLengthCheckEnabledInEpoch(epoch))
		assert.True(t, handler.IsFixAsyncCallbackCheckFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsSaveToSystemAccountFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsCheckFrozenCollectionFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsSendAlwaysFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsValueLengthCheckFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsCheckTransferFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsTransferToMetaFlagEnabledInEpoch(epoch))
		assert.True(t, handler.IsESDTNFTImprovementV1FlagEnabledInEpoch(epoch))
		assert.True(t, handler.FixDelegationChangeOwnerOnAccountEnabledInEpoch(epoch))
		assert.True(t, handler.IsSCProcessorV2FlagEnabledInEpoch(epoch))
	})
	t.Run("flags with < should be set", func(t *testing.T) {
		t.Parallel()

		cfg := createEnableEpochsConfig()
		handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
		require.False(t, check.IfNil(handler))

		assert.False(t, handler.IsSCDeployFlagEnabledInEpoch(0))
		assert.False(t, handler.IsBuiltInFunctionsFlagEnabledInEpoch(0))
		assert.False(t, handler.IsRelayedTransactionsFlagEnabledInEpoch(0))
		assert.False(t, handler.IsPenalizedTooMuchGasFlagEnabledInEpoch(0))
		assert.False(t, handler.IsSwitchJailWaitingFlagEnabledInEpoch(0))
		assert.False(t, handler.IsBelowSignedThresholdFlagEnabledInEpoch(0))
		assert.False(t, handler.IsSwitchHysteresisForMinNodesFlagEnabledInEpoch(0))
		assert.False(t, handler.IsSwitchHysteresisForMinNodesFlagEnabledInSpecificEpochOnly(0)) // epoch == limit
		assert.False(t, handler.IsTransactionSignedWithTxHashFlagEnabledInEpoch(0))
		assert.False(t, handler.IsMetaProtectionFlagEnabledInEpoch(0))
		assert.False(t, handler.IsAheadOfTimeGasUsageFlagEnabledInEpoch(0))
		assert.False(t, handler.IsGasPriceModifierFlagEnabledInEpoch(0))
		assert.False(t, handler.IsRepairCallbackFlagEnabledInEpoch(0))
		assert.False(t, handler.IsBalanceWaitingListsFlagEnabledInEpoch(0))
		assert.False(t, handler.IsReturnDataToLastTransferFlagEnabledAfterEpoch(0))
		assert.False(t, handler.IsSenderInOutTransferFlagEnabledInEpoch(0))
		assert.False(t, handler.IsStakeFlagEnabledInEpoch(0))
		assert.False(t, handler.IsStakingV2FlagEnabledInEpoch(0))
		assert.False(t, handler.IsStakingV2OwnerFlagEnabledInSpecificEpochOnly(0)) // epoch == limit
		assert.False(t, handler.IsStakingV2FlagEnabledAfterEpoch(0))
		assert.False(t, handler.IsDoubleKeyProtectionFlagEnabledInEpoch(0))
		assert.False(t, handler.IsESDTFlagEnabledInEpoch(0))
		assert.False(t, handler.IsESDTFlagEnabledInSpecificEpochOnly(0)) // epoch == limit
		assert.False(t, handler.IsGovernanceFlagEnabledInEpoch(0))
		assert.False(t, handler.IsGovernanceFlagEnabledInSpecificEpochOnly(0)) // epoch == limit
		assert.False(t, handler.IsDelegationManagerFlagEnabledInEpoch(0))
		assert.False(t, handler.IsDelegationSmartContractFlagEnabledInEpoch(0))
		assert.False(t, handler.IsDelegationSmartContractFlagEnabledInSpecificEpochOnly(0)) // epoch == limit
		assert.False(t, handler.IsCorrectLastUnJailedFlagEnabledInEpoch(0))
		assert.False(t, handler.IsCorrectLastUnJailedFlagEnabledInSpecificEpochOnly(0)) // epoch == limit
		assert.False(t, handler.IsRelayedTransactionsV2FlagEnabledInEpoch(0))
		assert.False(t, handler.IsUnBondTokensV2FlagEnabledInEpoch(0))
		assert.False(t, handler.IsSaveJailedAlwaysFlagEnabledInEpoch(0))
		assert.False(t, handler.IsReDelegateBelowMinCheckFlagEnabledInEpoch(0))
		assert.False(t, handler.IsValidatorToDelegationFlagEnabledInEpoch(0))
		assert.False(t, handler.IsWaitingListFixFlagEnabledInEpoch(0))
		assert.False(t, handler.IsIncrementSCRNonceInMultiTransferFlagEnabledInEpoch(0))
		assert.False(t, handler.IsESDTMultiTransferFlagEnabledInEpoch(0))
		assert.True(t, handler.IsGlobalMintBurnFlagEnabledInEpoch(0))
		assert.False(t, handler.IsESDTTransferRoleFlagEnabledInEpoch(0))
		assert.False(t, handler.IsBuiltInFunctionOnMetaFlagEnabledInEpoch(0))
		assert.False(t, handler.IsComputeRewardCheckpointFlagEnabledInEpoch(0))
		assert.False(t, handler.IsSCRSizeInvariantCheckFlagEnabledInEpoch(0))
		assert.True(t, handler.IsBackwardCompSaveKeyValueFlagEnabledInEpoch(0))
		assert.False(t, handler.IsESDTNFTCreateOnMultiShardFlagEnabledInEpoch(0))
		assert.False(t, handler.IsMetaESDTSetFlagEnabledInEpoch(0))
		assert.False(t, handler.IsAddTokensToDelegationFlagEnabledInEpoch(0))
		assert.False(t, handler.IsMultiESDTTransferFixOnCallBackFlagEnabledInEpoch(0))
		assert.False(t, handler.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledInEpoch(0))
		assert.False(t, handler.IsCorrectFirstQueuedFlagEnabledInEpoch(0))
		assert.False(t, handler.IsDeleteDelegatorAfterClaimRewardsFlagEnabledInEpoch(0))
		assert.False(t, handler.IsFixOOGReturnCodeFlagEnabledInEpoch(0))
		assert.False(t, handler.IsRemoveNonUpdatedStorageFlagEnabledInEpoch(0))
		assert.False(t, handler.IsOptimizeNFTStoreFlagEnabledInEpoch(0))
		assert.False(t, handler.IsCreateNFTThroughExecByCallerFlagEnabledInEpoch(0))
		assert.False(t, handler.IsStopDecreasingValidatorRatingWhenStuckFlagEnabledInEpoch(0))
		assert.False(t, handler.IsFrontRunningProtectionFlagEnabledInEpoch(0))
		assert.False(t, handler.IsPayableBySCFlagEnabledInEpoch(0))
		assert.False(t, handler.IsCleanUpInformativeSCRsFlagEnabledInEpoch(0))
		assert.False(t, handler.IsStorageAPICostOptimizationFlagEnabledInEpoch(0))
		assert.False(t, handler.IsESDTRegisterAndSetAllRolesFlagEnabledInEpoch(0))
		assert.False(t, handler.IsScheduledMiniBlocksFlagEnabledInEpoch(0))
		assert.False(t, handler.IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledInEpoch(0))
		assert.False(t, handler.IsDoNotReturnOldBlockInBlockchainHookFlagEnabledInEpoch(0))
		assert.True(t, handler.IsAddFailedRelayedTxToInvalidMBsFlag())
		assert.False(t, handler.IsSCRSizeInvariantOnBuiltInResultFlagEnabledInEpoch(0))
		assert.False(t, handler.IsCheckCorrectTokenIDForTransferRoleFlagEnabledInEpoch(0))
		assert.False(t, handler.IsDisableExecByCallerFlagEnabledInEpoch(0))
		assert.False(t, handler.IsRefactorContextFlagEnabledInEpoch(0))
		assert.False(t, handler.IsFailExecutionOnEveryAPIErrorFlagEnabledInEpoch(0))
		assert.False(t, handler.IsManagedCryptoAPIsFlagEnabledInEpoch(0))
		assert.False(t, handler.IsCheckFunctionArgumentFlagEnabledInEpoch(0))
		assert.False(t, handler.IsCheckExecuteOnReadOnlyFlagEnabledInEpoch(0))
		assert.False(t, handler.IsESDTMetadataContinuousCleanupFlagEnabledInEpoch(0))
		assert.False(t, handler.IsChangeDelegationOwnerFlagEnabledInEpoch(0))
		assert.False(t, handler.IsMiniBlockPartialExecutionFlagEnabledInEpoch(0))
		assert.False(t, handler.IsFixAsyncCallBackArgsListFlagEnabledInEpoch(0))
		assert.False(t, handler.IsFixOldTokenLiquidityEnabledInEpoch(0))
		assert.False(t, handler.IsRuntimeMemStoreLimitEnabledInEpoch(0))
		assert.False(t, handler.IsSetSenderInEeiOutputTransferFlagEnabledInEpoch(0))
		assert.False(t, handler.IsRefactorPeersMiniBlocksFlagEnabledInEpoch(0))
		assert.False(t, handler.IsMaxBlockchainHookCountersFlagEnabledInEpoch(0))
		assert.False(t, handler.IsWipeSingleNFTLiquidityDecreaseEnabledInEpoch(0))
		assert.False(t, handler.IsAlwaysSaveTokenMetaDataEnabledInEpoch(0))
		assert.False(t, handler.IsRuntimeCodeSizeFixEnabledInEpoch(0))
		assert.False(t, handler.IsRelayedNonceFixEnabledInEpoch(0))
		assert.False(t, handler.IsSetGuardianEnabledInEpoch(0))
		assert.False(t, handler.IsAutoBalanceDataTriesEnabledInEpoch(0))
		assert.False(t, handler.IsKeepExecOrderOnCreatedSCRsEnabledInEpoch(0))
		assert.False(t, handler.IsMultiClaimOnDelegationEnabledInEpoch(0))
		assert.False(t, handler.IsChangeUsernameEnabledInEpoch(0))
		assert.False(t, handler.IsConsistentTokensValuesLengthCheckEnabledInEpoch(0))
		assert.False(t, handler.IsFixAsyncCallbackCheckFlagEnabledInEpoch(0))
		assert.False(t, handler.IsSaveToSystemAccountFlagEnabledInEpoch(0))
		assert.False(t, handler.IsCheckFrozenCollectionFlagEnabledInEpoch(0))
		assert.False(t, handler.IsSendAlwaysFlagEnabledInEpoch(0))
		assert.False(t, handler.IsValueLengthCheckFlagEnabledInEpoch(0))
		assert.False(t, handler.IsCheckTransferFlagEnabledInEpoch(0))
		assert.False(t, handler.IsTransferToMetaFlagEnabledInEpoch(0))
		assert.False(t, handler.IsESDTNFTImprovementV1FlagEnabledInEpoch(0))
		assert.False(t, handler.FixDelegationChangeOwnerOnAccountEnabledInEpoch(0))
		assert.False(t, handler.IsSCProcessorV2FlagEnabledInEpoch(0))
	})
}

func TestNewEnableEpochsHandler_Getters(t *testing.T) {
	t.Parallel()

	cfg := createEnableEpochsConfig()
	handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
	require.NotNil(t, handler)

	currentEpoch := uint32(1234)
	handler.EpochConfirmed(currentEpoch, 0)

	require.Equal(t, currentEpoch, handler.GetCurrentEpoch())
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
