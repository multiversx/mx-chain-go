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

// updateVirtualRecord updates the virtualBalance of a virtualAccountRecord and handles the nonces
func (virtualRecord *virtualAccountRecord) updateVirtualRecord(breadcrumb *accountBreadcrumb) {
	virtualRecord.virtualBalance.accumulateConsumedBalance(breadcrumb)

	if !breadcrumb.lastNonce.HasValue {
		// in a certain tracked block, we can have breadcrumbs for accounts that are only used as relayers and never as senders
		// those accounts don't have a set nonce, here we treat those cases
		return
	}

	if !virtualRecord.initialNonce.HasValue {
		// a virtual record is created accumulating information from at least one breadcrumb
		// we could have a breadcrumb for the same account in two different tracked blocks
		// a breadcrumb without nonce (if first occurred in record creation) leads to a virtual record without nonce
		// here we treat the virtual records which don't have a nonce yet.
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
	// this solves the next scenario:
	// two different breadcrumbs for the same sender
	// in the first tracked block, the last tx of that sender has nonce X -> virtual record, for the moment, has nonce X
	// in the second tracked block, the last tx of that sender has nonce X + N
	// the virtual record has to be updated with the maximum between current value (X) and (X+N) value
	// increasing by 1 helps knowing that we can select txs starting from this nonce of the virtual record
	virtualRecord.initialNonce = core.OptionalUint64{
		Value:    max(breadcrumb.lastNonce.Value, virtualRecord.initialNonce.Value) + 1,
		HasValue: true,
	}
}

func (virtualRecord *virtualAccountRecord) getInitialBalance() *big.Int {
	return virtualRecord.virtualBalance.getInitialBalance()
}

func (virtualRecord *virtualAccountRecord) getConsumedBalance() *big.Int {
	return virtualRecord.virtualBalance.getConsumedBalance()
}
