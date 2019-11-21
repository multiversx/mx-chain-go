package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type GasHandlerMock struct {
	InitCalled                          func()
	SetGasConsumedCalled                func(gasConsumed uint64, hash []byte)
	SetGasRefundedCalled                func(gasRefunded uint64, hash []byte)
	GasConsumedCalled                   func(hash []byte) uint64
	GasRefundedCalled                   func(hash []byte) uint64
	TotalGasConsumedCalled              func() uint64
	TotalGasRefundedCalled              func() uint64
	RemoveConsumedCalled                func(hashes [][]byte)
	RemoveRefundedCalled                func(hashes [][]byte)
	ComputeGasConsumedByMiniBlockCalled func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error)
	ComputeGasConsumedByTxCalled        func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error)
}

func (ghm *GasHandlerMock) Init() {
	ghm.InitCalled()
}

func (ghm *GasHandlerMock) SetGasConsumed(gasConsumed uint64, hash []byte) {
	ghm.SetGasConsumedCalled(gasConsumed, hash)
}

func (ghm *GasHandlerMock) SetGasRefunded(gasRefunded uint64, hash []byte) {
	ghm.SetGasRefundedCalled(gasRefunded, hash)
}

func (ghm *GasHandlerMock) GasConsumed(hash []byte) uint64 {
	return ghm.GasConsumedCalled(hash)
}

func (ghm *GasHandlerMock) GasRefunded(hash []byte) uint64 {
	return ghm.GasRefundedCalled(hash)
}

func (ghm *GasHandlerMock) TotalGasConsumed() uint64 {
	return ghm.TotalGasConsumedCalled()
}

func (ghm *GasHandlerMock) TotalGasRefunded() uint64 {
	return ghm.TotalGasRefundedCalled()
}

func (ghm *GasHandlerMock) RemoveGasConsumed(hashes [][]byte) {
	ghm.RemoveConsumedCalled(hashes)
}

func (ghm *GasHandlerMock) RemoveGasRefunded(hashes [][]byte) {
	ghm.RemoveRefundedCalled(hashes)
}

func (ghm *GasHandlerMock) ComputeGasConsumedByMiniBlock(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
	return ghm.ComputeGasConsumedByMiniBlockCalled(miniBlock, mapHashTx)
}

func (ghm *GasHandlerMock) ComputeGasConsumedByTx(txSenderShardId uint32, txReceiverShardId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
	return ghm.ComputeGasConsumedByTxCalled(txSenderShardId, txReceiverShardId, txHandler)
}

func (ghm *GasHandlerMock) IsInterfaceNil() bool {
	return ghm == nil
}
