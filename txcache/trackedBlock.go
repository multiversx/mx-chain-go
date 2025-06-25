package txcache

import "math/big"

type trackedBlock struct {
	nonce                uint64
	hash                 []byte
	rootHash             []byte
	prevHash             []byte
	breadcrumbsByAddress map[string]*accountBreadcrumb
}

func newTrackedBlock(nonce uint64, blockHash []byte, rootHash []byte, prevHash []byte) *trackedBlock {
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

func (tb *trackedBlock) compileBreadcrumb(tx *WrappedTransaction) {
	sender := tx.Tx.GetSndAddr()
	feePayer := tx.FeePayer
	nonce := tx.Tx.GetNonce()

	// compile for sender
	senderBreadcrumb := tb.getBreadcrumb(string(sender), nonce)
	transferredValue := tx.TransferredValue
	if transferredValue != nil {
		senderBreadcrumb.consumedBalance.Add(senderBreadcrumb.consumedBalance, transferredValue)
	}
	senderBreadcrumb.lastNonce = nonce

	// compile for fee payer
	if feePayer != nil {
		feePayerBreadcrumb := tb.getBreadcrumb(string(feePayer), nonce)
		fee := tx.Fee
		if fee != nil {
			feePayerBreadcrumb.consumedBalance.Add(feePayerBreadcrumb.consumedBalance, fee)
		}
		feePayerBreadcrumb.lastNonce = nonce
	}
}

func (tb *trackedBlock) getBreadcrumb(address string, nonce uint64) *accountBreadcrumb {
	breadCrumb, ok := tb.breadcrumbsByAddress[address]
	if ok {
		return breadCrumb
	} else {
		breadcrumb := &accountBreadcrumb{
			initialNonce:    nonce,
			lastNonce:       nonce,
			consumedBalance: big.NewInt(0),
		}

		tb.breadcrumbsByAddress[address] = breadcrumb

		return breadcrumb
	}
}
