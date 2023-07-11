package enablers

import (
	"sync"

	coreAtomic "github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("common/enablers")

type enableEpochsHandler struct {
	*epochFlagsHolder
	enableEpochsConfig config.EnableEpochs
	currentEpoch       uint32
	epochMut           sync.RWMutex
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
	handler.epochMut.Lock()
	handler.currentEpoch = epoch
	handler.epochMut.Unlock()

	// TODO[Sorin]: remove the lines below with epochFlags.go
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
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.MultiClaimOnDelegationEnableEpoch, handler.multiClaimOnDelegationFlag, "multiClaimOnDelegationFlag", epoch, handler.enableEpochsConfig.MultiClaimOnDelegationEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.KeepExecOrderOnCreatedSCRsEnableEpoch, handler.keepExecOrderOnCreatedSCRsFlag, "keepExecOrderOnCreatedSCRsFlag", epoch, handler.enableEpochsConfig.KeepExecOrderOnCreatedSCRsEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ChangeUsernameEnableEpoch, handler.changeUsernameFlag, "changeUsername", epoch, handler.enableEpochsConfig.ChangeUsernameEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.ConsistentTokensValuesLengthCheckEnableEpoch, handler.consistentTokensValuesCheckFlag, "consistentTokensValuesCheckFlag", epoch, handler.enableEpochsConfig.ConsistentTokensValuesLengthCheckEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.AutoBalanceDataTriesEnableEpoch, handler.autoBalanceDataTriesFlag, "autoBalanceDataTriesFlag", epoch, handler.enableEpochsConfig.AutoBalanceDataTriesEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.FixDelegationChangeOwnerOnAccountEnableEpoch, handler.fixDelegationChangeOwnerOnAccountFlag, "fixDelegationChangeOwnerOnAccountFlag", epoch, handler.enableEpochsConfig.FixDelegationChangeOwnerOnAccountEnableEpoch)
	handler.setFlagValue(epoch >= handler.enableEpochsConfig.SCProcessorV2EnableEpoch, handler.scProcessorV2Flag, "scProcessorV2Flag", epoch, handler.enableEpochsConfig.SCProcessorV2EnableEpoch)
}

func (handler *enableEpochsHandler) setFlagValue(value bool, flag *coreAtomic.Flag, flagName string, epoch uint32, flagEpoch uint32) {
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

// GetCurrentEpoch returns the current epoch
func (handler *enableEpochsHandler) GetCurrentEpoch() uint32 {
	handler.epochMut.RLock()
	currentEpoch := handler.currentEpoch
	handler.epochMut.RUnlock()

	return currentEpoch
}

// IsSCDeployFlagEnabledInEpoch returns true if SCDeployEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsSCDeployFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.SCDeployEnableEpoch
}

// IsBuiltInFunctionsFlagEnabledInEpoch returns true if BuiltInFunctionsEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsBuiltInFunctionsFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.BuiltInFunctionsEnableEpoch
}

// IsRelayedTransactionsFlagEnabledInEpoch returns true if RelayedTransactionsEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsRelayedTransactionsFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.RelayedTransactionsEnableEpoch
}

// IsPenalizedTooMuchGasFlagEnabledInEpoch returns true if PenalizedTooMuchGasEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsPenalizedTooMuchGasFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.PenalizedTooMuchGasEnableEpoch
}

// IsSwitchJailWaitingFlagEnabledInEpoch returns true if SwitchJailWaitingEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsSwitchJailWaitingFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.SwitchJailWaitingEnableEpoch
}

// IsBelowSignedThresholdFlagEnabledInEpoch returns true if BelowSignedThresholdEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsBelowSignedThresholdFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.BelowSignedThresholdEnableEpoch
}

// IsSwitchHysteresisForMinNodesFlagEnabledInSpecificEpochOnly returns true if SwitchHysteresisForMinNodesEnableEpoch is the provided epoch
func (handler *enableEpochsHandler) IsSwitchHysteresisForMinNodesFlagEnabledInSpecificEpochOnly(epoch uint32) bool {
	return epoch == handler.enableEpochsConfig.SwitchHysteresisForMinNodesEnableEpoch
}

