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

func createEnableEpochsConfig() config.EnableEpochs {
	return config.EnableEpochs{
		SCDeployEnableEpoch:                               1,
		BuiltInFunctionsEnableEpoch:                       2,
		RelayedTransactionsEnableEpoch:                    3,
		PenalizedTooMuchGasEnableEpoch:                    4,
		SwitchJailWaitingEnableEpoch:                      5,
		BelowSignedThresholdEnableEpoch:                   6,
		SwitchHysteresisForMinNodesEnableEpoch:            7,
		TransactionSignedWithTxHashEnableEpoch:            8,
		MetaProtectionEnableEpoch:                         9,
		AheadOfTimeGasUsageEnableEpoch:                    10,
		GasPriceModifierEnableEpoch:                       11,
		RepairCallbackEnableEpoch:                         12,
		BlockGasAndFeesReCheckEnableEpoch:                 13,
		BalanceWaitingListsEnableEpoch:                    14,
		ReturnDataToLastTransferEnableEpoch:               15,
		SenderInOutTransferEnableEpoch:                    16,
		StakeEnableEpoch:                                  17,
		StakingV2EnableEpoch:                              18,
		DoubleKeyProtectionEnableEpoch:                    19,
		ESDTEnableEpoch:                                   20,
		GovernanceEnableEpoch:                             21,
		DelegationManagerEnableEpoch:                      22,
		DelegationSmartContractEnableEpoch:                23,
		CorrectLastUnjailedEnableEpoch:                    24,
		RelayedTransactionsV2EnableEpoch:                  25,
		UnbondTokensV2EnableEpoch:                         26,
		SaveJailedAlwaysEnableEpoch:                       27,
		ReDelegateBelowMinCheckEnableEpoch:                28,
		ValidatorToDelegationEnableEpoch:                  29,
		WaitingListFixEnableEpoch:                         30,
		IncrementSCRNonceInMultiTransferEnableEpoch:       31,
		ESDTMultiTransferEnableEpoch:                      32,
		GlobalMintBurnDisableEpoch:                        33,
		ESDTTransferRoleEnableEpoch:                       34,
		BuiltInFunctionOnMetaEnableEpoch:                  35,
		ComputeRewardCheckpointEnableEpoch:                36,
		SCRSizeInvariantCheckEnableEpoch:                  37,
		BackwardCompSaveKeyValueEnableEpoch:               38,
		ESDTNFTCreateOnMultiShardEnableEpoch:              39,
		MetaESDTSetEnableEpoch:                            40,
		AddTokensToDelegationEnableEpoch:                  41,
		MultiESDTTransferFixOnCallBackOnEnableEpoch:       42,
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch:       43,
		CorrectFirstQueuedEpoch:                           44,
		DeleteDelegatorAfterClaimRewardsEnableEpoch:       45,
		FixOOGReturnCodeEnableEpoch:                       46,
		RemoveNonUpdatedStorageEnableEpoch:                47,
		OptimizeNFTStoreEnableEpoch:                       48,
		CreateNFTThroughExecByCallerEnableEpoch:           49,
		StopDecreasingValidatorRatingWhenStuckEnableEpoch: 50,
		FrontRunningProtectionEnableEpoch:                 51,
		IsPayableBySCEnableEpoch:                          52,
		CleanUpInformativeSCRsEnableEpoch:                 53,
		StorageAPICostOptimizationEnableEpoch:             54,
		TransformToMultiShardCreateEnableEpoch:            55,
		ESDTRegisterAndSetAllRolesEnableEpoch:             56,
		ScheduledMiniBlocksEnableEpoch:                    57,
		CorrectJailedNotUnstakedEmptyQueueEpoch:           58,
		DoNotReturnOldBlockInBlockchainHookEnableEpoch:    59,
		AddFailedRelayedTxToInvalidMBsDisableEpoch:        60,
		SCRSizeInvariantOnBuiltInResultEnableEpoch:        61,
		CheckCorrectTokenIDForTransferRoleEnableEpoch:     62,
		DisableExecByCallerEnableEpoch:                    63,
		RefactorContextEnableEpoch:                        64,
		FailExecutionOnEveryAPIErrorEnableEpoch:           65,
		ManagedCryptoAPIsEnableEpoch:                      66,
		CheckFunctionArgumentEnableEpoch:                  67,
		CheckExecuteOnReadOnlyEnableEpoch:                 68,
		ESDTMetadataContinuousCleanupEnableEpoch:          69,
		MiniBlockPartialExecutionEnableEpoch:              70,
		FixAsyncCallBackArgsListEnableEpoch:               71,
		FixOldTokenLiquidityEnableEpoch:                   72,
		RuntimeMemStoreLimitEnableEpoch:                   73,
		SetSenderInEeiOutputTransferEnableEpoch:           74,
		RefactorPeersMiniBlocksEnableEpoch:                75,
		MaxBlockchainHookCountersEnableEpoch:              76,
		WipeSingleNFTLiquidityDecreaseEnableEpoch:         77,
		AlwaysSaveTokenMetaDataEnableEpoch:                78,
		RuntimeCodeSizeFixEnableEpoch:                     79,
		RelayedNonceFixEnableEpoch:                        80,
		SetGuardianEnableEpoch:                            81,
		AutoBalanceDataTriesEnableEpoch:                   82,
		KeepExecOrderOnCreatedSCRsEnableEpoch:             83,
		MultiClaimOnDelegationEnableEpoch:                 84,
		ChangeUsernameEnableEpoch:                         85,
		ConsistentTokensValuesLengthCheckEnableEpoch:      86,
		FixDelegationChangeOwnerOnAccountEnableEpoch:      87,
		SCProcessorV2EnableEpoch:                          88,
		DeterministicSortOnValidatorsInfoEnableEpoch:      89,
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

func TestNewEnableEpochsHandler_Getters(t *testing.T) {
	t.Parallel()

	cfg := createEnableEpochsConfig()
	handler, _ := NewEnableEpochsHandler(cfg, &epochNotifier.EpochNotifierStub{})
	require.NotNil(t, handler)

	currentEpoch := uint32(1234)
	handler.EpochConfirmed(currentEpoch, 0)

	require.Equal(t, currentEpoch, handler.GetCurrentEpoch())
	require.Equal(t, cfg.ScheduledMiniBlocksEnableEpoch, handler.ScheduledMiniBlocksEnableEpoch())
	assert.Equal(t, cfg.BlockGasAndFeesReCheckEnableEpoch, handler.BlockGasAndFeesReCheckEnableEpoch())
	require.Equal(t, cfg.StakingV2EnableEpoch, handler.StakingV2EnableEpoch())
	require.Equal(t, cfg.SwitchJailWaitingEnableEpoch, handler.SwitchJailWaitingEnableEpoch())
	require.Equal(t, cfg.BalanceWaitingListsEnableEpoch, handler.BalanceWaitingListsEnableEpoch())
	require.Equal(t, cfg.WaitingListFixEnableEpoch, handler.WaitingListFixEnableEpoch())
	require.Equal(t, cfg.MultiESDTTransferFixOnCallBackOnEnableEpoch, handler.MultiESDTTransferAsyncCallBackEnableEpoch())
	require.Equal(t, cfg.FixOOGReturnCodeEnableEpoch, handler.FixOOGReturnCodeEnableEpoch())
	require.Equal(t, cfg.RemoveNonUpdatedStorageEnableEpoch, handler.RemoveNonUpdatedStorageEnableEpoch())
	require.Equal(t, cfg.CreateNFTThroughExecByCallerEnableEpoch, handler.CreateNFTThroughExecByCallerEnableEpoch())
	require.Equal(t, cfg.FailExecutionOnEveryAPIErrorEnableEpoch, handler.FixFailExecutionOnErrorEnableEpoch())
	require.Equal(t, cfg.ManagedCryptoAPIsEnableEpoch, handler.ManagedCryptoAPIEnableEpoch())
	require.Equal(t, cfg.DisableExecByCallerEnableEpoch, handler.DisableExecByCallerEnableEpoch())
	require.Equal(t, cfg.RefactorContextEnableEpoch, handler.RefactorContextEnableEpoch())
	require.Equal(t, cfg.CheckExecuteOnReadOnlyEnableEpoch, handler.CheckExecuteReadOnlyEnableEpoch())
	require.Equal(t, cfg.StorageAPICostOptimizationEnableEpoch, handler.StorageAPICostOptimizationEnableEpoch())
	require.Equal(t, cfg.MiniBlockPartialExecutionEnableEpoch, handler.MiniBlockPartialExecutionEnableEpoch())
	require.Equal(t, cfg.RefactorPeersMiniBlocksEnableEpoch, handler.RefactorPeersMiniBlocksEnableEpoch())
	require.Equal(t, cfg.RelayedNonceFixEnableEpoch, handler.RelayedNonceFixEnableEpoch())
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
}

func TestEnableEpochsHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var handler *enableEpochsHandler
	require.True(t, handler.IsInterfaceNil())

	handler, _ = NewEnableEpochsHandler(createEnableEpochsConfig(), &epochNotifier.EpochNotifierStub{})
	require.False(t, handler.IsInterfaceNil())
}
