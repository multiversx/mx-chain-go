package dataPool

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var log = logger.NewDefaultLogger()

// NonceToHashCacher is a wrapper over a storage.Cacher interface that has the key of type uint64
// and value of type byte slice
type NonceToHashCacher struct {
	cacher         storage.Cacher
	nonceConverter typeConverters.Uint64ByteSliceConverter
}

// NewNonceToHashCacher returns a new instance of NonceToHashCacher
func NewNonceToHashCacher(
	cacher storage.Cacher,
	nonceConverter typeConverters.Uint64ByteSliceConverter) (*NonceToHashCacher, error) {

	if cacher == nil {
		return nil, data.ErrNilCacher
	}

	if nonceConverter == nil {
		return nil, data.ErrNilNonceConverter
	}

	return &NonceToHashCacher{
		cacher:         cacher,
		nonceConverter: nonceConverter,
	}, nil
}

//func NewNonceToHashCacher

// Clear is used to completely clear the cache.
func (nthc *NonceToHashCacher) Clear() {
	nthc.cacher.Clear()
}

// Put adds a value to the cache.  Returns true if an eviction occurred.
func (nthc *NonceToHashCacher) Put(nonce uint64, hash []byte) (evicted bool) {
	return nthc.cacher.Put(nthc.nonceConverter.ToByteSlice(nonce), hash)
}

// Get looks up for a nonce in cache.
func (nthc *NonceToHashCacher) Get(nonce uint64) ([]byte, bool) {
	val, ok := nthc.cacher.Get(nthc.nonceConverter.ToByteSlice(nonce))

	if !ok {
		return []byte(nil), ok
	}

	return val.([]byte), ok
}

// Has checks if a nonce is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (nthc *NonceToHashCacher) Has(nonce uint64) bool {
	return nthc.cacher.Has(nthc.nonceConverter.ToByteSlice(nonce))
}

// Peek returns the nonce value (or nil if not found) without updating
// the "recently used"-ness of the nonce.
func (nthc *NonceToHashCacher) Peek(nonce uint64) ([]byte, bool) {
	val, ok := nthc.cacher.Peek(nthc.nonceConverter.ToByteSlice(nonce))

	if !ok {
		return []byte(nil), ok
	}

	return val.([]byte), ok
}

// HasOrAdd checks if a nonce is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the hash.
// Returns whether found and whether an eviction occurred.
func (nthc *NonceToHashCacher) HasOrAdd(nonce uint64, hash []byte) (ok, evicted bool) {
	return nthc.cacher.HasOrAdd(nthc.nonceConverter.ToByteSlice(nonce), hash)
}

// Remove removes the provided nonce from the cache.
func (nthc *NonceToHashCacher) Remove(nonce uint64) {
	nthc.cacher.Remove(nthc.nonceConverter.ToByteSlice(nonce))
}

// RemoveOldest removes the oldest item from the cache.
func (nthc *NonceToHashCacher) RemoveOldest() {
	nthc.cacher.RemoveOldest()
}

// Keys returns a slice of nonces from the cache, from oldest to newest.
func (nthc *NonceToHashCacher) Keys() []uint64 {
	keys := nthc.cacher.Keys()

	nonces := make([]uint64, 0)

	for _, key := range keys {
		nonce, err := nthc.nonceConverter.ToUint64(key)

		if err == nil {
			nonces = append(nonces, nonce)
		}
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
		nonce, err := nthc.nonceConverter.ToUint64(key)
		if err == nil {
			handler(nonce)
		}
	}

	nthc.cacher.RegisterHandler(handlerWrapper)
}
