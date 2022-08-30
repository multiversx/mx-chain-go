package testscommon

// EnableEpochsHandlerStub -
type EnableEpochsHandlerStub struct {
	ResetPenalizedTooMuchGasFlagCalled                           func()
	BlockGasAndFeesReCheckEnableEpochField                       uint32
	StakingV2EnableEpochField                                    uint32
	ScheduledMiniBlocksEnableEpochField                          uint32
	SwitchJailWaitingEnableEpochField                            uint32
	BalanceWaitingListsEnableEpochField                          uint32
	WaitingListFixEnableEpochField                               uint32
	MultiESDTTransferAsyncCallBackEnableEpochField               uint32
	FixOOGReturnCodeEnableEpochField                             uint32
	RemoveNonUpdatedStorageEnableEpochField                      uint32
	CreateNFTThroughExecByCallerEnableEpochField                 uint32
	FixFailExecutionOnErrorEnableEpochField                      uint32
	ManagedCryptoAPIEnableEpochField                             uint32
	DisableExecByCallerEnableEpochField                          uint32
	RefactorContextEnableEpochField                              uint32
	CheckExecuteReadOnlyEnableEpochField                         uint32
	StorageAPICostOptimizationEnableEpochField                   uint32
	MiniBlockPartialExecutionEnableEpochField                    uint32
	RefactorPeersMiniBlocksEnableEpochField                      uint32
	IsSCDeployFlagEnabledField                                   bool
	IsBuiltInFunctionsFlagEnabledField                           bool
	IsRelayedTransactionsFlagEnabledField                        bool
	IsPenalizedTooMuchGasFlagEnabledField                        bool
	IsSwitchJailWaitingFlagEnabledField                          bool
	IsBelowSignedThresholdFlagEnabledField                       bool
	IsSwitchHysteresisForMinNodesFlagEnabledField                bool
	IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpochField bool
	IsTransactionSignedWithTxHashFlagEnabledField                bool
	IsMetaProtectionFlagEnabledField                             bool
	IsAheadOfTimeGasUsageFlagEnabledField                        bool
	IsGasPriceModifierFlagEnabledField                           bool
	IsRepairCallbackFlagEnabledField                             bool
	IsBalanceWaitingListsFlagEnabledField                        bool
	IsReturnDataToLastTransferFlagEnabledField                   bool
	IsSenderInOutTransferFlagEnabledField                        bool
	IsStakeFlagEnabledField                                      bool
	IsStakingV2FlagEnabledField                                  bool
	IsStakingV2OwnerFlagEnabledField                             bool
	IsStakingV2FlagEnabledForActivationEpochCompletedField       bool
	IsDoubleKeyProtectionFlagEnabledField                        bool
	IsESDTFlagEnabledField                                       bool
	IsESDTFlagEnabledForCurrentEpochField                        bool
	IsGovernanceFlagEnabledField                                 bool
	IsGovernanceFlagEnabledForCurrentEpochField                  bool
	IsDelegationManagerFlagEnabledField                          bool
	IsDelegationSmartContractFlagEnabledField                    bool
	IsDelegationSmartContractFlagForCurrentEpochEnabledField     bool
	IsCorrectLastUnJailedFlagEnabledField                        bool
	IsCorrectLastUnJailedFlagEnabledForCurrentEpochField         bool
	IsRelayedTransactionsV2FlagEnabledField                      bool
	IsUnBondTokensV2FlagEnabledField                             bool
	IsSaveJailedAlwaysFlagEnabledField                           bool
	IsReDelegateBelowMinCheckFlagEnabledField                    bool
	IsValidatorToDelegationFlagEnabledField                      bool
	IsWaitingListFixFlagEnabledField                             bool
	IsIncrementSCRNonceInMultiTransferFlagEnabledField           bool
	IsESDTMultiTransferFlagEnabledField                          bool
	IsGlobalMintBurnFlagEnabledField                             bool
	IsESDTTransferRoleFlagEnabledField                           bool
	IsBuiltInFunctionOnMetaFlagEnabledField                      bool
	IsComputeRewardCheckpointFlagEnabledField                    bool
	IsSCRSizeInvariantCheckFlagEnabledField                      bool
	IsBackwardCompSaveKeyValueFlagEnabledField                   bool
	IsESDTNFTCreateOnMultiShardFlagEnabledField                  bool
	IsMetaESDTSetFlagEnabledField                                bool
	IsAddTokensToDelegationFlagEnabledField                      bool
	IsMultiESDTTransferFixOnCallBackFlagEnabledField             bool
	IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledField           bool
	IsCorrectFirstQueuedFlagEnabledField                         bool
	IsDeleteDelegatorAfterClaimRewardsFlagEnabledField           bool
	IsFixOOGReturnCodeFlagEnabledField                           bool
	IsRemoveNonUpdatedStorageFlagEnabledField                    bool
	IsOptimizeNFTStoreFlagEnabledField                           bool
	IsCreateNFTThroughExecByCallerFlagEnabledField               bool
	IsStopDecreasingValidatorRatingWhenStuckFlagEnabledField     bool
	IsFrontRunningProtectionFlagEnabledField                     bool
	IsPayableBySCFlagEnabledField                                bool
	IsCleanUpInformativeSCRsFlagEnabledField                     bool
	IsStorageAPICostOptimizationFlagEnabledField                 bool
	IsESDTRegisterAndSetAllRolesFlagEnabledField                 bool
	IsScheduledMiniBlocksFlagEnabledField                        bool
	IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledField         bool
	IsDoNotReturnOldBlockInBlockchainHookFlagEnabledField        bool
	IsAddFailedRelayedTxToInvalidMBsFlagField                    bool
	IsSCRSizeInvariantOnBuiltInResultFlagEnabledField            bool
	IsCheckCorrectTokenIDForTransferRoleFlagEnabledField         bool
	IsFailExecutionOnEveryAPIErrorFlagEnabledField               bool
	IsHeartbeatDisableFlagEnabledField                           bool
	IsMiniBlockPartialExecutionFlagEnabledField                  bool
	IsManagedCryptoAPIsFlagEnabledField                          bool
	IsESDTMetadataContinuousCleanupFlagEnabledField              bool
	IsDisableExecByCallerFlagEnabledField                        bool
	IsRefactorContextFlagEnabledField                            bool
	IsCheckFunctionArgumentFlagEnabledField                      bool
	IsCheckExecuteOnReadOnlyFlagEnabledField                     bool
	IsFixAsyncCallbackCheckFlagEnabledField                      bool
	IsSaveToSystemAccountFlagEnabledField                        bool
	IsCheckFrozenCollectionFlagEnabledField                      bool
	IsSendAlwaysFlagEnabledField                                 bool
	IsValueLengthCheckFlagEnabledField                           bool
	IsCheckTransferFlagEnabledField                              bool
	IsTransferToMetaFlagEnabledField                             bool
	IsESDTNFTImprovementV1FlagEnabledField                       bool
	IsSetSenderInEeiOutputTransferFlagEnabledField               bool
	IsChangeDelegationOwnerFlagEnabledField                      bool
	IsRefactorPeersMiniBlocksFlagEnabledField                    bool
}

