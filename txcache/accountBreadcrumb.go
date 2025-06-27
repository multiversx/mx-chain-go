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

func (breadcrumb *accountBreadcrumb) updateBreadcrumb(transferredValue *big.Int, lastNonce core.OptionalUint64) {
	if transferredValue != nil {
		breadcrumb.consumedBalance.Add(breadcrumb.consumedBalance, transferredValue)
	}
	if lastNonce.HasValue {
		breadcrumb.lastNonce = lastNonce
	}
}
