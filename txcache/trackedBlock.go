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

// TODO create the breadcrumbs directly in the constructor
func newTrackedBlock(nonce uint64, blockHash []byte, rootHash []byte,
	prevHash []byte) *trackedBlock {

	return &trackedBlock{
		nonce:                nonce,
		hash:                 blockHash,
		rootHash:             rootHash,
		prevHash:             prevHash,
		breadcrumbsByAddress: make(map[string]*accountBreadcrumb),
	}
}

func (tb *trackedBlock) sameNonce(trackedBlock1 *trackedBlock) bool {
	return tb.nonce == trackedBlock1.nonce
}

func (tb *trackedBlock) compileBreadcrumbs(txs []*WrappedTransaction) {
	for _, tx := range txs {
		tb.compileBreadcrumb(tx)
	}
}

// TODO add validation when compiling breadcrumb
func (tb *trackedBlock) compileBreadcrumb(tx *WrappedTransaction) {
	sender := tx.Tx.GetSndAddr()
	feePayer := tx.FeePayer
	initialNonce := tx.Tx.GetNonce()
	latestNonce := tx.Tx.GetNonce()

	// compile for sender
	senderBreadcrumb := tb.getOrCreateBreadcrumb(string(sender), core.OptionalUint64{
		Value:    initialNonce,
		HasValue: true,
	})
	transferredValue := tx.TransferredValue
	senderBreadcrumb.updateBreadcrumb(transferredValue, core.OptionalUint64{
		Value:    latestNonce,
		HasValue: true,
	})

	// compile for fee payer
	if feePayer != nil {
		feePayerBreadcrumb := tb.getOrCreateBreadcrumb(string(feePayer), core.OptionalUint64{
			Value:    0,
			HasValue: false,
		})
		fee := tx.Fee
		feePayerBreadcrumb.updateBreadcrumb(fee, core.OptionalUint64{
			Value:    0,
			HasValue: false,
		})
	}
}

func (tb *trackedBlock) getOrCreateBreadcrumb(address string, nonce core.OptionalUint64) *accountBreadcrumb {
	breadCrumb, ok := tb.breadcrumbsByAddress[address]
	if ok {
		return breadCrumb
	}

	breadcrumb := &accountBreadcrumb{
		initialNonce:    nonce,
		lastNonce:       nonce,
		consumedBalance: big.NewInt(0),
	}
	tb.breadcrumbsByAddress[address] = breadcrumb

	return breadcrumb
}
