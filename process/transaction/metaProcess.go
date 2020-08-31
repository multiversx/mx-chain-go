package transaction

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ process.TransactionProcessor = (*metaTxProcessor)(nil)

// txProcessor implements TransactionProcessor interface and can modify account states according to a transaction
type metaTxProcessor struct {
	*baseTxProcessor
	txTypeHandler process.TxTypeHandler
}

// NewMetaTxProcessor creates a new txProcessor engine
func NewMetaTxProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accounts state.AccountsAdapter,
	pubkeyConv core.PubkeyConverter,
	shardCoordinator sharding.Coordinator,
	scProcessor process.SmartContractProcessor,
	txTypeHandler process.TxTypeHandler,
	economicsFee process.FeeHandler,
	relayedTxEnableEpoch uint32,
	penalizedTooMuchGasEnableEpoch uint32,
	epochNotifier process.EpochNotifier,
) (*metaTxProcessor, error) {

	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(pubkeyConv) {
		return nil, process.ErrNilPubkeyConverter
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
	if check.IfNil(epochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	baseTxProcess := &baseTxProcessor{
		accounts:                       accounts,
		shardCoordinator:               shardCoordinator,
		pubkeyConv:                     pubkeyConv,
		economicsFee:                   economicsFee,
		hasher:                         hasher,
		marshalizer:                    marshalizer,
		scProcessor:                    scProcessor,
		relayedTxEnableEpoch:           relayedTxEnableEpoch,
		penalizedTooMuchGasEnableEpoch: penalizedTooMuchGasEnableEpoch,
	}

	metaTxProc := &metaTxProcessor{
		baseTxProcessor: baseTxProcess,
		txTypeHandler:   txTypeHandler,
	}

	epochNotifier.RegisterNotifyHandler(metaTxProc)

	return metaTxProc, nil
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (txProc *metaTxProcessor) ProcessTransaction(tx *transaction.Transaction) (vmcommon.ReturnCode, error) {
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
		txProc.pubkeyConv,
	)

	err = txProc.checkTxValues(tx, acntSnd, acntDst)
	if err != nil {
		return 0, err
	}

	txType := txProc.txTypeHandler.ComputeTransactionType(tx)

	switch txType {
	case process.SCDeployment:
		return txProc.processSCDeployment(tx, tx.SndAddr)
	case process.SCInvoking:
		return txProc.processSCInvoking(tx, tx.SndAddr, tx.RcvAddr)
	}

	snapshot := txProc.accounts.JournalLen()
	err = txProc.scProcessor.ProcessIfError(acntSnd, txHash, tx, process.ErrWrongTransaction.Error(), nil, snapshot)
	if err != nil {
		return 0, err
	}

	return vmcommon.UserError, nil
}

func (txProc *metaTxProcessor) processSCDeployment(
	tx *transaction.Transaction,
	adrSrc []byte,
) (vmcommon.ReturnCode, error) {
	// getAccounts returns acntSrc not nil if the adrSrc is in the node shard, the same, acntDst will be not nil
	// if adrDst is in the node shard. If an error occurs it will be signaled in err variable.
	acntSrc, err := txProc.getAccountFromAddress(adrSrc)
	if err != nil {
		return 0, err
	}

	return txProc.scProcessor.DeploySmartContract(tx, acntSrc)
}

func (txProc *metaTxProcessor) processSCInvoking(
	tx *transaction.Transaction,
	adrSrc, adrDst []byte,
) (vmcommon.ReturnCode, error) {
	// getAccounts returns acntSrc not nil if the adrSrc is in the node shard, the same, acntDst will be not nil
	// if adrDst is in the node shard. If an error occurs it will be signaled in err variable.
	acntSrc, acntDst, err := txProc.getAccounts(adrSrc, adrDst)
	if err != nil {
		return 0, err
	}

	return txProc.scProcessor.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
}

// IsInterfaceNil returns true if there is no value under the interface
func (txProc *metaTxProcessor) IsInterfaceNil() bool {
	return txProc == nil
}
