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

func newVirtualSelectionSession(session SelectionSession) *virtualSelectionSession {
	return &virtualSelectionSession{
		session:                  session,
		virtualAccountsByAddress: make(map[string]*virtualAccountRecord),
	}
}

func (virtualSession *virtualSelectionSession) getRecord(address []byte) (*virtualAccountRecord, error) {
	virtualRecord, ok := virtualSession.virtualAccountsByAddress[string(address)]
	if ok {
		return virtualRecord, nil
	}

	account, err := virtualSession.session.GetAccountState(address)
	if err != nil {
		log.Debug("virtualSelectionSession.getRecord",
			"address", address,
			"err", err)
		return nil, err
	}

	initialNonce := account.GetNonce()
	initialBalance := account.GetBalance()
	virtualRecord = newVirtualAccountRecord(
		core.OptionalUint64{
			Value:    initialNonce,
			HasValue: true,
		},
		initialBalance,
	)

	virtualSession.virtualAccountsByAddress[string(address)] = virtualRecord
	return virtualRecord, nil
}

func (virtualSession *virtualSelectionSession) getNonce(address []byte) (uint64, error) {
	account, err := virtualSession.getRecord(address)
	if err != nil {
		log.Debug("virtualSelectionSession.getNonce",
			"address", address,
			"err", err)
		return 0, err
	}

	return account.initialNonce.Value, nil
}

func (virtualSession *virtualSelectionSession) accumulateConsumedBalance(tx *WrappedTransaction) error {
	sender := tx.Tx.GetSndAddr()
	feePayer := tx.FeePayer

	senderRecord, err := virtualSession.getRecord(sender)
	if err != nil {
		log.Warn("accumulateConsumedBalance.getRecord sender",
			"sender", sender,
			"err", err)
		return err
	}

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
