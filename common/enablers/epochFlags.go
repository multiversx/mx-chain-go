package enablers

import (
	"github.com/multiversx/mx-chain-core-go/core/atomic"
)

type epochFlagsHolder struct {
	builtInFunctionsFlag                        *atomic.Flag
	relayedTransactionsFlag                     *atomic.Flag
	penalizedTooMuchGasFlag                     *atomic.Flag
	switchJailWaitingFlag                       *atomic.Flag
	belowSignedThresholdFlag                    *atomic.Flag
	switchHysteresisForMinNodesFlag             *atomic.Flag
	switchHysteresisForMinNodesCurrentEpochFlag *atomic.Flag
	transactionSignedWithTxHashFlag             *atomic.Flag
	metaProtectionFlag                          *atomic.Flag
	aheadOfTimeGasUsageFlag                     *atomic.Flag
	gasPriceModifierFlag                        *atomic.Flag
	repairCallbackFlag                          *atomic.Flag
	balanceWaitingListsFlag                     *atomic.Flag
	returnDataToLastTransferFlag                *atomic.Flag
	senderInOutTransferFlag                     *atomic.Flag
	stakeFlag                                   *atomic.Flag
	stakingV2Flag                               *atomic.Flag
	stakingV2OwnerFlag                          *atomic.Flag
	stakingV2GreaterEpochFlag                   *atomic.Flag
	doubleKeyProtectionFlag                     *atomic.Flag
	esdtFlag                                    *atomic.Flag
	esdtCurrentEpochFlag                        *atomic.Flag
	governanceFlag                              *atomic.Flag
	governanceCurrentEpochFlag                  *atomic.Flag
	delegationManagerFlag                       *atomic.Flag
	delegationSmartContractFlag                 *atomic.Flag
	delegationSmartContractCurrentEpochFlag     *atomic.Flag
	correctLastUnJailedFlag                     *atomic.Flag
	correctLastUnJailedCurrentEpochFlag         *atomic.Flag
	relayedTransactionsV2Flag                   *atomic.Flag
	unBondTokensV2Flag                          *atomic.Flag
	saveJailedAlwaysFlag                        *atomic.Flag
	reDelegateBelowMinCheckFlag                 *atomic.Flag
	validatorToDelegationFlag                   *atomic.Flag
	waitingListFixFlag                          *atomic.Flag
	incrementSCRNonceInMultiTransferFlag        *atomic.Flag
	esdtMultiTransferFlag                       *atomic.Flag
	globalMintBurnFlag                          *atomic.Flag
	esdtTransferRoleFlag                        *atomic.Flag
	builtInFunctionOnMetaFlag                   *atomic.Flag
	computeRewardCheckpointFlag                 *atomic.Flag
	scrSizeInvariantCheckFlag                   *atomic.Flag
	backwardCompSaveKeyValueFlag                *atomic.Flag
	esdtNFTCreateOnMultiShardFlag               *atomic.Flag
	metaESDTSetFlag                             *atomic.Flag
	addTokensToDelegationFlag                   *atomic.Flag
	multiESDTTransferFixOnCallBackFlag          *atomic.Flag
	optimizeGasUsedInCrossMiniBlocksFlag        *atomic.Flag
	correctFirstQueuedFlag                      *atomic.Flag
	deleteDelegatorAfterClaimRewardsFlag        *atomic.Flag
	fixOOGReturnCodeFlag                        *atomic.Flag
	removeNonUpdatedStorageFlag                 *atomic.Flag
	optimizeNFTStoreFlag                        *atomic.Flag
	createNFTThroughExecByCallerFlag            *atomic.Flag
	stopDecreasingValidatorRatingWhenStuckFlag  *atomic.Flag
	frontRunningProtectionFlag                  *atomic.Flag
	isPayableBySCFlag                           *atomic.Flag
	cleanUpInformativeSCRsFlag                  *atomic.Flag
	storageAPICostOptimizationFlag              *atomic.Flag
	esdtRegisterAndSetAllRolesFlag              *atomic.Flag
	scheduledMiniBlocksFlag                     *atomic.Flag
	correctJailedNotUnStakedEmptyQueueFlag      *atomic.Flag
	doNotReturnOldBlockInBlockchainHookFlag     *atomic.Flag
	addFailedRelayedTxToInvalidMBsFlag          *atomic.Flag
	scrSizeInvariantOnBuiltInResultFlag         *atomic.Flag
	checkCorrectTokenIDForTransferRoleFlag      *atomic.Flag
	failExecutionOnEveryAPIErrorFlag            *atomic.Flag
	isMiniBlockPartialExecutionFlag             *atomic.Flag
	managedCryptoAPIsFlag                       *atomic.Flag
	esdtMetadataContinuousCleanupFlag           *atomic.Flag
	disableExecByCallerFlag                     *atomic.Flag
	refactorContextFlag                         *atomic.Flag
	checkFunctionArgumentFlag                   *atomic.Flag
	checkExecuteOnReadOnlyFlag                  *atomic.Flag
	setSenderInEeiOutputTransferFlag            *atomic.Flag
	changeDelegationOwnerFlag                   *atomic.Flag
	refactorPeersMiniBlocksFlag                 *atomic.Flag
	scProcessorV2Flag                           *atomic.Flag
	fixAsyncCallBackArgsList                    *atomic.Flag
	fixOldTokenLiquidity                        *atomic.Flag
	runtimeMemStoreLimitFlag                    *atomic.Flag
	runtimeCodeSizeFixFlag                      *atomic.Flag
	maxBlockchainHookCountersFlag               *atomic.Flag
	wipeSingleNFTLiquidityDecreaseFlag          *atomic.Flag
	alwaysSaveTokenMetaDataFlag                 *atomic.Flag
	setGuardianFlag                             *atomic.Flag
	relayedNonceFixFlag                         *atomic.Flag
	keepExecOrderOnCreatedSCRsFlag              *atomic.Flag
	multiClaimOnDelegationFlag                  *atomic.Flag
	changeUsernameFlag                          *atomic.Flag
	consistentTokensValuesCheckFlag             *atomic.Flag
	autoBalanceDataTriesFlag                    *atomic.Flag
	fixDelegationChangeOwnerOnAccountFlag       *atomic.Flag
}

