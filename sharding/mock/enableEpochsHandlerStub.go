package mock

type EnableEpochsHandlerStub struct {
	WaitingListFixEnableEpochField uint32
}

// BlockGasAndFeesReCheckEnableEpoch returns 0
func (stub *EnableEpochsHandlerStub) BlockGasAndFeesReCheckEnableEpoch() uint32 {
	return 0
}

// StakingV2EnableEpoch returns 0
func (stub *EnableEpochsHandlerStub) StakingV2EnableEpoch() uint32 {
	return 0
}

// ScheduledMiniBlocksEnableEpoch returns 0
func (stub *EnableEpochsHandlerStub) ScheduledMiniBlocksEnableEpoch() uint32 {
	return 0
}

// SwitchJailWaitingEnableEpoch returns 0
func (stub *EnableEpochsHandlerStub) SwitchJailWaitingEnableEpoch() uint32 {
	return 0
}

// BalanceWaitingListsEnableEpoch returns WaitingListFixEnableEpochField
func (stub *EnableEpochsHandlerStub) BalanceWaitingListsEnableEpoch() uint32 {
	return 0
}

// WaitingListFixEnableEpoch returns WaitingListFixEnableEpochField
func (stub *EnableEpochsHandlerStub) WaitingListFixEnableEpoch() uint32 {
	return stub.WaitingListFixEnableEpochField
}

// IsSCDeployFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsSCDeployFlagEnabled() bool {
	return false
}

// IsBuiltInFunctionsFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsBuiltInFunctionsFlagEnabled() bool {
	return false
}

// IsRelayedTransactionsFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsRelayedTransactionsFlagEnabled() bool {
	return false
}

// IsPenalizedTooMuchGasFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsPenalizedTooMuchGasFlagEnabled() bool {
	return false
}

// ResetPenalizedTooMuchGasFlag does nothing
func (stub *EnableEpochsHandlerStub) ResetPenalizedTooMuchGasFlag() {
}

// IsSwitchJailWaitingFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsSwitchJailWaitingFlagEnabled() bool {
	return false
}

// IsBelowSignedThresholdFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsBelowSignedThresholdFlagEnabled() bool {
	return false
}

// IsSwitchHysteresisForMinNodesFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsSwitchHysteresisForMinNodesFlagEnabled() bool {
	return false
}

// IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch returns false
func (stub *EnableEpochsHandlerStub) IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch() bool {
	return false
}

// IsTransactionSignedWithTxHashFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsTransactionSignedWithTxHashFlagEnabled() bool {
	return false
}

// IsMetaProtectionFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsMetaProtectionFlagEnabled() bool {
	return false
}

// IsAheadOfTimeGasUsageFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsAheadOfTimeGasUsageFlagEnabled() bool {
	return false
}

// IsGasPriceModifierFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsGasPriceModifierFlagEnabled() bool {
	return false
}

// IsRepairCallbackFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsRepairCallbackFlagEnabled() bool {
	return false
}

// IsBalanceWaitingListsFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsBalanceWaitingListsFlagEnabled() bool {
	return false
}

// IsReturnDataToLastTransferFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsReturnDataToLastTransferFlagEnabled() bool {
	return false
}

// IsSenderInOutTransferFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsSenderInOutTransferFlagEnabled() bool {
	return false
}

// IsStakeFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsStakeFlagEnabled() bool {
	return false
}

// IsStakingV2FlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsStakingV2FlagEnabled() bool {
	return false
}

// IsStakingV2OwnerFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsStakingV2OwnerFlagEnabled() bool {
	return false
}

// IsStakingV2FlagEnabledForActivationEpochCompleted returns false
func (stub *EnableEpochsHandlerStub) IsStakingV2FlagEnabledForActivationEpochCompleted() bool {
	return false
}

// IsDoubleKeyProtectionFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsDoubleKeyProtectionFlagEnabled() bool {
	return false
}

// IsESDTFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsESDTFlagEnabled() bool {
	return false
}

// IsESDTFlagEnabledForCurrentEpoch returns false
func (stub *EnableEpochsHandlerStub) IsESDTFlagEnabledForCurrentEpoch() bool {
	return false
}

// IsGovernanceFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsGovernanceFlagEnabled() bool {
	return false
}

// IsGovernanceFlagEnabledForCurrentEpoch returns false
func (stub *EnableEpochsHandlerStub) IsGovernanceFlagEnabledForCurrentEpoch() bool {
	return false
}

// IsDelegationManagerFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsDelegationManagerFlagEnabled() bool {
	return false
}

// IsDelegationSmartContractFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsDelegationSmartContractFlagEnabled() bool {
	return false
}

// IsDelegationSmartContractFlagEnabledForCurrentEpoch returns false
func (stub *EnableEpochsHandlerStub) IsDelegationSmartContractFlagEnabledForCurrentEpoch() bool {
	return false
}

// IsCorrectLastUnJailedFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsCorrectLastUnJailedFlagEnabled() bool {
	return false
}

