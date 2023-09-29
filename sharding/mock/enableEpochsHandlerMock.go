package mock

// EnableEpochsHandlerMock -
type EnableEpochsHandlerMock struct {
	WaitingListFixEnableEpochField            uint32
	RefactorPeersMiniBlocksEnableEpochField   uint32
	IsRefactorPeersMiniBlocksFlagEnabledField bool
	IsSCProcessorV2FlagEnabledField           bool
	IsFixOldTokenLiquidityFlagEnabledField    bool
}

// BlockGasAndFeesReCheckEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) BlockGasAndFeesReCheckEnableEpoch() uint32 {
	return 0
}

// StakingV2EnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) StakingV2EnableEpoch() uint32 {
	return 0
}

// ScheduledMiniBlocksEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) ScheduledMiniBlocksEnableEpoch() uint32 {
	return 0
}

// SwitchJailWaitingEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) SwitchJailWaitingEnableEpoch() uint32 {
	return 0
}

// BalanceWaitingListsEnableEpoch returns WaitingListFixEnableEpochField
func (mock *EnableEpochsHandlerMock) BalanceWaitingListsEnableEpoch() uint32 {
	return 0
}

// WaitingListFixEnableEpoch returns WaitingListFixEnableEpochField
func (mock *EnableEpochsHandlerMock) WaitingListFixEnableEpoch() uint32 {
	return mock.WaitingListFixEnableEpochField
}

// MultiESDTTransferAsyncCallBackEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) MultiESDTTransferAsyncCallBackEnableEpoch() uint32 {
	return 0
}

// FixOOGReturnCodeEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) FixOOGReturnCodeEnableEpoch() uint32 {
	return 0
}

// RemoveNonUpdatedStorageEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) RemoveNonUpdatedStorageEnableEpoch() uint32 {
	return 0
}

// CreateNFTThroughExecByCallerEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) CreateNFTThroughExecByCallerEnableEpoch() uint32 {
	return 0
}

// FixFailExecutionOnErrorEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) FixFailExecutionOnErrorEnableEpoch() uint32 {
	return 0
}

// ManagedCryptoAPIEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) ManagedCryptoAPIEnableEpoch() uint32 {
	return 0
}

// DisableExecByCallerEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) DisableExecByCallerEnableEpoch() uint32 {
	return 0
}

// RefactorContextEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) RefactorContextEnableEpoch() uint32 {
	return 0
}

// CheckExecuteReadOnlyEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) CheckExecuteReadOnlyEnableEpoch() uint32 {
	return 0
}

// StorageAPICostOptimizationEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) StorageAPICostOptimizationEnableEpoch() uint32 {
	return 0
}

// MiniBlockPartialExecutionEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) MiniBlockPartialExecutionEnableEpoch() uint32 {
	return 0
}

// RefactorPeersMiniBlocksEnableEpoch returns 0
func (mock *EnableEpochsHandlerMock) RefactorPeersMiniBlocksEnableEpoch() uint32 {
	return mock.RefactorPeersMiniBlocksEnableEpochField
}

// IsSCDeployFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsSCDeployFlagEnabled() bool {
	return false
}

// IsBuiltInFunctionsFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsBuiltInFunctionsFlagEnabled() bool {
	return false
}

// IsRelayedTransactionsFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsRelayedTransactionsFlagEnabled() bool {
	return false
}

// IsPenalizedTooMuchGasFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsPenalizedTooMuchGasFlagEnabled() bool {
	return false
}

// ResetPenalizedTooMuchGasFlag does nothing
func (mock *EnableEpochsHandlerMock) ResetPenalizedTooMuchGasFlag() {
}

// IsSwitchJailWaitingFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsSwitchJailWaitingFlagEnabled() bool {
	return false
}

// IsBelowSignedThresholdFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsBelowSignedThresholdFlagEnabled() bool {
	return false
}

// IsSwitchHysteresisForMinNodesFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsSwitchHysteresisForMinNodesFlagEnabled() bool {
	return false
}

// IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch returns false
func (mock *EnableEpochsHandlerMock) IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch() bool {
	return false
}

// IsTransactionSignedWithTxHashFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsTransactionSignedWithTxHashFlagEnabled() bool {
	return false
}

// IsMetaProtectionFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsMetaProtectionFlagEnabled() bool {
	return false
}

// IsAheadOfTimeGasUsageFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsAheadOfTimeGasUsageFlagEnabled() bool {
	return false
}

// IsGasPriceModifierFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsGasPriceModifierFlagEnabled() bool {
	return false
}

// IsRepairCallbackFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsRepairCallbackFlagEnabled() bool {
	return false
}

// IsBalanceWaitingListsFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsBalanceWaitingListsFlagEnabled() bool {
	return false
}

// IsReturnDataToLastTransferFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsReturnDataToLastTransferFlagEnabled() bool {
	return false
}

// IsSenderInOutTransferFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsSenderInOutTransferFlagEnabled() bool {
	return false
}

// IsStakeFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsStakeFlagEnabled() bool {
	return false
}

// IsStakingV2FlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsStakingV2FlagEnabled() bool {
	return false
}

// IsStakingV2OwnerFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsStakingV2OwnerFlagEnabled() bool {
	return false
}

// IsStakingV2FlagEnabledForActivationEpochCompleted returns false
func (mock *EnableEpochsHandlerMock) IsStakingV2FlagEnabledForActivationEpochCompleted() bool {
	return false
}

// IsDoubleKeyProtectionFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsDoubleKeyProtectionFlagEnabled() bool {
	return false
}

// IsESDTFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsESDTFlagEnabled() bool {
	return false
}

// IsESDTFlagEnabledForCurrentEpoch returns false
func (mock *EnableEpochsHandlerMock) IsESDTFlagEnabledForCurrentEpoch() bool {
	return false
}

// IsGovernanceFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsGovernanceFlagEnabled() bool {
	return false
}

// IsGovernanceFlagEnabledForCurrentEpoch returns false
func (mock *EnableEpochsHandlerMock) IsGovernanceFlagEnabledForCurrentEpoch() bool {
	return false
}

// IsDelegationManagerFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsDelegationManagerFlagEnabled() bool {
	return false
}

// IsDelegationSmartContractFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsDelegationSmartContractFlagEnabled() bool {
	return false
}

// IsDelegationSmartContractFlagEnabledForCurrentEpoch returns false
func (mock *EnableEpochsHandlerMock) IsDelegationSmartContractFlagEnabledForCurrentEpoch() bool {
	return false
}

// IsCorrectLastUnJailedFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsCorrectLastUnJailedFlagEnabled() bool {
	return false
}

// IsCorrectLastUnJailedFlagEnabledForCurrentEpoch returns false
func (mock *EnableEpochsHandlerMock) IsCorrectLastUnJailedFlagEnabledForCurrentEpoch() bool {
	return false
}

// IsRelayedTransactionsV2FlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsRelayedTransactionsV2FlagEnabled() bool {
	return false
}

// IsUnBondTokensV2FlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsUnBondTokensV2FlagEnabled() bool {
	return false
}

// IsSaveJailedAlwaysFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsSaveJailedAlwaysFlagEnabled() bool {
	return false
}

// IsReDelegateBelowMinCheckFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsReDelegateBelowMinCheckFlagEnabled() bool {
	return false
}

// IsValidatorToDelegationFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsValidatorToDelegationFlagEnabled() bool {
	return false
}

// IsWaitingListFixFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsWaitingListFixFlagEnabled() bool {
	return false
}

// IsIncrementSCRNonceInMultiTransferFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsIncrementSCRNonceInMultiTransferFlagEnabled() bool {
	return false
}

// IsESDTMultiTransferFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsESDTMultiTransferFlagEnabled() bool {
	return false
}

// IsGlobalMintBurnFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsGlobalMintBurnFlagEnabled() bool {
	return false
}

// IsESDTTransferRoleFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsESDTTransferRoleFlagEnabled() bool {
	return false
}

// IsBuiltInFunctionOnMetaFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsBuiltInFunctionOnMetaFlagEnabled() bool {
	return false
}

// IsComputeRewardCheckpointFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsComputeRewardCheckpointFlagEnabled() bool {
	return false
}

// IsSCRSizeInvariantCheckFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsSCRSizeInvariantCheckFlagEnabled() bool {
	return false
}

// IsBackwardCompSaveKeyValueFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsBackwardCompSaveKeyValueFlagEnabled() bool {
	return false
}

// IsESDTNFTCreateOnMultiShardFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsESDTNFTCreateOnMultiShardFlagEnabled() bool {
	return false
}

// IsMetaESDTSetFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsMetaESDTSetFlagEnabled() bool {
	return false
}

// IsAddTokensToDelegationFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsAddTokensToDelegationFlagEnabled() bool {
	return false
}

// IsMultiESDTTransferFixOnCallBackFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsMultiESDTTransferFixOnCallBackFlagEnabled() bool {
	return false
}

// IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled() bool {
	return false
}

// IsCorrectFirstQueuedFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsCorrectFirstQueuedFlagEnabled() bool {
	return false
}

// IsDeleteDelegatorAfterClaimRewardsFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsDeleteDelegatorAfterClaimRewardsFlagEnabled() bool {
	return false
}

// IsFixOOGReturnCodeFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsFixOOGReturnCodeFlagEnabled() bool {
	return false
}

// IsRemoveNonUpdatedStorageFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsRemoveNonUpdatedStorageFlagEnabled() bool {
	return false
}

// IsOptimizeNFTStoreFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsOptimizeNFTStoreFlagEnabled() bool {
	return false
}

// IsCreateNFTThroughExecByCallerFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsCreateNFTThroughExecByCallerFlagEnabled() bool {
	return false
}

// IsStopDecreasingValidatorRatingWhenStuckFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsStopDecreasingValidatorRatingWhenStuckFlagEnabled() bool {
	return false
}

// IsFrontRunningProtectionFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsFrontRunningProtectionFlagEnabled() bool {
	return false
}

// IsPayableBySCFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsPayableBySCFlagEnabled() bool {
	return false
}

// IsCleanUpInformativeSCRsFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsCleanUpInformativeSCRsFlagEnabled() bool {
	return false
}

// IsStorageAPICostOptimizationFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsStorageAPICostOptimizationFlagEnabled() bool {
	return false
}

// IsESDTRegisterAndSetAllRolesFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsESDTRegisterAndSetAllRolesFlagEnabled() bool {
	return false
}

// IsScheduledMiniBlocksFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsScheduledMiniBlocksFlagEnabled() bool {
	return false
}

// IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled() bool {
	return false
}

// IsDoNotReturnOldBlockInBlockchainHookFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsDoNotReturnOldBlockInBlockchainHookFlagEnabled() bool {
	return false
}

// IsAddFailedRelayedTxToInvalidMBsFlag returns false
func (mock *EnableEpochsHandlerMock) IsAddFailedRelayedTxToInvalidMBsFlag() bool {
	return false
}

// IsSCRSizeInvariantOnBuiltInResultFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsSCRSizeInvariantOnBuiltInResultFlagEnabled() bool {
	return false
}

// IsCheckCorrectTokenIDForTransferRoleFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsCheckCorrectTokenIDForTransferRoleFlagEnabled() bool {
	return false
}

// IsFailExecutionOnEveryAPIErrorFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsFailExecutionOnEveryAPIErrorFlagEnabled() bool {
	return false
}

// IsMiniBlockPartialExecutionFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsMiniBlockPartialExecutionFlagEnabled() bool {
	return false
}

// IsManagedCryptoAPIsFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsManagedCryptoAPIsFlagEnabled() bool {
	return false
}

// IsESDTMetadataContinuousCleanupFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsESDTMetadataContinuousCleanupFlagEnabled() bool {
	return false
}

// IsDisableExecByCallerFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsDisableExecByCallerFlagEnabled() bool {
	return false
}

// IsRefactorContextFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsRefactorContextFlagEnabled() bool {
	return false
}

// IsCheckFunctionArgumentFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsCheckFunctionArgumentFlagEnabled() bool {
	return false
}

// IsCheckExecuteOnReadOnlyFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsCheckExecuteOnReadOnlyFlagEnabled() bool {
	return false
}

// IsFixAsyncCallbackCheckFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsFixAsyncCallbackCheckFlagEnabled() bool {
	return false
}

// IsSaveToSystemAccountFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsSaveToSystemAccountFlagEnabled() bool {
	return false
}

// IsCheckFrozenCollectionFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsCheckFrozenCollectionFlagEnabled() bool {
	return false
}

// IsSendAlwaysFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsSendAlwaysFlagEnabled() bool {
	return false
}

// IsValueLengthCheckFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsValueLengthCheckFlagEnabled() bool {
	return false
}

// IsCheckTransferFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsCheckTransferFlagEnabled() bool {
	return false
}

// IsTransferToMetaFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsTransferToMetaFlagEnabled() bool {
	return false
}

// IsESDTNFTImprovementV1FlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsESDTNFTImprovementV1FlagEnabled() bool {
	return false
}

