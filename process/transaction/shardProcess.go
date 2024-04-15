package transaction

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var log = logger.GetOrCreate("process/transaction")
var _ process.TransactionProcessor = (*txProcessor)(nil)

// RefundGasMessage is the message returned in the data field of a receipt,
// for move balance transactions that provide more gas than needed
const RefundGasMessage = "refundedGas"

type relayedFees struct {
	totalFee, remainingFee, relayerFee *big.Int
}

// txProcessor implements TransactionProcessor interface and can modify account states according to a transaction
type txProcessor struct {
	*baseTxProcessor
	txFeeHandler         process.TransactionFeeHandler
	txTypeHandler        process.TxTypeHandler
	receiptForwarder     process.IntermediateTransactionHandler
	badTxForwarder       process.IntermediateTransactionHandler
	argsParser           process.ArgumentsParser
	scrForwarder         process.IntermediateTransactionHandler
	signMarshalizer      marshal.Marshalizer
	enableEpochsHandler  common.EnableEpochsHandler
	txLogsProcessor      process.TransactionLogProcessor
	relayedTxV3Processor process.RelayedTxV3Processor
}

// ArgsNewTxProcessor defines the arguments needed for new tx processor
type ArgsNewTxProcessor struct {
	Accounts             state.AccountsAdapter
	Hasher               hashing.Hasher
	PubkeyConv           core.PubkeyConverter
	Marshalizer          marshal.Marshalizer
	SignMarshalizer      marshal.Marshalizer
	ShardCoordinator     sharding.Coordinator
	ScProcessor          process.SmartContractProcessor
	TxFeeHandler         process.TransactionFeeHandler
	TxTypeHandler        process.TxTypeHandler
	EconomicsFee         process.FeeHandler
	ReceiptForwarder     process.IntermediateTransactionHandler
	BadTxForwarder       process.IntermediateTransactionHandler
	ArgsParser           process.ArgumentsParser
	ScrForwarder         process.IntermediateTransactionHandler
	EnableRoundsHandler  process.EnableRoundsHandler
	EnableEpochsHandler  common.EnableEpochsHandler
	TxVersionChecker     process.TxVersionCheckerHandler
	GuardianChecker      process.GuardianChecker
	TxLogsProcessor      process.TransactionLogProcessor
	RelayedTxV3Processor process.RelayedTxV3Processor
}

// NewTxProcessor creates a new txProcessor engine
func NewTxProcessor(args ArgsNewTxProcessor) (*txProcessor, error) {
	if check.IfNil(args.Accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.PubkeyConv) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.ScProcessor) {
		return nil, process.ErrNilSmartContractProcessor
	}
	if check.IfNil(args.TxFeeHandler) {
		return nil, process.ErrNilUnsignedTxHandler
	}
	if check.IfNil(args.TxTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(args.EconomicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.ReceiptForwarder) {
		return nil, process.ErrNilReceiptHandler
	}
	if check.IfNil(args.BadTxForwarder) {
		return nil, process.ErrNilBadTxHandler
	}
	if check.IfNil(args.ArgsParser) {
		return nil, process.ErrNilArgumentParser
	}
	if check.IfNil(args.ScrForwarder) {
		return nil, process.ErrNilIntermediateTransactionHandler
	}
	if check.IfNil(args.SignMarshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.EnableRoundsHandler) {
		return nil, process.ErrNilEnableRoundsHandler
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.PenalizedTooMuchGasFlag,
		common.MetaProtectionFlag,
		common.AddFailedRelayedTxToInvalidMBsFlag,
		common.RelayedTransactionsFlag,
		common.RelayedTransactionsV2Flag,
		common.RelayedNonceFixFlag,
		common.RelayedTransactionsV3Flag,
		common.FixRelayedMoveBalanceFlag,
	})
	if err != nil {
		return nil, err
	}
	if check.IfNil(args.TxVersionChecker) {
		return nil, process.ErrNilTransactionVersionChecker
	}
	if check.IfNil(args.GuardianChecker) {
		return nil, process.ErrNilGuardianChecker
	}
	if check.IfNil(args.TxLogsProcessor) {
		return nil, process.ErrNilTxLogsProcessor
	}
	if check.IfNil(args.RelayedTxV3Processor) {
		return nil, process.ErrNilRelayedTxV3Processor
	}

	baseTxProcess := &baseTxProcessor{
		accounts:            args.Accounts,
		shardCoordinator:    args.ShardCoordinator,
		pubkeyConv:          args.PubkeyConv,
		economicsFee:        args.EconomicsFee,
		hasher:              args.Hasher,
		marshalizer:         args.Marshalizer,
		scProcessor:         args.ScProcessor,
		enableEpochsHandler: args.EnableEpochsHandler,
		txVersionChecker:    args.TxVersionChecker,
		guardianChecker:     args.GuardianChecker,
	}

	txProc := &txProcessor{
		baseTxProcessor:      baseTxProcess,
		txFeeHandler:         args.TxFeeHandler,
		txTypeHandler:        args.TxTypeHandler,
		receiptForwarder:     args.ReceiptForwarder,
		badTxForwarder:       args.BadTxForwarder,
		argsParser:           args.ArgsParser,
		scrForwarder:         args.ScrForwarder,
		signMarshalizer:      args.SignMarshalizer,
		enableEpochsHandler:  args.EnableEpochsHandler,
		txLogsProcessor:      args.TxLogsProcessor,
		relayedTxV3Processor: args.RelayedTxV3Processor,
	}

	return txProc, nil
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (txProc *txProcessor) ProcessTransaction(tx *transaction.Transaction) (vmcommon.ReturnCode, error) {
	if check.IfNil(tx) {
		return 0, process.ErrNilTransaction
	}

	acntSnd, acntDst, err := txProc.getAccounts(tx.SndAddr, tx.RcvAddr)
	if err != nil {
		return 0, err
	}

	txHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
	if err != nil {
		return 0, err
	}

	process.DisplayProcessTxDetails(
		"ProcessTransaction: sender account details",
		acntSnd,
		tx,
		txHash,
		txProc.pubkeyConv,
	)

	txType, dstShardTxType := txProc.txTypeHandler.ComputeTransactionType(tx)
	err = txProc.checkTxValues(tx, acntSnd, acntDst, false)
	if err != nil {
		if errors.Is(err, process.ErrInsufficientFunds) {
			receiptErr := txProc.executingFailedTransaction(tx, acntSnd, err)
			if receiptErr != nil {
				return 0, receiptErr
			}
		}

		if errors.Is(err, process.ErrUserNameDoesNotMatch) && txProc.enableEpochsHandler.IsFlagEnabled(common.RelayedTransactionsFlag) {
			receiptErr := txProc.executingFailedTransaction(tx, acntSnd, err)
			if receiptErr != nil {
				return vmcommon.UserError, receiptErr
			}
		}

		if errors.Is(err, process.ErrUserNameDoesNotMatchInCrossShardTx) {
			errProcessIfErr := txProc.processIfTxErrorCrossShard(tx, err.Error())
			if errProcessIfErr != nil {
				return 0, errProcessIfErr
			}
			return vmcommon.UserError, nil
		}
		return vmcommon.UserError, err
	}

	switch txType {
	case process.MoveBalance:
		err = txProc.processMoveBalance(tx, acntSnd, acntDst, dstShardTxType, nil, false)
		if err != nil {
			return vmcommon.UserError, txProc.executeAfterFailedMoveBalanceTransaction(tx, err)
		}
		return vmcommon.Ok, err
	case process.SCDeployment:
		return txProc.processSCDeployment(tx, acntSnd)
	case process.SCInvoking:
		return txProc.processSCInvoking(tx, acntSnd, acntDst)
	case process.BuiltInFunctionCall:
		return txProc.processBuiltInFunctionCall(tx, acntSnd, acntDst)
	case process.RelayedTx:
		return txProc.processRelayedTx(tx, acntSnd, acntDst)
	case process.RelayedTxV2:
		return txProc.processRelayedTxV2(tx, acntSnd, acntDst)
	case process.RelayedTxV3:
		return txProc.processRelayedTxV3(tx, acntSnd)
	}

	return vmcommon.UserError, txProc.executingFailedTransaction(tx, acntSnd, process.ErrWrongTransaction)
}

