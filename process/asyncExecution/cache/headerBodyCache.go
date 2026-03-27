package cache

import (
	"slices"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/asyncExecution/cache")

const (
	defaultMaxQueueSize = 1000
)

// headerBodyCache stores header-body pairs by nonce for async execution.
//
// Optimization: signalBlockAdded is a buffered(1) channel that notifies the
// headersExecutor immediately when a new block is queued via AddOrReplace.
// This replaces the old 5ms polling loop, reducing execution start latency
// from 0-5ms to <1ms. The non-blocking send (select/default) ensures
// AddOrReplace never blocks even if the consumer hasn't drained the signal.
type headerBodyCache struct {
	mutex            sync.RWMutex
	cacheByNonce     map[uint64]HeaderBodyPair
	maxCacheSize     int
	signalBlockAdded chan struct{}
}

// NewHeaderBodyCache will create a new instance of cache
func NewHeaderBodyCache(config config.HeaderBodyCacheConfig) *headerBodyCache {
	cacheSize := config.Capacity
	if cacheSize == 0 {
		cacheSize = defaultMaxQueueSize
	}

	return &headerBodyCache{
		cacheByNonce:     make(map[uint64]HeaderBodyPair),
		mutex:            sync.RWMutex{},
		maxCacheSize:     cacheSize,
		signalBlockAdded: make(chan struct{}, 1),
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
	if pair.HeaderHash == nil {
		return common.ErrNilHeaderHash
	}

	c.mutex.Lock()

	headerNonce := pair.Header.GetNonce()
	_, found := c.cacheByNonce[headerNonce]
	if len(c.cacheByNonce) >= c.maxCacheSize && !found {
		currentSize := len(c.cacheByNonce)
		c.mutex.Unlock()
		log.Warn("async execution cache is full",
			"current size", currentSize,
			"max size", c.maxCacheSize,
			"rejected nonce", headerNonce,
		)
		return ErrCacheIsFull
	}

	c.cacheByNonce[headerNonce] = pair
	cacheSize := len(c.cacheByNonce)
	c.mutex.Unlock()

	log.Debug("headerBodyCache.AddOrReplace - block has been added",
		"round", pair.Header.GetRound(),
		"nonce", pair.Header.GetNonce(),
		"hash", pair.HeaderHash,
		"cache size", cacheSize)

	// Signal outside the lock to avoid coupling notification to lock scope
	select {
	case c.signalBlockAdded <- struct{}{}:
	default:
	}

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
	nonces := make([]uint64, 0)
	c.mutex.Lock()
	for nonce := range c.cacheByNonce {
		if nonce >= providedNonce {
			delete(c.cacheByNonce, nonce)
			nonces = append(nonces, nonce)
		}
	}
	c.mutex.Unlock()

	slices.Sort(nonces)

	return nonces
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

// GetSignalBlockAddedChan returns a receive-only channel that signals when a block is added.
func (c *headerBodyCache) GetSignalBlockAddedChan() <-chan struct{} {
	return c.signalBlockAdded
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *headerBodyCache) IsInterfaceNil() bool {
	return c == nil
}
