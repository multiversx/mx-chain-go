package pruning

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"runtime/debug"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/clean"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var _ storage.Storer = (*PruningStorer)(nil)

var log = logger.GetOrCreate("storage/pruning")

// maxNumEpochsToKeepIfAShardIsStuck represents the maximum number of epochs to be kept active if a shard remains stuck
// and requires data from older epochs
const maxNumEpochsToKeepIfAShardIsStuck = 5

// epochForDefaultEpochPrepareHdr represents the default epoch number for the meta block which is saved on EpochPrepareAction
// it is useful for checking if any metablock of this kind is received
const epochForDefaultEpochPrepareHdr = math.MaxUint32 - 7

// persisterData structure is used so the persister and its path can be kept in the same place
type persisterData struct {
	persister storage.Persister
	path      string
	epoch     uint32
	isClosed  bool
	sync.RWMutex
}

func (pd *persisterData) getIsClosed() bool {
	pd.RLock()
	defer pd.RUnlock()

	return pd.isClosed
}

func (pd *persisterData) setIsClosed(closed bool) {
	pd.Lock()
	pd.isClosed = closed
	pd.Unlock()
}

// Close closes the underlying persister
func (pd *persisterData) Close() error {
	pd.setIsClosed(true)
	err := pd.persister.Close()
	return err
}

func (pd *persisterData) getPersister() storage.Persister {
	pd.RLock()
	defer pd.RUnlock()

	return pd.persister
}

func (pd *persisterData) setPersisterAndIsClosed(persister storage.Persister, isClosed bool) {
	pd.Lock()
	pd.persister = persister
	pd.isClosed = isClosed
	pd.Unlock()
}

// PruningStorer represents a storer which creates a new persister for each epoch and removes older activePersisters
type PruningStorer struct {
	lock             sync.RWMutex
	shardCoordinator storage.ShardCoordinator
	activePersisters []*persisterData
	// it is mandatory to keep map of pointers for persistersMapByEpoch as a loaded pointer might get modified in inner functions
	persistersMapByEpoch map[uint32]*persisterData
	cacher               storage.Cacher
	// TODO remove bloom filter
	bloomFilter            storage.BloomFilter
	pathManager            storage.PathManagerHandler
	dbPath                 string
	persisterFactory       DbFactoryHandler
	mutEpochPrepareHdr     sync.RWMutex
	epochPrepareHdr        data.HeaderHandler
	oldDataCleanerProvider clean.OldDataCleanerProvider
	identifier             string
	numOfEpochsToKeep      uint32
	numOfActivePersisters  uint32
	epochForPutOperation   uint32
	pruningEnabled         bool
}

// NewPruningStorer will return a new instance of PruningStorer without sharded directories' naming scheme
func NewPruningStorer(args *StorerArgs) (*PruningStorer, error) {
	return initPruningStorer(args, "")
}

// initPruningStorer will create a PruningStorer with or without sharded directories' naming scheme
func initPruningStorer(
	args *StorerArgs,
	shardIDStr string,
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
	if check.IfNil(args.ShardCoordinator) {
		return nil, storage.ErrNilShardCoordinator
	}
	if check.IfNil(args.PathManager) {
		return nil, storage.ErrNilPathManager
	}
	if args.MaxBatchSize > int(args.CacheConf.Capacity) {
		return nil, storage.ErrCacheSizeIsLowerThanBatchSize
	}

	cache, err = storageUnit.NewCache(args.CacheConf)
	if err != nil {
		return nil, err
	}

	persisters, persistersMapByEpoch, err := initPersistersInEpoch(args, shardIDStr)
	if err != nil {
		return nil, err
	}

	identifier := args.Identifier
	if len(shardIDStr) > 0 {
		identifier += shardIDStr
	}

	pdb := &PruningStorer{
		pruningEnabled:         args.PruningEnabled,
		identifier:             identifier,
		activePersisters:       persisters,
		persisterFactory:       args.PersisterFactory,
		shardCoordinator:       args.ShardCoordinator,
		persistersMapByEpoch:   persistersMapByEpoch,
		cacher:                 cache,
		epochPrepareHdr:        &block.MetaBlock{Epoch: epochForDefaultEpochPrepareHdr},
		bloomFilter:            nil,
		epochForPutOperation:   args.StartingEpoch,
		pathManager:            args.PathManager,
		dbPath:                 args.DbPath,
		numOfEpochsToKeep:      args.NumOfEpochsToKeep,
		numOfActivePersisters:  args.NumOfActivePersisters,
		oldDataCleanerProvider: args.OldDataCleanerProvider,
	}

	if args.BloomFilterConf.Size != 0 { // if size is 0, that means an empty config was used so bloom filter will be nil
		bf, err = storageUnit.NewBloomFilter(args.BloomFilterConf)
		if err != nil {
			return nil, err
		}

		pdb.bloomFilter = bf
	}

	pdb.registerHandler(args.Notifier)

	return pdb, nil
}

