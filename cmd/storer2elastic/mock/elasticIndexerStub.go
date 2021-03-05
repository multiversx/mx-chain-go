package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ElasticIndexerStub -
type ElasticIndexerStub struct {
	SetTxLogsProcessorCalled    func(txLogsProc process.TransactionLogProcessorDatabase)
	SaveBlockCalled             func(args *indexer.ArgsSaveBlockData)
	SaveRoundsInfosCalled       func(roundsInfos []*indexer.RoundInfo)
	UpdateTPSCalled             func(tpsBenchmark statistics.TPSBenchmark)
	SaveValidatorsPubKeysCalled func(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRatingCalled  func(indexID string, infoRating []*indexer.ValidatorRatingInfo)
}

// SetTxLogsProcessor -
func (e *ElasticIndexerStub) SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase) {
	if e.SetTxLogsProcessorCalled != nil {
		e.SetTxLogsProcessorCalled(txLogsProc)
	}
}

// SaveBlock -
func (e *ElasticIndexerStub) SaveBlock(args *indexer.ArgsSaveBlockData) {
	if e.SaveBlockCalled != nil {
		e.SaveBlockCalled(args)
	}
}

// SaveRoundsInfo -
func (e *ElasticIndexerStub) SaveRoundsInfo(roundsInfos []*indexer.RoundInfo) {
	if e.SaveRoundsInfosCalled != nil {
		e.SaveRoundsInfosCalled(roundsInfos)
	}
}

// UpdateTPS -
func (e *ElasticIndexerStub) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	if e.UpdateTPSCalled != nil {
		e.UpdateTPSCalled(tpsBenchmark)
	}
}

// SaveValidatorsPubKeys -
func (e *ElasticIndexerStub) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) {
	if e.SaveValidatorsPubKeysCalled != nil {
		e.SaveValidatorsPubKeysCalled(validatorsPubKeys, epoch)
	}
}

// SaveValidatorsRating -
func (e *ElasticIndexerStub) SaveValidatorsRating(indexID string, infoRating []*indexer.ValidatorRatingInfo) {
	if e.SaveValidatorsRatingCalled != nil {
		e.SaveValidatorsRatingCalled(indexID, infoRating)
	}
}

// Close -
func (e *ElasticIndexerStub) Close() error {
	return nil
}

// RevertIndexedBlock -
func (e *ElasticIndexerStub) RevertIndexedBlock(_ data.HeaderHandler, _ data.BodyHandler) {
}

// SaveAccounts -
func (e *ElasticIndexerStub) SaveAccounts(_ uint64, _ []state.UserAccountHandler) {
}

// IsInterfaceNil -
func (e *ElasticIndexerStub) IsInterfaceNil() bool {
	return e == nil
}

// IsNilIndexer -
func (e *ElasticIndexerStub) IsNilIndexer() bool {
	return false
}
