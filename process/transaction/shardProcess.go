package transaction

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("process/transaction")
var _ process.TransactionProcessor = (*txProcessor)(nil)

// RefundGasMessage is the message returned in the data field of a receipt,
// for move balance transactions that provide more gas than needed
const RefundGasMessage = "refundedGas"

// txProcessor implements TransactionProcessor interface and can modify account states according to a transaction
type txProcessor struct {
	*baseTxProcessor
	txFeeHandler                   process.TransactionFeeHandler
	txTypeHandler                  process.TxTypeHandler
	receiptForwarder               process.IntermediateTransactionHandler
	badTxForwarder                 process.IntermediateTransactionHandler
	argsParser                     process.ArgumentsParser
	scrForwarder                   process.IntermediateTransactionHandler
	signMarshalizer                marshal.Marshalizer
	flagRelayedTx                  atomic.Flag
	flagPenalizedTooMuchGas        atomic.Flag
	flagMetaProtection             atomic.Flag
	relayedTxEnableEpoch           uint32
	penalizedTooMuchGasEnableEpoch uint32
	metaProtectionEnableEpoch      uint32
}

// ArgsNewTxProcessor defines the arguments needed for new tx processor
type ArgsNewTxProcessor struct {
	Accounts                       state.AccountsAdapter
	Hasher                         hashing.Hasher
	PubkeyConv                     core.PubkeyConverter
	Marshalizer                    marshal.Marshalizer
	SignMarshalizer                marshal.Marshalizer
	ShardCoordinator               sharding.Coordinator
	ScProcessor                    process.SmartContractProcessor
	TxFeeHandler                   process.TransactionFeeHandler
	TxTypeHandler                  process.TxTypeHandler
	EconomicsFee                   process.FeeHandler
	ReceiptForwarder               process.IntermediateTransactionHandler
	BadTxForwarder                 process.IntermediateTransactionHandler
	ArgsParser                     process.ArgumentsParser
	ScrForwarder                   process.IntermediateTransactionHandler
	RelayedTxEnableEpoch           uint32
	PenalizedTooMuchGasEnableEpoch uint32
	MetaProtectionEnableEpoch      uint32
	EpochNotifier                  process.EpochNotifier
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
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	baseTxProcess := &baseTxProcessor{
		accounts:         args.Accounts,
		shardCoordinator: args.ShardCoordinator,
		pubkeyConv:       args.PubkeyConv,
		economicsFee:     args.EconomicsFee,
		hasher:           args.Hasher,
		marshalizer:      args.Marshalizer,
		scProcessor:      args.ScProcessor,
	}

	txProc := &txProcessor{
		baseTxProcessor:                baseTxProcess,
		txFeeHandler:                   args.TxFeeHandler,
		txTypeHandler:                  args.TxTypeHandler,
		receiptForwarder:               args.ReceiptForwarder,
		badTxForwarder:                 args.BadTxForwarder,
		argsParser:                     args.ArgsParser,
		scrForwarder:                   args.ScrForwarder,
		signMarshalizer:                args.SignMarshalizer,
		relayedTxEnableEpoch:           args.RelayedTxEnableEpoch,
		penalizedTooMuchGasEnableEpoch: args.PenalizedTooMuchGasEnableEpoch,
		metaProtectionEnableEpoch:      args.MetaProtectionEnableEpoch,
	}

	args.EpochNotifier.RegisterNotifyHandler(txProc)

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

	process.DisplayProcessTxDetails(
		"ProcessTransaction: sender account details",
		acntSnd,
		tx,
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
		err = txProc.processMoveBalance(tx, acntSnd, acntDst, dstShardTxType, false)
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
	}

	return vmcommon.UserError, txProc.executingFailedTransaction(tx, acntSnd, process.ErrWrongTransaction)
}

