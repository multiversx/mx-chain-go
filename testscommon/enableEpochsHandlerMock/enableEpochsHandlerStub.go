package enableEpochsHandlerMock

import "sync"

// EnableEpochsHandlerStub -
type EnableEpochsHandlerStub struct {
	sync.RWMutex
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
	IsFixAsyncCallBackArgsListFlagEnabledField                   bool
	IsFixOldTokenLiquidityEnabledField                           bool
	IsRuntimeMemStoreLimitEnabledField                           bool
	IsRuntimeCodeSizeFixEnabledField                             bool
	IsMaxBlockchainHookCountersFlagEnabledField                  bool
	IsWipeSingleNFTLiquidityDecreaseEnabledField                 bool
	IsAlwaysSaveTokenMetaDataEnabledField                        bool
	IsSetGuardianEnabledField                                    bool
	IsRelayedNonceFixEnabledField                                bool
	IsKeepExecOrderOnCreatedSCRsEnabledField                     bool
	IsMultiClaimOnDelegationEnabledField                         bool
	IsChangeUsernameEnabledField                                 bool
	IsConsistentTokensValuesLengthCheckEnabledField              bool
	IsAutoBalanceDataTriesEnabledField                           bool
	FixDelegationChangeOwnerOnAccountEnabledField                bool
	IsConsensusModelV2EnabledField                               bool
}

// ResetPenalizedTooMuchGasFlag -
func (stub *EnableEpochsHandlerStub) ResetPenalizedTooMuchGasFlag() {
	if stub.ResetPenalizedTooMuchGasFlagCalled != nil {
		stub.ResetPenalizedTooMuchGasFlagCalled()
	}
}

// BlockGasAndFeesReCheckEnableEpoch -
func (stub *EnableEpochsHandlerStub) BlockGasAndFeesReCheckEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.BlockGasAndFeesReCheckEnableEpochField
}

// StakingV2EnableEpoch -
func (stub *EnableEpochsHandlerStub) StakingV2EnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.StakingV2EnableEpochField
}

// ScheduledMiniBlocksEnableEpoch -
func (stub *EnableEpochsHandlerStub) ScheduledMiniBlocksEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.ScheduledMiniBlocksEnableEpochField
}

// SwitchJailWaitingEnableEpoch -
func (stub *EnableEpochsHandlerStub) SwitchJailWaitingEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.SwitchJailWaitingEnableEpochField
}

// BalanceWaitingListsEnableEpoch -
func (stub *EnableEpochsHandlerStub) BalanceWaitingListsEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.BalanceWaitingListsEnableEpochField
}

// WaitingListFixEnableEpoch -
func (stub *EnableEpochsHandlerStub) WaitingListFixEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.WaitingListFixEnableEpochField
}

// MultiESDTTransferAsyncCallBackEnableEpoch -
func (stub *EnableEpochsHandlerStub) MultiESDTTransferAsyncCallBackEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.MultiESDTTransferAsyncCallBackEnableEpochField
}

// FixOOGReturnCodeEnableEpoch -
func (stub *EnableEpochsHandlerStub) FixOOGReturnCodeEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.FixOOGReturnCodeEnableEpochField
}

// RemoveNonUpdatedStorageEnableEpoch -
func (stub *EnableEpochsHandlerStub) RemoveNonUpdatedStorageEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.RemoveNonUpdatedStorageEnableEpochField
}

// CreateNFTThroughExecByCallerEnableEpoch -
func (stub *EnableEpochsHandlerStub) CreateNFTThroughExecByCallerEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.CreateNFTThroughExecByCallerEnableEpochField
}

// FixFailExecutionOnErrorEnableEpoch -
func (stub *EnableEpochsHandlerStub) FixFailExecutionOnErrorEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.FixFailExecutionOnErrorEnableEpochField
}

// ManagedCryptoAPIEnableEpoch -
func (stub *EnableEpochsHandlerStub) ManagedCryptoAPIEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.ManagedCryptoAPIEnableEpochField
}