// ResetPenalizedTooMuchGasFlag -
func (stub *EnableEpochsHandlerStub) ResetPenalizedTooMuchGasFlag() {
	if stub.ResetPenalizedTooMuchGasFlagCalled != nil {
		stub.ResetPenalizedTooMuchGasFlagCalled()
	}
}

// BlockGasAndFeesReCheckEnableEpoch -
func (stub *EnableEpochsHandlerStub) BlockGasAndFeesReCheckEnableEpoch() uint32 {
	return stub.BlockGasAndFeesReCheckEnableEpochField
}

// StakingV2EnableEpoch -
func (stub *EnableEpochsHandlerStub) StakingV2EnableEpoch() uint32 {
	return stub.StakingV2EnableEpochField
}

// ScheduledMiniBlocksEnableEpoch -
func (stub *EnableEpochsHandlerStub) ScheduledMiniBlocksEnableEpoch() uint32 {
	return stub.ScheduledMiniBlocksEnableEpochField
}

// SwitchJailWaitingEnableEpoch -
func (stub *EnableEpochsHandlerStub) SwitchJailWaitingEnableEpoch() uint32 {
	return stub.SwitchJailWaitingEnableEpochField
}

// BalanceWaitingListsEnableEpoch -
func (stub *EnableEpochsHandlerStub) BalanceWaitingListsEnableEpoch() uint32 {
	return stub.BalanceWaitingListsEnableEpochField
}

// WaitingListFixEnableEpoch -
func (stub *EnableEpochsHandlerStub) WaitingListFixEnableEpoch() uint32 {
	return stub.WaitingListFixEnableEpochField
}