func initPersistersInEpoch(
	args *StorerArgs,
	shardIDStr string,
) ([]*persisterData, map[uint32]*persisterData, error) {
	var persisters []*persisterData
	persistersMapByEpoch := make(map[uint32]*persisterData)

	if !args.PruningEnabled {
		p, err := createPersisterDataForEpoch(args, 0, shardIDStr)
		if err != nil {
			return nil, nil, err
		}
		persisters = append(persisters, p)
		return persisters, persistersMapByEpoch, nil
	}

	if args.NumOfEpochsToKeep < args.NumOfActivePersisters {
		return nil, nil, fmt.Errorf("invalid epochs configuration")
	}

	oldestEpochKeep := int64(args.StartingEpoch) - int64(args.NumOfEpochsToKeep) + 1
	if oldestEpochKeep < 0 {
		oldestEpochKeep = 0
	}
	oldestEpochActive := int64(args.StartingEpoch) - int64(args.NumOfActivePersisters) + 1
	if oldestEpochActive < 0 {
		oldestEpochActive = 0
	}

	log.Debug("initPersistersInEpoch",
		"StartingEpoch", args.StartingEpoch,
		"NumOfEpochsToKeep", args.NumOfEpochsToKeep,
		"oldestEpochKeep", oldestEpochKeep,
		"NumOfActivePersisters", args.NumOfActivePersisters,
		"oldestEpochActive", oldestEpochActive,
	)

	for epoch := int64(args.StartingEpoch); epoch >= oldestEpochKeep; epoch-- {
		log.Debug("initPersistersInEpoch(): createPersisterDataForEpoch", "identifier", args.Identifier, "epoch", epoch, "shardID", shardIDStr)
		p, err := createPersisterDataForEpoch(args, uint32(epoch), shardIDStr)
		if err != nil {
			return nil, nil, err
		}

		persistersMapByEpoch[uint32(epoch)] = p

		if epoch < oldestEpochActive {
			err = p.Close()
			if err != nil {
				log.Debug("persister.Close()", "identifier", args.Identifier, "error", err.Error())
			}
		} else {
			persisters = append(persisters, p)
			log.Debug("appended a pruning active persister", "epoch", epoch, "identifier", args.Identifier)
		}
	}

	return persisters, persistersMapByEpoch, nil
}

// Put adds data to both cache and persistence medium and updates the bloom filter
func (ps *PruningStorer) Put(key, data []byte) error {
	ps.cacher.Put(key, data, len(data))

	persisterToUse := ps.getPersisterToUse()
	return ps.doPutInPersister(key, data, persisterToUse.getPersister())
}

// PutWithoutCache adds data to persistence medium and updates the bloom filter
func (ps *PruningStorer) PutWithoutCache(key, data []byte) error {
	persisterToUse := ps.getPersisterToUse()

	err := persisterToUse.getPersister().Put(key, data)
	if err != nil {
		return err
	}

	if ps.bloomFilter != nil {
		ps.bloomFilter.Add(key)
	}

	return nil
}

