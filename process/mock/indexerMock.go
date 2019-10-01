package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
)

// IndexerMock is a mock implementation fot the Indexer interface
type IndexerMock struct {
	SaveBlockCalled func(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler)
}

func (im *IndexerMock) SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler, signersIndexes []uint64) {
	if im.SaveBlockCalled != nil {
		im.SaveBlockCalled(body, header, txPool)
	}
}

func (im *IndexerMock) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	panic("implement me")
}

func (im *IndexerMock) SaveRoundInfo(round int64, shardId uint32, signersIndexes []uint64, blockWasProposed bool) {
	return
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

func (im *IndexerMock) IsNilIndexer() bool {
	return true
}
