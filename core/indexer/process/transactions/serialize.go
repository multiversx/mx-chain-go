package transactions

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
)

// SerializeScResults -
func (tdp *txDatabaseProcessor) SerializeScResults(scResults []*types.ScResult, bulkSizeThreshold int) ([]bytes.Buffer, error) {
	var err error
	var buff bytes.Buffer

	buffSlice := make([]bytes.Buffer, 0)
	for _, sc := range scResults {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, sc.Hash, "\n"))
		serializedData, errPrepareSc := json.Marshal(sc)
		if errPrepareSc != nil {
			log.Warn("indexer: marshal",
				"error", "could not serialize sc results, will skip indexing",
				"hash", sc.Hash)
			continue
		}

		// append a newline for each element
		serializedData = append(serializedData, "\n"...)

		buffLenWithCurrentScResults := buff.Len() + len(meta) + len(serializedData)
		if buffLenWithCurrentScResults > bulkSizeThreshold && buff.Len() != 0 {
			buffSlice = append(buffSlice, buff)
			buff = bytes.Buffer{}
		}

		buff.Grow(len(meta) + len(serializedData))
		_, err = buff.Write(meta)
		if err != nil {
			log.Warn("elastic search: serialize bulk smart contract results, write meta", "error", err.Error())
			return nil, err
		}
		_, err = buff.Write(serializedData)
		if err != nil {
			log.Warn("elastic search: serialize bulk smart contract results, write serialized sc results", "error", err.Error())
			return nil, err
		}

	}

	// check if the last buffer contains data
	if buff.Len() != 0 {
		buffSlice = append(buffSlice, buff)
	}

	return buffSlice, nil
}

// SerializeReceipts -
func (tdp *txDatabaseProcessor) SerializeReceipts(receipts []*types.Receipt, bulkSizeThreshold int) ([]bytes.Buffer, error) {
	var err error
	var buff bytes.Buffer

	buffSlice := make([]bytes.Buffer, 0)
	for _, rec := range receipts {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, rec.Hash, "\n"))
		serializedData, errPrepareSc := json.Marshal(rec)
		if errPrepareSc != nil {
			log.Warn("indexer: marshal",
				"error", "could not serialize receipts, will skip indexing",
				"hash", rec.Hash)
			continue
		}

		// append a newline for each element
		serializedData = append(serializedData, "\n"...)

		buffLenWithCurrentReceipt := buff.Len() + len(meta) + len(serializedData)
		if buffLenWithCurrentReceipt > bulkSizeThreshold && buff.Len() != 0 {
			buffSlice = append(buffSlice, buff)
			buff = bytes.Buffer{}
		}

		buff.Grow(len(meta) + len(serializedData))
		_, err = buff.Write(meta)
		if err != nil {
			log.Warn("elastic search: serialize bulk receipts, write meta", "error", err.Error())
			return nil, err
		}
		_, err = buff.Write(serializedData)
		if err != nil {
			log.Warn("elastic search: serialize bulk receipts, write serialized receipt", "error", err.Error())
			return nil, err
		}

	}

	// check if the last buffer contains data
	if buff.Len() != 0 {
		buffSlice = append(buffSlice, buff)
	}

	return buffSlice, nil
}

