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

// updateVirtualRecord updates the virtualBalance of a virtualAccountRecord and handles the nonces.
// The updateVirtualRecord is used only once, when the virtual record is created from a global account breadcrumb.
func (virtualRecord *virtualAccountRecord) updateVirtualRecord(globalBreadcrumb *globalAccountBreadcrumb) {
	virtualRecord.accumulateConsumedBalance(globalBreadcrumb.consumedBalance)

	if !globalBreadcrumb.lastNonce.HasValue {
		// We can have global breadcrumbs for accounts that are only used as relayers and never as senders.
		// The global breadcrumbs of those accounts don't have a set nonce, here we treat those cases.
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
