package esdtSupply

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
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

func (lg *logsGetter) getLogsBasedOnBody(blockBody data.BodyHandler) (map[string]data.LogHandler, error) {
	body, ok := blockBody.(*block.Body)
	if !ok {
		return nil, errCannotCastToBlockBody
	}

	logsDB := make(map[string]data.LogHandler)
	for _, mb := range body.MiniBlocks {
		shouldIgnore := mb.Type != block.TxBlock && mb.Type != block.SmartContractResultBlock
		if shouldIgnore {
			continue
		}

		dbLogsMb, err := lg.getLogsBasedOnMB(mb)
		if err != nil {
			return nil, err
		}

		logsDB = mergeLogsMap(logsDB, dbLogsMb)
	}

	return logsDB, nil
}

func (lg *logsGetter) getLogsBasedOnMB(mb *block.MiniBlock) (map[string]data.LogHandler, error) {
	dbLogs := make(map[string]data.LogHandler)
	for _, txHash := range mb.TxHashes {
		txLog, ok, err := lg.getTxLog(txHash)
		if err != nil {
			return nil, err
		}

		if !ok {
			continue
		}

		dbLogs[string(txHash)] = txLog
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

func mergeLogsMap(m1, m2 map[string]data.LogHandler) map[string]data.LogHandler {
	finalMap := make(map[string]data.LogHandler, len(m1)+len(m2))

	for key, value := range m1 {
		finalMap[key] = value
	}

	for key, value := range m2 {
		finalMap[key] = value
	}

	return finalMap
}
