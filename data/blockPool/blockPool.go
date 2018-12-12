package blockPool

import (
	"encoding/binary"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var defaultCacherConfig = &storage.CacheConfig{
	Size: 100,
	Type: storage.LRUCache,
}

// BlockPool holds cashers of headers and bodies
type BlockPool struct {
	headerStore storage.Cacher
	bodyStore   storage.Cacher

	cacherConfig *storage.CacheConfig

	headerHandlerLock sync.RWMutex
	addHeaderHandlers []func(nonce uint64)

	bodyHandlerLock sync.RWMutex
	addBodyHandlers []func(nonce uint64)
}

// NewBlockPool creates a BlockPool object which contains an empty pool of headers and bodies
func NewBlockPool(cacherConfig *storage.CacheConfig) *BlockPool {
	if cacherConfig == nil {
		cacherConfig = defaultCacherConfig
	}

	bp := &BlockPool{cacherConfig: cacherConfig}

	var err error

	bp.headerStore, err = storage.NewCache(cacherConfig.Type, cacherConfig.Size)

	if err != nil {
		return nil
	}

	bp.bodyStore, err = storage.NewCache(cacherConfig.Type, cacherConfig.Size)

	if err != nil {
		return nil
	}

	return bp
}

// HeaderStore method returns the header store which contains block nonces and headers
func (bp *BlockPool) HeaderStore() (c storage.Cacher) {
	return bp.headerStore
}

// BodyStore method returns the body store which contains block nonces and bodies
func (bp *BlockPool) BodyStore() (c storage.Cacher) {
	return bp.bodyStore
}

// AddHeader method adds a block header to the corresponding pool
func (bp *BlockPool) AddHeader(nonce uint64, hdr *block.Header) {
	mp := bp.HeaderStore()

	key := make([]byte, 8)
	binary.PutUvarint(key, nonce)

	found, _ := mp.HasOrAdd(key, hdr)

	if bp.addHeaderHandlers != nil && !found {
		for _, handler := range bp.addHeaderHandlers {
			go handler(nonce)
		}
	}
}

// AddBody method adds a block body to the corresponding pool
func (bp *BlockPool) AddBody(nonce uint64, blk *block.Block) {
	mp := bp.BodyStore()

	key := make([]byte, 8)
	binary.PutUvarint(key, nonce)

	found, _ := mp.HasOrAdd(key, blk)

	if bp.addBodyHandlers != nil && !found {
		for _, handler := range bp.addBodyHandlers {
			go handler(nonce)
		}
	}
}

// Header method returns the block header with given nonce
func (bp *BlockPool) Header(nonce uint64) (interface{}, bool) {
	mp := bp.HeaderStore()

	key := make([]byte, 8)
	binary.PutUvarint(key, nonce)

	return mp.Get(key)
}

// Body method returns the block body with given nonce
func (bp *BlockPool) Body(nonce uint64) (interface{}, bool) {
	mp := bp.BodyStore()

	key := make([]byte, 8)
	binary.PutUvarint(key, nonce)

	return mp.Get(key)
}

// RemoveHeader method removes a block header with given nonce
func (bp *BlockPool) RemoveHeader(nonce uint64) {
	key := make([]byte, 8)
	binary.PutUvarint(key, nonce)

	bp.headerStore.Remove(key)
}

// RemoveBody method removes a block body with given nonce
func (bp *BlockPool) RemoveBody(nonce uint64) {
	key := make([]byte, 8)
	binary.PutUvarint(key, nonce)

	bp.bodyStore.Remove(key)
}

// ClearHeaderStore deletes all the block headers from the coresponding pool
func (bp *BlockPool) ClearHeaderStore() {
	bp.headerStore.Clear()
}

// ClearBodyStore deletes all the block bodies from the coresponding pool
func (bp *BlockPool) ClearBodyStore() {
	bp.bodyStore.Clear()
}

// RegisterHeaderHandler will register a new call back function wich will be notified when a new header will be added
func (bp *BlockPool) RegisterHeaderHandler(headerHandler func(nonce uint64)) {
	bp.headerHandlerLock.Lock()

	if bp.addHeaderHandlers == nil {
		bp.addHeaderHandlers = make([]func(nonce uint64), 0)
	}

	bp.addHeaderHandlers = append(bp.addHeaderHandlers, headerHandler)

	bp.headerHandlerLock.Unlock()
}

// RegisterBodyHandler will register a new call back function wich will be notified when a new body will be added
func (bp *BlockPool) RegisterBodyHandler(bodyHandler func(nonce uint64)) {
	bp.bodyHandlerLock.Lock()

	if bp.addBodyHandlers == nil {
		bp.addBodyHandlers = make([]func(nonce uint64), 0)
	}

	bp.addBodyHandlers = append(bp.addBodyHandlers, bodyHandler)

	bp.bodyHandlerLock.Unlock()
}
