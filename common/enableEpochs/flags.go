package enableEpochs

import "github.com/ElrondNetwork/elrond-go-core/core/atomic"

type flagsHolder struct {
	scDeployFlag                                *atomic.Flag
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
	heartbeatDisableFlag                        *atomic.Flag
	isMiniBlockPartialExecutionFlag             *atomic.Flag
}

func newFlagsHolder() *flagsHolder {
	return &flagsHolder{
		scDeployFlag:                                &atomic.Flag{},
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
		heartbeatDisableFlag:                        &atomic.Flag{},
		isMiniBlockPartialExecutionFlag:             &atomic.Flag{},
	}
}

// IsSCDeployFlagEnabled returns true if scDeployFlag is enabled
func (fh *flagsHolder) IsSCDeployFlagEnabled() bool {
	return fh.scDeployFlag.IsSet()
}

// IsBuiltInFunctionsFlagEnabled returns true if builtInFunctionsFlag is enabled
func (fh *flagsHolder) IsBuiltInFunctionsFlagEnabled() bool {
	return fh.builtInFunctionsFlag.IsSet()
}

// IsRelayedTransactionsFlagEnabled returns true if relayedTransactionsFlag is enabled
func (fh *flagsHolder) IsRelayedTransactionsFlagEnabled() bool {
	return fh.relayedTransactionsFlag.IsSet()
}

// IsPenalizedTooMuchGasFlagEnabled returns true if penalizedTooMuchGasFlag is enabled
func (fh *flagsHolder) IsPenalizedTooMuchGasFlagEnabled() bool {
	return fh.penalizedTooMuchGasFlag.IsSet()
}

// ResetPenalizedTooMuchGasFlag resets the penalizedTooMuchGasFlag
func (fh *flagsHolder) ResetPenalizedTooMuchGasFlag() {
	fh.penalizedTooMuchGasFlag.Reset()
}

// IsSwitchJailWaitingFlagEnabled returns true if switchJailWaitingFlag is enabled
func (fh *flagsHolder) IsSwitchJailWaitingFlagEnabled() bool {
	return fh.switchJailWaitingFlag.IsSet()
}

// IsBelowSignedThresholdFlagEnabled returns true if belowSignedThresholdFlag is enabled
func (fh *flagsHolder) IsBelowSignedThresholdFlagEnabled() bool {
	return fh.belowSignedThresholdFlag.IsSet()
}

// IsSwitchHysteresisForMinNodesFlagEnabled returns true if switchHysteresisForMinNodesFlag is enabled
func (fh *flagsHolder) IsSwitchHysteresisForMinNodesFlagEnabled() bool {
	return fh.switchHysteresisForMinNodesFlag.IsSet()
}

// IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch returns true if switchHysteresisForMinNodesCurrentEpochFlag is enabled
func (fh *flagsHolder) IsSwitchHysteresisForMinNodesFlagEnabledForCurrentEpoch() bool {
	return fh.switchHysteresisForMinNodesCurrentEpochFlag.IsSet()
}

// IsTransactionSignedWithTxHashFlagEnabled returns true if transactionSignedWithTxHashFlag is enabled
func (fh *flagsHolder) IsTransactionSignedWithTxHashFlagEnabled() bool {
	return fh.transactionSignedWithTxHashFlag.IsSet()
}

// IsMetaProtectionFlagEnabled returns true if metaProtectionFlag is enabled
func (fh *flagsHolder) IsMetaProtectionFlagEnabled() bool {
	return fh.metaProtectionFlag.IsSet()
}

// IsAheadOfTimeGasUsageFlagEnabled returns true if aheadOfTimeGasUsageFlag is enabled
func (fh *flagsHolder) IsAheadOfTimeGasUsageFlagEnabled() bool {
	return fh.aheadOfTimeGasUsageFlag.IsSet()
}

// IsGasPriceModifierFlagEnabled returns true if gasPriceModifierFlag is enabled
func (fh *flagsHolder) IsGasPriceModifierFlagEnabled() bool {
	return fh.gasPriceModifierFlag.IsSet()
}

