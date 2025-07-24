package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

type accountBreadcrumb struct {
	initialNonce    core.OptionalUint64
	lastNonce       core.OptionalUint64
	initialBalance  *big.Int
	consumedBalance *big.Int
}

func newAccountBreadcrumb(
	initialNonce core.OptionalUint64,
	initialBalance *big.Int,
	consumedBalance *big.Int,
) *accountBreadcrumb {
	if consumedBalance == nil {
		consumedBalance = big.NewInt(0)
	}
	return &accountBreadcrumb{
		initialNonce:    initialNonce,
		lastNonce:       core.OptionalUint64{HasValue: false},
		initialBalance:  initialBalance,
		consumedBalance: consumedBalance,
	}
}

func (breadcrumb *accountBreadcrumb) accumulateConsumedBalance(transferredValue *big.Int) {
	if transferredValue != nil {
		_ = breadcrumb.consumedBalance.Add(breadcrumb.consumedBalance, transferredValue)
	}
}

func (breadcrumb *accountBreadcrumb) updateLastNonce(lastNonce core.OptionalUint64) error {
	if !lastNonce.HasValue {
		return errReceivedLastNonceNotSet
	}
	if breadcrumb.lastNonce.HasValue && breadcrumb.lastNonce.Value+1 != lastNonce.Value {
		// validate that we have txs with sequential nonces inside the tracked block used for breadcrumbs
		return errNonceGap
	}

	breadcrumb.lastNonce = lastNonce
	return nil
}

func (breadcrumb *accountBreadcrumb) verifyContinuityBetweenAccountBreadcrumbs(
	previousBreadcrumbAsSender *accountBreadcrumb,
) bool {
	return previousBreadcrumbAsSender == nil || previousBreadcrumbAsSender.lastNonce.Value+1 == breadcrumb.initialNonce.Value
}

func (breadcrumb *accountBreadcrumb) verifyContinuityWithSessionNonce(sessionNonce uint64) bool {
	return breadcrumb.initialNonce.Value == sessionNonce
}

func (breadcrumb *accountBreadcrumb) hasUnknownNonce() bool {
	return !breadcrumb.initialNonce.HasValue && !breadcrumb.lastNonce.HasValue
}
