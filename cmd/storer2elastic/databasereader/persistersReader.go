package databasereader

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/storage"
)

func (dr *databaseReader) GetHeaders(dbInfo *DatabaseInfo) ([]data.HeaderHandler, error) {
	hdrStorer := shardBlocksStorer
	if dbInfo.Shard == core.MetachainShardId {
		hdrStorer = metaHeadersStorer
	}

	shardIDStr := fmt.Sprintf("%d", dbInfo.Shard)
	if shardIDStr == fmt.Sprintf("%d", core.MetachainShardId) {
		shardIDStr = "metachain"
	}
	persisterPath := filepath.Join(dr.dbPathWithChainID, fmt.Sprintf("Epoch_%d", dbInfo.Epoch), fmt.Sprintf("Shard_%s", shardIDStr), hdrStorer)
	hdrPersister, err := dr.persisterFactory.Create(persisterPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = hdrPersister.Close()
	}()

	records := make([]core.KeyValueHolder, 0)
	recordsChannel := hdrPersister.Iterate()

	for rec := range recordsChannel {
		records = append(records, rec)
	}

	hdrs := make([]data.HeaderHandler, 0)
	for _, rec := range records {
		hdrBytes := rec.Value()
		var hdr data.HeaderHandler
		var errUmarshal error
		if hdrStorer == shardBlocksStorer {
			hdr, errUmarshal = dr.unmarshalShardHeader(hdrBytes)
		} else {
			hdr, errUmarshal = dr.unmarshalMetaBlock(hdrBytes)
		}
		if errUmarshal != nil {
			log.Warn("error unmarshalling header", "error", errUmarshal)
			continue
		}

		hdrs = append(hdrs, hdr)
	}

	if len(hdrs) == 0 {
		return nil, errors.New("no header")
	}
	return hdrs, nil
}

func (dr *databaseReader) unmarshalShardHeader(hdrBytes []byte) (*block.Header, error) {
	blck := &block.Header{}
	err := dr.marshalizer.Unmarshal(blck, hdrBytes)
	if err != nil {
		return nil, err
	}

	return blck, nil
}

func (dr *databaseReader) unmarshalMetaBlock(hdrBytes []byte) (*block.MetaBlock, error) {
	blck := &block.MetaBlock{}
	err := dr.marshalizer.Unmarshal(blck, hdrBytes)
	if err != nil {
		return nil, err
	}

	return blck, nil
}

func (dr *databaseReader) LoadPersister(dbInfo *DatabaseInfo, unit string) (storage.Persister, error) {
	shardIDStr := fmt.Sprintf("%d", dbInfo.Shard)
	if shardIDStr == fmt.Sprintf("%d", core.MetachainShardId) {
		shardIDStr = "metachain"
	}

	persisterPath := filepath.Join(dr.dbPathWithChainID, fmt.Sprintf("Epoch_%d", dbInfo.Epoch), fmt.Sprintf("Shard_%s", shardIDStr), unit)
	return dr.persisterFactory.Create(persisterPath)
}
