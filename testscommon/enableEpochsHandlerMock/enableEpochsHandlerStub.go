package enableEpochsHandlerMock

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
)

// EnableEpochsHandlerStub -
type EnableEpochsHandlerStub struct {
	sync.RWMutex
	GetCurrentEpochCalled                                             func() uint32
	BlockGasAndFeesReCheckEnableEpochField                            uint32
	StakingV2EnableEpochField                                         uint32
	ScheduledMiniBlocksEnableEpochField                               uint32
	SwitchJailWaitingEnableEpochField                                 uint32
	BalanceWaitingListsEnableEpochField                               uint32
	WaitingListFixEnableEpochField                                    uint32
	MultiESDTTransferAsyncCallBackEnableEpochField                    uint32
	FixOOGReturnCodeEnableEpochField                                  uint32
	RemoveNonUpdatedStorageEnableEpochField                           uint32
	CreateNFTThroughExecByCallerEnableEpochField                      uint32
	FixFailExecutionOnErrorEnableEpochField                           uint32
	ManagedCryptoAPIEnableEpochField                                  uint32
	DisableExecByCallerEnableEpochField                               uint32
	RefactorContextEnableEpochField                                   uint32
	CheckExecuteReadOnlyEnableEpochField                              uint32
	StorageAPICostOptimizationEnableEpochField                        uint32
	MiniBlockPartialExecutionEnableEpochField                         uint32
	RefactorPeersMiniBlocksEnableEpochField                           uint32
	IsSCDeployFlagEnabledInEpochCalled                                func(epoch uint32) bool
	IsBuiltInFunctionsFlagEnabledInEpochCalled                        func(epoch uint32) bool
	IsRelayedTransactionsFlagEnabledInEpochCalled                     func(epoch uint32) bool
	IsPenalizedTooMuchGasFlagEnabledInEpochCalled                     func(epoch uint32) bool
	IsSwitchJailWaitingFlagEnabledInEpochCalled                       func(epoch uint32) bool
	IsBelowSignedThresholdFlagEnabledInEpochCalled                    func(epoch uint32) bool
	IsSwitchHysteresisForMinNodesFlagEnabledInSpecificEpochOnlyCalled func(epoch uint32) bool
	IsTransactionSignedWithTxHashFlagEnabledInEpochCalled             func(epoch uint32) bool
	IsMetaProtectionFlagEnabledInEpochCalled                          func(epoch uint32) bool
	IsAheadOfTimeGasUsageFlagEnabledInEpochCalled                     func(epoch uint32) bool
	IsGasPriceModifierFlagEnabledInEpochCalled                        func(epoch uint32) bool
	IsRepairCallbackFlagEnabledInEpochCalled                          func(epoch uint32) bool
	IsReturnDataToLastTransferFlagEnabledAfterEpochCalled             func(epoch uint32) bool
	IsSenderInOutTransferFlagEnabledInEpochCalled                     func(epoch uint32) bool
	IsStakeFlagEnabledInEpochCalled                                   func(epoch uint32) bool
	IsStakingV2FlagEnabledInEpochCalled                               func(epoch uint32) bool
	IsStakingV2OwnerFlagEnabledInSpecificEpochOnlyCalled              func(epoch uint32) bool
	IsStakingV2FlagEnabledAfterEpochCalled                            func(epoch uint32) bool
	IsDoubleKeyProtectionFlagEnabledInEpochCalled                     func(epoch uint32) bool
	IsESDTFlagEnabledInEpochCalled                                    func(epoch uint32) bool
	IsESDTFlagEnabledInSpecificEpochOnlyCalled                        func(epoch uint32) bool
	IsGovernanceFlagEnabledInEpochCalled                              func(epoch uint32) bool
	IsGovernanceFlagEnabledInSpecificEpochOnlyCalled                  func(epoch uint32) bool
	IsDelegationManagerFlagEnabledInEpochCalled                       func(epoch uint32) bool
	IsDelegationSmartContractFlagEnabledInEpochCalled                 func(epoch uint32) bool
	IsDelegationSmartContractFlagEnabledInSpecificEpochOnlyCalled     func(epoch uint32) bool
	IsCorrectLastUnJailedFlagEnabledInEpochCalled                     func(epoch uint32) bool
	IsCorrectLastUnJailedFlagEnabledInSpecificEpochOnlyCalled         func(epoch uint32) bool
	IsRelayedTransactionsV2FlagEnabledInEpochCalled                   func(epoch uint32) bool
	IsUnBondTokensV2FlagEnabledInEpochCalled                          func(epoch uint32) bool
	IsSaveJailedAlwaysFlagEnabledInEpochCalled                        func(epoch uint32) bool
	IsReDelegateBelowMinCheckFlagEnabledInEpochCalled                 func(epoch uint32) bool
	IsValidatorToDelegationFlagEnabledInEpochCalled                   func(epoch uint32) bool
	IsIncrementSCRNonceInMultiTransferFlagEnabledInEpochCalled        func(epoch uint32) bool
	IsESDTMultiTransferFlagEnabledInEpochCalled                       func(epoch uint32) bool
	IsGlobalMintBurnFlagEnabledInEpochCalled                          func(epoch uint32) bool
	IsESDTTransferRoleFlagEnabledInEpochCalled                        func(epoch uint32) bool
	IsBuiltInFunctionOnMetaFlagEnabledInEpochCalled                   func(epoch uint32) bool
	IsComputeRewardCheckpointFlagEnabledInEpochCalled                 func(epoch uint32) bool
	IsSCRSizeInvariantCheckFlagEnabledInEpochCalled                   func(epoch uint32) bool
	IsBackwardCompSaveKeyValueFlagEnabledInEpochCalled                func(epoch uint32) bool
	IsESDTNFTCreateOnMultiShardFlagEnabledInEpochCalled               func(epoch uint32) bool
	IsMetaESDTSetFlagEnabledInEpochCalled                             func(epoch uint32) bool
	IsAddTokensToDelegationFlagEnabledInEpochCalled                   func(epoch uint32) bool
	IsMultiESDTTransferFixOnCallBackFlagEnabledInEpochCalled          func(epoch uint32) bool
	IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledInEpochCalled        func(epoch uint32) bool
	IsCorrectFirstQueuedFlagEnabledInEpochCalled                      func(epoch uint32) bool
	IsDeleteDelegatorAfterClaimRewardsFlagEnabledInEpochCalled        func(epoch uint32) bool
	IsRemoveNonUpdatedStorageFlagEnabledInEpochCalled                 func(epoch uint32) bool
	IsOptimizeNFTStoreFlagEnabledInEpochCalled                        func(epoch uint32) bool
	IsCreateNFTThroughExecByCallerFlagEnabledInEpochCalled            func(epoch uint32) bool
	IsStopDecreasingValidatorRatingWhenStuckFlagEnabledInEpochCalled  func(epoch uint32) bool
	IsFrontRunningProtectionFlagEnabledInEpochCalled                  func(epoch uint32) bool
	IsPayableBySCFlagEnabledInEpochCalled                             func(epoch uint32) bool
	IsCleanUpInformativeSCRsFlagEnabledInEpochCalled                  func(epoch uint32) bool
	IsStorageAPICostOptimizationFlagEnabledInEpochCalled              func(epoch uint32) bool
	IsESDTRegisterAndSetAllRolesFlagEnabledInEpochCalled              func(epoch uint32) bool
	IsScheduledMiniBlocksFlagEnabledInEpochCalled                     func(epoch uint32) bool
	IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledInEpochCalled      func(epoch uint32) bool
	IsDoNotReturnOldBlockInBlockchainHookFlagEnabledInEpochCalled     func(epoch uint32) bool
	IsAddFailedRelayedTxToInvalidMBsFlagEnabledInEpochCalled          func(epoch uint32) bool
	IsSCRSizeInvariantOnBuiltInResultFlagEnabledInEpochCalled         func(epoch uint32) bool
	IsCheckCorrectTokenIDForTransferRoleFlagEnabledInEpochCalled      func(epoch uint32) bool
	IsFailExecutionOnEveryAPIErrorFlagEnabledInEpochCalled            func(epoch uint32) bool
	IsMiniBlockPartialExecutionFlagEnabledInEpochCalled               func(epoch uint32) bool
	IsManagedCryptoAPIsFlagEnabledInEpochCalled                       func(epoch uint32) bool
	IsESDTMetadataContinuousCleanupFlagEnabledInEpochCalled           func(epoch uint32) bool
	IsDisableExecByCallerFlagEnabledInEpochCalled                     func(epoch uint32) bool
	IsRefactorContextFlagEnabledInEpochCalled                         func(epoch uint32) bool
	IsCheckFunctionArgumentFlagEnabledInEpochCalled                   func(epoch uint32) bool
	IsCheckExecuteOnReadOnlyFlagEnabledInEpochCalled                  func(epoch uint32) bool
	IsSetSenderInEeiOutputTransferFlagEnabledInEpochCalled            func(epoch uint32) bool
	IsFixAsyncCallbackCheckFlagEnabledInEpochCalled                   func(epoch uint32) bool
	IsSaveToSystemAccountFlagEnabledInEpochCalled                     func(epoch uint32) bool
	IsCheckFrozenCollectionFlagEnabledInEpochCalled                   func(epoch uint32) bool
	IsSendAlwaysFlagEnabledInEpochCalled                              func(epoch uint32) bool
	IsValueLengthCheckFlagEnabledInEpochCalled                        func(epoch uint32) bool
	IsCheckTransferFlagEnabledInEpochCalled                           func(epoch uint32) bool
	IsTransferToMetaFlagEnabledInEpochCalled                          func(epoch uint32) bool
	IsESDTNFTImprovementV1FlagEnabledInEpochCalled                    func(epoch uint32) bool
	IsChangeDelegationOwnerFlagEnabledInEpochCalled                   func(epoch uint32) bool
	IsRefactorPeersMiniBlocksFlagEnabledInEpochCalled                 func(epoch uint32) bool
	IsSCProcessorV2FlagEnabledInEpochCalled                           func(epoch uint32) bool
	IsFixAsyncCallBackArgsListFlagEnabledInEpochCalled                func(epoch uint32) bool
	IsFixOldTokenLiquidityEnabledInEpochCalled                        func(epoch uint32) bool
	IsRuntimeMemStoreLimitEnabledInEpochCalled                        func(epoch uint32) bool
	IsRuntimeCodeSizeFixEnabledInEpochCalled                          func(epoch uint32) bool
	IsMaxBlockchainHookCountersFlagEnabledInEpochCalled               func(epoch uint32) bool
	IsWipeSingleNFTLiquidityDecreaseEnabledInEpochCalled              func(epoch uint32) bool
	IsAlwaysSaveTokenMetaDataEnabledInEpochCalled                     func(epoch uint32) bool
	IsSetGuardianEnabledInEpochCalled                                 func(epoch uint32) bool
	IsRelayedNonceFixEnabledInEpochCalled                             func(epoch uint32) bool
	IsConsistentTokensValuesLengthCheckEnabledInEpochCalled           func(epoch uint32) bool
	IsKeepExecOrderOnCreatedSCRsEnabledInEpochCalled                  func(epoch uint32) bool
	IsMultiClaimOnDelegationEnabledInEpochCalled                      func(epoch uint32) bool
	IsChangeUsernameEnabledInEpochCalled                              func(epoch uint32) bool
	IsAutoBalanceDataTriesEnabledInEpochCalled                        func(epoch uint32) bool
	FixDelegationChangeOwnerOnAccountEnabledInEpochCalled             func(epoch uint32) bool
	IsFixOOGReturnCodeFlagEnabledInEpochCalled                        func(epoch uint32) bool
	IsDeterministicSortOnValidatorsInfoFixEnabledInEpochCalled        func(epoch uint32) bool
	IsFlagDefinedCalled                                               func(flag core.EnableEpochFlag) bool
	IsFlagEnabledInCurrentEpochCalled                                 func(flag core.EnableEpochFlag) bool
	IsFlagEnabledInEpochCalled                                        func(flag core.EnableEpochFlag, epoch uint32) bool
}

