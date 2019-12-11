package transaction

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// txProcessor implements TransactionProcessor interface and can modify account states according to a transaction
type txProcessor struct {
	*baseTxProcessor
	hasher           hashing.Hasher
	scProcessor      process.SmartContractProcessor
	marshalizer      marshal.Marshalizer
	txFeeHandler     process.TransactionFeeHandler
	txTypeHandler    process.TxTypeHandler
	shardCoordinator sharding.Coordinator
	economicsFee     process.FeeHandler
	receiptForwarder process.IntermediateTransactionHandler
	badTxForwarder   process.IntermediateTransactionHandler
}

// NewTxProcessor creates a new txProcessor engine
func NewTxProcessor(
	accounts state.AccountsAdapter,
	hasher hashing.Hasher,
	addressConv state.AddressConverter,
	marshalizer marshal.Marshalizer,
	shardCoordinator sharding.Coordinator,
	scProcessor process.SmartContractProcessor,
	txFeeHandler process.TransactionFeeHandler,
	txTypeHandler process.TxTypeHandler,
	economicsFee process.FeeHandler,
	receiptForwarder process.IntermediateTransactionHandler,
	badTxForwarder process.IntermediateTransactionHandler,
) (*txProcessor, error) {

	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(addressConv) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(scProcessor) {
		return nil, process.ErrNilSmartContractProcessor
	}
	if check.IfNil(txFeeHandler) {
		return nil, process.ErrNilUnsignedTxHandler
	}
	if check.IfNil(txTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(receiptForwarder) {
		return nil, process.ErrNilReceiptHandler
	}
	if check.IfNil(badTxForwarder) {
		return nil, process.ErrNilBadTxHandler
	}

	baseTxProcess := &baseTxProcessor{
		accounts:         accounts,
		shardCoordinator: shardCoordinator,
		adrConv:          addressConv,
	}

	return &txProcessor{
		baseTxProcessor:  baseTxProcess,
		hasher:           hasher,
		marshalizer:      marshalizer,
		scProcessor:      scProcessor,
		txFeeHandler:     txFeeHandler,
		txTypeHandler:    txTypeHandler,
		economicsFee:     economicsFee,
		receiptForwarder: receiptForwarder,
		badTxForwarder:   badTxForwarder,
	}, nil
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (txProc *txProcessor) ProcessTransaction(tx *transaction.Transaction) error {
	if check.IfNil(tx) {
		return process.ErrNilTransaction
	}

	adrSrc, adrDst, err := txProc.getAddresses(tx)
	if err != nil {
		return err
	}

	acntSnd, err := txProc.getAccountFromAddress(adrSrc)
	if err != nil {
		return err
	}

	err = txProc.checkTxValues(tx, acntSnd)
	if err != nil {
		return err
	}

	txType, err := txProc.txTypeHandler.ComputeTransactionType(tx)
	if err != nil {
		return err
	}

	switch txType {
	case process.MoveBalance:
		return txProc.processMoveBalance(tx, adrSrc, adrDst)
	case process.SCDeployment:
		return txProc.processSCDeployment(tx, adrSrc)
	case process.SCInvoking:
		return txProc.processSCInvoking(tx, adrSrc, adrDst)
	}

	return process.ErrWrongTransaction
}

func (txProc *txProcessor) createReceiptsWhenFail(tx *transaction.Transaction, acntSnd *state.Account) error {
	if check.IfNil(acntSnd) {
		return nil
	}

	cost := txProc.economicsFee.ComputeFee(tx)
	operation := big.NewInt(0)

	err := acntSnd.SetBalanceWithJournal(operation.Sub(acntSnd.Balance, cost))
	if err != nil {
		return err
	}

	err = txProc.badTxForwarder.AddIntermediateTransactions([]data.TransactionHandler{tx})
	if err != nil {
		return err
	}

	txHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
	if err != nil {
		return err
	}

	rpt := &receipt.Receipt{
		Value:   big.NewInt(0).Set(cost),
		SndAddr: tx.SndAddr,
		Data:    "invalidTransaction",
		TxHash:  txHash,
	}

	err = txProc.receiptForwarder.AddIntermediateTransactions([]data.TransactionHandler{rpt})
	if err != nil {
		return err
	}

	return process.ErrFailedTransaction
}

func (txProc *txProcessor) createReceiptWithReturnedGas(tx *transaction.Transaction, acntSnd *state.Account) error {
	if check.IfNil(acntSnd) {
		return nil
	}

	totalProvided := big.NewInt(0)
	totalProvided.Mul(big.NewInt(0).SetUint64(tx.GasPrice), big.NewInt(0).SetUint64(tx.GasLimit))

	actualCost := txProc.economicsFee.ComputeFee(tx)
	refundValue := big.NewInt(0).Sub(totalProvided, actualCost)

	zero := big.NewInt(0)
	if refundValue.Cmp(zero) == 0 {
		return nil
	}

	txHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
	if err != nil {
		return err
	}

	rpt := &receipt.Receipt{
		Value:   big.NewInt(0).Set(refundValue),
		SndAddr: tx.SndAddr,
		Data:    "refundedGas",
		TxHash:  txHash,
	}

	err = txProc.receiptForwarder.AddIntermediateTransactions([]data.TransactionHandler{rpt})
	if err != nil {
		return err
	}

	return nil
}

func (txProc *txProcessor) processTxFee(tx *transaction.Transaction, acntSnd *state.Account) (*big.Int, error) {
	if acntSnd == nil {
		return big.NewInt(0), nil
	}

	err := txProc.economicsFee.CheckValidityTxValues(tx)
	if err != nil {
		return nil, err
	}

	cost := txProc.economicsFee.ComputeFee(tx)

	operation := big.NewInt(0)
	err = acntSnd.SetBalanceWithJournal(operation.Sub(acntSnd.Balance, cost))
	if err != nil {
		return nil, err
	}

	return cost, nil
}

func (txProc *txProcessor) processMoveBalance(
	tx *transaction.Transaction,
	adrSrc, adrDst state.AddressContainer,
) error {

	// getAccounts returns acntSrc not nil if the adrSrc is in the node shard, the same, acntDst will be not nil
	// if adrDst is in the node shard. If an error occurs it will be signaled in err variable.
	acntSrc, acntDst, err := txProc.getAccounts(adrSrc, adrDst)
	if err != nil {
		return err
	}

	txFee, err := txProc.processTxFee(tx, acntSrc)
	if err != nil {
		return err
	}

	value := tx.Value

	err = txProc.moveBalances(acntSrc, acntDst, value)
	if err != nil {
		return err
	}

	// is sender address in node shard
	if acntSrc != nil {
		err = txProc.increaseNonce(acntSrc)
		if err != nil {
			return err
		}
	}

	err = txProc.createReceiptWithReturnedGas(tx, acntSrc)
	if err != nil {
		return err
	}

	txProc.txFeeHandler.ProcessTransactionFee(txFee)

	return nil
}

func (txProc *txProcessor) processSCDeployment(
	tx *transaction.Transaction,
	adrSrc state.AddressContainer,
) error {
	// getAccounts returns acntSrc not nil if the adrSrc is in the node shard, the same, acntDst will be not nil
	// if adrDst is in the node shard. If an error occurs it will be signaled in err variable.
	acntSrc, err := txProc.getAccountFromAddress(adrSrc)
	if err != nil {
		return err
	}

	err = txProc.scProcessor.DeploySmartContract(tx, acntSrc)
	return err
}

func (txProc *txProcessor) processSCInvoking(
	tx *transaction.Transaction,
	adrSrc, adrDst state.AddressContainer,
) error {
	// getAccounts returns acntSrc not nil if the adrSrc is in the node shard, the same, acntDst will be not nil
	// if adrDst is in the node shard. If an error occurs it will be signaled in err variable.
	acntSrc, acntDst, err := txProc.getAccounts(adrSrc, adrDst)
	if err != nil {
		return err
	}

	err = txProc.scProcessor.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	return err
}

func (txProc *txProcessor) moveBalances(acntSrc, acntDst *state.Account,
	value *big.Int,
) error {
	operation1 := big.NewInt(0)
	operation2 := big.NewInt(0)

	// is sender address in node shard
	if acntSrc != nil {
		err := acntSrc.SetBalanceWithJournal(operation1.Sub(acntSrc.Balance, value))
		if err != nil {
			return err
		}
	}

	// is receiver address in node shard
	if acntDst != nil {
		err := acntDst.SetBalanceWithJournal(operation2.Add(acntDst.Balance, value))
		if err != nil {
			return err
		}
	}

	return nil
}

func (txProc *txProcessor) increaseNonce(acntSrc *state.Account) error {
	return acntSrc.SetNonceWithJournal(acntSrc.Nonce + 1)
}

// IsInterfaceNil returns true if there is no value under the interface
func (txProc *txProcessor) IsInterfaceNil() bool {
	if txProc == nil {
		return true
	}
	return false
}
