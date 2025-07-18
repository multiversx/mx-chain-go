package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

type virtualAccountRecord struct {
	initialNonce    core.OptionalUint64
	initialBalance  *big.Int
	consumedBalance *big.Int
}

func newVirtualAccountRecord(initialNonce core.OptionalUint64, initialBalance *big.Int) *virtualAccountRecord {
	return &virtualAccountRecord{
		initialNonce:    initialNonce,
		initialBalance:  initialBalance,
		consumedBalance: big.NewInt(0),
	}
}

// TODO refactor this path, split into more functions
func (virtualRecord *virtualAccountRecord) updateVirtualRecord(breadcrumb *accountBreadcrumb) {
	_ = virtualRecord.consumedBalance.Add(virtualRecord.consumedBalance, breadcrumb.consumedBalance)

	if !breadcrumb.lastNonce.HasValue {
		return
	}

	if !virtualRecord.initialNonce.HasValue {
		virtualRecord.setInitialNonce(breadcrumb)
		return
	}

	virtualRecord.updateInitialNonce(breadcrumb)
}

func (virtualRecord *virtualAccountRecord) setInitialNonce(breadcrumb *accountBreadcrumb) {
	virtualRecord.initialNonce = core.OptionalUint64{
		Value:    breadcrumb.lastNonce.Value + 1,
		HasValue: true,
	}
}

func (virtualRecord *virtualAccountRecord) updateInitialNonce(breadcrumb *accountBreadcrumb) {
	virtualRecord.initialNonce = core.OptionalUint64{
		Value:    max(breadcrumb.lastNonce.Value, virtualRecord.initialNonce.Value) + 1,
		HasValue: true,
	}
}
