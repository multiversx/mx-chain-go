package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

type virtualAccountRecord struct {
	initialNonce   core.OptionalUint64
	virtualBalance *virtualAccountBalance
}

func newVirtualAccountRecord(initialNonce core.OptionalUint64, initialBalance *big.Int) *virtualAccountRecord {
	return &virtualAccountRecord{
		initialNonce:   initialNonce,
		virtualBalance: newVirtualAccountBalance(initialBalance),
	}
}

// TODO refactor this path, split into more functions
func (virtualRecord *virtualAccountRecord) updateVirtualRecord(breadcrumb *accountBreadcrumb) {
	virtualRecord.virtualBalance.accumulateConsumedBalance(breadcrumb)

	if !virtualRecord.initialNonce.HasValue {
		virtualRecord.initialNonce = breadcrumb.lastNonce
	}

	if breadcrumb.initialNonce.HasValue {
		virtualRecord.initialNonce = core.OptionalUint64{
			Value:    max(breadcrumb.lastNonce.Value, virtualRecord.initialNonce.Value) + 1,
			HasValue: true,
		}
	}
}

func (virtualRecord *virtualAccountRecord) getInitialBalance() *big.Int {
	return virtualRecord.virtualBalance.getInitialBalance()
}

func (virtualRecord *virtualAccountRecord) getConsumedBalance() *big.Int {
	return virtualRecord.virtualBalance.getConsumedBalance()
}
