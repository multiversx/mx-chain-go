package libp2p

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

// MutexHolder holds a cache of mutexes: pairs of (key, *sync.Mutex)
type MutexHolder struct {
	//generalMutex is used to serialize the access to the already concurrent safe lrucache
	generalMutex sync.Mutex
	mutexes      *lrucache.LRUCache
}

// NewMutexHolder creates a new instance of MutexHolder with specified capacity.
func NewMutexHolder(mutexesCapacity int) (*MutexHolder, error) {
	mh := &MutexHolder{}
	var err error
	mh.mutexes, err = lrucache.NewCache(mutexesCapacity)
	if err != nil {
		return nil, err
	}

	return mh, nil
}

// Get returns a mutex for the provided key. If the key was not found, it will create a new mutex, save it in the
// cache and returns it.
func (mh *MutexHolder) Get(key string) *sync.Mutex {
	mh.generalMutex.Lock()
	defer mh.generalMutex.Unlock()

	sliceKey := []byte(key)
	val, ok := mh.mutexes.Get(sliceKey)
	if !ok {
		newMutex := &sync.Mutex{}
		mh.mutexes.Put(sliceKey, newMutex)
		return newMutex
	}

	mutex, ok := val.(*sync.Mutex)
	if !ok {
		newMutex := &sync.Mutex{}
		mh.mutexes.Put(sliceKey, newMutex)
		return newMutex
	}

	return mutex
}