// IsTransactionSignedWithTxHashFlagEnabledInEpoch returns true if TransactionSignedWithTxHashEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsTransactionSignedWithTxHashFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.TransactionSignedWithTxHashEnableEpoch
}

// IsMetaProtectionFlagEnabledInEpoch returns true if MetaProtectionEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsMetaProtectionFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.MetaProtectionEnableEpoch
}

// IsAheadOfTimeGasUsageFlagEnabledInEpoch returns true if AheadOfTimeGasUsageEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsAheadOfTimeGasUsageFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.AheadOfTimeGasUsageEnableEpoch
}

// IsGasPriceModifierFlagEnabledInEpoch returns true if GasPriceModifierEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsGasPriceModifierFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.GasPriceModifierEnableEpoch
}

// IsRepairCallbackFlagEnabledInEpoch returns true if RepairCallbackEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsRepairCallbackFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.RepairCallbackEnableEpoch
}

// IsReturnDataToLastTransferFlagEnabledAfterEpoch returns true if ReturnDataToLastTransferEnableEpoch is lower or equal with the provided epoch
func (handler *enableEpochsHandler) IsReturnDataToLastTransferFlagEnabledAfterEpoch(epoch uint32) bool {
	return epoch > handler.enableEpochsConfig.ReturnDataToLastTransferEnableEpoch
}

// IsSenderInOutTransferFlagEnabledInEpoch returns true if SenderInOutTransferEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsSenderInOutTransferFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.SenderInOutTransferEnableEpoch
}

// IsStakeFlagEnabledInEpoch returns true if StakeEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsStakeFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.StakeEnableEpoch
}

// IsStakingV2FlagEnabledInEpoch returns true if StakingV2EnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsStakingV2FlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.StakingV2EnableEpoch
}

// IsStakingV2OwnerFlagEnabledInSpecificEpochOnly returns true if StakingV2EnableEpoch is the provided epoch
func (handler *enableEpochsHandler) IsStakingV2OwnerFlagEnabledInSpecificEpochOnly(epoch uint32) bool {
	return epoch == handler.enableEpochsConfig.StakingV2EnableEpoch
}

// IsStakingV2FlagEnabledAfterEpoch returns true if StakingV2EnableEpoch is lower or equal with the provided epoch
func (handler *enableEpochsHandler) IsStakingV2FlagEnabledAfterEpoch(epoch uint32) bool {
	return epoch > handler.enableEpochsConfig.StakingV2EnableEpoch
}

// IsDoubleKeyProtectionFlagEnabledInEpoch returns true if DoubleKeyProtectionEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsDoubleKeyProtectionFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.DoubleKeyProtectionEnableEpoch
}

// IsESDTFlagEnabledInEpoch returns true if ESDTEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsESDTFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.ESDTEnableEpoch
}

// IsESDTFlagEnabledInSpecificEpochOnly returns true if ESDTEnableEpoch is the provided epoch
func (handler *enableEpochsHandler) IsESDTFlagEnabledInSpecificEpochOnly(epoch uint32) bool {
	return epoch == handler.enableEpochsConfig.ESDTEnableEpoch
}

// IsGovernanceFlagEnabledInEpoch returns true if GovernanceEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsGovernanceFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.GovernanceEnableEpoch
}

// IsGovernanceFlagEnabledInSpecificEpochOnly returns true if GovernanceEnableEpoch is the provided epoch
func (handler *enableEpochsHandler) IsGovernanceFlagEnabledInSpecificEpochOnly(epoch uint32) bool {
	return epoch == handler.enableEpochsConfig.GovernanceEnableEpoch
}

// IsDelegationManagerFlagEnabledInEpoch returns true if DelegationManagerEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsDelegationManagerFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.DelegationManagerEnableEpoch
}