func newEpochFlagsHolder() *epochFlagsHolder {
	return &epochFlagsHolder{
		builtInFunctionsFlag:                        &atomic.Flag{},
		relayedTransactionsFlag:                     &atomic.Flag{},
		penalizedTooMuchGasFlag:                     &atomic.Flag{},
		switchJailWaitingFlag:                       &atomic.Flag{},
		belowSignedThresholdFlag:                    &atomic.Flag{},
		switchHysteresisForMinNodesFlag:             &atomic.Flag{},
		switchHysteresisForMinNodesCurrentEpochFlag: &atomic.Flag{},
		transactionSignedWithTxHashFlag:             &atomic.Flag{},
		metaProtectionFlag:                          &atomic.Flag{},
		aheadOfTimeGasUsageFlag:                     &atomic.Flag{},
		gasPriceModifierFlag:                        &atomic.Flag{},
		repairCallbackFlag:                          &atomic.Flag{},
		balanceWaitingListsFlag:                     &atomic.Flag{},
		returnDataToLastTransferFlag:                &atomic.Flag{},
		senderInOutTransferFlag:                     &atomic.Flag{},
		stakeFlag:                                   &atomic.Flag{},
		stakingV2Flag:                               &atomic.Flag{},
		stakingV2OwnerFlag:                          &atomic.Flag{},
		stakingV2GreaterEpochFlag:                   &atomic.Flag{},
		doubleKeyProtectionFlag:                     &atomic.Flag{},
		esdtFlag:                                    &atomic.Flag{},
		esdtCurrentEpochFlag:                        &atomic.Flag{},
		governanceFlag:                              &atomic.Flag{},
		governanceCurrentEpochFlag:                  &atomic.Flag{},
		delegationManagerFlag:                       &atomic.Flag{},
		delegationSmartContractFlag:                 &atomic.Flag{},
		delegationSmartContractCurrentEpochFlag:     &atomic.Flag{},
		correctLastUnJailedFlag:                     &atomic.Flag{},
		correctLastUnJailedCurrentEpochFlag:         &atomic.Flag{},
		relayedTransactionsV2Flag:                   &atomic.Flag{},
		unBondTokensV2Flag:                          &atomic.Flag{},
		saveJailedAlwaysFlag:                        &atomic.Flag{},
		reDelegateBelowMinCheckFlag:                 &atomic.Flag{},
		validatorToDelegationFlag:                   &atomic.Flag{},
		waitingListFixFlag:                          &atomic.Flag{},
		incrementSCRNonceInMultiTransferFlag:        &atomic.Flag{},
		esdtMultiTransferFlag:                       &atomic.Flag{},
		globalMintBurnFlag:                          &atomic.Flag{},
		esdtTransferRoleFlag:                        &atomic.Flag{},
		builtInFunctionOnMetaFlag:                   &atomic.Flag{},
		computeRewardCheckpointFlag:                 &atomic.Flag{},
		scrSizeInvariantCheckFlag:                   &atomic.Flag{},
		backwardCompSaveKeyValueFlag:                &atomic.Flag{},
		esdtNFTCreateOnMultiShardFlag:               &atomic.Flag{},
		metaESDTSetFlag:                             &atomic.Flag{},
		addTokensToDelegationFlag:                   &atomic.Flag{},
		multiESDTTransferFixOnCallBackFlag:          &atomic.Flag{},
		optimizeGasUsedInCrossMiniBlocksFlag:        &atomic.Flag{},
		correctFirstQueuedFlag:                      &atomic.Flag{},
		deleteDelegatorAfterClaimRewardsFlag:        &atomic.Flag{},
		fixOOGReturnCodeFlag:                        &atomic.Flag{},
		removeNonUpdatedStorageFlag:                 &atomic.Flag{},
		optimizeNFTStoreFlag:                        &atomic.Flag{},
		createNFTThroughExecByCallerFlag:            &atomic.Flag{},
		stopDecreasingValidatorRatingWhenStuckFlag:  &atomic.Flag{},
		frontRunningProtectionFlag:                  &atomic.Flag{},
		isPayableBySCFlag:                           &atomic.Flag{},
		cleanUpInformativeSCRsFlag:                  &atomic.Flag{},
		storageAPICostOptimizationFlag:              &atomic.Flag{},
		esdtRegisterAndSetAllRolesFlag:              &atomic.Flag{},
		scheduledMiniBlocksFlag:                     &atomic.Flag{},
		correctJailedNotUnStakedEmptyQueueFlag:      &atomic.Flag{},
		doNotReturnOldBlockInBlockchainHookFlag:     &atomic.Flag{},
		addFailedRelayedTxToInvalidMBsFlag:          &atomic.Flag{},
		scrSizeInvariantOnBuiltInResultFlag:         &atomic.Flag{},
		checkCorrectTokenIDForTransferRoleFlag:      &atomic.Flag{},
		failExecutionOnEveryAPIErrorFlag:            &atomic.Flag{},
		isMiniBlockPartialExecutionFlag:             &atomic.Flag{},
		managedCryptoAPIsFlag:                       &atomic.Flag{},
		esdtMetadataContinuousCleanupFlag:           &atomic.Flag{},
		disableExecByCallerFlag:                     &atomic.Flag{},
		refactorContextFlag:                         &atomic.Flag{},
		checkFunctionArgumentFlag:                   &atomic.Flag{},
		checkExecuteOnReadOnlyFlag:                  &atomic.Flag{},
		setSenderInEeiOutputTransferFlag:            &atomic.Flag{},
		changeDelegationOwnerFlag:                   &atomic.Flag{},
		refactorPeersMiniBlocksFlag:                 &atomic.Flag{},
		scProcessorV2Flag:                           &atomic.Flag{},
		fixAsyncCallBackArgsList:                    &atomic.Flag{},
		fixOldTokenLiquidity:                        &atomic.Flag{},
		runtimeMemStoreLimitFlag:                    &atomic.Flag{},
		runtimeCodeSizeFixFlag:                      &atomic.Flag{},
		maxBlockchainHookCountersFlag:               &atomic.Flag{},
		wipeSingleNFTLiquidityDecreaseFlag:          &atomic.Flag{},
		alwaysSaveTokenMetaDataFlag:                 &atomic.Flag{},
		setGuardianFlag:                             &atomic.Flag{},
		relayedNonceFixFlag:                         &atomic.Flag{},
		keepExecOrderOnCreatedSCRsFlag:              &atomic.Flag{},
		consistentTokensValuesCheckFlag:             &atomic.Flag{},
		multiClaimOnDelegationFlag:                  &atomic.Flag{},
		changeUsernameFlag:                          &atomic.Flag{},
		autoBalanceDataTriesFlag:                    &atomic.Flag{},
		fixDelegationChangeOwnerOnAccountFlag:       &atomic.Flag{},
	}
}

