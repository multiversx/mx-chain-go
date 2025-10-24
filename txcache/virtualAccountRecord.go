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
