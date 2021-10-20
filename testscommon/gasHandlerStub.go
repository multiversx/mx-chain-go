package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

// GasHandlerStub -
type GasHandlerStub struct {
	InitCalled                          func()
	SetGasConsumedCalled                func(gasConsumed uint64, hash []byte)
	SetGasConsumedAsScheduledCalled     func(gasConsumed uint64, hash []byte)
	SetGasRefundedCalled                func(gasRefunded uint64, hash []byte)
	GasConsumedCalled                   func(hash []byte) uint64
	GasRefundedCalled                   func(hash []byte) uint64
	TotalGasConsumedCalled              func() uint64
	TotalGasConsumedAsScheduledCalled   func() uint64
	TotalGasRefundedCalled              func() uint64
	RemoveGasConsumedCalled             func(hashes [][]byte)
	RemoveGasConsumedAsScheduledCalled  func(hashes [][]byte)
	RemoveGasRefundedCalled             func(hashes [][]byte)
	ComputeGasConsumedByMiniBlockCalled func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error)
	ComputeGasConsumedByTxCalled        func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error)
}

// Init -
func (ghs *GasHandlerStub) Init() {
	if ghs.InitCalled != nil {
		ghs.InitCalled()
	}
}

// SetGasConsumed -
func (ghs *GasHandlerStub) SetGasConsumed(gasConsumed uint64, hash []byte) {
	if ghs.SetGasConsumedCalled != nil {
		ghs.SetGasConsumedCalled(gasConsumed, hash)
	}
}

// SetGasConsumedAsScheduled -
func (ghs *GasHandlerStub) SetGasConsumedAsScheduled(gasConsumed uint64, hash []byte) {
	if ghs.SetGasConsumedAsScheduledCalled != nil {
		ghs.SetGasConsumedAsScheduledCalled(gasConsumed, hash)
	}
}

// SetGasRefunded -
func (ghs *GasHandlerStub) SetGasRefunded(gasRefunded uint64, hash []byte) {
	if ghs.SetGasRefundedCalled != nil {
		ghs.SetGasRefundedCalled(gasRefunded, hash)
	}
}

// GasConsumed -
func (ghs *GasHandlerStub) GasConsumed(hash []byte) uint64 {
	if ghs.GasConsumedCalled != nil {
		return ghs.GasConsumedCalled(hash)
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

// TotalGasConsumed -
func (ghs *GasHandlerStub) TotalGasConsumed() uint64 {
	if ghs.TotalGasConsumedCalled != nil {
		return ghs.TotalGasConsumedCalled()
	}
	return 0
}

// TotalGasConsumedAsScheduled -
func (ghs *GasHandlerStub) TotalGasConsumedAsScheduled() uint64 {
	if ghs.TotalGasConsumedAsScheduledCalled != nil {
		return ghs.TotalGasConsumedAsScheduledCalled()
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

// RemoveGasConsumed -
func (ghs *GasHandlerStub) RemoveGasConsumed(hashes [][]byte) {
	if ghs.RemoveGasConsumedCalled != nil {
		ghs.RemoveGasConsumedCalled(hashes)
	}
}

// RemoveGasConsumedAsScheduled -
func (ghs *GasHandlerStub) RemoveGasConsumedAsScheduled(hashes [][]byte) {
	if ghs.RemoveGasConsumedAsScheduledCalled != nil {
		ghs.RemoveGasConsumedAsScheduledCalled(hashes)
	}
}

// RemoveGasRefunded -
func (ghs *GasHandlerStub) RemoveGasRefunded(hashes [][]byte) {
	if ghs.RemoveGasRefundedCalled != nil {
		ghs.RemoveGasRefundedCalled(hashes)
	}
}

// ComputeGasConsumedByMiniBlock -
func (ghs *GasHandlerStub) ComputeGasConsumedByMiniBlock(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
	if ghs.ComputeGasConsumedByMiniBlockCalled != nil {
		return ghs.ComputeGasConsumedByMiniBlockCalled(miniBlock, mapHashTx)
	}
	return 0, 0, nil
}

// ComputeGasConsumedByTx -
func (ghs *GasHandlerStub) ComputeGasConsumedByTx(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
	if ghs.ComputeGasConsumedByTxCalled != nil {
		return ghs.ComputeGasConsumedByTxCalled(txSenderShardId, txReceiverShardId, txHandler)
	}
	return 0, 0, nil
}

// IsInterfaceNil -
func (ghs *GasHandlerStub) IsInterfaceNil() bool {
	return ghs == nil
}