// IsFlagDefined -
func (stub *EnableEpochsHandlerStub) IsFlagDefined(flag core.EnableEpochFlag) bool {
	if stub.IsFlagDefinedCalled != nil {
		return stub.IsFlagDefinedCalled(flag)
	}
	return false
}

// IsFlagEnabledInCurrentEpoch -
func (stub *EnableEpochsHandlerStub) IsFlagEnabledInCurrentEpoch(flag core.EnableEpochFlag) bool {
	if stub.IsFlagEnabledInCurrentEpochCalled != nil {
		return stub.IsFlagEnabledInCurrentEpochCalled(flag)
	}
	return false
}

// IsFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsFlagEnabledInEpoch(flag core.EnableEpochFlag, epoch uint32) bool {
	if stub.IsFlagEnabledInEpochCalled != nil {
		return stub.IsFlagEnabledInEpochCalled(flag, epoch)
	}
	return false
}

// GetCurrentEpoch -
func (stub *EnableEpochsHandlerStub) GetCurrentEpoch() uint32 {
	if stub.GetCurrentEpochCalled != nil {
		return stub.GetCurrentEpochCalled()
	}
	return 0
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

// IsSCDeployFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsSCDeployFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsSCDeployFlagEnabledInEpochCalled != nil {
		return stub.IsSCDeployFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsBuiltInFunctionsFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsBuiltInFunctionsFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsBuiltInFunctionsFlagEnabledInEpochCalled != nil {
		return stub.IsBuiltInFunctionsFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsRelayedTransactionsFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsRelayedTransactionsFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsRelayedTransactionsFlagEnabledInEpochCalled != nil {
		return stub.IsRelayedTransactionsFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsPenalizedTooMuchGasFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsPenalizedTooMuchGasFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsPenalizedTooMuchGasFlagEnabledInEpochCalled != nil {
		return stub.IsPenalizedTooMuchGasFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsSwitchJailWaitingFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsSwitchJailWaitingFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsSwitchJailWaitingFlagEnabledInEpochCalled != nil {
		return stub.IsSwitchJailWaitingFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsBelowSignedThresholdFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsBelowSignedThresholdFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsBelowSignedThresholdFlagEnabledInEpochCalled != nil {
		return stub.IsBelowSignedThresholdFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsSwitchHysteresisForMinNodesFlagEnabledInSpecificEpochOnly -
func (stub *EnableEpochsHandlerStub) IsSwitchHysteresisForMinNodesFlagEnabledInSpecificEpochOnly(epoch uint32) bool {
	if stub.IsSwitchHysteresisForMinNodesFlagEnabledInSpecificEpochOnlyCalled != nil {
		return stub.IsSwitchHysteresisForMinNodesFlagEnabledInSpecificEpochOnlyCalled(epoch)
	}
	return false
}

// IsTransactionSignedWithTxHashFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsTransactionSignedWithTxHashFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsTransactionSignedWithTxHashFlagEnabledInEpochCalled != nil {
		return stub.IsTransactionSignedWithTxHashFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsMetaProtectionFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsMetaProtectionFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsMetaProtectionFlagEnabledInEpochCalled != nil {
		return stub.IsMetaProtectionFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsAheadOfTimeGasUsageFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsAheadOfTimeGasUsageFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsAheadOfTimeGasUsageFlagEnabledInEpochCalled != nil {
		return stub.IsAheadOfTimeGasUsageFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsGasPriceModifierFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsGasPriceModifierFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsGasPriceModifierFlagEnabledInEpochCalled != nil {
		return stub.IsGasPriceModifierFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsRepairCallbackFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsRepairCallbackFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsRepairCallbackFlagEnabledInEpochCalled != nil {
		return stub.IsRepairCallbackFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsReturnDataToLastTransferFlagEnabledAfterEpoch -
func (stub *EnableEpochsHandlerStub) IsReturnDataToLastTransferFlagEnabledAfterEpoch(epoch uint32) bool {
	if stub.IsReturnDataToLastTransferFlagEnabledAfterEpochCalled != nil {
		return stub.IsReturnDataToLastTransferFlagEnabledAfterEpochCalled(epoch)
	}
	return false
}

// IsSenderInOutTransferFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsSenderInOutTransferFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsSenderInOutTransferFlagEnabledInEpochCalled != nil {
		return stub.IsSenderInOutTransferFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsStakeFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsStakeFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsStakeFlagEnabledInEpochCalled != nil {
		return stub.IsStakeFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsStakingV2FlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsStakingV2FlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsStakingV2FlagEnabledInEpochCalled != nil {
		return stub.IsStakingV2FlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsStakingV2OwnerFlagEnabledInSpecificEpochOnly -
func (stub *EnableEpochsHandlerStub) IsStakingV2OwnerFlagEnabledInSpecificEpochOnly(epoch uint32) bool {
	if stub.IsStakingV2OwnerFlagEnabledInSpecificEpochOnlyCalled != nil {
		return stub.IsStakingV2OwnerFlagEnabledInSpecificEpochOnlyCalled(epoch)
	}
	return false
}

// IsStakingV2FlagEnabledAfterEpoch -
func (stub *EnableEpochsHandlerStub) IsStakingV2FlagEnabledAfterEpoch(epoch uint32) bool {
	if stub.IsStakingV2FlagEnabledAfterEpochCalled != nil {
		return stub.IsStakingV2FlagEnabledAfterEpochCalled(epoch)
	}
	return false
}

// IsDoubleKeyProtectionFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsDoubleKeyProtectionFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsDoubleKeyProtectionFlagEnabledInEpochCalled != nil {
		return stub.IsDoubleKeyProtectionFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsESDTFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsESDTFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsESDTFlagEnabledInEpochCalled != nil {
		return stub.IsESDTFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsESDTFlagEnabledInSpecificEpochOnly -
func (stub *EnableEpochsHandlerStub) IsESDTFlagEnabledInSpecificEpochOnly(epoch uint32) bool {
	if stub.IsESDTFlagEnabledInSpecificEpochOnlyCalled != nil {
		return stub.IsESDTFlagEnabledInSpecificEpochOnlyCalled(epoch)
	}
	return false
}

// IsGovernanceFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsGovernanceFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsGovernanceFlagEnabledInEpochCalled != nil {
		return stub.IsGovernanceFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsGovernanceFlagEnabledInSpecificEpochOnly -
func (stub *EnableEpochsHandlerStub) IsGovernanceFlagEnabledInSpecificEpochOnly(epoch uint32) bool {
	if stub.IsGovernanceFlagEnabledInSpecificEpochOnlyCalled != nil {
		return stub.IsGovernanceFlagEnabledInSpecificEpochOnlyCalled(epoch)
	}
	return false
}

// IsDelegationManagerFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsDelegationManagerFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsDelegationManagerFlagEnabledInEpochCalled != nil {
		return stub.IsDelegationManagerFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsDelegationSmartContractFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsDelegationSmartContractFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsDelegationSmartContractFlagEnabledInEpochCalled != nil {
		return stub.IsDelegationSmartContractFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsDelegationSmartContractFlagEnabledInSpecificEpochOnly -
func (stub *EnableEpochsHandlerStub) IsDelegationSmartContractFlagEnabledInSpecificEpochOnly(epoch uint32) bool {
	if stub.IsDelegationSmartContractFlagEnabledInSpecificEpochOnlyCalled != nil {
		return stub.IsDelegationSmartContractFlagEnabledInSpecificEpochOnlyCalled(epoch)
	}
	return false
}

// IsCorrectLastUnJailedFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsCorrectLastUnJailedFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsCorrectLastUnJailedFlagEnabledInEpochCalled != nil {
		return stub.IsCorrectLastUnJailedFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsCorrectLastUnJailedFlagEnabledInSpecificEpochOnly -
func (stub *EnableEpochsHandlerStub) IsCorrectLastUnJailedFlagEnabledInSpecificEpochOnly(epoch uint32) bool {
	if stub.IsCorrectLastUnJailedFlagEnabledInSpecificEpochOnlyCalled != nil {
		return stub.IsCorrectLastUnJailedFlagEnabledInSpecificEpochOnlyCalled(epoch)
	}
	return false
}

// IsRelayedTransactionsV2FlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsRelayedTransactionsV2FlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsRelayedTransactionsV2FlagEnabledInEpochCalled != nil {
		return stub.IsRelayedTransactionsV2FlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsUnBondTokensV2FlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsUnBondTokensV2FlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsUnBondTokensV2FlagEnabledInEpochCalled != nil {
		return stub.IsUnBondTokensV2FlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsSaveJailedAlwaysFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsSaveJailedAlwaysFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsSaveJailedAlwaysFlagEnabledInEpochCalled != nil {
		return stub.IsSaveJailedAlwaysFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsReDelegateBelowMinCheckFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsReDelegateBelowMinCheckFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsReDelegateBelowMinCheckFlagEnabledInEpochCalled != nil {
		return stub.IsReDelegateBelowMinCheckFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsValidatorToDelegationFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsValidatorToDelegationFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsValidatorToDelegationFlagEnabledInEpochCalled != nil {
		return stub.IsValidatorToDelegationFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsIncrementSCRNonceInMultiTransferFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsIncrementSCRNonceInMultiTransferFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsIncrementSCRNonceInMultiTransferFlagEnabledInEpochCalled != nil {
		return stub.IsIncrementSCRNonceInMultiTransferFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsESDTMultiTransferFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsESDTMultiTransferFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsESDTMultiTransferFlagEnabledInEpochCalled != nil {
		return stub.IsESDTMultiTransferFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsGlobalMintBurnFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsGlobalMintBurnFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsGlobalMintBurnFlagEnabledInEpochCalled != nil {
		return stub.IsGlobalMintBurnFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsESDTTransferRoleFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsESDTTransferRoleFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsESDTTransferRoleFlagEnabledInEpochCalled != nil {
		return stub.IsESDTTransferRoleFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsBuiltInFunctionOnMetaFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsBuiltInFunctionOnMetaFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsBuiltInFunctionOnMetaFlagEnabledInEpochCalled != nil {
		return stub.IsBuiltInFunctionOnMetaFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsComputeRewardCheckpointFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsComputeRewardCheckpointFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsComputeRewardCheckpointFlagEnabledInEpochCalled != nil {
		return stub.IsComputeRewardCheckpointFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsSCRSizeInvariantCheckFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsSCRSizeInvariantCheckFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsSCRSizeInvariantCheckFlagEnabledInEpochCalled != nil {
		return stub.IsSCRSizeInvariantCheckFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsBackwardCompSaveKeyValueFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsBackwardCompSaveKeyValueFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsBackwardCompSaveKeyValueFlagEnabledInEpochCalled != nil {
		return stub.IsBackwardCompSaveKeyValueFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsESDTNFTCreateOnMultiShardFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsESDTNFTCreateOnMultiShardFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsESDTNFTCreateOnMultiShardFlagEnabledInEpochCalled != nil {
		return stub.IsESDTNFTCreateOnMultiShardFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsMetaESDTSetFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsMetaESDTSetFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsMetaESDTSetFlagEnabledInEpochCalled != nil {
		return stub.IsMetaESDTSetFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsAddTokensToDelegationFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsAddTokensToDelegationFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsAddTokensToDelegationFlagEnabledInEpochCalled != nil {
		return stub.IsAddTokensToDelegationFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsMultiESDTTransferFixOnCallBackFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsMultiESDTTransferFixOnCallBackFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsMultiESDTTransferFixOnCallBackFlagEnabledInEpochCalled != nil {
		return stub.IsMultiESDTTransferFixOnCallBackFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledInEpochCalled != nil {
		return stub.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsCorrectFirstQueuedFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsCorrectFirstQueuedFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsCorrectFirstQueuedFlagEnabledInEpochCalled != nil {
		return stub.IsCorrectFirstQueuedFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsDeleteDelegatorAfterClaimRewardsFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsDeleteDelegatorAfterClaimRewardsFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsDeleteDelegatorAfterClaimRewardsFlagEnabledInEpochCalled != nil {
		return stub.IsDeleteDelegatorAfterClaimRewardsFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsRemoveNonUpdatedStorageFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsRemoveNonUpdatedStorageFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsRemoveNonUpdatedStorageFlagEnabledInEpochCalled != nil {
		return stub.IsRemoveNonUpdatedStorageFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsOptimizeNFTStoreFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsOptimizeNFTStoreFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsOptimizeNFTStoreFlagEnabledInEpochCalled != nil {
		return stub.IsOptimizeNFTStoreFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsCreateNFTThroughExecByCallerFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsCreateNFTThroughExecByCallerFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsCreateNFTThroughExecByCallerFlagEnabledInEpochCalled != nil {
		return stub.IsCreateNFTThroughExecByCallerFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsStopDecreasingValidatorRatingWhenStuckFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsStopDecreasingValidatorRatingWhenStuckFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsStopDecreasingValidatorRatingWhenStuckFlagEnabledInEpochCalled != nil {
		return stub.IsStopDecreasingValidatorRatingWhenStuckFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsFrontRunningProtectionFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsFrontRunningProtectionFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsFrontRunningProtectionFlagEnabledInEpochCalled != nil {
		return stub.IsFrontRunningProtectionFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsPayableBySCFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsPayableBySCFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsPayableBySCFlagEnabledInEpochCalled != nil {
		return stub.IsPayableBySCFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsCleanUpInformativeSCRsFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsCleanUpInformativeSCRsFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsCleanUpInformativeSCRsFlagEnabledInEpochCalled != nil {
		return stub.IsCleanUpInformativeSCRsFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsStorageAPICostOptimizationFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsStorageAPICostOptimizationFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsStorageAPICostOptimizationFlagEnabledInEpochCalled != nil {
		return stub.IsStorageAPICostOptimizationFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsESDTRegisterAndSetAllRolesFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsESDTRegisterAndSetAllRolesFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsESDTRegisterAndSetAllRolesFlagEnabledInEpochCalled != nil {
		return stub.IsESDTRegisterAndSetAllRolesFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsScheduledMiniBlocksFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsScheduledMiniBlocksFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsScheduledMiniBlocksFlagEnabledInEpochCalled != nil {
		return stub.IsScheduledMiniBlocksFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledInEpochCalled != nil {
		return stub.IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsDoNotReturnOldBlockInBlockchainHookFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsDoNotReturnOldBlockInBlockchainHookFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsDoNotReturnOldBlockInBlockchainHookFlagEnabledInEpochCalled != nil {
		return stub.IsDoNotReturnOldBlockInBlockchainHookFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsAddFailedRelayedTxToInvalidMBsFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsAddFailedRelayedTxToInvalidMBsFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsAddFailedRelayedTxToInvalidMBsFlagEnabledInEpochCalled != nil {
		return stub.IsAddFailedRelayedTxToInvalidMBsFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsSCRSizeInvariantOnBuiltInResultFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsSCRSizeInvariantOnBuiltInResultFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsSCRSizeInvariantOnBuiltInResultFlagEnabledInEpochCalled != nil {
		return stub.IsSCRSizeInvariantOnBuiltInResultFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsCheckCorrectTokenIDForTransferRoleFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsCheckCorrectTokenIDForTransferRoleFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsCheckCorrectTokenIDForTransferRoleFlagEnabledInEpochCalled != nil {
		return stub.IsCheckCorrectTokenIDForTransferRoleFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsFailExecutionOnEveryAPIErrorFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsFailExecutionOnEveryAPIErrorFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsFailExecutionOnEveryAPIErrorFlagEnabledInEpochCalled != nil {
		return stub.IsFailExecutionOnEveryAPIErrorFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsMiniBlockPartialExecutionFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsMiniBlockPartialExecutionFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsMiniBlockPartialExecutionFlagEnabledInEpochCalled != nil {
		return stub.IsMiniBlockPartialExecutionFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsManagedCryptoAPIsFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsManagedCryptoAPIsFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsManagedCryptoAPIsFlagEnabledInEpochCalled != nil {
		return stub.IsManagedCryptoAPIsFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsESDTMetadataContinuousCleanupFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsESDTMetadataContinuousCleanupFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsESDTMetadataContinuousCleanupFlagEnabledInEpochCalled != nil {
		return stub.IsESDTMetadataContinuousCleanupFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsDisableExecByCallerFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsDisableExecByCallerFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsDisableExecByCallerFlagEnabledInEpochCalled != nil {
		return stub.IsDisableExecByCallerFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsRefactorContextFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsRefactorContextFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsRefactorContextFlagEnabledInEpochCalled != nil {
		return stub.IsRefactorContextFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsCheckFunctionArgumentFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsCheckFunctionArgumentFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsCheckFunctionArgumentFlagEnabledInEpochCalled != nil {
		return stub.IsCheckFunctionArgumentFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsCheckExecuteOnReadOnlyFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsCheckExecuteOnReadOnlyFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsCheckExecuteOnReadOnlyFlagEnabledInEpochCalled != nil {
		return stub.IsCheckExecuteOnReadOnlyFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsSetSenderInEeiOutputTransferFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsSetSenderInEeiOutputTransferFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsSetSenderInEeiOutputTransferFlagEnabledInEpochCalled != nil {
		return stub.IsSetSenderInEeiOutputTransferFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsFixAsyncCallbackCheckFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsFixAsyncCallbackCheckFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsFixAsyncCallbackCheckFlagEnabledInEpochCalled != nil {
		return stub.IsFixAsyncCallbackCheckFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsSaveToSystemAccountFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsSaveToSystemAccountFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsSaveToSystemAccountFlagEnabledInEpochCalled != nil {
		return stub.IsSaveToSystemAccountFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsCheckFrozenCollectionFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsCheckFrozenCollectionFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsCheckFrozenCollectionFlagEnabledInEpochCalled != nil {
		return stub.IsCheckFrozenCollectionFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsSendAlwaysFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsSendAlwaysFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsSendAlwaysFlagEnabledInEpochCalled != nil {
		return stub.IsSendAlwaysFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsValueLengthCheckFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsValueLengthCheckFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsValueLengthCheckFlagEnabledInEpochCalled != nil {
		return stub.IsValueLengthCheckFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsCheckTransferFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsCheckTransferFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsCheckTransferFlagEnabledInEpochCalled != nil {
		return stub.IsCheckTransferFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsTransferToMetaFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsTransferToMetaFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsTransferToMetaFlagEnabledInEpochCalled != nil {
		return stub.IsTransferToMetaFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsESDTNFTImprovementV1FlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsESDTNFTImprovementV1FlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsESDTNFTImprovementV1FlagEnabledInEpochCalled != nil {
		return stub.IsESDTNFTImprovementV1FlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsChangeDelegationOwnerFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsChangeDelegationOwnerFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsChangeDelegationOwnerFlagEnabledInEpochCalled != nil {
		return stub.IsChangeDelegationOwnerFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsRefactorPeersMiniBlocksFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsRefactorPeersMiniBlocksFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsRefactorPeersMiniBlocksFlagEnabledInEpochCalled != nil {
		return stub.IsRefactorPeersMiniBlocksFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsSCProcessorV2FlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsSCProcessorV2FlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsSCProcessorV2FlagEnabledInEpochCalled != nil {
		return stub.IsSCProcessorV2FlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsFixAsyncCallBackArgsListFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsFixAsyncCallBackArgsListFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsFixAsyncCallBackArgsListFlagEnabledInEpochCalled != nil {
		return stub.IsFixAsyncCallBackArgsListFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsFixOldTokenLiquidityEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsFixOldTokenLiquidityEnabledInEpoch(epoch uint32) bool {
	if stub.IsFixOldTokenLiquidityEnabledInEpochCalled != nil {
		return stub.IsFixOldTokenLiquidityEnabledInEpochCalled(epoch)
	}
	return false
}

// IsRuntimeMemStoreLimitEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsRuntimeMemStoreLimitEnabledInEpoch(epoch uint32) bool {
	if stub.IsRuntimeMemStoreLimitEnabledInEpochCalled != nil {
		return stub.IsRuntimeMemStoreLimitEnabledInEpochCalled(epoch)
	}
	return false
}

// IsRuntimeCodeSizeFixEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsRuntimeCodeSizeFixEnabledInEpoch(epoch uint32) bool {
	if stub.IsRuntimeCodeSizeFixEnabledInEpochCalled != nil {
		return stub.IsRuntimeCodeSizeFixEnabledInEpochCalled(epoch)
	}
	return false
}

// IsMaxBlockchainHookCountersFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsMaxBlockchainHookCountersFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsMaxBlockchainHookCountersFlagEnabledInEpochCalled != nil {
		return stub.IsMaxBlockchainHookCountersFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsWipeSingleNFTLiquidityDecreaseEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsWipeSingleNFTLiquidityDecreaseEnabledInEpoch(epoch uint32) bool {
	if stub.IsWipeSingleNFTLiquidityDecreaseEnabledInEpochCalled != nil {
		return stub.IsWipeSingleNFTLiquidityDecreaseEnabledInEpochCalled(epoch)
	}
	return false
}

// IsAlwaysSaveTokenMetaDataEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsAlwaysSaveTokenMetaDataEnabledInEpoch(epoch uint32) bool {
	if stub.IsAlwaysSaveTokenMetaDataEnabledInEpochCalled != nil {
		return stub.IsAlwaysSaveTokenMetaDataEnabledInEpochCalled(epoch)
	}
	return false
}

// IsSetGuardianEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsSetGuardianEnabledInEpoch(epoch uint32) bool {
	if stub.IsSetGuardianEnabledInEpochCalled != nil {
		return stub.IsSetGuardianEnabledInEpochCalled(epoch)
	}
	return false
}

// IsRelayedNonceFixEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsRelayedNonceFixEnabledInEpoch(epoch uint32) bool {
	if stub.IsRelayedNonceFixEnabledInEpochCalled != nil {
		return stub.IsRelayedNonceFixEnabledInEpochCalled(epoch)
	}
	return false
}

// IsConsistentTokensValuesLengthCheckEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsConsistentTokensValuesLengthCheckEnabledInEpoch(epoch uint32) bool {
	if stub.IsConsistentTokensValuesLengthCheckEnabledInEpochCalled != nil {
		return stub.IsConsistentTokensValuesLengthCheckEnabledInEpochCalled(epoch)
	}
	return false
}

// IsKeepExecOrderOnCreatedSCRsEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsKeepExecOrderOnCreatedSCRsEnabledInEpoch(epoch uint32) bool {
	if stub.IsKeepExecOrderOnCreatedSCRsEnabledInEpochCalled != nil {
		return stub.IsKeepExecOrderOnCreatedSCRsEnabledInEpochCalled(epoch)
	}
	return false
}

// IsMultiClaimOnDelegationEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsMultiClaimOnDelegationEnabledInEpoch(epoch uint32) bool {
	if stub.IsMultiClaimOnDelegationEnabledInEpochCalled != nil {
		return stub.IsMultiClaimOnDelegationEnabledInEpochCalled(epoch)
	}
	return false
}

// IsChangeUsernameEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsChangeUsernameEnabledInEpoch(epoch uint32) bool {
	if stub.IsChangeUsernameEnabledInEpochCalled != nil {
		return stub.IsChangeUsernameEnabledInEpochCalled(epoch)
	}
	return false
}

// IsAutoBalanceDataTriesEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsAutoBalanceDataTriesEnabledInEpoch(epoch uint32) bool {
	if stub.IsAutoBalanceDataTriesEnabledInEpochCalled != nil {
		return stub.IsAutoBalanceDataTriesEnabledInEpochCalled(epoch)
	}
	return false
}

// FixDelegationChangeOwnerOnAccountEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) FixDelegationChangeOwnerOnAccountEnabledInEpoch(epoch uint32) bool {
	if stub.FixDelegationChangeOwnerOnAccountEnabledInEpochCalled != nil {
		return stub.FixDelegationChangeOwnerOnAccountEnabledInEpochCalled(epoch)
	}
	return false
}

// IsFixOOGReturnCodeFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsFixOOGReturnCodeFlagEnabledInEpoch(epoch uint32) bool {
	if stub.IsFixOOGReturnCodeFlagEnabledInEpochCalled != nil {
		return stub.IsFixOOGReturnCodeFlagEnabledInEpochCalled(epoch)
	}
	return false
}

// IsDeterministicSortOnValidatorsInfoFixEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsDeterministicSortOnValidatorsInfoFixEnabledInEpoch(epoch uint32) bool {
	if stub.IsDeterministicSortOnValidatorsInfoFixEnabledInEpochCalled != nil {
		return stub.IsDeterministicSortOnValidatorsInfoFixEnabledInEpochCalled(epoch)
	}
	return false
}

// IsInterfaceNil -
func (stub *EnableEpochsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
