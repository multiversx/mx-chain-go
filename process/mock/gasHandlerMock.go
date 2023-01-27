package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// GasHandlerMock -
type GasHandlerMock struct {
	InitCalled                          func()
	ResetCalled                         func(key []byte)
	SetGasProvidedCalled                func(gasProvided uint64, hash []byte)
	SetGasProvidedAsScheduledCalled     func(gasProvided uint64, hash []byte)
	SetGasRefundedCalled                func(gasRefunded uint64, hash []byte)
	SetGasPenalizedCalled               func(gasPenalized uint64, hash []byte)
	GasProvidedCalled                   func(hash []byte) uint64
	GasProvidedAsScheduledCalled        func(hash []byte) uint64
	GasRefundedCalled                   func(hash []byte) uint64
	GasPenalizedCalled                  func(hash []byte) uint64
	TotalGasProvidedCalled              func() uint64
	TotalGasProvidedAsScheduledCalled   func() uint64
	TotalGasProvidedWithScheduledCalled func() uint64
	TotalGasRefundedCalled              func() uint64
	TotalGasPenalizedCalled             func() uint64
	RemoveGasProvidedCalled             func(hashes [][]byte)
	RemoveGasProvidedAsScheduledCalled  func(hashes [][]byte)
	RemoveGasRefundedCalled             func(hashes [][]byte)
	RemoveGasPenalizedCalled            func(hashes [][]byte)
	RestoreGasSinceLastResetCalled      func(key []byte)
	ComputeGasProvidedByMiniBlockCalled func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error)
	ComputeGasProvidedByTxCalled        func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error)
}

// Init -
func (ghm *GasHandlerMock) Init() {
	if ghm.InitCalled != nil {
		ghm.InitCalled()
	}
}

// Reset -
func (ghm *GasHandlerMock) Reset(key []byte) {
	if ghm.ResetCalled != nil {
		ghm.ResetCalled(key)
	}
}

// SetGasProvided -
func (ghm *GasHandlerMock) SetGasProvided(gasProvided uint64, hash []byte) {
	if ghm.SetGasProvidedCalled != nil {
		ghm.SetGasProvidedCalled(gasProvided, hash)
	}
}

// SetGasProvidedAsScheduled -
func (ghm *GasHandlerMock) SetGasProvidedAsScheduled(gasProvided uint64, hash []byte) {
	if ghm.SetGasProvidedAsScheduledCalled != nil {
		ghm.SetGasProvidedAsScheduledCalled(gasProvided, hash)
	}
}

// SetGasRefunded -
func (ghm *GasHandlerMock) SetGasRefunded(gasRefunded uint64, hash []byte) {
	if ghm.SetGasRefundedCalled != nil {
		ghm.SetGasRefundedCalled(gasRefunded, hash)
	}
}

// SetGasPenalized -
func (ghm *GasHandlerMock) SetGasPenalized(gasPenalized uint64, hash []byte) {
	if ghm.SetGasPenalizedCalled != nil {
		ghm.SetGasPenalizedCalled(gasPenalized, hash)
	}
}

// GasProvided -
func (ghm *GasHandlerMock) GasProvided(hash []byte) uint64 {
	if ghm.GasProvidedCalled != nil {
		return ghm.GasProvidedCalled(hash)
	}
	return 0
}

// GasProvidedAsScheduled -
func (ghm *GasHandlerMock) GasProvidedAsScheduled(hash []byte) uint64 {
	if ghm.GasProvidedAsScheduledCalled != nil {
		return ghm.GasProvidedAsScheduledCalled(hash)
	}
	return 0
}

// GasRefunded -
func (ghm *GasHandlerMock) GasRefunded(hash []byte) uint64 {
	if ghm.GasRefundedCalled != nil {
		return ghm.GasRefundedCalled(hash)
	}
	return 0
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
	if ghm.TotalGasProvidedCalled != nil {
		return ghm.TotalGasProvidedCalled()
	}
	return 0
}

// TotalGasProvidedAsScheduled -
func (ghm *GasHandlerMock) TotalGasProvidedAsScheduled() uint64 {
	if ghm.TotalGasProvidedAsScheduledCalled != nil {
		return ghm.TotalGasProvidedAsScheduledCalled()
	}
	return 0
}

// TotalGasProvidedWithScheduled -
func (ghm *GasHandlerMock) TotalGasProvidedWithScheduled() uint64 {
	if ghm.TotalGasProvidedWithScheduledCalled != nil {
		return ghm.TotalGasProvidedWithScheduledCalled()
	}
	return 0
}

// TotalGasRefunded -
func (ghm *GasHandlerMock) TotalGasRefunded() uint64 {
	if ghm.TotalGasRefundedCalled != nil {
		return ghm.TotalGasRefundedCalled()
	}
	return 0
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
	if ghm.RemoveGasProvidedCalled != nil {
		ghm.RemoveGasProvidedCalled(hashes)
	}
}

// RemoveGasProvidedAsScheduled -
func (ghm *GasHandlerMock) RemoveGasProvidedAsScheduled(hashes [][]byte) {
	if ghm.RemoveGasProvidedAsScheduledCalled != nil {
		ghm.RemoveGasProvidedAsScheduledCalled(hashes)
	}
}

// RemoveGasRefunded -
func (ghm *GasHandlerMock) RemoveGasRefunded(hashes [][]byte) {
	if ghm.RemoveGasRefundedCalled != nil {
		ghm.RemoveGasRefundedCalled(hashes)
	}
}

// RemoveGasPenalized -
func (ghm *GasHandlerMock) RemoveGasPenalized(hashes [][]byte) {
	if ghm.RemoveGasPenalizedCalled != nil {
		ghm.RemoveGasPenalizedCalled(hashes)
	}
}

// RestoreGasSinceLastReset -
func (ghm *GasHandlerMock) RestoreGasSinceLastReset(key []byte) {
	if ghm.RestoreGasSinceLastResetCalled != nil {
		ghm.RestoreGasSinceLastResetCalled(key)
	}
}

// ComputeGasProvidedByMiniBlock -
func (ghm *GasHandlerMock) ComputeGasProvidedByMiniBlock(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
	if ghm.ComputeGasProvidedByMiniBlockCalled != nil {
		return ghm.ComputeGasProvidedByMiniBlockCalled(miniBlock, mapHashTx)
	}
	return 0, 0, nil
}

// ComputeGasProvidedByTx -
func (ghm *GasHandlerMock) ComputeGasProvidedByTx(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
	if ghm.ComputeGasProvidedByTxCalled != nil {
		return ghm.ComputeGasProvidedByTxCalled(txSenderShardId, txReceiverShardId, txHandler)
	}
	return 0, 0, nil
}

// IsInterfaceNil -
func (ghm *GasHandlerMock) IsInterfaceNil() bool {
	return ghm == nil
}