// IsBuiltInFunctionsFlagEnabled returns true if builtInFunctionsFlag is enabled
func (holder *epochFlagsHolder) IsBuiltInFunctionsFlagEnabled() bool {
	return holder.builtInFunctionsFlag.IsSet()
}

// IsRelayedTransactionsFlagEnabled returns true if relayedTransactionsFlag is enabled
func (holder *epochFlagsHolder) IsRelayedTransactionsFlagEnabled() bool {
	return holder.relayedTransactionsFlag.IsSet()
}

// IsPenalizedTooMuchGasFlagEnabled returns true if penalizedTooMuchGasFlag is enabled
func (holder *epochFlagsHolder) IsPenalizedTooMuchGasFlagEnabled() bool {
	return holder.penalizedTooMuchGasFlag.IsSet()
}

// ResetPenalizedTooMuchGasFlag resets the penalizedTooMuchGasFlag
func (holder *epochFlagsHolder) ResetPenalizedTooMuchGasFlag() {
	holder.penalizedTooMuchGasFlag.Reset()
}

// IsSwitchJailWaitingFlagEnabled returns true if switchJailWaitingFlag is enabled
func (holder *epochFlagsHolder) IsSwitchJailWaitingFlagEnabled() bool {
	return holder.switchJailWaitingFlag.IsSet()
}

// IsBelowSignedThresholdFlagEnabled returns true if belowSignedThresholdFlag is enabled
func (holder *epochFlagsHolder) IsBelowSignedThresholdFlagEnabled() bool {
	return holder.belowSignedThresholdFlag.IsSet()
}

// IsSwitchHysteresisForMinNodesFlagEnabled returns true if switchHysteresisForMinNodesFlag is enabled
func (holder *epochFlagsHolder) IsSwitchHysteresisForMinNodesFlagEnabled() bool {
	return holder.switchHysteresisForMinNodesFlag.IsSet()
}

// IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch returns true if switchHysteresisForMinNodesCurrentEpochFlag is enabled
func (holder *epochFlagsHolder) IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch() bool {
	return holder.switchHysteresisForMinNodesCurrentEpochFlag.IsSet()
}

// IsTransactionSignedWithTxHashFlagEnabled returns true if transactionSignedWithTxHashFlag is enabled
func (holder *epochFlagsHolder) IsTransactionSignedWithTxHashFlagEnabled() bool {
	return holder.transactionSignedWithTxHashFlag.IsSet()
}

// IsMetaProtectionFlagEnabled returns true if metaProtectionFlag is enabled
func (holder *epochFlagsHolder) IsMetaProtectionFlagEnabled() bool {
	return holder.metaProtectionFlag.IsSet()
}

// IsAheadOfTimeGasUsageFlagEnabled returns true if aheadOfTimeGasUsageFlag is enabled
func (holder *epochFlagsHolder) IsAheadOfTimeGasUsageFlagEnabled() bool {
	return holder.aheadOfTimeGasUsageFlag.IsSet()
}

