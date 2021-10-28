package disabled

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// PostProcessorTxsHandler implements PostProcessorTxsHandler interface but does nothing as it is a disabled component
type PostProcessorTxsHandler struct {
}

// Init does nothing as it is a disabled component
func (ppth *PostProcessorTxsHandler) Init() {
}

// AddPostProcessorTx does nothing as it is a disabled component
func (ppth *PostProcessorTxsHandler) AddPostProcessorTx(_ []byte, _ data.TransactionHandler) bool {
	return true
}

// GetPostProcessorTx does nothing as it is a disabled component
func (ppth *PostProcessorTxsHandler) GetPostProcessorTx(_ []byte) data.TransactionHandler {
	return nil
}

// IsPostProcessorTxAdded does nothing as it is a disabled component
func (ppth *PostProcessorTxsHandler) IsPostProcessorTxAdded(_ []byte) bool {
	return false
}

// SetTransactionCoordinator does nothing as it is a disabled component
func (ppth *PostProcessorTxsHandler) SetTransactionCoordinator(_ process.TransactionCoordinator) {
}

// GetProcessedResults does nothing as it is a disabled component
func (ppth *PostProcessorTxsHandler) GetProcessedResults() map[block.Type]map[uint32][]*process.TxInfo {
	return nil
}

// InitProcessedResults does nothing as it is a disabled component
func (ppth *PostProcessorTxsHandler) InitProcessedResults() {
}

// IsInterfaceNil returns true if underlying object is nil
func (ppth *PostProcessorTxsHandler) IsInterfaceNil() bool {
	return ppth == nil
}
