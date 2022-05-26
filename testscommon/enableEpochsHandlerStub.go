package testscommon

type EnableEpochsHandlerStub struct {
	BlockGasAndFeesReCheckEnableEpochCalled                   func() uint32
	IsSCDeployFlagEnabledCalled                               func() bool
	IsBuiltInFunctionsFlagEnabledCalled                       func() bool
	IsRelayedTransactionsFlagEnabledCalled                    func() bool
	IsPenalizedTooMuchGasFlagEnabledCalled                    func() bool
	IsSwitchJailWaitingFlagEnabledCalled                      func() bool
	IsBelowSignedThresholdFlagEnabledCalled                   func() bool
	IsSwitchHysteresisForMinNodesFlagEnabledCalled            func() bool
	IsTransactionSignedWithTxHashFlagEnabledCalled            func() bool
	IsMetaProtectionFlagEnabledCalled                         func() bool
	IsAheadOfTimeGasUsageFlagEnabledCalled                    func() bool
	IsGasPriceModifierFlagEnabledCalled                       func() bool
	IsRepairCallbackFlagEnabledCalled                         func() bool
	IsBalanceWaitingListsFlagEnabledCalled                    func() bool
	IsReturnDataToLastTransferFlagEnabledCalled               func() bool
	IsSenderInOutTransferFlagEnabledCalled                    func() bool
	IsStakeFlagEnabledCalled                                  func() bool
	IsStakingV2FlagEnabledCalled                              func() bool
	IsStakingV2OwnerFlagEnabledCalled                         func() bool
	IsStakingV2DelegationFlagEnabledCalled                    func() bool
	IsDoubleKeyProtectionFlagEnabledCalled                    func() bool
	IsESDTFlagEnabledCalled                                   func() bool
	IsESDTFlagEnabledForCurrentEpochCalled                    func() bool
	IsGovernanceFlagEnabledCalled                             func() bool
	IsGovernanceFlagEnabledForCurrentEpochCalled              func() bool
	IsDelegationManagerFlagEnabledCalled                      func() bool
	IsDelegationSmartContractFlagEnabledCalled                func() bool
	IsCorrectLastUnjailedFlagEnabledCalled                    func() bool
	IsCorrectLastUnjailedFlagEnabledForCurrentEpochCalled     func() bool
	IsRelayedTransactionsV2FlagEnabledCalled                  func() bool
	IsUnbondTokensV2FlagEnabledCalled                         func() bool
	IsSaveJailedAlwaysFlagEnabledCalled                       func() bool
	IsReDelegateBelowMinCheckFlagEnabledCalled                func() bool
	IsValidatorToDelegationFlagEnabledCalled                  func() bool
	IsWaitingListFixFlagEnabledCalled                         func() bool
	IsIncrementSCRNonceInMultiTransferFlagEnabledCalled       func() bool
	IsESDTMultiTransferFlagEnabledCalled                      func() bool
	IsGlobalMintBurnFlagEnabledCalled                         func() bool
	IsESDTTransferRoleFlagEnabledCalled                       func() bool
	IsBuiltInFunctionOnMetaFlagEnabledCalled                  func() bool
	IsComputeRewardCheckpointFlagEnabledCalled                func() bool
	IsSCRSizeInvariantCheckFlagEnabledCalled                  func() bool
	IsBackwardCompSaveKeyValueFlagEnabledCalled               func() bool
	IsESDTNFTCreateOnMultiShardFlagEnabledCalled              func() bool
	IsMetaESDTSetFlagEnabledCalled                            func() bool
	IsAddTokensToDelegationFlagEnabledCalled                  func() bool
	IsMultiESDTTransferFixOnCallBackFlagEnabledCalled         func() bool
	IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledCalled       func() bool
	IsCorrectFirstQueuedFlagEnabledCalled                     func() bool
	IsDeleteDelegatorAfterClaimRewardsFlagEnabledCalled       func() bool
	IsFixOOGReturnCodeFlagEnabledCalled                       func() bool
	IsRemoveNonUpdatedStorageFlagEnabledCalled                func() bool
	IsOptimizeNFTStoreFlagEnabledCalled                       func() bool
	IsCreateNFTThroughExecByCallerFlagEnabledCalled           func() bool
	IsStopDecreasingValidatorRatingWhenStuckFlagEnabledCalled func() bool
	IsFrontRunningProtectionFlagEnabledCalled                 func() bool
	IsPayableBySCFlagEnabledCalled                            func() bool
	IsCleanUpInformativeSCRsFlagEnabledCalled                 func() bool
	IsStorageAPICostOptimizationFlagEnabledCalled             func() bool
	IsESDTRegisterAndSetAllRolesFlagEnabledCalled             func() bool
	IsScheduledMiniBlocksFlagEnabledCalled                    func() bool
	IsCorrectJailedNotUnstakedEmptyQueueFlagEnabledCalled     func() bool
	IsDoNotReturnOldBlockInBlockchainHookFlagEnabledCalled    func() bool
	IsSCRSizeInvariantOnBuiltInResultFlagEnabledCalled        func() bool
	IsCheckCorrectTokenIDForTransferRoleFlagEnabledCalled     func() bool
	IsFailExecutionOnEveryAPIErrorFlagEnabledCalled           func() bool
	IsHeartbeatDisableFlagEnabledCalled                       func() bool
	IsMiniBlockPartialExecutionFlagEnabledCalled              func() bool
}

