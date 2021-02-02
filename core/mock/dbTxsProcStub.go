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

	SerializeReceiptsCalled             func(receipts []*types.Receipt) ([]*bytes.Buffer, error)
	SerializeTransactionsCalled         func(transactions []*types.Transaction, selfShardID uint32, mbsHashInDB map[string]bool) ([]*bytes.Buffer, error)
	SerializeScResultsCalled            func(scResults []*types.ScResult) ([]*bytes.Buffer, error)
	GetRewardsTxsHashesHexEncodedCalled func(header data.HeaderHandler, body *block.Body) []string
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
func (d *DBTxsProcStub) SerializeReceipts(receipts []*types.Receipt) ([]*bytes.Buffer, error) {
	if d.SerializeReceiptsCalled != nil {
		return d.SerializeReceiptsCalled(receipts)
	}

	return nil, nil
}

// SerializeTransactions -
func (d *DBTxsProcStub) SerializeTransactions(transactions []*types.Transaction, selfShardID uint32, mbsHashInDB map[string]bool) ([]*bytes.Buffer, error) {
	if d.SerializeTransactionsCalled != nil {
		return d.SerializeTransactionsCalled(transactions, selfShardID, mbsHashInDB)
	}

	return nil, nil
}

// SerializeScResults -
func (d *DBTxsProcStub) SerializeScResults(scResults []*types.ScResult) ([]*bytes.Buffer, error) {
	if d.SerializeScResultsCalled != nil {
		return d.SerializeScResultsCalled(scResults)
	}

	return nil, nil
}

// GetRewardsTxsHashesHexEncoded -
func (d *DBTxsProcStub) GetRewardsTxsHashesHexEncoded(header data.HeaderHandler, body *block.Body) []string {
	if d.GetRewardsTxsHashesHexEncodedCalled != nil {
		return d.GetRewardsTxsHashesHexEncodedCalled(header, body)
	}

	return nil
}