// IsCorrectLastUnJailedFlagEnabledForCurrentEpoch returns false
func (stub *EnableEpochsHandlerStub) IsCorrectLastUnJailedFlagEnabledForCurrentEpoch() bool {
	return false
}

// IsRelayedTransactionsV2FlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsRelayedTransactionsV2FlagEnabled() bool {
	return false
}

// IsUnBondTokensV2FlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsUnBondTokensV2FlagEnabled() bool {
	return false
}

// IsSaveJailedAlwaysFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsSaveJailedAlwaysFlagEnabled() bool {
	return false
}

// IsReDelegateBelowMinCheckFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsReDelegateBelowMinCheckFlagEnabled() bool {
	return false
}

// IsValidatorToDelegationFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsValidatorToDelegationFlagEnabled() bool {
	return false
}

// IsWaitingListFixFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsWaitingListFixFlagEnabled() bool {
	return false
}

// IsIncrementSCRNonceInMultiTransferFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsIncrementSCRNonceInMultiTransferFlagEnabled() bool {
	return false
}

// IsESDTMultiTransferFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsESDTMultiTransferFlagEnabled() bool {
	return false
}

// IsGlobalMintBurnFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsGlobalMintBurnFlagEnabled() bool {
	return false
}

// IsESDTTransferRoleFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsESDTTransferRoleFlagEnabled() bool {
	return false
}

// IsBuiltInFunctionOnMetaFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsBuiltInFunctionOnMetaFlagEnabled() bool {
	return false
}

// IsComputeRewardCheckpointFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsComputeRewardCheckpointFlagEnabled() bool {
	return false
}

// IsSCRSizeInvariantCheckFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsSCRSizeInvariantCheckFlagEnabled() bool {
	return false
}

// IsBackwardCompSaveKeyValueFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsBackwardCompSaveKeyValueFlagEnabled() bool {
	return false
}

// IsESDTNFTCreateOnMultiShardFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsESDTNFTCreateOnMultiShardFlagEnabled() bool {
	return false
}

// IsMetaESDTSetFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsMetaESDTSetFlagEnabled() bool {
	return false
}

// IsAddTokensToDelegationFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsAddTokensToDelegationFlagEnabled() bool {
	return false
}

// IsMultiESDTTransferFixOnCallBackFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsMultiESDTTransferFixOnCallBackFlagEnabled() bool {
	return false
}

// IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled() bool {
	return false
}

// IsCorrectFirstQueuedFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsCorrectFirstQueuedFlagEnabled() bool {
	return false
}

// IsDeleteDelegatorAfterClaimRewardsFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsDeleteDelegatorAfterClaimRewardsFlagEnabled() bool {
	return false
}

// IsFixOOGReturnCodeFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsFixOOGReturnCodeFlagEnabled() bool {
	return false
}

// IsRemoveNonUpdatedStorageFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsRemoveNonUpdatedStorageFlagEnabled() bool {
	return false
}

// IsOptimizeNFTStoreFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsOptimizeNFTStoreFlagEnabled() bool {
	return false
}

// IsCreateNFTThroughExecByCallerFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsCreateNFTThroughExecByCallerFlagEnabled() bool {
	return false
}

// IsStopDecreasingValidatorRatingWhenStuckFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsStopDecreasingValidatorRatingWhenStuckFlagEnabled() bool {
	return false
}

// IsFrontRunningProtectionFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsFrontRunningProtectionFlagEnabled() bool {
	return false
}

// IsPayableBySCFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsPayableBySCFlagEnabled() bool {
	return false
}

// IsCleanUpInformativeSCRsFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsCleanUpInformativeSCRsFlagEnabled() bool {
	return false
}

// IsStorageAPICostOptimizationFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsStorageAPICostOptimizationFlagEnabled() bool {
	return false
}

// IsESDTRegisterAndSetAllRolesFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsESDTRegisterAndSetAllRolesFlagEnabled() bool {
	return false
}

// IsScheduledMiniBlocksFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsScheduledMiniBlocksFlagEnabled() bool {
	return false
}

// IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled() bool {
	return false
}

// IsDoNotReturnOldBlockInBlockchainHookFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsDoNotReturnOldBlockInBlockchainHookFlagEnabled() bool {
	return false
}

// IsAddFailedRelayedTxToInvalidMBsFlag returns false
func (stub *EnableEpochsHandlerStub) IsAddFailedRelayedTxToInvalidMBsFlag() bool {
	return false
}

// IsSCRSizeInvariantOnBuiltInResultFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsSCRSizeInvariantOnBuiltInResultFlagEnabled() bool {
	return false
}

// IsCheckCorrectTokenIDForTransferRoleFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsCheckCorrectTokenIDForTransferRoleFlagEnabled() bool {
	return false
}

// IsFailExecutionOnEveryAPIErrorFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsFailExecutionOnEveryAPIErrorFlagEnabled() bool {
	return false
}

// IsHeartbeatDisableFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsHeartbeatDisableFlagEnabled() bool {
	return false
}

// IsMiniBlockPartialExecutionFlagEnabled returns false
func (stub *EnableEpochsHandlerStub) IsMiniBlockPartialExecutionFlagEnabled() bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *EnableEpochsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