// IsDelegationSmartContractFlagEnabledInEpoch returns true if DelegationSmartContractEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsDelegationSmartContractFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.DelegationSmartContractEnableEpoch
}

// IsDelegationSmartContractFlagEnabledInSpecificEpochOnly returns true if DelegationSmartContractEnableEpoch is the provided epoch
func (handler *enableEpochsHandler) IsDelegationSmartContractFlagEnabledInSpecificEpochOnly(epoch uint32) bool {
	return epoch == handler.enableEpochsConfig.DelegationSmartContractEnableEpoch
}

// IsCorrectLastUnJailedFlagEnabledInEpoch returns true if CorrectLastUnjailedEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsCorrectLastUnJailedFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch
}

// IsCorrectLastUnJailedFlagEnabledInSpecificEpochOnly returns true if CorrectLastUnjailedEnableEpoch is the provided epoch
func (handler *enableEpochsHandler) IsCorrectLastUnJailedFlagEnabledInSpecificEpochOnly(epoch uint32) bool {
	return epoch == handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch
}

// IsRelayedTransactionsV2FlagEnabledInEpoch returns true if RelayedTransactionsV2EnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsRelayedTransactionsV2FlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.RelayedTransactionsV2EnableEpoch
}

// IsUnBondTokensV2FlagEnabledInEpoch returns true if UnbondTokensV2EnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsUnBondTokensV2FlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.UnbondTokensV2EnableEpoch
}

// IsSaveJailedAlwaysFlagEnabledInEpoch returns true if SaveJailedAlwaysEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsSaveJailedAlwaysFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.SaveJailedAlwaysEnableEpoch
}

// IsReDelegateBelowMinCheckFlagEnabledInEpoch returns true if ReDelegateBelowMinCheckEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsReDelegateBelowMinCheckFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.ReDelegateBelowMinCheckEnableEpoch
}

// IsValidatorToDelegationFlagEnabledInEpoch returns true if ValidatorToDelegationEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsValidatorToDelegationFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.ValidatorToDelegationEnableEpoch
}

// IsWaitingListFixFlagEnabledInEpoch returns true if WaitingListFixEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsWaitingListFixFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.WaitingListFixEnableEpoch
}

// IsIncrementSCRNonceInMultiTransferFlagEnabledInEpoch returns true if IncrementSCRNonceInMultiTransferEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsIncrementSCRNonceInMultiTransferFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.IncrementSCRNonceInMultiTransferEnableEpoch
}

// IsESDTMultiTransferFlagEnabledInEpoch returns true if ESDTMultiTransferEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsESDTMultiTransferFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.ESDTMultiTransferEnableEpoch
}

// IsGlobalMintBurnFlagEnabledInEpoch returns true if the provided epoch is lower than GlobalMintBurnDisableEpoch
func (handler *enableEpochsHandler) IsGlobalMintBurnFlagEnabledInEpoch(epoch uint32) bool {
	return epoch < handler.enableEpochsConfig.GlobalMintBurnDisableEpoch
}

// IsESDTTransferRoleFlagEnabledInEpoch returns true if ESDTTransferRoleEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsESDTTransferRoleFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.ESDTTransferRoleEnableEpoch
}

// IsBuiltInFunctionOnMetaFlagEnabledInEpoch returns true if BuiltInFunctionOnMetaEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsBuiltInFunctionOnMetaFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.BuiltInFunctionOnMetaEnableEpoch
}

// IsComputeRewardCheckpointFlagEnabledInEpoch returns true if ComputeRewardCheckpointEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsComputeRewardCheckpointFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.ComputeRewardCheckpointEnableEpoch
}

// IsSCRSizeInvariantCheckFlagEnabledInEpoch returns true if SCRSizeInvariantCheckEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsSCRSizeInvariantCheckFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.SCRSizeInvariantCheckEnableEpoch
}