// BlockGasAndFeesReCheckEnableEpoch -
func (stub *EnableEpochsHandlerStub) BlockGasAndFeesReCheckEnableEpoch() uint32 {
	if stub.BlockGasAndFeesReCheckEnableEpochCalled != nil {
		return stub.BlockGasAndFeesReCheckEnableEpochCalled()
	}

	return 0
}

// IsSCDeployFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSCDeployFlagEnabled() bool {
	if stub.IsSCDeployFlagEnabledCalled != nil {
		return stub.IsSCDeployFlagEnabledCalled()
	}

	return false
}

// IsBuiltInFunctionsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBuiltInFunctionsFlagEnabled() bool {
	if stub.IsBuiltInFunctionsFlagEnabledCalled != nil {
		return stub.IsBuiltInFunctionsFlagEnabledCalled()
	}

	return false
}

// IsRelayedTransactionsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRelayedTransactionsFlagEnabled() bool {
	if stub.IsRelayedTransactionsFlagEnabledCalled != nil {
		return stub.IsRelayedTransactionsFlagEnabledCalled()
	}

	return false
}

// IsPenalizedTooMuchGasFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsPenalizedTooMuchGasFlagEnabled() bool {
	if stub.IsPenalizedTooMuchGasFlagEnabledCalled != nil {
		return stub.IsPenalizedTooMuchGasFlagEnabledCalled()
	}

	return false
}

// IsSwitchJailWaitingFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSwitchJailWaitingFlagEnabled() bool {
	if stub.IsSwitchJailWaitingFlagEnabledCalled != nil {
		return stub.IsSwitchJailWaitingFlagEnabledCalled()
	}

	return false
}

// IsBelowSignedThresholdFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBelowSignedThresholdFlagEnabled() bool {
	if stub.IsBelowSignedThresholdFlagEnabledCalled != nil {
		return stub.IsBelowSignedThresholdFlagEnabledCalled()
	}

	return false
}

// IsSwitchHysteresisForMinNodesFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSwitchHysteresisForMinNodesFlagEnabled() bool {
	if stub.IsSwitchHysteresisForMinNodesFlagEnabledCalled != nil {
		return stub.IsSwitchHysteresisForMinNodesFlagEnabledCalled()
	}

	return false
}

// IsTransactionSignedWithTxHashFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsTransactionSignedWithTxHashFlagEnabled() bool {
	if stub.IsTransactionSignedWithTxHashFlagEnabledCalled != nil {
		return stub.IsTransactionSignedWithTxHashFlagEnabledCalled()
	}

	return false
}

// IsMetaProtectionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsMetaProtectionFlagEnabled() bool {
	if stub.IsMetaProtectionFlagEnabledCalled != nil {
		return stub.IsMetaProtectionFlagEnabledCalled()
	}

	return false
}

// IsAheadOfTimeGasUsageFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsAheadOfTimeGasUsageFlagEnabled() bool {
	if stub.IsAheadOfTimeGasUsageFlagEnabledCalled != nil {
		return stub.IsAheadOfTimeGasUsageFlagEnabledCalled()
	}

	return false
}

// IsGasPriceModifierFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsGasPriceModifierFlagEnabled() bool {
	if stub.IsGasPriceModifierFlagEnabledCalled != nil {
		return stub.IsGasPriceModifierFlagEnabledCalled()
	}

	return false
}

// IsRepairCallbackFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRepairCallbackFlagEnabled() bool {
	if stub.IsRepairCallbackFlagEnabledCalled != nil {
		return stub.IsRepairCallbackFlagEnabledCalled()
	}

	return false
}

// IsBalanceWaitingListsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBalanceWaitingListsFlagEnabled() bool {
	if stub.IsBalanceWaitingListsFlagEnabledCalled != nil {
		return stub.IsBalanceWaitingListsFlagEnabledCalled()
	}

	return false
}

// IsReturnDataToLastTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsReturnDataToLastTransferFlagEnabled() bool {
	if stub.IsReturnDataToLastTransferFlagEnabledCalled != nil {
		return stub.IsReturnDataToLastTransferFlagEnabledCalled()
	}

	return false
}

// IsSenderInOutTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSenderInOutTransferFlagEnabled() bool {
	if stub.IsSenderInOutTransferFlagEnabledCalled != nil {
		return stub.IsSenderInOutTransferFlagEnabledCalled()
	}

	return false
}

// IsStakeFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStakeFlagEnabled() bool {
	if stub.IsStakeFlagEnabledCalled != nil {
		return stub.IsStakeFlagEnabledCalled()
	}

	return false
}

// IsStakingV2FlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStakingV2FlagEnabled() bool {
	if stub.IsStakingV2FlagEnabledCalled != nil {
		return stub.IsStakingV2FlagEnabledCalled()
	}

	return false
}

// IsStakingV2OwnerFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStakingV2OwnerFlagEnabled() bool {
	if stub.IsStakingV2OwnerFlagEnabledCalled != nil {
		return stub.IsStakingV2OwnerFlagEnabledCalled()
	}

	return false
}

// IsStakingV2DelegationFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStakingV2DelegationFlagEnabled() bool {
	if stub.IsStakingV2DelegationFlagEnabledCalled != nil {
		return stub.IsStakingV2DelegationFlagEnabledCalled()
	}

	return false
}

// IsDoubleKeyProtectionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDoubleKeyProtectionFlagEnabled() bool {
	if stub.IsDoubleKeyProtectionFlagEnabledCalled != nil {
		return stub.IsDoubleKeyProtectionFlagEnabledCalled()
	}

	return false
}

// IsESDTFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTFlagEnabled() bool {
	if stub.IsESDTFlagEnabledCalled != nil {
		return stub.IsESDTFlagEnabledCalled()
	}

	return false
}

// IsESDTFlagEnabledForCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsESDTFlagEnabledForCurrentEpoch() bool {
	if stub.IsESDTFlagEnabledForCurrentEpochCalled != nil {
		return stub.IsESDTFlagEnabledForCurrentEpochCalled()
	}

	return false
}

// IsGovernanceFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsGovernanceFlagEnabled() bool {
	if stub.IsGovernanceFlagEnabledCalled != nil {
		return stub.IsGovernanceFlagEnabledCalled()
	}

	return false
}

// IsGovernanceFlagEnabledForCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsGovernanceFlagEnabledForCurrentEpoch() bool {
	if stub.IsGovernanceFlagEnabledForCurrentEpochCalled != nil {
		return stub.IsGovernanceFlagEnabledForCurrentEpochCalled()
	}

	return false
}

// IsDelegationManagerFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDelegationManagerFlagEnabled() bool {
	if stub.IsDelegationManagerFlagEnabledCalled != nil {
		return stub.IsDelegationManagerFlagEnabledCalled()
	}

	return false
}

// IsDelegationSmartContractFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDelegationSmartContractFlagEnabled() bool {
	if stub.IsDelegationSmartContractFlagEnabledCalled != nil {
		return stub.IsDelegationSmartContractFlagEnabledCalled()
	}

	return false
}

// IsCorrectLastUnjailedFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCorrectLastUnjailedFlagEnabled() bool {
	if stub.IsCorrectLastUnjailedFlagEnabledCalled != nil {
		return stub.IsCorrectLastUnjailedFlagEnabledCalled()
	}

	return false
}

// IsCorrectLastUnjailedFlagEnabledForCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsCorrectLastUnjailedFlagEnabledForCurrentEpoch() bool {
	if stub.IsCorrectLastUnjailedFlagEnabledForCurrentEpochCalled != nil {
		return stub.IsCorrectLastUnjailedFlagEnabledForCurrentEpochCalled()
	}

	return false
}

// IsRelayedTransactionsV2FlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRelayedTransactionsV2FlagEnabled() bool {
	if stub.IsRelayedTransactionsV2FlagEnabledCalled != nil {
		return stub.IsRelayedTransactionsV2FlagEnabledCalled()
	}

	return false
}

// IsUnbondTokensV2FlagEnabled -
func (stub *EnableEpochsHandlerStub) IsUnbondTokensV2FlagEnabled() bool {
	if stub.IsUnbondTokensV2FlagEnabledCalled != nil {
		return stub.IsUnbondTokensV2FlagEnabledCalled()
	}

	return false
}

// IsSaveJailedAlwaysFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSaveJailedAlwaysFlagEnabled() bool {
	if stub.IsSaveJailedAlwaysFlagEnabledCalled != nil {
		return stub.IsSaveJailedAlwaysFlagEnabledCalled()
	}

	return false
}

// IsReDelegateBelowMinCheckFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsReDelegateBelowMinCheckFlagEnabled() bool {
	if stub.IsReDelegateBelowMinCheckFlagEnabledCalled != nil {
		return stub.IsReDelegateBelowMinCheckFlagEnabledCalled()
	}

	return false
}

// IsValidatorToDelegationFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsValidatorToDelegationFlagEnabled() bool {
	if stub.IsValidatorToDelegationFlagEnabledCalled != nil {
		return stub.IsValidatorToDelegationFlagEnabledCalled()
	}

	return false
}

// IsWaitingListFixFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsWaitingListFixFlagEnabled() bool {
	if stub.IsWaitingListFixFlagEnabledCalled != nil {
		return stub.IsWaitingListFixFlagEnabledCalled()
	}

	return false
}

// IsIncrementSCRNonceInMultiTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsIncrementSCRNonceInMultiTransferFlagEnabled() bool {
	if stub.IsIncrementSCRNonceInMultiTransferFlagEnabledCalled != nil {
		return stub.IsIncrementSCRNonceInMultiTransferFlagEnabledCalled()
	}

	return false
}

// IsESDTMultiTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTMultiTransferFlagEnabled() bool {
	if stub.IsESDTMultiTransferFlagEnabledCalled != nil {
		return stub.IsESDTMultiTransferFlagEnabledCalled()
	}

	return false
}

// IsGlobalMintBurnFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsGlobalMintBurnFlagEnabled() bool {
	if stub.IsGlobalMintBurnFlagEnabledCalled != nil {
		return stub.IsGlobalMintBurnFlagEnabledCalled()
	}

	return false
}

// IsESDTTransferRoleFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTTransferRoleFlagEnabled() bool {
	if stub.IsESDTTransferRoleFlagEnabledCalled != nil {
		return stub.IsESDTTransferRoleFlagEnabledCalled()
	}

	return false
}

// IsBuiltInFunctionOnMetaFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBuiltInFunctionOnMetaFlagEnabled() bool {
	if stub.IsBuiltInFunctionOnMetaFlagEnabledCalled != nil {
		return stub.IsBuiltInFunctionOnMetaFlagEnabledCalled()
	}

	return false
}

// IsComputeRewardCheckpointFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsComputeRewardCheckpointFlagEnabled() bool {
	if stub.IsComputeRewardCheckpointFlagEnabledCalled != nil {
		return stub.IsComputeRewardCheckpointFlagEnabledCalled()
	}

	return false
}

// IsSCRSizeInvariantCheckFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSCRSizeInvariantCheckFlagEnabled() bool {
	if stub.IsSCRSizeInvariantCheckFlagEnabledCalled != nil {
		return stub.IsSCRSizeInvariantCheckFlagEnabledCalled()
	}

	return false
}

// IsBackwardCompSaveKeyValueFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBackwardCompSaveKeyValueFlagEnabled() bool {
	if stub.IsBackwardCompSaveKeyValueFlagEnabledCalled != nil {
		return stub.IsBackwardCompSaveKeyValueFlagEnabledCalled()
	}

	return false
}

// IsESDTNFTCreateOnMultiShardFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTNFTCreateOnMultiShardFlagEnabled() bool {
	if stub.IsESDTNFTCreateOnMultiShardFlagEnabledCalled != nil {
		return stub.IsESDTNFTCreateOnMultiShardFlagEnabledCalled()
	}

	return false
}

// IsMetaESDTSetFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsMetaESDTSetFlagEnabled() bool {
	if stub.IsMetaESDTSetFlagEnabledCalled != nil {
		return stub.IsMetaESDTSetFlagEnabledCalled()
	}

	return false
}

// IsAddTokensToDelegationFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsAddTokensToDelegationFlagEnabled() bool {
	if stub.IsAddTokensToDelegationFlagEnabledCalled != nil {
		return stub.IsAddTokensToDelegationFlagEnabledCalled()
	}

	return false
}

// IsMultiESDTTransferFixOnCallBackFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsMultiESDTTransferFixOnCallBackFlagEnabled() bool {
	if stub.IsMultiESDTTransferFixOnCallBackFlagEnabledCalled != nil {
		return stub.IsMultiESDTTransferFixOnCallBackFlagEnabledCalled()
	}

	return false
}

// IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled() bool {
	if stub.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledCalled != nil {
		return stub.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledCalled()
	}

	return false
}

// IsCorrectFirstQueuedFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCorrectFirstQueuedFlagEnabled() bool {
	if stub.IsCorrectFirstQueuedFlagEnabledCalled != nil {
		return stub.IsCorrectFirstQueuedFlagEnabledCalled()
	}

	return false
}

// IsDeleteDelegatorAfterClaimRewardsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDeleteDelegatorAfterClaimRewardsFlagEnabled() bool {
	if stub.IsDeleteDelegatorAfterClaimRewardsFlagEnabledCalled != nil {
		return stub.IsDeleteDelegatorAfterClaimRewardsFlagEnabledCalled()
	}

	return false
}

// IsFixOOGReturnCodeFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsFixOOGReturnCodeFlagEnabled() bool {
	if stub.IsFixOOGReturnCodeFlagEnabledCalled != nil {
		return stub.IsFixOOGReturnCodeFlagEnabledCalled()
	}

	return false
}

// IsRemoveNonUpdatedStorageFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRemoveNonUpdatedStorageFlagEnabled() bool {
	if stub.IsRemoveNonUpdatedStorageFlagEnabledCalled != nil {
		return stub.IsRemoveNonUpdatedStorageFlagEnabledCalled()
	}

	return false
}

// IsOptimizeNFTStoreFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsOptimizeNFTStoreFlagEnabled() bool {
	if stub.IsOptimizeNFTStoreFlagEnabledCalled != nil {
		return stub.IsOptimizeNFTStoreFlagEnabledCalled()
	}

	return false
}

// IsCreateNFTThroughExecByCallerFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCreateNFTThroughExecByCallerFlagEnabled() bool {
	if stub.IsCreateNFTThroughExecByCallerFlagEnabledCalled != nil {
		return stub.IsCreateNFTThroughExecByCallerFlagEnabledCalled()
	}

	return false
}

// IsStopDecreasingValidatorRatingWhenStuckFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStopDecreasingValidatorRatingWhenStuckFlagEnabled() bool {
	if stub.IsStopDecreasingValidatorRatingWhenStuckFlagEnabledCalled != nil {
		return stub.IsStopDecreasingValidatorRatingWhenStuckFlagEnabledCalled()
	}

	return false
}

// IsFrontRunningProtectionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsFrontRunningProtectionFlagEnabled() bool {
	if stub.IsFrontRunningProtectionFlagEnabledCalled != nil {
		return stub.IsFrontRunningProtectionFlagEnabledCalled()
	}

	return false
}

// IsPayableBySCFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsPayableBySCFlagEnabled() bool {
	if stub.IsPayableBySCFlagEnabledCalled != nil {
		return stub.IsPayableBySCFlagEnabledCalled()
	}

	return false
}

// IsCleanUpInformativeSCRsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCleanUpInformativeSCRsFlagEnabled() bool {
	if stub.IsCleanUpInformativeSCRsFlagEnabledCalled != nil {
		return stub.IsCleanUpInformativeSCRsFlagEnabledCalled()
	}

	return false
}

// IsStorageAPICostOptimizationFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStorageAPICostOptimizationFlagEnabled() bool {
	if stub.IsStorageAPICostOptimizationFlagEnabledCalled != nil {
		return stub.IsStorageAPICostOptimizationFlagEnabledCalled()
	}

	return false
}

// IsESDTRegisterAndSetAllRolesFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTRegisterAndSetAllRolesFlagEnabled() bool {
	if stub.IsESDTRegisterAndSetAllRolesFlagEnabledCalled != nil {
		return stub.IsESDTRegisterAndSetAllRolesFlagEnabledCalled()
	}

	return false
}

// IsScheduledMiniBlocksFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsScheduledMiniBlocksFlagEnabled() bool {
	if stub.IsScheduledMiniBlocksFlagEnabledCalled != nil {
		return stub.IsScheduledMiniBlocksFlagEnabledCalled()
	}

	return false
}

// IsCorrectJailedNotUnstakedEmptyQueueFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCorrectJailedNotUnstakedEmptyQueueFlagEnabled() bool {
	if stub.IsCorrectJailedNotUnstakedEmptyQueueFlagEnabledCalled != nil {
		return stub.IsCorrectJailedNotUnstakedEmptyQueueFlagEnabledCalled()
	}

	return false
}

// IsDoNotReturnOldBlockInBlockchainHookFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDoNotReturnOldBlockInBlockchainHookFlagEnabled() bool {
	if stub.IsDoNotReturnOldBlockInBlockchainHookFlagEnabledCalled != nil {
		return stub.IsDoNotReturnOldBlockInBlockchainHookFlagEnabledCalled()
	}

	return false
}

// IsSCRSizeInvariantOnBuiltInResultFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSCRSizeInvariantOnBuiltInResultFlagEnabled() bool {
	if stub.IsSCRSizeInvariantOnBuiltInResultFlagEnabledCalled != nil {
		return stub.IsSCRSizeInvariantOnBuiltInResultFlagEnabledCalled()
	}

	return false
}

// IsCheckCorrectTokenIDForTransferRoleFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCheckCorrectTokenIDForTransferRoleFlagEnabled() bool {
	if stub.IsCheckCorrectTokenIDForTransferRoleFlagEnabledCalled != nil {
		return stub.IsCheckCorrectTokenIDForTransferRoleFlagEnabledCalled()
	}

	return false
}

// IsFailExecutionOnEveryAPIErrorFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsFailExecutionOnEveryAPIErrorFlagEnabled() bool {
	if stub.IsFailExecutionOnEveryAPIErrorFlagEnabledCalled != nil {
		return stub.IsFailExecutionOnEveryAPIErrorFlagEnabledCalled()
	}

	return false
}

// IsHeartbeatDisableFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsHeartbeatDisableFlagEnabled() bool {
	if stub.IsHeartbeatDisableFlagEnabledCalled != nil {
		return stub.IsHeartbeatDisableFlagEnabledCalled()
	}

	return false
}

// IsMiniBlockPartialExecutionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsMiniBlockPartialExecutionFlagEnabled() bool {
	if stub.IsMiniBlockPartialExecutionFlagEnabledCalled != nil {
		return stub.IsMiniBlockPartialExecutionFlagEnabledCalled()
	}

	return false
}

// IsInterfaceNil -
func (stub *EnableEpochsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