// IsSetSenderInEeiOutputTransferFlagEnabled -
func (mock *EnableEpochsHandlerMock) IsSetSenderInEeiOutputTransferFlagEnabled() bool {
	return false
}

// IsChangeDelegationOwnerFlagEnabled -
func (mock *EnableEpochsHandlerMock) IsChangeDelegationOwnerFlagEnabled() bool {
	return false
}

// IsRefactorPeersMiniBlocksFlagEnabled returns false
func (mock *EnableEpochsHandlerMock) IsRefactorPeersMiniBlocksFlagEnabled() bool {
	return mock.IsRefactorPeersMiniBlocksFlagEnabledField
}

// IsSCProcessorV2FlagEnabled -
func (mock *EnableEpochsHandlerMock) IsSCProcessorV2FlagEnabled() bool {
	return mock.IsSCProcessorV2FlagEnabledField
}

// IsFixAsyncCallBackArgsListFlagEnabled -
func (mock *EnableEpochsHandlerMock) IsFixAsyncCallBackArgsListFlagEnabled() bool {
	return false
}

// IsFixOldTokenLiquidityEnabled -
func (mock *EnableEpochsHandlerMock) IsFixOldTokenLiquidityEnabled() bool {
	return false
}

// IsRuntimeMemStoreLimitEnabled -
func (mock *EnableEpochsHandlerMock) IsRuntimeMemStoreLimitEnabled() bool {
	return false
}

// IsRuntimeCodeSizeFixEnabled -
func (mock *EnableEpochsHandlerMock) IsRuntimeCodeSizeFixEnabled() bool {
	return false
}

// IsMaxBlockchainHookCountersFlagEnabled -
func (mock *EnableEpochsHandlerMock) IsMaxBlockchainHookCountersFlagEnabled() bool {
	return false
}

// IsWipeSingleNFTLiquidityDecreaseEnabled -
func (mock *EnableEpochsHandlerMock) IsWipeSingleNFTLiquidityDecreaseEnabled() bool {
	return false
}

// IsAlwaysSaveTokenMetaDataEnabled -
func (mock *EnableEpochsHandlerMock) IsAlwaysSaveTokenMetaDataEnabled() bool {
	return false
}

// IsConsensusModelV2Enabled -
func (mock *EnableEpochsHandlerMock) IsConsensusModelV2Enabled() bool {
	return false
}

// IsSetGuardianEnabled returns false
func (mock *EnableEpochsHandlerMock) IsSetGuardianEnabled() bool {
	return false
}

// IsScToScEventLogEnabled returns false
func (mock *EnableEpochsHandlerMock) IsScToScEventLogEnabled() bool {
	return false
}

// IsRelayedNonceFixEnabled -
func (mock *EnableEpochsHandlerMock) IsRelayedNonceFixEnabled() bool {
	return false
}

// IsKeepExecOrderOnCreatedSCRsEnabled -
func (mock *EnableEpochsHandlerMock) IsKeepExecOrderOnCreatedSCRsEnabled() bool {
	return false
}

// IsMultiClaimOnDelegationEnabled -
func (mock *EnableEpochsHandlerMock) IsMultiClaimOnDelegationEnabled() bool {
	return false
}

// IsChangeUsernameEnabled -
func (mock *EnableEpochsHandlerMock) IsChangeUsernameEnabled() bool {
	return false
}

// IsConsistentTokensValuesLengthCheckEnabled -
func (mock *EnableEpochsHandlerMock) IsConsistentTokensValuesLengthCheckEnabled() bool {
	return false
}

// IsAutoBalanceDataTriesEnabled -
func (mock *EnableEpochsHandlerMock) IsAutoBalanceDataTriesEnabled() bool {
	return false
}

// FixDelegationChangeOwnerOnAccountEnabled -
func (mock *EnableEpochsHandlerMock) FixDelegationChangeOwnerOnAccountEnabled() bool {
	return false
}

// IsDeterministicSortOnValidatorsInfoFixEnabled -
func (mock *EnableEpochsHandlerMock) IsDeterministicSortOnValidatorsInfoFixEnabled() bool {
	return false
}

// IsDynamicGasCostForDataTrieStorageLoadEnabled -
func (mock *EnableEpochsHandlerMock) IsDynamicGasCostForDataTrieStorageLoadEnabled() bool {
	return false
}

// NFTStopCreateEnabled -
func (mock *EnableEpochsHandlerMock) NFTStopCreateEnabled() bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (mock *EnableEpochsHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}
