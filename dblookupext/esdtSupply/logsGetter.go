package esdtSupply

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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

func (lg *logsGetter) getLogsBasedOnBody(blockBody data.BodyHandler) (map[string]*vmcommon.LogEntry, error) {
	body, ok := blockBody.(*block.Body)
	if !ok {
		return nil, errCannotCastToBlockBody
	}

	logsDB := make(map[string]*vmcommon.LogEntry)
	for _, mb := range body.MiniBlocks {
		shouldIgnore := mb.Type != block.TxBlock && mb.Type != block.SmartContractResultBlock
		if shouldIgnore {
			continue
		}

		dbLogsMb := lg.getLogsBasedOnMB(mb)

		logsDB = mergeLogsMap(logsDB, dbLogsMb)
	}

	return logsDB, nil
}

func (lg *logsGetter) getLogsBasedOnMB(mb *block.MiniBlock) map[string]*vmcommon.LogEntry {
	dbLogs := make(map[string]*vmcommon.LogEntry)
	for _, txHash := range mb.TxHashes {
		txLog, ok := lg.getTxLog(txHash)
		if !ok {
			continue
		}

		dbLogs[string(txHash)] = txLog
	}

	return dbLogs
}

func (lg *logsGetter) getTxLog(txHash []byte) (*vmcommon.LogEntry, bool) {
	logBytes, err := lg.logsStorer.Get(txHash)
	if err != nil {
		return nil, false
	}

	logFromDB := &vmcommon.LogEntry{}
	err = lg.marshalizer.Unmarshal(logFromDB, logBytes)
	if err != nil {
		log.Warn("logsGetter.getTxLog cannot unmarshal log",
			"error", err,
			"txHash", hex.EncodeToString(txHash),
		)

		return nil, false
	}

	return logFromDB, true
}

func mergeLogsMap(m1, m2 map[string]*vmcommon.LogEntry) map[string]*vmcommon.LogEntry {
	finalMap := make(map[string]*vmcommon.LogEntry, len(m1)+len(m2))

	for key, value := range m1 {
		finalMap[key] = value
	}

	for key, value := range m2 {
		finalMap[key] = value
	}

	return finalMap
}