func (txProc *txProcessor) executeAfterFailedMoveBalanceTransaction(
	tx *transaction.Transaction,
	txError error,
) error {
	if core.IsGetNodeFromDBError(txError) {
		return txError
	}

	acntSnd, err := txProc.getAccountFromAddress(tx.SndAddr)
	if err != nil {
		return err
	}

	if errors.Is(txError, process.ErrInvalidMetaTransaction) || errors.Is(txError, process.ErrAccountNotPayable) {
		snapshot := txProc.accounts.JournalLen()
		var txHash []byte
		txHash, err = core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
		if err != nil {
			return err
		}

		err = txProc.scProcessor.ProcessIfError(acntSnd, txHash, tx, txError.Error(), nil, snapshot, 0)
		if err != nil {
			return err
		}

		if check.IfNil(acntSnd) {
			return nil
		}

		err = txProc.badTxForwarder.AddIntermediateTransactions([]data.TransactionHandler{tx})
		if err != nil {
			return err
		}

		return process.ErrFailedTransaction
	}

	return txError
}

func (txProc *txProcessor) executingFailedTransaction(
	tx *transaction.Transaction,
	acntSnd state.UserAccountHandler,
	txError error,
) error {
	if check.IfNil(acntSnd) {
		return nil
	}

	txFee := txProc.economicsFee.ComputeFeeForProcessing(tx, tx.GasLimit)
	if txProc.enableEpochsHandler.IsFlagEnabled(common.FixRelayedMoveBalanceFlag) {
		moveBalanceGasLimit := txProc.economicsFee.ComputeGasLimit(tx)
		gasToUse := tx.GetGasLimit() - moveBalanceGasLimit
		moveBalanceUserFee := txProc.economicsFee.ComputeMoveBalanceFee(tx)
		processingUserFee := txProc.economicsFee.ComputeFeeForProcessing(tx, gasToUse)
		txFee = big.NewInt(0).Add(moveBalanceUserFee, processingUserFee)
	}
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

	log.Trace("executingFailedTransaction", "fail reason(error)", txError, "tx hash", txHash)

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

func (txProc *txProcessor) createReceiptWithReturnedGas(
	txHash []byte,
	tx *transaction.Transaction,
	acntSnd state.UserAccountHandler,
	moveBalanceCost *big.Int,
	totalProvided *big.Int,
	destShardTxType process.TransactionType,
	isUserTxOfRelayed bool,
) error {
	if check.IfNil(acntSnd) || isUserTxOfRelayed {
		return nil
	}
	shouldCreateReceiptBackwardCompatible := !txProc.enableEpochsHandler.IsFlagEnabled(common.MetaProtectionFlag) && core.IsSmartContractAddress(tx.RcvAddr)
	if destShardTxType != process.MoveBalance || shouldCreateReceiptBackwardCompatible {
		return nil
	}

	refundValue := big.NewInt(0).Sub(totalProvided, moveBalanceCost)

	zero := big.NewInt(0)
	if refundValue.Cmp(zero) == 0 {
		return nil
	}

	rpt := &receipt.Receipt{
		Value:   big.NewInt(0).Set(refundValue),
		SndAddr: tx.SndAddr,
		Data:    []byte(RefundGasMessage),
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
	dstShardTxType process.TransactionType,
	isUserTxOfRelayed bool,
) (*big.Int, *big.Int, error) {
	if check.IfNil(acntSnd) {
		return big.NewInt(0), big.NewInt(0), nil
	}

	if isUserTxOfRelayed {
		totalCost := txProc.economicsFee.ComputeFeeForProcessing(tx, tx.GasLimit)
		if txProc.enableEpochsHandler.IsFlagEnabled(common.FixRelayedMoveBalanceFlag) {
			moveBalanceGasLimit := txProc.economicsFee.ComputeGasLimit(tx)
			gasToUse := tx.GetGasLimit() - moveBalanceGasLimit
			moveBalanceUserFee := txProc.economicsFee.ComputeMoveBalanceFee(tx)
			processingUserFee := txProc.economicsFee.ComputeFeeForProcessing(tx, gasToUse)
			totalCost = big.NewInt(0).Add(moveBalanceUserFee, processingUserFee)
		}
		err := acntSnd.SubFromBalance(totalCost)
		if err != nil {
			return nil, nil, err
		}

		if dstShardTxType == process.MoveBalance {
			return totalCost, totalCost, nil
		}

		moveBalanceGasLimit := txProc.economicsFee.ComputeGasLimit(tx)
		currentShardFee := txProc.economicsFee.ComputeFeeForProcessing(tx, moveBalanceGasLimit)
		return currentShardFee, totalCost, nil
	}

	moveBalanceFee := txProc.economicsFee.ComputeMoveBalanceFee(tx)
	totalCost := txProc.economicsFee.ComputeTxFee(tx)

	if !txProc.enableEpochsHandler.IsFlagEnabled(common.PenalizedTooMuchGasFlag) {
		totalCost = core.SafeMul(tx.GasLimit, tx.GasPrice)
	}

	isCrossShardSCCall := check.IfNil(acntDst) && len(tx.GetData()) > 0 && core.IsSmartContractAddress(tx.GetRcvAddr())
	if dstShardTxType != process.MoveBalance ||
		(!txProc.enableEpochsHandler.IsFlagEnabled(common.MetaProtectionFlag) && isCrossShardSCCall) {

		err := acntSnd.SubFromBalance(totalCost)
		if err != nil {
			return nil, nil, err
		}
	} else {
		err := acntSnd.SubFromBalance(moveBalanceFee)
		if err != nil {
			return nil, nil, err
		}
	}

	return moveBalanceFee, totalCost, nil
}

func (txProc *txProcessor) checkIfValidTxToMetaChain(
	tx *transaction.Transaction,
	adrDst []byte,
) error {

	destShardId := txProc.shardCoordinator.ComputeId(adrDst)
	if destShardId != core.MetachainShardId {
		return nil
	}

	// it is not allowed to send transactions to metachain if those are not of type smart contract
	if len(tx.GetData()) == 0 {
		return process.ErrInvalidMetaTransaction
	}

	if txProc.enableEpochsHandler.IsFlagEnabled(common.MetaProtectionFlag) {
		// additional check
		if tx.GasLimit < txProc.economicsFee.ComputeGasLimit(tx)+core.MinMetaTxExtraGasCost {
			return fmt.Errorf("%w: not enough gas", process.ErrInvalidMetaTransaction)
		}
	}

	return nil
}

func (txProc *txProcessor) processMoveBalance(
	tx *transaction.Transaction,
	acntSrc, acntDst state.UserAccountHandler,
	destShardTxType process.TransactionType,
	originalTxHash []byte,
	isUserTxOfRelayed bool,
) error {

	moveBalanceCost, totalCost, err := txProc.processTxFee(tx, acntSrc, acntDst, destShardTxType, isUserTxOfRelayed)
	if err != nil {
		return err
	}

	// is sender address in node shard
	if !check.IfNil(acntSrc) {
		acntSrc.IncreaseNonce(1)
		err = acntSrc.SubFromBalance(tx.Value)
		if err != nil {
			return err
		}

		err = txProc.accounts.SaveAccount(acntSrc)
		if err != nil {
			return err
		}
	}

	isPayable, err := txProc.scProcessor.IsPayable(tx.SndAddr, tx.RcvAddr)
	if err != nil {
		return err
	}
	if !isPayable {
		return process.ErrAccountNotPayable
	}

	err = txProc.checkIfValidTxToMetaChain(tx, tx.RcvAddr)
	if err != nil {
		return err
	}

	// is receiver address in node shard
	if !check.IfNil(acntDst) {
		err = acntDst.AddToBalance(tx.Value)
		if err != nil {
			return err
		}

		err = txProc.accounts.SaveAccount(acntDst)
		if err != nil {
			return err
		}
	}

	txHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
	if err != nil {
		return err
	}

	err = txProc.createReceiptWithReturnedGas(txHash, tx, acntSrc, moveBalanceCost, totalCost, destShardTxType, isUserTxOfRelayed)
	if err != nil {
		return err
	}

	if isUserTxOfRelayed {
		txProc.txFeeHandler.ProcessTransactionFeeRelayedUserTx(moveBalanceCost, big.NewInt(0), txHash, originalTxHash)
	} else {
		txProc.txFeeHandler.ProcessTransactionFee(moveBalanceCost, big.NewInt(0), txHash)
	}

	return nil
}

func (txProc *txProcessor) processSCDeployment(
	tx *transaction.Transaction,
	acntSrc state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	return txProc.scProcessor.DeploySmartContract(tx, acntSrc)
}

func (txProc *txProcessor) processSCInvoking(
	tx *transaction.Transaction,
	acntSrc, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	return txProc.scProcessor.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
}

func (txProc *txProcessor) processBuiltInFunctionCall(
	tx *transaction.Transaction,
	acntSrc, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	return txProc.scProcessor.ExecuteBuiltInFunction(tx, acntSrc, acntDst)
}

func makeUserTxFromRelayedTxV2Args(args [][]byte) *transaction.Transaction {
	userTx := &transaction.Transaction{}
	userTx.RcvAddr = args[0]
	userTx.Nonce = big.NewInt(0).SetBytes(args[1]).Uint64()
	userTx.Data = args[2]
	userTx.Signature = args[3]
	userTx.Value = big.NewInt(0)
	return userTx
}

func (txProc *txProcessor) finishExecutionOfRelayedTx(
	relayerAcnt, acntDst state.UserAccountHandler,
	tx *transaction.Transaction,
	userTx *transaction.Transaction,
) (vmcommon.ReturnCode, error) {
	computedFees := txProc.computeRelayedTxFees(tx, userTx)
	err := txProc.processTxAtRelayer(relayerAcnt, computedFees.totalFee, computedFees.relayerFee, tx)
	if err != nil {
		return 0, err
	}

	if check.IfNil(acntDst) {
		return vmcommon.Ok, nil
	}

	err = txProc.addFeeAndValueToDest(acntDst, tx, computedFees.remainingFee)
	if err != nil {
		return 0, err
	}

	return txProc.processUserTx(tx, userTx, tx.Value, tx.Nonce)
}

func (txProc *txProcessor) processTxAtRelayer(
	relayerAcnt state.UserAccountHandler,
	totalFee *big.Int,
	relayerFee *big.Int,
	tx *transaction.Transaction,
) error {
	if !check.IfNil(relayerAcnt) {
		err := relayerAcnt.SubFromBalance(tx.GetValue())
		if err != nil {
			return err
		}

		err = relayerAcnt.SubFromBalance(totalFee)
		if err != nil {
			return err
		}

		relayerAcnt.IncreaseNonce(1)
		err = txProc.accounts.SaveAccount(relayerAcnt)
		if err != nil {
			return err
		}

		txHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
		if err != nil {
			return err
		}

		txProc.txFeeHandler.ProcessTransactionFee(relayerFee, big.NewInt(0), txHash)
	}

	return nil
}

func (txProc *txProcessor) addFeeAndValueToDest(acntDst state.UserAccountHandler, tx *transaction.Transaction, remainingFee *big.Int) error {
	err := acntDst.AddToBalance(tx.GetValue())
	if err != nil {
		return err
	}

	err = acntDst.AddToBalance(remainingFee)
	if err != nil {
		return err
	}

	return txProc.accounts.SaveAccount(acntDst)
}

func (txProc *txProcessor) processRelayedTxV3(
	tx *transaction.Transaction,
	relayerAcnt state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	if !txProc.enableEpochsHandler.IsFlagEnabled(common.RelayedTransactionsV3Flag) {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, process.ErrRelayedTxV3Disabled)
	}
	if check.IfNil(relayerAcnt) {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, process.ErrNilRelayerAccount)
	}
	err := txProc.relayedTxV3Processor.CheckRelayedTx(tx)
	if err != nil {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, err)
	}

	// process fees on both relayer and sender
	sendersBalancesSnapshot, err := txProc.processInnerTxsFeesAfterSnapshot(tx, relayerAcnt)
	if err != nil {
		txProc.resetBalancesToSnapshot(sendersBalancesSnapshot)
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, err)
	}

	innerTxs := tx.GetInnerTransactions()

	var innerTxRetCode vmcommon.ReturnCode
	var innerTxErr error
	executedUserTxs := make([]*transaction.Transaction, 0)
	for _, innerTx := range innerTxs {
		innerTxRetCode, innerTxErr = txProc.finishExecutionOfInnerTx(tx, innerTx)
		if innerTxErr != nil || innerTxRetCode != vmcommon.Ok {
			break
		}

		executedUserTxs = append(executedUserTxs, innerTx)
	}

	allUserTxsSucceeded := len(executedUserTxs) == len(innerTxs) && innerTxErr == nil && innerTxRetCode == vmcommon.Ok
	// if all user transactions were executed, return success
	if allUserTxsSucceeded {
		return vmcommon.Ok, nil
	}

	defer func() {
		// reset all senders to the snapshot took before starting the execution
		txProc.resetBalancesToSnapshot(sendersBalancesSnapshot)
	}()

	// if the first one failed, return last error
	// the current transaction should have been already reverted
	if len(executedUserTxs) == 0 {
		return innerTxRetCode, innerTxErr
	}

	originalTxHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
	if err != nil {
		return vmcommon.UserError, err
	}

	defer func() {
		executedHashed := make([][]byte, 0)
		for _, executedUserTx := range executedUserTxs {
			txHash, errHash := core.CalculateHash(txProc.marshalizer, txProc.hasher, executedUserTx)
			if errHash != nil {
				continue
			}
			executedHashed = append(executedHashed, txHash)
		}

		txProc.txFeeHandler.RevertFees(executedHashed)
	}()

	// if one or more user transactions were executed before one of them failed, revert all, including the fees transferred
	// the current transaction should have been already reverted
	var lastErr error
	revertedTxsCnt := 0
	for _, executedUserTx := range executedUserTxs {
		errRemove := txProc.removeValueAndConsumedFeeFromUser(executedUserTx, tx.Value, originalTxHash, tx, process.ErrSubsequentInnerTransactionFailed)
		if errRemove != nil {
			lastErr = errRemove
			continue
		}

		revertedTxsCnt++
	}

	if lastErr != nil {
		log.Warn("failed to revert all previous executed inner transactions, last error = %w, "+
			"total transactions = %d, num of transactions reverted = %d",
			lastErr,
			len(executedUserTxs),
			revertedTxsCnt)

		return vmcommon.UserError, lastErr
	}

	return vmcommon.UserError, process.ErrInvalidInnerTransactions
}

