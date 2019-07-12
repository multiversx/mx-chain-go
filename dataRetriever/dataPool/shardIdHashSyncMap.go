package dataPool

import "sync"

// ShardIdHashSyncMap is a simple wrapper over a sync map that has specialized methods for load, store, range and so on
type ShardIdHashSyncMap struct {
	innerMap sync.Map
}

// Load returns the hash stored in the map for a shardId, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (sihsm *ShardIdHashSyncMap) Load(shardId uint32) ([]byte, bool) {
	value, ok := sihsm.innerMap.Load(shardId)
	if !ok {
		return nil, ok
	}

	return value.([]byte), ok
}

// Store sets the hash for a shardId.
func (sihsm *ShardIdHashSyncMap) Store(shardId uint32, hash []byte) {
	sihsm.innerMap.Store(shardId, hash)
}

// Range calls f sequentially for each shardId and hash present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func (sihsm *ShardIdHashSyncMap) Range(f func(shardId uint32, hash []byte) bool) {
	if f == nil {
		return
	}

	sihsm.innerMap.Range(func(key, value interface{}) bool {
		shardId := key.(uint32)
		hash := value.([]byte)

		return f(shardId, hash)
	})
}

// Delete deletes the value for a key.
func (sihsm *ShardIdHashSyncMap) Delete(shardId uint32) {
	sihsm.innerMap.Delete(shardId)
}
