package enablers

import (
	"runtime/debug"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("common/enablers")

type flagEnabledInEpoch = func(epoch uint32) bool

type flagHandler struct {
	isActiveInEpoch flagEnabledInEpoch
	activationEpoch uint32
}

type enableEpochsHandler struct {
	allFlagsDefined    map[core.EnableEpochFlag]flagHandler
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
		enableEpochsConfig: enableEpochsConfig,
	}

	handler.createAllFlagsMap()

	epochNotifier.RegisterNotifyHandler(handler)

	return handler, nil
}

func (handler *enableEpochsHandler) createAllFlagsMap() {
	handler.allFlagsDefined = map[core.EnableEpochFlag]flagHandler{
		common.SCDeployFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SCDeployEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.SCDeployEnableEpoch,
		},
		common.BuiltInFunctionsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.BuiltInFunctionsEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.BuiltInFunctionsEnableEpoch,
		},
		common.RelayedTransactionsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RelayedTransactionsEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.RelayedTransactionsEnableEpoch,
		},
		common.PenalizedTooMuchGasFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.PenalizedTooMuchGasEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.PenalizedTooMuchGasEnableEpoch,
		},
		common.SwitchJailWaitingFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SwitchJailWaitingEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.SwitchJailWaitingEnableEpoch,
		},
		common.BelowSignedThresholdFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.BelowSignedThresholdEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.BelowSignedThresholdEnableEpoch,
		},
		common.SwitchHysteresisForMinNodesFlagInSpecificEpochOnly: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch == handler.enableEpochsConfig.SwitchHysteresisForMinNodesEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.SwitchHysteresisForMinNodesEnableEpoch,
		},
		common.TransactionSignedWithTxHashFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.TransactionSignedWithTxHashEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.TransactionSignedWithTxHashEnableEpoch,
		},
		common.MetaProtectionFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MetaProtectionEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.MetaProtectionEnableEpoch,
		},
		common.AheadOfTimeGasUsageFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.AheadOfTimeGasUsageEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.AheadOfTimeGasUsageEnableEpoch,
		},
		common.GasPriceModifierFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.GasPriceModifierEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.GasPriceModifierEnableEpoch,
		},
		common.RepairCallbackFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RepairCallbackEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.RepairCallbackEnableEpoch,
		},
		common.ReturnDataToLastTransferFlagAfterEpoch: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch > handler.enableEpochsConfig.ReturnDataToLastTransferEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ReturnDataToLastTransferEnableEpoch,
		},
		common.SenderInOutTransferFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SenderInOutTransferEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.SenderInOutTransferEnableEpoch,
		},
		common.StakeFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.StakeEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.StakeEnableEpoch,
		},
		common.StakingV2Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.StakingV2EnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.StakingV2EnableEpoch,
		},
		common.StakingV2OwnerFlagInSpecificEpochOnly: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch == handler.enableEpochsConfig.StakingV2EnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.StakingV2EnableEpoch,
		},
		common.StakingV2FlagAfterEpoch: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch > handler.enableEpochsConfig.StakingV2EnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.StakingV2EnableEpoch,
		},
		common.DoubleKeyProtectionFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DoubleKeyProtectionEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.DoubleKeyProtectionEnableEpoch,
		},
		common.ESDTFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ESDTEnableEpoch,
		},
		common.ESDTFlagInSpecificEpochOnly: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch == handler.enableEpochsConfig.ESDTEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ESDTEnableEpoch,
		},
		common.GovernanceFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.GovernanceEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.GovernanceEnableEpoch,
		},
		common.GovernanceFlagInSpecificEpochOnly: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch == handler.enableEpochsConfig.GovernanceEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.GovernanceEnableEpoch,
		},
		common.DelegationManagerFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DelegationManagerEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.DelegationManagerEnableEpoch,
		},
		common.DelegationSmartContractFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DelegationSmartContractEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.DelegationSmartContractEnableEpoch,
		},
		common.DelegationSmartContractFlagInSpecificEpochOnly: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch == handler.enableEpochsConfig.DelegationSmartContractEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.DelegationSmartContractEnableEpoch,
		},
		common.CorrectLastUnJailedFlagInSpecificEpochOnly: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch == handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch,
		},
		common.CorrectLastUnJailedFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch,
		},
		common.RelayedTransactionsV2Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RelayedTransactionsV2EnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.RelayedTransactionsV2EnableEpoch,
		},
		common.UnBondTokensV2Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.UnbondTokensV2EnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.UnbondTokensV2EnableEpoch,
		},
		common.SaveJailedAlwaysFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SaveJailedAlwaysEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.SaveJailedAlwaysEnableEpoch,
		},
		common.ReDelegateBelowMinCheckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ReDelegateBelowMinCheckEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ReDelegateBelowMinCheckEnableEpoch,
		},
		common.ValidatorToDelegationFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ValidatorToDelegationEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ValidatorToDelegationEnableEpoch,
		},
		common.IncrementSCRNonceInMultiTransferFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.IncrementSCRNonceInMultiTransferEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.IncrementSCRNonceInMultiTransferEnableEpoch,
		},
		common.ESDTMultiTransferFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTMultiTransferEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ESDTMultiTransferEnableEpoch,
		},
		common.ESDTNFTImprovementV1Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTMultiTransferEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ESDTMultiTransferEnableEpoch,
		},
		common.GlobalMintBurnFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch < handler.enableEpochsConfig.GlobalMintBurnDisableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.GlobalMintBurnDisableEpoch,
		},
		common.ESDTTransferRoleFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTTransferRoleEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ESDTTransferRoleEnableEpoch,
		},
		common.BuiltInFunctionOnMetaFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.BuiltInFunctionOnMetaEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.BuiltInFunctionOnMetaEnableEpoch,
		},
		common.TransferToMetaFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.BuiltInFunctionOnMetaEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.BuiltInFunctionOnMetaEnableEpoch,
		},
		common.ComputeRewardCheckpointFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ComputeRewardCheckpointEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ComputeRewardCheckpointEnableEpoch,
		},
		common.SCRSizeInvariantCheckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SCRSizeInvariantCheckEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.SCRSizeInvariantCheckEnableEpoch,
		},
		common.BackwardCompSaveKeyValueFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch < handler.enableEpochsConfig.BackwardCompSaveKeyValueEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.BackwardCompSaveKeyValueEnableEpoch,
		},
		common.ESDTNFTCreateOnMultiShardFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTNFTCreateOnMultiShardEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ESDTNFTCreateOnMultiShardEnableEpoch,
		},
		common.MetaESDTSetFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MetaESDTSetEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.MetaESDTSetEnableEpoch,
		},
		common.AddTokensToDelegationFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.AddTokensToDelegationEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.AddTokensToDelegationEnableEpoch,
		},
		common.MultiESDTTransferFixOnCallBackFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MultiESDTTransferFixOnCallBackOnEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.MultiESDTTransferFixOnCallBackOnEnableEpoch,
		},
		common.OptimizeGasUsedInCrossMiniBlocksFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.OptimizeGasUsedInCrossMiniBlocksEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.OptimizeGasUsedInCrossMiniBlocksEnableEpoch,
		},
		common.CorrectFirstQueuedFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CorrectFirstQueuedEpoch
			},
			activationEpoch: handler.enableEpochsConfig.CorrectFirstQueuedEpoch,
		},
		common.DeleteDelegatorAfterClaimRewardsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DeleteDelegatorAfterClaimRewardsEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.DeleteDelegatorAfterClaimRewardsEnableEpoch,
		},
		common.RemoveNonUpdatedStorageFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RemoveNonUpdatedStorageEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.RemoveNonUpdatedStorageEnableEpoch,
		},
		common.OptimizeNFTStoreFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch,
		},
		common.SaveToSystemAccountFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch,
		},
		common.CheckFrozenCollectionFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch,
		},
		common.ValueLengthCheckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch,
		},
		common.CheckTransferFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch,
		},
		common.CreateNFTThroughExecByCallerFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CreateNFTThroughExecByCallerEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.CreateNFTThroughExecByCallerEnableEpoch,
		},
		common.StopDecreasingValidatorRatingWhenStuckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.StopDecreasingValidatorRatingWhenStuckEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.StopDecreasingValidatorRatingWhenStuckEnableEpoch,
		},
		common.FrontRunningProtectionFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FrontRunningProtectionEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.FrontRunningProtectionEnableEpoch,
		},
		common.PayableBySCFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.IsPayableBySCEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.IsPayableBySCEnableEpoch,
		},
		common.CleanUpInformativeSCRsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CleanUpInformativeSCRsEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.CleanUpInformativeSCRsEnableEpoch,
		},
		common.StorageAPICostOptimizationFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.StorageAPICostOptimizationEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.StorageAPICostOptimizationEnableEpoch,
		},
		common.ESDTRegisterAndSetAllRolesFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTRegisterAndSetAllRolesEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ESDTRegisterAndSetAllRolesEnableEpoch,
		},
		common.ScheduledMiniBlocksFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ScheduledMiniBlocksEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ScheduledMiniBlocksEnableEpoch,
		},
		common.CorrectJailedNotUnStakedEmptyQueueFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CorrectJailedNotUnstakedEmptyQueueEpoch
			},
			activationEpoch: handler.enableEpochsConfig.CorrectJailedNotUnstakedEmptyQueueEpoch,
		},
		common.DoNotReturnOldBlockInBlockchainHookFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DoNotReturnOldBlockInBlockchainHookEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.DoNotReturnOldBlockInBlockchainHookEnableEpoch,
		},
		common.AddFailedRelayedTxToInvalidMBsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch < handler.enableEpochsConfig.AddFailedRelayedTxToInvalidMBsDisableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.AddFailedRelayedTxToInvalidMBsDisableEpoch,
		},
		common.SCRSizeInvariantOnBuiltInResultFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SCRSizeInvariantOnBuiltInResultEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.SCRSizeInvariantOnBuiltInResultEnableEpoch,
		},
		common.CheckCorrectTokenIDForTransferRoleFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CheckCorrectTokenIDForTransferRoleEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.CheckCorrectTokenIDForTransferRoleEnableEpoch,
		},
		common.FailExecutionOnEveryAPIErrorFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FailExecutionOnEveryAPIErrorEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.FailExecutionOnEveryAPIErrorEnableEpoch,
		},
		common.MiniBlockPartialExecutionFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MiniBlockPartialExecutionEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.MiniBlockPartialExecutionEnableEpoch,
		},
		common.ManagedCryptoAPIsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ManagedCryptoAPIsEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ManagedCryptoAPIsEnableEpoch,
		},
		common.ESDTMetadataContinuousCleanupFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch,
		},
		common.FixAsyncCallbackCheckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch,
		},
		common.SendAlwaysFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch,
		},
		common.ChangeDelegationOwnerFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch,
		},
		common.DisableExecByCallerFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DisableExecByCallerEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.DisableExecByCallerEnableEpoch,
		},
		common.RefactorContextFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RefactorContextEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.RefactorContextEnableEpoch,
		},
		common.CheckFunctionArgumentFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CheckFunctionArgumentEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.CheckFunctionArgumentEnableEpoch,
		},
		common.CheckExecuteOnReadOnlyFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CheckExecuteOnReadOnlyEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.CheckExecuteOnReadOnlyEnableEpoch,
		},
		common.SetSenderInEeiOutputTransferFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SetSenderInEeiOutputTransferEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.SetSenderInEeiOutputTransferEnableEpoch,
		},
		common.RefactorPeersMiniBlocksFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RefactorPeersMiniBlocksEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.RefactorPeersMiniBlocksEnableEpoch,
		},
		common.SCProcessorV2Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SCProcessorV2EnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.SCProcessorV2EnableEpoch,
		},
		common.FixAsyncCallBackArgsListFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixAsyncCallBackArgsListEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.FixAsyncCallBackArgsListEnableEpoch,
		},
		common.FixOldTokenLiquidityFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixOldTokenLiquidityEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.FixOldTokenLiquidityEnableEpoch,
		},
		common.RuntimeMemStoreLimitFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RuntimeMemStoreLimitEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.RuntimeMemStoreLimitEnableEpoch,
		},
		common.RuntimeCodeSizeFixFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RuntimeCodeSizeFixEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.RuntimeCodeSizeFixEnableEpoch,
		},
		common.MaxBlockchainHookCountersFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MaxBlockchainHookCountersEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.MaxBlockchainHookCountersEnableEpoch,
		},
		common.WipeSingleNFTLiquidityDecreaseFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.WipeSingleNFTLiquidityDecreaseEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.WipeSingleNFTLiquidityDecreaseEnableEpoch,
		},
		common.AlwaysSaveTokenMetaDataFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.AlwaysSaveTokenMetaDataEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.AlwaysSaveTokenMetaDataEnableEpoch,
		},
		common.SetGuardianFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SetGuardianEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.SetGuardianEnableEpoch,
		},
		common.RelayedNonceFixFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RelayedNonceFixEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.RelayedNonceFixEnableEpoch,
		},
		common.ConsistentTokensValuesLengthCheckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ConsistentTokensValuesLengthCheckEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ConsistentTokensValuesLengthCheckEnableEpoch,
		},
		common.KeepExecOrderOnCreatedSCRsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.KeepExecOrderOnCreatedSCRsEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.KeepExecOrderOnCreatedSCRsEnableEpoch,
		},
		common.MultiClaimOnDelegationFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MultiClaimOnDelegationEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.MultiClaimOnDelegationEnableEpoch,
		},
		common.ChangeUsernameFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ChangeUsernameEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ChangeUsernameEnableEpoch,
		},
		common.AutoBalanceDataTriesFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.AutoBalanceDataTriesEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.AutoBalanceDataTriesEnableEpoch,
		},
		common.MigrateDataTrieFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MigrateDataTrieEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.MigrateDataTrieEnableEpoch,
		},
		common.FixDelegationChangeOwnerOnAccountFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixDelegationChangeOwnerOnAccountEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.FixDelegationChangeOwnerOnAccountEnableEpoch,
		},
		common.FixOOGReturnCodeFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixOOGReturnCodeEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.FixOOGReturnCodeEnableEpoch,
		},
		common.DeterministicSortOnValidatorsInfoFixFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DeterministicSortOnValidatorsInfoEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.DeterministicSortOnValidatorsInfoEnableEpoch,
		},
		common.DynamicGasCostForDataTrieStorageLoadFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DynamicGasCostForDataTrieStorageLoadEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.DynamicGasCostForDataTrieStorageLoadEnableEpoch,
		},
		common.ScToScLogEventFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ScToScLogEventEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ScToScLogEventEnableEpoch,
		},
		common.BlockGasAndFeesReCheckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.BlockGasAndFeesReCheckEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.BlockGasAndFeesReCheckEnableEpoch,
		},
		common.BalanceWaitingListsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.BalanceWaitingListsEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.BalanceWaitingListsEnableEpoch,
		},
		common.WaitingListFixFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.WaitingListFixEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.WaitingListFixEnableEpoch,
		},
		common.NFTStopCreateFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.NFTStopCreateEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.NFTStopCreateEnableEpoch,
		},
		common.FixGasRemainingForSaveKeyValueFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixGasRemainingForSaveKeyValueBuiltinFunctionEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.FixGasRemainingForSaveKeyValueBuiltinFunctionEnableEpoch,
		},
		common.IsChangeOwnerAddressCrossShardThroughSCFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ChangeOwnerAddressCrossShardThroughSCEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.ChangeOwnerAddressCrossShardThroughSCEnableEpoch,
		},
		common.CurrentRandomnessOnSortingFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CurrentRandomnessOnSortingEnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.CurrentRandomnessOnSortingEnableEpoch,
		},
		common.CryptoAPIV2Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CryptoAPIV2EnableEpoch
			},
			activationEpoch: handler.enableEpochsConfig.CryptoAPIV2EnableEpoch,
		},
	}
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (handler *enableEpochsHandler) EpochConfirmed(epoch uint32, _ uint64) {
	handler.epochMut.Lock()
	handler.currentEpoch = epoch
	handler.epochMut.Unlock()
}

