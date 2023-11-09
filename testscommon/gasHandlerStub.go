package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// GasHandlerStub -
type GasHandlerStub struct {
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
func (ghs *GasHandlerStub) Init() {
	if ghs.InitCalled != nil {
		ghs.InitCalled()
	}
}

// Reset -
func (ghs *GasHandlerStub) Reset(key []byte) {
	if ghs.ResetCalled != nil {
		ghs.ResetCalled(key)
	}
}

// SetGasProvided -
func (ghs *GasHandlerStub) SetGasProvided(gasProvided uint64, hash []byte) {
	if ghs.SetGasProvidedCalled != nil {
		ghs.SetGasProvidedCalled(gasProvided, hash)
	}
}

// SetGasProvidedAsScheduled -
func (ghs *GasHandlerStub) SetGasProvidedAsScheduled(gasProvided uint64, hash []byte) {
	if ghs.SetGasProvidedAsScheduledCalled != nil {
		ghs.SetGasProvidedAsScheduledCalled(gasProvided, hash)
	}
}

// SetGasRefunded -
func (ghs *GasHandlerStub) SetGasRefunded(gasRefunded uint64, hash []byte) {
	if ghs.SetGasRefundedCalled != nil {
		ghs.SetGasRefundedCalled(gasRefunded, hash)
	}
}

// SetGasPenalized -
func (ghs *GasHandlerStub) SetGasPenalized(gasPenalized uint64, hash []byte) {
	if ghs.SetGasPenalizedCalled != nil {
		ghs.SetGasPenalizedCalled(gasPenalized, hash)
	}
}

// GasProvided -
func (ghs *GasHandlerStub) GasProvided(hash []byte) uint64 {
	if ghs.GasProvidedCalled != nil {
		return ghs.GasProvidedCalled(hash)
	}
	return 0
}

// GasProvidedAsScheduled -
func (ghs *GasHandlerStub) GasProvidedAsScheduled(hash []byte) uint64 {
	if ghs.GasProvidedAsScheduledCalled != nil {
		return ghs.GasProvidedAsScheduledCalled(hash)
	}
	return 0
}

// GasRefunded -
func (ghs *GasHandlerStub) GasRefunded(hash []byte) uint64 {
	if ghs.GasRefundedCalled != nil {
		return ghs.GasRefundedCalled(hash)
	}
	return 0
}

// GasPenalized -
func (ghs *GasHandlerStub) GasPenalized(hash []byte) uint64 {
	if ghs.GasPenalizedCalled != nil {
		return ghs.GasPenalizedCalled(hash)
	}
	return 0
}

// TotalGasProvided -
func (ghs *GasHandlerStub) TotalGasProvided() uint64 {
	if ghs.TotalGasProvidedCalled != nil {
		return ghs.TotalGasProvidedCalled()
	}
	return 0
}

// TotalGasProvidedAsScheduled -
func (ghs *GasHandlerStub) TotalGasProvidedAsScheduled() uint64 {
	if ghs.TotalGasProvidedAsScheduledCalled != nil {
		return ghs.TotalGasProvidedAsScheduledCalled()
	}
	return 0
}

// TotalGasProvidedWithScheduled -
func (ghs *GasHandlerStub) TotalGasProvidedWithScheduled() uint64 {
	if ghs.TotalGasProvidedWithScheduledCalled != nil {
		return ghs.TotalGasProvidedWithScheduledCalled()
	}
	return 0
}

// TotalGasRefunded -
func (ghs *GasHandlerStub) TotalGasRefunded() uint64 {
	if ghs.TotalGasRefundedCalled != nil {
		return ghs.TotalGasRefundedCalled()
	}
	return 0
}

// TotalGasPenalized -
func (ghs *GasHandlerStub) TotalGasPenalized() uint64 {
	if ghs.TotalGasPenalizedCalled != nil {
		return ghs.TotalGasPenalizedCalled()
	}
	return 0
}

// RemoveGasProvided -
func (ghs *GasHandlerStub) RemoveGasProvided(hashes [][]byte) {
	if ghs.RemoveGasProvidedCalled != nil {
		ghs.RemoveGasProvidedCalled(hashes)
	}
}

// RemoveGasProvidedAsScheduled -
func (ghs *GasHandlerStub) RemoveGasProvidedAsScheduled(hashes [][]byte) {
	if ghs.RemoveGasProvidedAsScheduledCalled != nil {
		ghs.RemoveGasProvidedAsScheduledCalled(hashes)
	}
}

// RemoveGasRefunded -
func (ghs *GasHandlerStub) RemoveGasRefunded(hashes [][]byte) {
	if ghs.RemoveGasRefundedCalled != nil {
		ghs.RemoveGasRefundedCalled(hashes)
	}
}

// RemoveGasPenalized -
func (ghs *GasHandlerStub) RemoveGasPenalized(hashes [][]byte) {
	if ghs.RemoveGasPenalizedCalled != nil {
		ghs.RemoveGasPenalizedCalled(hashes)
	}
}

// RestoreGasSinceLastReset -
func (ghs *GasHandlerStub) RestoreGasSinceLastReset(key []byte) {
	if ghs.RestoreGasSinceLastResetCalled != nil {
		ghs.RestoreGasSinceLastResetCalled(key)
	}
}

// ComputeGasProvidedByMiniBlock -
func (ghs *GasHandlerStub) ComputeGasProvidedByMiniBlock(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
	if ghs.ComputeGasProvidedByMiniBlockCalled != nil {
		return ghs.ComputeGasProvidedByMiniBlockCalled(miniBlock, mapHashTx)
	}
	return 0, 0, nil
}

// ComputeGasProvidedByTx -
func (ghs *GasHandlerStub) ComputeGasProvidedByTx(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
	if ghs.ComputeGasProvidedByTxCalled != nil {
		return ghs.ComputeGasProvidedByTxCalled(txSenderShardId, txReceiverShardId, txHandler)
	}
	return 0, 0, nil
}

// IsInterfaceNil -
func (ghs *GasHandlerStub) IsInterfaceNil() bool {
	return ghs == nil
}
