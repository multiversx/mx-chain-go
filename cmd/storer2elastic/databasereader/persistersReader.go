package databasereader

import (
	"fmt"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// GetHeaders returns all the headers found in meta block or shard header units
func (dr *databaseReader) GetHeaders(dbInfo *DatabaseInfo) ([]data.HeaderHandler, error) {
	hdrStorer := dr.generalConfig.BlockHeaderStorage.DB.FilePath
	shardIDStr := fmt.Sprintf("%d", dbInfo.Shard)
	if dbInfo.Shard == core.MetachainShardId {
		hdrStorer = dr.generalConfig.MetaBlockStorage.DB.FilePath
		shardIDStr = "metachain"
	}

	persisterPath := filepath.Join(
		dr.dbPathWithChainID, fmt.Sprintf("%s%d", epochDirectoryPrefix, dbInfo.Epoch),
		fmt.Sprintf("%s%s", shardDirectoryPrefix, shardIDStr),
		hdrStorer,
	)

	hdrPersister, err := dr.persisterFactory.Create(persisterPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = hdrPersister.Close()
	}()

	hdrs := make([]data.HeaderHandler, 0)
	recordsRangeHandler := func(key []byte, value []byte) bool {
		hdr, errGetHdr := dr.getHeader(hdrStorer, value)
		if errGetHdr != nil {
			log.Warn("error fetching a header", "key", key, "error", errGetHdr)
		} else {
			hdrs = append(hdrs, hdr)
		}
		return true
	}

	hdrPersister.RangeKeys(recordsRangeHandler)

	if len(hdrs) == 0 {
		return nil, ErrNoHeader
	}

	return hdrs, nil
}

func (dr *databaseReader) getHeader(hdrStorer string, value []byte) (data.HeaderHandler, error) {
	var hdr data.HeaderHandler
	var errUmarshal error
	if hdrStorer == dr.generalConfig.BlockHeaderStorage.DB.FilePath {
		hdr, errUmarshal = dr.unmarshalShardHeader(value)
	} else {
		hdr, errUmarshal = dr.unmarshalMetaBlock(value)
	}

	return hdr, errUmarshal
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

// LoadPersister will load the persister based on the database information and the unit
func (dr *databaseReader) LoadPersister(dbInfo *DatabaseInfo, unit string) (storage.Persister, error) {
	shardIDStr := fmt.Sprintf("%d", dbInfo.Shard)
	if shardIDStr == fmt.Sprintf("%d", core.MetachainShardId) {
		shardIDStr = "metachain"
	}

	persisterPath := filepath.Join(dr.dbPathWithChainID, fmt.Sprintf("Epoch_%d", dbInfo.Epoch), fmt.Sprintf("Shard_%s", shardIDStr), unit)
	return dr.persisterFactory.Create(persisterPath)
}

// LoadStaticPersister will load the static persister based on the database information and the unit
func (dr *databaseReader) LoadStaticPersister(dbInfo *DatabaseInfo, unit string) (storage.Persister, error) {
	shardIDStr := fmt.Sprintf("%d", dbInfo.Shard)
	if shardIDStr == fmt.Sprintf("%d", core.MetachainShardId) {
		shardIDStr = "metachain"
	}

	persisterPath := filepath.Join(dr.dbPathWithChainID, "Static", fmt.Sprintf("Shard_%s", shardIDStr), unit)
	return dr.persisterFactory.Create(persisterPath)
}