// IsBackwardCompSaveKeyValueFlagEnabledInEpoch returns true if the provided epoch is lower than BackwardCompSaveKeyValueEnableEpoch
func (handler *enableEpochsHandler) IsBackwardCompSaveKeyValueFlagEnabledInEpoch(epoch uint32) bool {
	return epoch < handler.enableEpochsConfig.BackwardCompSaveKeyValueEnableEpoch
}

// IsESDTNFTCreateOnMultiShardFlagEnabledInEpoch returns true if ESDTNFTCreateOnMultiShardEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsESDTNFTCreateOnMultiShardFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.ESDTNFTCreateOnMultiShardEnableEpoch
}

// IsMetaESDTSetFlagEnabledInEpoch returns true if MetaESDTSetEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsMetaESDTSetFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.MetaESDTSetEnableEpoch
}

// IsAddTokensToDelegationFlagEnabledInEpoch returns true if AddTokensToDelegationEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsAddTokensToDelegationFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.AddTokensToDelegationEnableEpoch
}

// IsMultiESDTTransferFixOnCallBackFlagEnabledInEpoch returns true if MultiESDTTransferFixOnCallBackOnEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsMultiESDTTransferFixOnCallBackFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.MultiESDTTransferFixOnCallBackOnEnableEpoch
}

// IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledInEpoch returns true if OptimizeGasUsedInCrossMiniBlocksEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsOptimizeGasUsedInCrossMiniBlocksFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.OptimizeGasUsedInCrossMiniBlocksEnableEpoch
}

// IsCorrectFirstQueuedFlagEnabledInEpoch returns true if CorrectFirstQueuedEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsCorrectFirstQueuedFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.CorrectFirstQueuedEpoch
}

// IsDeleteDelegatorAfterClaimRewardsFlagEnabledInEpoch returns true if DeleteDelegatorAfterClaimRewardsEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsDeleteDelegatorAfterClaimRewardsFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.DeleteDelegatorAfterClaimRewardsEnableEpoch
}

// IsFixOOGReturnCodeFlagEnabledInEpoch returns true if FixOOGReturnCodeEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsFixOOGReturnCodeFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.FixOOGReturnCodeEnableEpoch
}

// IsRemoveNonUpdatedStorageFlagEnabledInEpoch returns true if RemoveNonUpdatedStorageEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsRemoveNonUpdatedStorageFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.RemoveNonUpdatedStorageEnableEpoch
}

// IsOptimizeNFTStoreFlagEnabledInEpoch returns true if OptimizeNFTStoreEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsOptimizeNFTStoreFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch
}

// IsCreateNFTThroughExecByCallerFlagEnabledInEpoch returns true if CreateNFTThroughExecByCallerEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsCreateNFTThroughExecByCallerFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.CreateNFTThroughExecByCallerEnableEpoch
}

// IsStopDecreasingValidatorRatingWhenStuckFlagEnabledInEpoch returns true if StopDecreasingValidatorRatingWhenStuckEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsStopDecreasingValidatorRatingWhenStuckFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.StopDecreasingValidatorRatingWhenStuckEnableEpoch
}

// IsFrontRunningProtectionFlagEnabledInEpoch returns true if FrontRunningProtectionEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsFrontRunningProtectionFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.FrontRunningProtectionEnableEpoch
}

// IsPayableBySCFlagEnabledInEpoch returns true if IsPayableBySCEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsPayableBySCFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.IsPayableBySCEnableEpoch
}

// IsCleanUpInformativeSCRsFlagEnabledInEpoch returns true if CleanUpInformativeSCRsEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsCleanUpInformativeSCRsFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.CleanUpInformativeSCRsEnableEpoch
}

// IsStorageAPICostOptimizationFlagEnabledInEpoch returns true if StorageAPICostOptimizationEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsStorageAPICostOptimizationFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.StorageAPICostOptimizationEnableEpoch
}

// IsESDTRegisterAndSetAllRolesFlagEnabledInEpoch returns true if ESDTRegisterAndSetAllRolesEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsESDTRegisterAndSetAllRolesFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.ESDTRegisterAndSetAllRolesEnableEpoch
}

