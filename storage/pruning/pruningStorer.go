package pruning

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var log = logger.GetOrCreate("storage/pruning")

// DefaultEpochDirectoryName represents the naming pattern for epoch directories
const DefaultEpochDirectoryName = "Epoch"

type persisterData struct {
	persister storage.Persister
	path      string
}

// PruningStorer represents a storer which creates a new persister for each epoch and removes older persisters
type PruningStorer struct {
	lock                  sync.RWMutex
	fullArchive           bool
	batcher               storage.Batcher
	persisters            []*persisterData
	closedPersistersPaths []string
	cacher                storage.Cacher
	bloomFilter           storage.BloomFilter
	dbPath                string
	persisterFactory      DbFactoryHandler
	numOfEpochsToKeep     uint32
	numOfActivePersisters uint32
	identifier            string
}

// NewPruningStorer will return a new instance of PruningStorer without sharded directories' naming scheme
func NewPruningStorer(args *PruningStorerArgs) (*PruningStorer, error) {
	return initPruningStorer(args, "")
}

// NewShardedPruningStorer will return a new instance of PruningStorer with sharded directories' naming scheme
func NewShardedPruningStorer(
	args *PruningStorerArgs,
	shardId uint32,
) (*PruningStorer, error) {
	shardIdStr := fmt.Sprintf("%d", shardId)
	return initPruningStorer(args, shardIdStr)
}

// initPruningStorer will create a PruningStorer with or without sharded directories' naming scheme
func initPruningStorer(
	args *PruningStorerArgs,
	shardIdStr string,
) (*PruningStorer, error) {
	var cache storage.Cacher
	var db storage.Persister
	var bf storage.BloomFilter
	var err error

	defer func() {
		if err != nil && db != nil {
			_ = db.Destroy()
		}
	}()

	if args.NumOfActivePersisters < 1 {
		return nil, storage.ErrInvalidNumberOfPersisters
	}
	if check.IfNil(args.Notifier) {
		return nil, storage.ErrNilEpochStartNotifier
	}
	if check.IfNil(args.PersisterFactory) {
		return nil, storage.ErrNilPersisterFactory
	}

	cache, err = storageUnit.NewCache(args.CacheConf.Type, args.CacheConf.Size, args.CacheConf.Shards)
	if err != nil {
		return nil, err
	}

	filePath := args.DbPath
	if len(shardIdStr) > 0 {
		filePath = filePath + shardIdStr
	}
	db, err = args.PersisterFactory.Create(filePath)

	var persisters []*persisterData
	persisters = append(persisters, &persisterData{
		persister: db,
		path:      filePath,
	})

	if reflect.DeepEqual(args.BloomFilterConf, storageUnit.BloomConfig{}) {
		pdb := &PruningStorer{
			identifier:            args.Identifier,
			fullArchive:           args.FullArchive,
			persisters:            persisters,
			persisterFactory:      args.PersisterFactory,
			closedPersistersPaths: make([]string, 0),
			cacher:                cache,
			bloomFilter:           nil,
			dbPath:                filePath,
			numOfEpochsToKeep:     args.NumOfEpochsToKeep,
			numOfActivePersisters: args.NumOfActivePersisters,
		}
		err = pdb.persisters[0].persister.Init()
		if err != nil {
			return nil, err
		}

		pdb.registerHandler(args.Notifier)

		return pdb, nil
	}

	bf, err = storageUnit.NewBloomFilter(args.BloomFilterConf)
	if err != nil {
		return nil, err
	}

	pdb := &PruningStorer{
		identifier:            args.Identifier,
		fullArchive:           args.FullArchive,
		persisters:            persisters,
		persisterFactory:      args.PersisterFactory,
		closedPersistersPaths: make([]string, 0),
		cacher:                cache,
		bloomFilter:           bf,
		dbPath:                filePath,
		numOfEpochsToKeep:     args.NumOfEpochsToKeep,
		numOfActivePersisters: args.NumOfActivePersisters,
	}

	err = pdb.persisters[0].persister.Init()
	if err != nil {
		return nil, err
	}

	pdb.registerHandler(args.Notifier)

	return pdb, nil
}

