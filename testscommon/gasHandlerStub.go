package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

// GasHandlerStub -
type GasHandlerStub struct {
	InitCalled                          func()
	SetGasConsumedCalled                func(gasConsumed uint64, hash []byte)
	SetGasRefundedCalled                func(gasRefunded uint64, hash []byte)
	GasConsumedCalled                   func(hash []byte) uint64
	GasRefundedCalled                   func(hash []byte) uint64
	TotalGasConsumedCalled              func() uint64
	TotalGasRefundedCalled              func() uint64
	RemoveGasConsumedCalled             func(hashes [][]byte)
	RemoveGasRefundedCalled             func(hashes [][]byte)
	ComputeGasConsumedByMiniBlockCalled func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error)
	ComputeGasConsumedByTxCalled        func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error)
}

// Init -
func (ghm *GasHandlerStub) Init() {
	ghm.InitCalled()
}

// SetGasConsumed -
func (ghm *GasHandlerStub) SetGasConsumed(gasConsumed uint64, hash []byte) {
	if ghm.SetGasConsumedCalled != nil {
		ghm.SetGasConsumedCalled(gasConsumed, hash)
	}
}

// SetGasRefunded -
func (ghm *GasHandlerStub) SetGasRefunded(gasRefunded uint64, hash []byte) {
	if ghm.SetGasRefundedCalled != nil {
		ghm.SetGasRefundedCalled(gasRefunded, hash)
	}
}

// GasConsumed -
func (ghm *GasHandlerStub) GasConsumed(hash []byte) uint64 {
	return ghm.GasConsumedCalled(hash)
}

// GasRefunded -
func (ghm *GasHandlerStub) GasRefunded(hash []byte) uint64 {
	if ghm.GasRefundedCalled != nil {
		return ghm.GasRefundedCalled(hash)
	}
	return 0
}

// TotalGasConsumed -
func (ghm *GasHandlerStub) TotalGasConsumed() uint64 {
	if ghm.TotalGasConsumedCalled != nil {
		return ghm.TotalGasConsumedCalled()
	}
	return 0
}

// TotalGasRefunded -
func (ghm *GasHandlerStub) TotalGasRefunded() uint64 {
	return ghm.TotalGasRefundedCalled()
}

// RemoveGasConsumed -
func (ghm *GasHandlerStub) RemoveGasConsumed(hashes [][]byte) {
	ghm.RemoveGasConsumedCalled(hashes)
}

// RemoveGasRefunded -
func (ghm *GasHandlerStub) RemoveGasRefunded(hashes [][]byte) {
	ghm.RemoveGasRefundedCalled(hashes)
}

// ComputeGasConsumedByMiniBlock -
func (ghm *GasHandlerStub) ComputeGasConsumedByMiniBlock(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
	return ghm.ComputeGasConsumedByMiniBlockCalled(miniBlock, mapHashTx)
}

// ComputeGasConsumedByTx -
func (ghm *GasHandlerStub) ComputeGasConsumedByTx(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
	if ghm.ComputeGasConsumedByTxCalled != nil {
		return ghm.ComputeGasConsumedByTxCalled(txSenderShardId, txReceiverShardId, txHandler)
	}
	return 0, 0, nil
}

// IsInterfaceNil -
func (ghm *GasHandlerStub) IsInterfaceNil() bool {
	return ghm == nil
}
