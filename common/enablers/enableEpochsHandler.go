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
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SCDeployEnableEpoch, handler.scDeployFlag, "scDeployFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.BuiltInFunctionsEnableEpoch, handler.builtInFunctionsFlag, "builtInFunctionsFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RelayedTransactionsEnableEpoch, handler.relayedTransactionsFlag, "relayedTransactionsFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.PenalizedTooMuchGasEnableEpoch, handler.penalizedTooMuchGasFlag, "penalizedTooMuchGasFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SwitchJailWaitingEnableEpoch, handler.switchJailWaitingFlag, "switchJailWaitingFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.BelowSignedThresholdEnableEpoch, handler.belowSignedThresholdFlag, "belowSignedThresholdFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SwitchHysteresisForMinNodesEnableEpoch, handler.switchHysteresisForMinNodesFlag, "switchHysteresisForMinNodesFlag")
	handler.setFlagValue(epoch == handler.enableEpochsConfig.SwitchHysteresisForMinNodesEnableEpoch, handler.switchHysteresisForMinNodesCurrentEpochFlag, "switchHysteresisForMinNodesCurrentEpochFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.TransactionSignedWithTxHashEnableEpoch, handler.transactionSignedWithTxHashFlag, "transactionSignedWithTxHashFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.MetaProtectionEnableEpoch, handler.metaProtectionFlag, "metaProtectionFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.AheadOfTimeGasUsageEnableEpoch, handler.aheadOfTimeGasUsageFlag, "aheadOfTimeGasUsageFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.GasPriceModifierEnableEpoch, handler.gasPriceModifierFlag, "gasPriceModifierFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RepairCallbackEnableEpoch, handler.repairCallbackFlag, "repairCallbackFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.BalanceWaitingListsEnableEpoch, handler.balanceWaitingListsFlag, "balanceWaitingListsFlag")
	handler.setFlagValue(epoch > handler.enableEpochsConfig.ReturnDataToLastTransferEnableEpoch, handler.returnDataToLastTransferFlag, "returnDataToLastTransferFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SenderInOutTransferEnableEpoch, handler.senderInOutTransferFlag, "senderInOutTransferFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.StakeEnableEpoch, handler.stakeFlag, "stakeFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.StakingV2EnableEpoch, handler.stakingV2Flag, "stakingV2Flag")
	handler.setFlagValue(epoch == handler.enableEpochsConfig.StakingV2EnableEpoch, handler.stakingV2OwnerFlag, "stakingV2OwnerFlag")
	handler.setFlagValue(epoch > handler.enableEpochsConfig.StakingV2EnableEpoch, handler.stakingV2GreaterEpochFlag, "stakingV2GreaterEpochFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.DoubleKeyProtectionEnableEpoch, handler.doubleKeyProtectionFlag, "doubleKeyProtectionFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTEnableEpoch, handler.esdtFlag, "esdtFlag")
	handler.setFlagValue(epoch == handler.enableEpochsConfig.ESDTEnableEpoch, handler.esdtCurrentEpochFlag, "esdtCurrentEpochFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.GovernanceEnableEpoch, handler.governanceFlag, "governanceFlag")
	handler.setFlagValue(epoch == handler.enableEpochsConfig.GovernanceEnableEpoch, handler.governanceCurrentEpochFlag, "governanceCurrentEpochFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.DelegationManagerEnableEpoch, handler.delegationManagerFlag, "delegationManagerFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.DelegationSmartContractEnableEpoch, handler.delegationSmartContractFlag, "delegationSmartContractFlag")
	handler.setFlagValue(epoch == handler.enableEpochsConfig.DelegationSmartContractEnableEpoch, handler.delegationSmartContractCurrentEpochFlag, "delegationSmartContractCurrentEpochFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch, handler.correctLastUnJailedFlag, "correctLastUnJailedFlag")
	handler.setFlagValue(epoch == handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch, handler.correctLastUnJailedCurrentEpochFlag, "correctLastUnJailedCurrentEpochFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RelayedTransactionsV2EnableEpoch, handler.relayedTransactionsV2Flag, "relayedTransactionsV2Flag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.UnbondTokensV2EnableEpoch, handler.unBondTokensV2Flag, "unBondTokensV2Flag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SaveJailedAlwaysEnableEpoch, handler.saveJailedAlwaysFlag, "saveJailedAlwaysFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ReDelegateBelowMinCheckEnableEpoch, handler.reDelegateBelowMinCheckFlag, "reDelegateBelowMinCheckFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ValidatorToDelegationEnableEpoch, handler.validatorToDelegationFlag, "validatorToDelegationFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.WaitingListFixEnableEpoch, handler.waitingListFixFlag, "waitingListFixFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.IncrementSCRNonceInMultiTransferEnableEpoch, handler.incrementSCRNonceInMultiTransferFlag, "incrementSCRNonceInMultiTransferFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTMultiTransferEnableEpoch, handler.esdtMultiTransferFlag, "esdtMultiTransferFlag")
	handler.setFlagValue(epoch < handler.enableEpochsConfig.GlobalMintBurnDisableEpoch, handler.globalMintBurnFlag, "globalMintBurnFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTTransferRoleEnableEpoch, handler.esdtTransferRoleFlag, "esdtTransferRoleFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.BuiltInFunctionOnMetaEnableEpoch, handler.builtInFunctionOnMetaFlag, "builtInFunctionOnMetaFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ComputeRewardCheckpointEnableEpoch, handler.computeRewardCheckpointFlag, "computeRewardCheckpointFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SCRSizeInvariantCheckEnableEpoch, handler.scrSizeInvariantCheckFlag, "scrSizeInvariantCheckFlag")
	handler.setFlagValue(epoch < handler.enableEpochsConfig.BackwardCompSaveKeyValueEnableEpoch, handler.backwardCompSaveKeyValueFlag, "backwardCompSaveKeyValueFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTNFTCreateOnMultiShardEnableEpoch, handler.esdtNFTCreateOnMultiShardFlag, "esdtNFTCreateOnMultiShardFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.MetaESDTSetEnableEpoch, handler.metaESDTSetFlag, "metaESDTSetFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.AddTokensToDelegationEnableEpoch, handler.addTokensToDelegationFlag, "addTokensToDelegationFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.MultiESDTTransferFixOnCallBackOnEnableEpoch, handler.multiESDTTransferFixOnCallBackFlag, "multiESDTTransferFixOnCallBackFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.OptimizeGasUsedInCrossMiniBlocksEnableEpoch, handler.optimizeGasUsedInCrossMiniBlocksFlag, "optimizeGasUsedInCrossMiniBlocksFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CorrectFirstQueuedEpoch, handler.correctFirstQueuedFlag, "correctFirstQueuedFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.DeleteDelegatorAfterClaimRewardsEnableEpoch, handler.deleteDelegatorAfterClaimRewardsFlag, "deleteDelegatorAfterClaimRewardsFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.FixOOGReturnCodeEnableEpoch, handler.fixOOGReturnCodeFlag, "fixOOGReturnCodeFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RemoveNonUpdatedStorageEnableEpoch, handler.removeNonUpdatedStorageFlag, "removeNonUpdatedStorageFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch, handler.optimizeNFTStoreFlag, "optimizeNFTStoreFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CreateNFTThroughExecByCallerEnableEpoch, handler.createNFTThroughExecByCallerFlag, "createNFTThroughExecByCallerFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.StopDecreasingValidatorRatingWhenStuckEnableEpoch, handler.stopDecreasingValidatorRatingWhenStuckFlag, "stopDecreasingValidatorRatingWhenStuckFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.FrontRunningProtectionEnableEpoch, handler.frontRunningProtectionFlag, "frontRunningProtectionFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.IsPayableBySCEnableEpoch, handler.isPayableBySCFlag, "isPayableBySCFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CleanUpInformativeSCRsEnableEpoch, handler.cleanUpInformativeSCRsFlag, "cleanUpInformativeSCRsFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.StorageAPICostOptimizationEnableEpoch, handler.storageAPICostOptimizationFlag, "storageAPICostOptimizationFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTRegisterAndSetAllRolesEnableEpoch, handler.esdtRegisterAndSetAllRolesFlag, "esdtRegisterAndSetAllRolesFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ScheduledMiniBlocksEnableEpoch, handler.scheduledMiniBlocksFlag, "scheduledMiniBlocksFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CorrectJailedNotUnstakedEmptyQueueEpoch, handler.correctJailedNotUnStakedEmptyQueueFlag, "correctJailedNotUnStakedEmptyQueueFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.DoNotReturnOldBlockInBlockchainHookEnableEpoch, handler.doNotReturnOldBlockInBlockchainHookFlag, "doNotReturnOldBlockInBlockchainHookFlag")
	handler.setFlagValue(epoch < handler.enableEpochsConfig.AddFailedRelayedTxToInvalidMBsDisableEpoch, handler.addFailedRelayedTxToInvalidMBsFlag, "addFailedRelayedTxToInvalidMBsFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SCRSizeInvariantOnBuiltInResultEnableEpoch, handler.scrSizeInvariantOnBuiltInResultFlag, "scrSizeInvariantOnBuiltInResultFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CheckCorrectTokenIDForTransferRoleEnableEpoch, handler.checkCorrectTokenIDForTransferRoleFlag, "checkCorrectTokenIDForTransferRoleFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.FailExecutionOnEveryAPIErrorEnableEpoch, handler.failExecutionOnEveryAPIErrorFlag, "failExecutionOnEveryAPIErrorFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.MiniBlockPartialExecutionEnableEpoch, handler.isMiniBlockPartialExecutionFlag, "isMiniBlockPartialExecutionFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ManagedCryptoAPIsEnableEpoch, handler.managedCryptoAPIsFlag, "managedCryptoAPIsFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch, handler.esdtMetadataContinuousCleanupFlag, "esdtMetadataContinuousCleanupFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.DisableExecByCallerEnableEpoch, handler.disableExecByCallerFlag, "disableExecByCallerFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RefactorContextEnableEpoch, handler.refactorContextFlag, "refactorContextFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CheckFunctionArgumentEnableEpoch, handler.checkFunctionArgumentFlag, "checkFunctionArgumentFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.CheckExecuteOnReadOnlyEnableEpoch, handler.checkExecuteOnReadOnlyFlag, "checkExecuteOnReadOnlyFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SetSenderInEeiOutputTransferEnableEpoch, handler.setSenderInEeiOutputTransferFlag, "setSenderInEeiOutputTransferFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch, handler.changeDelegationOwnerFlag, "changeDelegationOwnerFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RefactorPeersMiniBlocksEnableEpoch, handler.refactorPeersMiniBlocksFlag, "refactorPeersMiniBlocksFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.FixAsyncCallBackArgsListEnableEpoch, handler.fixAsyncCallBackArgsList, "fixAsyncCallBackArgsList")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.FixOldTokenLiquidityEnableEpoch, handler.fixOldTokenLiquidity, "fixOldTokenLiquidity")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RuntimeMemStoreLimitEnableEpoch, handler.runtimeMemStoreLimitFlag, "runtimeMemStoreLimitFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RuntimeCodeSizeFixEnableEpoch, handler.runtimeCodeSizeFixFlag, "runtimeCodeSizeFixFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.MaxBlockchainHookCountersEnableEpoch, handler.maxBlockchainHookCountersFlag, "maxBlockchainHookCountersFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.WipeSingleNFTLiquidityDecreaseEnableEpoch, handler.wipeSingleNFTLiquidityDecreaseFlag, "wipeSingleNFTLiquidityDecreaseFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.AlwaysSaveTokenMetaDataEnableEpoch, handler.alwaysSaveTokenMetaDataFlag, "alwaysSaveTokenMetaDataFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SetGuardianEnableEpoch, handler.setGuardianFlag, "setGuardianFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.RelayedNonceFixEnableEpoch, handler.relayedNonceFixFlag, "relayedNonceFixFlag")
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.DeterministicSortOnValidatorsInfoEnableEpoch, handler.deterministicSortOnValidatorsInfoFixFlag, "deterministicSortOnValidatorsInfoFixFlag")
}

func (handler *enableEpochsHandler) setFlagValue(value bool, flag *atomic.Flag, flagName string) {
	flag.SetValue(value)
	log.Debug("EpochConfirmed", "flag", flagName, "enabled", flag.IsSet())
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