// IsScheduledMiniBlocksFlagEnabledInEpoch returns true if ScheduledMiniBlocksEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsScheduledMiniBlocksFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.ScheduledMiniBlocksEnableEpoch
}

// IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledInEpoch returns true if CorrectJailedNotUnstakedEmptyQueueEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsCorrectJailedNotUnStakedEmptyQueueFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.CorrectJailedNotUnstakedEmptyQueueEpoch
}

// IsDoNotReturnOldBlockInBlockchainHookFlagEnabledInEpoch returns true if DoNotReturnOldBlockInBlockchainHookEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsDoNotReturnOldBlockInBlockchainHookFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.DoNotReturnOldBlockInBlockchainHookEnableEpoch
}

// IsAddFailedRelayedTxToInvalidMBsFlagEnabledInEpoch returns true if the provided epoch is lower than AddFailedRelayedTxToInvalidMBsDisableEpoch
func (handler *enableEpochsHandler) IsAddFailedRelayedTxToInvalidMBsFlagEnabledInEpoch(epoch uint32) bool {
	return epoch < handler.enableEpochsConfig.AddFailedRelayedTxToInvalidMBsDisableEpoch
}

// IsSCRSizeInvariantOnBuiltInResultFlagEnabledInEpoch returns true if SCRSizeInvariantOnBuiltInResultEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsSCRSizeInvariantOnBuiltInResultFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.SCRSizeInvariantOnBuiltInResultEnableEpoch
}

// IsCheckCorrectTokenIDForTransferRoleFlagEnabledInEpoch returns true if CheckCorrectTokenIDForTransferRoleEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsCheckCorrectTokenIDForTransferRoleFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.CheckCorrectTokenIDForTransferRoleEnableEpoch
}

// IsFailExecutionOnEveryAPIErrorFlagEnabledInEpoch returns true if FailExecutionOnEveryAPIErrorEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsFailExecutionOnEveryAPIErrorFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.FailExecutionOnEveryAPIErrorEnableEpoch
}

// IsMiniBlockPartialExecutionFlagEnabledInEpoch returns true if MiniBlockPartialExecutionEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsMiniBlockPartialExecutionFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.MiniBlockPartialExecutionEnableEpoch
}

// IsManagedCryptoAPIsFlagEnabledInEpoch returns true if ManagedCryptoAPIsEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsManagedCryptoAPIsFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.ManagedCryptoAPIsEnableEpoch
}

// IsESDTMetadataContinuousCleanupFlagEnabledInEpoch returns true if ESDTMetadataContinuousCleanupEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsESDTMetadataContinuousCleanupFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch
}

// IsDisableExecByCallerFlagEnabledInEpoch returns true if DisableExecByCallerEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsDisableExecByCallerFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.DisableExecByCallerEnableEpoch
}

// IsRefactorContextFlagEnabledInEpoch returns true if RefactorContextEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsRefactorContextFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.RefactorContextEnableEpoch
}

// IsCheckFunctionArgumentFlagEnabledInEpoch returns true if CheckFunctionArgumentEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsCheckFunctionArgumentFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.CheckFunctionArgumentEnableEpoch
}

// IsCheckExecuteOnReadOnlyFlagEnabledInEpoch returns true if CheckExecuteOnReadOnlyEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsCheckExecuteOnReadOnlyFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.CheckExecuteOnReadOnlyEnableEpoch
}

// IsSetSenderInEeiOutputTransferFlagEnabledInEpoch returns true if SetSenderInEeiOutputTransferEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsSetSenderInEeiOutputTransferFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.SetSenderInEeiOutputTransferEnableEpoch
}

// IsFixAsyncCallbackCheckFlagEnabledInEpoch returns true if ESDTMetadataContinuousCleanupEnableEpoch is lower than the provided epoch
// this is a duplicate for ESDTMetadataContinuousCleanupEnableEpoch needed for consistency into vm-common
func (handler *enableEpochsHandler) IsFixAsyncCallbackCheckFlagEnabledInEpoch(epoch uint32) bool {
	return handler.IsESDTMetadataContinuousCleanupFlagEnabledInEpoch(epoch)
}

