package esdtSupply

import (
	"encoding/hex"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage"
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

func (lg *logsGetter) getLogsBasedOnBody(blockBody data.BodyHandler) (map[string]*data.LogData, error) {
	body, ok := blockBody.(*block.Body)
	if !ok {
		return nil, errCannotCastToBlockBody
	}

	logsDB := make(map[string]*data.LogData)
	for _, mb := range body.MiniBlocks {
		shouldIgnore := mb.Type != block.TxBlock && mb.Type != block.SmartContractResultBlock
		if shouldIgnore {
			continue
		}

		dbLogsMb, err := lg.getLogsBasedOnMB(mb)
		if err != nil {
			return nil, err
		}

		for _, logData := range dbLogsMb {
			logsDB[logData.TxHash] = logData
		}
	}

	return logsDB, nil
}

func (lg *logsGetter) getLogsBasedOnMB(mb *block.MiniBlock) ([]*data.LogData, error) {
	dbLogs := make([]*data.LogData, 0)
	for _, txHash := range mb.TxHashes {
		txLog, ok, err := lg.getTxLog(txHash)
		if err != nil {
			return nil, err
		}

		if !ok {
			continue
		}

		dbLogs = append(dbLogs, &data.LogData{
			LogHandler: txLog,
			TxHash:     string(txHash),
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