// IsGasPriceModifierFlagEnabled returns true if gasPriceModifierFlag is enabled
func (holder *epochFlagsHolder) IsGasPriceModifierFlagEnabled() bool {
	return holder.gasPriceModifierFlag.IsSet()
}

// IsRepairCallbackFlagEnabled returns true if repairCallbackFlag is enabled
func (holder *epochFlagsHolder) IsRepairCallbackFlagEnabled() bool {
	return holder.repairCallbackFlag.IsSet()
}

// IsBalanceWaitingListsFlagEnabled returns true if balanceWaitingListsFlag is enabled
func (holder *epochFlagsHolder) IsBalanceWaitingListsFlagEnabled() bool {
	return holder.balanceWaitingListsFlag.IsSet()
}

// IsReturnDataToLastTransferFlagEnabled returns true if returnDataToLastTransferFlag is enabled
func (holder *epochFlagsHolder) IsReturnDataToLastTransferFlagEnabled() bool {
	return holder.returnDataToLastTransferFlag.IsSet()
}

// IsSenderInOutTransferFlagEnabled returns true if senderInOutTransferFlag is enabled
func (holder *epochFlagsHolder) IsSenderInOutTransferFlagEnabled() bool {
	return holder.senderInOutTransferFlag.IsSet()
}

// IsStakeFlagEnabled returns true if stakeFlag is enabled
func (holder *epochFlagsHolder) IsStakeFlagEnabled() bool {
	return holder.stakeFlag.IsSet()
}

// IsStakingV2FlagEnabled returns true if stakingV2Flag is enabled
func (holder *epochFlagsHolder) IsStakingV2FlagEnabled() bool {
	return holder.stakingV2Flag.IsSet()
}

// IsStakingV2OwnerFlagEnabled returns true if stakingV2OwnerFlag is enabled
func (holder *epochFlagsHolder) IsStakingV2OwnerFlagEnabled() bool {
	return holder.stakingV2OwnerFlag.IsSet()
}

// IsStakingV2FlagEnabledForActivationEpochCompleted returns true if stakingV2GreaterEpochFlag is enabled (epoch is greater than the one used for staking v2 activation)
func (holder *epochFlagsHolder) IsStakingV2FlagEnabledForActivationEpochCompleted() bool {
	return holder.stakingV2GreaterEpochFlag.IsSet()
}

// IsDoubleKeyProtectionFlagEnabled returns true if doubleKeyProtectionFlag is enabled
func (holder *epochFlagsHolder) IsDoubleKeyProtectionFlagEnabled() bool {
	return holder.doubleKeyProtectionFlag.IsSet()
}

// IsESDTFlagEnabled returns true if esdtFlag is enabled
func (holder *epochFlagsHolder) IsESDTFlagEnabled() bool {
	return holder.esdtFlag.IsSet()
}

// IsESDTFlagEnabledForCurrentEpoch returns true if esdtCurrentEpochFlag is enabled
func (holder *epochFlagsHolder) IsESDTFlagEnabledForCurrentEpoch() bool {
	return holder.esdtCurrentEpochFlag.IsSet()
}

// IsGovernanceFlagEnabled returns true if governanceFlag is enabled
func (holder *epochFlagsHolder) IsGovernanceFlagEnabled() bool {
	return holder.governanceFlag.IsSet()
}

// IsGovernanceFlagEnabledForCurrentEpoch returns true if governanceCurrentEpochFlag is enabled
func (holder *epochFlagsHolder) IsGovernanceFlagEnabledForCurrentEpoch() bool {
	return holder.governanceCurrentEpochFlag.IsSet()
}

// IsDelegationManagerFlagEnabled returns true if delegationManagerFlag is enabled
func (holder *epochFlagsHolder) IsDelegationManagerFlagEnabled() bool {
	return holder.delegationManagerFlag.IsSet()
}

// IsDelegationSmartContractFlagEnabled returns true if delegationSmartContractFlag is enabled
func (holder *epochFlagsHolder) IsDelegationSmartContractFlagEnabled() bool {
	return holder.delegationSmartContractFlag.IsSet()
}

// IsDelegationSmartContractFlagEnabledForCurrentEpoch returns true if delegationSmartContractCurrentEpochFlag is enabled
func (holder *epochFlagsHolder) IsDelegationSmartContractFlagEnabledForCurrentEpoch() bool {
	return holder.delegationSmartContractCurrentEpochFlag.IsSet()
}

// IsCorrectLastUnJailedFlagEnabled returns true if correctLastUnJailedFlag is enabled
func (holder *epochFlagsHolder) IsCorrectLastUnJailedFlagEnabled() bool {
	return holder.correctLastUnJailedFlag.IsSet()
}

// IsCorrectLastUnJailedFlagEnabledForCurrentEpoch returns true if correctLastUnJailedCurrentEpochFlag is enabled
func (holder *epochFlagsHolder) IsCorrectLastUnJailedFlagEnabledForCurrentEpoch() bool {
	return holder.correctLastUnJailedCurrentEpochFlag.IsSet()
}

// IsRelayedTransactionsV2FlagEnabled returns true if relayedTransactionsV2Flag is enabled
func (holder *epochFlagsHolder) IsRelayedTransactionsV2FlagEnabled() bool {
	return holder.relayedTransactionsV2Flag.IsSet()
}

