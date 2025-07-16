package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

type accountBreadcrumb struct {
	initialNonce    core.OptionalUint64
	lastNonce       core.OptionalUint64
	consumedBalance *big.Int
}

func newAccountBreadcrumb(
	initialNonce core.OptionalUint64,
	lastNonce core.OptionalUint64,
	consumedBalance *big.Int,
) *accountBreadcrumb {
	if consumedBalance == nil {
		consumedBalance = big.NewInt(0)
	}
	return &accountBreadcrumb{
		initialNonce:    initialNonce,
		lastNonce:       lastNonce,
		consumedBalance: consumedBalance,
	}
}

func (breadcrumb *accountBreadcrumb) accumulateConsumedBalance(transferredValue *big.Int) {
	if transferredValue != nil {
		_ = breadcrumb.consumedBalance.Add(breadcrumb.consumedBalance, transferredValue)
	}
}

func (breadcrumb *accountBreadcrumb) updateNonce(lastNonce core.OptionalUint64) {
	if lastNonce.HasValue {
		breadcrumb.lastNonce = lastNonce
	}
}

func (breadcrumb *accountBreadcrumb) verifyContinuityBetweenAccountBreadcrumbs(
	previousBreadcrumbAsSender *accountBreadcrumb,
) bool {
	return previousBreadcrumbAsSender == nil || previousBreadcrumbAsSender.lastNonce.Value+1 == breadcrumb.initialNonce.Value
}

func (breadcrumb *accountBreadcrumb) verifyContinuityWithSessionNonce(sessionNonce uint64) bool {
	return breadcrumb.initialNonce.Value == sessionNonce
}

func (breadcrumb *accountBreadcrumb) hasUnkownNonce() bool {
	return !breadcrumb.initialNonce.HasValue && !breadcrumb.lastNonce.HasValue
}