func (ps *PruningStorer) getPersisterToUse() *persisterData {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	persisterToUse := ps.activePersisters[0]
	if !ps.pruningEnabled {
		return persisterToUse
	}

	persisterInSetEpoch, ok := ps.persistersMapByEpoch[ps.epochForPutOperation]
	if ok && !persisterInSetEpoch.getIsClosed() {
		persisterToUse = persisterInSetEpoch
	} else {
		log.Debug("active persister not found",
			"epoch", ps.epochForPutOperation,
			"used", persisterToUse.epoch)
	}

	return persisterToUse
}

func (ps *PruningStorer) doPutInPersister(key, data []byte, persister storage.Persister) error {
	err := persister.Put(key, data)
	if err != nil {
		ps.cacher.Remove(key)
		return err
	}

	if ps.bloomFilter != nil {
		ps.bloomFilter.Add(key)
	}

	return nil
}

// PutInEpoch adds data to specified epoch
func (ps *PruningStorer) PutInEpoch(key, data []byte, epoch uint32) error {
	ps.cacher.Put(key, data, len(data))

	ps.lock.RLock()
	pd, exists := ps.persistersMapByEpoch[epoch]
	ps.lock.RUnlock()
	if !exists {
		return fmt.Errorf("put in epoch: persister for epoch %d not found", epoch)
	}

	persister, closePersister, err := ps.createAndInitPersisterIfClosedProtected(pd)
	if err != nil {
		return err
	}
	defer closePersister()

	return ps.doPutInPersister(key, data, persister)
}

func (ps *PruningStorer) createAndInitPersisterIfClosedProtected(pd *persisterData) (storage.Persister, func(), error) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	return ps.createAndInitPersisterIfClosedUnprotected(pd)
}

func (ps *PruningStorer) createAndInitPersisterIfClosedUnprotected(pd *persisterData) (storage.Persister, func(), error) {
	isOpen := !pd.getIsClosed()
	if isOpen {
		noopClose := func() {}
		return pd.getPersister(), noopClose, nil
	}

	return ps.createAndInitPersister(pd)
}

func (ps *PruningStorer) createAndInitPersister(pd *persisterData) (storage.Persister, func(), error) {
	isOpen := !pd.getIsClosed()
	if isOpen {
		noopClose := func() {}
		return pd.getPersister(), noopClose, nil
	}

	persister, err := ps.persisterFactory.Create(pd.path)
	if err != nil {
		log.Warn("createAndInitPersister()", "error", err.Error())
		return nil, nil, err
	}

	closeFunc := func() {
		err = pd.Close()
		if err != nil {
			log.Warn("createAndInitPersister(): persister.Close()", "error", err.Error())
		}
	}

	pd.setPersisterAndIsClosed(persister, false)

	return persister, closeFunc, nil
}

// Get searches the key in the cache. In case it is not found, it verifies with the bloom filter
// if the key may be in the db. If bloom filter confirms then it further searches in the databases.
func (ps *PruningStorer) Get(key []byte) ([]byte, error) {
	v, ok := ps.cacher.Get(key)
	if ok {
		return v.([]byte), nil
	}

	if ps.bloomFilter != nil && !ps.bloomFilter.MayContain(key) {
		return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
	}

	// not found in cache
	// search it in active persisters
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	for idx := uint32(0); (idx < ps.numOfActivePersisters) && (idx < uint32(len(ps.activePersisters))); idx++ {
		val, err := ps.activePersisters[idx].persister.Get(key)
		if err != nil {
			continue
		}

		// if found in persistence unit, add it to cache and return
		_ = ps.cacher.Put(key, val, len(val))
		return val, nil
	}

	return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
}

// GetFromOldEpochsWithoutCache searches the old epochs for the given key without updating the cache
func (ps *PruningStorer) GetFromOldEpochsWithoutCache(key []byte) ([]byte, error) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	if ps.bloomFilter != nil && !ps.bloomFilter.MayContain(key) {
		return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
	}

	for idx := uint32(1); (idx < ps.numOfActivePersisters) && (idx < uint32(len(ps.activePersisters))); idx++ {
		val, err := ps.activePersisters[idx].persister.Get(key)
		if err != nil {
			continue
		}

		return val, nil
	}

	return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
}

