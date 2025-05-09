package transaction

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ process.TransactionProcessor = (*metaTxProcessor)(nil)

// txProcessor implements TransactionProcessor interface and can modify account states according to a transaction
type metaTxProcessor struct {
	*baseTxProcessor
	enableEpochsHandler common.EnableEpochsHandler
}

// ArgsNewMetaTxProcessor defines the arguments needed for new meta tx processor
type ArgsNewMetaTxProcessor struct {
	Hasher              hashing.Hasher
	Marshalizer         marshal.Marshalizer
	Accounts            state.AccountsAdapter
	PubkeyConv          core.PubkeyConverter
	ShardCoordinator    sharding.Coordinator
	ScProcessor         process.SmartContractProcessor
	TxTypeHandler       process.TxTypeHandler
	EconomicsFee        process.FeeHandler
	EnableEpochsHandler common.EnableEpochsHandler
	TxVersionChecker    process.TxVersionCheckerHandler
	GuardianChecker     process.GuardianChecker
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
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.PenalizedTooMuchGasFlag,
		common.ESDTFlag,
		common.FixRelayedBaseCostFlag,
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
		txTypeHandler:       args.TxTypeHandler,
	}

	txProc := &metaTxProcessor{
		baseTxProcessor:     baseTxProcess,
		enableEpochsHandler: args.EnableEpochsHandler,
	}

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
		txHash,
		txProc.pubkeyConv,
	)

	err = txProc.checkTxValues(tx, acntSnd, acntDst, false)
	if err != nil {
		if errors.Is(err, process.ErrUserNameDoesNotMatchInCrossShardTx) {
			errProcessIfErr := txProc.processIfTxErrorCrossShard(tx, err.Error())
			if errProcessIfErr != nil {
				return 0, errProcessIfErr
			}
			return vmcommon.UserError, nil
		}
		return 0, err
	}

	txType, _, _ := txProc.txTypeHandler.ComputeTransactionType(tx)
	switch txType {
	case process.SCDeployment:
		return txProc.processSCDeployment(tx, tx.SndAddr)
	case process.SCInvoking:
		return txProc.processSCInvoking(tx, tx.SndAddr, tx.RcvAddr)
	case process.BuiltInFunctionCall:
		if txProc.enableEpochsHandler.IsFlagEnabled(common.ESDTFlag) {
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

// IsInterfaceNil returns true if there is no value under the interface
func (txProc *metaTxProcessor) IsInterfaceNil() bool {
	return txProc == nil
}
