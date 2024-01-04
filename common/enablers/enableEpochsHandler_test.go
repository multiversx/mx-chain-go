package enablers

import (
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

LEAVING BUILDING ERROR HERE TO REMEBER TO DELETE BuiltInFunctionOnMeta + WaitingListFixEnableEpoch

func createEnableEpochsConfig() config.EnableEpochs {
	return config.EnableEpochs{
		SCDeployEnableEpoch:                                      1,
		BuiltInFunctionsEnableEpoch:                              2,
		RelayedTransactionsEnableEpoch:                           3,
		PenalizedTooMuchGasEnableEpoch:                           4,
		SwitchJailWaitingEnableEpoch:                             5,
		BelowSignedThresholdEnableEpoch:                          6,
		SwitchHysteresisForMinNodesEnableEpoch:                   7,
		TransactionSignedWithTxHashEnableEpoch:                   8,
		MetaProtectionEnableEpoch:                                9,
		AheadOfTimeGasUsageEnableEpoch:                           10,
		GasPriceModifierEnableEpoch:                              11,
		RepairCallbackEnableEpoch:                                12,
		BlockGasAndFeesReCheckEnableEpoch:                        13,
		BalanceWaitingListsEnableEpoch:                           14,
		ReturnDataToLastTransferEnableEpoch:                      15,
		SenderInOutTransferEnableEpoch:                           16,
		StakeEnableEpoch:                                         17,
		StakingV2EnableEpoch:                                     18,
		DoubleKeyProtectionEnableEpoch:                           19,
		ESDTEnableEpoch:                                          20,
		GovernanceEnableEpoch:                                    21,
		DelegationManagerEnableEpoch:                             22,
		DelegationSmartContractEnableEpoch:                       23,
		CorrectLastUnjailedEnableEpoch:                           24,
		RelayedTransactionsV2EnableEpoch:                         25,
		UnbondTokensV2EnableEpoch:                                26,
		SaveJailedAlwaysEnableEpoch:                              27,
		ReDelegateBelowMinCheckEnableEpoch:                       28,
		ValidatorToDelegationEnableEpoch:                         29,
		WaitingListFixEnableEpoch:                                30,
		IncrementSCRNonceInMultiTransferEnableEpoch:              31,
		ESDTMultiTransferEnableEpoch:                             32,
		GlobalMintBurnDisableEpoch:                               33,
		ESDTTransferRoleEnableEpoch:                              34,
		BuiltInFunctionOnMetaEnableEpoch:                         35,
		ComputeRewardCheckpointEnableEpoch:                       36,
		SCRSizeInvariantCheckEnableEpoch:                         37,
		BackwardCompSaveKeyValueEnableEpoch:                      38,
		ESDTNFTCreateOnMultiShardEnableEpoch:                     39,
		MetaESDTSetEnableEpoch:                                   40,
		AddTokensToDelegationEnableEpoch:                         41,
		MultiESDTTransferFixOnCallBackOnEnableEpoch:              42,
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch:              43,
		CorrectFirstQueuedEpoch:                                  44,
		DeleteDelegatorAfterClaimRewardsEnableEpoch:              45,
		FixOOGReturnCodeEnableEpoch:                              46,
		RemoveNonUpdatedStorageEnableEpoch:                       47,
		OptimizeNFTStoreEnableEpoch:                              48,
		CreateNFTThroughExecByCallerEnableEpoch:                  49,
		StopDecreasingValidatorRatingWhenStuckEnableEpoch:        50,
		FrontRunningProtectionEnableEpoch:                        51,
		IsPayableBySCEnableEpoch:                                 52,
		CleanUpInformativeSCRsEnableEpoch:                        53,
		StorageAPICostOptimizationEnableEpoch:                    54,
		TransformToMultiShardCreateEnableEpoch:                   55,
		ESDTRegisterAndSetAllRolesEnableEpoch:                    56,
		ScheduledMiniBlocksEnableEpoch:                           57,
		CorrectJailedNotUnstakedEmptyQueueEpoch:                  58,
		DoNotReturnOldBlockInBlockchainHookEnableEpoch:           59,
		AddFailedRelayedTxToInvalidMBsDisableEpoch:               60,
		SCRSizeInvariantOnBuiltInResultEnableEpoch:               61,
		CheckCorrectTokenIDForTransferRoleEnableEpoch:            62,
		DisableExecByCallerEnableEpoch:                           63,
		RefactorContextEnableEpoch:                               64,
		FailExecutionOnEveryAPIErrorEnableEpoch:                  65,
		ManagedCryptoAPIsEnableEpoch:                             66,
		CheckFunctionArgumentEnableEpoch:                         67,
		CheckExecuteOnReadOnlyEnableEpoch:                        68,
		ESDTMetadataContinuousCleanupEnableEpoch:                 69,
		MiniBlockPartialExecutionEnableEpoch:                     70,
		FixAsyncCallBackArgsListEnableEpoch:                      71,
		FixOldTokenLiquidityEnableEpoch:                          72,
		RuntimeMemStoreLimitEnableEpoch:                          73,
		SetSenderInEeiOutputTransferEnableEpoch:                  74,
		RefactorPeersMiniBlocksEnableEpoch:                       75,
		MaxBlockchainHookCountersEnableEpoch:                     76,
		WipeSingleNFTLiquidityDecreaseEnableEpoch:                77,
		AlwaysSaveTokenMetaDataEnableEpoch:                       78,
		RuntimeCodeSizeFixEnableEpoch:                            79,
		RelayedNonceFixEnableEpoch:                               80,
		SetGuardianEnableEpoch:                                   81,
		AutoBalanceDataTriesEnableEpoch:                          82,
		KeepExecOrderOnCreatedSCRsEnableEpoch:                    83,
		MultiClaimOnDelegationEnableEpoch:                        84,
		ChangeUsernameEnableEpoch:                                85,
		ConsistentTokensValuesLengthCheckEnableEpoch:             86,
		FixDelegationChangeOwnerOnAccountEnableEpoch:             87,
		SCProcessorV2EnableEpoch:                                 88,
		DeterministicSortOnValidatorsInfoEnableEpoch:             89,
		DynamicGasCostForDataTrieStorageLoadEnableEpoch:          90,
		ScToScLogEventEnableEpoch:                                91,
		NFTStopCreateEnableEpoch:                                 92,
		FixGasRemainingForSaveKeyValueBuiltinFunctionEnableEpoch: 93,
		ChangeOwnerAddressCrossShardThroughSCEnableEpoch:         94,
		StakeLimitsEnableEpoch:                                   95,
		StakingV4Step1EnableEpoch:                                96,
		StakingV4Step2EnableEpoch:                                97,
		StakingV4Step3EnableEpoch:                                98,
	}
}

func TestNewEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil epoch notifier should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewEnableEpochsHandler(createEnableEpochsConfig(), nil)
		assert.Equal(t, process.ErrNilEpochNotifier, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		handler, err := NewEnableEpochsHandler(createEnableEpochsConfig(), &epochNotifier.EpochNotifierStub{
			RegisterNotifyHandlerCalled: func(handler vmcommon.EpochSubscriberHandler) {
				wasCalled = true
			},
		})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
		assert.True(t, wasCalled)
	})
}

