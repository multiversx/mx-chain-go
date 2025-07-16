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

func (virtualRecord *virtualAccountRecord) updateVirtualRecord(breadcrumb *accountBreadcrumb) {
	_ = virtualRecord.consumedBalance.Add(virtualRecord.consumedBalance, breadcrumb.consumedBalance)

	if !virtualRecord.initialNonce.HasValue {
		virtualRecord.initialNonce = breadcrumb.initialNonce
		return
	}

	if breadcrumb.initialNonce.HasValue {
		virtualRecord.initialNonce = core.OptionalUint64{
			Value:    max(breadcrumb.initialNonce.Value, virtualRecord.initialNonce.Value),
			HasValue: true,
		}
	}
}