// DisableExecByCallerEnableEpoch -
func (stub *EnableEpochsHandlerStub) DisableExecByCallerEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.DisableExecByCallerEnableEpochField
}

// RefactorContextEnableEpoch -
func (stub *EnableEpochsHandlerStub) RefactorContextEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.RefactorContextEnableEpochField
}

// CheckExecuteReadOnlyEnableEpoch -
func (stub *EnableEpochsHandlerStub) CheckExecuteReadOnlyEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.CheckExecuteReadOnlyEnableEpochField
}

// StorageAPICostOptimizationEnableEpoch -
func (stub *EnableEpochsHandlerStub) StorageAPICostOptimizationEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.StorageAPICostOptimizationEnableEpochField
}

// MiniBlockPartialExecutionEnableEpoch -
func (stub *EnableEpochsHandlerStub) MiniBlockPartialExecutionEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.MiniBlockPartialExecutionEnableEpochField
}

// RefactorPeersMiniBlocksEnableEpoch -
func (stub *EnableEpochsHandlerStub) RefactorPeersMiniBlocksEnableEpoch() uint32 {
	stub.RLock()
	defer stub.RUnlock()

	return stub.RefactorPeersMiniBlocksEnableEpochField
}

// IsSCDeployFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSCDeployFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsSCDeployFlagEnabledField
}

// IsBuiltInFunctionsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBuiltInFunctionsFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsBuiltInFunctionsFlagEnabledField
}

// IsRelayedTransactionsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRelayedTransactionsFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsRelayedTransactionsFlagEnabledField
}

// IsPenalizedTooMuchGasFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsPenalizedTooMuchGasFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsPenalizedTooMuchGasFlagEnabledField
}

// IsSwitchJailWaitingFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSwitchJailWaitingFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsSwitchJailWaitingFlagEnabledField
}

// IsBelowSignedThresholdFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBelowSignedThresholdFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsBelowSignedThresholdFlagEnabledField
}

// IsSwitchHysteresisForMinNodesFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSwitchHysteresisForMinNodesFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsSwitchHysteresisForMinNodesFlagEnabledField
}

// IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpochField
}

// IsTransactionSignedWithTxHashFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsTransactionSignedWithTxHashFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsTransactionSignedWithTxHashFlagEnabledField
}

// IsMetaProtectionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsMetaProtectionFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsMetaProtectionFlagEnabledField
}

// IsAheadOfTimeGasUsageFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsAheadOfTimeGasUsageFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsAheadOfTimeGasUsageFlagEnabledField
}

// IsGasPriceModifierFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsGasPriceModifierFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsGasPriceModifierFlagEnabledField
}

// IsRepairCallbackFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRepairCallbackFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsRepairCallbackFlagEnabledField
}

// IsBalanceWaitingListsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBalanceWaitingListsFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsBalanceWaitingListsFlagEnabledField
}

// IsReturnDataToLastTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsReturnDataToLastTransferFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsReturnDataToLastTransferFlagEnabledField
}

// IsSenderInOutTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSenderInOutTransferFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsSenderInOutTransferFlagEnabledField
}

// IsStakeFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStakeFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsStakeFlagEnabledField
}

// IsStakingV2FlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStakingV2FlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsStakingV2FlagEnabledField
}

// IsStakingV2OwnerFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStakingV2OwnerFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsStakingV2OwnerFlagEnabledField
}

// IsStakingV2FlagEnabledForActivationEpochCompleted -
func (stub *EnableEpochsHandlerStub) IsStakingV2FlagEnabledForActivationEpochCompleted() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsStakingV2FlagEnabledForActivationEpochCompletedField
}

// IsDoubleKeyProtectionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDoubleKeyProtectionFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsDoubleKeyProtectionFlagEnabledField
}

// IsESDTFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsESDTFlagEnabledField
}

// IsESDTFlagEnabledForCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsESDTFlagEnabledForCurrentEpoch() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsESDTFlagEnabledForCurrentEpochField
}

// IsGovernanceFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsGovernanceFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsGovernanceFlagEnabledField
}

// IsGovernanceFlagEnabledForCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsGovernanceFlagEnabledForCurrentEpoch() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsGovernanceFlagEnabledForCurrentEpochField
}

// IsDelegationManagerFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDelegationManagerFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsDelegationManagerFlagEnabledField
}

// IsDelegationSmartContractFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDelegationSmartContractFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsDelegationSmartContractFlagEnabledField
}

// IsDelegationSmartContractFlagEnabledForCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsDelegationSmartContractFlagEnabledForCurrentEpoch() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsDelegationSmartContractFlagForCurrentEpochEnabledField
}

// IsCorrectLastUnJailedFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCorrectLastUnJailedFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsCorrectLastUnJailedFlagEnabledField
}

// IsCorrectLastUnJailedFlagEnabledForCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsCorrectLastUnJailedFlagEnabledForCurrentEpoch() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsCorrectLastUnJailedFlagEnabledForCurrentEpochField
}

// IsRelayedTransactionsV2FlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRelayedTransactionsV2FlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsRelayedTransactionsV2FlagEnabledField
}

// IsUnBondTokensV2FlagEnabled -
func (stub *EnableEpochsHandlerStub) IsUnBondTokensV2FlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsUnBondTokensV2FlagEnabledField
}

// IsSaveJailedAlwaysFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSaveJailedAlwaysFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsSaveJailedAlwaysFlagEnabledField
}

// IsReDelegateBelowMinCheckFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsReDelegateBelowMinCheckFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsReDelegateBelowMinCheckFlagEnabledField
}

// IsValidatorToDelegationFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsValidatorToDelegationFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsValidatorToDelegationFlagEnabledField
}

// IsWaitingListFixFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsWaitingListFixFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsWaitingListFixFlagEnabledField
}

// IsIncrementSCRNonceInMultiTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsIncrementSCRNonceInMultiTransferFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsIncrementSCRNonceInMultiTransferFlagEnabledField
}

// IsESDTMultiTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTMultiTransferFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsESDTMultiTransferFlagEnabledField
}

// IsGlobalMintBurnFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsGlobalMintBurnFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsGlobalMintBurnFlagEnabledField
}

// IsESDTTransferRoleFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTTransferRoleFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsESDTTransferRoleFlagEnabledField
}

// IsBuiltInFunctionOnMetaFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBuiltInFunctionOnMetaFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsBuiltInFunctionOnMetaFlagEnabledField
}

// IsComputeRewardCheckpointFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsComputeRewardCheckpointFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsComputeRewardCheckpointFlagEnabledField
}

// IsSCRSizeInvariantCheckFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSCRSizeInvariantCheckFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsSCRSizeInvariantCheckFlagEnabledField
}

// IsBackwardCompSaveKeyValueFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsBackwardCompSaveKeyValueFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsBackwardCompSaveKeyValueFlagEnabledField
}

// IsESDTNFTCreateOnMultiShardFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTNFTCreateOnMultiShardFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsESDTNFTCreateOnMultiShardFlagEnabledField
}

// IsMetaESDTSetFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsMetaESDTSetFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsMetaESDTSetFlagEnabledField
}

// IsAddTokensToDelegationFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsAddTokensToDelegationFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsAddTokensToDelegationFlagEnabledField
}

// IsMultiESDTTransferFixOnCallBackFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsMultiESDTTransferFixOnCallBackFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsMultiESDTTransferFixOnCallBackFlagEnabledField
}

// IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledField
}

// IsCorrectFirstQueuedFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCorrectFirstQueuedFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsCorrectFirstQueuedFlagEnabledField
}

// IsDeleteDelegatorAfterClaimRewardsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDeleteDelegatorAfterClaimRewardsFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsDeleteDelegatorAfterClaimRewardsFlagEnabledField
}

// IsFixOOGReturnCodeFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsFixOOGReturnCodeFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsFixOOGReturnCodeFlagEnabledField
}

// IsRemoveNonUpdatedStorageFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRemoveNonUpdatedStorageFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsRemoveNonUpdatedStorageFlagEnabledField
}

// IsOptimizeNFTStoreFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsOptimizeNFTStoreFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsOptimizeNFTStoreFlagEnabledField
}

// IsCreateNFTThroughExecByCallerFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCreateNFTThroughExecByCallerFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsCreateNFTThroughExecByCallerFlagEnabledField
}

// IsStopDecreasingValidatorRatingWhenStuckFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStopDecreasingValidatorRatingWhenStuckFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsStopDecreasingValidatorRatingWhenStuckFlagEnabledField
}

// IsFrontRunningProtectionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsFrontRunningProtectionFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsFrontRunningProtectionFlagEnabledField
}

// IsPayableBySCFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsPayableBySCFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsPayableBySCFlagEnabledField
}

// IsCleanUpInformativeSCRsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCleanUpInformativeSCRsFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsCleanUpInformativeSCRsFlagEnabledField
}

// IsStorageAPICostOptimizationFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsStorageAPICostOptimizationFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsStorageAPICostOptimizationFlagEnabledField
}

// IsESDTRegisterAndSetAllRolesFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTRegisterAndSetAllRolesFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsESDTRegisterAndSetAllRolesFlagEnabledField
}

// IsScheduledMiniBlocksFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsScheduledMiniBlocksFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsScheduledMiniBlocksFlagEnabledField
}

// IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledField
}

// IsDoNotReturnOldBlockInBlockchainHookFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDoNotReturnOldBlockInBlockchainHookFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsDoNotReturnOldBlockInBlockchainHookFlagEnabledField
}

// IsAddFailedRelayedTxToInvalidMBsFlag -
func (stub *EnableEpochsHandlerStub) IsAddFailedRelayedTxToInvalidMBsFlag() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsAddFailedRelayedTxToInvalidMBsFlagField
}

// IsSCRSizeInvariantOnBuiltInResultFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSCRSizeInvariantOnBuiltInResultFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsSCRSizeInvariantOnBuiltInResultFlagEnabledField
}

// IsCheckCorrectTokenIDForTransferRoleFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCheckCorrectTokenIDForTransferRoleFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsCheckCorrectTokenIDForTransferRoleFlagEnabledField
}

// IsFailExecutionOnEveryAPIErrorFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsFailExecutionOnEveryAPIErrorFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsFailExecutionOnEveryAPIErrorFlagEnabledField
}

// IsMiniBlockPartialExecutionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsMiniBlockPartialExecutionFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsMiniBlockPartialExecutionFlagEnabledField
}

// IsManagedCryptoAPIsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsManagedCryptoAPIsFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsManagedCryptoAPIsFlagEnabledField
}

// IsESDTMetadataContinuousCleanupFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTMetadataContinuousCleanupFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsESDTMetadataContinuousCleanupFlagEnabledField
}

// IsDisableExecByCallerFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsDisableExecByCallerFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsDisableExecByCallerFlagEnabledField
}

// IsRefactorContextFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRefactorContextFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsRefactorContextFlagEnabledField
}

// IsCheckFunctionArgumentFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCheckFunctionArgumentFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsCheckFunctionArgumentFlagEnabledField
}

// IsCheckExecuteOnReadOnlyFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCheckExecuteOnReadOnlyFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsCheckExecuteOnReadOnlyFlagEnabledField
}

// IsFixAsyncCallbackCheckFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsFixAsyncCallbackCheckFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsFixAsyncCallbackCheckFlagEnabledField
}

// IsSaveToSystemAccountFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSaveToSystemAccountFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsSaveToSystemAccountFlagEnabledField
}

// IsCheckFrozenCollectionFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCheckFrozenCollectionFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsCheckFrozenCollectionFlagEnabledField
}

// IsSendAlwaysFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSendAlwaysFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsSendAlwaysFlagEnabledField
}

// IsValueLengthCheckFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsValueLengthCheckFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsValueLengthCheckFlagEnabledField
}

// IsCheckTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsCheckTransferFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsCheckTransferFlagEnabledField
}

// IsTransferToMetaFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsTransferToMetaFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsTransferToMetaFlagEnabledField
}

// IsESDTNFTImprovementV1FlagEnabled -
func (stub *EnableEpochsHandlerStub) IsESDTNFTImprovementV1FlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsESDTNFTImprovementV1FlagEnabledField
}

// IsSetSenderInEeiOutputTransferFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsSetSenderInEeiOutputTransferFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsSetSenderInEeiOutputTransferFlagEnabledField
}

// IsChangeDelegationOwnerFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsChangeDelegationOwnerFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsChangeDelegationOwnerFlagEnabledField
}

// IsRefactorPeersMiniBlocksFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsRefactorPeersMiniBlocksFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsRefactorPeersMiniBlocksFlagEnabledField
}

// IsFixAsyncCallBackArgsListFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsFixAsyncCallBackArgsListFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsFixAsyncCallBackArgsListFlagEnabledField
}

// IsFixOldTokenLiquidityEnabled -
func (stub *EnableEpochsHandlerStub) IsFixOldTokenLiquidityEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsFixOldTokenLiquidityEnabledField
}

// IsRuntimeMemStoreLimitEnabled -
func (stub *EnableEpochsHandlerStub) IsRuntimeMemStoreLimitEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsRuntimeMemStoreLimitEnabledField
}

// IsRuntimeCodeSizeFixEnabled -
func (stub *EnableEpochsHandlerStub) IsRuntimeCodeSizeFixEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsRuntimeCodeSizeFixEnabledField
}

// IsMaxBlockchainHookCountersFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsMaxBlockchainHookCountersFlagEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsMaxBlockchainHookCountersFlagEnabledField
}

// IsWipeSingleNFTLiquidityDecreaseEnabled -
func (stub *EnableEpochsHandlerStub) IsWipeSingleNFTLiquidityDecreaseEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsWipeSingleNFTLiquidityDecreaseEnabledField
}

// IsAlwaysSaveTokenMetaDataEnabled -
func (stub *EnableEpochsHandlerStub) IsAlwaysSaveTokenMetaDataEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsAlwaysSaveTokenMetaDataEnabledField
}

// IsSetGuardianEnabled -
func (stub *EnableEpochsHandlerStub) IsSetGuardianEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsSetGuardianEnabledField
}

// IsRelayedNonceFixEnabled -
func (stub *EnableEpochsHandlerStub) IsRelayedNonceFixEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsRelayedNonceFixEnabledField
}

// IsKeepExecOrderOnCreatedSCRsEnabled -
func (stub *EnableEpochsHandlerStub) IsKeepExecOrderOnCreatedSCRsEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsKeepExecOrderOnCreatedSCRsEnabledField
}

// IsMultiClaimOnDelegationEnabled -
func (stub *EnableEpochsHandlerStub) IsMultiClaimOnDelegationEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsMultiClaimOnDelegationEnabledField
}

// IsChangeUsernameEnabled -
func (stub *EnableEpochsHandlerStub) IsChangeUsernameEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsChangeUsernameEnabledField
}

// IsConsistentTokensValuesLengthCheckEnabled -
func (stub *EnableEpochsHandlerStub) IsConsistentTokensValuesLengthCheckEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsConsistentTokensValuesLengthCheckEnabledField
}

// IsAutoBalanceDataTriesEnabled -
func (stub *EnableEpochsHandlerStub) IsAutoBalanceDataTriesEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsAutoBalanceDataTriesEnabledField
}

// FixDelegationChangeOwnerOnAccountEnabled -
func (stub *EnableEpochsHandlerStub) FixDelegationChangeOwnerOnAccountEnabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.FixDelegationChangeOwnerOnAccountEnabledField
}

// IsConsensusModelV2Enabled -
func (stub *EnableEpochsHandlerStub) IsConsensusModelV2Enabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsConsensusModelV2EnabledField
}

// IsInterfaceNil -
func (stub *EnableEpochsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