// Put adds data to both cache and persistence medium and updates the bloom filter
func (ps *PruningStorer) Put(key, data []byte) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	ps.cacher.Put(key, data)

	err := ps.persisters[0].persister.Put(key, data)
	if err != nil {
		ps.cacher.Remove(key)
		return err
	}

	if ps.bloomFilter != nil {
		ps.bloomFilter.Add(key)
	}

	return err
}

// Get searches the key in the cache. In case it is not found, it searches
// for the key in bloom filter first and if found
// it further searches it in the associated databases.
// In case it is found in the database, the cache is updated with the value as well.
func (ps *PruningStorer) Get(key []byte) ([]byte, error) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	v, ok := ps.cacher.Get(key)
	var err error

	if !ok {
		// not found in cache
		// search it in second persistence medium
		found := false
		for idx := uint32(0); (idx < ps.numOfActivePersisters) && (idx < uint32(len(ps.persisters))); idx++ {
			if ps.bloomFilter == nil || ps.bloomFilter.MayContain(key) == true {
				v, err = ps.persisters[idx].persister.Get(key)

				if err != nil {
					log.Debug(ps.identifier+" pruning db - get",
						"error", err.Error())
					continue
				}

				found = true
				// if found in persistence unit, add it in cache
				ps.cacher.Put(key, v)
				break
			}
		}
		if !found && len(ps.closedPersistersPaths) > 0 {
			res, err := ps.getFromClosedPersisters(key)
			if err != nil {
				log.Debug("get from closed persisters", "error", err.Error())
			} else {
				ps.cacher.Put(key, res)
				return res, nil
			}
		}
		if !found {
			return nil, errors.New(fmt.Sprintf("%s: key %s not found", ps.identifier, base64.StdEncoding.EncodeToString(key)))
		}
	}

	return v.([]byte), nil
}

func (ps *PruningStorer) getFromClosedPersisters(key []byte) ([]byte, error) {
	for _, path := range ps.closedPersistersPaths {
		persister, err := ps.persisterFactory.Create(path)
		if err != nil {
			log.Debug("open old persister", "error", err.Error())
			continue
		}

		err = persister.Init()
		if err != nil {
			log.Debug("init old persister", "error", err.Error())
			continue
		}

		res, errToRet := persister.Get(key)

		err = persister.Close()
		if err != nil {
			log.Debug("close old persister", "error", err.Error())
		}

		if errToRet == nil {
			return res, nil
		}

		log.Debug("get from old persister", "error", errToRet.Error())
	}

	return nil, errors.New("key not found")
}

// Has checks if the key is in the Unit.
// It first checks the cache. If it is not found, it checks the bloom filter
// and if present it checks the db
func (ps *PruningStorer) Has(key []byte) error {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	has := ps.cacher.Has(key)
	if has {
		return nil
	}

	if ps.bloomFilter == nil || ps.bloomFilter.MayContain(key) == true {
		for _, persister := range ps.persisters {
			if persister.persister.Has(key) != nil {
				continue
			}

			return nil
		}
	}

	return storage.ErrKeyNotFound
}

// Remove removes the data associated to the given key from both cache and persistence medium
func (ps *PruningStorer) Remove(key []byte) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	ps.cacher.Remove(key)
	return ps.persisters[0].persister.Remove(key)
}

// ClearCache cleans up the entire cache
func (ps *PruningStorer) ClearCache() {
	ps.cacher.Clear()
}

// DestroyUnit cleans up the bloom filter, the cache, and the dbs
func (ps *PruningStorer) DestroyUnit() error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if ps.bloomFilter != nil {
		ps.bloomFilter.Clear()
	}

	ps.cacher.Clear()

	var err error
	numOfPersistersRemoved := 0
	totalNumOfPersisters := len(ps.persisters)
	for _, persister := range ps.persisters {
		err = persister.persister.Destroy()
		if err != nil {
			log.Debug("pruning db: destroy",
				"error", err.Error())
			continue
		}
		numOfPersistersRemoved++
	}

	if numOfPersistersRemoved != totalNumOfPersisters {
		return errors.New(fmt.Sprintf("couldn't destroy all persisters. %d/%d destroyed",
			numOfPersistersRemoved,
			totalNumOfPersisters,
		))
	}
	return ps.persisters[0].persister.Destroy()
}

