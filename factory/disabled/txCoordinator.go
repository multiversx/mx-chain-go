package disabled

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// TxCoordinator implements the TransactionCoordinator interface but does nothing as it is disabled
type TxCoordinator struct {
}

// CreatePostProcessMiniBlocks does nothing as it is disabled
func (txCoordinator *TxCoordinator) CreatePostProcessMiniBlocks() block.MiniBlockSlice {
	return make(block.MiniBlockSlice, 0)
}

// CreateReceiptsHash does nothing as it is disabled
func (txCoordinator *TxCoordinator) CreateReceiptsHash() ([]byte, error) {
	return nil, nil
}

// ComputeTransactionType does nothing as it is disabled
func (txCoordinator *TxCoordinator) ComputeTransactionType(_ data.TransactionHandler) (process.TransactionType, process.TransactionType) {
	return 0, 0
}

// RequestMiniBlocks does nothing as it is disabled
func (txCoordinator *TxCoordinator) RequestMiniBlocks(_ data.HeaderHandler) {
}

// RequestBlockTransactions does nothing as it is disabled
func (txCoordinator *TxCoordinator) RequestBlockTransactions(_ *block.Body) {
}

// IsDataPreparedForProcessing does nothing as it is disabled
func (txCoordinator *TxCoordinator) IsDataPreparedForProcessing(_ func() time.Duration) error {
	return nil
}

// SaveTxsToStorage does nothing as it is disabled
func (txCoordinator *TxCoordinator) SaveTxsToStorage(_ *block.Body) error {
	return nil
}

// RestoreBlockDataFromStorage does nothing as it is disabled
func (txCoordinator *TxCoordinator) RestoreBlockDataFromStorage(_ *block.Body) (int, error) {
	return 0, nil
}

// RemoveBlockDataFromPool does nothing as it is disabled
func (txCoordinator *TxCoordinator) RemoveBlockDataFromPool(_ *block.Body) error {
	return nil
}

// RemoveTxsFromPool does nothing as it is disabled
func (txCoordinator *TxCoordinator) RemoveTxsFromPool(_ *block.Body) error {
	return nil
}

// ProcessBlockTransaction does nothing as it is disabled
func (txCoordinator *TxCoordinator) ProcessBlockTransaction(_ *block.Body, _ func() time.Duration) error {
	return nil
}

// CreateBlockStarted does nothing as it is disabled
func (txCoordinator *TxCoordinator) CreateBlockStarted() {
}

// CreateMbsAndProcessCrossShardTransactionsDstMe does nothing as it is disabled
func (txCoordinator *TxCoordinator) CreateMbsAndProcessCrossShardTransactionsDstMe(
	_ data.HeaderHandler,
	_ map[string]struct{},
	_ func() bool,
) (block.MiniBlockSlice, uint32, bool, error) {
	return make(block.MiniBlockSlice, 0), 0, false, nil
}

// CreateMbsAndProcessTransactionsFromMe does nothing as it is disabled
func (txCoordinator *TxCoordinator) CreateMbsAndProcessTransactionsFromMe(_ func() bool) block.MiniBlockSlice {
	return make(block.MiniBlockSlice, 0)
}

// CreateMarshalizedData does nothing as it is disabled
func (txCoordinator *TxCoordinator) CreateMarshalizedData(_ *block.Body) map[string][][]byte {
	return make(map[string][][]byte)
}

// GetAllCurrentUsedTxs does nothing as it is disabled
func (txCoordinator *TxCoordinator) GetAllCurrentUsedTxs(_ block.Type) map[string]data.TransactionHandler {
	return make(map[string]data.TransactionHandler)
}

// VerifyCreatedBlockTransactions does nothing as it is disabled
func (txCoordinator *TxCoordinator) VerifyCreatedBlockTransactions(_ data.HeaderHandler, _ *block.Body) error {
	return nil
}

// CreateMarshalizedReceipts does nothing as it is disabled
func (txCoordinator *TxCoordinator) CreateMarshalizedReceipts() ([]byte, error) {
	return nil, nil
}

// VerifyCreatedMiniBlocks does nothing as it is disabled
func (txCoordinator *TxCoordinator) VerifyCreatedMiniBlocks(_ data.HeaderHandler, _ *block.Body) error {
	return nil
}

// AddIntermediateTransactions does nothing as it is disabled
func (txCoordinator *TxCoordinator) AddIntermediateTransactions(_ map[block.Type][]data.TransactionHandler) error {
	return nil
}

// GetAllIntermediateTxs does nothing as it is disabled
func (txCoordinator *TxCoordinator) GetAllIntermediateTxs() map[block.Type]map[string]data.TransactionHandler {
	return make(map[block.Type]map[string]data.TransactionHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (txCoordinator *TxCoordinator) IsInterfaceNil() bool {
	return txCoordinator == nil
}