// GetFromLastEpoch searches only the last epoch storer for the given key
func (ps *PruningStorer) GetFromLastEpoch(key []byte) ([]byte, error) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	if ps.bloomFilter != nil && !ps.bloomFilter.MayContain(key) {
		return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
	}

	if len(ps.activePersisters) < 2 {
		return nil, fmt.Errorf("key %s not found in %s", hex.EncodeToString(key), ps.identifier)
	}

	return ps.activePersisters[1].persister.Get(key)
}

// Close will close PruningStorer
func (ps *PruningStorer) Close() error {
	closedSuccessfully := true
	for _, pd := range ps.activePersisters {
		err := pd.Close()

		if err != nil {
			log.Warn("cannot close pd", "error", err)
			closedSuccessfully = false
		}
	}

	ps.cacher.Clear()

	if closedSuccessfully {
		log.Debug("successfully closed pruningStorer", "identifier", ps.identifier)
		return nil
	}

	return storage.ErrClosingPersisters
}

// GetFromEpoch will search a key only in the persister for the given epoch
func (ps *PruningStorer) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	// TODO: this will be used when requesting from resolvers
	v, ok := ps.cacher.Get(key)
	if ok {
		return v.([]byte), nil
	}

	ps.lock.RLock()
	pd, exists := ps.persistersMapByEpoch[epoch]
	ps.lock.RUnlock()
	if !exists {
		return nil, fmt.Errorf("key %s not found in %s",
			hex.EncodeToString(key), ps.identifier)
	}

	persister, closePersister, err := ps.createAndInitPersisterIfClosedProtected(pd)
	if err != nil {
		return nil, err
	}
	defer closePersister()

	res, err := persister.Get(key)
	if err == nil {
		return res, nil
	}

	log.Debug("get from closed persister",
		"id", ps.identifier,
		"epoch", epoch,
		"key", key,
		"error", err.Error())

	return nil, fmt.Errorf("key %s not found in %s",
		hex.EncodeToString(key), ps.identifier)

}

// GetBulkFromEpoch will return a slice of keys only in the persister for the given epoch
func (ps *PruningStorer) GetBulkFromEpoch(keys [][]byte, epoch uint32) (map[string][]byte, error) {
	ps.lock.RLock()
	pd, exists := ps.persistersMapByEpoch[epoch]
	ps.lock.RUnlock()
	if !exists {
		log.Warn("get from removed persister",
			"id", ps.identifier,
			"epoch", epoch)
		return nil, errors.New("persister does not exist")
	}

	persisterToRead, closePersister, err := ps.createAndInitPersisterIfClosedProtected(pd)
	if err != nil {
		return nil, err
	}
	defer closePersister()

	returnMap := make(map[string][]byte)
	for _, key := range keys {
		v, ok := ps.cacher.Get(key)
		if ok {
			returnMap[string(key)] = v.([]byte)
			continue
		}

		res, errGet := persisterToRead.Get(key)
		if errGet != nil {
			log.Warn("cannot get from persister",
				"hash", hex.EncodeToString(key),
				"error", errGet.Error(),
			)
			continue
		}

		returnMap[string(key)] = res
	}

	return returnMap, nil
}

// SearchFirst will search a given key in all the active persisters, from the newest to the oldest
func (ps *PruningStorer) SearchFirst(key []byte) ([]byte, error) {
	v, ok := ps.cacher.Get(key)
	if ok {
		return v.([]byte), nil
	}

	var res []byte
	var err error

	ps.lock.RLock()
	defer ps.lock.RUnlock()
	for _, pd := range ps.activePersisters {
		res, err = pd.getPersister().Get(key)
		if err == nil {
			return res, nil
		}
	}

	return nil, fmt.Errorf("%w - SearchFirst, unit = %s, key = %s, num active persisters = %d",
		storage.ErrKeyNotFound,
		ps.identifier,
		hex.EncodeToString(key),
		len(ps.activePersisters),
	)
}

// Has checks if the key is in the Unit.
// It first checks the cache. If it is not found, it checks the bloom filter
// and if present it checks the db
func (ps *PruningStorer) Has(key []byte) error {
	has := ps.cacher.Has(key)
	if has {
		return nil
	}

	ps.lock.RLock()
	defer ps.lock.RUnlock()
	if ps.bloomFilter == nil || ps.bloomFilter.MayContain(key) {
		for _, persister := range ps.activePersisters {
			if persister.getPersister().Has(key) != nil {
				continue
			}

			return nil
		}
	}

	return storage.ErrKeyNotFound
}

