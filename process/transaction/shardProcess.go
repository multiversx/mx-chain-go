package transaction

import (
	"errors"
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

var _ process.TransactionProcessor = (*txProcessor)(nil)

// txProcessor implements TransactionProcessor interface and can modify account states according to a transaction
type txProcessor struct {
	*baseTxProcessor
	txFeeHandler     process.TransactionFeeHandler
	txTypeHandler    process.TxTypeHandler
	receiptForwarder process.IntermediateTransactionHandler
	badTxForwarder   process.IntermediateTransactionHandler
}

// NewTxProcessor creates a new txProcessor engine
func NewTxProcessor(
	accounts state.AccountsAdapter,
	hasher hashing.Hasher,
	pubkeyConv core.PubkeyConverter,
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
	if check.IfNil(pubkeyConv) {
		return nil, process.ErrNilPubkeyConverter
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
		pubkeyConv:       pubkeyConv,
		economicsFee:     economicsFee,
		hasher:           hasher,
		marshalizer:      marshalizer,
		scProcessor:      scProcessor,
	}

	return &txProcessor{
		baseTxProcessor:  baseTxProcess,
		txFeeHandler:     txFeeHandler,
		txTypeHandler:    txTypeHandler,
		receiptForwarder: receiptForwarder,
		badTxForwarder:   badTxForwarder,
	}, nil
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (txProc *txProcessor) ProcessTransaction(tx *transaction.Transaction) error {
	if check.IfNil(tx) {
		return process.ErrNilTransaction
	}

	acntSnd, acntDst, err := txProc.getAccounts(tx.SndAddr, tx.RcvAddr)
	if err != nil {
		return err
	}

	process.DisplayProcessTxDetails(
		"ProcessTransaction: sender account details",
		acntSnd,
		tx,
		txProc.pubkeyConv,
	)

	err = txProc.checkTxValues(tx, acntSnd, acntDst)
	if err != nil {
		if errors.Is(err, process.ErrInsufficientFunds) {
			receiptErr := txProc.executingFailedTransaction(tx, acntSnd, err)
			if receiptErr != nil {
				return receiptErr
			}
		}

		if errors.Is(err, process.ErrUserNameDoesNotMatchInCrossShardTx) {
			errProcessIfErr := txProc.processIfTxErrorCrossShard(tx, err.Error())
			if errProcessIfErr != nil {
				return errProcessIfErr
			}
			return nil
		}
		return err
	}

	txType := txProc.txTypeHandler.ComputeTransactionType(tx)
	switch txType {
	case process.MoveBalance:
		return txProc.processMoveBalance(tx, tx.SndAddr, tx.RcvAddr)
	case process.SCDeployment:
		return txProc.processSCDeployment(tx, tx.SndAddr)
	case process.SCInvoking:
		return txProc.processSCInvoking(tx, tx.SndAddr, tx.RcvAddr)
	case process.BuiltInFunctionCall:
		return txProc.processSCInvoking(tx, tx.SndAddr, tx.RcvAddr)
	}

	return process.ErrWrongTransaction
}

func (txProc *txProcessor) executingFailedTransaction(
	tx *transaction.Transaction,
	acntSnd state.UserAccountHandler,
	txError error,
) error {
	if check.IfNil(acntSnd) {
		return nil
	}

	txFee := txProc.economicsFee.ComputeFee(tx)
	err := acntSnd.SubFromBalance(txFee)
	if err != nil {
		return err
	}

	acntSnd.IncreaseNonce(1)
	err = txProc.badTxForwarder.AddIntermediateTransactions([]data.TransactionHandler{tx})
	if err != nil {
		return err
	}

	txHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
	if err != nil {
		return err
	}

	rpt := &receipt.Receipt{
		Value:   big.NewInt(0).Set(txFee),
		SndAddr: tx.SndAddr,
		Data:    []byte(txError.Error()),
		TxHash:  txHash,
	}

	err = txProc.receiptForwarder.AddIntermediateTransactions([]data.TransactionHandler{rpt})
	if err != nil {
		return err
	}

	txProc.txFeeHandler.ProcessTransactionFee(txFee, big.NewInt(0), txHash)

	err = txProc.accounts.SaveAccount(acntSnd)
	if err != nil {
		return err
	}

	return process.ErrFailedTransaction
}

func (txProc *txProcessor) createReceiptWithReturnedGas(txHash []byte, tx *transaction.Transaction, acntSnd state.UserAccountHandler) error {
	if check.IfNil(acntSnd) {
		return nil
	}
	if core.IsSmartContractAddress(tx.RcvAddr) {
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

	rpt := &receipt.Receipt{
		Value:   big.NewInt(0).Set(refundValue),
		SndAddr: tx.SndAddr,
		Data:    []byte("refundedGas"),
		TxHash:  txHash,
	}

	err := txProc.receiptForwarder.AddIntermediateTransactions([]data.TransactionHandler{rpt})
	if err != nil {
		return err
	}

	return nil
}

func (txProc *txProcessor) processTxFee(
	tx *transaction.Transaction,
	acntSnd, acntDst state.UserAccountHandler,
) (*big.Int, error) {
	if check.IfNil(acntSnd) {
		return big.NewInt(0), nil
	}

	cost := txProc.economicsFee.ComputeFee(tx)

	isCrossShardSCCall := check.IfNil(acntDst) && len(tx.GetData()) > 0 && core.IsSmartContractAddress(tx.GetRcvAddr())
	if isCrossShardSCCall {
		totalCost := big.NewInt(0).Mul(big.NewInt(0).SetUint64(tx.GetGasLimit()), big.NewInt(0).SetUint64(tx.GetGasPrice()))
		err := acntSnd.SubFromBalance(totalCost)
		if err != nil {
			return nil, err
		}
	} else {
		err := acntSnd.SubFromBalance(cost)
		if err != nil {
			return nil, err
		}
	}

	return cost, nil
}

func (txProc *txProcessor) checkIfValidTxToMetaChain(
	tx *transaction.Transaction,
	acntSnd state.UserAccountHandler,
	adrDst []byte,
) error {

	destShardId := txProc.shardCoordinator.ComputeId(adrDst)
	if destShardId != core.MetachainShardId {
		return nil
	}

	// it is not allowed to send transactions to metachain if those are not of type smart contract
	if len(tx.GetData()) == 0 {
		return txProc.executingFailedTransaction(tx, acntSnd, process.ErrInvalidMetaTransaction)
	}

	return nil
}

func (txProc *txProcessor) processMoveBalance(
	tx *transaction.Transaction,
	adrSrc, adrDst []byte,
) error {

	// getAccounts returns acntSrc not nil if the adrSrc is in the node shard, the same, acntDst will be not nil
	// if adrDst is in the node shard. If an error occurs it will be signaled in err variable.
	acntSrc, acntDst, err := txProc.getAccounts(adrSrc, adrDst)
	if err != nil {
		return err
	}

	err = txProc.checkIfValidTxToMetaChain(tx, acntSrc, adrDst)
	if err != nil {
		return err
	}

	txFee, err := txProc.processTxFee(tx, acntSrc, acntDst)
	if err != nil {
		return err
	}

	err = txProc.moveBalances(acntSrc, acntDst, tx.GetValue())
	if err != nil {
		return err
	}

	// is sender address in node shard
	if acntSrc != nil {
		acntSrc.IncreaseNonce(1)
	}

	txHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
	if err != nil {
		return err
	}

	err = txProc.createReceiptWithReturnedGas(txHash, tx, acntSrc)
	if err != nil {
		return err
	}

	txProc.txFeeHandler.ProcessTransactionFee(txFee, big.NewInt(0), txHash)

	return txProc.saveAccounts(acntSrc, acntDst)
}

func (txProc *txProcessor) saveAccounts(acntSnd, acntDst state.AccountHandler) error {
	if !check.IfNil(acntSnd) {
		err := txProc.accounts.SaveAccount(acntSnd)
		if err != nil {
			return err
		}
	}

	if !check.IfNil(acntDst) {
		err := txProc.accounts.SaveAccount(acntDst)
		if err != nil {
			return err
		}
	}

	return nil
}

func (txProc *txProcessor) processSCDeployment(
	tx *transaction.Transaction,
	adrSrc []byte,
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
	adrSrc, adrDst []byte,
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

func (txProc *txProcessor) moveBalances(
	acntSrc, acntDst state.UserAccountHandler,
	value *big.Int,
) error {
	// is sender address in node shard
	if !check.IfNil(acntSrc) {
		err := acntSrc.SubFromBalance(value)
		if err != nil {
			return err
		}
	}

	// is receiver address in node shard
	if !check.IfNil(acntDst) {
		err := acntDst.AddToBalance(value)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (txProc *txProcessor) IsInterfaceNil() bool {
	return txProc == nil
}
