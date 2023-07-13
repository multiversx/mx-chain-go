package enablers

import (
	"github.com/multiversx/mx-chain-core-go/core/atomic"
)

type epochFlagsHolder struct {
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
		relayedNonceFixFlag:                   &atomic.Flag{},
		keepExecOrderOnCreatedSCRsFlag:        &atomic.Flag{},
		consistentTokensValuesCheckFlag:       &atomic.Flag{},
		multiClaimOnDelegationFlag:            &atomic.Flag{},
		changeUsernameFlag:                    &atomic.Flag{},
		autoBalanceDataTriesFlag:              &atomic.Flag{},
		fixDelegationChangeOwnerOnAccountFlag: &atomic.Flag{},
	}
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
