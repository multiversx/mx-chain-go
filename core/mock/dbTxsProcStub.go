package mock

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	dataProcess "github.com/ElrondNetwork/elrond-go/process"
)

// DBTxsProcStub -
type DBTxsProcStub struct {
	PrepareTransactionsForDatabaseCalled func(
		body *block.Body,
		header data.HeaderHandler,
		txPool map[string]data.TransactionHandler,
	) ([]*types.Transaction, []*types.ScResult, []*types.Receipt, map[string]*types.AlteredAccount)

	SerializeReceiptsCalled     func(receipts []*types.Receipt, bulkSizeThreshold int) ([]bytes.Buffer, error)
	SerializeTransactionsCalled func(transactions []*types.Transaction, selfShardID uint32, mbsHashInDB map[string]bool, bulkSizeThreshold int) ([]bytes.Buffer, error)
	SerializeScResultsCalled    func(scResults []*types.ScResult, bulkSizeThreshold int) ([]bytes.Buffer, error)
}

// PrepareTransactionsForDatabase -
func (d *DBTxsProcStub) PrepareTransactionsForDatabase(
	body *block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
) ([]*types.Transaction, []*types.ScResult, []*types.Receipt, map[string]*types.AlteredAccount) {
	if d.PrepareTransactionsForDatabaseCalled != nil {
		return d.PrepareTransactionsForDatabaseCalled(body, header, txPool)
	}

	return nil, nil, nil, nil
}

// SetTxLogsProcessor -
func (d *DBTxsProcStub) SetTxLogsProcessor(_ dataProcess.TransactionLogProcessorDatabase) {
}

// SerializeReceipts -
func (d *DBTxsProcStub) SerializeReceipts(receipts []*types.Receipt, bulkSizeThreshold int) ([]bytes.Buffer, error) {
	if d.SerializeReceiptsCalled != nil {
		return d.SerializeReceiptsCalled(receipts, bulkSizeThreshold)
	}

	return nil, nil
}

// SerializeTransactions -
func (d *DBTxsProcStub) SerializeTransactions(transactions []*types.Transaction, selfShardID uint32, mbsHashInDB map[string]bool, bulkSizeThreshold int) ([]bytes.Buffer, error) {
	if d.SerializeTransactionsCalled != nil {
		return d.SerializeTransactionsCalled(transactions, selfShardID, mbsHashInDB, bulkSizeThreshold)
	}

	return nil, nil
}

// SerializeScResults -
func (d *DBTxsProcStub) SerializeScResults(scResults []*types.ScResult, bulkSizeThreshold int) ([]bytes.Buffer, error) {
	if d.SerializeScResultsCalled != nil {
		return d.SerializeScResultsCalled(scResults, bulkSizeThreshold)
	}

	return nil, nil
}