// IsRepairCallbackFlagEnabled returns true if repairCallbackFlag is enabled
func (fh *flagsHolder) IsRepairCallbackFlagEnabled() bool {
	return fh.repairCallbackFlag.IsSet()
}

// IsBalanceWaitingListsFlagEnabled returns true if balanceWaitingListsFlag is enabled
func (fh *flagsHolder) IsBalanceWaitingListsFlagEnabled() bool {
	return fh.balanceWaitingListsFlag.IsSet()
}

// IsReturnDataToLastTransferFlagEnabled returns true if returnDataToLastTransferFlag is enabled
func (fh *flagsHolder) IsReturnDataToLastTransferFlagEnabled() bool {
	return fh.returnDataToLastTransferFlag.IsSet()
}

// IsSenderInOutTransferFlagEnabled returns true if senderInOutTransferFlag is enabled
func (fh *flagsHolder) IsSenderInOutTransferFlagEnabled() bool {
	return fh.senderInOutTransferFlag.IsSet()
}

// IsStakeFlagEnabled returns true if stakeFlag is enabled
func (fh *flagsHolder) IsStakeFlagEnabled() bool {
	return fh.stakeFlag.IsSet()
}

// IsStakingV2FlagEnabled returns true if stakingV2Flag is enabled
func (fh *flagsHolder) IsStakingV2FlagEnabled() bool {
	return fh.stakingV2Flag.IsSet()
}

// IsStakingV2OwnerFlagEnabled returns true if stakingV2OwnerFlag is enabled
func (fh *flagsHolder) IsStakingV2OwnerFlagEnabled() bool {
	return fh.stakingV2OwnerFlag.IsSet()
}

// IsStakingV2FlagEnabledForActivationEpochCompleted returns true if stakingV2GreaterEpochFlag is enabled (epoch is greater than the one used for staking v2 activation)
func (fh *flagsHolder) IsStakingV2FlagEnabledForActivationEpochCompleted() bool {
	return fh.stakingV2GreaterEpochFlag.IsSet()
}

// IsDoubleKeyProtectionFlagEnabled returns true if doubleKeyProtectionFlag is enabled
func (fh *flagsHolder) IsDoubleKeyProtectionFlagEnabled() bool {
	return fh.doubleKeyProtectionFlag.IsSet()
}

// IsESDTFlagEnabled returns true if esdtFlag is enabled
func (fh *flagsHolder) IsESDTFlagEnabled() bool {
	return fh.esdtFlag.IsSet()
}

// IsESDTFlagEnabledForCurrentEpoch returns true if esdtCurrentEpochFlag is enabled
func (fh *flagsHolder) IsESDTFlagEnabledForCurrentEpoch() bool {
	return fh.esdtCurrentEpochFlag.IsSet()
}

// IsGovernanceFlagEnabled returns true if governanceFlag is enabled
func (fh *flagsHolder) IsGovernanceFlagEnabled() bool {
	return fh.governanceFlag.IsSet()
}

// IsGovernanceFlagEnabledForCurrentEpoch returns true if governanceCurrentEpochFlag is enabled
func (fh *flagsHolder) IsGovernanceFlagEnabledForCurrentEpoch() bool {
	return fh.governanceCurrentEpochFlag.IsSet()
}

// IsDelegationManagerFlagEnabled returns true if delegationManagerFlag is enabled
func (fh *flagsHolder) IsDelegationManagerFlagEnabled() bool {
	return fh.delegationManagerFlag.IsSet()
}

// IsDelegationSmartContractFlagEnabled returns true if delegationSmartContractFlag is enabled
func (fh *flagsHolder) IsDelegationSmartContractFlagEnabled() bool {
	return fh.delegationSmartContractFlag.IsSet()
}

// IsDelegationSmartContractFlagEnabledForCurrentEpoch returns true if delegationSmartContractCurrentEpochFlag is enabled
func (fh *flagsHolder) IsDelegationSmartContractFlagEnabledForCurrentEpoch() bool {
	return fh.delegationSmartContractCurrentEpochFlag.IsSet()
}