func (txProc *txProcessor) finishExecutionOfInnerTx(
	tx *transaction.Transaction,
	innerTx *transaction.Transaction,
) (vmcommon.ReturnCode, error) {
	acntSnd, err := txProc.getAccountFromAddress(innerTx.SndAddr)
	if err != nil {
		return vmcommon.UserError, err
	}

	if check.IfNil(acntSnd) {
		return vmcommon.Ok, nil
	}

	return txProc.processUserTx(tx, innerTx, tx.Value, tx.Nonce)
}

func (txProc *txProcessor) processRelayedTxV2(
	tx *transaction.Transaction,
	relayerAcnt, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	if !txProc.enableEpochsHandler.IsFlagEnabled(common.RelayedTransactionsV2Flag) {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, process.ErrRelayedTxV2Disabled)
	}
	if tx.GetValue().Cmp(big.NewInt(0)) != 0 {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, process.ErrRelayedTxV2ZeroVal)
	}

	_, args, err := txProc.argsParser.ParseCallData(string(tx.GetData()))
	if err != nil {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, err)
	}
	if len(args) != 4 {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, process.ErrInvalidArguments)
	}

	userTx := makeUserTxFromRelayedTxV2Args(args)
	userTx.GasPrice = tx.GasPrice
	userTx.GasLimit = tx.GasLimit - txProc.economicsFee.ComputeGasLimit(tx)
	userTx.SndAddr = tx.RcvAddr

	return txProc.finishExecutionOfRelayedTx(relayerAcnt, acntDst, tx, userTx)
}

