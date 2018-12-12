package dataPool

import (
	"encoding/binary"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var log = logger.NewDefaultLogger()

// NonceToHashCacher is a wrapper over a storage.Cacher interface that has the key of type uint64
// and value of type byte slice
type NonceToHashCacher struct {
	cacher storage.Cacher
}

// NewNonceToHashCacher returns a new instance of NonceToHashCacher
func NewNonceToHashCacher(cacherConfig storage.CacheConfig) (*NonceToHashCacher, error) {
	cacher, err := storage.NewCache(cacherConfig.Type, cacherConfig.Size)
	if err != nil {
		return nil, err
	}

	return &NonceToHashCacher{cacher: cacher}, nil
}

//cacher, err := storage.NewCache(cacherConfig.Type, cacherConfig.Size)
//func NewNonceToHashCacher

// Clear is used to completely clear the cache.
func (nthc *NonceToHashCacher) Clear() {
	nthc.cacher.Clear()
}

// Put adds a value to the cache.  Returns true if an eviction occurred.
func (nthc *NonceToHashCacher) Put(nonce uint64, hash []byte) (evicted bool) {
	return nthc.cacher.Put(nthc.nonceToByteArray(nonce), hash)
}

// Get looks up for a nonce in cache.
func (nthc *NonceToHashCacher) Get(nonce uint64) (hash []byte, ok bool) {
	val, ok := nthc.cacher.Get(nthc.nonceToByteArray(nonce))

	if val != nil {
		hash = val.([]byte)
	}

	return
}

// Has checks if a nonce is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (nthc *NonceToHashCacher) Has(nonce uint64) bool {
	return nthc.cacher.Has(nthc.nonceToByteArray(nonce))
}

// Peek returns the nonce value (or nil if not found) without updating
// the "recently used"-ness of the nonce.
func (nthc *NonceToHashCacher) Peek(nonce uint64) (hash []byte, ok bool) {
	val, ok := nthc.cacher.Peek(nthc.nonceToByteArray(nonce))

	if val != nil {
		hash = val.([]byte)
	}

	return
}

// HasOrAdd checks if a nonce is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the hash.
// Returns whether found and whether an eviction occurred.
func (nthc *NonceToHashCacher) HasOrAdd(nonce uint64, hash []byte) (ok, evicted bool) {
	return nthc.cacher.HasOrAdd(nthc.nonceToByteArray(nonce), hash)
}

// Remove removes the provided nonce from the cache.
func (nthc *NonceToHashCacher) Remove(nonce uint64) {
	nthc.cacher.Remove(nthc.nonceToByteArray(nonce))
}

// RemoveOldest removes the oldest item from the cache.
func (nthc *NonceToHashCacher) RemoveOldest() {
	nthc.cacher.RemoveOldest()
}

// Keys returns a slice of nonce from the cache, from oldest to newest.
func (nthc *NonceToHashCacher) Keys() []uint64 {
	keys := nthc.cacher.Keys()

	nonces := make([]uint64, 0)

	for _, key := range keys {
		nonces = append(nonces, nthc.byteArrayToNonce(key))
	}

	return nonces
}

// Len returns the number of items in the cache.
func (nthc *NonceToHashCacher) Len() int {
	return nthc.cacher.Len()
}

// RegisterHandler registers a new handler to be called when a new data is added
func (nthc *NonceToHashCacher) RegisterHandler(handler func(nonce uint64)) {
	if handler == nil {
		log.Error("attempt to register a nil handler to a cacher object")
		return
	}

	handlerWrapper := func(key []byte) {
		nonce := nthc.byteArrayToNonce(key)
		handler(nonce)
	}

	nthc.cacher.RegisterHandler(handlerWrapper)
}

func (nthc *NonceToHashCacher) nonceToByteArray(nonce uint64) []byte {
	buff := make([]byte, 8)
	binary.BigEndian.PutUint64(buff, nonce)
	return buff
}

func (nthc *NonceToHashCacher) byteArrayToNonce(buff []byte) uint64 {
	return binary.BigEndian.Uint64(buff)
}
