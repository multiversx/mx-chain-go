package miniblocks

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
)

// SerializeBulkMiniBlocks -
func (mp *miniblocksProcessor) SerializeBulkMiniBlocks(
	bulkMbs []*types.Miniblock,
	existsInDb map[string]bool,
) *bytes.Buffer {
	buff := &bytes.Buffer{}
	for _, mb := range bulkMbs {
		meta, serializedData, err := mp.prepareMiniblockData(mb, existsInDb[mb.Hash])
		if err != nil {
			log.Warn("miniblocksProcessor.SerializeBulkMiniBlocks cannot prepare miniblock data", "error", err)
		}

		putInBufferMiniblockData(buff, meta, serializedData)
	}

	return buff
}

func (mp *miniblocksProcessor) prepareMiniblockData(miniblockDB *types.Miniblock, isInDB bool) ([]byte, []byte, error) {
	if isInDB {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, miniblockDB.Hash, "_doc", "\n"))
		serializedData, err := json.Marshal(miniblockDB)

		return meta, serializedData, err
	}

	// prepare data for update operation
	meta := []byte(fmt.Sprintf(`{ "update" : { "_id" : "%s" } }%s`, miniblockDB.Hash, "\n"))
	if mp.selfShardID == miniblockDB.SenderShardID {
		// prepare for update sender block hash
		serializedData := []byte(fmt.Sprintf(`{ "doc" : { "senderBlockHash" : "%s" } }`, miniblockDB.SenderBlockHash))

		return meta, serializedData, nil
	}

	// prepare for update receiver block hash
	serializedData := []byte(fmt.Sprintf(`{ "doc" : { "receiverBlockHash" : "%s" } }`, miniblockDB.ReceiverBlockHash))

	return meta, serializedData, nil
}

func putInBufferMiniblockData(buff *bytes.Buffer, meta, serializedData []byte) {
	serializedData = append(serializedData, "\n"...)
	buff.Grow(len(meta) + len(serializedData))
	_, err := buff.Write(meta)
	if err != nil {
		log.Warn("elastic search: serialize bulk miniblocks, write meta", "error", err.Error())
	}
	_, err = buff.Write(serializedData)
	if err != nil {
		log.Warn("elastic search: serialize bulk miniblocks, write serialized miniblock", "error", err.Error())
	}

	return
}
