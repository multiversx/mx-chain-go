package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type GasHandlerMock struct {
	InitGasConsumedCalled               func()
	AddGasConsumedCalled                func(gasConsumed uint64)
	SetGasConsumedCalled                func(gasConsumed uint64)
	GetGasConsumedCalled                func() uint64
	ComputeGasConsumedByMiniBlockCalled func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error)
	ComputeGasConsumedInShardCalled     func(shId uint32, sndShId uint32, rcvShId uint32, gasConsumedInSndSh uint64, gasConsumedInRcvSh uint64) (uint64, error)
	ComputeGasConsumedByTxCalled        func(txSndShId uint32, txRcvShId uint32, txHandler data.TransactionHandler) (uint64, uint64, error)
	IsMaxGasLimitReachedCalled          func(gasConsumedByTxInSndSh uint64, gasConsumedByTxInRcvSh uint64, gasConsumedByTxInSlfSh uint64, currentGasConsumedByMiniBlockInSndSh uint64, currentGasConsumedByMiniBlockInRcvSh uint64, currentGasConsumedByBlockInSlfSh uint64) bool
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

func (ghm *GasHandlerMock) GetGasConsumed() uint64 {
	return ghm.GetGasConsumedCalled()
}

func (ghm *GasHandlerMock) IsInterfaceNil() bool {
	if ghm == nil {
		return true
	}

	return false
}

func (ghm *GasHandlerMock) ComputeGasConsumedByMiniBlock(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
	return ghm.ComputeGasConsumedByMiniBlockCalled(miniBlock, mapHashTx)
}

func (ghm *GasHandlerMock) ComputeGasConsumedInShard(shId uint32, sndShId uint32, rcvShId uint32, gasConsumedInSndSh uint64, gasConsumedInRcvSh uint64) (uint64, error) {
	return ghm.ComputeGasConsumedInShardCalled(shId, sndShId, rcvShId, gasConsumedInSndSh, gasConsumedInRcvSh)
}

func (ghm *GasHandlerMock) ComputeGasConsumedByTx(txSndShId uint32, txRcvShId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
	return ghm.ComputeGasConsumedByTxCalled(txSndShId, txRcvShId, txHandler)
}

func (ghm *GasHandlerMock) IsMaxGasLimitReached(gasConsumedByTxInSndSh uint64, gasConsumedByTxInRcvSh uint64, gasConsumedByTxInSlfSh uint64, currentGasConsumedByMiniBlockInSndSh uint64, currentGasConsumedByMiniBlockInRcvSh uint64, currentGasConsumedByBlockInSlfSh uint64) bool {
	return ghm.IsMaxGasLimitReachedCalled(gasConsumedByTxInSndSh, gasConsumedByTxInRcvSh, gasConsumedByTxInSlfSh, currentGasConsumedByMiniBlockInSndSh, currentGasConsumedByMiniBlockInRcvSh, currentGasConsumedByBlockInSlfSh)
}