// SerializeTransactions -
func (tdp *txDatabaseProcessor) SerializeTransactions(
	transactions []*types.Transaction,
	selfShardID uint32,
	mbsHashInDB map[string]bool,
	bulkSizeThreshold int,
) ([]bytes.Buffer, error) {
	var err error

	var buff bytes.Buffer
	buffSlice := make([]bytes.Buffer, 0)
	for _, tx := range transactions {
		isMBOfTxInDB := mbsHashInDB[tx.MBHash]
		meta, serializedData, errPrepareTx := prepareSerializedDataForATransaction(tx, selfShardID, isMBOfTxInDB)
		if errPrepareTx != nil {
			log.Warn("error preparing transaction for indexing", "tx hash", tx.Hash, "error", err)
			return nil, errPrepareTx
		}

		// append a newline for each element
		serializedData = append(serializedData, "\n"...)

		buffLenWithCurrentTx := buff.Len() + len(meta) + len(serializedData)
		if buffLenWithCurrentTx > bulkSizeThreshold && buff.Len() != 0 {
			buffSlice = append(buffSlice, buff)
			buff = bytes.Buffer{}
		}

		buff.Grow(len(meta) + len(serializedData))
		_, err = buff.Write(meta)
		if err != nil {
			log.Warn("elastic search: serialize bulk tx, write meta", "error", err.Error())
			return nil, err

		}
		_, err = buff.Write(serializedData)
		if err != nil {
			log.Warn("elastic search: serialize bulk tx, write serialized tx", "error", err.Error())
			return nil, err
		}
	}

	// check if the last buffer contains data
	if buff.Len() != 0 {
		buffSlice = append(buffSlice, buff)
	}

	return buffSlice, nil
}

func prepareSerializedDataForATransaction(
	tx *types.Transaction,
	selfShardID uint32,
	_ bool,
) ([]byte, []byte, error) {
	metaData := []byte(fmt.Sprintf(`{"update":{"_id":"%s", "_type": "_doc"}}%s`, tx.Hash, "\n"))

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize transaction, will skip indexing",
			"tx hash", tx.Hash)
		return nil, nil, err
	}

	if isIntraShardOrInvalid(tx, selfShardID) {
		// if transaction is intra-shard, use basic insert as data can be re-written at forks
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, tx.Hash, "_doc", "\n"))
		log.Trace("indexer tx is intra shard or invalid tx", "meta", string(meta), "marshaledTx", string(marshaledTx))

		return meta, marshaledTx, nil
	}

	if !isCrossShardDstMe(tx, selfShardID) {
		// if transaction is cross-shard and current shard ID is source, use upsert without updating anything
		serializedData :=
			[]byte(fmt.Sprintf(`{"script":{"source":"return"},"upsert":%s}`,
				string(marshaledTx)))
		log.Trace("indexer tx is on sender shard", "metaData", string(metaData), "serializedData", string(serializedData))

		return metaData, serializedData, nil
	}

	serializedData, err := prepareCrossShardTxForDestinationSerialized(tx, marshaledTx)
	if err != nil {
		return nil, nil, err
	}

	log.Trace("indexer tx is on destination shard", "metaData", string(metaData), "serializedData", string(serializedData))

	return metaData, serializedData, nil
}

func prepareCrossShardTxForDestinationSerialized(tx *types.Transaction, marshaledTx []byte) ([]byte, error) {
	// if transaction is cross-shard and current shard ID is destination, use upsert with updating fields
	marshaledLogs, err := json.Marshal(tx.Logs)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize transaction log, will skip indexing",
			"tx hash", tx.Hash)
		return nil, err
	}

	marshaledTimestamp, err := json.Marshal(tx.Timestamp)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize timestamp, will skip indexing",
			"tx hash", tx.Hash)
		return nil, err
	}

	serializedData := []byte(fmt.Sprintf(`{"script":{"source":"`+
		`ctx._source.status = params.status;`+
		`ctx._source.miniBlockHash = params.miniBlockHash;`+
		`ctx._source.log = params.log;`+
		`ctx._source.timestamp = params.timestamp;`+
		`ctx._source.gasUsed = params.gasUsed;`+
		`ctx._source.fee = params.fee;`+
		`ctx._source.hasScResults = params.hasScResults;`+
		`","lang": "painless","params":`+
		`{"status": "%s", "miniBlockHash": "%s", "logs": %s, "timestamp": %s, "gasUsed": %d, "fee": "%s", "hasScResults": %t}},"upsert":%s}`,
		tx.Status, tx.MBHash, string(marshaledLogs), string(marshaledTimestamp), tx.GasUsed, tx.Fee, tx.HasSCR, string(marshaledTx)))

	return serializedData, nil
}