// IsCorrectLastUnJailedFlagEnabled returns true if correctLastUnJailedFlag is enabled
func (fh *flagsHolder) IsCorrectLastUnJailedFlagEnabled() bool {
	return fh.correctLastUnJailedFlag.IsSet()
}

// IsCorrectLastUnJailedFlagEnabledForCurrentEpoch returns true if correctLastUnJailedCurrentEpochFlag is enabled
func (fh *flagsHolder) IsCorrectLastUnJailedFlagEnabledForCurrentEpoch() bool {
	return fh.correctLastUnJailedCurrentEpochFlag.IsSet()
}

// IsRelayedTransactionsV2FlagEnabled returns true if relayedTransactionsV2Flag is enabled
func (fh *flagsHolder) IsRelayedTransactionsV2FlagEnabled() bool {
	return fh.relayedTransactionsV2Flag.IsSet()
}

// IsUnBondTokensV2FlagEnabled returns true if unBondTokensV2Flag is enabled
func (fh *flagsHolder) IsUnBondTokensV2FlagEnabled() bool {
	return fh.unBondTokensV2Flag.IsSet()
}

// IsSaveJailedAlwaysFlagEnabled returns true if saveJailedAlwaysFlag is enabled
func (fh *flagsHolder) IsSaveJailedAlwaysFlagEnabled() bool {
	return fh.saveJailedAlwaysFlag.IsSet()
}

// IsReDelegateBelowMinCheckFlagEnabled returns true if reDelegateBelowMinCheckFlag is enabled
func (fh *flagsHolder) IsReDelegateBelowMinCheckFlagEnabled() bool {
	return fh.reDelegateBelowMinCheckFlag.IsSet()
}

// IsValidatorToDelegationFlagEnabled returns true if validatorToDelegationFlag is enabled
func (fh *flagsHolder) IsValidatorToDelegationFlagEnabled() bool {
	return fh.validatorToDelegationFlag.IsSet()
}

// IsWaitingListFixFlagEnabled returns true if waitingListFixFlag is enabled
func (fh *flagsHolder) IsWaitingListFixFlagEnabled() bool {
	return fh.waitingListFixFlag.IsSet()
}

// IsIncrementSCRNonceInMultiTransferFlagEnabled returns true if incrementSCRNonceInMultiTransferFlag is enabled
func (fh *flagsHolder) IsIncrementSCRNonceInMultiTransferFlagEnabled() bool {
	return fh.incrementSCRNonceInMultiTransferFlag.IsSet()
}

// IsESDTMultiTransferFlagEnabled returns true if esdtMultiTransferFlag is enabled
func (fh *flagsHolder) IsESDTMultiTransferFlagEnabled() bool {
	return fh.esdtMultiTransferFlag.IsSet()
}

// IsGlobalMintBurnFlagEnabled returns true if globalMintBurnFlag is enabled
func (fh *flagsHolder) IsGlobalMintBurnFlagEnabled() bool {
	return fh.globalMintBurnFlag.IsSet()
}

// IsESDTTransferRoleFlagEnabled returns true if esdtTransferRoleFlag is enabled
func (fh *flagsHolder) IsESDTTransferRoleFlagEnabled() bool {
	return fh.esdtTransferRoleFlag.IsSet()
}

// IsBuiltInFunctionOnMetaFlagEnabled returns true if builtInFunctionOnMetaFlag is enabled
func (fh *flagsHolder) IsBuiltInFunctionOnMetaFlagEnabled() bool {
	return fh.builtInFunctionOnMetaFlag.IsSet()
}

// IsComputeRewardCheckpointFlagEnabled returns true if computeRewardCheckpointFlag is enabled
func (fh *flagsHolder) IsComputeRewardCheckpointFlagEnabled() bool {
	return fh.computeRewardCheckpointFlag.IsSet()
}

// IsSCRSizeInvariantCheckFlagEnabled returns true if scrSizeInvariantCheckFlag is enabled
func (fh *flagsHolder) IsSCRSizeInvariantCheckFlagEnabled() bool {
	return fh.scrSizeInvariantCheckFlag.IsSet()
}

