package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

type virtualAccountRecord struct {
	initialNonce   core.OptionalUint64
	virtualBalance *virtualAccountBalance
}

func newVirtualAccountRecord(initialNonce core.OptionalUint64, initialBalance *big.Int) (*virtualAccountRecord, error) {
	virtualBalance, err := newVirtualAccountBalance(initialBalance)
	if err != nil {
		return nil, err
	}

	return &virtualAccountRecord{
		initialNonce:   initialNonce,
		virtualBalance: virtualBalance,
	}, nil
}

// updateVirtualRecord updates the virtualBalance of a virtualAccountRecord and handles the nonces
func (virtualRecord *virtualAccountRecord) updateVirtualRecord(breadcrumb *accountBreadcrumb) {
	virtualRecord.virtualBalance.accumulateConsumedBalance(breadcrumb.consumedBalance)

	if !breadcrumb.lastNonce.HasValue {
		// In a certain tracked block, we can have breadcrumbs for accounts that are only used as relayers and never as senders.
		// Breadcrumb of those accounts don't have a set nonce, here we treat those cases.
		return
	}

	if !virtualRecord.initialNonce.HasValue {
		// A virtual record is created by accumulating information from at least one breadcrumb.
		// We could have a breadcrumb for the same account in two different tracked blocks.
		// A breadcrumb without nonce (if first occurred in record creation) leads to a virtual record without nonce.
		// Here we treat the virtual records which don't have a nonce yet.
		virtualRecord.setInitialNonceOnFirstUpdate(breadcrumb)
		return
	}

	virtualRecord.setInitialNonceOnSubsequentUpdate(breadcrumb)
}

func (virtualRecord *virtualAccountRecord) setInitialNonceOnFirstUpdate(breadcrumb *accountBreadcrumb) {
	virtualRecord.initialNonce = core.OptionalUint64{
		Value:    breadcrumb.lastNonce.Value + 1,
		HasValue: true,
	}
}

func (virtualRecord *virtualAccountRecord) setInitialNonceOnSubsequentUpdate(breadcrumb *accountBreadcrumb) {
	// This solves the common scenario when we have at least two different breadcrumbs for the same sender.
	// For example:
	// In a first tracked block, the last tx of a certain sender has nonce X.
	// This means that its virtual record, for the moment, has nonce X.
	// In a second tracked block, the last tx of the same sender has nonce X + N.
	// This means that the virtual record has to be updated with the maximum between current value (X) and (X+N) value.
	// The resulted value + 1 indicates the starting nonce which can be used for selection for that certain sender.
	virtualRecord.initialNonce = core.OptionalUint64{
		Value:    max(breadcrumb.lastNonce.Value, virtualRecord.initialNonce.Value) + 1,
		HasValue: true,
	}
}

func (virtualRecord *virtualAccountRecord) getInitialNonce() (uint64, error) {
	if !virtualRecord.initialNonce.HasValue {
		log.Debug("virtualAccountRecord.getInitialNonce",
			"err", errNonceNotSet)
		return 0, errNonceNotSet
	}

	return virtualRecord.initialNonce.Value, nil
}

func (virtualRecord *virtualAccountRecord) getInitialBalance() *big.Int {
	return virtualRecord.virtualBalance.getInitialBalance()
}

func (virtualRecord *virtualAccountRecord) getConsumedBalance() *big.Int {
	return virtualRecord.virtualBalance.getConsumedBalance()
}

func (virtualRecord *virtualAccountRecord) accumulateConsumedBalance(consumedBalance *big.Int) {
	virtualRecord.virtualBalance.accumulateConsumedBalance(consumedBalance)
}
