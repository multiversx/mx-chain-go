package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/process"
)

// StorageDataIndexerStub -
type StorageDataIndexerStub struct {
	SetTxLogsProcessorCalled    func(txLogsProc process.TransactionLogProcessorDatabase)
	SaveBlockCalled             func(args *indexer.ArgsSaveBlockData)
	SaveRoundsInfosCalled       func(roundsInfos []*indexer.RoundInfo)
	UpdateTPSCalled             func(tpsBenchmark statistics.TPSBenchmark)
	SaveValidatorsPubKeysCalled func(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRatingCalled  func(indexID string, infoRating []*indexer.ValidatorRatingInfo)
}

// SaveBlock -
func (e *StorageDataIndexerStub) SaveBlock(args *indexer.ArgsSaveBlockData) {
	if e.SaveBlockCalled != nil {
		e.SaveBlockCalled(args)
	}
}

// SaveRoundsInfo -
func (e *StorageDataIndexerStub) SaveRoundsInfo(roundsInfos []*indexer.RoundInfo) {
	if e.SaveRoundsInfosCalled != nil {
		e.SaveRoundsInfosCalled(roundsInfos)
	}
}

// UpdateTPS -
func (e *StorageDataIndexerStub) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	if e.UpdateTPSCalled != nil {
		e.UpdateTPSCalled(tpsBenchmark)
	}
}

// SaveValidatorsPubKeys -
func (e *StorageDataIndexerStub) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) {
	if e.SaveValidatorsPubKeysCalled != nil {
		e.SaveValidatorsPubKeysCalled(validatorsPubKeys, epoch)
	}
}

// SaveValidatorsRating -
func (e *StorageDataIndexerStub) SaveValidatorsRating(indexID string, infoRating []*indexer.ValidatorRatingInfo) {
	if e.SaveValidatorsRatingCalled != nil {
		e.SaveValidatorsRatingCalled(indexID, infoRating)
	}
}

// IsInterfaceNil -
func (e *StorageDataIndexerStub) IsInterfaceNil() bool {
	return e == nil
}
