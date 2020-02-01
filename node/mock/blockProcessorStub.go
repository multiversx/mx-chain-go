package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
)

// BlockProcessorStub mocks the implementation for a blockProcessor
type BlockProcessorStub struct {
	ProcessBlockCalled               func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlockCalled                func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountStateCalled         func()
	CreateGenesisBlockCalled         func(balances map[string]*big.Int) (data.HeaderHandler, error)
	CreateBlockBodyCalled            func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error)
	RestoreBlockIntoPoolsCalled      func(header data.HeaderHandler, body data.BodyHandler) error
	SetOnRequestTransactionCalled    func(f func(destShardID uint32, txHash []byte))
	ApplyBodyToHeaderCalled          func(header data.HeaderHandler, body data.BodyHandler) (data.BodyHandler, error)
	MarshalizedDataToBroadcastCalled func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBodyAndHeaderCalled   func(dta []byte) (data.BodyHandler, data.HeaderHandler)
	DecodeBlockBodyCalled            func(dta []byte) data.BodyHandler
	DecodeBlockHeaderCalled          func(dta []byte) data.HeaderHandler
	AddLastNotarizedHdrCalled        func(shardId uint32, processedHdr data.HeaderHandler)
	CreateNewHeaderCalled            func() data.HeaderHandler
	RevertStateToBlockCalled         func(header data.HeaderHandler) error
}

func (bps *BlockProcessorStub) RestoreLastNotarizedHrdsToGenesis() {
}

func (bps *BlockProcessorStub) SetNumProcessedObj(numObj uint64) {
}

// ProcessBlock mocks pocessing a block
func (bps *BlockProcessorStub) ProcessBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
	return bps.ProcessBlockCalled(blockChain, header, body, haveTime)
}

// CommitBlock mocks the commit of a block
func (bps *BlockProcessorStub) CommitBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
	return bps.CommitBlockCalled(blockChain, header, body)
}

// RevertAccountState mocks revert of the accounts state
func (bps *BlockProcessorStub) RevertAccountState() {
	bps.RevertAccountStateCalled()
}

// CreateGenesisBlock mocks the creation of a genesis block body
func (bps *BlockProcessorStub) CreateGenesisBlock(balances map[string]*big.Int) (data.HeaderHandler, error) {
	return bps.CreateGenesisBlockCalled(balances)
}

// RevertStateToBlock recreates thee state tries to the root hashes indicated by the provided header
func (bps *BlockProcessorStub) RevertStateToBlock(header data.HeaderHandler) error {
	if bps.RevertStateToBlockCalled != nil {
		return bps.RevertStateToBlockCalled(header)
	}
	return nil
}

// CreateBlockBody mocks the creation of a transaction block body
func (bps *BlockProcessorStub) CreateBlockBody(initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error) {
	return bps.CreateBlockBodyCalled(initialHdrData, haveTime)
}

func (bps *BlockProcessorStub) RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error {
	return bps.RestoreBlockIntoPoolsCalled(header, body)
}

func (bps *BlockProcessorStub) ApplyBodyToHeader(header data.HeaderHandler, body data.BodyHandler) (data.BodyHandler, error) {
	return bps.ApplyBodyToHeaderCalled(header, body)
}

func (bps *BlockProcessorStub) MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
	return bps.MarshalizedDataToBroadcastCalled(header, body)
}

func (bps *BlockProcessorStub) DecodeBlockBodyAndHeader(dta []byte) (data.BodyHandler, data.HeaderHandler) {
	return bps.DecodeBlockBodyAndHeaderCalled(dta)
}

func (bps *BlockProcessorStub) DecodeBlockBody(dta []byte) data.BodyHandler {
	return bps.DecodeBlockBodyCalled(dta)
}

func (bps *BlockProcessorStub) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	return bps.DecodeBlockHeaderCalled(dta)
}

func (bps *BlockProcessorStub) AddLastNotarizedHdr(shardId uint32, processedHdr data.HeaderHandler) {
	bps.AddLastNotarizedHdrCalled(shardId, processedHdr)
}

func (bps *BlockProcessorStub) SetConsensusData(randomness []byte, round uint64, epoch uint32, shardId uint32) {
	panic("implement me")
}

// CreateNewHeader creates a new header
func (bps BlockProcessorStub) CreateNewHeader() data.HeaderHandler {
	return bps.CreateNewHeaderCalled()
}

func (bps *BlockProcessorStub) ApplyProcessedMiniBlocks(miniBlocks *processedMb.ProcessedMiniBlockTracker) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (bps *BlockProcessorStub) IsInterfaceNil() bool {
	return bps == nil
}
