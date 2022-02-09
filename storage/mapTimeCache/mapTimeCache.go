package mapTimeCache

import (
	"bytes"
	"context"
	"encoding/gob"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
)

var ltcLog = logger.GetOrCreate("storage/maptimecache")

// ArgMapTimeCacher is the argument used to create a new mapTimeCacher
type ArgMapTimeCacher struct {
	DefaultSpan time.Duration
	CacheExpiry time.Duration
}

// mapTimeCacher implements a map cache with eviction and inner TimeCacher
type mapTimeCacher struct {
	sync.RWMutex
	dataMap         map[string]interface{}
	timeCache       storage.TimeCacher
	cacheExpiry     time.Duration
	defaultTimeSpan time.Duration
	cancelFunc      func()
}

// NewMapTimeCache creates a new mapTimeCacher
func NewMapTimeCache(arg ArgMapTimeCacher) *mapTimeCacher {
	return &mapTimeCacher{
		dataMap:         make(map[string]interface{}),
		timeCache:       timecache.NewTimeCache(arg.DefaultSpan),
		cacheExpiry:     arg.CacheExpiry,
		defaultTimeSpan: arg.DefaultSpan,
	}
}

// StartSweeping starts a go routine which handles sweeping the time cache
func (mtc *mapTimeCacher) StartSweeping() {
	mtc.timeCache.RegisterHandler(mtc)

	var ctx context.Context
	ctx, mtc.cancelFunc = context.WithCancel(context.Background())

	go func(ctx context.Context) {
		timer := time.NewTimer(mtc.cacheExpiry)
		defer timer.Stop()

		for {
			timer.Reset(mtc.cacheExpiry)

			select {
			case <-timer.C:
				mtc.timeCache.Sweep()
			case <-ctx.Done():
				ltcLog.Info("closing sweep go routine...")
				return
			}
		}
	}(ctx)
}

// OnSweep is the handler called on Sweep method
func (mtc *mapTimeCacher) OnSweep(key []byte) {
	if key == nil {
		return
	}

	mtc.Lock()
	defer mtc.Unlock()

	delete(mtc.dataMap, string(key))
}

// Clear deletes all stored data
func (mtc *mapTimeCacher) Clear() {
	mtc.Lock()
	defer mtc.Unlock()

	mtc.dataMap = make(map[string]interface{})
}

// Put adds a value to the cache. Returns true if an eviction occurred
func (mtc *mapTimeCacher) Put(key []byte, value interface{}, _ int) (evicted bool) {
	mtc.Lock()
	defer mtc.Unlock()

	_, evicted = mtc.dataMap[string(key)]
	mtc.dataMap[string(key)] = value
	if evicted {
		mtc.upsertToTimeCache(key)
		return true
	}

	mtc.addToTimeCache(key)
	return false
}

// Get returns a key's value from the cache
func (mtc *mapTimeCacher) Get(key []byte) (value interface{}, ok bool) {
	mtc.RLock()
	defer mtc.RUnlock()

	v, ok := mtc.dataMap[string(key)]
	return v, ok
}

// Has checks if a key is in the cache
func (mtc *mapTimeCacher) Has(key []byte) bool {
	mtc.RLock()
	defer mtc.RUnlock()

	_, ok := mtc.dataMap[string(key)]
	return ok
}

// Peek returns a key's value from the cache
func (mtc *mapTimeCacher) Peek(key []byte) (value interface{}, ok bool) {
	mtc.RLock()
	defer mtc.RUnlock()

	v, ok := mtc.dataMap[string(key)]
	return v, ok
}

// HasOrAdd checks if a key is in the cache.
// If key exists, does not update the value. Otherwise, adds the key-value in the cache
func (mtc *mapTimeCacher) HasOrAdd(key []byte, value interface{}, _ int) (has, added bool) {
	mtc.Lock()
	defer mtc.Unlock()

	_, ok := mtc.dataMap[string(key)]
	if ok {
		return true, false
	}

	mtc.dataMap[string(key)] = value
	mtc.upsertToTimeCache(key)

	return false, true
}

// Remove removes the key from cache
func (mtc *mapTimeCacher) Remove(key []byte) {
	mtc.Lock()
	defer mtc.Unlock()

	delete(mtc.dataMap, string(key))
}

// Keys returns all keys from cache
func (mtc *mapTimeCacher) Keys() [][]byte {
	mtc.RLock()
	defer mtc.RUnlock()

	keys := make([][]byte, len(mtc.dataMap))
	idx := 0
	for k := range mtc.dataMap {
		keys[idx] = []byte(k)
		idx++
	}
	return keys
}

// Len returns the size of the cache
func (mtc *mapTimeCacher) Len() int {
	mtc.RLock()
	defer mtc.RUnlock()

	return len(mtc.dataMap)
}

// SizeInBytesContained returns the size in bytes of all contained elements
func (mtc *mapTimeCacher) SizeInBytesContained() uint64 {
	mtc.RLock()
	defer mtc.RUnlock()

	totalSize := 0
	b := new(bytes.Buffer)
	for _, v := range mtc.dataMap {
		err := gob.NewEncoder(b).Encode(v)
		if err != nil {
			ltcLog.Error(err.Error())
		} else {
			totalSize += b.Len()
		}
	}

	return uint64(totalSize)
}

// MaxSize returns the maximum number of items which can be stored in cache.
func (mtc *mapTimeCacher) MaxSize() int {
	return 10000
}

// RegisterHandler -
func (mtc *mapTimeCacher) RegisterHandler(_ func(key []byte, value interface{}), _ string) {
}

// UnRegisterHandler -
func (mtc *mapTimeCacher) UnRegisterHandler(_ string) {
}

// Close will close the internal sweep go routine
func (mtc *mapTimeCacher) Close() error {
	if mtc.cancelFunc != nil {
		mtc.cancelFunc()
	}

	return nil
}

func (mtc *mapTimeCacher) addToTimeCache(key []byte) {
	err := mtc.timeCache.Add(string(key))
	if err != nil {
		ltcLog.Error("could not add key", "key", string(key))
	}
}

func (mtc *mapTimeCacher) upsertToTimeCache(key []byte) {
	err := mtc.timeCache.Upsert(string(key), mtc.defaultTimeSpan)
	if err != nil {
		ltcLog.Error("could not upsert timestamp for key", "key", string(key))
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (mtc *mapTimeCacher) IsInterfaceNil() bool {
	return mtc == nil
}
