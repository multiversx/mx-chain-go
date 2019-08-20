package dataPool

import (
	"bytes"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.DefaultLogger()

type nonceSyncMapCacher struct {
	mergeMut             sync.Mutex
	cacher               storage.Cacher
	nonceConverter       typeConverters.Uint64ByteSliceConverter
	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func(nonce uint64, shardId uint32, value []byte)
}

// NewNonceSyncMapCacher returns a new instance of nonceSyncMapCacher
func NewNonceSyncMapCacher(
	cacher storage.Cacher,
	nonceConverter typeConverters.Uint64ByteSliceConverter,
) (*nonceSyncMapCacher, error) {

	if cacher == nil {
		return nil, dataRetriever.ErrNilCacher
	}
	if nonceConverter == nil {
		return nil, dataRetriever.ErrNilUint64ByteSliceConverter
	}

	return &nonceSyncMapCacher{
		cacher:            cacher,
		nonceConverter:    nonceConverter,
		addedDataHandlers: make([]func(nonce uint64, shardId uint32, value []byte), 0),
	}, nil
}

// Clear is used to completely clear the cache.
func (nspc *nonceSyncMapCacher) Clear() {
	nspc.cacher.Clear()
}

// Get looks up for a nonce in cache.
func (nspc *nonceSyncMapCacher) Get(nonce uint64) (dataRetriever.ShardIdHashMap, bool) {
	val, ok := nspc.cacher.Peek(nspc.nonceConverter.ToByteSlice(nonce))
	if !ok {
		return nil, ok
	}

	syncMap, ok := val.(*ShardIdHashSyncMap)
	if !ok {
		return nil, ok
	}

	return syncMap, ok
}

// Merge will append existing values from src map. If the keys already exists in the existing map, their values
// will be overwritten. If the existing map is nil, a new map will created and all values from src map will be copied.
func (nspc *nonceSyncMapCacher) Merge(nonce uint64, src dataRetriever.ShardIdHashMap) {
	if src == nil {
		return
	}

	nspc.mergeMut.Lock()
	defer nspc.mergeMut.Unlock()

	shouldRewriteMap := false
	val, ok := nspc.cacher.Peek(nspc.nonceConverter.ToByteSlice(nonce))
	if !ok {
		val = &ShardIdHashSyncMap{}
		shouldRewriteMap = true
	}

	syncMap := val.(*ShardIdHashSyncMap)

	if shouldRewriteMap {
		nspc.cacher.Put(nspc.nonceConverter.ToByteSlice(nonce), syncMap)
	}

	nspc.copySyncMap(nonce, syncMap, src)
}

func (nspc *nonceSyncMapCacher) copySyncMap(nonce uint64, dest dataRetriever.ShardIdHashMap, src dataRetriever.ShardIdHashMap) {
	src.Range(func(shardId uint32, hash []byte) bool {
		existingVal, exists := dest.Load(shardId)
		if !exists {
			//new key with value
			dest.Store(shardId, hash)
			nspc.callAddedDataHandlers(nonce, shardId, hash)
			return true
		}

		if !bytes.Equal(existingVal, hash) {
			//value mismatch
			dest.Store(shardId, hash)
			nspc.callAddedDataHandlers(nonce, shardId, hash)
			return true
		}

		return true
	})
}

// Remove removes the nonce-shardId-hash tuple using the nonce and shardId
func (nspc *nonceSyncMapCacher) Remove(nonce uint64, shardId uint32) {
	val, ok := nspc.cacher.Peek(nspc.nonceConverter.ToByteSlice(nonce))
	if !ok {
		return
	}

	syncMap, ok := val.(*ShardIdHashSyncMap)
	if !ok {
		return
	}

	syncMap.Delete(shardId)
}

// RegisterHandler registers a new handler to be called when a new data is added
func (nspc *nonceSyncMapCacher) RegisterHandler(handler func(nonce uint64, shardId uint32, value []byte)) {
	if handler == nil {
		log.Error("attempt to register a nil handler to a cacher object")
		return
	}

	nspc.mutAddedDataHandlers.Lock()
	nspc.addedDataHandlers = append(nspc.addedDataHandlers, handler)
	nspc.mutAddedDataHandlers.Unlock()
}

func (nspc *nonceSyncMapCacher) callAddedDataHandlers(nonce uint64, shardId uint32, val []byte) {
	nspc.mutAddedDataHandlers.RLock()
	for _, handler := range nspc.addedDataHandlers {
		go handler(nonce, shardId, val)
	}
	nspc.mutAddedDataHandlers.RUnlock()
}

// Has returns true if a map is found for provided nonce ans shardId
func (nspc *nonceSyncMapCacher) Has(nonce uint64, shardId uint32) bool {
	val, ok := nspc.cacher.Peek(nspc.nonceConverter.ToByteSlice(nonce))
	if !ok {
		return false
	}

	syncMap, ok := val.(*ShardIdHashSyncMap)
	if !ok {
		return false
	}

	_, exists := syncMap.Load(shardId)

	return exists
}