// IsBackwardCompSaveKeyValueFlagEnabled returns true if backwardCompSaveKeyValueFlag is enabled
func (fh *flagsHolder) IsBackwardCompSaveKeyValueFlagEnabled() bool {
	return fh.backwardCompSaveKeyValueFlag.IsSet()
}

// IsESDTNFTCreateOnMultiShardFlagEnabled returns true if esdtNFTCreateOnMultiShardFlag is enabled
func (fh *flagsHolder) IsESDTNFTCreateOnMultiShardFlagEnabled() bool {
	return fh.esdtNFTCreateOnMultiShardFlag.IsSet()
}

// IsMetaESDTSetFlagEnabled returns true if metaESDTSetFlag is enabled
func (fh *flagsHolder) IsMetaESDTSetFlagEnabled() bool {
	return fh.metaESDTSetFlag.IsSet()
}

// IsAddTokensToDelegationFlagEnabled returns true if addTokensToDelegationFlag is enabled
func (fh *flagsHolder) IsAddTokensToDelegationFlagEnabled() bool {
	return fh.addTokensToDelegationFlag.IsSet()
}

// IsMultiESDTTransferFixOnCallBackFlagEnabled returns true if multiESDTTransferFixOnCallBackFlag is enabled
func (fh *flagsHolder) IsMultiESDTTransferFixOnCallBackFlagEnabled() bool {
	return fh.multiESDTTransferFixOnCallBackFlag.IsSet()
}

// IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled returns true if optimizeGasUsedInCrossMiniBlocksFlag is enabled
func (fh *flagsHolder) IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled() bool {
	return fh.optimizeGasUsedInCrossMiniBlocksFlag.IsSet()
}

// IsCorrectFirstQueuedFlagEnabled returns true if correctFirstQueuedFlag is enabled
func (fh *flagsHolder) IsCorrectFirstQueuedFlagEnabled() bool {
	return fh.correctFirstQueuedFlag.IsSet()
}

// IsDeleteDelegatorAfterClaimRewardsFlagEnabled returns true if deleteDelegatorAfterClaimRewardsFlag is enabled
func (fh *flagsHolder) IsDeleteDelegatorAfterClaimRewardsFlagEnabled() bool {
	return fh.deleteDelegatorAfterClaimRewardsFlag.IsSet()
}

// IsFixOOGReturnCodeFlagEnabled returns true if fixOOGReturnCodeFlag is enabled
func (fh *flagsHolder) IsFixOOGReturnCodeFlagEnabled() bool {
	return fh.fixOOGReturnCodeFlag.IsSet()
}

// IsRemoveNonUpdatedStorageFlagEnabled returns true if removeNonUpdatedStorageFlag is enabled
func (fh *flagsHolder) IsRemoveNonUpdatedStorageFlagEnabled() bool {
	return fh.removeNonUpdatedStorageFlag.IsSet()
}

// IsOptimizeNFTStoreFlagEnabled returns true if removeNonUpdatedStorageFlag is enabled
func (fh *flagsHolder) IsOptimizeNFTStoreFlagEnabled() bool {
	return fh.optimizeNFTStoreFlag.IsSet()
}

// IsCreateNFTThroughExecByCallerFlagEnabled returns true if createNFTThroughExecByCallerFlag is enabled
func (fh *flagsHolder) IsCreateNFTThroughExecByCallerFlagEnabled() bool {
	return fh.createNFTThroughExecByCallerFlag.IsSet()
}

// IsStopDecreasingValidatorRatingWhenStuckFlagEnabled returns true if stopDecreasingValidatorRatingWhenStuckFlag is enabled
func (fh *flagsHolder) IsStopDecreasingValidatorRatingWhenStuckFlagEnabled() bool {
	return fh.stopDecreasingValidatorRatingWhenStuckFlag.IsSet()
}

// IsFrontRunningProtectionFlagEnabled returns true if frontRunningProtectionFlag is enabled
func (fh *flagsHolder) IsFrontRunningProtectionFlagEnabled() bool {
	return fh.frontRunningProtectionFlag.IsSet()
}