// IsSaveToSystemAccountFlagEnabledInEpoch returns true if OptimizeNFTStoreEnableEpoch is lower than the provided epoch
// this is a duplicate for OptimizeNFTStoreEnableEpoch needed for consistency into vm-common
func (handler *enableEpochsHandler) IsSaveToSystemAccountFlagEnabledInEpoch(epoch uint32) bool {
	return handler.IsOptimizeNFTStoreFlagEnabledInEpoch(epoch)
}

// IsCheckFrozenCollectionFlagEnabledInEpoch returns true if OptimizeNFTStoreEnableEpoch is lower than the provided epoch
// this is a duplicate for OptimizeNFTStoreEnableEpoch needed for consistency into vm-common
func (handler *enableEpochsHandler) IsCheckFrozenCollectionFlagEnabledInEpoch(epoch uint32) bool {
	return handler.IsOptimizeNFTStoreFlagEnabledInEpoch(epoch)
}

// IsSendAlwaysFlagEnabledInEpoch returns true if ESDTMetadataContinuousCleanupEnableEpoch is lower than the provided epoch
// this is a duplicate for ESDTMetadataContinuousCleanupEnableEpoch needed for consistency into vm-common
func (handler *enableEpochsHandler) IsSendAlwaysFlagEnabledInEpoch(epoch uint32) bool {
	return handler.IsESDTMetadataContinuousCleanupFlagEnabledInEpoch(epoch)
}

// IsValueLengthCheckFlagEnabledInEpoch returns true if OptimizeNFTStoreEnableEpoch is lower than the provided epoch
// this is a duplicate for OptimizeNFTStoreEnableEpoch needed for consistency into vm-common
func (handler *enableEpochsHandler) IsValueLengthCheckFlagEnabledInEpoch(epoch uint32) bool {
	return handler.IsOptimizeNFTStoreFlagEnabledInEpoch(epoch)
}

// IsCheckTransferFlagEnabledInEpoch returns true if OptimizeNFTStoreEnableEpoch is lower than the provided epoch
// this is a duplicate for OptimizeNFTStoreEnableEpoch needed for consistency into vm-common
func (handler *enableEpochsHandler) IsCheckTransferFlagEnabledInEpoch(epoch uint32) bool {
	return handler.IsOptimizeNFTStoreFlagEnabledInEpoch(epoch)
}

// IsTransferToMetaFlagEnabledInEpoch returns true if BuiltInFunctionOnMetaEnableEpoch is lower than the provided epoch
// this is a duplicate for BuiltInFunctionOnMetaEnableEpoch needed for consistency into vm-common
func (handler *enableEpochsHandler) IsTransferToMetaFlagEnabledInEpoch(epoch uint32) bool {
	return handler.IsBuiltInFunctionOnMetaFlagEnabledInEpoch(epoch)
}

// IsESDTNFTImprovementV1FlagEnabledInEpoch returns true if ESDTMultiTransferEnableEpoch is lower than the provided epoch
// this is a duplicate for ESDTMultiTransferEnableEpoch needed for consistency into vm-common
func (handler *enableEpochsHandler) IsESDTNFTImprovementV1FlagEnabledInEpoch(epoch uint32) bool {
	return handler.IsESDTMultiTransferFlagEnabledInEpoch(epoch)
}

// IsChangeDelegationOwnerFlagEnabledInEpoch returns true if the change delegation owner feature is enabled
// this is a duplicate for ESDTMetadataContinuousCleanupEnableEpoch needed for consistency into vm-common
func (handler *enableEpochsHandler) IsChangeDelegationOwnerFlagEnabledInEpoch(epoch uint32) bool {
	return handler.IsESDTMetadataContinuousCleanupFlagEnabledInEpoch(epoch)
}