func (txProc *txProcessor) processRelayedTx(
	tx *transaction.Transaction,
	relayerAcnt, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {

	_, args, err := txProc.argsParser.ParseCallData(string(tx.GetData()))
	if err != nil {
		return 0, err
	}
	if len(args) != 1 {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, process.ErrInvalidArguments)
	}
	if !txProc.enableEpochsHandler.IsFlagEnabled(common.RelayedTransactionsFlag) {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, process.ErrRelayedTxDisabled)
	}

	userTx := &transaction.Transaction{}
	err = txProc.signMarshalizer.Unmarshal(userTx, args[0])
	if err != nil {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, err)
	}
	if !bytes.Equal(userTx.SndAddr, tx.RcvAddr) {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, process.ErrRelayedTxBeneficiaryDoesNotMatchReceiver)
	}

	if userTx.Value.Cmp(tx.Value) < 0 {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, process.ErrRelayedTxValueHigherThenUserTxValue)
	}
	if userTx.GasPrice != tx.GasPrice {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, process.ErrRelayedGasPriceMissmatch)
	}

	remainingGasLimit := tx.GasLimit - txProc.economicsFee.ComputeGasLimit(tx)
	if userTx.GasLimit != remainingGasLimit {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, process.ErrRelayedTxGasLimitMissmatch)
	}

	return txProc.finishExecutionOfRelayedTx(relayerAcnt, acntDst, tx, userTx)
}