// IsPayableBySCFlagEnabled returns true if isPayableBySCFlag is enabled
func (fh *flagsHolder) IsPayableBySCFlagEnabled() bool {
	return fh.isPayableBySCFlag.IsSet()
}

// IsCleanUpInformativeSCRsFlagEnabled returns true if cleanUpInformativeSCRsFlag is enabled
func (fh *flagsHolder) IsCleanUpInformativeSCRsFlagEnabled() bool {
	return fh.cleanUpInformativeSCRsFlag.IsSet()
}

// IsStorageAPICostOptimizationFlagEnabled returns true if storageAPICostOptimizationFlag is enabled
func (fh *flagsHolder) IsStorageAPICostOptimizationFlagEnabled() bool {
	return fh.storageAPICostOptimizationFlag.IsSet()
}

// IsESDTRegisterAndSetAllRolesFlagEnabled returns true if esdtRegisterAndSetAllRolesFlag is enabled
func (fh *flagsHolder) IsESDTRegisterAndSetAllRolesFlagEnabled() bool {
	return fh.esdtRegisterAndSetAllRolesFlag.IsSet()
}

// IsScheduledMiniBlocksFlagEnabled returns true if scheduledMiniBlocksFlag is enabled
func (fh *flagsHolder) IsScheduledMiniBlocksFlagEnabled() bool {
	return fh.scheduledMiniBlocksFlag.IsSet()
}

// IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled returns true if correctJailedNotUnStakedEmptyQueueFlag is enabled
func (fh *flagsHolder) IsCorrectJailedNotUnStakedEmptyQueueFlagEnabled() bool {
	return fh.correctJailedNotUnStakedEmptyQueueFlag.IsSet()
}

// IsDoNotReturnOldBlockInBlockchainHookFlagEnabled returns true if doNotReturnOldBlockInBlockchainHookFlag is enabled
func (fh *flagsHolder) IsDoNotReturnOldBlockInBlockchainHookFlagEnabled() bool {
	return fh.doNotReturnOldBlockInBlockchainHookFlag.IsSet()
}

// IsAddFailedRelayedTxToInvalidMBsFlag returns true if addFailedRelayedTxToInvalidMBsFlag is enabled
func (fh *flagsHolder) IsAddFailedRelayedTxToInvalidMBsFlag() bool {
	return fh.addFailedRelayedTxToInvalidMBsFlag.IsSet()
}

// IsSCRSizeInvariantOnBuiltInResultFlagEnabled returns true if scrSizeInvariantOnBuiltInResultFlag is enabled
func (fh *flagsHolder) IsSCRSizeInvariantOnBuiltInResultFlagEnabled() bool {
	return fh.scrSizeInvariantOnBuiltInResultFlag.IsSet()
}

// IsCheckCorrectTokenIDForTransferRoleFlagEnabled returns true if checkCorrectTokenIDForTransferRoleFlag is enabled
func (fh *flagsHolder) IsCheckCorrectTokenIDForTransferRoleFlagEnabled() bool {
	return fh.checkCorrectTokenIDForTransferRoleFlag.IsSet()
}

// IsFailExecutionOnEveryAPIErrorFlagEnabled returns true if failExecutionOnEveryAPIErrorFlag is enabled
func (fh *flagsHolder) IsFailExecutionOnEveryAPIErrorFlagEnabled() bool {
	return fh.failExecutionOnEveryAPIErrorFlag.IsSet()
}

// IsHeartbeatDisableFlagEnabled returns true if heartbeatDisableFlag is enabled
func (fh *flagsHolder) IsHeartbeatDisableFlagEnabled() bool {
	return fh.heartbeatDisableFlag.IsSet()
}

// IsMiniBlockPartialExecutionFlagEnabled returns true if isMiniBlockPartialExecutionFlag is enabled
func (fh *flagsHolder) IsMiniBlockPartialExecutionFlagEnabled() bool {
	return fh.isMiniBlockPartialExecutionFlag.IsSet()
}
