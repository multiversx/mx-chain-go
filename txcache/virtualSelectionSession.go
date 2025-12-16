package txcache

import (
	"bytes"
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
	var feePayerRecord *virtualAccountRecord

	// check if there's a need to search for another record
	if tx.isFeePayerSameAsSender() {
		feePayerRecord = senderRecord
	} else {
		feePayer := tx.FeePayer
		record, err := virtualSession.getRecord(feePayer)
		if err != nil {
			log.Warn("accumulateConsumedBalance.getRecord feePayer",
				"feePayer", feePayer,
				"err", err)
			return err
		}

		// should affect the record of fee payer
		feePayerRecord = record
	}

	fee := tx.Fee
	if fee != nil {
		feePayerRecord.accumulateConsumedBalance(fee)
	}

	// getting the record of the gee payer might generate an unexpected failure.
	// this means that the transaction will not be selected.
	// accumulate the transferred value only if there isn't any error until here.
	transferredValue := tx.TransferredValue
	if transferredValue != nil {
		senderRecord.accumulateConsumedBalance(transferredValue)
	}

	return nil
}

func (virtualSession *virtualSelectionSession) consumedBalanceExceedsInitialBalance(address []byte, value *big.Int) bool {
	record, err := virtualSession.getRecord(address)
	if err != nil {
		log.Debug("virtualSelectionSession.consumedBalanceExceedsInitialBalance",
			"err", err)
		return true
	}

	consumedBalance := record.getConsumedBalance()
	futureConsumedBalance := new(big.Int).Add(consumedBalance, value)
	initialBalance := record.getInitialBalance()

	willBalanceBeExceeded := futureConsumedBalance.Cmp(initialBalance) > 0
	if willBalanceBeExceeded {
		logSelect.Trace("virtualSelectionSession.consumedBalanceExceedsInitialBalance",
			"initialBalance", initialBalance,
			"consumedBalance", consumedBalance,
		)
	}
	return willBalanceBeExceeded
}

// the selection of transactions has to be as restrictive as proposing the block.
// this means that we should check not only if fee exceeds the balance of relayer, but also if the transferred value exceeds it
func (virtualSession *virtualSelectionSession) detectWillBalanceBeExceeded(tx *WrappedTransaction) bool {
	if tx == nil {
		log.Debug("virtualSelectionSession.detectWillBalanceBeExceeded nil wrapped transaction")
		return true
	}

	if tx.Tx == nil {
		log.Debug("virtualSelectionSession.detectWillBalanceBeExceeded nil tx")
		return true
	}

	sender := tx.Tx.GetSndAddr()
	transferredValue := tx.TransferredValue
	if transferredValue == nil {
		log.Trace("virtualSelectionSession.detectWillBalanceBeExceeded nil transferredValue")
		return true
	}

	if virtualSession.consumedBalanceExceedsInitialBalance(sender, transferredValue) {
		logSelect.Debug("virtualSelectionSession.detectWillBalanceBeExceeded balance exceeded  by transferred value",
			"txHash", tx.TxHash,
			"sender", sender,
			"transferredValue", transferredValue,
		)
		return true
	}

	feePayer := tx.FeePayer
	fee := tx.Fee
	if fee == nil {
		log.Debug("virtualSelectionSession.detectWillBalanceBeExceeded nil fee")
		return true
	}

	if virtualSession.consumedBalanceExceedsInitialBalance(feePayer, fee) {
		logSelect.Trace("virtualSelectionSession.detectWillBalanceBeExceeded balance exceeded by fee",
			"txHash", tx.TxHash,
			"feePayer", feePayer,
			"fee", fee,
		)
		return true
	}

	accumulatedBalance := big.NewInt(0).Add(transferredValue, fee)
	if bytes.Equal(sender, feePayer) && virtualSession.consumedBalanceExceedsInitialBalance(sender, accumulatedBalance) {
		logSelect.Trace("virtualSelectionSession.detectWillBalanceBeExceeded balance exceeded by sum of transferred value and fee",
			"txHash", tx.TxHash,
			"sender", sender,
			"transferredValue", transferredValue,
			"fee", fee,
		)
		return true
	}

	return false
}

func (virtualSession *virtualSelectionSession) isIncorrectlyGuarded(tx data.TransactionHandler) bool {
	return virtualSession.session.IsIncorrectlyGuarded(tx)
}
