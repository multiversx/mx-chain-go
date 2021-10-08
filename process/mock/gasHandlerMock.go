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

// Init -
func (ghm *GasHandlerMock) Init() {
	if ghm.InitCalled != nil {
		ghm.InitCalled()
	}
}

// SetGasConsumed -
func (ghm *GasHandlerMock) SetGasConsumed(gasConsumed uint64, hash []byte) {
	if ghm.SetGasConsumedCalled != nil {
		ghm.SetGasConsumedCalled(gasConsumed, hash)
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

// GasConsumed -
func (ghm *GasHandlerMock) GasConsumed(hash []byte) uint64 {
	if ghm.GasConsumedCalled != nil {
		return ghm.GasConsumedCalled(hash)
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

// TotalGasConsumed -
func (ghm *GasHandlerMock) TotalGasConsumed() uint64 {
	if ghm.TotalGasConsumedCalled != nil {
		return ghm.TotalGasConsumedCalled()
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

// RemoveGasConsumed -
func (ghm *GasHandlerMock) RemoveGasConsumed(hashes [][]byte) {
	if ghm.RemoveGasConsumedCalled != nil {
		ghm.RemoveGasConsumedCalled(hashes)
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

// ComputeGasConsumedByMiniBlock -
func (ghm *GasHandlerMock) ComputeGasConsumedByMiniBlock(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
	if ghm.ComputeGasConsumedByMiniBlockCalled != nil {
		return ghm.ComputeGasConsumedByMiniBlockCalled(miniBlock, mapHashTx)
	}
	return 0, 0, nil
}

// ComputeGasConsumedByTx -
func (ghm *GasHandlerMock) ComputeGasConsumedByTx(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
	if ghm.ComputeGasConsumedByTxCalled != nil {
		return ghm.ComputeGasConsumedByTxCalled(txSenderShardId, txReceiverShardId, txHandler)
	}
	return 0, 0, nil
}

// IsInterfaceNil -
func (ghm *GasHandlerMock) IsInterfaceNil() bool {
	return ghm == nil
}
