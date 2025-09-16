package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
)

type virtualSelectionSession struct {
	session                  SelectionSession
	virtualAccountsByAddress map[string]*virtualAccountRecord
}

func newVirtualSelectionSession(session SelectionSession, virtualAccountsByAddress map[string]*virtualAccountRecord) *virtualSelectionSession {
	return &virtualSelectionSession{
		session:                  session,
		virtualAccountsByAddress: virtualAccountsByAddress,
	}
}

func (virtualSession *virtualSelectionSession) getRecord(address []byte) (*virtualAccountRecord, error) {
	virtualRecord, ok := virtualSession.virtualAccountsByAddress[string(address)]
	if ok {
		return virtualRecord, nil
	}

	virtualRecord, err := virtualSession.createAccountRecord(address)
	if err != nil {
		log.Warn("virtualSelectionSession.getRecord: error when creating virtual account record",
			"address", address,
			"err", err)
		return nil, err
	}

	// We handle records corresponding to new (missing) accounts, as well (see "createAccountRecord").
	virtualSession.virtualAccountsByAddress[string(address)] = virtualRecord
	return virtualRecord, nil
}

func (virtualSession *virtualSelectionSession) createAccountRecord(address []byte) (*virtualAccountRecord, error) {
	initialNonce, initialBalance, _, err := virtualSession.session.GetAccountNonceAndBalance(address)
	if err != nil {
		return nil, err
	}

	return newVirtualAccountRecord(
		core.OptionalUint64{
			Value:    initialNonce,
			HasValue: true,
		},
		initialBalance,
	)
}

func (virtualSession *virtualSelectionSession) getNonceForAccountRecord(accountRecord *virtualAccountRecord) (uint64, error) {
	return accountRecord.getInitialNonce()
}

func (virtualSession *virtualSelectionSession) accumulateConsumedBalance(tx *WrappedTransaction, senderRecord *virtualAccountRecord) error {
	feePayer := tx.FeePayer

	feePayerRecord, err := virtualSession.getRecord(feePayer)
	if err != nil {
		log.Warn("accumulateConsumedBalance.getRecord feePayer",
			"feePayer", feePayer,
			"err", err)
		return err
	}

	transferredValue := tx.TransferredValue
	if transferredValue != nil {
		consumedBalance := senderRecord.getConsumedBalance()
		consumedBalance.Add(consumedBalance, transferredValue)
	}

	fee := tx.Fee
	if fee != nil {
		consumedBalance := feePayerRecord.getConsumedBalance()
		consumedBalance.Add(consumedBalance, fee)
	}

	return nil
}

func (virtualSession *virtualSelectionSession) detectWillFeeExceedBalance(tx *WrappedTransaction) bool {
	fee := tx.Fee
	if fee == nil {
		// unexpected failure
		log.Debug("virtualSelectionSession.detectWillFeeExceedBalance nil fee")
		return false
	}

	// Here, we are not interested into an eventual transfer of value (we only check if there's enough balance to pay the transaction fee).
	feePayer := tx.FeePayer
	feePayerRecord, err := virtualSession.getRecord(feePayer)
	if err != nil {
		log.Debug("virtualSelectionSession.detectWillFeeExceedBalance",
			"err", err)
		return false
	}

	consumedBalance := feePayerRecord.getConsumedBalance()
	futureConsumedBalance := new(big.Int).Add(consumedBalance, fee)
	feePayerBalance := feePayerRecord.getInitialBalance()

	willFeeExceedBalance := futureConsumedBalance.Cmp(feePayerBalance) > 0
	if willFeeExceedBalance {
		logSelect.Trace("virtualSelectionSession.detectWillFeeExceedBalance",
			"tx", tx.TxHash,
			"feePayer", feePayer,
			"initialBalance", feePayerBalance,
			"consumedBalance", consumedBalance,
		)
	}

	return willFeeExceedBalance
}

func (virtualSession *virtualSelectionSession) isIncorrectlyGuarded(tx data.TransactionHandler) bool {
	return virtualSession.session.IsIncorrectlyGuarded(tx)
}