func (txProc *txProcessor) computeRelayedTxFees(tx, userTx *transaction.Transaction) relayedFees {
	relayerFee := txProc.economicsFee.ComputeMoveBalanceFee(tx)
	totalFee := txProc.economicsFee.ComputeTxFee(tx)
	if txProc.enableEpochsHandler.IsFlagEnabled(common.FixRelayedMoveBalanceFlag) {
		moveBalanceGasLimit := txProc.economicsFee.ComputeGasLimit(userTx)
		gasToUse := userTx.GetGasLimit() - moveBalanceGasLimit
		moveBalanceUserFee := txProc.economicsFee.ComputeMoveBalanceFee(userTx)
		processingUserFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, gasToUse)
		userFee := big.NewInt(0).Add(moveBalanceUserFee, processingUserFee)

		totalFee = totalFee.Add(relayerFee, userFee)
	}
	remainingFee := big.NewInt(0).Sub(totalFee, relayerFee)

	computedFees := relayedFees{
		totalFee:     totalFee,
		remainingFee: remainingFee,
		relayerFee:   relayerFee,
	}

	return computedFees
}

func (txProc *txProcessor) removeValueAndConsumedFeeFromUser(
	userTx *transaction.Transaction,
	relayedTxValue *big.Int,
	originalTxHash []byte,
	originalTx *transaction.Transaction,
	executionErr error,
) error {
	userAcnt, err := txProc.getAccountFromAddress(userTx.SndAddr)
	if err != nil {
		return err
	}
	if check.IfNil(userAcnt) {
		return process.ErrNilUserAccount
	}
	err = userAcnt.SubFromBalance(relayedTxValue)
	if err != nil {
		return err
	}

	consumedFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, userTx.GasLimit)
	if txProc.enableEpochsHandler.IsFlagEnabled(common.FixRelayedMoveBalanceFlag) {
		moveBalanceGasLimit := txProc.economicsFee.ComputeGasLimit(userTx)
		gasToUse := userTx.GetGasLimit() - moveBalanceGasLimit
		moveBalanceUserFee := txProc.economicsFee.ComputeMoveBalanceFee(userTx)
		processingUserFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, gasToUse)
		consumedFee = big.NewInt(0).Add(moveBalanceUserFee, processingUserFee)
	}
	err = userAcnt.SubFromBalance(consumedFee)
	if err != nil {
		return err
	}

	if txProc.shouldIncreaseNonce(executionErr) {
		userAcnt.IncreaseNonce(1)
	}

	err = txProc.addNonExecutableLog(executionErr, originalTxHash, originalTx)
	if err != nil {
		return err
	}

	err = txProc.accounts.SaveAccount(userAcnt)
	if err != nil {
		return err
	}

	return nil
}

