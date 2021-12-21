package esdtSupply

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type logsGetter struct {
	logsStorer  storage.Storer
	marshalizer marshal.Marshalizer
}

func newLogsGetter(
	marshalizer marshal.Marshalizer,
	logsStorer storage.Storer,
) *logsGetter {
	return &logsGetter{
		logsStorer:  logsStorer,
		marshalizer: marshalizer,
	}
}

func (lg *logsGetter) getLogsBasedOnBody(blockBody data.BodyHandler) ([]indexer.LogData, error) {
	body, ok := blockBody.(*block.Body)
	if !ok {
		return nil, errCannotCastToBlockBody
	}

	logsDB := make([]indexer.LogData, 0)
	for _, mb := range body.MiniBlocks {
		shouldIgnore := mb.Type != block.TxBlock && mb.Type != block.SmartContractResultBlock
		if shouldIgnore {
			continue
		}

		dbLogsMb, err := lg.getLogsBasedOnMB(mb)
		if err != nil {
			return nil, err
		}

		logsDB = mergeLogsSlices(logsDB, dbLogsMb)
	}

	return logsDB, nil
}

func (lg *logsGetter) getLogsBasedOnMB(mb *block.MiniBlock) ([]indexer.LogData, error) {
	dbLogs := make([]indexer.LogData, 0)
	for _, txHash := range mb.TxHashes {
		txLog, ok, err := lg.getTxLog(txHash)
		if err != nil {
			return nil, err
		}

		if !ok {
			continue
		}

		dbLogs = append(dbLogs, indexer.LogData{
			LogHandler: txLog,
			TxHash: string(txHash),
		})
	}

	return dbLogs, nil
}

func (lg *logsGetter) getTxLog(txHash []byte) (data.LogHandler, bool, error) {
	logBytes, err := lg.logsStorer.Get(txHash)
	if err != nil {
		return nil, false, nil
	}

	logFromDB := &transaction.Log{}
	err = lg.marshalizer.Unmarshal(logFromDB, logBytes)
	if err != nil {
		log.Warn("logsGetter.getTxLog cannot unmarshal log",
			"error", err,
			"txHash", hex.EncodeToString(txHash),
		)

		return nil, false, err
	}

	return logFromDB, true, nil
}

func mergeLogsSlices(m1, m2 []indexer.LogData) []indexer.LogData {
	logsMap := make(map[string]indexer.LogData, len(m1)+len(m2))

	for _, value := range m1 {
		logsMap[value.TxHash] = value
	}

	for _, value := range m2 {
		logsMap[value.TxHash] = value
	}

	finalSlice := make([]indexer.LogData, 0)
	for _, logData := range logsMap {
		finalSlice = append(finalSlice, logData)
	}

	return finalSlice
}
