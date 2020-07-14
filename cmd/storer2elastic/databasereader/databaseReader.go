package databasereader

import (
	"errors"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
)

var log = logger.GetOrCreate("databasereader")

const shardBlocksStorer = "BlockHeaders"
const metaHeadersStorer = "MetaBlock"

type DatabaseInfo struct {
	Epoch uint32
	Shard uint32
}

type databaseReader struct {
	directoryReader   storage.DirectoryReaderHandler
	marshalizer       marshal.Marshalizer
	persisterFactory  storage.PersisterFactory
	dbPathWithChainID string
}

// NewDatabaseReader will return a new instance of databaseReader
func NewDatabaseReader(dbFilePath string, marshalizer marshal.Marshalizer, persisterFactory storage.PersisterFactory) (*databaseReader, error) {
	if len(dbFilePath) == 0 {
		return nil, ErrEmptyDbFilePath
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(persisterFactory) {
		return nil, ErrNilPersisterFactory
	}

	return &databaseReader{
		directoryReader:   factory.NewDirectoryReader(),
		marshalizer:       marshalizer,
		persisterFactory:  persisterFactory,
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

	sort.Slice(dbs, func(i, j int) bool {
		return dbs[i].Epoch < dbs[j].Epoch
	})

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

// IsInterfaceNil returns true if there is no value under the interface
func (dr *databaseReader) IsInterfaceNil() bool {
	return dr == nil
}