func (txProc *txProcessor) addNonExecutableLog(executionErr error, originalTxHash []byte, originalTx data.TransactionHandler) error {
	if !isNonExecutableError(executionErr) {
		return nil
	}

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(core.SignalErrorOperation),
		Address:    originalTx.GetRcvAddr(),
	}

	return txProc.txLogsProcessor.SaveLog(originalTxHash, originalTx, []*vmcommon.LogEntry{logEntry})
}

func (txProc *txProcessor) processMoveBalanceCostRelayedUserTx(
	userTx *transaction.Transaction,
	userScr *smartContractResult.SmartContractResult,
	userAcc state.UserAccountHandler,
	originalTxHash []byte,
) error {
	moveBalanceGasLimit := txProc.economicsFee.ComputeGasLimit(userTx)
	moveBalanceUserFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, moveBalanceGasLimit)
	if txProc.enableEpochsHandler.IsFlagEnabled(common.FixRelayedMoveBalanceFlag) {
		gasToUse := userTx.GetGasLimit() - moveBalanceGasLimit
		moveBalanceUserFee = txProc.economicsFee.ComputeMoveBalanceFee(userTx)
		processingUserFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, gasToUse)
		moveBalanceUserFee = moveBalanceUserFee.Add(moveBalanceUserFee, processingUserFee)
	}

	userScrHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, userScr)
	if err != nil {
		return err
	}

	txProc.txFeeHandler.ProcessTransactionFeeRelayedUserTx(moveBalanceUserFee, big.NewInt(0), userScrHash, originalTxHash)
	return userAcc.SubFromBalance(moveBalanceUserFee)
}

func (txProc *txProcessor) processUserTx(
	originalTx *transaction.Transaction,
	userTx *transaction.Transaction,
	relayedTxValue *big.Int,
	relayedNonce uint64,
) (vmcommon.ReturnCode, error) {

	acntSnd, acntDst, err := txProc.getAccounts(userTx.SndAddr, userTx.RcvAddr)
	if err != nil {
		return 0, err
	}

	var originalTxHash []byte
	originalTxHash, err = core.CalculateHash(txProc.marshalizer, txProc.hasher, originalTx)
	if err != nil {
		return 0, err
	}

	relayerAdr := originalTx.SndAddr
	txType, dstShardTxType := txProc.txTypeHandler.ComputeTransactionType(userTx)
	err = txProc.checkTxValues(userTx, acntSnd, acntDst, true)
	if err != nil {
		errRemove := txProc.removeValueAndConsumedFeeFromUser(userTx, relayedTxValue, originalTxHash, originalTx, err)
		if errRemove != nil {
			return vmcommon.UserError, errRemove
		}
		return vmcommon.UserError, txProc.executeFailedRelayedUserTx(
			userTx,
			relayerAdr,
			relayedTxValue,
			relayedNonce,
			originalTx,
			originalTxHash,
			err.Error())
	}

	scrFromTx, err := txProc.makeSCRFromUserTx(userTx, relayerAdr, relayedTxValue, originalTxHash, false)
	if err != nil {
		return 0, err
	}

	returnCode := vmcommon.Ok
	switch txType {
	case process.MoveBalance:
		err = txProc.processMoveBalance(userTx, acntSnd, acntDst, dstShardTxType, originalTxHash, true)
	case process.SCDeployment:
		err = txProc.processMoveBalanceCostRelayedUserTx(userTx, scrFromTx, acntSnd, originalTxHash)
		if err != nil {
			break
		}

		returnCode, err = txProc.scProcessor.DeploySmartContract(scrFromTx, acntSnd)
	case process.SCInvoking:
		err = txProc.processMoveBalanceCostRelayedUserTx(userTx, scrFromTx, acntSnd, originalTxHash)
		if err != nil {
			break
		}

		returnCode, err = txProc.scProcessor.ExecuteSmartContractTransaction(scrFromTx, acntSnd, acntDst)
	case process.BuiltInFunctionCall:
		err = txProc.processMoveBalanceCostRelayedUserTx(userTx, scrFromTx, acntSnd, originalTxHash)
		if err != nil {
			break
		}

		returnCode, err = txProc.scProcessor.ExecuteBuiltInFunction(scrFromTx, acntSnd, acntDst)
	default:
		err = process.ErrWrongTransaction
		errRemove := txProc.removeValueAndConsumedFeeFromUser(userTx, relayedTxValue, originalTxHash, originalTx, err)
		if errRemove != nil {
			return vmcommon.UserError, errRemove
		}
		return vmcommon.UserError, txProc.executeFailedRelayedUserTx(
			userTx,
			relayerAdr,
			relayedTxValue,
			relayedNonce,
			originalTx,
			originalTxHash,
			err.Error())
	}

	if errors.Is(err, process.ErrInvalidMetaTransaction) || errors.Is(err, process.ErrAccountNotPayable) {
		return vmcommon.UserError, txProc.executeFailedRelayedUserTx(
			userTx,
			relayerAdr,
			relayedTxValue,
			relayedNonce,
			originalTx,
			originalTxHash,
			err.Error())
	}

	if errors.Is(err, process.ErrFailedTransaction) {
		// in case of failed inner user tx transaction we should just simply return execution failed and
		// not failed transaction - as the actual transaction (the relayed we correctly executed) and thus
		// it should not lend in the invalid miniblock
		return vmcommon.ExecutionFailed, nil
	}

	if err != nil {
		log.Error("processUserTx", "protocolError", err)
		return vmcommon.ExecutionFailed, err
	}

	// no need to add the smart contract result From TX to the intermediate transactions in case of error
	// returning value is resolved inside smart contract processor or above by executeFailedRelayedUserTx
	if returnCode != vmcommon.Ok {
		return returnCode, nil
	}

	err = txProc.scrForwarder.AddIntermediateTransactions([]data.TransactionHandler{scrFromTx})
	if err != nil {
		return 0, err
	}

	return vmcommon.Ok, nil
}

