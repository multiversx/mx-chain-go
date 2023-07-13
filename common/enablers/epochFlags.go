package enablers

import (
	"github.com/multiversx/mx-chain-core-go/core/atomic"
)

type epochFlagsHolder struct {
	refactorPeersMiniBlocksFlag           *atomic.Flag
	scProcessorV2Flag                     *atomic.Flag
	fixAsyncCallBackArgsList              *atomic.Flag
	fixOldTokenLiquidity                  *atomic.Flag
	runtimeMemStoreLimitFlag              *atomic.Flag
	runtimeCodeSizeFixFlag                *atomic.Flag
	maxBlockchainHookCountersFlag         *atomic.Flag
	wipeSingleNFTLiquidityDecreaseFlag    *atomic.Flag
	alwaysSaveTokenMetaDataFlag           *atomic.Flag
	setGuardianFlag                       *atomic.Flag
	relayedNonceFixFlag                   *atomic.Flag
	keepExecOrderOnCreatedSCRsFlag        *atomic.Flag
	multiClaimOnDelegationFlag            *atomic.Flag
	changeUsernameFlag                    *atomic.Flag
	consistentTokensValuesCheckFlag       *atomic.Flag
	autoBalanceDataTriesFlag              *atomic.Flag
	fixDelegationChangeOwnerOnAccountFlag *atomic.Flag
}

func newEpochFlagsHolder() *epochFlagsHolder {
	return &epochFlagsHolder{
		refactorPeersMiniBlocksFlag:           &atomic.Flag{},
		scProcessorV2Flag:                     &atomic.Flag{},
		fixAsyncCallBackArgsList:              &atomic.Flag{},
		fixOldTokenLiquidity:                  &atomic.Flag{},
		runtimeMemStoreLimitFlag:              &atomic.Flag{},
		runtimeCodeSizeFixFlag:                &atomic.Flag{},
		maxBlockchainHookCountersFlag:         &atomic.Flag{},
		wipeSingleNFTLiquidityDecreaseFlag:    &atomic.Flag{},
		alwaysSaveTokenMetaDataFlag:           &atomic.Flag{},
		setGuardianFlag:                       &atomic.Flag{},
		relayedNonceFixFlag:                   &atomic.Flag{},
		keepExecOrderOnCreatedSCRsFlag:        &atomic.Flag{},
		consistentTokensValuesCheckFlag:       &atomic.Flag{},
		multiClaimOnDelegationFlag:            &atomic.Flag{},
		changeUsernameFlag:                    &atomic.Flag{},
		autoBalanceDataTriesFlag:              &atomic.Flag{},
		fixDelegationChangeOwnerOnAccountFlag: &atomic.Flag{},
	}
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
