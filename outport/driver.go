package outport

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/outport/marshaling"
	"github.com/ElrondNetwork/elrond-go/outport/messages"
)

var log = logger.GetOrCreate("outport")

var _ Driver = (*outportDriver)(nil)

type mapOfTxs = map[string]data.TransactionHandler

type outportDriver struct {
	config        config.OutportConfig
	txCoordinator TransactionCoordinator
	logsProcessor TransactionLogProcessor
	sender        sender
}

// newOutportDriver creates a new outport driver
func newOutportDriver(
	config config.OutportConfig,
	txCoordinator TransactionCoordinator,
	logsProcessor TransactionLogProcessor,
	sender sender,
) (*outportDriver, error) {
	if check.IfNil(txCoordinator) {
		return nil, ErrNilTxCoordinator
	}
	if check.IfNil(logsProcessor) {
		return nil, ErrNilLogsProcessor
	}
	if check.IfNil(sender) {
		return nil, ErrNilSender
	}

	return &outportDriver{
		config:        config,
		txCoordinator: txCoordinator,
		logsProcessor: logsProcessor,
		sender:        sender,
	}, nil
}

// DigestCommittedBlock digests a block
func (driver *outportDriver) DigestCommittedBlock(headerHash []byte, header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	message := messages.NewMessageCommittedBlock(header)
	message.RegularTransactions = marshaling.NewSerializableMapStringTransactionHandler(driver.getRegularTransactions())
	message.SmartContractResults = marshaling.NewSerializableMapStringTransactionHandler(driver.getSmartContractResults())
	message.RewardTransactions = marshaling.NewSerializableMapStringTransactionHandler(driver.getRewardTransactions())
	message.InvalidTransactions = marshaling.NewSerializableMapStringTransactionHandler(driver.getInvalidTransactions())
	message.Receipts = marshaling.NewSerializableMapStringTransactionHandler(driver.getReceipts())
	message.SmartContractLogs = nil

	length, err := driver.sender.Send(message)
	if err != nil {
		log.Warn("DigestCommittedBlock(): could not send message", "err", err)
	} else {
		log.Trace("DigestCommittedBlock(): message sent", "length", length)
	}
}

func (driver *outportDriver) getRegularTransactions() mapOfTxs {
	filter := driver.config.Filter
	if !filter.WithRegularTransactions {
		return make(mapOfTxs)
	}

	return driver.txCoordinator.GetAllCurrentUsedTxs(block.TxBlock)
}

func (driver *outportDriver) getSmartContractResults() mapOfTxs {
	filter := driver.config.Filter
	if !filter.WithSmartContractResults {
		return make(mapOfTxs)
	}

	return driver.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)
}

func (driver *outportDriver) getRewardTransactions() mapOfTxs {
	filter := driver.config.Filter
	if !filter.WithRewardTransactions {
		return make(mapOfTxs)
	}

	return driver.txCoordinator.GetAllCurrentUsedTxs(block.RewardsBlock)
}

func (driver *outportDriver) getInvalidTransactions() mapOfTxs {
	filter := driver.config.Filter
	if !filter.WithInvalidTransactions {
		return make(mapOfTxs)
	}

	return driver.txCoordinator.GetAllCurrentUsedTxs(block.InvalidBlock)
}

func (driver *outportDriver) getReceipts() mapOfTxs {
	filter := driver.config.Filter
	if !filter.WithRewardTransactions {
		return make(mapOfTxs)
	}

	return driver.txCoordinator.GetAllCurrentUsedTxs(block.ReceiptBlock)
}

func (driver *outportDriver) getSmartContractLogs() []data.LogHandler {
	filter := driver.config.Filter
	if !filter.WithSmartContractLogs {
		return make([]data.LogHandler, 0)
	}

	// TODO: return driver.logsProcessor.GetLog(...)
	// TODO: Get all instead.
	return make([]data.LogHandler, 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (driver *outportDriver) IsInterfaceNil() bool {
	return driver == nil
}