func TestNewEnableEpochsHandler_GetCurrentEpoch(t *testing.T) {
	t.Parallel()

	cfg := createEnableEpochsConfig()
	handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
	require.NotNil(t, handler)

	currentEpoch := uint32(1234)
	handler.EpochConfirmed(currentEpoch, 0)

	require.Equal(t, currentEpoch, handler.GetCurrentEpoch())
}

func TestEnableEpochsHandler_IsFlagDefined(t *testing.T) {
	t.Parallel()

	cfg := createEnableEpochsConfig()
	handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
	require.NotNil(t, handler)

	require.True(t, handler.IsFlagDefined(common.SCDeployFlag))
	require.False(t, handler.IsFlagDefined("new flag"))
}

func TestEnableEpochsHandler_IsFlagEnabledInEpoch(t *testing.T) {
	t.Parallel()

	cfg := createEnableEpochsConfig()
	handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
	require.NotNil(t, handler)

	require.True(t, handler.IsFlagEnabledInEpoch(common.BuiltInFunctionsFlag, cfg.BuiltInFunctionsEnableEpoch))
	require.True(t, handler.IsFlagEnabledInEpoch(common.BuiltInFunctionsFlag, cfg.BuiltInFunctionsEnableEpoch+1))
	require.False(t, handler.IsFlagEnabledInEpoch(common.BuiltInFunctionsFlag, cfg.BuiltInFunctionsEnableEpoch-1))
	require.False(t, handler.IsFlagEnabledInEpoch("new flag", 0))
}