// IsFlagDefined checks if a specific flag is supported by the current version of mx-chain-core-go
func (handler *enableEpochsHandler) IsFlagDefined(flag core.EnableEpochFlag) bool {
	_, found := handler.allFlagsDefined[flag]
	if found {
		return true
	}

	log.Error("programming error, flag is not defined",
		"flag", flag,
		"stack trace", string(debug.Stack()))
	return false
}

// IsFlagEnabled returns true if the provided flag is enabled in the current epoch
func (handler *enableEpochsHandler) IsFlagEnabled(flag core.EnableEpochFlag) bool {
	handler.epochMut.RLock()
	currentEpoch := handler.currentEpoch
	handler.epochMut.RUnlock()

	return handler.IsFlagEnabledInEpoch(flag, currentEpoch)
}

// IsFlagEnabledInEpoch returns true if the provided flag is enabled in the provided epoch
func (handler *enableEpochsHandler) IsFlagEnabledInEpoch(flag core.EnableEpochFlag, epoch uint32) bool {
	fh, found := handler.allFlagsDefined[flag]
	if !found {
		log.Warn("IsFlagEnabledInEpoch: programming error, got unknown flag",
			"flag", flag,
			"epoch", epoch,
			"stack trace", string(debug.Stack()))
		return false
	}

	return fh.isActiveInEpoch(epoch)
}

// GetActivationEpoch returns the activation epoch of the provided flag
func (handler *enableEpochsHandler) GetActivationEpoch(flag core.EnableEpochFlag) uint32 {
	fh, found := handler.allFlagsDefined[flag]
	if !found {
		log.Warn("GetActivationEpoch: programming error, got unknown flag",
			"flag", flag,
			"stack trace", string(debug.Stack()))
		return 0
	}

	return fh.activationEpoch
}

// GetCurrentEpoch returns the current epoch
func (handler *enableEpochsHandler) GetCurrentEpoch() uint32 {
	handler.epochMut.RLock()
	currentEpoch := handler.currentEpoch
	handler.epochMut.RUnlock()

	return currentEpoch
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *enableEpochsHandler) IsInterfaceNil() bool {
	return handler == nil
}
