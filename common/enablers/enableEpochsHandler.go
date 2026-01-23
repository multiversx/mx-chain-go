package enablers

import (
	"runtime/debug"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

var log = logger.GetOrCreate("common/enablers")

type flagEnabledInEpoch = func(epoch uint32) bool

type flagHandler struct {
	isActiveInEpoch     flagEnabledInEpoch
	activationEpoch     uint32
	activationEpochName string
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
			activationEpoch:     handler.enableEpochsConfig.SCDeployEnableEpoch,
			activationEpochName: "SCDeployEnableEpoch",
		},
		common.BuiltInFunctionsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.BuiltInFunctionsEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.BuiltInFunctionsEnableEpoch,
			activationEpochName: "BuiltInFunctionsEnableEpoch",
		},
		common.RelayedTransactionsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RelayedTransactionsEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.RelayedTransactionsEnableEpoch,
			activationEpochName: "RelayedTransactionsEnableEpoch",
		},
		common.PenalizedTooMuchGasFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.PenalizedTooMuchGasEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.PenalizedTooMuchGasEnableEpoch,
			activationEpochName: "PenalizedTooMuchGasEnableEpoch",
		},
		common.SwitchJailWaitingFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SwitchJailWaitingEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.SwitchJailWaitingEnableEpoch,
			activationEpochName: "SwitchJailWaitingEnableEpoch",
		},
		common.BelowSignedThresholdFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.BelowSignedThresholdEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.BelowSignedThresholdEnableEpoch,
			activationEpochName: "BelowSignedThresholdEnableEpoch",
		},
		common.SwitchHysteresisForMinNodesFlagInSpecificEpochOnly: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch == handler.enableEpochsConfig.SwitchHysteresisForMinNodesEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.SwitchHysteresisForMinNodesEnableEpoch,
			activationEpochName: "SwitchHysteresisForMinNodesEnableEpoch",
		},
		common.TransactionSignedWithTxHashFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.TransactionSignedWithTxHashEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.TransactionSignedWithTxHashEnableEpoch,
			activationEpochName: "TransactionSignedWithTxHashEnableEpoch",
		},
		common.MetaProtectionFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MetaProtectionEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.MetaProtectionEnableEpoch,
			activationEpochName: "MetaProtectionEnableEpoch",
		},
		common.AheadOfTimeGasUsageFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.AheadOfTimeGasUsageEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.AheadOfTimeGasUsageEnableEpoch,
			activationEpochName: "AheadOfTimeGasUsageEnableEpoch",
		},
		common.GasPriceModifierFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.GasPriceModifierEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.GasPriceModifierEnableEpoch,
			activationEpochName: "GasPriceModifierEnableEpoch",
		},
		common.RepairCallbackFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RepairCallbackEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.RepairCallbackEnableEpoch,
			activationEpochName: "RepairCallbackEnableEpoch",
		},
		common.ReturnDataToLastTransferFlagAfterEpoch: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch > handler.enableEpochsConfig.ReturnDataToLastTransferEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ReturnDataToLastTransferEnableEpoch,
			activationEpochName: "ReturnDataToLastTransferEnableEpoch",
		},
		common.SenderInOutTransferFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SenderInOutTransferEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.SenderInOutTransferEnableEpoch,
			activationEpochName: "SenderInOutTransferEnableEpoch",
		},
		common.StakeFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.StakeEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.StakeEnableEpoch,
			activationEpochName: "StakeEnableEpoch",
		},
		common.StakingV2Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.StakingV2EnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.StakingV2EnableEpoch,
			activationEpochName: "StakingV2EnableEpoch",
		},
		common.StakingV2OwnerFlagInSpecificEpochOnly: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch == handler.enableEpochsConfig.StakingV2EnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.StakingV2EnableEpoch,
			activationEpochName: "StakingV2EnableEpoch",
		},
		common.StakingV2FlagAfterEpoch: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch > handler.enableEpochsConfig.StakingV2EnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.StakingV2EnableEpoch,
			activationEpochName: "StakingV2EnableEpoch",
		},
		common.DoubleKeyProtectionFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DoubleKeyProtectionEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.DoubleKeyProtectionEnableEpoch,
			activationEpochName: "DoubleKeyProtectionEnableEpoch",
		},
		common.ESDTFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ESDTEnableEpoch,
			activationEpochName: "ESDTEnableEpoch",
		},
		common.ESDTFlagInSpecificEpochOnly: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch == handler.enableEpochsConfig.ESDTEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ESDTEnableEpoch,
			activationEpochName: "ESDTEnableEpoch",
		},
		common.GovernanceFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.GovernanceEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.GovernanceEnableEpoch,
			activationEpochName: "GovernanceEnableEpoch",
		},
		common.GovernanceFlagInSpecificEpochOnly: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch == handler.enableEpochsConfig.GovernanceEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.GovernanceEnableEpoch,
			activationEpochName: "GovernanceEnableEpoch",
		},
		common.GovernanceDisableProposeFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.GovernanceDisableProposeEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.GovernanceDisableProposeEnableEpoch,
			activationEpochName: "GovernanceDisableProposeEnableEpoch",
		},
		common.GovernanceFixesFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.GovernanceFixesEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.GovernanceFixesEnableEpoch,
			activationEpochName: "GovernanceFixesEnableEpoch",
		},
		common.DelegationManagerFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DelegationManagerEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.DelegationManagerEnableEpoch,
			activationEpochName: "DelegationManagerEnableEpoch",
		},
		common.DelegationSmartContractFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DelegationSmartContractEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.DelegationSmartContractEnableEpoch,
			activationEpochName: "DelegationSmartContractEnableEpoch",
		},
		common.DelegationSmartContractFlagInSpecificEpochOnly: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch == handler.enableEpochsConfig.DelegationSmartContractEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.DelegationSmartContractEnableEpoch,
			activationEpochName: "DelegationSmartContractEnableEpoch",
		},
		common.CorrectLastUnJailedFlagInSpecificEpochOnly: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch == handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch,
			activationEpochName: "CorrectLastUnjailedEnableEpoch",
		},
		common.CorrectLastUnJailedFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.CorrectLastUnjailedEnableEpoch,
			activationEpochName: "CorrectLastUnjailedEnableEpoch",
		},
		common.RelayedTransactionsV2Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RelayedTransactionsV2EnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.RelayedTransactionsV2EnableEpoch,
			activationEpochName: "RelayedTransactionsV2EnableEpoch",
		},
		common.UnBondTokensV2Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.UnbondTokensV2EnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.UnbondTokensV2EnableEpoch,
			activationEpochName: "UnbondTokensV2EnableEpoch",
		},
		common.SaveJailedAlwaysFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SaveJailedAlwaysEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.SaveJailedAlwaysEnableEpoch,
			activationEpochName: "SaveJailedAlwaysEnableEpoch",
		},
		common.ReDelegateBelowMinCheckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ReDelegateBelowMinCheckEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ReDelegateBelowMinCheckEnableEpoch,
			activationEpochName: "ReDelegateBelowMinCheckEnableEpoch",
		},
		common.ValidatorToDelegationFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ValidatorToDelegationEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ValidatorToDelegationEnableEpoch,
			activationEpochName: "ValidatorToDelegationEnableEpoch",
		},
		common.IncrementSCRNonceInMultiTransferFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.IncrementSCRNonceInMultiTransferEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.IncrementSCRNonceInMultiTransferEnableEpoch,
			activationEpochName: "IncrementSCRNonceInMultiTransferEnableEpoch",
		},
		common.ESDTMultiTransferFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTMultiTransferEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ESDTMultiTransferEnableEpoch,
			activationEpochName: "ESDTMultiTransferEnableEpoch",
		},
		common.ESDTNFTImprovementV1Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTMultiTransferEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ESDTMultiTransferEnableEpoch,
			activationEpochName: "ESDTMultiTransferEnableEpoch",
		},
		common.GlobalMintBurnFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch < handler.enableEpochsConfig.GlobalMintBurnDisableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.GlobalMintBurnDisableEpoch,
			activationEpochName: "GlobalMintBurnDisableEpoch",
		},
		common.ESDTTransferRoleFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTTransferRoleEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ESDTTransferRoleEnableEpoch,
			activationEpochName: "ESDTTransferRoleEnableEpoch",
		},
		common.ComputeRewardCheckpointFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ComputeRewardCheckpointEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ComputeRewardCheckpointEnableEpoch,
			activationEpochName: "ComputeRewardCheckpointEnableEpoch",
		},
		common.SCRSizeInvariantCheckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SCRSizeInvariantCheckEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.SCRSizeInvariantCheckEnableEpoch,
			activationEpochName: "SCRSizeInvariantCheckEnableEpoch",
		},
		common.BackwardCompSaveKeyValueFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch < handler.enableEpochsConfig.BackwardCompSaveKeyValueEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.BackwardCompSaveKeyValueEnableEpoch,
			activationEpochName: "BackwardCompSaveKeyValueEnableEpoch",
		},
		common.ESDTNFTCreateOnMultiShardFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTNFTCreateOnMultiShardEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ESDTNFTCreateOnMultiShardEnableEpoch,
			activationEpochName: "ESDTNFTCreateOnMultiShardEnableEpoch",
		},
		common.MetaESDTSetFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MetaESDTSetEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.MetaESDTSetEnableEpoch,
			activationEpochName: "MetaESDTSetEnableEpoch",
		},
		common.AddTokensToDelegationFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.AddTokensToDelegationEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.AddTokensToDelegationEnableEpoch,
			activationEpochName: "AddTokensToDelegationEnableEpoch",
		},
		common.MultiESDTTransferFixOnCallBackFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MultiESDTTransferFixOnCallBackOnEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.MultiESDTTransferFixOnCallBackOnEnableEpoch,
			activationEpochName: "MultiESDTTransferFixOnCallBackOnEnableEpoch",
		},
		common.OptimizeGasUsedInCrossMiniBlocksFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.OptimizeGasUsedInCrossMiniBlocksEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.OptimizeGasUsedInCrossMiniBlocksEnableEpoch,
			activationEpochName: "OptimizeGasUsedInCrossMiniBlocksEnableEpoch",
		},
		common.CorrectFirstQueuedFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CorrectFirstQueuedEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.CorrectFirstQueuedEpoch,
			activationEpochName: "CorrectFirstQueuedEpoch",
		},
		common.DeleteDelegatorAfterClaimRewardsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DeleteDelegatorAfterClaimRewardsEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.DeleteDelegatorAfterClaimRewardsEnableEpoch,
			activationEpochName: "DeleteDelegatorAfterClaimRewardsEnableEpoch",
		},
		common.RemoveNonUpdatedStorageFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RemoveNonUpdatedStorageEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.RemoveNonUpdatedStorageEnableEpoch,
			activationEpochName: "RemoveNonUpdatedStorageEnableEpoch",
		},
		common.OptimizeNFTStoreFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch,
			activationEpochName: "OptimizeNFTStoreEnableEpoch",
		},
		common.SaveToSystemAccountFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch,
			activationEpochName: "OptimizeNFTStoreEnableEpoch",
		},
		common.CheckFrozenCollectionFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch,
			activationEpochName: "OptimizeNFTStoreEnableEpoch",
		},
		common.ValueLengthCheckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch,
			activationEpochName: "OptimizeNFTStoreEnableEpoch",
		},
		common.CheckTransferFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.OptimizeNFTStoreEnableEpoch,
			activationEpochName: "OptimizeNFTStoreEnableEpoch",
		},
		common.CreateNFTThroughExecByCallerFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CreateNFTThroughExecByCallerEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.CreateNFTThroughExecByCallerEnableEpoch,
			activationEpochName: "CreateNFTThroughExecByCallerEnableEpoch",
		},
		common.StopDecreasingValidatorRatingWhenStuckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.StopDecreasingValidatorRatingWhenStuckEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.StopDecreasingValidatorRatingWhenStuckEnableEpoch,
			activationEpochName: "StopDecreasingValidatorRatingWhenStuckEnableEpoch",
		},
		common.FrontRunningProtectionFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FrontRunningProtectionEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.FrontRunningProtectionEnableEpoch,
			activationEpochName: "FrontRunningProtectionEnableEpoch",
		},
		common.PayableBySCFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.IsPayableBySCEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.IsPayableBySCEnableEpoch,
			activationEpochName: "IsPayableBySCEnableEpoch",
		},
		common.CleanUpInformativeSCRsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CleanUpInformativeSCRsEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.CleanUpInformativeSCRsEnableEpoch,
			activationEpochName: "CleanUpInformativeSCRsEnableEpoch",
		},
		common.StorageAPICostOptimizationFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.StorageAPICostOptimizationEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.StorageAPICostOptimizationEnableEpoch,
			activationEpochName: "StorageAPICostOptimizationEnableEpoch",
		},
		common.ESDTRegisterAndSetAllRolesFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTRegisterAndSetAllRolesEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ESDTRegisterAndSetAllRolesEnableEpoch,
			activationEpochName: "ESDTRegisterAndSetAllRolesEnableEpoch",
		},
		common.ScheduledMiniBlocksFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ScheduledMiniBlocksEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ScheduledMiniBlocksEnableEpoch,
			activationEpochName: "ScheduledMiniBlocksEnableEpoch",
		},
		common.CorrectJailedNotUnStakedEmptyQueueFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CorrectJailedNotUnstakedEmptyQueueEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.CorrectJailedNotUnstakedEmptyQueueEpoch,
			activationEpochName: "CorrectJailedNotUnstakedEmptyQueueEpoch",
		},
		common.DoNotReturnOldBlockInBlockchainHookFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DoNotReturnOldBlockInBlockchainHookEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.DoNotReturnOldBlockInBlockchainHookEnableEpoch,
			activationEpochName: "DoNotReturnOldBlockInBlockchainHookEnableEpoch",
		},
		common.AddFailedRelayedTxToInvalidMBsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch < handler.enableEpochsConfig.AddFailedRelayedTxToInvalidMBsDisableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.AddFailedRelayedTxToInvalidMBsDisableEpoch,
			activationEpochName: "AddFailedRelayedTxToInvalidMBsDisableEpoch",
		},
		common.SCRSizeInvariantOnBuiltInResultFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SCRSizeInvariantOnBuiltInResultEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.SCRSizeInvariantOnBuiltInResultEnableEpoch,
			activationEpochName: "SCRSizeInvariantOnBuiltInResultEnableEpoch",
		},
		common.CheckCorrectTokenIDForTransferRoleFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CheckCorrectTokenIDForTransferRoleEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.CheckCorrectTokenIDForTransferRoleEnableEpoch,
			activationEpochName: "CheckCorrectTokenIDForTransferRoleEnableEpoch",
		},
		common.FailExecutionOnEveryAPIErrorFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FailExecutionOnEveryAPIErrorEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.FailExecutionOnEveryAPIErrorEnableEpoch,
			activationEpochName: "FailExecutionOnEveryAPIErrorEnableEpoch",
		},
		common.MiniBlockPartialExecutionFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MiniBlockPartialExecutionEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.MiniBlockPartialExecutionEnableEpoch,
			activationEpochName: "MiniBlockPartialExecutionEnableEpoch",
		},
		common.ManagedCryptoAPIsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ManagedCryptoAPIsEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ManagedCryptoAPIsEnableEpoch,
			activationEpochName: "ManagedCryptoAPIsEnableEpoch",
		},
		common.ESDTMetadataContinuousCleanupFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch,
			activationEpochName: "ESDTMetadataContinuousCleanupEnableEpoch",
		},
		common.FixAsyncCallbackCheckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch,
			activationEpochName: "ESDTMetadataContinuousCleanupEnableEpoch",
		},
		common.SendAlwaysFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch,
			activationEpochName: "ESDTMetadataContinuousCleanupEnableEpoch",
		},
		common.ChangeDelegationOwnerFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ESDTMetadataContinuousCleanupEnableEpoch,
			activationEpochName: "ESDTMetadataContinuousCleanupEnableEpoch",
		},
		common.DisableExecByCallerFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DisableExecByCallerEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.DisableExecByCallerEnableEpoch,
			activationEpochName: "DisableExecByCallerEnableEpoch",
		},
		common.RefactorContextFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RefactorContextEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.RefactorContextEnableEpoch,
			activationEpochName: "RefactorContextEnableEpoch",
		},
		common.CheckFunctionArgumentFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CheckFunctionArgumentEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.CheckFunctionArgumentEnableEpoch,
			activationEpochName: "CheckFunctionArgumentEnableEpoch",
		},
		common.CheckExecuteOnReadOnlyFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CheckExecuteOnReadOnlyEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.CheckExecuteOnReadOnlyEnableEpoch,
			activationEpochName: "CheckExecuteOnReadOnlyEnableEpoch",
		},
		common.SetSenderInEeiOutputTransferFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SetSenderInEeiOutputTransferEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.SetSenderInEeiOutputTransferEnableEpoch,
			activationEpochName: "SetSenderInEeiOutputTransferEnableEpoch",
		},
		common.RefactorPeersMiniBlocksFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RefactorPeersMiniBlocksEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.RefactorPeersMiniBlocksEnableEpoch,
			activationEpochName: "RefactorPeersMiniBlocksEnableEpoch",
		},
		common.SCProcessorV2Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SCProcessorV2EnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.SCProcessorV2EnableEpoch,
			activationEpochName: "SCProcessorV2EnableEpoch",
		},
		common.FixAsyncCallBackArgsListFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixAsyncCallBackArgsListEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.FixAsyncCallBackArgsListEnableEpoch,
			activationEpochName: "FixAsyncCallBackArgsListEnableEpoch",
		},
		common.FixOldTokenLiquidityFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixOldTokenLiquidityEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.FixOldTokenLiquidityEnableEpoch,
			activationEpochName: "FixOldTokenLiquidityEnableEpoch",
		},
		common.RuntimeMemStoreLimitFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RuntimeMemStoreLimitEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.RuntimeMemStoreLimitEnableEpoch,
			activationEpochName: "RuntimeMemStoreLimitEnableEpoch",
		},
		common.RuntimeCodeSizeFixFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RuntimeCodeSizeFixEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.RuntimeCodeSizeFixEnableEpoch,
			activationEpochName: "RuntimeCodeSizeFixEnableEpoch",
		},
		common.MaxBlockchainHookCountersFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MaxBlockchainHookCountersEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.MaxBlockchainHookCountersEnableEpoch,
			activationEpochName: "MaxBlockchainHookCountersEnableEpoch",
		},
		common.WipeSingleNFTLiquidityDecreaseFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.WipeSingleNFTLiquidityDecreaseEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.WipeSingleNFTLiquidityDecreaseEnableEpoch,
			activationEpochName: "WipeSingleNFTLiquidityDecreaseEnableEpoch",
		},
		common.AlwaysSaveTokenMetaDataFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.AlwaysSaveTokenMetaDataEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.AlwaysSaveTokenMetaDataEnableEpoch,
			activationEpochName: "AlwaysSaveTokenMetaDataEnableEpoch",
		},
		common.SetGuardianFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SetGuardianEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.SetGuardianEnableEpoch,
			activationEpochName: "SetGuardianEnableEpoch",
		},
		common.RelayedNonceFixFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RelayedNonceFixEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.RelayedNonceFixEnableEpoch,
			activationEpochName: "RelayedNonceFixEnableEpoch",
		},
		common.ConsistentTokensValuesLengthCheckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ConsistentTokensValuesLengthCheckEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ConsistentTokensValuesLengthCheckEnableEpoch,
			activationEpochName: "ConsistentTokensValuesLengthCheckEnableEpoch",
		},
		common.KeepExecOrderOnCreatedSCRsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.KeepExecOrderOnCreatedSCRsEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.KeepExecOrderOnCreatedSCRsEnableEpoch,
			activationEpochName: "KeepExecOrderOnCreatedSCRsEnableEpoch",
		},
		common.MultiClaimOnDelegationFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MultiClaimOnDelegationEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.MultiClaimOnDelegationEnableEpoch,
			activationEpochName: "MultiClaimOnDelegationEnableEpoch",
		},
		common.ChangeUsernameFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ChangeUsernameEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ChangeUsernameEnableEpoch,
			activationEpochName: "ChangeUsernameEnableEpoch",
		},
		common.AutoBalanceDataTriesFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.AutoBalanceDataTriesEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.AutoBalanceDataTriesEnableEpoch,
			activationEpochName: "AutoBalanceDataTriesEnableEpoch",
		},
		common.MigrateDataTrieFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MigrateDataTrieEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.MigrateDataTrieEnableEpoch,
			activationEpochName: "MigrateDataTrieEnableEpoch",
		},
		common.FixDelegationChangeOwnerOnAccountFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixDelegationChangeOwnerOnAccountEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.FixDelegationChangeOwnerOnAccountEnableEpoch,
			activationEpochName: "FixDelegationChangeOwnerOnAccountEnableEpoch",
		},
		common.FixOOGReturnCodeFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixOOGReturnCodeEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.FixOOGReturnCodeEnableEpoch,
			activationEpochName: "FixOOGReturnCodeEnableEpoch",
		},
		common.DeterministicSortOnValidatorsInfoFixFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DeterministicSortOnValidatorsInfoEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.DeterministicSortOnValidatorsInfoEnableEpoch,
			activationEpochName: "DeterministicSortOnValidatorsInfoEnableEpoch",
		},
		common.DynamicGasCostForDataTrieStorageLoadFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DynamicGasCostForDataTrieStorageLoadEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.DynamicGasCostForDataTrieStorageLoadEnableEpoch,
			activationEpochName: "DynamicGasCostForDataTrieStorageLoadEnableEpoch",
		},
		common.ScToScLogEventFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ScToScLogEventEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ScToScLogEventEnableEpoch,
			activationEpochName: "ScToScLogEventEnableEpoch",
		},
		common.BlockGasAndFeesReCheckFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.BlockGasAndFeesReCheckEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.BlockGasAndFeesReCheckEnableEpoch,
			activationEpochName: "BlockGasAndFeesReCheckEnableEpoch",
		},
		common.BalanceWaitingListsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.BalanceWaitingListsEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.BalanceWaitingListsEnableEpoch,
			activationEpochName: "BalanceWaitingListsEnableEpoch",
		},
		common.NFTStopCreateFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.NFTStopCreateEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.NFTStopCreateEnableEpoch,
			activationEpochName: "NFTStopCreateEnableEpoch",
		},
		common.FixGasRemainingForSaveKeyValueFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixGasRemainingForSaveKeyValueBuiltinFunctionEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.FixGasRemainingForSaveKeyValueBuiltinFunctionEnableEpoch,
			activationEpochName: "FixGasRemainingForSaveKeyValueBuiltinFunctionEnableEpoch",
		},
		common.IsChangeOwnerAddressCrossShardThroughSCFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ChangeOwnerAddressCrossShardThroughSCEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ChangeOwnerAddressCrossShardThroughSCEnableEpoch,
			activationEpochName: "ChangeOwnerAddressCrossShardThroughSCEnableEpoch",
		},
		common.CurrentRandomnessOnSortingFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CurrentRandomnessOnSortingEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.CurrentRandomnessOnSortingEnableEpoch,
			activationEpochName: "CurrentRandomnessOnSortingEnableEpoch",
		},
		common.StakeLimitsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.StakeLimitsEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.StakeLimitsEnableEpoch,
			activationEpochName: "StakeLimitsEnableEpoch",
		},
		common.StakingV4Step1Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch == handler.enableEpochsConfig.StakingV4Step1EnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.StakingV4Step1EnableEpoch,
			activationEpochName: "StakingV4Step1EnableEpoch",
		},
		common.StakingV4Step2Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.StakingV4Step2EnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.StakingV4Step2EnableEpoch,
			activationEpochName: "StakingV4Step2EnableEpoch",
		},
		common.StakingV4Step3Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.StakingV4Step3EnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.StakingV4Step3EnableEpoch,
			activationEpochName: "StakingV4Step3EnableEpoch",
		},
		common.CleanupAuctionOnLowWaitingListFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CleanupAuctionOnLowWaitingListEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.CleanupAuctionOnLowWaitingListEnableEpoch,
			activationEpochName: "CleanupAuctionOnLowWaitingListEnableEpoch",
		},
		common.StakingV4StartedFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.StakingV4Step1EnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.StakingV4Step1EnableEpoch,
			activationEpochName: "StakingV4Step1EnableEpoch",
		},
		common.AlwaysMergeContextsInEEIFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.AlwaysMergeContextsInEEIEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.AlwaysMergeContextsInEEIEnableEpoch,
			activationEpochName: "AlwaysMergeContextsInEEIEnableEpoch",
		},
		common.UseGasBoundedShouldFailExecutionFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.UseGasBoundedShouldFailExecutionEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.UseGasBoundedShouldFailExecutionEnableEpoch,
			activationEpochName: "UseGasBoundedShouldFailExecutionEnableEpoch",
		},
		common.DynamicESDTFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.DynamicESDTEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.DynamicESDTEnableEpoch,
			activationEpochName: "DynamicESDTEnableEpoch",
		},
		common.EGLDInESDTMultiTransferFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.EGLDInMultiTransferEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.EGLDInMultiTransferEnableEpoch,
			activationEpochName: "EGLDInMultiTransferEnableEpoch",
		},
		common.CryptoOpcodesV2Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CryptoOpcodesV2EnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.CryptoOpcodesV2EnableEpoch,
			activationEpochName: "CryptoOpcodesV2EnableEpoch",
		},
		common.UnJailCleanupFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.UnJailCleanupEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.UnJailCleanupEnableEpoch,
			activationEpochName: "UnJailCleanupEnableEpoch",
		},
		common.FixRelayedBaseCostFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixRelayedBaseCostEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.FixRelayedBaseCostEnableEpoch,
			activationEpochName: "FixRelayedBaseCostEnableEpoch",
		},
		common.MultiESDTNFTTransferAndExecuteByUserFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MultiESDTNFTTransferAndExecuteByUserEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.MultiESDTNFTTransferAndExecuteByUserEnableEpoch,
			activationEpochName: "MultiESDTNFTTransferAndExecuteByUserEnableEpoch",
		},
		common.FixRelayedMoveBalanceToNonPayableSCFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixRelayedMoveBalanceToNonPayableSCEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.FixRelayedMoveBalanceToNonPayableSCEnableEpoch,
			activationEpochName: "FixRelayedMoveBalanceToNonPayableSCEnableEpoch",
		},
		common.RelayedTransactionsV3Flag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RelayedTransactionsV3EnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.RelayedTransactionsV3EnableEpoch,
			activationEpochName: "RelayedTransactionsV3EnableEpoch",
		},
		common.RelayedTransactionsV3FixESDTTransferFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RelayedTransactionsV3FixESDTTransferEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.RelayedTransactionsV3FixESDTTransferEnableEpoch,
			activationEpochName: "RelayedTransactionsV3FixESDTTransferEnableEpoch",
		},
		common.AndromedaFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.AndromedaEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.AndromedaEnableEpoch,
			activationEpochName: "AndromedaEnableEpoch",
		},
		common.SupernovaFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.SupernovaEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.SupernovaEnableEpoch,
			activationEpochName: "SupernovaEnableEpoch",
		},
		// TODO: move it to activation round
		common.CheckBuiltInCallOnTransferValueAndFailExecutionFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.CheckBuiltInCallOnTransferValueAndFailEnableRound
			},
			activationEpoch:     handler.enableEpochsConfig.CheckBuiltInCallOnTransferValueAndFailEnableRound,
			activationEpochName: "CheckBuiltInCallOnTransferValueAndFailEnableRound",
		},
		common.AutomaticActivationOfNodesDisableFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.AutomaticActivationOfNodesDisableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.AutomaticActivationOfNodesDisableEpoch,
			activationEpochName: "AutomaticActivationOfNodesDisableEpoch",
		},
		common.MaskInternalDependenciesErrorsFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.MaskVMInternalDependenciesErrorsEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.MaskVMInternalDependenciesErrorsEnableEpoch,
			activationEpochName: "MaskVMInternalDependenciesErrorsEnableEpoch",
		},
		common.FixBackTransferOPCODEFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixBackTransferOPCODEEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.FixBackTransferOPCODEEnableEpoch,
			activationEpochName: "FixBackTransferOPCODEEnableEpoch",
		},
		common.ValidationOnGobDecodeFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.ValidationOnGobDecodeEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.ValidationOnGobDecodeEnableEpoch,
			activationEpochName: "ValidationOnGobDecodeEnableEpoch",
		},
		common.BarnardOpcodesFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.BarnardOpcodesEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.BarnardOpcodesEnableEpoch,
			activationEpochName: "BarnardOpcodesEnableEpoch",
		},
		common.FixGetBalanceFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.FixGetBalanceEnableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.FixGetBalanceEnableEpoch,
			activationEpochName: "FixGetBalanceEnableEpoch",
		},
		common.RelayedTransactionsV1V2DisableFlag: {
			isActiveInEpoch: func(epoch uint32) bool {
				return epoch >= handler.enableEpochsConfig.RelayedTransactionsV1V2DisableEpoch
			},
			activationEpoch:     handler.enableEpochsConfig.RelayedTransactionsV1V2DisableEpoch,
			activationEpochName: "RelayedTransactionsV1V2DisableEpoch",
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

// StakingV4Step2EnableEpoch returns the epoch when stakingV4 becomes active
func (handler *enableEpochsHandler) StakingV4Step2EnableEpoch() uint32 {
	return handler.enableEpochsConfig.StakingV4Step2EnableEpoch
}

// StakingV4Step1EnableEpoch returns the epoch when stakingV4 phase1 becomes active
func (handler *enableEpochsHandler) StakingV4Step1EnableEpoch() uint32 {
	return handler.enableEpochsConfig.StakingV4Step1EnableEpoch
}

// GetAllEnableEpochs returns all flags with their activation epochs
func (handler *enableEpochsHandler) GetAllEnableEpochs() map[string]uint32 {
	result := make(map[string]uint32, len(handler.allFlagsDefined))
	for _, fh := range handler.allFlagsDefined {
		result[fh.activationEpochName] = fh.activationEpoch
	}
	return result
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *enableEpochsHandler) IsInterfaceNil() bool {
	return handler == nil
}