// SetEpochForPutOperation will set the epoch to be used when using the put operation
func (ps *PruningStorer) SetEpochForPutOperation(epoch uint32) {
	ps.lock.Lock()
	ps.epochForPutOperation = epoch
	ps.lock.Unlock()
}

// Remove removes the data associated to the given key from both cache and persistence medium
func (ps *PruningStorer) Remove(key []byte) error {
	var err error
	ps.cacher.Remove(key)

	ps.lock.RLock()
	defer ps.lock.RUnlock()
	for _, pd := range ps.activePersisters {
		err = pd.persister.Remove(key)
		if err == nil {
			return nil
		}
	}

	return err
}

// ClearCache cleans up the entire cache
func (ps *PruningStorer) ClearCache() {
	ps.cacher.Clear()
}

// GetOldestEpoch returns the oldest epoch from current configuration
func (ps *PruningStorer) GetOldestEpoch() (uint32, error) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	if len(ps.persistersMapByEpoch) == 0 {
		return 0, fmt.Errorf("no epoch configuration in pruning storer with identifer %s", ps.identifier)
	}

	oldestEpoch := uint32(math.MaxUint32)
	for epoch := range ps.persistersMapByEpoch {
		if epoch < oldestEpoch {
			oldestEpoch = epoch
		}
	}

	log.Debug("pruningStorer.GetOldestEpoch", "identifier", ps.identifier, "epoch", oldestEpoch)

	return oldestEpoch, nil
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
	totalNumOfPersisters := len(ps.persistersMapByEpoch)
	for _, pd := range ps.persistersMapByEpoch {
		if pd.getIsClosed() {
			err = pd.getPersister().DestroyClosed()
		} else {
			err = pd.getPersister().Destroy()
		}

		if err != nil {
			log.Debug("pruning db: destroy",
				"error", err.Error())
			continue
		}
		numOfPersistersRemoved++
	}

	if numOfPersistersRemoved != totalNumOfPersisters {
		log.Debug("error destroying pruning db",
			"identifier", ps.identifier,
			"destroyed", numOfPersistersRemoved,
			"total", totalNumOfPersisters)
		return storage.ErrDestroyingUnit
	}

	return nil
}

// registerHandler will register a new function to the epoch start notifier
func (ps *PruningStorer) registerHandler(handler EpochStartNotifier) {
	subscribeHandler := notifier.NewHandlerForEpochStart(
		func(hdr data.HeaderHandler) {
			err := ps.changeEpoch(hdr)
			if err != nil {
				log.Warn("change epoch in storer", "error", err.Error())
			}
		},
		func(metaHdr data.HeaderHandler) {
			err := ps.saveHeaderForEpochStartPrepare(metaHdr)
			if err != nil {
				log.Warn("prepare epoch change in storer", "error", err.Error())
			}
		},
		common.StorerOrder)

	handler.RegisterHandler(subscribeHandler)
}

func (ps *PruningStorer) saveHeaderForEpochStartPrepare(header data.HeaderHandler) error {
	ps.mutEpochPrepareHdr.Lock()
	defer ps.mutEpochPrepareHdr.Unlock()

	var ok bool
	ps.epochPrepareHdr, ok = header.(*block.MetaBlock)
	if !ok {
		return storage.ErrWrongTypeAssertion
	}

	return nil
}