// IsRefactorPeersMiniBlocksFlagEnabledInEpoch returns true if RefactorPeersMiniBlocksEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsRefactorPeersMiniBlocksFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.RefactorPeersMiniBlocksEnableEpoch
}

// IsSCProcessorV2FlagEnabledInEpoch returns true if SCProcessorV2EnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsSCProcessorV2FlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.SCProcessorV2EnableEpoch
}

// IsFixAsyncCallBackArgsListFlagEnabledInEpoch returns true if FixAsyncCallBackArgsListEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsFixAsyncCallBackArgsListFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.FixAsyncCallBackArgsListEnableEpoch
}

// IsFixOldTokenLiquidityEnabledInEpoch returns true if FixOldTokenLiquidityEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsFixOldTokenLiquidityEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.FixOldTokenLiquidityEnableEpoch
}

// IsRuntimeMemStoreLimitEnabledInEpoch returns true if RuntimeMemStoreLimitEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsRuntimeMemStoreLimitEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.RuntimeMemStoreLimitEnableEpoch
}

// IsRuntimeCodeSizeFixEnabledInEpoch returns true if RuntimeCodeSizeFixEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsRuntimeCodeSizeFixEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.RuntimeCodeSizeFixEnableEpoch
}

// IsMaxBlockchainHookCountersFlagEnabledInEpoch returns true if MaxBlockchainHookCountersEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsMaxBlockchainHookCountersFlagEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.MaxBlockchainHookCountersEnableEpoch
}

// IsWipeSingleNFTLiquidityDecreaseEnabledInEpoch returns true if WipeSingleNFTLiquidityDecreaseEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsWipeSingleNFTLiquidityDecreaseEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.WipeSingleNFTLiquidityDecreaseEnableEpoch
}

// IsAlwaysSaveTokenMetaDataEnabledInEpoch returns true if AlwaysSaveTokenMetaDataEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsAlwaysSaveTokenMetaDataEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.AlwaysSaveTokenMetaDataEnableEpoch
}

// IsSetGuardianEnabledInEpoch returns true if setGuardianFlag is lower than the provided epoch
func (handler *enableEpochsHandler) IsSetGuardianEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.SetGuardianEnableEpoch
}

// IsRelayedNonceFixEnabledInEpoch returns true if RelayedNonceFixEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsRelayedNonceFixEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.RelayedNonceFixEnableEpoch
}

// IsConsistentTokensValuesLengthCheckEnabledInEpoch returns true if ConsistentTokensValuesLengthCheckEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsConsistentTokensValuesLengthCheckEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.ConsistentTokensValuesLengthCheckEnableEpoch
}

// IsKeepExecOrderOnCreatedSCRsEnabledInEpoch returns true if KeepExecOrderOnCreatedSCRsEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsKeepExecOrderOnCreatedSCRsEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.KeepExecOrderOnCreatedSCRsEnableEpoch
}

// IsMultiClaimOnDelegationEnabledInEpoch returns true if multi claim on delegation is enabled
func (handler *enableEpochsHandler) IsMultiClaimOnDelegationEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.MultiClaimOnDelegationEnableEpoch
}

// IsChangeUsernameEnabledInEpoch returns true if ChangeUsernameEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsChangeUsernameEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.ChangeUsernameEnableEpoch
}

// IsAutoBalanceDataTriesEnabledInEpoch returns true if AutoBalanceDataTriesEnableEpoch is lower than the provided epoch
func (handler *enableEpochsHandler) IsAutoBalanceDataTriesEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.AutoBalanceDataTriesEnableEpoch
}

// FixDelegationChangeOwnerOnAccountEnabledInEpoch returns true if the fix for the delegation change owner on account is enabled
func (handler *enableEpochsHandler) FixDelegationChangeOwnerOnAccountEnabledInEpoch(epoch uint32) bool {
	return epoch >= handler.enableEpochsConfig.FixDelegationChangeOwnerOnAccountEnableEpoch
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *enableEpochsHandler) IsInterfaceNil() bool {
	return handler == nil
}
