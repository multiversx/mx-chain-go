package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// IndexerMock is a mock implementation fot the Indexer interface
type IndexerMock struct {
	SaveBlockCalled func(body block.Body, header *block.Header)
}

func (im *IndexerMock) SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler) {
	panic("implement me")
}

func (im *IndexerMock) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	panic("implement me")
}

func (im *IndexerMock) SaveRoundInfo(round int64, shardId uint32, signersIndexes []uint64) {
	panic("implement me")
}

func (im *IndexerMock) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte) {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (im *IndexerMock) IsInterfaceNil() bool {
	if im == nil {
		return true
	}
	return false
}