// changeEpoch will handle creating a new persister and removing of the older ones
func (ps *PruningStorer) changeEpoch(header data.HeaderHandler) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	epoch := header.GetEpoch()
	log.Debug("PruningStorer - change epoch", "unit", ps.identifier, "epoch", epoch, "bytes in cache", ps.cacher.SizeInBytesContained())
	// if pruning is not enabled, don't create new persisters, but use the same one instead
	if !ps.pruningEnabled {
		log.Debug("PruningStorer - change epoch - pruning is disabled")
		return nil
	}

	_, ok := ps.persistersMapByEpoch[epoch]
	if ok {
		err := ps.changeEpochWithExisting(epoch)
		if err != nil {
			log.Warn("change epoch", "epoch", epoch, "error", err)
			return err
		}
		log.Debug("change epoch pruning storer success", "persister", ps.identifier, "epoch", epoch)

		return nil
	}

	shardID := core.GetShardIDString(ps.shardCoordinator.SelfId())
	filePath := ps.pathManager.PathForEpoch(shardID, epoch, ps.identifier)
	db, err := ps.persisterFactory.Create(filePath)
	if err != nil {
		log.Warn("change epoch", "persister", ps.identifier, "error", err.Error())
		return err
	}

	newPersister := &persisterData{
		persister: db,
		epoch:     epoch,
		path:      filePath,
		isClosed:  false,
	}

	singleItemPersisters := []*persisterData{newPersister}

	ps.activePersisters = append(singleItemPersisters, ps.activePersisters...)
	ps.persistersMapByEpoch[epoch] = newPersister

	wasExtended := ps.extendSavedEpochsIfNeeded(header)
	if wasExtended {
		if len(ps.activePersisters) > int(ps.numOfActivePersisters) {
			log.Debug("PruningStorer - skip closing and destroying persisters due to a stuck shard -",
				"current epoch", epoch,
				"num active persisters", len(ps.activePersisters),
				"default maximum num active persisters", ps.numOfActivePersisters,
				"oldest epoch in storage", ps.activePersisters[len(ps.activePersisters)-1].epoch)
		}
		return nil
	}

	err = ps.closePersisters(epoch)
	if err != nil {
		log.Warn("closing persisters", "error", err.Error())
		return err
	}
	return nil
}

// should be called under mutex protection
func (ps *PruningStorer) extendSavedEpochsIfNeeded(header data.HeaderHandler) bool {
	epoch := header.GetEpoch()
	metaBlock, mbOk := header.(*block.MetaBlock)
	if !mbOk {
		ps.mutEpochPrepareHdr.RLock()
		epochPrepareHdr := ps.epochPrepareHdr
		ps.mutEpochPrepareHdr.RUnlock()
		if epochPrepareHdr != nil {
			var ok bool
			metaBlock, ok = epochPrepareHdr.(*block.MetaBlock)
			if !ok {
				log.Warn("PruningStorer.extendSavedEpochsIfNeeded", "error", "invalid type assertion")
				return false
			}
		} else {
			return false
		}
	}
	shouldExtend := metaBlock.Epoch != epochForDefaultEpochPrepareHdr
	if !shouldExtend {
		return false
	}

	oldestEpochInCurrentSetting := ps.activePersisters[len(ps.activePersisters)-1].epoch

	oldestEpochToKeep := computeOldestEpoch(metaBlock)
	shouldKeepOlderEpochsIfShardIsStuck := epoch-oldestEpochToKeep < maxNumEpochsToKeepIfAShardIsStuck
	if oldestEpochToKeep <= oldestEpochInCurrentSetting && shouldKeepOlderEpochsIfShardIsStuck {
		err := ps.extendActivePersisters(oldestEpochToKeep, oldestEpochInCurrentSetting)
		if err != nil {
			log.Warn("PruningStorer - extend epochs", "error", err)
			return false
		}

		return true
	}

	return false
}

// should be called under mutex protection
func (ps *PruningStorer) changeEpochWithExisting(epoch uint32) error {
	numActivePersisters := ps.numOfActivePersisters

	activePersisters := make([]*persisterData, 0, numActivePersisters)

	oldestEpochActive := int64(epoch) - int64(numActivePersisters) + 1
	if oldestEpochActive < 0 {
		oldestEpochActive = 0
	}

	persisters := make([]*persisterData, 0)
	for e := int64(epoch); e >= oldestEpochActive; e-- {
		p, ok := ps.persistersMapByEpoch[uint32(e)]
		if !ok {
			return nil
		}
		persisters = append(persisters, p)
	}

	for _, p := range persisters {
		if p.getIsClosed() {
			db, errCreate := ps.persisterFactory.Create(p.path)
			if errCreate != nil {
				return errCreate
			}

			p.setPersisterAndIsClosed(db, false)
		}

		activePersisters = append(activePersisters, p)
	}

	ps.activePersisters = activePersisters

	return nil
}