func (txProc *txProcessor) executeAfterFailedMoveBalanceTransaction(
	tx *transaction.Transaction,
	txError error,
) error {
	acntSnd, err := txProc.getAccountFromAddress(tx.SndAddr)
	if err != nil {
		return err
	}

	if txError == process.ErrInvalidMetaTransaction || txError == process.ErrAccountNotPayable {
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

	txFee := txProc.economicsFee.ComputeTxFee(tx)
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
	shouldCreateReceiptBackwardCompatible := !txProc.flagMetaProtection.IsSet() && core.IsSmartContractAddress(tx.RcvAddr)
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
	isCrossShardSCCall := check.IfNil(acntDst) && len(tx.GetData()) > 0 && core.IsSmartContractAddress(tx.GetRcvAddr())
	if dstShardTxType != process.MoveBalance || (!txProc.flagMetaProtection.IsSet() && isCrossShardSCCall) {
		totalCost := txProc.economicsFee.ComputeTxFee(tx)
		if !txProc.flagPenalizedTooMuchGas.IsSet() {
			totalCost = core.SafeMul(tx.GasLimit, tx.GasPrice)
		}

		err := acntSnd.SubFromBalance(totalCost)
		if err != nil {
			return nil, nil, err
		}

		return moveBalanceFee, totalCost, nil
	}

	err := acntSnd.SubFromBalance(moveBalanceFee)
	if err != nil {
		return nil, nil, err
	}

	return moveBalanceFee, moveBalanceFee, nil
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

	if txProc.flagMetaProtection.IsSet() {
		// additional check
		if tx.GasLimit < txProc.economicsFee.ComputeGasLimit(tx)+core.MinMetaTxExtraGasCost {
			return process.ErrInvalidMetaTransaction
		}
	}

	return nil
}

func (txProc *txProcessor) processMoveBalance(
	tx *transaction.Transaction,
	acntSrc, acntDst state.UserAccountHandler,
	destShardTxType process.TransactionType,
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

	isPayable, err := txProc.scProcessor.IsPayable(tx.RcvAddr)
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

	txProc.txFeeHandler.ProcessTransactionFee(moveBalanceCost, big.NewInt(0), txHash)

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

	if !txProc.flagRelayedTx.IsSet() {
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

	totalFee, remainingFee, relayerFee, remainingGasLimit := txProc.computeRelayedTxFees(tx)
	if userTx.GasLimit != remainingGasLimit {
		return vmcommon.UserError, txProc.executingFailedTransaction(tx, relayerAcnt, process.ErrRelayedTxGasLimitMissmatch)
	}

	txHash, err := core.CalculateHash(txProc.marshalizer, txProc.hasher, tx)
	if err != nil {
		return 0, err
	}

	if !check.IfNil(relayerAcnt) {
		err = relayerAcnt.SubFromBalance(tx.GetValue())
		if err != nil {
			return 0, err
		}

		err = relayerAcnt.SubFromBalance(totalFee)
		if err != nil {
			return 0, err
		}

		relayerAcnt.IncreaseNonce(1)
		err = txProc.accounts.SaveAccount(relayerAcnt)
		if err != nil {
			return 0, err
		}

		txProc.txFeeHandler.ProcessTransactionFee(relayerFee, big.NewInt(0), txHash)
	}

	if check.IfNil(acntDst) {
		return vmcommon.Ok, nil
	}

	err = acntDst.AddToBalance(tx.GetValue())
	if err != nil {
		return 0, err
	}

	err = acntDst.AddToBalance(remainingFee)
	if err != nil {
		return 0, err
	}

	err = txProc.accounts.SaveAccount(acntDst)
	if err != nil {
		return 0, err
	}

	return txProc.processUserTx(tx, userTx, tx.Value, tx.Nonce, txHash)
}

func (txProc *txProcessor) computeRelayedTxFees(tx *transaction.Transaction) (*big.Int, *big.Int, *big.Int, uint64) {
	relayerGasLimit := txProc.economicsFee.ComputeGasLimit(tx)
	relayerFee := txProc.economicsFee.ComputeMoveBalanceFee(tx)
	totalFee := txProc.economicsFee.ComputeTxFee(tx)
	remainingFee := big.NewInt(0).Sub(totalFee, relayerFee)

	return totalFee, remainingFee, relayerFee, tx.GasLimit - relayerGasLimit
}

func (txProc *txProcessor) removeValueAndConsumedFeeFromUser(
	userTx *transaction.Transaction,
	relayedTxValue *big.Int,
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
	err = userAcnt.SubFromBalance(consumedFee)
	if err != nil {
		return err
	}
	userAcnt.IncreaseNonce(1)

	err = txProc.accounts.SaveAccount(userAcnt)
	if err != nil {
		return err
	}

	return nil
}

func (txProc *txProcessor) takeMoveBalanceCostOutOfUser(
	userTx *transaction.Transaction,
	userAcc state.UserAccountHandler,
) error {
	moveBalanceGasLimit := txProc.economicsFee.ComputeGasLimit(userTx)
	moveBalanceUserFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, moveBalanceGasLimit)
	return userAcc.SubFromBalance(moveBalanceUserFee)
}

func (txProc *txProcessor) processUserTx(
	originalTx *transaction.Transaction,
	userTx *transaction.Transaction,
	relayedTxValue *big.Int,
	relayedNonce uint64,
	txHash []byte,
) (vmcommon.ReturnCode, error) {

	acntSnd, acntDst, err := txProc.getAccounts(userTx.SndAddr, userTx.RcvAddr)
	if err != nil {
		return 0, err
	}

	relayerAdr := originalTx.SndAddr
	txType, dstShardTxType := txProc.txTypeHandler.ComputeTransactionType(userTx)
	err = txProc.checkTxValues(userTx, acntSnd, acntDst, true)
	if err != nil {
		errRemove := txProc.removeValueAndConsumedFeeFromUser(userTx, relayedTxValue)
		if errRemove != nil {
			return vmcommon.UserError, errRemove
		}
		return vmcommon.UserError, txProc.executeFailedRelayedTransaction(
			userTx,
			relayerAdr,
			relayedTxValue,
			relayedNonce,
			originalTx,
			txHash,
			err.Error())
	}

	scrFromTx, err := txProc.makeSCRFromUserTx(userTx, relayerAdr, relayedTxValue, txHash)
	if err != nil {
		return 0, err
	}

	returnCode := vmcommon.Ok
	switch txType {
	case process.MoveBalance:
		err = txProc.processMoveBalance(userTx, acntSnd, acntDst, dstShardTxType, true)
	case process.SCDeployment:
		err = txProc.takeMoveBalanceCostOutOfUser(userTx, acntSnd)
		if err != nil {
			break
		}

		returnCode, err = txProc.scProcessor.DeploySmartContract(scrFromTx, acntSnd)
	case process.SCInvoking:
		err = txProc.takeMoveBalanceCostOutOfUser(userTx, acntSnd)
		if err != nil {
			break
		}

		returnCode, err = txProc.scProcessor.ExecuteSmartContractTransaction(scrFromTx, acntSnd, acntDst)
	case process.BuiltInFunctionCall:
		err = txProc.takeMoveBalanceCostOutOfUser(userTx, acntSnd)
		if err != nil {
			break
		}

		returnCode, err = txProc.scProcessor.ExecuteBuiltInFunction(scrFromTx, acntSnd, acntDst)
	default:
		err = process.ErrWrongTransaction
		errRemove := txProc.removeValueAndConsumedFeeFromUser(userTx, relayedTxValue)
		if errRemove != nil {
			return vmcommon.UserError, errRemove
		}
		return vmcommon.UserError, txProc.executeFailedRelayedTransaction(
			userTx,
			relayerAdr,
			relayedTxValue,
			relayedNonce,
			originalTx,
			txHash,
			err.Error())
	}

	if errors.Is(err, process.ErrInvalidMetaTransaction) || errors.Is(err, process.ErrAccountNotPayable) {
		return vmcommon.UserError, txProc.executeFailedRelayedTransaction(
			userTx,
			relayerAdr,
			relayedTxValue,
			relayedNonce,
			originalTx,
			txHash,
			err.Error())
	}

	if err != nil {
		log.Error("processUserTx", "error", err)
		return vmcommon.ExecutionFailed, err
	}

	// no need to add the smart contract result From TX to the intermediate transactions in case of error
	// returning value is resolved inside smart contract processor or above by executeFailedRelayedTransaction
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
) (*smartContractResult.SmartContractResult, error) {
	scr := &smartContractResult.SmartContractResult{
		Nonce:          tx.Nonce,
		Value:          tx.Value,
		RcvAddr:        tx.RcvAddr,
		SndAddr:        tx.SndAddr,
		RelayerAddr:    relayerAdr,
		RelayedValue:   big.NewInt(0).Set(relayedTxValue),
		Data:           tx.Data,
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
		GasLimit:       tx.GasLimit,
		GasPrice:       tx.GasPrice,
		CallType:       vmcommon.DirectCall,
	}

	var err error
	scr.GasLimit, err = core.SafeSubUint64(scr.GasLimit, txProc.economicsFee.ComputeGasLimit(tx))
	if err != nil {
		return nil, err
	}

	return scr, nil
}

func (txProc *txProcessor) executeFailedRelayedTransaction(
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
	senderShardID := txProc.shardCoordinator.ComputeId(userTx.SndAddr)
	if senderShardID != txProc.shardCoordinator.SelfId() {
		moveBalanceGasLimit := txProc.economicsFee.ComputeGasLimit(userTx)
		moveBalanceUserFee := txProc.economicsFee.ComputeFeeForProcessing(userTx, moveBalanceGasLimit)
		totalFee.Sub(totalFee, moveBalanceUserFee)
	}

	txProc.txFeeHandler.ProcessTransactionFee(totalFee, big.NewInt(0), originalTxHash)

	if !check.IfNil(relayerAcnt) {
		err = relayerAcnt.AddToBalance(scrForRelayer.Value)
		if err != nil {
			return err
		}

		err = txProc.badTxForwarder.AddIntermediateTransactions([]data.TransactionHandler{originalTx})
		if err != nil {
			return err
		}

		err = txProc.accounts.SaveAccount(relayerAcnt)
		if err != nil {
			return err
		}

		return process.ErrFailedTransaction
	}

	return nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (txProc *txProcessor) EpochConfirmed(epoch uint32) {
	txProc.flagRelayedTx.Toggle(epoch >= txProc.relayedTxEnableEpoch)
	log.Debug("txProcessor: relayed transactions", "enabled", txProc.flagRelayedTx.IsSet())

	txProc.flagPenalizedTooMuchGas.Toggle(epoch >= txProc.penalizedTooMuchGasEnableEpoch)
	log.Debug("txProcessor: penalized too much gas", "enabled", txProc.flagPenalizedTooMuchGas.IsSet())

	txProc.flagMetaProtection.Toggle(epoch >= txProc.metaProtectionEnableEpoch)
	log.Debug("txProcessor: meta protection", "enabled", txProc.flagMetaProtection.IsSet())
}

// IsInterfaceNil returns true if there is no value under the interface
func (txProc *txProcessor) IsInterfaceNil() bool {
	return txProc == nil
}
