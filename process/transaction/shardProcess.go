package transaction

import (
	"math/big"

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
) (*txProcessor, error) {

	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if addressConv == nil || addressConv.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if scProcessor == nil || scProcessor.IsInterfaceNil() {
		return nil, process.ErrNilSmartContractProcessor
	}
	if txFeeHandler == nil || txFeeHandler.IsInterfaceNil() {
		return nil, process.ErrNilUnsignedTxHandler
	}
	if txTypeHandler == nil || txTypeHandler.IsInterfaceNil() {
		return nil, process.ErrNilTxTypeHandler
	}
	if economicsFee == nil || economicsFee.IsInterfaceNil() {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	baseTxProcess := &baseTxProcessor{
		accounts:         accounts,
		shardCoordinator: shardCoordinator,
		adrConv:          addressConv,
	}

	return &txProcessor{
		baseTxProcessor: baseTxProcess,
		hasher:          hasher,
		marshalizer:     marshalizer,
		scProcessor:     scProcessor,
		txFeeHandler:    txFeeHandler,
		txTypeHandler:   txTypeHandler,
		economicsFee:    economicsFee,
	}, nil
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (txProc *txProcessor) ProcessTransaction(tx *transaction.Transaction, roundIndex uint64) error {
	if tx == nil || tx.IsInterfaceNil() {
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
		return txProc.processSCDeployment(tx, adrSrc, roundIndex)
	case process.SCInvoking:
		return txProc.processSCInvoking(tx, adrSrc, adrDst, roundIndex)
	}

	return process.ErrWrongTransaction
}

func (txProc *txProcessor) processTxFee(tx *transaction.Transaction, acntSnd *state.Account) (*big.Int, error) {
	if acntSnd == nil {
		return nil, nil
	}

	err := txProc.economicsFee.CheckValidityTxValues(tx)
	if err != nil {
		return nil, err
	}

	cost := txProc.economicsFee.ComputeFee(tx)
	if acntSnd.Balance.Cmp(cost) < 0 {
		return nil, process.ErrInsufficientFunds
	}

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

	value := tx.GetValue()

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

	txProc.txFeeHandler.ProcessTransactionFee(txFee)

	return nil
}

func (txProc *txProcessor) processSCDeployment(
	tx *transaction.Transaction,
	adrSrc state.AddressContainer,
	roundIndex uint64,
) error {
	// getAccounts returns acntSrc not nil if the adrSrc is in the node shard, the same, acntDst will be not nil
	// if adrDst is in the node shard. If an error occurs it will be signaled in err variable.
	acntSrc, err := txProc.getAccountFromAddress(adrSrc)
	if err != nil {
		return err
	}

	err = txProc.scProcessor.DeploySmartContract(tx, acntSrc, roundIndex)
	return err
}

func (txProc *txProcessor) processSCInvoking(
	tx *transaction.Transaction,
	adrSrc, adrDst state.AddressContainer,
	roundIndex uint64,
) error {
	// getAccounts returns acntSrc not nil if the adrSrc is in the node shard, the same, acntDst will be not nil
	// if adrDst is in the node shard. If an error occurs it will be signaled in err variable.
	acntSrc, acntDst, err := txProc.getAccounts(adrSrc, adrDst)
	if err != nil {
		return err
	}

	err = txProc.scProcessor.ExecuteSmartContractTransaction(tx, acntSrc, acntDst, roundIndex)
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
