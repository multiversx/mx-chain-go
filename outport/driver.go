package outport

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/outport/marshaling"
)

var log = logger.GetOrCreate("outport")

var _ Driver = (*outportDriver)(nil)

type outportDriver struct {
	config        config.OutportConfig
	txCoordinator TransactionCoordinator
	logsProcessor TransactionLogProcessor
}

// NewOutportDriver creates a new outport driver
func NewOutportDriver(
	config config.OutportConfig,
	txCoordinator TransactionCoordinator,
	logsProcessor TransactionLogProcessor,
	marshalizer marshaling.Marshalizer,
) (*outportDriver, error) {
	if check.IfNil(txCoordinator) {
		return nil, ErrNilTxCoordinator
	}
	if check.IfNil(logsProcessor) {
		return nil, ErrNilLogsProcessor
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}

	return &outportDriver{
		config:        config,
		txCoordinator: txCoordinator,
		logsProcessor: logsProcessor,
	}, nil
}

// DigestBlock digests a block
func (driver *outportDriver) DigestCommittedBlock(header data.HeaderHandler, body data.BodyHandler) {
	if check.IfNil(header) {
		return
	}
	if check.IfNil(body) {
		return
	}

	txPool := driver.txCoordinator.GetAllCurrentUsedTxs(block.TxBlock)
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
func (driver *outportDriver) IsInterfaceNil() bool {
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
