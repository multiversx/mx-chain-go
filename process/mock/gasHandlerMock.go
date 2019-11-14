package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type GasHandlerMock struct {
	InitGasConsumedCalled               func()
	AddGasConsumedCalled                func(gasConsumed uint64)
	SetGasConsumedCalled                func(gasConsumed uint64)
	GasConsumedCalled                   func() uint64
	ComputeGasConsumedByMiniBlockCalled func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error)
	ComputeGasConsumedByTxCalled        func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error)
}

func (ghm *GasHandlerMock) InitGasConsumed() {
	ghm.InitGasConsumedCalled()
}

func (ghm *GasHandlerMock) AddGasConsumed(gasConsumed uint64) {
	ghm.AddGasConsumedCalled(gasConsumed)
}

func (ghm *GasHandlerMock) SetGasConsumed(gasConsumed uint64) {
	ghm.SetGasConsumedCalled(gasConsumed)
}

func (ghm *GasHandlerMock) GasConsumed() uint64 {
	return ghm.GasConsumedCalled()
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
