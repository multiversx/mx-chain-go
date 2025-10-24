package preprocMocks

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

// TxsForBlockStub -
type TxsForBlockStub struct {
	ResetCalled                            func()
	AddTransactionCalled                   func(txHash []byte, tx data.TransactionHandler, senderShardID uint32, receiverShardID uint32)
	WaitForRequestedDataCalled             func(waitTime time.Duration) error
	GetTxInfoByHashCalled                  func(hash []byte) (*process.TxInfo, bool)
	GetAllCurrentUsedTxsCalled             func() map[string]data.TransactionHandler
	GetMissingTxsCountCalled               func() int
	ReceivedTransactionCalled              func(txHash []byte, tx data.TransactionHandler)
	HasMissingTransactionsCalled           func() bool
	ComputeExistingAndRequestMissingCalled func(body *block.Body, isMiniBlockCorrect func(block.Type) bool, txPool dataRetriever.ShardedDataCacherNotifier, onRequestTxs func(shardID uint32, txHashes [][]byte)) int
}

// Reset -
func (tfbs *TxsForBlockStub) Reset() {
	if tfbs.ResetCalled != nil {
		tfbs.ResetCalled()
	}
}

// AddTransaction -
func (tfbs *TxsForBlockStub) AddTransaction(txHash []byte, tx data.TransactionHandler, senderShardID uint32, receiverShardID uint32) {
	if tfbs.AddTransactionCalled != nil {
		tfbs.AddTransactionCalled(txHash, tx, senderShardID, receiverShardID)
	}
}

// WaitForRequestedData -
func (tfbs *TxsForBlockStub) WaitForRequestedData(waitTime time.Duration) error {
	if tfbs.WaitForRequestedDataCalled != nil {
		return tfbs.WaitForRequestedDataCalled(waitTime)
	}
	return nil
}

// GetTxInfoByHash -
func (tfbs *TxsForBlockStub) GetTxInfoByHash(hash []byte) (*process.TxInfo, bool) {
	if tfbs.GetTxInfoByHashCalled != nil {
		return tfbs.GetTxInfoByHashCalled(hash)
	}
	return nil, false
}

// GetAllCurrentUsedTxs -
func (tfbs *TxsForBlockStub) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	if tfbs.GetAllCurrentUsedTxsCalled != nil {
		return tfbs.GetAllCurrentUsedTxsCalled()
	}
	return nil
}

// GetMissingTxsCount -
func (tfbs *TxsForBlockStub) GetMissingTxsCount() int {
	if tfbs.GetMissingTxsCountCalled != nil {
		return tfbs.GetMissingTxsCountCalled()
	}
	return 0
}

// ReceivedTransaction -
func (tfbs *TxsForBlockStub) ReceivedTransaction(txHash []byte, tx data.TransactionHandler) {
	if tfbs.ReceivedTransactionCalled != nil {
		tfbs.ReceivedTransactionCalled(txHash, tx)
	}
}

// HasMissingTransactions -
func (tfbs *TxsForBlockStub) HasMissingTransactions() bool {
	if tfbs.HasMissingTransactionsCalled != nil {
		return tfbs.HasMissingTransactionsCalled()
	}
	return false
}

// ComputeExistingAndRequestMissing -
func (tfbs *TxsForBlockStub) ComputeExistingAndRequestMissing(body *block.Body, isMiniBlockCorrect func(block.Type) bool, txPool dataRetriever.ShardedDataCacherNotifier, onRequestTxs func(shardID uint32, txHashes [][]byte)) int {
	if tfbs.ComputeExistingAndRequestMissingCalled != nil {
		return tfbs.ComputeExistingAndRequestMissingCalled(body, isMiniBlockCorrect, txPool, onRequestTxs)
	}
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (tfbs *TxsForBlockStub) IsInterfaceNil() bool {
	return tfbs == nil
}