// IsUnBondTokensV2FlagEnabled returns true if unBondTokensV2Flag is enabled
func (holder *epochFlagsHolder) IsUnBondTokensV2FlagEnabled() bool {
	return holder.unBondTokensV2Flag.IsSet()
}

// IsSaveJailedAlwaysFlagEnabled returns true if saveJailedAlwaysFlag is enabled
func (holder *epochFlagsHolder) IsSaveJailedAlwaysFlagEnabled() bool {
	return holder.saveJailedAlwaysFlag.IsSet()
}

// IsReDelegateBelowMinCheckFlagEnabled returns true if reDelegateBelowMinCheckFlag is enabled
func (holder *epochFlagsHolder) IsReDelegateBelowMinCheckFlagEnabled() bool {
	return holder.reDelegateBelowMinCheckFlag.IsSet()
}

// IsValidatorToDelegationFlagEnabled returns true if validatorToDelegationFlag is enabled
func (holder *epochFlagsHolder) IsValidatorToDelegationFlagEnabled() bool {
	return holder.validatorToDelegationFlag.IsSet()
}

// IsWaitingListFixFlagEnabled returns true if waitingListFixFlag is enabled
func (holder *epochFlagsHolder) IsWaitingListFixFlagEnabled() bool {
	return holder.waitingListFixFlag.IsSet()
}

// IsIncrementSCRNonceInMultiTransferFlagEnabled returns true if incrementSCRNonceInMultiTransferFlag is enabled
func (holder *epochFlagsHolder) IsIncrementSCRNonceInMultiTransferFlagEnabled() bool {
	return holder.incrementSCRNonceInMultiTransferFlag.IsSet()
}

// IsESDTMultiTransferFlagEnabled returns true if esdtMultiTransferFlag is enabled
func (holder *epochFlagsHolder) IsESDTMultiTransferFlagEnabled() bool {
	return holder.esdtMultiTransferFlag.IsSet()
}

// IsGlobalMintBurnFlagEnabled returns true if globalMintBurnFlag is enabled
func (holder *epochFlagsHolder) IsGlobalMintBurnFlagEnabled() bool {
	return holder.globalMintBurnFlag.IsSet()
}

// IsESDTTransferRoleFlagEnabled returns true if esdtTransferRoleFlag is enabled
func (holder *epochFlagsHolder) IsESDTTransferRoleFlagEnabled() bool {
	return holder.esdtTransferRoleFlag.IsSet()
}

// IsBuiltInFunctionOnMetaFlagEnabled returns true if builtInFunctionOnMetaFlag is enabled
func (holder *epochFlagsHolder) IsBuiltInFunctionOnMetaFlagEnabled() bool {
	return holder.builtInFunctionOnMetaFlag.IsSet()
}

// IsComputeRewardCheckpointFlagEnabled returns true if computeRewardCheckpointFlag is enabled
func (holder *epochFlagsHolder) IsComputeRewardCheckpointFlagEnabled() bool {
	return holder.computeRewardCheckpointFlag.IsSet()
}

// IsSCRSizeInvariantCheckFlagEnabled returns true if scrSizeInvariantCheckFlag is enabled
func (holder *epochFlagsHolder) IsSCRSizeInvariantCheckFlagEnabled() bool {
	return holder.scrSizeInvariantCheckFlag.IsSet()
}

// IsBackwardCompSaveKeyValueFlagEnabled returns true if backwardCompSaveKeyValueFlag is enabled
func (holder *epochFlagsHolder) IsBackwardCompSaveKeyValueFlagEnabled() bool {
	return holder.backwardCompSaveKeyValueFlag.IsSet()
}

// IsESDTNFTCreateOnMultiShardFlagEnabled returns true if esdtNFTCreateOnMultiShardFlag is enabled
func (holder *epochFlagsHolder) IsESDTNFTCreateOnMultiShardFlagEnabled() bool {
	return holder.esdtNFTCreateOnMultiShardFlag.IsSet()
}

// IsMetaESDTSetFlagEnabled returns true if metaESDTSetFlag is enabled
func (holder *epochFlagsHolder) IsMetaESDTSetFlagEnabled() bool {
	return holder.metaESDTSetFlag.IsSet()
}

// IsAddTokensToDelegationFlagEnabled returns true if addTokensToDelegationFlag is enabled
func (holder *epochFlagsHolder) IsAddTokensToDelegationFlagEnabled() bool {
	return holder.addTokensToDelegationFlag.IsSet()
}

// IsMultiESDTTransferFixOnCallBackFlagEnabled returns true if multiESDTTransferFixOnCallBackFlag is enabled
func (holder *epochFlagsHolder) IsMultiESDTTransferFixOnCallBackFlagEnabled() bool {
	return holder.multiESDTTransferFixOnCallBackFlag.IsSet()
}

// IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled returns true if optimizeGasUsedInCrossMiniBlocksFlag is enabled
func (holder *epochFlagsHolder) IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled() bool {
	return holder.optimizeGasUsedInCrossMiniBlocksFlag.IsSet()
}

// IsCorrectFirstQueuedFlagEnabled returns true if correctFirstQueuedFlag is enabled
func (holder *epochFlagsHolder) IsCorrectFirstQueuedFlagEnabled() bool {
	return holder.correctFirstQueuedFlag.IsSet()
}