func TestEnableEpochsHandler_IsFlagEnabled(t *testing.T) {
	t.Parallel()

	cfg := createEnableEpochsConfig()
	handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
	require.NotNil(t, handler)

	require.False(t, handler.IsFlagEnabled(common.SetGuardianFlag))
	handler.EpochConfirmed(cfg.SetGuardianEnableEpoch, 0)
	require.True(t, handler.IsFlagEnabled(common.SetGuardianFlag))
	handler.EpochConfirmed(cfg.SetGuardianEnableEpoch+1, 0)
	require.True(t, handler.IsFlagEnabled(common.SetGuardianFlag))

	handler.EpochConfirmed(math.MaxUint32, 0)
	require.True(t, handler.IsFlagEnabled(common.SCDeployFlag))
	require.True(t, handler.IsFlagEnabled(common.BuiltInFunctionsFlag))
	require.True(t, handler.IsFlagEnabled(common.RelayedTransactionsFlag))
	require.True(t, handler.IsFlagEnabled(common.PenalizedTooMuchGasFlag))
	require.True(t, handler.IsFlagEnabled(common.SwitchJailWaitingFlag))
	require.True(t, handler.IsFlagEnabled(common.BelowSignedThresholdFlag))
	require.False(t, handler.IsFlagEnabled(common.SwitchHysteresisForMinNodesFlagInSpecificEpochOnly)) // ==
	require.True(t, handler.IsFlagEnabled(common.TransactionSignedWithTxHashFlag))
	require.True(t, handler.IsFlagEnabled(common.MetaProtectionFlag))
	require.True(t, handler.IsFlagEnabled(common.AheadOfTimeGasUsageFlag))
	require.True(t, handler.IsFlagEnabled(common.GasPriceModifierFlag))
	require.True(t, handler.IsFlagEnabled(common.RepairCallbackFlag))
	require.True(t, handler.IsFlagEnabled(common.ReturnDataToLastTransferFlagAfterEpoch))
	require.True(t, handler.IsFlagEnabled(common.SenderInOutTransferFlag))
	require.True(t, handler.IsFlagEnabled(common.StakeFlag))
	require.True(t, handler.IsFlagEnabled(common.StakingV2Flag))
	require.False(t, handler.IsFlagEnabled(common.StakingV2OwnerFlagInSpecificEpochOnly)) // ==
	require.True(t, handler.IsFlagEnabled(common.StakingV2FlagAfterEpoch))
	require.True(t, handler.IsFlagEnabled(common.DoubleKeyProtectionFlag))
	require.True(t, handler.IsFlagEnabled(common.ESDTFlag))
	require.False(t, handler.IsFlagEnabled(common.ESDTFlagInSpecificEpochOnly)) // ==
	require.True(t, handler.IsFlagEnabled(common.GovernanceFlag))
	require.False(t, handler.IsFlagEnabled(common.GovernanceFlagInSpecificEpochOnly)) // ==
	require.True(t, handler.IsFlagEnabled(common.DelegationManagerFlag))
	require.True(t, handler.IsFlagEnabled(common.DelegationSmartContractFlag))
	require.False(t, handler.IsFlagEnabled(common.DelegationSmartContractFlagInSpecificEpochOnly)) // ==
	require.False(t, handler.IsFlagEnabled(common.CorrectLastUnJailedFlagInSpecificEpochOnly))     // ==
	require.True(t, handler.IsFlagEnabled(common.CorrectLastUnJailedFlag))
	require.True(t, handler.IsFlagEnabled(common.RelayedTransactionsV2Flag))
	require.True(t, handler.IsFlagEnabled(common.UnBondTokensV2Flag))
	require.True(t, handler.IsFlagEnabled(common.SaveJailedAlwaysFlag))
	require.True(t, handler.IsFlagEnabled(common.ReDelegateBelowMinCheckFlag))
	require.True(t, handler.IsFlagEnabled(common.ValidatorToDelegationFlag))
	require.True(t, handler.IsFlagEnabled(common.IncrementSCRNonceInMultiTransferFlag))
	require.True(t, handler.IsFlagEnabled(common.ESDTMultiTransferFlag))
	require.False(t, handler.IsFlagEnabled(common.GlobalMintBurnFlag)) // <
	require.True(t, handler.IsFlagEnabled(common.ESDTTransferRoleFlag))
	require.True(t, handler.IsFlagEnabled(common.BuiltInFunctionOnMetaFlag))
	require.True(t, handler.IsFlagEnabled(common.ComputeRewardCheckpointFlag))
	require.True(t, handler.IsFlagEnabled(common.SCRSizeInvariantCheckFlag))
	require.False(t, handler.IsFlagEnabled(common.BackwardCompSaveKeyValueFlag)) // <
	require.True(t, handler.IsFlagEnabled(common.ESDTNFTCreateOnMultiShardFlag))
	require.True(t, handler.IsFlagEnabled(common.MetaESDTSetFlag))
	require.True(t, handler.IsFlagEnabled(common.AddTokensToDelegationFlag))
	require.True(t, handler.IsFlagEnabled(common.MultiESDTTransferFixOnCallBackFlag))
	require.True(t, handler.IsFlagEnabled(common.OptimizeGasUsedInCrossMiniBlocksFlag))
	require.True(t, handler.IsFlagEnabled(common.CorrectFirstQueuedFlag))
	require.True(t, handler.IsFlagEnabled(common.DeleteDelegatorAfterClaimRewardsFlag))
	require.True(t, handler.IsFlagEnabled(common.RemoveNonUpdatedStorageFlag))
	require.True(t, handler.IsFlagEnabled(common.OptimizeNFTStoreFlag))
	require.True(t, handler.IsFlagEnabled(common.CreateNFTThroughExecByCallerFlag))
	require.True(t, handler.IsFlagEnabled(common.StopDecreasingValidatorRatingWhenStuckFlag))
	require.True(t, handler.IsFlagEnabled(common.FrontRunningProtectionFlag))
	require.True(t, handler.IsFlagEnabled(common.PayableBySCFlag))
	require.True(t, handler.IsFlagEnabled(common.CleanUpInformativeSCRsFlag))
	require.True(t, handler.IsFlagEnabled(common.StorageAPICostOptimizationFlag))
	require.True(t, handler.IsFlagEnabled(common.ESDTRegisterAndSetAllRolesFlag))
	require.True(t, handler.IsFlagEnabled(common.ScheduledMiniBlocksFlag))
	require.True(t, handler.IsFlagEnabled(common.CorrectJailedNotUnStakedEmptyQueueFlag))
	require.True(t, handler.IsFlagEnabled(common.DoNotReturnOldBlockInBlockchainHookFlag))
	require.False(t, handler.IsFlagEnabled(common.AddFailedRelayedTxToInvalidMBsFlag)) // <
	require.True(t, handler.IsFlagEnabled(common.SCRSizeInvariantOnBuiltInResultFlag))
	require.True(t, handler.IsFlagEnabled(common.CheckCorrectTokenIDForTransferRoleFlag))
	require.True(t, handler.IsFlagEnabled(common.FailExecutionOnEveryAPIErrorFlag))
	require.True(t, handler.IsFlagEnabled(common.MiniBlockPartialExecutionFlag))
	require.True(t, handler.IsFlagEnabled(common.ManagedCryptoAPIsFlag))
	require.True(t, handler.IsFlagEnabled(common.ESDTMetadataContinuousCleanupFlag))
	require.True(t, handler.IsFlagEnabled(common.DisableExecByCallerFlag))
	require.True(t, handler.IsFlagEnabled(common.RefactorContextFlag))
	require.True(t, handler.IsFlagEnabled(common.CheckFunctionArgumentFlag))
	require.True(t, handler.IsFlagEnabled(common.CheckExecuteOnReadOnlyFlag))
	require.True(t, handler.IsFlagEnabled(common.SetSenderInEeiOutputTransferFlag))
	require.True(t, handler.IsFlagEnabled(common.FixAsyncCallbackCheckFlag))
	require.True(t, handler.IsFlagEnabled(common.SaveToSystemAccountFlag))
	require.True(t, handler.IsFlagEnabled(common.CheckFrozenCollectionFlag))
	require.True(t, handler.IsFlagEnabled(common.SendAlwaysFlag))
	require.True(t, handler.IsFlagEnabled(common.ValueLengthCheckFlag))
	require.True(t, handler.IsFlagEnabled(common.CheckTransferFlag))
	require.True(t, handler.IsFlagEnabled(common.TransferToMetaFlag))
	require.True(t, handler.IsFlagEnabled(common.ESDTNFTImprovementV1Flag))
	require.True(t, handler.IsFlagEnabled(common.ChangeDelegationOwnerFlag))
	require.True(t, handler.IsFlagEnabled(common.RefactorPeersMiniBlocksFlag))
	require.True(t, handler.IsFlagEnabled(common.SCProcessorV2Flag))
	require.True(t, handler.IsFlagEnabled(common.FixAsyncCallBackArgsListFlag))
	require.True(t, handler.IsFlagEnabled(common.FixOldTokenLiquidityFlag))
	require.True(t, handler.IsFlagEnabled(common.RuntimeMemStoreLimitFlag))
	require.True(t, handler.IsFlagEnabled(common.RuntimeCodeSizeFixFlag))
	require.True(t, handler.IsFlagEnabled(common.MaxBlockchainHookCountersFlag))
	require.True(t, handler.IsFlagEnabled(common.WipeSingleNFTLiquidityDecreaseFlag))
	require.True(t, handler.IsFlagEnabled(common.AlwaysSaveTokenMetaDataFlag))
	require.True(t, handler.IsFlagEnabled(common.SetGuardianFlag))
	require.True(t, handler.IsFlagEnabled(common.RelayedNonceFixFlag))
	require.True(t, handler.IsFlagEnabled(common.ConsistentTokensValuesLengthCheckFlag))
	require.True(t, handler.IsFlagEnabled(common.KeepExecOrderOnCreatedSCRsFlag))
	require.True(t, handler.IsFlagEnabled(common.MultiClaimOnDelegationFlag))
	require.True(t, handler.IsFlagEnabled(common.ChangeUsernameFlag))
	require.True(t, handler.IsFlagEnabled(common.AutoBalanceDataTriesFlag))
	require.True(t, handler.IsFlagEnabled(common.FixDelegationChangeOwnerOnAccountFlag))
	require.True(t, handler.IsFlagEnabled(common.FixOOGReturnCodeFlag))
	require.True(t, handler.IsFlagEnabled(common.DeterministicSortOnValidatorsInfoFixFlag))
	require.True(t, handler.IsFlagEnabled(common.DynamicGasCostForDataTrieStorageLoadFlag))
	require.True(t, handler.IsFlagEnabled(common.ScToScLogEventFlag))
	require.True(t, handler.IsFlagEnabled(common.BlockGasAndFeesReCheckFlag))
	require.True(t, handler.IsFlagEnabled(common.BalanceWaitingListsFlag))
	require.True(t, handler.IsFlagEnabled(common.WaitingListFixFlag))
	require.True(t, handler.IsFlagEnabled(common.NFTStopCreateFlag))
	require.True(t, handler.IsFlagEnabled(common.FixGasRemainingForSaveKeyValueFlag))
	require.True(t, handler.IsFlagEnabled(common.IsChangeOwnerAddressCrossShardThroughSCFlag))
	assert.True(t, handler.IsStakeLimitsFlagEnabled())
	assert.True(t, handler.IsStakingV4Step1Enabled())
	assert.True(t, handler.IsStakingV4Step2Enabled())
	assert.True(t, handler.IsStakingV4Step3Enabled())
	assert.False(t, handler.IsStakingQueueEnabled())
	assert.True(t, handler.IsStakingV4Started())
}

