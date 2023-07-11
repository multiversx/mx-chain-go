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
		handler.EpochConfirmed(currentEpoch, 0)

		assert.True(t, handler.IsSCDeployFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsBuiltInFunctionsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRelayedTransactionsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsPenalizedTooMuchGasFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSwitchJailWaitingFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsBelowSignedThresholdFlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsSwitchHysteresisForMinNodesFlagEnabledInSpecificEpochOnly(currentEpoch)) // epoch == limit
		assert.True(t, handler.IsTransactionSignedWithTxHashFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsMetaProtectionFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsAheadOfTimeGasUsageFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsGasPriceModifierFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRepairCallbackFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsBalanceWaitingListsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsReturnDataToLastTransferFlagEnabledAfterEpoch(math.MaxUint32))
		assert.True(t, handler.IsSenderInOutTransferFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsStakeFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsStakingV2FlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsStakingV2OwnerFlagEnabledInSpecificEpochOnly(currentEpoch)) // epoch == limit
		assert.True(t, handler.IsStakingV2FlagEnabledAfterEpoch(math.MaxUint32))
		assert.True(t, handler.IsDoubleKeyProtectionFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTFlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsESDTFlagEnabledInSpecificEpochOnly(currentEpoch)) // epoch == limit
		assert.True(t, handler.IsGovernanceFlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsGovernanceFlagEnabledInSpecificEpochOnly(currentEpoch)) // epoch == limit
		assert.True(t, handler.IsDelegationManagerFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsDelegationSmartContractFlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsDelegationSmartContractFlagEnabledInSpecificEpochOnly(currentEpoch)) // epoch == limit
		assert.True(t, handler.IsCorrectLastUnJailedFlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsCorrectLastUnJailedFlagEnabledInSpecificEpochOnly(currentEpoch)) // epoch == limit
		assert.True(t, handler.IsRelayedTransactionsV2FlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsUnBondTokensV2FlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSaveJailedAlwaysFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsReDelegateBelowMinCheckFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsValidatorToDelegationFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsWaitingListFixFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsIncrementSCRNonceInMultiTransferFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTMultiTransferFlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsGlobalMintBurnFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTTransferRoleFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsBuiltInFunctionOnMetaFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsComputeRewardCheckpointFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSCRSizeInvariantCheckFlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsBackwardCompSaveKeyValueFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTNFTCreateOnMultiShardFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsMetaESDTSetFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsAddTokensToDelegationFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsMultiESDTTransferFixOnCallBackFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCorrectFirstQueuedFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsDeleteDelegatorAfterClaimRewardsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsFixOOGReturnCodeFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRemoveNonUpdatedStorageFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsOptimizeNFTStoreFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCreateNFTThroughExecByCallerFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsStopDecreasingValidatorRatingWhenStuckFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsFrontRunningProtectionFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsPayableBySCFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCleanUpInformativeSCRsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsStorageAPICostOptimizationFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTRegisterAndSetAllRolesFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsScheduledMiniBlocksFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsDoNotReturnOldBlockInBlockchainHookFlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsAddFailedRelayedTxToInvalidMBsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSCRSizeInvariantOnBuiltInResultFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCheckCorrectTokenIDForTransferRoleFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsDisableExecByCallerFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRefactorContextFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsFailExecutionOnEveryAPIErrorFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsManagedCryptoAPIsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCheckFunctionArgumentFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCheckExecuteOnReadOnlyFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTMetadataContinuousCleanupFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsChangeDelegationOwnerFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsMiniBlockPartialExecutionFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsFixAsyncCallBackArgsListFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsFixOldTokenLiquidityEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRuntimeMemStoreLimitEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSetSenderInEeiOutputTransferFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRefactorPeersMiniBlocksFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsMaxBlockchainHookCountersFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsWipeSingleNFTLiquidityDecreaseEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsAlwaysSaveTokenMetaDataEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRuntimeCodeSizeFixEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRelayedNonceFixEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSetGuardianEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsAutoBalanceDataTriesEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsKeepExecOrderOnCreatedSCRsEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsMultiClaimOnDelegationEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsChangeUsernameEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsConsistentTokensValuesLengthCheckEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsFixAsyncCallbackCheckFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSaveToSystemAccountFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCheckFrozenCollectionFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSendAlwaysFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsValueLengthCheckFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCheckTransferFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsTransferToMetaFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTNFTImprovementV1FlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.FixDelegationChangeOwnerOnAccountEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSCProcessorV2FlagEnabledInEpoch(math.MaxUint32))
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

		assert.True(t, handler.IsSCDeployFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsBuiltInFunctionsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRelayedTransactionsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsPenalizedTooMuchGasFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSwitchJailWaitingFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsBelowSignedThresholdFlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsSwitchHysteresisForMinNodesFlagEnabledInSpecificEpochOnly(math.MaxUint32)) // epoch == limit
		assert.True(t, handler.IsTransactionSignedWithTxHashFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsMetaProtectionFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsAheadOfTimeGasUsageFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsGasPriceModifierFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRepairCallbackFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsBalanceWaitingListsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsReturnDataToLastTransferFlagEnabledAfterEpoch(math.MaxUint32-1))
		assert.True(t, handler.IsSenderInOutTransferFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsStakeFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsStakingV2FlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsStakingV2OwnerFlagEnabledInSpecificEpochOnly(math.MaxUint32)) // epoch == limit
		assert.False(t, handler.IsStakingV2FlagEnabledAfterEpoch(math.MaxUint32-1))
		assert.True(t, handler.IsDoubleKeyProtectionFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTFlagEnabledInSpecificEpochOnly(math.MaxUint32)) // epoch == limit
		assert.True(t, handler.IsGovernanceFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsGovernanceFlagEnabledInSpecificEpochOnly(math.MaxUint32)) // epoch == limit
		assert.True(t, handler.IsDelegationManagerFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsDelegationSmartContractFlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsDelegationSmartContractFlagEnabledInSpecificEpochOnly(math.MaxUint32)) // epoch == limit
		assert.True(t, handler.IsCorrectLastUnJailedFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCorrectLastUnJailedFlagEnabledInSpecificEpochOnly(math.MaxUint32)) // epoch == limit
		assert.True(t, handler.IsRelayedTransactionsV2FlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsUnBondTokensV2FlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSaveJailedAlwaysFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsReDelegateBelowMinCheckFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsValidatorToDelegationFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsWaitingListFixFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsIncrementSCRNonceInMultiTransferFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTMultiTransferFlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsGlobalMintBurnFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTTransferRoleFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsBuiltInFunctionOnMetaFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsComputeRewardCheckpointFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSCRSizeInvariantCheckFlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsBackwardCompSaveKeyValueFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTNFTCreateOnMultiShardFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsMetaESDTSetFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsAddTokensToDelegationFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsMultiESDTTransferFixOnCallBackFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCorrectFirstQueuedFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsDeleteDelegatorAfterClaimRewardsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsFixOOGReturnCodeFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRemoveNonUpdatedStorageFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsOptimizeNFTStoreFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCreateNFTThroughExecByCallerFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsStopDecreasingValidatorRatingWhenStuckFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsFrontRunningProtectionFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsPayableBySCFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCleanUpInformativeSCRsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsStorageAPICostOptimizationFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTRegisterAndSetAllRolesFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsScheduledMiniBlocksFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsDoNotReturnOldBlockInBlockchainHookFlagEnabledInEpoch(math.MaxUint32))
		assert.False(t, handler.IsAddFailedRelayedTxToInvalidMBsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSCRSizeInvariantOnBuiltInResultFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCheckCorrectTokenIDForTransferRoleFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsDisableExecByCallerFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRefactorContextFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsFailExecutionOnEveryAPIErrorFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsManagedCryptoAPIsFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCheckFunctionArgumentFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCheckExecuteOnReadOnlyFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTMetadataContinuousCleanupFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsChangeDelegationOwnerFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsMiniBlockPartialExecutionFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsFixAsyncCallBackArgsListFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsFixOldTokenLiquidityEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRuntimeMemStoreLimitEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSetSenderInEeiOutputTransferFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRefactorPeersMiniBlocksFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsMaxBlockchainHookCountersFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsWipeSingleNFTLiquidityDecreaseEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsAlwaysSaveTokenMetaDataEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRuntimeCodeSizeFixEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsRelayedNonceFixEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSetGuardianEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsAutoBalanceDataTriesEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsKeepExecOrderOnCreatedSCRsEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsMultiClaimOnDelegationEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsChangeUsernameEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsConsistentTokensValuesLengthCheckEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsFixAsyncCallbackCheckFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSaveToSystemAccountFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCheckFrozenCollectionFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSendAlwaysFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsValueLengthCheckFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsCheckTransferFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsTransferToMetaFlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsESDTNFTImprovementV1FlagEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.FixDelegationChangeOwnerOnAccountEnabledInEpoch(math.MaxUint32))
		assert.True(t, handler.IsSCProcessorV2FlagEnabledInEpoch(math.MaxUint32))
	})
	t.Run("flags with < should be set", func(t *testing.T) {
		t.Parallel()

		epoch := uint32(0)
		cfg := createEnableEpochsConfig()
		handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
		require.False(t, check.IfNil(handler))

		handler.EpochConfirmed(epoch, 0)

		assert.False(t, handler.IsSCDeployFlagEnabledInEpoch(0))
		assert.False(t, handler.IsBuiltInFunctionsFlagEnabledInEpoch(0))
		assert.False(t, handler.IsRelayedTransactionsFlagEnabledInEpoch(0))
		assert.False(t, handler.IsPenalizedTooMuchGasFlagEnabledInEpoch(0))
		assert.False(t, handler.IsSwitchJailWaitingFlagEnabledInEpoch(0))
		assert.False(t, handler.IsBelowSignedThresholdFlagEnabledInEpoch(0))
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
