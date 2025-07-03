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

func (tb *trackedBlock) createOrUpdateVirtualRecords(
	session SelectionSession,
	skippedSenders map[string]struct{},
	sendersInContinuityWithSessionNonce map[string]struct{},
	accountPreviousBreadcrumb map[string]*accountBreadcrumb,
	virtualAccountsByAddress map[string]*virtualAccountRecord,
) error {
	for address, breadcrumb := range tb.breadcrumbsByAddress {
		_, ok := skippedSenders[address]
		if ok {
			continue
		}

		accountState, err := session.GetAccountState([]byte(address))
		if err != nil {
			log.Debug("selectionTracker.createVirtualSelectionSession",
				"err", err)
			return err
		}

		accountNonce := accountState.GetNonce()

		if !breadcrumb.isContinuous(address, accountNonce,
			sendersInContinuityWithSessionNonce, accountPreviousBreadcrumb) {
			skippedSenders[address] = struct{}{}
			delete(virtualAccountsByAddress, address)
			continue
		}

		breadcrumb.createOrUpdateVirtualRecord(virtualAccountsByAddress, accountState, address)
	}

	return nil
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
// TODO optimize the flow in the case when sender is also fee payer
func (tb *trackedBlock) compileBreadcrumb(tx *WrappedTransaction) {
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
	senderBreadcrumb.updateNonce(core.OptionalUint64{
		Value:    latestNonce,
		HasValue: true,
	})

	// compile for fee payer
	if feePayer != nil {
		feePayerBreadcrumb := tb.getOrCreateBreadcrumb(string(feePayer))
		fee := tx.Fee
		feePayerBreadcrumb.accumulateConsumedBalance(fee)
	}
}

func (tb *trackedBlock) getOrCreateBreadcrumb(address string) *accountBreadcrumb {
	breadCrumb, ok := tb.breadcrumbsByAddress[address]
	if ok {
		return breadCrumb
	}

	breadcrumb := newAccountBreadcrumb(core.OptionalUint64{
		Value:    0,
		HasValue: false,
	}, core.OptionalUint64{
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

	breadcrumb := newAccountBreadcrumb(nonce, nonce, big.NewInt(0))
	tb.breadcrumbsByAddress[address] = breadcrumb

	return breadcrumb
}