func TestEnableEpochsHandler_GetActivationEpoch(t *testing.T) {
	t.Parallel()

	cfg := createEnableEpochsConfig()
	handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
	require.NotNil(t, handler)

	require.Equal(t, uint32(0), handler.GetActivationEpoch("dummy flag"))
	require.Equal(t, cfg.SCDeployEnableEpoch, handler.GetActivationEpoch(common.SCDeployFlag))
	require.Equal(t, cfg.BuiltInFunctionsEnableEpoch, handler.GetActivationEpoch(common.BuiltInFunctionsFlag))
	require.Equal(t, cfg.RelayedTransactionsEnableEpoch, handler.GetActivationEpoch(common.RelayedTransactionsFlag))
	require.Equal(t, cfg.PenalizedTooMuchGasEnableEpoch, handler.GetActivationEpoch(common.PenalizedTooMuchGasFlag))
	require.Equal(t, cfg.SwitchJailWaitingEnableEpoch, handler.GetActivationEpoch(common.SwitchJailWaitingFlag))
	require.Equal(t, cfg.BelowSignedThresholdEnableEpoch, handler.GetActivationEpoch(common.BelowSignedThresholdFlag))
	require.Equal(t, cfg.TransactionSignedWithTxHashEnableEpoch, handler.GetActivationEpoch(common.TransactionSignedWithTxHashFlag))
	require.Equal(t, cfg.MetaProtectionEnableEpoch, handler.GetActivationEpoch(common.MetaProtectionFlag))
	require.Equal(t, cfg.AheadOfTimeGasUsageEnableEpoch, handler.GetActivationEpoch(common.AheadOfTimeGasUsageFlag))
	require.Equal(t, cfg.GasPriceModifierEnableEpoch, handler.GetActivationEpoch(common.GasPriceModifierFlag))
	require.Equal(t, cfg.RepairCallbackEnableEpoch, handler.GetActivationEpoch(common.RepairCallbackFlag))
	require.Equal(t, cfg.SenderInOutTransferEnableEpoch, handler.GetActivationEpoch(common.SenderInOutTransferFlag))
	require.Equal(t, cfg.StakeEnableEpoch, handler.GetActivationEpoch(common.StakeFlag))
	require.Equal(t, cfg.StakingV2EnableEpoch, handler.GetActivationEpoch(common.StakingV2Flag))
	require.Equal(t, cfg.DoubleKeyProtectionEnableEpoch, handler.GetActivationEpoch(common.DoubleKeyProtectionFlag))
	require.Equal(t, cfg.ESDTEnableEpoch, handler.GetActivationEpoch(common.ESDTFlag))
	require.Equal(t, cfg.GovernanceEnableEpoch, handler.GetActivationEpoch(common.GovernanceFlag))
	require.Equal(t, cfg.DelegationManagerEnableEpoch, handler.GetActivationEpoch(common.DelegationManagerFlag))
	require.Equal(t, cfg.DelegationSmartContractEnableEpoch, handler.GetActivationEpoch(common.DelegationSmartContractFlag))
	require.Equal(t, cfg.CorrectLastUnjailedEnableEpoch, handler.GetActivationEpoch(common.CorrectLastUnJailedFlag))
	require.Equal(t, cfg.RelayedTransactionsV2EnableEpoch, handler.GetActivationEpoch(common.RelayedTransactionsV2Flag))
	require.Equal(t, cfg.UnbondTokensV2EnableEpoch, handler.GetActivationEpoch(common.UnBondTokensV2Flag))
	require.Equal(t, cfg.SaveJailedAlwaysEnableEpoch, handler.GetActivationEpoch(common.SaveJailedAlwaysFlag))
	require.Equal(t, cfg.ReDelegateBelowMinCheckEnableEpoch, handler.GetActivationEpoch(common.ReDelegateBelowMinCheckFlag))
	require.Equal(t, cfg.ValidatorToDelegationEnableEpoch, handler.GetActivationEpoch(common.ValidatorToDelegationFlag))
	require.Equal(t, cfg.IncrementSCRNonceInMultiTransferEnableEpoch, handler.GetActivationEpoch(common.IncrementSCRNonceInMultiTransferFlag))
	require.Equal(t, cfg.ESDTMultiTransferEnableEpoch, handler.GetActivationEpoch(common.ESDTMultiTransferFlag))
	require.Equal(t, cfg.GlobalMintBurnDisableEpoch, handler.GetActivationEpoch(common.GlobalMintBurnFlag))
	require.Equal(t, cfg.ESDTTransferRoleEnableEpoch, handler.GetActivationEpoch(common.ESDTTransferRoleFlag))
	require.Equal(t, cfg.BuiltInFunctionOnMetaEnableEpoch, handler.GetActivationEpoch(common.BuiltInFunctionOnMetaFlag))
	require.Equal(t, cfg.ComputeRewardCheckpointEnableEpoch, handler.GetActivationEpoch(common.ComputeRewardCheckpointFlag))
	require.Equal(t, cfg.SCRSizeInvariantCheckEnableEpoch, handler.GetActivationEpoch(common.SCRSizeInvariantCheckFlag))
	require.Equal(t, cfg.BackwardCompSaveKeyValueEnableEpoch, handler.GetActivationEpoch(common.BackwardCompSaveKeyValueFlag))
	require.Equal(t, cfg.ESDTNFTCreateOnMultiShardEnableEpoch, handler.GetActivationEpoch(common.ESDTNFTCreateOnMultiShardFlag))
	require.Equal(t, cfg.MetaESDTSetEnableEpoch, handler.GetActivationEpoch(common.MetaESDTSetFlag))
	require.Equal(t, cfg.AddTokensToDelegationEnableEpoch, handler.GetActivationEpoch(common.AddTokensToDelegationFlag))
	require.Equal(t, cfg.MultiESDTTransferFixOnCallBackOnEnableEpoch, handler.GetActivationEpoch(common.MultiESDTTransferFixOnCallBackFlag))
	require.Equal(t, cfg.OptimizeGasUsedInCrossMiniBlocksEnableEpoch, handler.GetActivationEpoch(common.OptimizeGasUsedInCrossMiniBlocksFlag))
	require.Equal(t, cfg.CorrectFirstQueuedEpoch, handler.GetActivationEpoch(common.CorrectFirstQueuedFlag))
	require.Equal(t, cfg.DeleteDelegatorAfterClaimRewardsEnableEpoch, handler.GetActivationEpoch(common.DeleteDelegatorAfterClaimRewardsFlag))
	require.Equal(t, cfg.RemoveNonUpdatedStorageEnableEpoch, handler.GetActivationEpoch(common.RemoveNonUpdatedStorageFlag))
	require.Equal(t, cfg.OptimizeNFTStoreEnableEpoch, handler.GetActivationEpoch(common.OptimizeNFTStoreFlag))
	require.Equal(t, cfg.CreateNFTThroughExecByCallerEnableEpoch, handler.GetActivationEpoch(common.CreateNFTThroughExecByCallerFlag))
	require.Equal(t, cfg.StopDecreasingValidatorRatingWhenStuckEnableEpoch, handler.GetActivationEpoch(common.StopDecreasingValidatorRatingWhenStuckFlag))
	require.Equal(t, cfg.FrontRunningProtectionEnableEpoch, handler.GetActivationEpoch(common.FrontRunningProtectionFlag))
	require.Equal(t, cfg.IsPayableBySCEnableEpoch, handler.GetActivationEpoch(common.PayableBySCFlag))
	require.Equal(t, cfg.CleanUpInformativeSCRsEnableEpoch, handler.GetActivationEpoch(common.CleanUpInformativeSCRsFlag))
	require.Equal(t, cfg.StorageAPICostOptimizationEnableEpoch, handler.GetActivationEpoch(common.StorageAPICostOptimizationFlag))
	require.Equal(t, cfg.ESDTRegisterAndSetAllRolesEnableEpoch, handler.GetActivationEpoch(common.ESDTRegisterAndSetAllRolesFlag))
	require.Equal(t, cfg.ScheduledMiniBlocksEnableEpoch, handler.GetActivationEpoch(common.ScheduledMiniBlocksFlag))
	require.Equal(t, cfg.CorrectJailedNotUnstakedEmptyQueueEpoch, handler.GetActivationEpoch(common.CorrectJailedNotUnStakedEmptyQueueFlag))
	require.Equal(t, cfg.DoNotReturnOldBlockInBlockchainHookEnableEpoch, handler.GetActivationEpoch(common.DoNotReturnOldBlockInBlockchainHookFlag))
	require.Equal(t, cfg.AddFailedRelayedTxToInvalidMBsDisableEpoch, handler.GetActivationEpoch(common.AddFailedRelayedTxToInvalidMBsFlag))
	require.Equal(t, cfg.SCRSizeInvariantOnBuiltInResultEnableEpoch, handler.GetActivationEpoch(common.SCRSizeInvariantOnBuiltInResultFlag))
	require.Equal(t, cfg.CheckCorrectTokenIDForTransferRoleEnableEpoch, handler.GetActivationEpoch(common.CheckCorrectTokenIDForTransferRoleFlag))
	require.Equal(t, cfg.FailExecutionOnEveryAPIErrorEnableEpoch, handler.GetActivationEpoch(common.FailExecutionOnEveryAPIErrorFlag))
	require.Equal(t, cfg.MiniBlockPartialExecutionEnableEpoch, handler.GetActivationEpoch(common.MiniBlockPartialExecutionFlag))
	require.Equal(t, cfg.ManagedCryptoAPIsEnableEpoch, handler.GetActivationEpoch(common.ManagedCryptoAPIsFlag))
	require.Equal(t, cfg.ESDTMetadataContinuousCleanupEnableEpoch, handler.GetActivationEpoch(common.ESDTMetadataContinuousCleanupFlag))
	require.Equal(t, cfg.DisableExecByCallerEnableEpoch, handler.GetActivationEpoch(common.DisableExecByCallerFlag))
	require.Equal(t, cfg.RefactorContextEnableEpoch, handler.GetActivationEpoch(common.RefactorContextFlag))
	require.Equal(t, cfg.CheckFunctionArgumentEnableEpoch, handler.GetActivationEpoch(common.CheckFunctionArgumentFlag))
	require.Equal(t, cfg.CheckExecuteOnReadOnlyEnableEpoch, handler.GetActivationEpoch(common.CheckExecuteOnReadOnlyFlag))
	require.Equal(t, cfg.SetSenderInEeiOutputTransferEnableEpoch, handler.GetActivationEpoch(common.SetSenderInEeiOutputTransferFlag))
	require.Equal(t, cfg.ESDTMetadataContinuousCleanupEnableEpoch, handler.GetActivationEpoch(common.FixAsyncCallbackCheckFlag))
	require.Equal(t, cfg.OptimizeNFTStoreEnableEpoch, handler.GetActivationEpoch(common.SaveToSystemAccountFlag))
	require.Equal(t, cfg.OptimizeNFTStoreEnableEpoch, handler.GetActivationEpoch(common.CheckFrozenCollectionFlag))
	require.Equal(t, cfg.ESDTMetadataContinuousCleanupEnableEpoch, handler.GetActivationEpoch(common.SendAlwaysFlag))
	require.Equal(t, cfg.OptimizeNFTStoreEnableEpoch, handler.GetActivationEpoch(common.ValueLengthCheckFlag))
	require.Equal(t, cfg.OptimizeNFTStoreEnableEpoch, handler.GetActivationEpoch(common.CheckTransferFlag))
	require.Equal(t, cfg.BuiltInFunctionOnMetaEnableEpoch, handler.GetActivationEpoch(common.TransferToMetaFlag))
	require.Equal(t, cfg.ESDTMultiTransferEnableEpoch, handler.GetActivationEpoch(common.ESDTNFTImprovementV1Flag))
	require.Equal(t, cfg.ESDTMetadataContinuousCleanupEnableEpoch, handler.GetActivationEpoch(common.ChangeDelegationOwnerFlag))
	require.Equal(t, cfg.RefactorPeersMiniBlocksEnableEpoch, handler.GetActivationEpoch(common.RefactorPeersMiniBlocksFlag))
	require.Equal(t, cfg.SCProcessorV2EnableEpoch, handler.GetActivationEpoch(common.SCProcessorV2Flag))
	require.Equal(t, cfg.FixAsyncCallBackArgsListEnableEpoch, handler.GetActivationEpoch(common.FixAsyncCallBackArgsListFlag))
	require.Equal(t, cfg.FixOldTokenLiquidityEnableEpoch, handler.GetActivationEpoch(common.FixOldTokenLiquidityFlag))
	require.Equal(t, cfg.RuntimeMemStoreLimitEnableEpoch, handler.GetActivationEpoch(common.RuntimeMemStoreLimitFlag))
	require.Equal(t, cfg.RuntimeCodeSizeFixEnableEpoch, handler.GetActivationEpoch(common.RuntimeCodeSizeFixFlag))
	require.Equal(t, cfg.MaxBlockchainHookCountersEnableEpoch, handler.GetActivationEpoch(common.MaxBlockchainHookCountersFlag))
	require.Equal(t, cfg.WipeSingleNFTLiquidityDecreaseEnableEpoch, handler.GetActivationEpoch(common.WipeSingleNFTLiquidityDecreaseFlag))
	require.Equal(t, cfg.AlwaysSaveTokenMetaDataEnableEpoch, handler.GetActivationEpoch(common.AlwaysSaveTokenMetaDataFlag))
	require.Equal(t, cfg.SetGuardianEnableEpoch, handler.GetActivationEpoch(common.SetGuardianFlag))
	require.Equal(t, cfg.RelayedNonceFixEnableEpoch, handler.GetActivationEpoch(common.RelayedNonceFixFlag))
	require.Equal(t, cfg.ConsistentTokensValuesLengthCheckEnableEpoch, handler.GetActivationEpoch(common.ConsistentTokensValuesLengthCheckFlag))
	require.Equal(t, cfg.KeepExecOrderOnCreatedSCRsEnableEpoch, handler.GetActivationEpoch(common.KeepExecOrderOnCreatedSCRsFlag))
	require.Equal(t, cfg.MultiClaimOnDelegationEnableEpoch, handler.GetActivationEpoch(common.MultiClaimOnDelegationFlag))
	require.Equal(t, cfg.ChangeUsernameEnableEpoch, handler.GetActivationEpoch(common.ChangeUsernameFlag))
	require.Equal(t, cfg.AutoBalanceDataTriesEnableEpoch, handler.GetActivationEpoch(common.AutoBalanceDataTriesFlag))
	require.Equal(t, cfg.FixDelegationChangeOwnerOnAccountEnableEpoch, handler.GetActivationEpoch(common.FixDelegationChangeOwnerOnAccountFlag))
	require.Equal(t, cfg.FixOOGReturnCodeEnableEpoch, handler.GetActivationEpoch(common.FixOOGReturnCodeFlag))
	require.Equal(t, cfg.DeterministicSortOnValidatorsInfoEnableEpoch, handler.GetActivationEpoch(common.DeterministicSortOnValidatorsInfoFixFlag))
	require.Equal(t, cfg.DynamicGasCostForDataTrieStorageLoadEnableEpoch, handler.GetActivationEpoch(common.DynamicGasCostForDataTrieStorageLoadFlag))
	require.Equal(t, cfg.ScToScLogEventEnableEpoch, handler.GetActivationEpoch(common.ScToScLogEventFlag))
	require.Equal(t, cfg.BlockGasAndFeesReCheckEnableEpoch, handler.GetActivationEpoch(common.BlockGasAndFeesReCheckFlag))
	require.Equal(t, cfg.BalanceWaitingListsEnableEpoch, handler.GetActivationEpoch(common.BalanceWaitingListsFlag))
	require.Equal(t, cfg.WaitingListFixEnableEpoch, handler.GetActivationEpoch(common.WaitingListFixFlag))
	require.Equal(t, cfg.NFTStopCreateEnableEpoch, handler.GetActivationEpoch(common.NFTStopCreateFlag))
	require.Equal(t, cfg.ChangeOwnerAddressCrossShardThroughSCEnableEpoch, handler.GetActivationEpoch(common.IsChangeOwnerAddressCrossShardThroughSCFlag))
	require.Equal(t, cfg.FixGasRemainingForSaveKeyValueBuiltinFunctionEnableEpoch, handler.GetActivationEpoch(common.FixGasRemainingForSaveKeyValueFlag))
	assert.True(t, handler.IsStakeLimitsFlagEnabled())
	assert.True(t, handler.IsStakingV4Step1Enabled())
	assert.True(t, handler.IsStakingV4Step2Enabled())
	assert.True(t, handler.IsStakingV4Step3Enabled())
	assert.False(t, handler.IsStakingQueueEnabled())
	assert.True(t, handler.IsStakingV4Started())
}

func TestEnableEpochsHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var handler *enableEpochsHandler
	require.True(t, handler.IsInterfaceNil())

	handler, _ = NewEnableEpochsHandler(createEnableEpochsConfig(), &epochNotifier.EpochNotifierStub{})
	require.False(t, handler.IsInterfaceNil())
}
