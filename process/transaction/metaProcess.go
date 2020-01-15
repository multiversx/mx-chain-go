package transaction

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
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
	economicsFee process.FeeHandler,
) (*metaTxProcessor, error) {

	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(addressConv) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(scProcessor) {
		return nil, process.ErrNilSmartContractProcessor
	}
	if check.IfNil(txTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	baseTxProcess := &baseTxProcessor{
		accounts:         accounts,
		shardCoordinator: shardCoordinator,
		adrConv:          addressConv,
		economicsFee:     economicsFee,
	}

	return &metaTxProcessor{
		baseTxProcessor: baseTxProcess,
		scProcessor:     scProcessor,
		txTypeHandler:   txTypeHandler,
	}, nil
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (txProc *metaTxProcessor) ProcessTransaction(tx *transaction.Transaction) error {
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
		return txProc.processSCDeployment(tx, adrSrc)
	case process.SCInvoking:
		return txProc.processSCInvoking(tx, adrSrc, adrDst)
	}

	return process.ErrWrongTransaction
}

func (txProc *metaTxProcessor) processSCDeployment(
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

func (txProc *metaTxProcessor) processSCInvoking(
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

// IsInterfaceNil returns true if there is no value under the interface
func (txProc *metaTxProcessor) IsInterfaceNil() bool {
	return txProc == nil
}
