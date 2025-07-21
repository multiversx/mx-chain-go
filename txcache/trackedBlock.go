package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

type trackedBlock struct {
	nonce                uint64
	hash                 []byte
	rootHash             []byte
	prevHash             []byte
	breadcrumbsByAddress map[string]*accountBreadcrumb
}

func newTrackedBlock(
	nonce uint64,
	blockHash []byte,
	rootHash []byte,
	prevHash []byte,
	txs []*WrappedTransaction,
) (*trackedBlock, error) {

	tb := &trackedBlock{
		nonce:                nonce,
		hash:                 blockHash,
		rootHash:             rootHash,
		prevHash:             prevHash,
		breadcrumbsByAddress: make(map[string]*accountBreadcrumb),
	}

	err := tb.compileBreadcrumbs(txs)
	if err != nil {
		return nil, err
	}

	return tb, nil
}

func (tb *trackedBlock) sameNonce(trackedBlock1 *trackedBlock) bool {
	return tb.nonce == trackedBlock1.nonce
}

func (tb *trackedBlock) compileBreadcrumbs(txs []*WrappedTransaction) error {
	for _, tx := range txs {
		err := tb.compileBreadcrumb(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO add validation when compiling breadcrumb
// TODO optimize the flow in the case when sender is also fee payer
func (tb *trackedBlock) compileBreadcrumb(tx *WrappedTransaction) error {
	sender := tx.Tx.GetSndAddr()
	feePayer := tx.FeePayer
	initialNonce := tx.Tx.GetNonce()
	latestNonce := tx.Tx.GetNonce()

	// compile for sender
	senderBreadcrumb := tb.getOrCreateBreadcrumbWithNonce(string(sender), core.OptionalUint64{
		Value:    initialNonce,
		HasValue: true,
	})

	transferredValue := tx.TransferredValue
	senderBreadcrumb.accumulateConsumedBalance(transferredValue)

	err := senderBreadcrumb.updateLastNonce(core.OptionalUint64{
		Value:    latestNonce,
		HasValue: true,
	})
	if err != nil {
		return err
	}

	// compile for fee payer
	if feePayer != nil {
		feePayerBreadcrumb := tb.getOrCreateBreadcrumb(string(feePayer))
		fee := tx.Fee
		feePayerBreadcrumb.accumulateConsumedBalance(fee)
	}

	return nil
}

func (tb *trackedBlock) getOrCreateBreadcrumb(address string) *accountBreadcrumb {
	breadCrumb, ok := tb.breadcrumbsByAddress[address]
	if ok {
		return breadCrumb
	}

	breadcrumb := newAccountBreadcrumb(core.OptionalUint64{
		Value:    0,
		HasValue: false,
	}, big.NewInt(0))
	tb.breadcrumbsByAddress[address] = breadcrumb

	return breadcrumb
}

func (tb *trackedBlock) getOrCreateBreadcrumbWithNonce(address string, nonce core.OptionalUint64) *accountBreadcrumb {
	breadCrumb, ok := tb.breadcrumbsByAddress[address]
	if ok {
		return breadCrumb
	}

	breadcrumb := newAccountBreadcrumb(nonce, big.NewInt(0))
	tb.breadcrumbsByAddress[address] = breadcrumb

	return breadcrumb
}