func (txProc *baseTxProcessor) isCrossTxFromMe(adrSrc, adrDst []byte) bool {
	shardForSrc := txProc.shardCoordinator.ComputeId(adrSrc)
	shardForDst := txProc.shardCoordinator.ComputeId(adrDst)
	shardForCurrentNode := txProc.shardCoordinator.SelfId()

	srcInShard := shardForSrc == shardForCurrentNode
	dstInShard := shardForDst == shardForCurrentNode

	return srcInShard && !dstInShard
}

func (txProc *txProcessor) makeSCRFromUserTx(
	tx *transaction.Transaction,
	relayerAdr []byte,
	relayedTxValue *big.Int,
	txHash []byte,
	isRevertSCR bool,
) (*smartContractResult.SmartContractResult, error) {
	scrValue := tx.Value
	if isRevertSCR {
		scrValue = big.NewInt(0).Neg(tx.Value)
	}
	scr := &smartContractResult.SmartContractResult{
		Nonce:          tx.Nonce,
		Value:          scrValue,
		RcvAddr:        tx.RcvAddr,
		SndAddr:        tx.SndAddr,
		RelayerAddr:    relayerAdr,
		RelayedValue:   big.NewInt(0).Set(relayedTxValue),
		Data:           tx.Data,
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
		GasLimit:       tx.GasLimit,
		GasPrice:       tx.GasPrice,
		CallType:       vm.DirectCall,
	}

	var err error
	scr.GasLimit, err = core.SafeSubUint64(scr.GasLimit, txProc.economicsFee.ComputeGasLimit(tx))
	if err != nil {
		return nil, err
	}

	return scr, nil
}

