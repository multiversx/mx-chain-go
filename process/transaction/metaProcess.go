package transaction

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.TransactionProcessor = (*metaTxProcessor)(nil)

// txProcessor implements TransactionProcessor interface and can modify account states according to a transaction
type metaTxProcessor struct {
	*baseTxProcessor
	txTypeHandler   process.TxTypeHandler
	flagESDTEnabled atomic.Flag
	esdtEnableEpoch uint32
}

// ArgsNewMetaTxProcessor defines the arguments needed for new meta tx processor
type ArgsNewMetaTxProcessor struct {
	Hasher           hashing.Hasher
	Marshalizer      marshal.Marshalizer
	Accounts         state.AccountsAdapter
	PubkeyConv       core.PubkeyConverter
	ShardCoordinator sharding.Coordinator
	ScProcessor      process.SmartContractProcessor
	TxTypeHandler    process.TxTypeHandler
	EconomicsFee     process.FeeHandler
	ESDTEnableEpoch  uint32
	EpochNotifier    process.EpochNotifier
}

// NewMetaTxProcessor creates a new txProcessor engine
func NewMetaTxProcessor(args ArgsNewMetaTxProcessor) (*metaTxProcessor, error) {

	if check.IfNil(args.Accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.PubkeyConv) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.ScProcessor) {
		return nil, process.ErrNilSmartContractProcessor
	}
	if check.IfNil(args.TxTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(args.EconomicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	baseTxProcess := &baseTxProcessor{
		accounts:                args.Accounts,
		shardCoordinator:        args.ShardCoordinator,
		pubkeyConv:              args.PubkeyConv,
		economicsFee:            args.EconomicsFee,
		hasher:                  args.Hasher,
		marshalizer:             args.Marshalizer,
		scProcessor:             args.ScProcessor,
		flagPenalizedTooMuchGas: atomic.Flag{},
	}
	//backwards compatibility
	baseTxProcess.flagPenalizedTooMuchGas.Unset()

	txProc := &metaTxProcessor{
		baseTxProcessor: baseTxProcess,
		txTypeHandler:   args.TxTypeHandler,
		esdtEnableEpoch: args.ESDTEnableEpoch,
	}
	log.Debug("metaProcess: enable epoch for esdt", "epoch", txProc.esdtEnableEpoch)

	args.EpochNotifier.RegisterNotifyHandler(txProc)

	return txProc, nil
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

	err = txProc.checkTxValues(tx, acntSnd, acntDst, false)
	if err != nil {
		return 0, err
	}

	txType, _ := txProc.txTypeHandler.ComputeTransactionType(tx)

	switch txType {
	case process.SCDeployment:
		return txProc.processSCDeployment(tx, tx.SndAddr)
	case process.SCInvoking:
		return txProc.processSCInvoking(tx, tx.SndAddr, tx.RcvAddr)
	case process.BuiltInFunctionCall:
		if txProc.flagESDTEnabled.IsSet() {
			return txProc.processSCInvoking(tx, tx.SndAddr, tx.RcvAddr)
		}
	}

	snapshot := txProc.accounts.JournalLen()
	err = txProc.scProcessor.ProcessIfError(acntSnd, txHash, tx, process.ErrWrongTransaction.Error(), nil, snapshot, 0)
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

// EpochConfirmed is called whenever a new epoch is confirmed
func (txProc *metaTxProcessor) EpochConfirmed(epoch uint32, _ uint64) {
	txProc.flagESDTEnabled.Toggle(epoch >= txProc.esdtEnableEpoch)
	log.Debug("txProcessor: esdt", "enabled", txProc.flagESDTEnabled.IsSet())
}

// IsInterfaceNil returns true if there is no value under the interface
func (txProc *metaTxProcessor) IsInterfaceNil() bool {
	return txProc == nil
}
