package miniblocks

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
)

// SerializeBulkMiniBlocks -
func (mp *miniblocksProcessor) SerializeBulkMiniBlocks(
	hdrShardID uint32,
	bulkMbs []*types.Miniblock,
	getAlreadyIndexedItems func(hashes []string, index string) (map[string]bool, error),
	miniblocksIndex string,
) (bytes.Buffer, map[string]bool) {
	var err error
	var buff bytes.Buffer

	mbsHashes := make([]string, len(bulkMbs))
	for idx := range bulkMbs {
		mbsHashes[idx] = bulkMbs[idx].Hash
	}

	existsInDb, err := getAlreadyIndexedItems(mbsHashes, miniblocksIndex)
	if err != nil {
		log.Warn("indexer get indexed items miniblocks",
			"error", err.Error())
		return buff, make(map[string]bool)
	}

	for _, mb := range bulkMbs {
		var meta, serializedData []byte
		if !existsInDb[mb.Hash] {
			//insert miniblock in database
			meta = []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, mb.Hash, "_doc", "\n"))
			serializedData, err = json.Marshal(mb)
			if err != nil {
				log.Debug("indexer: marshal",
					"error", "could not serialize miniblock, will skip indexing",
					"mb hash", mb.Hash)
				continue
			}
		} else {
			// update miniblock
			meta = []byte(fmt.Sprintf(`{ "update" : { "_id" : "%s" } }%s`, mb.Hash, "\n"))
			if hdrShardID == mb.SenderShardID {
				// update sender block hash
				serializedData = []byte(fmt.Sprintf(`{ "doc" : { "senderBlockHash" : "%s" } }`, mb.SenderBlockHash))
			} else {
				// update receiver block hash
				serializedData = []byte(fmt.Sprintf(`{ "doc" : { "receiverBlockHash" : "%s" } }`, mb.ReceiverBlockHash))
			}
		}

		buff = prepareBufferMiniblocks(buff, meta, serializedData)
	}

	return buff, existsInDb
}

func prepareBufferMiniblocks(buff bytes.Buffer, meta, serializedData []byte) bytes.Buffer {
	// append a newline for each element
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

	return buff
}