// registerHandler will register a new function to the epoch start notifier
func (ps *PruningStorer) registerHandler(handler EpochStartNotifier) {
	subscribeHandler := notifier.MakeHandlerForEpochStart(func(hdr data.HeaderHandler) {
		err := ps.changeEpoch(hdr.GetEpoch())
		if err != nil {
			log.Warn("change epoch in storer", "error", err.Error())
		}
	})

	handler.RegisterHandler(subscribeHandler)
}

// changeEpoch will handle creating a new persister and removing of the older ones
func (ps *PruningStorer) changeEpoch(epoch uint32) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	filePath := ps.getNewFilePath(epoch)
	db, err := ps.persisterFactory.Create(filePath)
	if err != nil {
		log.Warn("change epoch error", "error - "+ps.identifier, err.Error())
		return err
	}

	singleItemPersisters := []*persisterData{
		{
			persister: db,
			path:      filePath,
		},
	}
	ps.persisters = append(singleItemPersisters, ps.persisters...)
	err = ps.persisters[0].persister.Init()
	if err != nil {
		return err
	}

	// recent persisters have to he closed for both scenarios: full archive or not
	if ps.numOfActivePersisters < uint32(len(ps.persisters)) {
		err = ps.persisters[ps.numOfActivePersisters].persister.Close()
		if err != nil {
			log.Error("error closing pers", "error", err.Error(), "id", ps.identifier)
			return err
		}
		ps.closedPersistersPaths = append(ps.closedPersistersPaths, ps.persisters[ps.numOfActivePersisters].path)
	}

	// if not is not full archive, destroy old persisters
	if !ps.fullArchive {
		if ps.numOfEpochsToKeep < uint32(len(ps.persisters)) {
			err = ps.persisters[ps.numOfEpochsToKeep].persister.DestroyClosed()
			if err != nil {
				log.Error("error destroy", "error", err.Error(), "id", ps.identifier)
				return err
			}
			//	removeDirectoryIfEmpty(ps.persisters[ps.numOfEpochsToKeep].path)
			ps.persisters = append(ps.persisters[:ps.numOfEpochsToKeep], ps.persisters[ps.numOfEpochsToKeep+1:]...)
		}
	}

	return nil
}

// getNewFilePath will return the file path for the new epoch. It uses regex to change the default path
func (ps *PruningStorer) getNewFilePath(epoch uint32) string {
	// using a regex to match the epoch directory name as placeholder followed by at least one digit
	rg := regexp.MustCompile("Epoch_\\d+")
	newEpochDirectoryName := fmt.Sprintf("%s_%d", DefaultEpochDirectoryName, epoch)
	return rg.ReplaceAllString(ps.dbPath, newEpochDirectoryName)
}

func removeDirectoryIfEmpty(path string) {
	elementsSplitBySeparator := strings.Split(path, string(os.PathSeparator))
	// the structure is this way :
	// workspace/db/Epoch_X/Shard_Y/DbName
	// we need to remove the last 3 if everything is empty

	epochDirectory := ""
	for idx := 0; idx < len(elementsSplitBySeparator)-2; idx++ {
		epochDirectory += elementsSplitBySeparator[idx] + string(os.PathSeparator)
	}

	shardDirectory := epochDirectory + elementsSplitBySeparator[len(elementsSplitBySeparator)-2]
	if isDirectoryEmpty(shardDirectory) {
		err := os.RemoveAll(shardDirectory)
		if err != nil {
			log.Debug("delete old db directory", "error", err.Error())
		}

		err = os.RemoveAll(epochDirectory)
		if err != nil {
			log.Debug("delete old db directory", "error", err.Error())
		}
	}
}

func isDirectoryEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true
	}

	return false // Either not empty or error, suits both cases
}

// IsInterfaceNil -
func (ps *PruningStorer) IsInterfaceNil() bool {
	return ps == nil
}