// IsDeleteDelegatorAfterClaimRewardsFlagEnabled returns true if deleteDelegatorAfterClaimRewardsFlag is enabled
func (holder *epochFlagsHolder) IsDeleteDelegatorAfterClaimRewardsFlagEnabled() bool {
	return holder.deleteDelegatorAfterClaimRewardsFlag.IsSet()
}

// IsFixOOGReturnCodeFlagEnabled returns true if fixOOGReturnCodeFlag is enabled
func (holder *epochFlagsHolder) IsFixOOGReturnCodeFlagEnabled() bool {
	return holder.fixOOGReturnCodeFlag.IsSet()
}

// IsRemoveNonUpdatedStorageFlagEnabled returns true if removeNonUpdatedStorageFlag is enabled
func (holder *epochFlagsHolder) IsRemoveNonUpdatedStorageFlagEnabled() bool {
	return holder.removeNonUpdatedStorageFlag.IsSet()
}

// IsOptimizeNFTStoreFlagEnabled returns true if removeNonUpdatedStorageFlag is enabled
func (holder *epochFlagsHolder) IsOptimizeNFTStoreFlagEnabled() bool {
	return holder.optimizeNFTStoreFlag.IsSet()
}

// IsCreateNFTThroughExecByCallerFlagEnabled returns true if createNFTThroughExecByCallerFlag is enabled
func (holder *epochFlagsHolder) IsCreateNFTThroughExecByCallerFlagEnabled() bool {
	return holder.createNFTThroughExecByCallerFlag.IsSet()
}

// IsStopDecreasingValidatorRatingWhenStuckFlagEnabled returns true if stopDecreasingValidatorRatingWhenStuckFlag is enabled
func (holder *epochFlagsHolder) IsStopDecreasingValidatorRatingWhenStuckFlagEnabled() bool {
	return holder.stopDecreasingValidatorRatingWhenStuckFlag.IsSet()
}

// IsFrontRunningProtectionFlagEnabled returns true if frontRunningProtectionFlag is enabled
func (holder *epochFlagsHolder) IsFrontRunningProtectionFlagEnabled() bool {
	return holder.frontRunningProtectionFlag.IsSet()
}

// IsPayableBySCFlagEnabled returns true if isPayableBySCFlag is enabled
func (holder *epochFlagsHolder) IsPayableBySCFlagEnabled() bool {
	return holder.isPayableBySCFlag.IsSet()
}

// IsCleanUpInformativeSCRsFlagEnabled returns true if cleanUpInformativeSCRsFlag is enabled
func (holder *epochFlagsHolder) IsCleanUpInformativeSCRsFlagEnabled() bool {
	return holder.cleanUpInformativeSCRsFlag.IsSet()
}

// IsStorageAPICostOptimizationFlagEnabled returns true if storageAPICostOptimizationFlag is enabled
func (holder *epochFlagsHolder) IsStorageAPICostOptimizationFlagEnabled() bool {
	return holder.storageAPICostOptimizationFlag.IsSet()
}

// IsESDTRegisterAndSetAllRolesFlagEnabled returns true if esdtRegisterAndSetAllRolesFlag is enabled
func (holder *epochFlagsHolder) IsESDTRegisterAndSetAllRolesFlagEnabled() bool {
	return holder.esdtRegisterAndSetAllRolesFlag.IsSet()
}

// IsScheduledMiniBlocksFlagEnabled returns true if scheduledMiniBlocksFlag is enabled
func (holder *epochFlagsHolder) IsScheduledMiniBlocksFlagEnabled() bool {
	return holder.scheduledMiniBlocksFlag.IsSet()
}

// IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled returns true if correctJailedNotUnStakedEmptyQueueFlag is enabled
func (holder *epochFlagsHolder) IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled() bool {
	return holder.correctJailedNotUnStakedEmptyQueueFlag.IsSet()
}

// IsDoNotReturnOldBlockInBlockchainHookFlagEnabled returns true if doNotReturnOldBlockInBlockchainHookFlag is enabled
func (holder *epochFlagsHolder) IsDoNotReturnOldBlockInBlockchainHookFlagEnabled() bool {
	return holder.doNotReturnOldBlockInBlockchainHookFlag.IsSet()
}

// IsAddFailedRelayedTxToInvalidMBsFlag returns true if addFailedRelayedTxToInvalidMBsFlag is enabled
func (holder *epochFlagsHolder) IsAddFailedRelayedTxToInvalidMBsFlag() bool {
	return holder.addFailedRelayedTxToInvalidMBsFlag.IsSet()
}

// IsSCRSizeInvariantOnBuiltInResultFlagEnabled returns true if scrSizeInvariantOnBuiltInResultFlag is enabled
func (holder *epochFlagsHolder) IsSCRSizeInvariantOnBuiltInResultFlagEnabled() bool {
	return holder.scrSizeInvariantOnBuiltInResultFlag.IsSet()
}

// IsCheckCorrectTokenIDForTransferRoleFlagEnabled returns true if checkCorrectTokenIDForTransferRoleFlag is enabled
func (holder *epochFlagsHolder) IsCheckCorrectTokenIDForTransferRoleFlagEnabled() bool {
	return holder.checkCorrectTokenIDForTransferRoleFlag.IsSet()
}