// MultiESDTTransferAsyncCallBackEnableEpoch -
func (stub *EnableEpochsHandlerStub) MultiESDTTransferAsyncCallBackEnableEpoch() uint32 {
	return stub.MultiESDTTransferAsyncCallBackEnableEpochField
}

// FixOOGReturnCodeEnableEpoch -
func (stub *EnableEpochsHandlerStub) FixOOGReturnCodeEnableEpoch() uint32 {
	return stub.FixOOGReturnCodeEnableEpochField
}

// RemoveNonUpdatedStorageEnableEpoch -
func (stub *EnableEpochsHandlerStub) RemoveNonUpdatedStorageEnableEpoch() uint32 {
	return stub.RemoveNonUpdatedStorageEnableEpochField
}

// CreateNFTThroughExecByCallerEnableEpoch -
func (stub *EnableEpochsHandlerStub) CreateNFTThroughExecByCallerEnableEpoch() uint32 {
	return stub.CreateNFTThroughExecByCallerEnableEpochField
}

// FixFailExecutionOnErrorEnableEpoch -
func (stub *EnableEpochsHandlerStub) FixFailExecutionOnErrorEnableEpoch() uint32 {
	return stub.FixFailExecutionOnErrorEnableEpochField
}

// ManagedCryptoAPIEnableEpoch -
func (stub *EnableEpochsHandlerStub) ManagedCryptoAPIEnableEpoch() uint32 {
	return stub.ManagedCryptoAPIEnableEpochField
}

// DisableExecByCallerEnableEpoch -
func (stub *EnableEpochsHandlerStub) DisableExecByCallerEnableEpoch() uint32 {
	return stub.DisableExecByCallerEnableEpochField
}

// RefactorContextEnableEpoch -
func (stub *EnableEpochsHandlerStub) RefactorContextEnableEpoch() uint32 {
	return stub.RefactorContextEnableEpochField
}

// CheckExecuteReadOnlyEnableEpoch -
func (stub *EnableEpochsHandlerStub) CheckExecuteReadOnlyEnableEpoch() uint32 {
	return stub.CheckExecuteReadOnlyEnableEpochField
}

// StorageAPICostOptimizationEnableEpoch -
func (stub *EnableEpochsHandlerStub) StorageAPICostOptimizationEnableEpoch() uint32 {
	return stub.StorageAPICostOptimizationEnableEpochField
}

// MiniBlockPartialExecutionEnableEpoch -
func (stub *EnableEpochsHandlerStub) MiniBlockPartialExecutionEnableEpoch() uint32 {
	return stub.MiniBlockPartialExecutionEnableEpochField
}

// RefactorPeersMiniBlocksEnableEpoch -
func (stub *EnableEpochsHandlerStub) RefactorPeersMiniBlocksEnableEpoch() uint32 {
	return stub.RefactorPeersMiniBlocksEnableEpochField
}

// IsSCDeployFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSCDeployFlagEnabled() bool {
	return stub.IsSCDeployFlagEnabledField
}

// IsBuiltInFunctionsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBuiltInFunctionsFlagEnabled() bool {
	return stub.IsBuiltInFunctionsFlagEnabledField
}

// IsRelayedTransactionsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRelayedTransactionsFlagEnabled() bool {
	return stub.IsRelayedTransactionsFlagEnabledField
}

// IsPenalizedTooMuchGasFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsPenalizedTooMuchGasFlagEnabled() bool {
	return stub.IsPenalizedTooMuchGasFlagEnabledField
}

// IsSwitchJailWaitingFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSwitchJailWaitingFlagEnabled() bool {
	return stub.IsSwitchJailWaitingFlagEnabledField
}

// IsBelowSignedThresholdFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBelowSignedThresholdFlagEnabled() bool {
	return stub.IsBelowSignedThresholdFlagEnabledField
}

// IsSwitchHysteresisForMinNodesFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSwitchHysteresisForMinNodesFlagEnabled() bool {
	return stub.IsSwitchHysteresisForMinNodesFlagEnabledField
}

// IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch() bool {
	return stub.IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpochField
}

// IsTransactionSignedWithTxHashFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsTransactionSignedWithTxHashFlagEnabled() bool {
	return stub.IsTransactionSignedWithTxHashFlagEnabledField
}

// IsMetaProtectionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsMetaProtectionFlagEnabled() bool {
	return stub.IsMetaProtectionFlagEnabledField
}

// IsAheadOfTimeGasUsageFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsAheadOfTimeGasUsageFlagEnabled() bool {
	return stub.IsAheadOfTimeGasUsageFlagEnabledField
}

// IsGasPriceModifierFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsGasPriceModifierFlagEnabled() bool {
	return stub.IsGasPriceModifierFlagEnabledField
}

// IsRepairCallbackFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRepairCallbackFlagEnabled() bool {
	return stub.IsRepairCallbackFlagEnabledField
}

// IsBalanceWaitingListsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBalanceWaitingListsFlagEnabled() bool {
	return stub.IsBalanceWaitingListsFlagEnabledField
}

// IsReturnDataToLastTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsReturnDataToLastTransferFlagEnabled() bool {
	return stub.IsReturnDataToLastTransferFlagEnabledField
}

// IsSenderInOutTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSenderInOutTransferFlagEnabled() bool {
	return stub.IsSenderInOutTransferFlagEnabledField
}

// IsStakeFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStakeFlagEnabled() bool {
	return stub.IsStakeFlagEnabledField
}

// IsStakingV2FlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStakingV2FlagEnabled() bool {
	return stub.IsStakingV2FlagEnabledField
}

// IsStakingV2OwnerFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStakingV2OwnerFlagEnabled() bool {
	return stub.IsStakingV2OwnerFlagEnabledField
}

// IsStakingV2FlagEnabledForActivationEpochCompleted -
func (stub *EnableEpochsHandlerStub) IsStakingV2FlagEnabledForActivationEpochCompleted() bool {
	return stub.IsStakingV2FlagEnabledForActivationEpochCompletedField
}

// IsDoubleKeyProtectionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDoubleKeyProtectionFlagEnabled() bool {
	return stub.IsDoubleKeyProtectionFlagEnabledField
}

// IsESDTFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTFlagEnabled() bool {
	return stub.IsESDTFlagEnabledField
}

// IsESDTFlagEnabledForCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsESDTFlagEnabledForCurrentEpoch() bool {
	return stub.IsESDTFlagEnabledForCurrentEpochField
}

// IsGovernanceFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsGovernanceFlagEnabled() bool {
	return stub.IsGovernanceFlagEnabledField
}

// IsGovernanceFlagEnabledForCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsGovernanceFlagEnabledForCurrentEpoch() bool {
	return stub.IsGovernanceFlagEnabledForCurrentEpochField
}

// IsDelegationManagerFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDelegationManagerFlagEnabled() bool {
	return stub.IsDelegationManagerFlagEnabledField
}

// IsDelegationSmartContractFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDelegationSmartContractFlagEnabled() bool {
	return stub.IsDelegationSmartContractFlagEnabledField
}

// IsDelegationSmartContractFlagEnabledForCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsDelegationSmartContractFlagEnabledForCurrentEpoch() bool {
	return stub.IsDelegationSmartContractFlagForCurrentEpochEnabledField
}

// IsCorrectLastUnJailedFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCorrectLastUnJailedFlagEnabled() bool {
	return stub.IsCorrectLastUnJailedFlagEnabledField
}

// IsCorrectLastUnJailedFlagEnabledForCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsCorrectLastUnJailedFlagEnabledForCurrentEpoch() bool {
	return stub.IsCorrectLastUnJailedFlagEnabledForCurrentEpochField
}

// IsRelayedTransactionsV2FlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRelayedTransactionsV2FlagEnabled() bool {
	return stub.IsRelayedTransactionsV2FlagEnabledField
}

// IsUnBondTokensV2FlagEnabled -
func (stub *EnableEpochsHandlerStub) IsUnBondTokensV2FlagEnabled() bool {
	return stub.IsUnBondTokensV2FlagEnabledField
}

// IsSaveJailedAlwaysFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSaveJailedAlwaysFlagEnabled() bool {
	return stub.IsSaveJailedAlwaysFlagEnabledField
}

// IsReDelegateBelowMinCheckFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsReDelegateBelowMinCheckFlagEnabled() bool {
	return stub.IsReDelegateBelowMinCheckFlagEnabledField
}

// IsValidatorToDelegationFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsValidatorToDelegationFlagEnabled() bool {
	return stub.IsValidatorToDelegationFlagEnabledField
}

// IsWaitingListFixFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsWaitingListFixFlagEnabled() bool {
	return stub.IsWaitingListFixFlagEnabledField
}

// IsIncrementSCRNonceInMultiTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsIncrementSCRNonceInMultiTransferFlagEnabled() bool {
	return stub.IsIncrementSCRNonceInMultiTransferFlagEnabledField
}

// IsESDTMultiTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTMultiTransferFlagEnabled() bool {
	return stub.IsESDTMultiTransferFlagEnabledField
}

// IsGlobalMintBurnFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsGlobalMintBurnFlagEnabled() bool {
	return stub.IsGlobalMintBurnFlagEnabledField
}

// IsESDTTransferRoleFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTTransferRoleFlagEnabled() bool {
	return stub.IsESDTTransferRoleFlagEnabledField
}

// IsBuiltInFunctionOnMetaFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBuiltInFunctionOnMetaFlagEnabled() bool {
	return stub.IsBuiltInFunctionOnMetaFlagEnabledField
}

// IsComputeRewardCheckpointFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsComputeRewardCheckpointFlagEnabled() bool {
	return stub.IsComputeRewardCheckpointFlagEnabledField
}

// IsSCRSizeInvariantCheckFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSCRSizeInvariantCheckFlagEnabled() bool {
	return stub.IsSCRSizeInvariantCheckFlagEnabledField
}

// IsBackwardCompSaveKeyValueFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBackwardCompSaveKeyValueFlagEnabled() bool {
	return stub.IsBackwardCompSaveKeyValueFlagEnabledField
}

// IsESDTNFTCreateOnMultiShardFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTNFTCreateOnMultiShardFlagEnabled() bool {
	return stub.IsESDTNFTCreateOnMultiShardFlagEnabledField
}

// IsMetaESDTSetFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsMetaESDTSetFlagEnabled() bool {
	return stub.IsMetaESDTSetFlagEnabledField
}

// IsAddTokensToDelegationFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsAddTokensToDelegationFlagEnabled() bool {
	return stub.IsAddTokensToDelegationFlagEnabledField
}

// IsMultiESDTTransferFixOnCallBackFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsMultiESDTTransferFixOnCallBackFlagEnabled() bool {
	return stub.IsMultiESDTTransferFixOnCallBackFlagEnabledField
}

// IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled() bool {
	return stub.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledField
}

// IsCorrectFirstQueuedFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCorrectFirstQueuedFlagEnabled() bool {
	return stub.IsCorrectFirstQueuedFlagEnabledField
}

// IsDeleteDelegatorAfterClaimRewardsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDeleteDelegatorAfterClaimRewardsFlagEnabled() bool {
	return stub.IsDeleteDelegatorAfterClaimRewardsFlagEnabledField
}

// IsFixOOGReturnCodeFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsFixOOGReturnCodeFlagEnabled() bool {
	return stub.IsFixOOGReturnCodeFlagEnabledField
}

// IsRemoveNonUpdatedStorageFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRemoveNonUpdatedStorageFlagEnabled() bool {
	return stub.IsRemoveNonUpdatedStorageFlagEnabledField
}

// IsOptimizeNFTStoreFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsOptimizeNFTStoreFlagEnabled() bool {
	return stub.IsOptimizeNFTStoreFlagEnabledField
}

// IsCreateNFTThroughExecByCallerFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCreateNFTThroughExecByCallerFlagEnabled() bool {
	return stub.IsCreateNFTThroughExecByCallerFlagEnabledField
}

// IsStopDecreasingValidatorRatingWhenStuckFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStopDecreasingValidatorRatingWhenStuckFlagEnabled() bool {
	return stub.IsStopDecreasingValidatorRatingWhenStuckFlagEnabledField
}

// IsFrontRunningProtectionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsFrontRunningProtectionFlagEnabled() bool {
	return stub.IsFrontRunningProtectionFlagEnabledField
}

// IsPayableBySCFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsPayableBySCFlagEnabled() bool {
	return stub.IsPayableBySCFlagEnabledField
}

// IsCleanUpInformativeSCRsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCleanUpInformativeSCRsFlagEnabled() bool {
	return stub.IsCleanUpInformativeSCRsFlagEnabledField
}

// IsStorageAPICostOptimizationFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStorageAPICostOptimizationFlagEnabled() bool {
	return stub.IsStorageAPICostOptimizationFlagEnabledField
}

// IsESDTRegisterAndSetAllRolesFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTRegisterAndSetAllRolesFlagEnabled() bool {
	return stub.IsESDTRegisterAndSetAllRolesFlagEnabledField
}

// IsScheduledMiniBlocksFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsScheduledMiniBlocksFlagEnabled() bool {
	return stub.IsScheduledMiniBlocksFlagEnabledField
}

// IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled() bool {
	return stub.IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledField
}