// should be called under mutex protection
func (ps *PruningStorer) extendActivePersisters(from uint32, to uint32) error {
	persisters := make([]*persisterData, 0)
	for e := int(to); e >= int(from); e-- {
		p, ok := ps.persistersMapByEpoch[uint32(e)]
		if !ok {
			return nil
		}
		persisters = append(persisters, p)
	}

	reOpenedPersisters := make([]*persisterData, 0)
	for _, p := range persisters {
		if p.getIsClosed() {
			persister, err := ps.persisterFactory.Create(p.path)
			if err != nil {
				return err
			}
			reOpenedPersisters = append(reOpenedPersisters, p)
			p.setPersisterAndIsClosed(persister, false)
		}
	}

	ps.activePersisters = append(ps.activePersisters, reOpenedPersisters...)

	return nil
}

// should be called under mutex protection
func (ps *PruningStorer) closePersisters(epoch uint32) error {
	// activePersisters outside the numOfActivePersisters border have to he closed for both scenarios: full archive or not
	persistersToClose := make([]*persisterData, 0)

	if ps.numOfActivePersisters < uint32(len(ps.activePersisters)) {
		for idx := int(ps.numOfActivePersisters); idx < len(ps.activePersisters); idx++ {
			persisterToClose := ps.activePersisters[idx]
			// remove it from the active persisters slice
			ps.activePersisters = ps.activePersisters[:ps.numOfActivePersisters]
			ps.persistersMapByEpoch[persisterToClose.epoch] = persisterToClose
			persistersToClose = append(persistersToClose, persisterToClose)
		}
	}

	if ps.oldDataCleanerProvider.ShouldClean() && uint32(len(ps.persistersMapByEpoch)) > ps.numOfEpochsToKeep {
		idxToRemove := epoch - ps.numOfEpochsToKeep
		for {
			_, exists := ps.persistersMapByEpoch[idxToRemove]
			if !exists {
				break
			}
			delete(ps.persistersMapByEpoch, idxToRemove)
			idxToRemove--
		}
	}

	for _, pd := range persistersToClose {
		err := pd.Close()
		if err != nil {
			log.Warn("error closing persister", "error", err.Error(), "id", ps.identifier)
		}
	}

	return nil
}

// RangeKeys does nothing as it is unable to iterate over multiple persisters
// RangeKeys -
func (ps *PruningStorer) RangeKeys(_ func(key []byte, val []byte) bool) {
	log.Error("improper use of PruningStorer.RangeKeys()")
	debug.PrintStack()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ps *PruningStorer) IsInterfaceNil() bool {
	return ps == nil
}

func createPersisterPathForEpoch(args *StorerArgs, epoch uint32, shard string) string {
	filePath := args.PathManager.PathForEpoch(core.GetShardIDString(args.ShardCoordinator.SelfId()), epoch, args.Identifier)
	if len(shard) > 0 {
		filePath += shard
	}

	return filePath
}

func createPersisterDataForEpoch(args *StorerArgs, epoch uint32, shard string) (*persisterData, error) {
	// TODO: if booting from storage in an epoch > 0, shardId needs to be taken from somewhere else
	// e.g. determined from directories in persister path or taken from boot storer
	filePath := createPersisterPathForEpoch(args, epoch, shard)

	db, err := args.PersisterFactory.Create(filePath)
	if err != nil {
		log.Warn("persister create error", "error", err.Error())
		return nil, err
	}

	p := &persisterData{
		persister: db,
		epoch:     epoch,
		path:      filePath,
		isClosed:  false,
	}

	return p, nil
}

func computeOldestEpoch(metaBlock *block.MetaBlock) uint32 {
	oldestEpoch := metaBlock.Epoch
	for _, lastHdr := range metaBlock.EpochStart.LastFinalizedHeaders {
		if lastHdr.Epoch < oldestEpoch {
			oldestEpoch = lastHdr.Epoch
		}
	}

	return oldestEpoch
}
