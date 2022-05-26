package enableEpochs

import (
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
)

var log = logger.GetOrCreate("enableEpochsHandler")

type enableEpochsHandler struct {
	*flagsHolder
	enableEpochsConfig config.EnableEpochs
}

// NewEnableEpochsHandler creates a new instance of enableEpochsHandler
func NewEnableEpochsHandler(enableEpochsConfig config.EnableEpochs, epochNotifier process.EpochNotifier) (*enableEpochsHandler, error) {
	if check.IfNil(epochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	handler := &enableEpochsHandler{
		flagsHolder:        &flagsHolder{},
		enableEpochsConfig: enableEpochsConfig,
	}

	epochNotifier.RegisterNotifyHandler(handler)

	return handler, nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (handler *enableEpochsHandler) EpochConfirmed(epoch uint32, _ uint64) {
	cfg := handler.enableEpochsConfig

	handler.setFlagValue(epoch >= cfg.SCDeployEnableEpoch, &handler.scDeployFlag, "scDeployFlag")
	handler.setFlagValue(epoch >= cfg.BuiltInFunctionsEnableEpoch, &handler.builtInFunctionsFlag, "builtInFunctionsFlag")
	handler.setFlagValue(epoch >= cfg.RelayedTransactionsEnableEpoch, &handler.relayedTransactionsFlag, "relayedTransactionsFlag")
	handler.setFlagValue(epoch >= cfg.PenalizedTooMuchGasEnableEpoch, &handler.penalizedTooMuchGasFlag, "penalizedTooMuchGasFlag")
	handler.setFlagValue(epoch >= cfg.SwitchJailWaitingEnableEpoch, &handler.switchJailWaitingFlag, "switchJailWaitingFlag")
	handler.setFlagValue(epoch >= cfg.BelowSignedThresholdEnableEpoch, &handler.belowSignedThresholdFlag, "belowSignedThresholdFlag")
	handler.setFlagValue(epoch >= cfg.SwitchHysteresisForMinNodesEnableEpoch, &handler.switchHysteresisForMinNodesFlag, "switchHysteresisForMinNodesFlag")
	handler.setFlagValue(epoch >= cfg.TransactionSignedWithTxHashEnableEpoch, &handler.transactionSignedWithTxHashFlag, "transactionSignedWithTxHashFlag")
	handler.setFlagValue(epoch >= cfg.MetaProtectionEnableEpoch, &handler.metaProtectionFlag, "metaProtectionFlag")
	handler.setFlagValue(epoch >= cfg.AheadOfTimeGasUsageEnableEpoch, &handler.aheadOfTimeGasUsageFlag, "aheadOfTimeGasUsageFlag")
	handler.setFlagValue(epoch >= cfg.GasPriceModifierEnableEpoch, &handler.gasPriceModifierFlag, "gasPriceModifierFlag")
	handler.setFlagValue(epoch >= cfg.RepairCallbackEnableEpoch, &handler.repairCallbackFlag, "repairCallbackFlag")
	handler.setFlagValue(epoch >= cfg.BalanceWaitingListsEnableEpoch, &handler.balanceWaitingListsFlag, "balanceWaitingListsFlag")
	handler.setFlagValue(epoch > cfg.ReturnDataToLastTransferEnableEpoch, &handler.returnDataToLastTransferFlag, "returnDataToLastTransferFlag")
	handler.setFlagValue(epoch >= cfg.SenderInOutTransferEnableEpoch, &handler.senderInOutTransferFlag, "senderInOutTransferFlag")
	handler.setFlagValue(epoch >= cfg.StakeEnableEpoch, &handler.stakeFlag, "stakeFlag")
	handler.setFlagValue(epoch >= cfg.StakingV2EnableEpoch, &handler.stakingV2Flag, "stakingV2Flag")
	handler.setFlagValue(epoch == cfg.StakingV2EnableEpoch, &handler.stakingV2OwnerFlag, "stakingV2OwnerFlag")
	handler.setFlagValue(epoch > cfg.StakingV2EnableEpoch, &handler.stakingV2DelegationFlag, "stakingV2DelegationFlag")
	handler.setFlagValue(epoch >= cfg.DoubleKeyProtectionEnableEpoch, &handler.doubleKeyProtectionFlag, "doubleKeyProtectionFlag")
	handler.setFlagValue(epoch >= cfg.ESDTEnableEpoch, &handler.esdtFlag, "esdtFlag")
	handler.setFlagValue(epoch == cfg.ESDTEnableEpoch, &handler.esdtCurrentEpochFlag, "esdtCurrentEpochFlag")
	handler.setFlagValue(epoch >= cfg.GovernanceEnableEpoch, &handler.governanceFlag, "governanceFlag")
	handler.setFlagValue(epoch == cfg.GovernanceEnableEpoch, &handler.governanceCurrentEpochFlag, "governanceCurrentEpochFlag")
	handler.setFlagValue(epoch >= cfg.DelegationManagerEnableEpoch, &handler.delegationManagerFlag, "delegationManagerFlag")
	handler.setFlagValue(epoch >= cfg.DelegationSmartContractEnableEpoch, &handler.delegationSmartContractFlag, "delegationSmartContractFlag")
	handler.setFlagValue(epoch >= cfg.CorrectLastUnjailedEnableEpoch, &handler.correctLastUnjailedFlag, "correctLastUnjailedFlag")
	handler.setFlagValue(epoch == cfg.CorrectLastUnjailedEnableEpoch, &handler.correctLastUnjailedCurrentEpochFlag, "correctLastUnjailedCurrentEpochFlag")
	handler.setFlagValue(epoch >= cfg.RelayedTransactionsV2EnableEpoch, &handler.relayedTransactionsV2Flag, "relayedTransactionsV2Flag")
	handler.setFlagValue(epoch >= cfg.UnbondTokensV2EnableEpoch, &handler.unbondTokensV2Flag, "unbondTokensV2Flag")
	handler.setFlagValue(epoch >= cfg.SaveJailedAlwaysEnableEpoch, &handler.saveJailedAlwaysFlag, "saveJailedAlwaysFlag")
	handler.setFlagValue(epoch >= cfg.ReDelegateBelowMinCheckEnableEpoch, &handler.reDelegateBelowMinCheckFlag, "reDelegateBelowMinCheckFlag")
	handler.setFlagValue(epoch >= cfg.ValidatorToDelegationEnableEpoch, &handler.validatorToDelegationFlag, "validatorToDelegationFlag")
	handler.setFlagValue(epoch >= cfg.WaitingListFixEnableEpoch, &handler.waitingListFixFlag, "waitingListFixFlag")
	handler.setFlagValue(epoch >= cfg.IncrementSCRNonceInMultiTransferEnableEpoch, &handler.incrementSCRNonceInMultiTransferFlag, "incrementSCRNonceInMultiTransferFlag")
	handler.setFlagValue(epoch >= cfg.ESDTMultiTransferEnableEpoch, &handler.esdtMultiTransferFlag, "esdtMultiTransferFlag")
	handler.setFlagValue(epoch < cfg.GlobalMintBurnDisableEpoch, &handler.globalMintBurnDisableFlag, "globalMintBurnDisableFlag")
	handler.setFlagValue(epoch >= cfg.ESDTTransferRoleEnableEpoch, &handler.esdtTransferRoleFlag, "esdtTransferRoleFlag")
	handler.setFlagValue(epoch >= cfg.BuiltInFunctionOnMetaEnableEpoch, &handler.builtInFunctionOnMetaFlag, "builtInFunctionOnMetaFlag")
	handler.setFlagValue(epoch >= cfg.ComputeRewardCheckpointEnableEpoch, &handler.computeRewardCheckpointFlag, "computeRewardCheckpointFlag")
	handler.setFlagValue(epoch >= cfg.SCRSizeInvariantCheckEnableEpoch, &handler.scrSizeInvariantCheckFlag, "scrSizeInvariantCheckFlag")
	handler.setFlagValue(epoch < cfg.BackwardCompSaveKeyValueEnableEpoch, &handler.backwardCompSaveKeyValueFlag, "backwardCompSaveKeyValueFlag")
	handler.setFlagValue(epoch >= cfg.ESDTNFTCreateOnMultiShardEnableEpoch, &handler.esdtNFTCreateOnMultiShardFlag, "esdtNFTCreateOnMultiShardFlag")
	handler.setFlagValue(epoch >= cfg.MetaESDTSetEnableEpoch, &handler.metaESDTSetFlag, "metaESDTSetFlag")
	handler.setFlagValue(epoch >= cfg.AddTokensToDelegationEnableEpoch, &handler.addTokensToDelegationFlag, "addTokensToDelegationFlag")
	handler.setFlagValue(epoch >= cfg.MultiESDTTransferFixOnCallBackOnEnableEpoch, &handler.multiESDTTransferFixOnCallBackFlag, "multiESDTTransferFixOnCallBackFlag")
	handler.setFlagValue(epoch >= cfg.OptimizeGasUsedInCrossMiniBlocksEnableEpoch, &handler.optimizeGasUsedInCrossMiniBlocksFlag, "optimizeGasUsedInCrossMiniBlocksFlag")
	handler.setFlagValue(epoch >= cfg.CorrectFirstQueuedEpoch, &handler.correctFirstQueuedFlag, "correctFirstQueuedFlag")
	handler.setFlagValue(epoch >= cfg.DeleteDelegatorAfterClaimRewardsEnableEpoch, &handler.deleteDelegatorAfterClaimRewardsFlag, "deleteDelegatorAfterClaimRewardsFlag")
	handler.setFlagValue(epoch >= cfg.FixOOGReturnCodeEnableEpoch, &handler.fixOOGReturnCodeFlag, "fixOOGReturnCodeFlag")
	handler.setFlagValue(epoch >= cfg.RemoveNonUpdatedStorageEnableEpoch, &handler.removeNonUpdatedStorageFlag, "removeNonUpdatedStorageFlag")
	handler.setFlagValue(epoch >= cfg.OptimizeNFTStoreEnableEpoch, &handler.optimizeNFTStoreFlag, "optimizeNFTStoreFlag")
	handler.setFlagValue(epoch >= cfg.CreateNFTThroughExecByCallerEnableEpoch, &handler.createNFTThroughExecByCallerFlag, "createNFTThroughExecByCallerFlag")
	handler.setFlagValue(epoch >= cfg.StopDecreasingValidatorRatingWhenStuckEnableEpoch, &handler.stopDecreasingValidatorRatingWhenStuckFlag, "stopDecreasingValidatorRatingWhenStuckFlag")
	handler.setFlagValue(epoch >= cfg.FrontRunningProtectionEnableEpoch, &handler.frontRunningProtectionFlag, "frontRunningProtectionFlag")
	handler.setFlagValue(epoch >= cfg.IsPayableBySCEnableEpoch, &handler.isPayableBySCFlag, "isPayableBySCFlag")
	handler.setFlagValue(epoch >= cfg.CleanUpInformativeSCRsEnableEpoch, &handler.cleanUpInformativeSCRsFlag, "cleanUpInformativeSCRsFlag")
	handler.setFlagValue(epoch >= cfg.StorageAPICostOptimizationEnableEpoch, &handler.storageAPICostOptimizationFlag, "storageAPICostOptimizationFlag")
	handler.setFlagValue(epoch >= cfg.ESDTRegisterAndSetAllRolesEnableEpoch, &handler.esdtRegisterAndSetAllRolesFlag, "esdtRegisterAndSetAllRolesFlag")
	handler.setFlagValue(epoch >= cfg.ScheduledMiniBlocksEnableEpoch, &handler.scheduledMiniBlocksFlag, "scheduledMiniBlocksFlag")
	handler.setFlagValue(epoch >= cfg.CorrectJailedNotUnstakedEmptyQueueEpoch, &handler.correctJailedNotUnstakedEmptyQueueFlag, "correctJailedNotUnstakedEmptyQueueFlag")
	handler.setFlagValue(epoch >= cfg.DoNotReturnOldBlockInBlockchainHookEnableEpoch, &handler.doNotReturnOldBlockInBlockchainHookFlag, "doNotReturnOldBlockInBlockchainHookFlag")
	handler.setFlagValue(epoch >= cfg.SCRSizeInvariantOnBuiltInResultEnableEpoch, &handler.scrSizeInvariantOnBuiltInResultFlag, "scrSizeInvariantOnBuiltInResultFlag")
	handler.setFlagValue(epoch >= cfg.CheckCorrectTokenIDForTransferRoleEnableEpoch, &handler.checkCorrectTokenIDForTransferRoleFlag, "checkCorrectTokenIDForTransferRoleFlag")
	handler.setFlagValue(epoch >= cfg.FailExecutionOnEveryAPIErrorEnableEpoch, &handler.failExecutionOnEveryAPIErrorFlag, "failExecutionOnEveryAPIErrorFlag")
	handler.setFlagValue(epoch >= cfg.HeartbeatDisableEpoch, &handler.heartbeatDisableFlag, "heartbeatDisableFlag")
	handler.setFlagValue(epoch >= cfg.MiniBlockPartialExecutionEnableEpoch, &handler.isMiniBlockPartialExecutionFlag, "isMiniBlockPartialExecutionFlag")
}

func (handler *enableEpochsHandler) setFlagValue(value bool, flag *atomic.Flag, flagName string) {
	flag.SetValue(value)
	log.Debug("EpochConfirmed", "flag", flagName, "enabled", flag.IsSet())
}

// BlockGasAndFeesReCheckEnableEpoch returns the epoch for block gas and fees recheck
func (handler *enableEpochsHandler) BlockGasAndFeesReCheckEnableEpoch() uint32 {
	return handler.enableEpochsConfig.BlockGasAndFeesReCheckEnableEpoch
}

// IsSCDeployFlagEnabled returns true if scDeployFlag is enabled
func (handler *enableEpochsHandler) IsSCDeployFlagEnabled() bool {
	return handler.scDeployFlag.IsSet()
}

// IsBuiltInFunctionsFlagEnabled returns true if builtInFunctionsFlag is enabled
func (handler *enableEpochsHandler) IsBuiltInFunctionsFlagEnabled() bool {
	return handler.builtInFunctionsFlag.IsSet()
}

// IsRelayedTransactionsFlagEnabled returns true if relayedTransactionsFlag is enabled
func (handler *enableEpochsHandler) IsRelayedTransactionsFlagEnabled() bool {
	return handler.relayedTransactionsFlag.IsSet()
}

// IsPenalizedTooMuchGasFlagEnabled returns true if penalizedTooMuchGasFlag is enabled
func (handler *enableEpochsHandler) IsPenalizedTooMuchGasFlagEnabled() bool {
	return handler.penalizedTooMuchGasFlag.IsSet()
}

// IsSwitchJailWaitingFlagEnabled returns true if switchJailWaitingFlag is enabled
func (handler *enableEpochsHandler) IsSwitchJailWaitingFlagEnabled() bool {
	return handler.switchJailWaitingFlag.IsSet()
}

// IsBelowSignedThresholdFlagEnabled returns true if belowSignedThresholdFlag is enabled
func (handler *enableEpochsHandler) IsBelowSignedThresholdFlagEnabled() bool {
	return handler.belowSignedThresholdFlag.IsSet()
}

// IsSwitchHysteresisForMinNodesFlagEnabled returns true if switchHysteresisForMinNodesFlag is enabled
func (handler *enableEpochsHandler) IsSwitchHysteresisForMinNodesFlagEnabled() bool {
	return handler.switchHysteresisForMinNodesFlag.IsSet()
}

// IsTransactionSignedWithTxHashFlagEnabled returns true if transactionSignedWithTxHashFlag is enabled
func (handler *enableEpochsHandler) IsTransactionSignedWithTxHashFlagEnabled() bool {
	return handler.transactionSignedWithTxHashFlag.IsSet()
}

// IsMetaProtectionFlagEnabled returns true if metaProtectionFlag is enabled
func (handler *enableEpochsHandler) IsMetaProtectionFlagEnabled() bool {
	return handler.metaProtectionFlag.IsSet()
}

// IsAheadOfTimeGasUsageFlagEnabled returns true if aheadOfTimeGasUsageFlag is enabled
func (handler *enableEpochsHandler) IsAheadOfTimeGasUsageFlagEnabled() bool {
	return handler.aheadOfTimeGasUsageFlag.IsSet()
}

// IsGasPriceModifierFlagEnabled returns true if gasPriceModifierFlag is enabled
func (handler *enableEpochsHandler) IsGasPriceModifierFlagEnabled() bool {
	return handler.gasPriceModifierFlag.IsSet()
}

// IsRepairCallbackFlagEnabled returns true if repairCallbackFlag is enabled
func (handler *enableEpochsHandler) IsRepairCallbackFlagEnabled() bool {
	return handler.repairCallbackFlag.IsSet()
}

// IsBalanceWaitingListsFlagEnabled returns true if balanceWaitingListsFlag is enabled
func (handler *enableEpochsHandler) IsBalanceWaitingListsFlagEnabled() bool {
	return handler.balanceWaitingListsFlag.IsSet()
}

// IsReturnDataToLastTransferFlagEnabled returns true if returnDataToLastTransferFlag is enabled
func (handler *enableEpochsHandler) IsReturnDataToLastTransferFlagEnabled() bool {
	return handler.returnDataToLastTransferFlag.IsSet()
}

// IsSenderInOutTransferFlagEnabled returns true if senderInOutTransferFlag is enabled
func (handler *enableEpochsHandler) IsSenderInOutTransferFlagEnabled() bool {
	return handler.senderInOutTransferFlag.IsSet()
}

// IsStakeFlagEnabled returns true if stakeFlag is enabled
func (handler *enableEpochsHandler) IsStakeFlagEnabled() bool {
	return handler.stakeFlag.IsSet()
}

// IsStakingV2FlagEnabled returns true if stakingV2Flag is enabled
func (handler *enableEpochsHandler) IsStakingV2FlagEnabled() bool {
	return handler.stakingV2Flag.IsSet()
}

// IsStakingV2OwnerFlagEnabled returns true if stakingV2OwnerFlag is enabled
func (handler *enableEpochsHandler) IsStakingV2OwnerFlagEnabled() bool {
	return handler.stakingV2OwnerFlag.IsSet()
}

// IsStakingV2DelegationFlagEnabled returns true if stakingV2DelegationFlag is enabled
func (handler *enableEpochsHandler) IsStakingV2DelegationFlagEnabled() bool {
	return handler.stakingV2DelegationFlag.IsSet()
}

// IsDoubleKeyProtectionFlagEnabled returns true if doubleKeyProtectionFlag is enabled
func (handler *enableEpochsHandler) IsDoubleKeyProtectionFlagEnabled() bool {
	return handler.doubleKeyProtectionFlag.IsSet()
}

// IsESDTFlagEnabled returns true if esdtFlag is enabled
func (handler *enableEpochsHandler) IsESDTFlagEnabled() bool {
	return handler.esdtFlag.IsSet()
}

// IsESDTFlagEnabledForCurrentEpoch returns true if esdtCurrentEpochFlag is enabled
func (handler *enableEpochsHandler) IsESDTFlagEnabledForCurrentEpoch() bool {
	return handler.esdtCurrentEpochFlag.IsSet()
}

// IsGovernanceFlagEnabled returns true if governanceFlag is enabled
func (handler *enableEpochsHandler) IsGovernanceFlagEnabled() bool {
	return handler.governanceFlag.IsSet()
}

// IsGovernanceFlagEnabledForCurrentEpoch returns true if governanceCurrentEpochFlag is enabled
func (handler *enableEpochsHandler) IsGovernanceFlagEnabledForCurrentEpoch() bool {
	return handler.governanceCurrentEpochFlag.IsSet()
}

// IsDelegationManagerFlagEnabled returns true if delegationManagerFlag is enabled
func (handler *enableEpochsHandler) IsDelegationManagerFlagEnabled() bool {
	return handler.delegationManagerFlag.IsSet()
}

// IsDelegationSmartContractFlagEnabled returns true if delegationSmartContractFlag is enabled
func (handler *enableEpochsHandler) IsDelegationSmartContractFlagEnabled() bool {
	return handler.delegationSmartContractFlag.IsSet()
}

// IsCorrectLastUnjailedFlagEnabled returns true if correctLastUnjailedFlag is enabled
func (handler *enableEpochsHandler) IsCorrectLastUnjailedFlagEnabled() bool {
	return handler.correctLastUnjailedFlag.IsSet()
}

// IsCorrectLastUnjailedFlagEnabledForCurrentEpoch returns true if correctLastUnjailedCurrentEpochFlag is enabled
func (handler *enableEpochsHandler) IsCorrectLastUnjailedFlagEnabledForCurrentEpoch() bool {
	return handler.correctLastUnjailedCurrentEpochFlag.IsSet()
}

// IsRelayedTransactionsV2FlagEnabled returns true if relayedTransactionsV2Flag is enabled
func (handler *enableEpochsHandler) IsRelayedTransactionsV2FlagEnabled() bool {
	return handler.relayedTransactionsV2Flag.IsSet()
}

// IsUnbondTokensV2FlagEnabled returns true if unbondTokensV2Flag is enabled
func (handler *enableEpochsHandler) IsUnbondTokensV2FlagEnabled() bool {
	return handler.unbondTokensV2Flag.IsSet()
}

// IsSaveJailedAlwaysFlagEnabled returns true if saveJailedAlwaysFlag is enabled
func (handler *enableEpochsHandler) IsSaveJailedAlwaysFlagEnabled() bool {
	return handler.saveJailedAlwaysFlag.IsSet()
}

// IsReDelegateBelowMinCheckFlagEnabled returns true if reDelegateBelowMinCheckFlag is enabled
func (handler *enableEpochsHandler) IsReDelegateBelowMinCheckFlagEnabled() bool {
	return handler.reDelegateBelowMinCheckFlag.IsSet()
}

// IsValidatorToDelegationFlagEnabled returns true if validatorToDelegationFlag is enabled
func (handler *enableEpochsHandler) IsValidatorToDelegationFlagEnabled() bool {
	return handler.validatorToDelegationFlag.IsSet()
}

// IsWaitingListFixFlagEnabled returns true if waitingListFixFlag is enabled
func (handler *enableEpochsHandler) IsWaitingListFixFlagEnabled() bool {
	return handler.waitingListFixFlag.IsSet()
}

// IsIncrementSCRNonceInMultiTransferFlagEnabled returns true if incrementSCRNonceInMultiTransferFlag is enabled
func (handler *enableEpochsHandler) IsIncrementSCRNonceInMultiTransferFlagEnabled() bool {
	return handler.incrementSCRNonceInMultiTransferFlag.IsSet()
}

// IsESDTMultiTransferFlagEnabled returns true if esdtMultiTransferFlag is enabled
func (handler *enableEpochsHandler) IsESDTMultiTransferFlagEnabled() bool {
	return handler.esdtMultiTransferFlag.IsSet()
}

// IsGlobalMintBurnDisableFlagEnabled returns true if globalMintBurnDisableFlag is enabled
func (handler *enableEpochsHandler) IsGlobalMintBurnDisableFlagEnabled() bool {
	return handler.globalMintBurnDisableFlag.IsSet()
}

// IsESDTTransferRoleFlagEnabled returns true if esdtTransferRoleFlag is enabled
func (handler *enableEpochsHandler) IsESDTTransferRoleFlagEnabled() bool {
	return handler.esdtTransferRoleFlag.IsSet()
}

// IsBuiltInFunctionOnMetaFlagEnabled returns true if builtInFunctionOnMetaFlag is enabled
func (handler *enableEpochsHandler) IsBuiltInFunctionOnMetaFlagEnabled() bool {
	return handler.builtInFunctionOnMetaFlag.IsSet()
}

// IsComputeRewardCheckpointFlagEnabled returns true if computeRewardCheckpointFlag is enabled
func (handler *enableEpochsHandler) IsComputeRewardCheckpointFlagEnabled() bool {
	return handler.computeRewardCheckpointFlag.IsSet()
}

// IsSCRSizeInvariantCheckFlagEnabled returns true if scrSizeInvariantCheckFlag is enabled
func (handler *enableEpochsHandler) IsSCRSizeInvariantCheckFlagEnabled() bool {
	return handler.scrSizeInvariantCheckFlag.IsSet()
}

// IsBackwardCompSaveKeyValueFlagEnabled returns true if backwardCompSaveKeyValueFlag is enabled
func (handler *enableEpochsHandler) IsBackwardCompSaveKeyValueFlagEnabled() bool {
	return handler.backwardCompSaveKeyValueFlag.IsSet()
}

// IsESDTNFTCreateOnMultiShardFlagEnabled returns true if esdtNFTCreateOnMultiShardFlag is enabled
func (handler *enableEpochsHandler) IsESDTNFTCreateOnMultiShardFlagEnabled() bool {
	return handler.esdtNFTCreateOnMultiShardFlag.IsSet()
}

// IsMetaESDTSetFlagEnabled returns true if metaESDTSetFlag is enabled
func (handler *enableEpochsHandler) IsMetaESDTSetFlagEnabled() bool {
	return handler.metaESDTSetFlag.IsSet()
}

// IsAddTokensToDelegationFlagEnabled returns true if addTokensToDelegationFlag is enabled
func (handler *enableEpochsHandler) IsAddTokensToDelegationFlagEnabled() bool {
	return handler.addTokensToDelegationFlag.IsSet()
}

// IsMultiESDTTransferFixOnCallBackFlagEnabled returns true if multiESDTTransferFixOnCallBackFlag is enabled
func (handler *enableEpochsHandler) IsMultiESDTTransferFixOnCallBackFlagEnabled() bool {
	return handler.multiESDTTransferFixOnCallBackFlag.IsSet()
}

// IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled returns true if optimizeGasUsedInCrossMiniBlocksFlag is enabled
func (handler *enableEpochsHandler) IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled() bool {
	return handler.optimizeGasUsedInCrossMiniBlocksFlag.IsSet()
}

// IsCorrectFirstQueuedFlagEnabled returns true if correctFirstQueuedFlag is enabled
func (handler *enableEpochsHandler) IsCorrectFirstQueuedFlagEnabled() bool {
	return handler.correctFirstQueuedFlag.IsSet()
}

// IsDeleteDelegatorAfterClaimRewardsFlagEnabled returns true if deleteDelegatorAfterClaimRewardsFlag is enabled
func (handler *enableEpochsHandler) IsDeleteDelegatorAfterClaimRewardsFlagEnabled() bool {
	return handler.deleteDelegatorAfterClaimRewardsFlag.IsSet()
}

// IsFixOOGReturnCodeFlagEnabled returns true if fixOOGReturnCodeFlag is enabled
func (handler *enableEpochsHandler) IsFixOOGReturnCodeFlagEnabled() bool {
	return handler.fixOOGReturnCodeFlag.IsSet()
}

// IsRemoveNonUpdatedStorageFlagEnabled returns true if removeNonUpdatedStorageFlag is enabled
func (handler *enableEpochsHandler) IsRemoveNonUpdatedStorageFlagEnabled() bool {
	return handler.removeNonUpdatedStorageFlag.IsSet()
}

// IsOptimizeNFTStoreFlagEnabled returns true if removeNonUpdatedStorageFlag is enabled
func (handler *enableEpochsHandler) IsOptimizeNFTStoreFlagEnabled() bool {
	return handler.optimizeNFTStoreFlag.IsSet()
}

// IsCreateNFTThroughExecByCallerFlagEnabled returns true if createNFTThroughExecByCallerFlag is enabled
func (handler *enableEpochsHandler) IsCreateNFTThroughExecByCallerFlagEnabled() bool {
	return handler.createNFTThroughExecByCallerFlag.IsSet()
}

// IsStopDecreasingValidatorRatingWhenStuckFlagEnabled returns true if stopDecreasingValidatorRatingWhenStuckFlag is enabled
func (handler *enableEpochsHandler) IsStopDecreasingValidatorRatingWhenStuckFlagEnabled() bool {
	return handler.stopDecreasingValidatorRatingWhenStuckFlag.IsSet()
}

// IsFrontRunningProtectionFlagEnabled returns true if frontRunningProtectionFlag is enabled
func (handler *enableEpochsHandler) IsFrontRunningProtectionFlagEnabled() bool {
	return handler.frontRunningProtectionFlag.IsSet()
}

// IsPayableBySCFlagEnabled returns true if isPayableBySCFlag is enabled
func (handler *enableEpochsHandler) IsPayableBySCFlagEnabled() bool {
	return handler.isPayableBySCFlag.IsSet()
}

// IsCleanUpInformativeSCRsFlagEnabled returns true if cleanUpInformativeSCRsFlag is enabled
func (handler *enableEpochsHandler) IsCleanUpInformativeSCRsFlagEnabled() bool {
	return handler.cleanUpInformativeSCRsFlag.IsSet()
}

// IsStorageAPICostOptimizationFlagEnabled returns true if storageAPICostOptimizationFlag is enabled
func (handler *enableEpochsHandler) IsStorageAPICostOptimizationFlagEnabled() bool {
	return handler.storageAPICostOptimizationFlag.IsSet()
}

// IsESDTRegisterAndSetAllRolesFlagEnabled returns true if esdtRegisterAndSetAllRolesFlag is enabled
func (handler *enableEpochsHandler) IsESDTRegisterAndSetAllRolesFlagEnabled() bool {
	return handler.esdtRegisterAndSetAllRolesFlag.IsSet()
}

// IsScheduledMiniBlocksFlagEnabled returns true if scheduledMiniBlocksFlag is enabled
func (handler *enableEpochsHandler) IsScheduledMiniBlocksFlagEnabled() bool {
	return handler.scheduledMiniBlocksFlag.IsSet()
}

// IsCorrectJailedNotUnstakedEmptyQueueFlagEnabled returns true if correctJailedNotUnstakedEmptyQueueFlag is enabled
func (handler *enableEpochsHandler) IsCorrectJailedNotUnstakedEmptyQueueFlagEnabled() bool {
	return handler.correctJailedNotUnstakedEmptyQueueFlag.IsSet()
}

// IsDoNotReturnOldBlockInBlockchainHookFlagEnabled returns true if doNotReturnOldBlockInBlockchainHookFlag is enabled
func (handler *enableEpochsHandler) IsDoNotReturnOldBlockInBlockchainHookFlagEnabled() bool {
	return handler.doNotReturnOldBlockInBlockchainHookFlag.IsSet()
}

// IsSCRSizeInvariantOnBuiltInResultFlagEnabled returns true if scrSizeInvariantOnBuiltInResultFlag is enabled
func (handler *enableEpochsHandler) IsSCRSizeInvariantOnBuiltInResultFlagEnabled() bool {
	return handler.scrSizeInvariantOnBuiltInResultFlag.IsSet()
}

// IsCheckCorrectTokenIDForTransferRoleFlagEnabled returns true if checkCorrectTokenIDForTransferRoleFlag is enabled
func (handler *enableEpochsHandler) IsCheckCorrectTokenIDForTransferRoleFlagEnabled() bool {
	return handler.checkCorrectTokenIDForTransferRoleFlag.IsSet()
}

// IsFailExecutionOnEveryAPIErrorFlagEnabled returns true if failExecutionOnEveryAPIErrorFlag is enabled
func (handler *enableEpochsHandler) IsFailExecutionOnEveryAPIErrorFlagEnabled() bool {
	return handler.failExecutionOnEveryAPIErrorFlag.IsSet()
}

// IsHeartbeatDisableFlagEnabled returns true if heartbeatDisableFlag is enabled
func (handler *enableEpochsHandler) IsHeartbeatDisableFlagEnabled() bool {
	return handler.heartbeatDisableFlag.IsSet()
}

// IsMiniBlockPartialExecutionFlagEnabled returns true if isMiniBlockPartialExecutionFlag is enabled
func (handler *enableEpochsHandler) IsMiniBlockPartialExecutionFlagEnabled() bool {
	return handler.isMiniBlockPartialExecutionFlag.IsSet()
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *enableEpochsHandler) IsInterfaceNil() bool {
	return handler == nil
}
