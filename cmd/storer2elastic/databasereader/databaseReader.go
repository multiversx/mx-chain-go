package databasereader

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/leveldb"
	leveldb2 "github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var log = logger.GetOrCreate("databasereader")

const shardBlocksStorer = "BlockHeaders"
const metaHeadersStorer = "MetaBlock"
const miniBlockStorer = "MiniBlocks"
const maxOpenFiles = 100

type DatabaseInfo struct {
	Epoch uint32
	Shard uint32
}

type databaseReader struct {
	directoryReader   storage.DirectoryReaderHandler
	marshalizer       marshal.Marshalizer
	dbPathWithChainID string
}

// NewDatabaseReader will return a new instance of databaseReader
func NewDatabaseReader(dbFilePath string, marshalizer marshal.Marshalizer) (*databaseReader, error) {
	if len(dbFilePath) == 0 {
		return nil, ErrEmptyDbFilePath
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}

	return &databaseReader{
		directoryReader:   factory.NewDirectoryReader(),
		marshalizer:       marshalizer,
		dbPathWithChainID: dbFilePath,
	}, nil
}

func (dr *databaseReader) GetDatabaseInfo() ([]*DatabaseInfo, error) {
	dbs := make([]*DatabaseInfo, 0)
	epochsDirectories, err := dr.directoryReader.ListDirectoriesAsString(dr.dbPathWithChainID)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile("[0-9]+")

	for _, dirname := range epochsDirectories {
		epochStr := re.FindString(dirname)
		epoch, err := strconv.ParseInt(epochStr, 10, 64)
		if err != nil {
			log.Warn("cannot parse epoch number from directory name", "directory name", dirname)
			continue
		}

		shardDirectories, err := dr.getShardDirectoriesForEpoch(filepath.Join(dr.dbPathWithChainID, dirname))
		if err != nil {
			log.Warn("cannot parse shard directories for epoch", "epoch directory name", dirname)
			continue
		}

		for _, shardDir := range shardDirectories {
			dbs = append(dbs,
				&DatabaseInfo{
					Epoch: uint32(epoch),
					Shard: shardDir,
				})
		}
	}

	if len(dbs) == 0 {
		return nil, errors.New("no database found")
	}

	return dbs, nil
}

func (dr *databaseReader) getShardDirectoriesForEpoch(dirEpoch string) ([]uint32, error) {
	shardIDs := make([]uint32, 0)
	directoriesNames, err := dr.directoryReader.ListDirectoriesAsString(dirEpoch)
	if err != nil {
		return nil, err
	}

	shardDirs := make([]string, 0, len(directoriesNames))
	for _, dirName := range directoriesNames {
		isShardDir := strings.HasPrefix(dirName, "Shard_")
		if !isShardDir {
			continue
		}

		shardDirs = append(shardDirs, dirName)
	}

	for _, fileName := range shardDirs {
		stringToSplitBy := "Shard_"
		splitSlice := strings.Split(fileName, stringToSplitBy)
		if len(splitSlice) < 2 {
			continue
		}

		shardID := core.MetachainShardId
		shardIdStr := splitSlice[1]
		if shardIdStr != "metachain" {
			u64, err := strconv.ParseUint(shardIdStr, 10, 32)
			if err != nil {
				continue
			}
			shardID = uint32(u64)
		}

		shardIDs = append(shardIDs, shardID)
	}

	return shardIDs, nil
}

// TODO : finish

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
	options := &opt.Options{
		// disable internal cache
		BlockCacheCapacity:     -1,
		OpenFilesCacheCapacity: maxOpenFiles,
	}
	lvlDb, err := leveldb2.OpenFile(persisterPath, options)
	if err != nil {
		return nil, err
	}

	baseLvlDb := leveldb.BaseLevelDb{DB: lvlDb}
	records := make([]core.KeyValHolder, 0)
	recordsChannel := baseLvlDb.Iterate()

	for rec := range recordsChannel {
		records = append(records, rec)
	}

	hdrs := make([]data.HeaderHandler, 0)
	for _, rec := range records {
		hdrBytes := rec.Val()
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

func (dr *databaseReader) GetMiniBlocks(dbInfo *DatabaseInfo) ([]*block.MiniBlock, error) {
	hdrStorer := miniBlockStorer

	shardIDStr := fmt.Sprintf("%d", dbInfo.Shard)
	if shardIDStr == fmt.Sprintf("%d", core.MetachainShardId) {
		shardIDStr = "metachain"
	}
	persisterPath := filepath.Join(dr.dbPathWithChainID, fmt.Sprintf("Epoch_%d", dbInfo.Epoch), fmt.Sprintf("Shard_%s", shardIDStr), hdrStorer)
	options := &opt.Options{
		// disable internal cache
		BlockCacheCapacity:     -1,
		OpenFilesCacheCapacity: maxOpenFiles,
	}
	lvlDb, err := leveldb2.OpenFile(persisterPath, options)
	if err != nil {
		return nil, err
	}

	baseLvlDb := leveldb.BaseLevelDb{DB: lvlDb}
	records := make([]core.KeyValHolder, 0)
	recordsChannel := baseLvlDb.Iterate()

	for rec := range recordsChannel {
		records = append(records, rec)
	}

	miniBlocks := make([]*block.MiniBlock, 0)
	for _, rec := range records {
		mb := &block.MiniBlock{}
		err := dr.marshalizer.Unmarshal(mb, rec.Val())
		if err != nil {
			log.Warn("cannot unmarshal miniblock", "error", err)
			continue
		}

		miniBlocks = append(miniBlocks, mb)
	}

	return miniBlocks, nil
}
