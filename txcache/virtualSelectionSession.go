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

type virtualAccountRecord struct {
	initialNonce    core.OptionalUint64
	initialBalance  *big.Int
	consumedBalance *big.Int
}

func newVirtualAccountRecord(initialNonce core.OptionalUint64, initialBalance *big.Int) *virtualAccountRecord {
	return &virtualAccountRecord{
		initialNonce:    initialNonce,
		initialBalance:  initialBalance,
		consumedBalance: big.NewInt(0),
	}
}

func newVirtualSelectionSession(session SelectionSession) *virtualSelectionSession {
	return &virtualSelectionSession{
		session:                  session,
		virtualAccountsByAddress: make(map[string]*virtualAccountRecord),
	}
}

func (virtualSession *virtualSelectionSession) getVirtualRecord(address []byte) (*virtualAccountRecord, error) {
	virtualRecord, ok := virtualSession.virtualAccountsByAddress[string(address)]
	if ok {
		return virtualRecord, nil
	}

	account, err := virtualSession.session.GetAccountState(address)
	if err != nil {
		log.Debug("virtualSelectionSession.getNonce",
			"address", address,
			"err", err)
		return nil, err
	}
	initialNonce := account.GetNonce()
	initialBalance := account.GetBalance()

	return newVirtualAccountRecord(core.OptionalUint64{
		Value:    initialNonce,
		HasValue: true}, initialBalance), nil
}

func (virtualSession *virtualSelectionSession) getNonce(address []byte) (uint64, error) {
	account, err := virtualSession.getVirtualRecord(address)
	if err != nil {
		log.Debug("virtualSelectionSession.getNonce",
			"address", address,
			"err", err)
		return 0, err
	}

	return account.initialNonce.Value, nil
}

func (virtualSession *virtualSelectionSession) detectWillFeeExceedBalance(tx *WrappedTransaction) bool {
	fee := tx.Fee
	if fee == nil {
		return false
	}

	// Here, we are not interested into an eventual transfer of value (we only check if there's enough balance to pay the transaction fee).
	feePayer := tx.FeePayer
	feePayerRecord, err := virtualSession.getVirtualRecord(feePayer)
	if err != nil {
		log.Debug("virtualSelectionSession.detectWillFeeExceedBalance",
			"err", err)
		return false
	}

	futureConsumedBalance := new(big.Int).Add(feePayerRecord.consumedBalance, fee)
	feePayerBalance := feePayerRecord.initialBalance

	willFeeExceedBalance := futureConsumedBalance.Cmp(feePayerBalance) > 0
	if willFeeExceedBalance {
		logSelect.Trace("virtualSelectionSession.detectWillFeeExceedBalance",
			"tx", tx.TxHash,
			"feePayer", feePayer,
			"initialBalance", feePayerRecord.initialBalance,
			"consumedBalance", feePayerRecord.consumedBalance,
		)
	}

	return willFeeExceedBalance
}

func (virtualSession *virtualSelectionSession) isIncorrectlyGuarded(tx data.TransactionHandler) bool {
	return virtualSession.session.IsIncorrectlyGuarded(tx)
}
