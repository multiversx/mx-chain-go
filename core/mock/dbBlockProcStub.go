package mock

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// DbBlockProcHandlerStub -
type DbBlockProcHandlerStub struct {
	PrepareBlockForDBCalled func(
		header data.HeaderHandler,
		signersIndexes []uint64,
		body *block.Body,
		notarizedHeadersHashes []string,
		sizeTxs int,
	) (*types.Block, error)

	SerializeBlockCalled func(elasticBlock *types.Block) (*bytes.Buffer, error)
}

// PrepareBlockForDB -
func (d *DbBlockProcHandlerStub) PrepareBlockForDB(
	header data.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	sizeTxs int,
) (*types.Block, error) {
	if d.PrepareBlockForDBCalled != nil {
		return d.PrepareBlockForDBCalled(header, signersIndexes, body, notarizedHeadersHashes, sizeTxs)
	}

	return &types.Block{}, nil
}

// SerializeBlock -
func (d *DbBlockProcHandlerStub) SerializeBlock(elasticBlock *types.Block) (*bytes.Buffer, error) {
	if d.SerializeBlockCalled != nil {
		return d.SerializeBlockCalled(elasticBlock)
	}

	return &bytes.Buffer{}, nil
}
