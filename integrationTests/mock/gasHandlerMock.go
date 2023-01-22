package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// GasHandlerMock -
type GasHandlerMock struct {
	InitCalled                          func()
	SetGasProvidedCalled                func(gasProvided uint64, hash []byte)
	SetGasRefundedCalled                func(gasRefunded uint64, hash []byte)
	SetGasPenalizedCalled               func(gasPenalized uint64, hash []byte)
	GasProvidedCalled                   func(hash []byte) uint64
	GasRefundedCalled                   func(hash []byte) uint64
	GasPenalizedCalled                  func(hash []byte) uint64
	TotalGasProvidedCalled              func() uint64
	TotalGasProvidedWithScheduledCalled func() uint64
	TotalGasRefundedCalled              func() uint64
	TotalGasPenalizedCalled             func() uint64
	RemoveGasProvidedCalled             func(hashes [][]byte)
	RemoveGasRefundedCalled             func(hashes [][]byte)
	RemoveGasPenalizedCalled            func(hashes [][]byte)
	ComputeGasProvidedByMiniBlockCalled func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error)
	ComputeGasProvidedByTxCalled        func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error)
}

// Init -
func (ghm *GasHandlerMock) Init() {
	ghm.InitCalled()
}

// SetGasProvided -
func (ghm *GasHandlerMock) SetGasProvided(gasProvided uint64, hash []byte) {
	ghm.SetGasProvidedCalled(gasProvided, hash)
}

// SetGasRefunded -
func (ghm *GasHandlerMock) SetGasRefunded(gasRefunded uint64, hash []byte) {
	ghm.SetGasRefundedCalled(gasRefunded, hash)
}

// SetGasPenalized -
func (ghm *GasHandlerMock) SetGasPenalized(gasPenalized uint64, hash []byte) {
	if ghm.SetGasPenalizedCalled != nil {
		ghm.SetGasPenalizedCalled(gasPenalized, hash)
	}
}

// GasProvided -
func (ghm *GasHandlerMock) GasProvided(hash []byte) uint64 {
	return ghm.GasProvidedCalled(hash)
}

// GasRefunded -
func (ghm *GasHandlerMock) GasRefunded(hash []byte) uint64 {
	return ghm.GasRefundedCalled(hash)
}

// GasPenalized -
func (ghm *GasHandlerMock) GasPenalized(hash []byte) uint64 {
	if ghm.GasPenalizedCalled != nil {
		return ghm.GasPenalizedCalled(hash)
	}
	return 0
}

// TotalGasProvided -
func (ghm *GasHandlerMock) TotalGasProvided() uint64 {
	return ghm.TotalGasProvidedCalled()
}

// TotalGasProvidedWithScheduled -
func (ghm *GasHandlerMock) TotalGasProvidedWithScheduled() uint64 {
	return ghm.TotalGasProvidedWithScheduledCalled()
}

// TotalGasRefunded -
func (ghm *GasHandlerMock) TotalGasRefunded() uint64 {
	return ghm.TotalGasRefundedCalled()
}

// TotalGasPenalized -
func (ghm *GasHandlerMock) TotalGasPenalized() uint64 {
	if ghm.TotalGasPenalizedCalled != nil {
		return ghm.TotalGasPenalizedCalled()
	}
	return 0
}

// RemoveGasProvided -
func (ghm *GasHandlerMock) RemoveGasProvided(hashes [][]byte) {
	ghm.RemoveGasProvidedCalled(hashes)
}

// RemoveGasRefunded -
func (ghm *GasHandlerMock) RemoveGasRefunded(hashes [][]byte) {
	ghm.RemoveGasRefundedCalled(hashes)
}

// RemoveGasPenalized -
func (ghm *GasHandlerMock) RemoveGasPenalized(hashes [][]byte) {
	if ghm.RemoveGasPenalizedCalled != nil {
		ghm.RemoveGasPenalizedCalled(hashes)
	}
}

// ComputeGasProvidedByMiniBlock -
func (ghm *GasHandlerMock) ComputeGasProvidedByMiniBlock(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
	return ghm.ComputeGasProvidedByMiniBlockCalled(miniBlock, mapHashTx)
}

// ComputeGasProvidedByTx -
func (ghm *GasHandlerMock) ComputeGasProvidedByTx(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
	return ghm.ComputeGasProvidedByTxCalled(txSenderShardId, txReceiverShardId, txHandler)
}

// IsInterfaceNil -
func (ghm *GasHandlerMock) IsInterfaceNil() bool {
	return ghm == nil
}