// IsFailExecutionOnEveryAPIErrorFlagEnabled returns true if failExecutionOnEveryAPIErrorFlag is enabled
func (holder *epochFlagsHolder) IsFailExecutionOnEveryAPIErrorFlagEnabled() bool {
	return holder.failExecutionOnEveryAPIErrorFlag.IsSet()
}

// IsMiniBlockPartialExecutionFlagEnabled returns true if isMiniBlockPartialExecutionFlag is enabled
func (holder *epochFlagsHolder) IsMiniBlockPartialExecutionFlagEnabled() bool {
	return holder.isMiniBlockPartialExecutionFlag.IsSet()
}

// IsManagedCryptoAPIsFlagEnabled returns true if managedCryptoAPIsFlag is enabled
func (holder *epochFlagsHolder) IsManagedCryptoAPIsFlagEnabled() bool {
	return holder.managedCryptoAPIsFlag.IsSet()
}

// IsESDTMetadataContinuousCleanupFlagEnabled returns true if esdtMetadataContinuousCleanupFlag is enabled
func (holder *epochFlagsHolder) IsESDTMetadataContinuousCleanupFlagEnabled() bool {
	return holder.esdtMetadataContinuousCleanupFlag.IsSet()
}

// IsDisableExecByCallerFlagEnabled returns true if disableExecByCallerFlag is enabled
func (holder *epochFlagsHolder) IsDisableExecByCallerFlagEnabled() bool {
	return holder.disableExecByCallerFlag.IsSet()
}

// IsRefactorContextFlagEnabled returns true if refactorContextFlag is enabled
func (holder *epochFlagsHolder) IsRefactorContextFlagEnabled() bool {
	return holder.refactorContextFlag.IsSet()
}

// IsCheckFunctionArgumentFlagEnabled returns true if checkFunctionArgumentFlag is enabled
func (holder *epochFlagsHolder) IsCheckFunctionArgumentFlagEnabled() bool {
	return holder.checkFunctionArgumentFlag.IsSet()
}

// IsCheckExecuteOnReadOnlyFlagEnabled returns true if checkExecuteOnReadOnlyFlag is enabled
func (holder *epochFlagsHolder) IsCheckExecuteOnReadOnlyFlagEnabled() bool {
	return holder.checkExecuteOnReadOnlyFlag.IsSet()
}

// IsSetSenderInEeiOutputTransferFlagEnabled returns true if setSenderInEeiOutputTransferFlag is enabled
func (holder *epochFlagsHolder) IsSetSenderInEeiOutputTransferFlagEnabled() bool {
	return holder.setSenderInEeiOutputTransferFlag.IsSet()
}

// IsFixAsyncCallbackCheckFlagEnabled returns true if esdtMetadataContinuousCleanupFlag is enabled
// this is a duplicate for ESDTMetadataContinuousCleanupEnableEpoch needed for consistency into vm-common
func (holder *epochFlagsHolder) IsFixAsyncCallbackCheckFlagEnabled() bool {
	return holder.esdtMetadataContinuousCleanupFlag.IsSet()
}

// IsSaveToSystemAccountFlagEnabled returns true if optimizeNFTStoreFlag is enabled
// this is a duplicate for OptimizeNFTStoreEnableEpoch needed for consistency into vm-common
func (holder *epochFlagsHolder) IsSaveToSystemAccountFlagEnabled() bool {
	return holder.optimizeNFTStoreFlag.IsSet()
}

// IsCheckFrozenCollectionFlagEnabled returns true if optimizeNFTStoreFlag is enabled
// this is a duplicate for OptimizeNFTStoreEnableEpoch needed for consistency into vm-common
func (holder *epochFlagsHolder) IsCheckFrozenCollectionFlagEnabled() bool {
	return holder.optimizeNFTStoreFlag.IsSet()
}

// IsSendAlwaysFlagEnabled returns true if esdtMetadataContinuousCleanupFlag is enabled
// this is a duplicate for ESDTMetadataContinuousCleanupEnableEpoch needed for consistency into vm-common
func (holder *epochFlagsHolder) IsSendAlwaysFlagEnabled() bool {
	return holder.esdtMetadataContinuousCleanupFlag.IsSet()
}

// IsValueLengthCheckFlagEnabled returns true if optimizeNFTStoreFlag is enabled
// this is a duplicate for OptimizeNFTStoreEnableEpoch needed for consistency into vm-common
func (holder *epochFlagsHolder) IsValueLengthCheckFlagEnabled() bool {
	return holder.optimizeNFTStoreFlag.IsSet()
}

// IsCheckTransferFlagEnabled returns true if optimizeNFTStoreFlag is enabled
// this is a duplicate for OptimizeNFTStoreEnableEpoch needed for consistency into vm-common
func (holder *epochFlagsHolder) IsCheckTransferFlagEnabled() bool {
	return holder.optimizeNFTStoreFlag.IsSet()
}

// IsTransferToMetaFlagEnabled returns true if builtInFunctionOnMetaFlag is enabled
// this is a duplicate for BuiltInFunctionOnMetaEnableEpoch needed for consistency into vm-common
func (holder *epochFlagsHolder) IsTransferToMetaFlagEnabled() bool {
	return holder.builtInFunctionOnMetaFlag.IsSet()
}

