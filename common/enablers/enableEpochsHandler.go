package enablers

import (
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("common/enablers")

type enableEpochsHandler struct {
	*epochFlagsHolder
	enableEpochsConfig config.EnableEpochs
}

// NewEnableEpochsHandler creates a new instance of enableEpochsHandler
func NewEnableEpochsHandler(enableEpochsConfig config.EnableEpochs, epochNotifier process.EpochNotifier) (*enableEpochsHandler, error) {
	if check.IfNil(epochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	handler := &enableEpochsHandler{
		epochFlagsHolder:   newEpochFlagsHolder(),
		enableEpochsConfig: enableEpochsConfig,
	}

	epochNotifier.RegisterNotifyHandler(handler)

	return handler, nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (handler *enableEpochsHandler) EpochConfirmed(epoch uint32, _ uint64) {
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SCDeployEnableEpoch, handler.scDeployFlag, "scDeployFlag", epoch, handler.enableEpochsConfig.SCDeployEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.BuiltInFunctionsEnableEpoch, handler.builtInFunctionsFlag, "builtInFunctionsFlag", epoch, handler.enableEpochsConfig.BuiltInFunctionsEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RelayedTransactionsEnableEpoch, handler.relayedTransactionsFlag, "relayedTransactionsFlag", epoch, handler.enableEpochsConfig.RelayedTransactionsEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.PenalizedTooMuchGasEnableEpoch, handler.penalizedTooMuchGasFlag, "penalizedTooMuchGasFlag", epoch, handler.enableEpochsConfig.PenalizedTooMuchGasEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SwitchJailWaitingEnableEpoch, handler.switchJailWaitingFlag, "switchJailWaitingFlag", epoch, handler.enableEpochsConfig.SwitchJailWaitingEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.BelowSignedThresholdEnableEpoch, handler.belowSignedThresholdFlag, "belowSignedThresholdFlag", epoch, handler.enableEpochsConfig.BelowSignedThresholdEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SwitchHysteresisForMinNodesEnableEpoch, handler.switchHysteresisForMinNodesFlag, "switchHysteresisForMinNodesFlag", epoch, handler.enableEpochsConfig.SwitchHysteresisForMinNodesEnableEpoch)
	handler.setFlagValue(epoch == handler.enableEpochsConfig.SwitchHysteresisForMinNodesEnableEpoch, handler.switchHysteresisForMinNodesCurrentEpochFlag, "switchHysteresisForMinNodesCurrentEpochFlag", epoch, handler.enableEpochsConfig.SwitchHysteresisForMinNodesEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.TransactionSignedWithTxHashEnableEpoch, handler.transactionSignedWithTxHashFlag, "transactionSignedWithTxHashFlag", epoch, handler.enableEpochsConfig.TransactionSignedWithTxHashEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.MetaProtectionEnableEpoch, handler.metaProtectionFlag, "metaProtectionFlag", epoch, handler.enableEpochsConfig.MetaProtectionEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.AheadOfTimeGasUsageEnableEpoch, handler.aheadOfTimeGasUsageFlag, "aheadOfTimeGasUsageFlag", epoch, handler.enableEpochsConfig.AheadOfTimeGasUsageEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.GasPriceModifierEnableEpoch, handler.gasPriceModifierFlag, "gasPriceModifierFlag", epoch, handler.enableEpochsConfig.GasPriceModifierEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RepairCallbackEnableEpoch, handler.repairCallbackFlag, "repairCallbackFlag", epoch, handler.enableEpochsConfig.RepairCallbackEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.BalanceWaitingListsEnableEpoch, handler.balanceWaitingListsFlag, "balanceWaitingListsFlag", epoch, handler.enableEpochsConfig.BalanceWaitingListsEnableEpoch)
	handler.setFlagValue(epoch > handler.enableEpochsConfig.ReturnDataToLastTransferEnableEpoch, handler.returnDataToLastTransferFlag, "returnDataToLastTransferFlag", epoch, handler.enableEpochsConfig.ReturnDataToLastTransferEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SenderInOutTransferEnableEpoch, handler.senderInOutTransferFlag, "senderInOutTransferFlag", epoch, handler.enableEpochsConfig.SenderInOutTransferEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.StakeEnableEpoch, handler.stakeFlag, "stakeFlag", epoch, handler.enableEpochsConfig.StakeEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.StakingV2EnableEpoch, handler.stakingV2Flag, "stakingV2Flag", epoch, handler.enableEpochsConfig.StakingV2EnableEpoch)
	handler.setFlagValue(epoch == handler.enableEpochsConfig.StakingV2EnableEpoch, handler.stakingV2OwnerFlag, "stakingV2OwnerFlag", epoch, handler.enableEpochsConfig.StakingV2EnableEpoch)
	handler.setFlagValue(epoch > handler.enableEpochsConfig.StakingV2EnableEpoch, handler.stakingV2GreaterEpochFlag, "stakingV2GreaterEpochFlag", epoch, handler.enableEpochsConfig.StakingV2EnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.DoubleKeyProtectionEnableEpoch, handler.doubleKeyProtectionFlag, "doubleKeyProtectionFlag", epoch, handler.enableEpochsConfig.DoubleKeyProtectionEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTEnableEpoch, handler.esdtFlag, "esdtFlag", epoch, handler.enableEpochsConfig.ESDTEnableEpoch)
	handler.setFlagValue(epoch == handler.enableEpochsConfig.ESDTEnableEpoch, handler.esdtCurrentEpochFlag, "esdtCurrentEpochFlag", epoch, handler.enableEpochsConfig.ESDTEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.GovernanceEnableEpoch, handler.governanceFlag, "governanceFlag", epoch, handler.enableEpochsConfig.GovernanceEnableEpoch)
	handler.setFlagValue(epoch == handler.enableEpochsConfig.GovernanceEnableEpoch, handler.governanceCurrentEpochFlag, "governanceCurrentEpochFlag", epoch, handler.enableEpochsConfig.GovernanceEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.DelegationManagerEnableEpoch, handler.delegationManagerFlag, "delegationManagerFlag", epoch, handler.enableEpochsConfig.DelegationManagerEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.DelegationSmartContractEnableEpoch, handler.delegationSmartContractFlag, "delegationSmartContractFlag", epoch, handler.enableEpochsConfig.DelegationSmartContractEnableEpoch)
	handler.setFlagValue(epoch == handler.enableEpochsConfig.DelegationSmartContractEnableEpoch, handler.delegationSmartContractCurrentEpochFlag, "delegationSmartContractCurrentEpochFlag", epoch, handler.enableEpochsConfig.DelegationSmartContractEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch, handler.correctLastUnJailedFlag, "correctLastUnJailedFlag", epoch, handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch)
	handler.setFlagValue(epoch == handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch, handler.correctLastUnJailedCurrentEpochFlag, "correctLastUnJailedCurrentEpochFlag", epoch, handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RelayedTransactionsV2EnableEpoch, handler.relayedTransactionsV2Flag, "relayedTransactionsV2Flag", epoch, handler.enableEpochsConfig.RelayedTransactionsV2EnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.UnbondTokensV2EnableEpoch, handler.unBondTokensV2Flag, "unBondTokensV2Flag", epoch, handler.enableEpochsConfig.UnbondTokensV2EnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SaveJailedAlwaysEnableEpoch, handler.saveJailedAlwaysFlag, "saveJailedAlwaysFlag", epoch, handler.enableEpochsConfig.SaveJailedAlwaysEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ReDelegateBelowMinCheckEnableEpoch, handler.reDelegateBelowMinCheckFlag, "reDelegateBelowMinCheckFlag", epoch, handler.enableEpochsConfig.ReDelegateBelowMinCheckEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ValidatorToDelegationEnableEpoch, handler.validatorToDelegationFlag, "validatorToDelegationFlag", epoch, handler.enableEpochsConfig.ValidatorToDelegationEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.WaitingListFixEnableEpoch, handler.waitingListFixFlag, "waitingListFixFlag", epoch, handler.enableEpochsConfig.WaitingListFixEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.IncrementSCRNonceInMultiTransferEnableEpoch, handler.incrementSCRNonceInMultiTransferFlag, "incrementSCRNonceInMultiTransferFlag", epoch, handler.enableEpochsConfig.IncrementSCRNonceInMultiTransferEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTMultiTransferEnableEpoch, handler.esdtMultiTransferFlag, "esdtMultiTransferFlag", epoch, handler.enableEpochsConfig.ESDTMultiTransferEnableEpoch)
	handler.setFlagValue(epoch < handler.enableEpochsConfig.GlobalMintBurnDisableEpoch, handler.globalMintBurnFlag, "globalMintBurnFlag", epoch, handler.enableEpochsConfig.GlobalMintBurnDisableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTTransferRoleEnableEpoch, handler.esdtTransferRoleFlag, "esdtTransferRoleFlag", epoch, handler.enableEpochsConfig.ESDTTransferRoleEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.BuiltInFunctionOnMetaEnableEpoch, handler.builtInFunctionOnMetaFlag, "builtInFunctionOnMetaFlag", epoch, handler.enableEpochsConfig.BuiltInFunctionOnMetaEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ComputeRewardCheckpointEnableEpoch, handler.computeRewardCheckpointFlag, "computeRewardCheckpointFlag", epoch, handler.enableEpochsConfig.ComputeRewardCheckpointEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SCRSizeInvariantCheckEnableEpoch, handler.scrSizeInvariantCheckFlag, "scrSizeInvariantCheckFlag", epoch, handler.enableEpochsConfig.SCRSizeInvariantCheckEnableEpoch)
	handler.setFlagValue(epoch < handler.enableEpochsConfig.BackwardCompSaveKeyValueEnableEpoch, handler.backwardCompSaveKeyValueFlag, "backwardCompSaveKeyValueFlag", epoch, handler.enableEpochsConfig.BackwardCompSaveKeyValueEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTNFTCreateOnMultiShardEnableEpoch, handler.esdtNFTCreateOnMultiShardFlag, "esdtNFTCreateOnMultiShardFlag", epoch, handler.enableEpochsConfig.ESDTNFTCreateOnMultiShardEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.MetaESDTSetEnableEpoch, handler.metaESDTSetFlag, "metaESDTSetFlag", epoch, handler.enableEpochsConfig.MetaESDTSetEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.AddTokensToDelegationEnableEpoch, handler.addTokensToDelegationFlag, "addTokensToDelegationFlag", epoch, handler.enableEpochsConfig.AddTokensToDelegationEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.MultiESDTTransferFixOnCallBackOnEnableEpoch, handler.multiESDTTransferFixOnCallBackFlag, "multiESDTTransferFixOnCallBackFlag", epoch, handler.enableEpochsConfig.MultiESDTTransferFixOnCallBackOnEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.OptimizeGasUsedInCrossMiniBlocksEnableEpoch, handler.optimizeGasUsedInCrossMiniBlocksFlag, "optimizeGasUsedInCrossMiniBlocksFlag", epoch, handler.enableEpochsConfig.OptimizeGasUsedInCrossMiniBlocksEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CorrectFirstQueuedEpoch, handler.correctFirstQueuedFlag, "correctFirstQueuedFlag", epoch, handler.enableEpochsConfig.CorrectFirstQueuedEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.DeleteDelegatorAfterClaimRewardsEnableEpoch, handler.deleteDelegatorAfterClaimRewardsFlag, "deleteDelegatorAfterClaimRewardsFlag", epoch, handler.enableEpochsConfig.DeleteDelegatorAfterClaimRewardsEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.FixOOGReturnCodeEnableEpoch, handler.fixOOGReturnCodeFlag, "fixOOGReturnCodeFlag", epoch, handler.enableEpochsConfig.FixOOGReturnCodeEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RemoveNonUpdatedStorageEnableEpoch, handler.removeNonUpdatedStorageFlag, "removeNonUpdatedStorageFlag", epoch, handler.enableEpochsConfig.RemoveNonUpdatedStorageEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch, handler.optimizeNFTStoreFlag, "optimizeNFTStoreFlag", epoch, handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CreateNFTThroughExecByCallerEnableEpoch, handler.createNFTThroughExecByCallerFlag, "createNFTThroughExecByCallerFlag", epoch, handler.enableEpochsConfig.CreateNFTThroughExecByCallerEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.StopDecreasingValidatorRatingWhenStuckEnableEpoch, handler.stopDecreasingValidatorRatingWhenStuckFlag, "stopDecreasingValidatorRatingWhenStuckFlag", epoch, handler.enableEpochsConfig.StopDecreasingValidatorRatingWhenStuckEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.FrontRunningProtectionEnableEpoch, handler.frontRunningProtectionFlag, "frontRunningProtectionFlag", epoch, handler.enableEpochsConfig.FrontRunningProtectionEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.IsPayableBySCEnableEpoch, handler.isPayableBySCFlag, "isPayableBySCFlag", epoch, handler.enableEpochsConfig.IsPayableBySCEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CleanUpInformativeSCRsEnableEpoch, handler.cleanUpInformativeSCRsFlag, "cleanUpInformativeSCRsFlag", epoch, handler.enableEpochsConfig.CleanUpInformativeSCRsEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.StorageAPICostOptimizationEnableEpoch, handler.storageAPICostOptimizationFlag, "storageAPICostOptimizationFlag", epoch, handler.enableEpochsConfig.StorageAPICostOptimizationEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTRegisterAndSetAllRolesEnableEpoch, handler.esdtRegisterAndSetAllRolesFlag, "esdtRegisterAndSetAllRolesFlag", epoch, handler.enableEpochsConfig.ESDTRegisterAndSetAllRolesEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ScheduledMiniBlocksEnableEpoch, handler.scheduledMiniBlocksFlag, "scheduledMiniBlocksFlag", epoch, handler.enableEpochsConfig.ScheduledMiniBlocksEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CorrectJailedNotUnstakedEmptyQueueEpoch, handler.correctJailedNotUnStakedEmptyQueueFlag, "correctJailedNotUnStakedEmptyQueueFlag", epoch, handler.enableEpochsConfig.CorrectJailedNotUnstakedEmptyQueueEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.DoNotReturnOldBlockInBlockchainHookEnableEpoch, handler.doNotReturnOldBlockInBlockchainHookFlag, "doNotReturnOldBlockInBlockchainHookFlag", epoch, handler.enableEpochsConfig.DoNotReturnOldBlockInBlockchainHookEnableEpoch)
	handler.setFlagValue(epoch < handler.enableEpochsConfig.AddFailedRelayedTxToInvalidMBsDisableEpoch, handler.addFailedRelayedTxToInvalidMBsFlag, "addFailedRelayedTxToInvalidMBsFlag", epoch, handler.enableEpochsConfig.AddFailedRelayedTxToInvalidMBsDisableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SCRSizeInvariantOnBuiltInResultEnableEpoch, handler.scrSizeInvariantOnBuiltInResultFlag, "scrSizeInvariantOnBuiltInResultFlag", epoch, handler.enableEpochsConfig.SCRSizeInvariantOnBuiltInResultEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CheckCorrectTokenIDForTransferRoleEnableEpoch, handler.checkCorrectTokenIDForTransferRoleFlag, "checkCorrectTokenIDForTransferRoleFlag", epoch, handler.enableEpochsConfig.CheckCorrectTokenIDForTransferRoleEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.FailExecutionOnEveryAPIErrorEnableEpoch, handler.failExecutionOnEveryAPIErrorFlag, "failExecutionOnEveryAPIErrorFlag", epoch, handler.enableEpochsConfig.FailExecutionOnEveryAPIErrorEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.MiniBlockPartialExecutionEnableEpoch, handler.isMiniBlockPartialExecutionFlag, "isMiniBlockPartialExecutionFlag", epoch, handler.enableEpochsConfig.MiniBlockPartialExecutionEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ManagedCryptoAPIsEnableEpoch, handler.managedCryptoAPIsFlag, "managedCryptoAPIsFlag", epoch, handler.enableEpochsConfig.ManagedCryptoAPIsEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch, handler.esdtMetadataContinuousCleanupFlag, "esdtMetadataContinuousCleanupFlag", epoch, handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.DisableExecByCallerEnableEpoch, handler.disableExecByCallerFlag, "disableExecByCallerFlag", epoch, handler.enableEpochsConfig.DisableExecByCallerEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RefactorContextEnableEpoch, handler.refactorContextFlag, "refactorContextFlag", epoch, handler.enableEpochsConfig.RefactorContextEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CheckFunctionArgumentEnableEpoch, handler.checkFunctionArgumentFlag, "checkFunctionArgumentFlag", epoch, handler.enableEpochsConfig.CheckFunctionArgumentEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CheckExecuteOnReadOnlyEnableEpoch, handler.checkExecuteOnReadOnlyFlag, "checkExecuteOnReadOnlyFlag", epoch, handler.enableEpochsConfig.CheckExecuteOnReadOnlyEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SetSenderInEeiOutputTransferEnableEpoch, handler.setSenderInEeiOutputTransferFlag, "setSenderInEeiOutputTransferFlag", epoch, handler.enableEpochsConfig.SetSenderInEeiOutputTransferEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch, handler.changeDelegationOwnerFlag, "changeDelegationOwnerFlag", epoch, handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RefactorPeersMiniBlocksEnableEpoch, handler.refactorPeersMiniBlocksFlag, "refactorPeersMiniBlocksFlag", epoch, handler.enableEpochsConfig.RefactorPeersMiniBlocksEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.FixAsyncCallBackArgsListEnableEpoch, handler.fixAsyncCallBackArgsList, "fixAsyncCallBackArgsList", epoch, handler.enableEpochsConfig.FixAsyncCallBackArgsListEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.FixOldTokenLiquidityEnableEpoch, handler.fixOldTokenLiquidity, "fixOldTokenLiquidity", epoch, handler.enableEpochsConfig.FixOldTokenLiquidityEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RuntimeMemStoreLimitEnableEpoch, handler.runtimeMemStoreLimitFlag, "runtimeMemStoreLimitFlag", epoch, handler.enableEpochsConfig.RuntimeMemStoreLimitEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RuntimeCodeSizeFixEnableEpoch, handler.runtimeCodeSizeFixFlag, "runtimeCodeSizeFixFlag", epoch, handler.enableEpochsConfig.RuntimeCodeSizeFixEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.MaxBlockchainHookCountersEnableEpoch, handler.maxBlockchainHookCountersFlag, "maxBlockchainHookCountersFlag", epoch, handler.enableEpochsConfig.MaxBlockchainHookCountersEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.WipeSingleNFTLiquidityDecreaseEnableEpoch, handler.wipeSingleNFTLiquidityDecreaseFlag, "wipeSingleNFTLiquidityDecreaseFlag", epoch, handler.enableEpochsConfig.WipeSingleNFTLiquidityDecreaseEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.AlwaysSaveTokenMetaDataEnableEpoch, handler.alwaysSaveTokenMetaDataFlag, "alwaysSaveTokenMetaDataFlag", epoch, handler.enableEpochsConfig.AlwaysSaveTokenMetaDataEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RelayedNonceFixEnableEpoch, handler.relayedNonceFixFlag, "relayedNonceFixFlag", epoch, handler.enableEpochsConfig.RelayedNonceFixEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SetGuardianEnableEpoch, handler.setGuardianFlag, "setGuardianFlag", epoch, handler.enableEpochsConfig.SetGuardianEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ScToScLogEventEnableEpoch, handler.scToScLogEventFlag, "setScToScLogEventFlag", epoch, handler.enableEpochsConfig.ScToScLogEventEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.MultiClaimOnDelegationEnableEpoch, handler.multiClaimOnDelegationFlag, "multiClaimOnDelegationFlag", epoch, handler.enableEpochsConfig.MultiClaimOnDelegationEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.KeepExecOrderOnCreatedSCRsEnableEpoch, handler.keepExecOrderOnCreatedSCRsFlag, "keepExecOrderOnCreatedSCRsFlag", epoch, handler.enableEpochsConfig.KeepExecOrderOnCreatedSCRsEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ChangeUsernameEnableEpoch, handler.changeUsernameFlag, "changeUsername", epoch, handler.enableEpochsConfig.ChangeUsernameEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ConsistentTokensValuesLengthCheckEnableEpoch, handler.consistentTokensValuesCheckFlag, "consistentTokensValuesCheckFlag", epoch, handler.enableEpochsConfig.ConsistentTokensValuesLengthCheckEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.AutoBalanceDataTriesEnableEpoch, handler.autoBalanceDataTriesFlag, "autoBalanceDataTriesFlag", epoch, handler.enableEpochsConfig.AutoBalanceDataTriesEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.FixDelegationChangeOwnerOnAccountEnableEpoch, handler.fixDelegationChangeOwnerOnAccountFlag, "fixDelegationChangeOwnerOnAccountFlag", epoch, handler.enableEpochsConfig.FixDelegationChangeOwnerOnAccountEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SCProcessorV2EnableEpoch, handler.scProcessorV2Flag, "scProcessorV2Flag", epoch, handler.enableEpochsConfig.SCProcessorV2EnableEpoch)
}

func (handler *enableEpochsHandler) setFlagValue(value bool, flag *atomic.Flag, flagName string, epoch uint32, flagEpoch uint32) {
	flag.SetValue(value)
	log.Debug("EpochConfirmed", "flag", flagName, "enabled", flag.IsSet(), "epoch", epoch, "flag epoch", flagEpoch)
}

// ScheduledMiniBlocksEnableEpoch returns the epoch when scheduled mini blocks becomes active
func (handler *enableEpochsHandler) ScheduledMiniBlocksEnableEpoch() uint32 {
	return handler.enableEpochsConfig.ScheduledMiniBlocksEnableEpoch
}

// BlockGasAndFeesReCheckEnableEpoch returns the epoch when block gas and fees recheck becomes active
func (handler *enableEpochsHandler) BlockGasAndFeesReCheckEnableEpoch() uint32 {
	return handler.enableEpochsConfig.BlockGasAndFeesReCheckEnableEpoch
}

// StakingV2EnableEpoch returns the epoch when staking v2 becomes active
func (handler *enableEpochsHandler) StakingV2EnableEpoch() uint32 {
	return handler.enableEpochsConfig.StakingV2EnableEpoch
}

// SwitchJailWaitingEnableEpoch returns the epoch for switch jail waiting
func (handler *enableEpochsHandler) SwitchJailWaitingEnableEpoch() uint32 {
	return handler.enableEpochsConfig.SwitchJailWaitingEnableEpoch
}

// BalanceWaitingListsEnableEpoch returns the epoch for balance waiting lists
func (handler *enableEpochsHandler) BalanceWaitingListsEnableEpoch() uint32 {
	return handler.enableEpochsConfig.BalanceWaitingListsEnableEpoch
}

// WaitingListFixEnableEpoch returns the epoch for waiting list fix
func (handler *enableEpochsHandler) WaitingListFixEnableEpoch() uint32 {
	return handler.enableEpochsConfig.WaitingListFixEnableEpoch
}

// MultiESDTTransferAsyncCallBackEnableEpoch returns the epoch when multi esdt transfer fix on callback becomes active
func (handler *enableEpochsHandler) MultiESDTTransferAsyncCallBackEnableEpoch() uint32 {
	return handler.enableEpochsConfig.MultiESDTTransferFixOnCallBackOnEnableEpoch
}

// FixOOGReturnCodeEnableEpoch returns the epoch when fix oog return code becomes active
func (handler *enableEpochsHandler) FixOOGReturnCodeEnableEpoch() uint32 {
	return handler.enableEpochsConfig.FixOOGReturnCodeEnableEpoch
}

// RemoveNonUpdatedStorageEnableEpoch returns the epoch for remove non updated storage
func (handler *enableEpochsHandler) RemoveNonUpdatedStorageEnableEpoch() uint32 {
	return handler.enableEpochsConfig.RemoveNonUpdatedStorageEnableEpoch
}

// CreateNFTThroughExecByCallerEnableEpoch returns the epoch when create nft through exec by caller becomes active
func (handler *enableEpochsHandler) CreateNFTThroughExecByCallerEnableEpoch() uint32 {
	return handler.enableEpochsConfig.CreateNFTThroughExecByCallerEnableEpoch
}

// FixFailExecutionOnErrorEnableEpoch returns the epoch when fail execution on error fix becomes active
func (handler *enableEpochsHandler) FixFailExecutionOnErrorEnableEpoch() uint32 {
	return handler.enableEpochsConfig.FailExecutionOnEveryAPIErrorEnableEpoch
}

// ManagedCryptoAPIEnableEpoch returns the epoch when managed crypto api becomes active
func (handler *enableEpochsHandler) ManagedCryptoAPIEnableEpoch() uint32 {
	return handler.enableEpochsConfig.ManagedCryptoAPIsEnableEpoch
}

// DisableExecByCallerEnableEpoch returns the epoch when disable exec by caller becomes active
func (handler *enableEpochsHandler) DisableExecByCallerEnableEpoch() uint32 {
	return handler.enableEpochsConfig.DisableExecByCallerEnableEpoch
}

// RefactorContextEnableEpoch returns the epoch when refactor context becomes active
func (handler *enableEpochsHandler) RefactorContextEnableEpoch() uint32 {
	return handler.enableEpochsConfig.RefactorContextEnableEpoch
}

// CheckExecuteReadOnlyEnableEpoch returns the epoch when check execute readonly becomes active
func (handler *enableEpochsHandler) CheckExecuteReadOnlyEnableEpoch() uint32 {
	return handler.enableEpochsConfig.CheckExecuteOnReadOnlyEnableEpoch
}

// StorageAPICostOptimizationEnableEpoch returns the epoch when storage api cost optimization becomes active
func (handler *enableEpochsHandler) StorageAPICostOptimizationEnableEpoch() uint32 {
	return handler.enableEpochsConfig.StorageAPICostOptimizationEnableEpoch
}

// MiniBlockPartialExecutionEnableEpoch returns the epoch when miniblock partial execution becomes active
func (handler *enableEpochsHandler) MiniBlockPartialExecutionEnableEpoch() uint32 {
	return handler.enableEpochsConfig.MiniBlockPartialExecutionEnableEpoch
}

// RefactorPeersMiniBlocksEnableEpoch returns the epoch when refactor of peers mini blocks becomes active
func (handler *enableEpochsHandler) RefactorPeersMiniBlocksEnableEpoch() uint32 {
	return handler.enableEpochsConfig.RefactorPeersMiniBlocksEnableEpoch
}

// RelayedNonceFixEnableEpoch returns the epoch when relayed nonce fix becomes active
func (handler *enableEpochsHandler) RelayedNonceFixEnableEpoch() uint32 {
	return handler.enableEpochsConfig.RelayedNonceFixEnableEpoch
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *enableEpochsHandler) IsInterfaceNil() bool {
	return handler == nil
}