// IsDoNotReturnOldBlockInBlockchainHookFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDoNotReturnOldBlockInBlockchainHookFlagEnabled() bool {
	return stub.IsDoNotReturnOldBlockInBlockchainHookFlagEnabledField
}

// IsAddFailedRelayedTxToInvalidMBsFlag -
func (stub *EnableEpochsHandlerStub) IsAddFailedRelayedTxToInvalidMBsFlag() bool {
	return stub.IsAddFailedRelayedTxToInvalidMBsFlagField
}

// IsSCRSizeInvariantOnBuiltInResultFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSCRSizeInvariantOnBuiltInResultFlagEnabled() bool {
	return stub.IsSCRSizeInvariantOnBuiltInResultFlagEnabledField
}

// IsCheckCorrectTokenIDForTransferRoleFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCheckCorrectTokenIDForTransferRoleFlagEnabled() bool {
	return stub.IsCheckCorrectTokenIDForTransferRoleFlagEnabledField
}

// IsFailExecutionOnEveryAPIErrorFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsFailExecutionOnEveryAPIErrorFlagEnabled() bool {
	return stub.IsFailExecutionOnEveryAPIErrorFlagEnabledField
}

// IsHeartbeatDisableFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsHeartbeatDisableFlagEnabled() bool {
	return stub.IsHeartbeatDisableFlagEnabledField
}

// IsMiniBlockPartialExecutionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsMiniBlockPartialExecutionFlagEnabled() bool {
	return stub.IsMiniBlockPartialExecutionFlagEnabledField
}

// IsManagedCryptoAPIsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsManagedCryptoAPIsFlagEnabled() bool {
	return stub.IsManagedCryptoAPIsFlagEnabledField
}

// IsESDTMetadataContinuousCleanupFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTMetadataContinuousCleanupFlagEnabled() bool {
	return stub.IsESDTMetadataContinuousCleanupFlagEnabledField
}

// IsDisableExecByCallerFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDisableExecByCallerFlagEnabled() bool {
	return stub.IsDisableExecByCallerFlagEnabledField
}

// IsRefactorContextFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRefactorContextFlagEnabled() bool {
	return stub.IsRefactorContextFlagEnabledField
}

// IsCheckFunctionArgumentFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCheckFunctionArgumentFlagEnabled() bool {
	return stub.IsCheckFunctionArgumentFlagEnabledField
}

// IsCheckExecuteOnReadOnlyFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCheckExecuteOnReadOnlyFlagEnabled() bool {
	return stub.IsCheckExecuteOnReadOnlyFlagEnabledField
}

// IsFixAsyncCallbackCheckFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsFixAsyncCallbackCheckFlagEnabled() bool {
	return stub.IsFixAsyncCallbackCheckFlagEnabledField
}

// IsSaveToSystemAccountFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSaveToSystemAccountFlagEnabled() bool {
	return stub.IsSaveToSystemAccountFlagEnabledField
}

// IsCheckFrozenCollectionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCheckFrozenCollectionFlagEnabled() bool {
	return stub.IsCheckFrozenCollectionFlagEnabledField
}

// IsSendAlwaysFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSendAlwaysFlagEnabled() bool {
	return stub.IsSendAlwaysFlagEnabledField
}

// IsValueLengthCheckFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsValueLengthCheckFlagEnabled() bool {
	return stub.IsValueLengthCheckFlagEnabledField
}

// IsCheckTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCheckTransferFlagEnabled() bool {
	return stub.IsCheckTransferFlagEnabledField
}

// IsTransferToMetaFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsTransferToMetaFlagEnabled() bool {
	return stub.IsTransferToMetaFlagEnabledField
}

// IsESDTNFTImprovementV1FlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTNFTImprovementV1FlagEnabled() bool {
	return stub.IsESDTNFTImprovementV1FlagEnabledField
}

// IsSetSenderInEeiOutputTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSetSenderInEeiOutputTransferFlagEnabled() bool {
	return stub.IsSetSenderInEeiOutputTransferFlagEnabledField
}

// IsChangeDelegationOwnerFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsChangeDelegationOwnerFlagEnabled() bool {
	return stub.IsChangeDelegationOwnerFlagEnabledField
}

// IsRefactorPeersMiniBlocksFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRefactorPeersMiniBlocksFlagEnabled() bool {
	return stub.IsRefactorPeersMiniBlocksFlagEnabledField
}

// IsInterfaceNil -
func (stub *EnableEpochsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