// IsESDTNFTImprovementV1FlagEnabled returns true if esdtMultiTransferFlag is enabled
// this is a duplicate for ESDTMultiTransferEnableEpoch needed for consistency into vm-common
func (holder *epochFlagsHolder) IsESDTNFTImprovementV1FlagEnabled() bool {
	return holder.esdtMultiTransferFlag.IsSet()
}

// IsChangeDelegationOwnerFlagEnabled returns true if the change delegation owner feature is enabled
func (holder *epochFlagsHolder) IsChangeDelegationOwnerFlagEnabled() bool {
	return holder.changeDelegationOwnerFlag.IsSet()
}

// IsRefactorPeersMiniBlocksFlagEnabled returns true if refactorPeersMiniBlocksFlag is enabled
func (holder *epochFlagsHolder) IsRefactorPeersMiniBlocksFlagEnabled() bool {
	return holder.refactorPeersMiniBlocksFlag.IsSet()
}

// IsSCProcessorV2FlagEnabled returns true if scProcessorV2Flag is enabled
func (holder *epochFlagsHolder) IsSCProcessorV2FlagEnabled() bool {
	return holder.scProcessorV2Flag.IsSet()
}

// IsFixAsyncCallBackArgsListFlagEnabled returns true if fixAsyncCallBackArgsList is enabled
func (holder *epochFlagsHolder) IsFixAsyncCallBackArgsListFlagEnabled() bool {
	return holder.fixAsyncCallBackArgsList.IsSet()
}

// IsFixOldTokenLiquidityEnabled returns true if fixOldTokenLiquidity is enabled
func (holder *epochFlagsHolder) IsFixOldTokenLiquidityEnabled() bool {
	return holder.fixOldTokenLiquidity.IsSet()
}

// IsRuntimeMemStoreLimitEnabled returns true if runtimeMemStoreLimitFlag is enabled
func (holder *epochFlagsHolder) IsRuntimeMemStoreLimitEnabled() bool {
	return holder.runtimeMemStoreLimitFlag.IsSet()
}

// IsRuntimeCodeSizeFixEnabled returns true if runtimeCodeSizeFixFlag is enabled
func (holder *epochFlagsHolder) IsRuntimeCodeSizeFixEnabled() bool {
	return holder.runtimeCodeSizeFixFlag.IsSet()
}

// IsMaxBlockchainHookCountersFlagEnabled returns true if maxBlockchainHookCountersFlagEnabled is enabled
func (holder *epochFlagsHolder) IsMaxBlockchainHookCountersFlagEnabled() bool {
	return holder.maxBlockchainHookCountersFlag.IsSet()
}

// IsWipeSingleNFTLiquidityDecreaseEnabled returns true if wipeSingleNFTLiquidityDecreaseFlag is enabled
func (holder *epochFlagsHolder) IsWipeSingleNFTLiquidityDecreaseEnabled() bool {
	return holder.wipeSingleNFTLiquidityDecreaseFlag.IsSet()
}

// IsAlwaysSaveTokenMetaDataEnabled returns true if alwaysSaveTokenMetaDataFlag is enabled
func (holder *epochFlagsHolder) IsAlwaysSaveTokenMetaDataEnabled() bool {
	return holder.alwaysSaveTokenMetaDataFlag.IsSet()
}

// IsSetGuardianEnabled returns true if setGuardianFlag is enabled
func (holder *epochFlagsHolder) IsSetGuardianEnabled() bool {
	return holder.setGuardianFlag.IsSet()
}

// IsRelayedNonceFixEnabled returns true if relayedNonceFixFlag is enabled
func (holder *epochFlagsHolder) IsRelayedNonceFixEnabled() bool {
	return holder.relayedNonceFixFlag.IsSet()
}

// IsConsistentTokensValuesLengthCheckEnabled returns true if consistentTokensValuesCheckFlag is enabled
func (holder *epochFlagsHolder) IsConsistentTokensValuesLengthCheckEnabled() bool {
	return holder.consistentTokensValuesCheckFlag.IsSet()
}

// IsKeepExecOrderOnCreatedSCRsEnabled returns true if keepExecOrderOnCreatedSCRsFlag is enabled
func (holder *epochFlagsHolder) IsKeepExecOrderOnCreatedSCRsEnabled() bool {
	return holder.keepExecOrderOnCreatedSCRsFlag.IsSet()
}

// IsMultiClaimOnDelegationEnabled returns true if multi claim on delegation is enabled
func (holder *epochFlagsHolder) IsMultiClaimOnDelegationEnabled() bool {
	return holder.multiClaimOnDelegationFlag.IsSet()
}

// IsChangeUsernameEnabled returns true if changeUsernameFlag is enabled
func (holder *epochFlagsHolder) IsChangeUsernameEnabled() bool {
	return holder.changeUsernameFlag.IsSet()
}

// IsAutoBalanceDataTriesEnabled returns true if autoBalanceDataTriesFlag is enabled
func (holder *epochFlagsHolder) IsAutoBalanceDataTriesEnabled() bool {
	return holder.autoBalanceDataTriesFlag.IsSet()
}

// FixDelegationChangeOwnerOnAccountEnabled returns true if the fix for the delegation change owner on account is enabled
func (holder *epochFlagsHolder) FixDelegationChangeOwnerOnAccountEnabled() bool {
	return holder.fixDelegationChangeOwnerOnAccountFlag.IsSet()
}
