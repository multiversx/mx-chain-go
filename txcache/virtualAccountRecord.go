package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

// TODO should refactor this; each account virtual record should be simply created from a global account breadcrumb, without further updates

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
func (virtualRecord *virtualAccountRecord) updateVirtualRecord(globalBreadcrumb *globalAccountBreadcrumb) {
	virtualRecord.virtualBalance.accumulateConsumedBalance(globalBreadcrumb.consumedBalance)

	if !globalBreadcrumb.lastNonce.HasValue {
		// In a certain tracked block, we can have breadcrumbs for accounts that are only used as relayers and never as senders.
		// Breadcrumb of those accounts don't have a set nonce, here we treat those cases.
		return
	}

	initialNonce := globalBreadcrumb.lastNonce.Value + 1
	virtualRecord.initialNonce = core.OptionalUint64{
		Value:    initialNonce,
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
