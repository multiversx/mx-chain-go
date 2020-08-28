package databasereader

import (
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("databasereader")

const shardDirectoryPrefix = "Shard_"
const epochDirectoryPrefix = "Epoch_"

// DatabaseInfo holds data about a specific database
type DatabaseInfo struct {
	Epoch uint32
	Shard uint32
}

// Args holds the arguments needed for creating a new database reader
type Args struct {
	DirectoryReader   storage.DirectoryReaderHandler
	GeneralConfig     config.Config
	Marshalizer       marshal.Marshalizer
	PersisterFactory  storage.PersisterFactory
	DbPathWithChainID string
}

type databaseReader struct {
	directoryReader   storage.DirectoryReaderHandler
	generalConfig     config.Config
	marshalizer       marshal.Marshalizer
	persisterFactory  storage.PersisterFactory
	dbPathWithChainID string
}

// New will return a new instance of databaseReader
func New(args Args) (*databaseReader, error) {
	if len(args.DbPathWithChainID) == 0 {
		return nil, ErrEmptyDbFilePath
	}
	if check.IfNil(args.DirectoryReader) {
		return nil, ErrNilDirectoryReader
	}
	if check.IfNil(args.Marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(args.PersisterFactory) {
		return nil, ErrNilPersisterFactory
	}

	return &databaseReader{
		directoryReader:   args.DirectoryReader,
		generalConfig:     args.GeneralConfig,
		marshalizer:       args.Marshalizer,
		persisterFactory:  args.PersisterFactory,
		dbPathWithChainID: args.DbPathWithChainID,
	}, nil
}

// GetDatabaseInfo returns all the databases' data found in the path specified on the constructor
func (dr *databaseReader) GetDatabaseInfo() ([]*DatabaseInfo, error) {
	dbs := make([]*DatabaseInfo, 0)
	epochsDirectories, err := dr.directoryReader.ListDirectoriesAsString(dr.dbPathWithChainID)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile("[0-9]+")

	for _, dirname := range epochsDirectories {
		epochStr := re.FindString(dirname)
		epoch, errParseInt := strconv.ParseInt(epochStr, 10, 64)
		if errParseInt != nil {
			log.Warn("cannot parse epoch number from directory name", "directory name", dirname)
			continue
		}

		shardDirectories, errGetShardDirs := dr.getShardDirectoriesForEpoch(filepath.Join(dr.dbPathWithChainID, dirname))
		if errGetShardDirs != nil {
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
		return nil, ErrNoDatabaseFound
	}

	sort.Slice(dbs, func(i, j int) bool {
		if dbs[i].Epoch < dbs[j].Epoch {
			return true
		}
		if dbs[i].Epoch > dbs[i].Epoch {
			return false
		}

		return dbs[i].Shard > dbs[j].Shard
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
		isShardDir := strings.HasPrefix(dirName, shardDirectoryPrefix)
		if !isShardDir {
			continue
		}

		shardDirs = append(shardDirs, dirName)
	}

	for _, fileName := range shardDirs {
		stringToSplitBy := shardDirectoryPrefix
		splitSlice := strings.Split(fileName, stringToSplitBy)
		if len(splitSlice) < 2 {
			continue
		}

		shardID := core.MetachainShardId
		shardIdStr := splitSlice[1]
		if shardIdStr != "metachain" {
			u64, errParseUint := strconv.ParseUint(shardIdStr, 10, 32)
			if errParseUint != nil {
				log.Warn("cannot parse shard ID from directory", "directory name", fileName)
				continue
			}
			shardID = uint32(u64)
		}

		shardIDs = append(shardIDs, shardID)
	}

	return shardIDs, nil
}

// GetStaticDatabaseInfo returns all the databases' data found in the path specified on the constructor
func (dr *databaseReader) GetStaticDatabaseInfo() ([]*DatabaseInfo, error) {
	dbs := make([]*DatabaseInfo, 0)
	dirname := "Static"
	shardDirectories, err := dr.getShardDirectoriesForEpoch(filepath.Join(dr.dbPathWithChainID, dirname))
	if err != nil {
		log.Warn("cannot parse shard directories for epoch", "epoch directory name", dirname)
		return nil, err
	}

	for _, shardDir := range shardDirectories {
		dbs = append(dbs,
			&DatabaseInfo{
				Epoch: 0,
				Shard: shardDir,
			})
	}

	if len(dbs) == 0 {
		return nil, ErrNoDatabaseFound
	}

	sort.Slice(dbs, func(i, j int) bool {
		if dbs[i].Epoch < dbs[j].Epoch {
			return true
		}
		if dbs[i].Epoch > dbs[i].Epoch {
			return false
		}

		return dbs[i].Shard > dbs[j].Shard
	})

	return dbs, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dr *databaseReader) IsInterfaceNil() bool {
	return dr == nil
}
