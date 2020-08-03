package outport

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
)

var log = logger.GetOrCreate("outport")

var _ Driver = (*OutportDriver)(nil)

type OutportDriver struct {
	txCoordinator TransactionCoordinator
	logsProcessor TransactionLogProcessor
}

func NewOutportDriver(txCoordinator TransactionCoordinator, logsProcessor TransactionLogProcessor) *OutportDriver {
	return &OutportDriver{
		txCoordinator: txCoordinator,
		logsProcessor: logsProcessor,
	}
}

// DigestBlock digests a block
func (driver *OutportDriver) DigestCommittedBlock(header data.HeaderHandler, body data.BodyHandler) {
	if check.IfNil(header) {
		return
	}
	if check.IfNil(body) {
		return
	}

	// txPool := txCoordinator.GetAllCurrentUsedTxs(block.TxBlock)
	// scPool := txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)
	// rewardPool := txCoordinator.GetAllCurrentUsedTxs(block.RewardsBlock)
	// invalidPool := txCoordinator.GetAllCurrentUsedTxs(block.InvalidBlock)
	// receiptPool := txCoordinator.GetAllCurrentUsedTxs(block.ReceiptBlock)

	// fmt.Println("txPool", txPool)
	// fmt.Println("scPool", scPool)
	// fmt.Println("rewardPool", rewardPool)
	// fmt.Println("invalidPool", invalidPool)
	// fmt.Println("receiptPool", receiptPool)
}

// IsInterfaceNil returns true if there is no value under the interface
func (driver *OutportDriver) IsInterfaceNil() bool {
	return driver == nil
}

// // TransactionsToDigest holds current transactions to digest
// type TransactionsToDigest struct {
// 	RegularTxs  map[string]data.TransactionHandler
// 	RewardTxs   map[string]data.TransactionHandler
// 	ScResults   map[string]data.TransactionHandler
// 	InvalidTxs  map[string]data.TransactionHandler
// 	ReceiptsTxs map[string]data.TransactionHandler
// }
