package transaction

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// txProcessor implements TransactionProcessor interface and can modify account states according to a transaction
type metaTxProcessor struct {
	*baseTxProcessor
	txTypeHandler process.TxTypeHandler
	scProcessor   process.SmartContractProcessor
}

// NewMetaTxProcessor creates a new txProcessor engine
func NewMetaTxProcessor(
	accounts state.AccountsAdapter,
	addressConv state.AddressConverter,
	shardCoordinator sharding.Coordinator,
	scProcessor process.SmartContractProcessor,
	txTypeHandler process.TxTypeHandler,
) (*metaTxProcessor, error) {

	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if addressConv == nil || addressConv.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if scProcessor == nil || scProcessor.IsInterfaceNil() {
		return nil, process.ErrNilSmartContractProcessor
	}
	if txTypeHandler == nil || txTypeHandler.IsInterfaceNil() {
		return nil, process.ErrNilTxTypeHandler
	}

	baseTxProcess := &baseTxProcessor{
		accounts:         accounts,
		shardCoordinator: shardCoordinator,
		adrConv:          addressConv,
	}

	return &metaTxProcessor{
		baseTxProcessor: baseTxProcess,
		scProcessor:     scProcessor,
		txTypeHandler:   txTypeHandler,
	}, nil
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (txProc *metaTxProcessor) ProcessTransaction(tx *transaction.Transaction, roundIndex uint64) error {
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
	case process.SCDeployment:
		return txProc.processSCDeployment(tx, adrSrc, roundIndex)
	case process.SCInvoking:
		return txProc.processSCInvoking(tx, adrSrc, adrDst, roundIndex)
	}

	return process.ErrWrongTransaction
}

func (txProc *metaTxProcessor) processSCDeployment(
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

func (txProc *metaTxProcessor) processSCInvoking(
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

// IsInterfaceNil returns true if there is no value under the interface
func (txProc *metaTxProcessor) IsInterfaceNil() bool {
	if txProc == nil {
		return true
	}
	return false
}
