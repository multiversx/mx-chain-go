package txcache

import (
	"bytes"
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
	session SelectionSession,
) (*trackedBlock, error) {

	tb := &trackedBlock{
		nonce:                nonce,
		hash:                 blockHash,
		rootHash:             rootHash,
		prevHash:             prevHash,
		breadcrumbsByAddress: make(map[string]*accountBreadcrumb),
	}

	err := tb.compileBreadcrumbs(txs, session)
	if err != nil {
		return nil, err
	}

	return tb, nil
}

func (tb *trackedBlock) sameNonce(trackedBlock1 *trackedBlock) bool {
	return tb.nonce == trackedBlock1.nonce
}

func (tb *trackedBlock) compileBreadcrumbs(txs []*WrappedTransaction, session SelectionSession) error {
	for _, tx := range txs {
		err := tb.compileBreadcrumb(tx, session)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO add validation when compiling breadcrumb
func (tb *trackedBlock) compileBreadcrumb(tx *WrappedTransaction, session SelectionSession) error {
	sender := tx.Tx.GetSndAddr()
	feePayer := tx.FeePayer
	initialNonce := tx.Tx.GetNonce()
	latestNonce := tx.Tx.GetNonce()

	accountState, err := session.GetAccountState(sender)
	if err != nil {
		return err
	}

	senderInitialBalance := accountState.GetBalance()

	// compile for sender
	senderBreadcrumb := tb.getOrCreateBreadcrumbWithNonce(string(sender), core.OptionalUint64{
		Value:    initialNonce,
		HasValue: true,
	}, senderInitialBalance)

	transferredValue := tx.TransferredValue
	senderBreadcrumb.accumulateConsumedBalance(transferredValue)

	err = senderBreadcrumb.updateLastNonce(core.OptionalUint64{
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

	if bytes.Equal(sender, feePayer) {
		fee := tx.Fee
		senderBreadcrumb.accumulateConsumedBalance(fee)
		return nil
	}

	accountState, err = session.GetAccountState(feePayer)
	if err != nil {
		return err
	}

	feePayerInitialBalance := accountState.GetBalance()
	feePayerBreadcrumb := tb.getOrCreateBreadcrumb(string(feePayer))
	fee := tx.Fee
	feePayerBreadcrumb.accumulateConsumedBalance(fee)
	return nil
}

func (tb *trackedBlock) getOrCreateBreadcrumb(address string, initialBalance *big.Int) *accountBreadcrumb {
	breadCrumb, ok := tb.breadcrumbsByAddress[address]
	if ok {
		return breadCrumb
	}

	breadcrumb := newAccountBreadcrumb(core.OptionalUint64{
		Value:    0,
		HasValue: false,
	}, initialBalance, big.NewInt(0))
	tb.breadcrumbsByAddress[address] = breadcrumb

	return breadcrumb
}

func (tb *trackedBlock) getOrCreateBreadcrumbWithNonce(
	address string,
	nonce core.OptionalUint64,
	initialBalance *big.Int) *accountBreadcrumb {
	breadCrumb, ok := tb.breadcrumbsByAddress[address]
	if ok {
		return breadCrumb
	}

	breadcrumb := newAccountBreadcrumb(nonce, initialBalance, big.NewInt(0))
	tb.breadcrumbsByAddress[address] = breadcrumb

	return breadcrumb
}
