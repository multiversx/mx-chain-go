package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
)

type BlockProcessorMock struct {
	ProcessBlockCalled               func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled         func()
	CreateGenesisBlockCalled         func(balances map[string]*big.Int) (data.HeaderHandler, error)
	CreateBlockCalled                func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error)
	RestoreBlockIntoPoolsCalled      func(header data.HeaderHandler, body data.BodyHandler) error
	noShards                         uint32
	SetOnRequestTransactionCalled    func(f func(destShardID uint32, txHash []byte))
	ApplyBodyToHeaderCalled          func(header data.HeaderHandler, body data.BodyHandler) error
	MarshalizedDataToBroadcastCalled func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBodyCalled            func(dta []byte) data.BodyHandler
	DecodeBlockHeaderCalled          func(dta []byte) data.HeaderHandler
	AddLastNotarizedHdrCalled        func(shardId uint32, processedHdr data.HeaderHandler)
	CreateNewHeaderCalled            func() data.HeaderHandler
	RevertStateToBlockCalled         func(header data.HeaderHandler) error
}

func (bpm *BlockProcessorMock) ApplyProcessedMiniBlocks(miniBlocks map[string]map[string]struct{}) {

}

func (bpm *BlockProcessorMock) RestoreLastNotarizedHrdsToGenesis() {

}

func (bpm *BlockProcessorMock) ProcessBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	return bpm.ProcessBlockCalled(blockChain, header, body, haveTime)
}

func (bpm *BlockProcessorMock) CommitBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
	return bpm.CommitBlockCalled(blockChain, header, body)
}

func (bpm *BlockProcessorMock) RevertAccountState() {
	bpm.RevertAccountStateCalled()
}

func (bpm *BlockProcessorMock) CreateNewHeader() data.HeaderHandler {
	return bpm.CreateNewHeaderCalled()
}

func (bpm *BlockProcessorMock) CreateGenesisBlock(balances map[string]*big.Int) (data.HeaderHandler, error) {
	return bpm.CreateGenesisBlockCalled(balances)
}

func (bpm *BlockProcessorMock) CreateBlockBody(initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error) {
	return bpm.CreateBlockCalled(initialHdrData, haveTime)
}

func (bpm *BlockProcessorMock) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	return bpm.RestoreBlockIntoPoolsCalled(header, body)
}

func (bpm *BlockProcessorMock) ApplyBodyToHeader(header data.HeaderHandler, body data.BodyHandler) error {
	return bpm.ApplyBodyToHeaderCalled(header, body)
}

func (bpm *BlockProcessorMock) MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
	return bpm.MarshalizedDataToBroadcastCalled(header, body)
}

func (bpm *BlockProcessorMock) DecodeBlockBody(dta []byte) data.BodyHandler {
	return bpm.DecodeBlockBodyCalled(dta)
}

func (bpm *BlockProcessorMock) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	return bpm.DecodeBlockHeaderCalled(dta)
}

func (bpm *BlockProcessorMock) AddLastNotarizedHdr(shardId uint32, processedHdr data.HeaderHandler) {
	bpm.AddLastNotarizedHdrCalled(shardId, processedHdr)
}

func (bpm *BlockProcessorMock) SetConsensusData(randomness []byte, round uint64, epoch uint32, shardId uint32) {
	panic("implement me")
}

// RevertStateToBlock recreates thee state tries to the root hashes indicated by the provided header
func (bpm *BlockProcessorMock) RevertStateToBlock(header data.HeaderHandler) error {
	if bpm.RevertStateToBlockCalled != nil {
		return bpm.RevertStateToBlockCalled(header)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bpm *BlockProcessorMock) IsInterfaceNil() bool {
	if bpm == nil {
		return true
	}
	return false
}
