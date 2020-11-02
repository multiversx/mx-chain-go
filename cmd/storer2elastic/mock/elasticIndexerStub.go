package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ElasticIndexerStub -
type ElasticIndexerStub struct {
	SetTxLogsProcessorCalled func(txLogsProc process.TransactionLogProcessorDatabase)
	SaveBlockCalled          func(body data.BodyHandler, header data.HeaderHandler,
		txPool map[string]data.TransactionHandler, signersIndexes []uint64, notarizedHeadersHashes []string, headerHash []byte)
	SaveRoundsInfosCalled       func(roundsInfos []workItems.RoundInfo)
	UpdateTPSCalled             func(tpsBenchmark statistics.TPSBenchmark)
	SaveValidatorsPubKeysCalled func(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRatingCalled  func(indexID string, infoRating []workItems.ValidatorRatingInfo)
}

// SetTxLogsProcessor -
func (e *ElasticIndexerStub) SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase) {
	if e.SetTxLogsProcessorCalled != nil {
		e.SetTxLogsProcessorCalled(txLogsProc)
	}
}

// SaveBlock -
func (e *ElasticIndexerStub) SaveBlock(body data.BodyHandler, header data.HeaderHandler,
	txPool map[string]data.TransactionHandler, signersIndexes []uint64, notarizedHeadersHashes []string, headerHash []byte) {
	if e.SaveBlockCalled != nil {
		e.SaveBlockCalled(body, header, txPool, signersIndexes, notarizedHeadersHashes, headerHash)
	}
}

// SaveRoundsInfo -
func (e *ElasticIndexerStub) SaveRoundsInfo(roundsInfos []workItems.RoundInfo) {
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
func (e *ElasticIndexerStub) SaveValidatorsRating(indexID string, infoRating []workItems.ValidatorRatingInfo) {
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
func (e *ElasticIndexerStub) SaveAccounts(_ []state.UserAccountHandler) {
}

// IsInterfaceNil -
func (e *ElasticIndexerStub) IsInterfaceNil() bool {
	return e == nil
}

// IsNilIndexer -
func (e *ElasticIndexerStub) IsNilIndexer() bool {
	return false
}
