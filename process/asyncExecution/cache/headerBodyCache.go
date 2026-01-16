package cache

import (
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
)

type headerBodyCache struct {
	mutex          sync.RWMutex
	cacheByNonce   map[uint64]HeaderBodyPair
	lastAddedNonce uint64
}

// NewHeaderBodyCache will create a new instance of cache
func NewHeaderBodyCache() *headerBodyCache {
	return &headerBodyCache{
		cacheByNonce: make(map[uint64]HeaderBodyPair),
		mutex:        sync.RWMutex{},
	}
}

// AddOrReplace will add or replace the provided pair based on header's nonce
func (c *headerBodyCache) AddOrReplace(pair HeaderBodyPair) error {
	if check.IfNil(pair.Header) {
		return common.ErrNilHeaderHandler
	}
	if check.IfNil(pair.Body) {
		return data.ErrNilBlockBody
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	headerNonce := pair.Header.GetNonce()
	c.cacheByNonce[headerNonce] = pair
	c.lastAddedNonce = headerNonce

	return nil
}

// GetByNonce will return the pair based on the provided nonce
func (c *headerBodyCache) GetByNonce(nonce uint64) (HeaderBodyPair, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	pair, found := c.cacheByNonce[nonce]

	return pair, found
}

// RemoveAtNonceAndHigher will remove all pairs with the provided nonce or higher
func (c *headerBodyCache) RemoveAtNonceAndHigher(providedNonce uint64) []uint64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	nonces := make([]uint64, 0)
	for nonce, _ := range c.cacheByNonce {
		if nonce >= providedNonce {
			delete(c.cacheByNonce, nonce)
			nonces = append(nonces, nonce)
		}
	}

	sort.Slice(nonces, func(i, j int) bool { return nonces[i] < nonces[j] })

	return nonces
}

// GetLastAdded will return the lat added nonce
func (c *headerBodyCache) GetLastAdded() (HeaderBodyPair, bool) {
	pair, found := c.cacheByNonce[c.lastAddedNonce]

	return pair, found
}

// Remove will remove a pair by provided nonce
func (c *headerBodyCache) Remove(nonce uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cacheByNonce, nonce)
}

// Clean will cleanup the cache
func (c *headerBodyCache) Clean() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cacheByNonce = make(map[uint64]HeaderBodyPair)
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *headerBodyCache) IsInterfaceNil() bool {
	return c == nil
}
