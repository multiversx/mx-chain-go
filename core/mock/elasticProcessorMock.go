package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ElasticProcessorMock -
type ElasticProcessorMock struct {
	SaveShardStatisticsCalled        func(tpsBenchmark statistics.TPSBenchmark) error
	SaveHeaderCalled                 func(header data.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, txsSize int) error
	RemoveHeaderCalled               func(header data.HeaderHandler) error
	RemoveMiniblocksCalled           func(header data.HeaderHandler, body *block.Body) error
	SaveMiniblocksCalled             func(header data.HeaderHandler, body *block.Body) (map[string]bool, error)
	SaveTransactionsCalled           func(body *block.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardID uint32, mbsInDb map[string]bool) error
	SaveValidatorsRatingCalled       func(index string, validatorsRatingInfo []workItems.ValidatorRatingInfo) error
	SaveRoundsInfoCalled             func(infos []workItems.RoundInfo) error
	SaveShardValidatorsPubKeysCalled func(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error
	SetTxLogsProcessorCalled         func(txLogsProc process.TransactionLogProcessorDatabase)
}

// SaveShardStatistics -
func (eim *ElasticProcessorMock) SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark) error {
	if eim.SaveShardStatisticsCalled != nil {
		return eim.SaveShardStatisticsCalled(tpsBenchmark)
	}
	return nil
}

// SaveHeader -
func (eim *ElasticProcessorMock) SaveHeader(header data.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, txsSize int) error {
	if eim.SaveHeaderCalled != nil {
		return eim.SaveHeaderCalled(header, signersIndexes, body, notarizedHeadersHashes, txsSize)
	}
	return nil
}

// RemoveHeader -
func (eim *ElasticProcessorMock) RemoveHeader(header data.HeaderHandler) error {
	if eim.RemoveHeaderCalled != nil {
		return eim.RemoveHeaderCalled(header)
	}
	return nil
}

// RemoveMiniblocks -
func (eim *ElasticProcessorMock) RemoveMiniblocks(header data.HeaderHandler, body *block.Body) error {
	if eim.RemoveMiniblocksCalled != nil {
		return eim.RemoveMiniblocksCalled(header, body)
	}
	return nil
}

// SaveMiniblocks -
func (eim *ElasticProcessorMock) SaveMiniblocks(header data.HeaderHandler, body *block.Body) (map[string]bool, error) {
	if eim.SaveMiniblocksCalled != nil {
		return eim.SaveMiniblocksCalled(header, body)
	}
	return nil, nil
}

// SaveTransactions -
func (eim *ElasticProcessorMock) SaveTransactions(body *block.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardID uint32, mbsInDb map[string]bool) error {
	if eim.SaveTransactionsCalled != nil {
		return eim.SaveTransactionsCalled(body, header, txPool, selfShardID, mbsInDb)
	}
	return nil
}

// SaveValidatorsRating -
func (eim *ElasticProcessorMock) SaveValidatorsRating(index string, validatorsRatingInfo []workItems.ValidatorRatingInfo) error {
	if eim.SaveValidatorsRatingCalled != nil {
		return eim.SaveValidatorsRatingCalled(index, validatorsRatingInfo)
	}
	return nil
}

// SaveRoundsInfo -
func (eim *ElasticProcessorMock) SaveRoundsInfo(info []workItems.RoundInfo) error {
	if eim.SaveRoundsInfoCalled != nil {
		return eim.SaveRoundsInfoCalled(info)
	}
	return nil
}

// SaveShardValidatorsPubKeys -
func (eim *ElasticProcessorMock) SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error {
	if eim.SaveShardValidatorsPubKeysCalled != nil {
		return eim.SaveShardValidatorsPubKeysCalled(shardID, epoch, shardValidatorsPubKeys)
	}
	return nil
}

// SetTxLogsProcessor -
func (eim *ElasticProcessorMock) SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase) {
	if eim.SetTxLogsProcessorCalled != nil {
		eim.SetTxLogsProcessorCalled(txLogsProc)
	}
}
