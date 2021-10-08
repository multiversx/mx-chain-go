package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

// GasHandlerMock -
type GasHandlerMock struct {
	InitCalled                          func()
	SetGasConsumedCalled                func(gasConsumed uint64, hash []byte)
	SetGasRefundedCalled                func(gasRefunded uint64, hash []byte)
	SetGasPenalizedCalled               func(gasPenalized uint64, hash []byte)
	GasConsumedCalled                   func(hash []byte) uint64
	GasRefundedCalled                   func(hash []byte) uint64
	GasPenalizedCalled                  func(hash []byte) uint64
	TotalGasConsumedCalled              func() uint64
	TotalGasRefundedCalled              func() uint64
	TotalGasPenalizedCalled             func() uint64
	RemoveGasConsumedCalled             func(hashes [][]byte)
	RemoveGasRefundedCalled             func(hashes [][]byte)
	RemoveGasPenalizedCalled            func(hashes [][]byte)
	ComputeGasConsumedByMiniBlockCalled func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error)
	ComputeGasConsumedByTxCalled        func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error)
	AddGasConsumedInSelfShardCalled     func(gasConsumed uint64)
	TotalGasConsumedInSelfShardCalled   func() uint64
}

// Init -
func (ghm *GasHandlerMock) Init() {
	ghm.InitCalled()
}

// SetGasConsumed -
func (ghm *GasHandlerMock) SetGasConsumed(gasConsumed uint64, hash []byte) {
	ghm.SetGasConsumedCalled(gasConsumed, hash)
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

// GasConsumed -
func (ghm *GasHandlerMock) GasConsumed(hash []byte) uint64 {
	return ghm.GasConsumedCalled(hash)
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

// TotalGasConsumed -
func (ghm *GasHandlerMock) TotalGasConsumed() uint64 {
	return ghm.TotalGasConsumedCalled()
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

// RemoveGasConsumed -
func (ghm *GasHandlerMock) RemoveGasConsumed(hashes [][]byte) {
	ghm.RemoveGasConsumedCalled(hashes)
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

// ComputeGasConsumedByMiniBlock -
func (ghm *GasHandlerMock) ComputeGasConsumedByMiniBlock(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
	return ghm.ComputeGasConsumedByMiniBlockCalled(miniBlock, mapHashTx)
}

// ComputeGasConsumedByTx -
func (ghm *GasHandlerMock) ComputeGasConsumedByTx(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
	return ghm.ComputeGasConsumedByTxCalled(txSenderShardId, txReceiverShardId, txHandler)
}

// AddTotalGasConsumedInSelfShard -
func (ghm *GasHandlerMock) AddGasConsumedInSelfShard(gasConsumed uint64) {
	if ghm.AddGasConsumedInSelfShardCalled != nil {
		ghm.AddGasConsumedInSelfShardCalled(gasConsumed)
	}
}

// TotalGasConsumedInSelfShard -
func (ghm *GasHandlerMock) TotalGasConsumedInSelfShard() uint64 {
	if ghm.TotalGasConsumedInSelfShardCalled != nil {
		return ghm.TotalGasConsumedInSelfShardCalled()
	}

	return 0
}

// IsInterfaceNil -
func (ghm *GasHandlerMock) IsInterfaceNil() bool {
	return ghm == nil
}
