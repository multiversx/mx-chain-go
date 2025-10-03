package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

type accountBreadcrumb struct {
	firstNonce      core.OptionalUint64
	lastNonce       core.OptionalUint64
	consumedBalance *big.Int
}

func newAccountBreadcrumb(
	initialNonce core.OptionalUint64,
) *accountBreadcrumb {
	return &accountBreadcrumb{
		firstNonce:      initialNonce,
		lastNonce:       core.OptionalUint64{HasValue: false},
		consumedBalance: big.NewInt(0),
	}
}

func (breadcrumb *accountBreadcrumb) accumulateConsumedBalance(transferredValue *big.Int) {
	if transferredValue != nil {
		_ = breadcrumb.consumedBalance.Add(breadcrumb.consumedBalance, transferredValue)
	}
}

func (breadcrumb *accountBreadcrumb) updateFirstNonce(initialNonce core.OptionalUint64) {
	breadcrumb.firstNonce = initialNonce
}

func (breadcrumb *accountBreadcrumb) updateNonceRange(lastNonce core.OptionalUint64) error {
	if !lastNonce.HasValue {
		return errReceivedLastNonceNotSet
	}
	if breadcrumb.lastNonce.HasValue && breadcrumb.lastNonce.Value+1 != lastNonce.Value {
		// validate that we have txs with sequential nonces inside the tracked block used for breadcrumbs
		return errNonceGap
	}

	// if the account was previously a relayer, the first nonce should not remain unset
	if !breadcrumb.firstNonce.HasValue {
		breadcrumb.updateFirstNonce(lastNonce)
	}

	breadcrumb.lastNonce = lastNonce
	return nil
}

func (breadcrumb *accountBreadcrumb) verifyContinuityBetweenAccountBreadcrumbs(previousBreadcrumbOfSender *accountBreadcrumb) bool {
	return previousBreadcrumbOfSender == nil || previousBreadcrumbOfSender.lastNonce.Value+1 == breadcrumb.firstNonce.Value
}

func (breadcrumb *accountBreadcrumb) verifyContinuityWithSessionNonce(sessionNonce uint64) bool {
	return breadcrumb.firstNonce.Value == sessionNonce
}

func (breadcrumb *accountBreadcrumb) hasUnknownNonce() bool {
	return !breadcrumb.firstNonce.HasValue && !breadcrumb.lastNonce.HasValue
}
