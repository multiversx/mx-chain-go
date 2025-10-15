package txcache

import (
	"bytes"

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
) *trackedBlock {
	return &trackedBlock{
		nonce:                nonce,
		hash:                 blockHash,
		rootHash:             rootHash,
		prevHash:             prevHash,
		breadcrumbsByAddress: make(map[string]*accountBreadcrumb),
	}
}

func (tb *trackedBlock) sameNonceOrBelow(otherBlock *trackedBlock) bool {
	return tb.nonce <= otherBlock.nonce
}

func (tb *trackedBlock) sameNonce(otherBlock *trackedBlock) bool {
	return tb.nonce == otherBlock.nonce
}

func (tb *trackedBlock) compileBreadcrumbs(txs []*WrappedTransaction) error {
	for _, tx := range txs {
		err := tb.compileBreadcrumb(tx)
		if err != nil {
			log.Debug("trackedBlock.compileBreadcrumbs failed",
				"err", err,
				"txHash", tx.TxHash,
				"sender", tx.Tx.GetSndAddr(),
				"nonce", tx.Tx.GetNonce(),
			)
			return err
		}
	}

	return nil
}

func (tb *trackedBlock) compileBreadcrumb(tx *WrappedTransaction) error {
	sender := tx.Tx.GetSndAddr()
	feePayer := tx.FeePayer
	initialNonce := tx.Tx.GetNonce()
	latestNonce := initialNonce

	// compile for sender
	senderBreadcrumb := tb.getOrCreateBreadcrumbWithNonce(string(sender), core.OptionalUint64{
		Value:    initialNonce,
		HasValue: true,
	})

	transferredValue := tx.TransferredValue
	senderBreadcrumb.accumulateConsumedBalance(transferredValue)

	err := senderBreadcrumb.updateNonceRange(core.OptionalUint64{
		Value:    latestNonce,
		HasValue: true,
	})
	if err != nil {
		return err
	}

	// compile for fee payer
	if feePayer == nil {
		return nil
	}

	isSenderTheFeePayer := bytes.Equal(sender, feePayer)
	if isSenderTheFeePayer {
		fee := tx.Fee
		senderBreadcrumb.accumulateConsumedBalance(fee)
		return nil
	}

	feePayerBreadcrumb := tb.getOrCreateBreadcrumb(string(feePayer))
	fee := tx.Fee
	feePayerBreadcrumb.accumulateConsumedBalance(fee)
	return nil
}

// getOrCreateBreadcrumbWithNonce is used on the flow of senders
func (tb *trackedBlock) getOrCreateBreadcrumbWithNonce(
	address string,
	nonce core.OptionalUint64,
) *accountBreadcrumb {
	breadCrumb, ok := tb.breadcrumbsByAddress[address]
	if ok {
		return breadCrumb
	}

	breadcrumb := newAccountBreadcrumb(nonce)
	tb.breadcrumbsByAddress[address] = breadcrumb

	return breadcrumb
}

// getOrCreateBreadcrumb is used on the flow of relayers
func (tb *trackedBlock) getOrCreateBreadcrumb(address string) *accountBreadcrumb {
	breadCrumb, ok := tb.breadcrumbsByAddress[address]
	if ok {
		return breadCrumb
	}

	breadcrumb := newAccountBreadcrumb(core.OptionalUint64{
		Value:    0,
		HasValue: false,
	})
	tb.breadcrumbsByAddress[address] = breadcrumb

	return breadcrumb
}