func (txProc *txProcessor) executeFailedRelayedUserTx(
	userTx *transaction.Transaction,
	relayerAdr []byte,
	relayedTxValue *big.Int,
	relayedNonce uint64,
	originalTx *transaction.Transaction,
	originalTxHash []byte,
	errorMsg string,
) error {
	scrForRelayer := &smartContractResult.SmartContractResult{
		Nonce:          relayedNonce,
		Value:          big.NewInt(0).Set(relayedTxValue),
		RcvAddr:        relayerAdr,
		SndAddr:        userTx.SndAddr,
		PrevTxHash:     originalTxHash,
		OriginalTxHash: originalTxHash,
		ReturnMessage:  []byte(errorMsg),
	}

	relayerAcnt, err := txProc.getAccountFromAddress(relayerAdr)
	if err != nil {
		return err
	}

	err = txProc.scrForwarder.AddIntermediateTransactions([]data.TransactionHandler{scrForRelayer})
	if err != nil {
		return err
	}

	totalFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, userTx.GasLimit)
	moveBalanceGasLimit := txProc.economicsFee.ComputeGasLimit(userTx)
	gasToUse := userTx.GetGasLimit() - moveBalanceGasLimit
	processingUserFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, gasToUse)
	if txProc.enableEpochsHandler.IsFlagEnabled(common.FixRelayedMoveBalanceFlag) {
		moveBalanceUserFee := txProc.economicsFee.ComputeMoveBalanceFee(userTx)
		totalFee = big.NewInt(0).Add(moveBalanceUserFee, processingUserFee)
	}

	senderShardID := txProc.shardCoordinator.ComputeId(userTx.SndAddr)
	if senderShardID != txProc.shardCoordinator.SelfId() {
		if txProc.enableEpochsHandler.IsFlagEnabled(common.FixRelayedMoveBalanceFlag) {
			totalFee.Sub(totalFee, processingUserFee)
		} else {
			moveBalanceUserFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, moveBalanceGasLimit)
			totalFee.Sub(totalFee, moveBalanceUserFee)
		}
	}

	txProc.txFeeHandler.ProcessTransactionFee(totalFee, big.NewInt(0), originalTxHash)

	if !check.IfNil(relayerAcnt) {
		err = relayerAcnt.AddToBalance(scrForRelayer.Value)
		if err != nil {
			return err
		}

		if txProc.enableEpochsHandler.IsFlagEnabled(common.AddFailedRelayedTxToInvalidMBsFlag) {
			err = txProc.badTxForwarder.AddIntermediateTransactions([]data.TransactionHandler{originalTx})
			if err != nil {
				return err
			}
		}

		err = txProc.accounts.SaveAccount(relayerAcnt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (txProc *txProcessor) shouldIncreaseNonce(executionErr error) bool {
	if !txProc.enableEpochsHandler.IsFlagEnabled(common.RelayedNonceFixFlag) {
		return true
	}

	if isNonExecutableError(executionErr) {
		return false
	}

	return true
}

func isNonExecutableError(executionErr error) bool {
	return errors.Is(executionErr, process.ErrLowerNonceInTransaction) ||
		errors.Is(executionErr, process.ErrHigherNonceInTransaction) ||
		errors.Is(executionErr, process.ErrTransactionNotExecutable)
}

func (txProc *txProcessor) processInnerTxsFeesAfterSnapshot(tx *transaction.Transaction, relayerAcnt state.UserAccountHandler) (map[state.UserAccountHandler]*big.Int, error) {
	relayerFee, totalFee := txProc.relayedTxV3Processor.ComputeRelayedTxFees(tx)
	err := txProc.processTxAtRelayer(relayerAcnt, totalFee, relayerFee, tx)
	if err != nil {
		return make(map[state.UserAccountHandler]*big.Int), err
	}

	uniqueSendersMap := txProc.relayedTxV3Processor.GetUniqueSendersRequiredFeesMap(tx.InnerTransactions)
	uniqueSendersSlice := mapToSlice(uniqueSendersMap)
	sendersBalancesSnapshot := make(map[state.UserAccountHandler]*big.Int, len(uniqueSendersMap))
	var lastTransferErr error
	for _, uniqueSender := range uniqueSendersSlice {
		totalFeesForSender := uniqueSendersMap[uniqueSender]
		senderAcnt, prevBalanceForSender, err := txProc.addFeesToDest([]byte(uniqueSender), totalFeesForSender)
		if err != nil {
			lastTransferErr = err
			break
		}

		sendersBalancesSnapshot[senderAcnt] = prevBalanceForSender
	}

	// if one error occurred, revert all transfers that succeeded and return error
	//if lastTransferErr != nil {
	//	for i := 0; i < lastIdx; i++ {
	//		uniqueSender := uniqueSendersSlice[i]
	//		totalFessSentForSender := uniqueSendersMap[uniqueSender]
	//		_, _, err = txProc.addFeesToDest([]byte(uniqueSender), big.NewInt(0).Neg(totalFessSentForSender))
	//		if err != nil {
	//			log.Warn("could not revert the fees transferred from relayer to sender",
	//				"sender", txProc.pubkeyConv.SilentEncode([]byte(uniqueSender), log),
	//				"relayer", txProc.pubkeyConv.SilentEncode(relayerAcnt.AddressBytes(), log))
	//		}
	//	}
	//}

	return sendersBalancesSnapshot, lastTransferErr
}

func (txProc *txProcessor) addFeesToDest(dstAddr []byte, feesForAllInnerTxs *big.Int) (state.UserAccountHandler, *big.Int, error) {
	acntDst, err := txProc.getAccountFromAddress(dstAddr)
	if err != nil {
		return nil, nil, err
	}

	if check.IfNil(acntDst) {
		return nil, nil, nil
	}

	prevBalance := acntDst.GetBalance()
	err = acntDst.AddToBalance(feesForAllInnerTxs)
	if err != nil {
		return nil, nil, err
	}

	return acntDst, prevBalance, txProc.accounts.SaveAccount(acntDst)
}

func (txProc *txProcessor) resetBalancesToSnapshot(snapshot map[state.UserAccountHandler]*big.Int) {
	for acnt, prevBalance := range snapshot {
		currentBalance := acnt.GetBalance()
		diff := big.NewInt(0).Sub(currentBalance, prevBalance)
		err := acnt.SubFromBalance(diff)
		if err != nil {
			log.Warn("could not reset sender to snapshot", "sender", txProc.pubkeyConv.SilentEncode(acnt.AddressBytes(), log))
			continue
		}

		err = txProc.accounts.SaveAccount(acnt)
		if err != nil {
			log.Warn("could not save account while resetting sender to snapshot", "sender", txProc.pubkeyConv.SilentEncode(acnt.AddressBytes(), log))
		}
	}
}

func mapToSlice(initialMap map[string]*big.Int) []string {
	newSlice := make([]string, 0, len(initialMap))
	for mapKey := range initialMap {
		newSlice = append(newSlice, mapKey)
	}

	return newSlice
}

// IsInterfaceNil returns true if there is no value under the interface
func (txProc *txProcessor) IsInterfaceNil() bool {
	return txProc == nil
}
